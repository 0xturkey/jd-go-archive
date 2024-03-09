package handlers

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/model"
	"github.com/0xturkey/jd-go/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

/*
func CreateTask(c *fiber.Ctx) error {
	validate := utils.NewValidator()

	type CreateTaskReq struct {
		WalletAddress      string     `json:"wallet_address" validate:"required,isEthAddress"`
		RpcUrl             string     `json:"rpc_url" validate:"required,isHTTPSURL"`
		EncodedTransaction string     `json:"encoded_transaction" validate:"required,isHexString"`
		DependentTaskId    *uuid.UUID `json:"dependent_task_id" validate:"omitempty,isValidUUID"`
		ScheduledAt        string     `json:"scheduled_at" validate:"required,isRFC3339Time"`
		TaskType           string     `json:"task_type" validate:"required,validateTaskType"`
	}
	json := new(CreateTaskReq)
	var err error
	if err = c.BodyParser(json); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid Request Data",
		})
	}

	err = validate.Struct(json)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": err.Error(),
		})
	}

	// find or create wallet reference
	userID := c.Locals("user_id").(uuid.UUID)
	wallet, err := FindOrCreateWallet(json.WalletAddress, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Unable to find or create wallet",
		})
	}

	scheduledAt, err := time.Parse(time.RFC3339, json.ScheduledAt)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid ScheduledAt",
		})
	}

	new := model.Task{
		ID:                 uuid.New(),
		UserRefer:          userID,
		WalletRefer:        wallet.ID,
		TaskStatus:         model.Scheduled,
		ScheduledAt:        scheduledAt,
		RpcUrl:             json.RpcUrl,
		EncodedTransaction: json.EncodedTransaction,
		DependentTaskID:    json.DependentTaskId,
		TaskType:           model.TaskType(json.TaskType),
	}
	db := database.DB
	db.Create(&new)

	return c.JSON(fiber.Map{
		"message": "success",
		"task":    new,
	})
}
*/

func GetTasks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	db := database.DB
	var tasks []model.Task
	db.Where("user_refer = ?", userID).Find(&tasks)
	return c.JSON(fiber.Map{
		"tasks": tasks,
	})
}

func refreshTaskStatus(task *model.Task) (*model.Task, error) {
	if task.TransactionID == nil || *task.TransactionID == "" {
		return task, fmt.Errorf("task %s has no transaction ID", task.ID)
	}

	if task.TaskStatus != model.InProgress {
		return task, nil
	}

	newStatus, err := CheckTaskStatus(task)
	if err != nil {
		log.Printf("Error checking status of tx %s: %s\n", task.ID, err)
	}

	if newStatus != task.TaskStatus {
		// get from db and then return
		db := database.DB
		var updatedTask model.Task
		result := db.First(&updatedTask, "id = ?", task.ID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Handle the case where the task is not found
				return nil, nil
			}
			// Handle any other possible error
			return nil, result.Error
		}
		return &updatedTask, nil
	}

	return task, nil
}

func checkTaskDependencies(dependencies *[]model.Task) (allFinished bool, latestTimestamp time.Time, err error) {
	allFinished = true

	for _, dep := range *dependencies {
		if dep.TransactionID == nil || *dep.TransactionID == "" {
			allFinished = false
			continue
		}

		updatedDep, err := refreshTaskStatus(&dep)
		if err != nil {
			return false, time.Time{}, err
		}

		updatedDepTime := time.Unix(updatedDep.UpdatedAt, 0)

		if updatedDepTime.After(latestTimestamp) {
			latestTimestamp = updatedDepTime
		}

		if dep.TaskStatus != model.Completed {
			allFinished = false
		}
	}

	return allFinished, latestTimestamp, nil
}

// runTask runs a task if all dependencies have finished.
func Run(task *model.Task) error {
	// Check all dependencies have finished
	if len(task.Dependencies) > 0 {
		log.Println("checking if all dependencies have finished")

		allFinished, latestTimestamp, err := checkTaskDependencies(&task.Dependencies)
		if err != nil {
			return err
		}

		log.Printf("whether all dependencies have finished: %v", allFinished)

		if !allFinished {
			log.Println("All dependencies have not finished. Returning")
			return nil
		}

		minMinutesBetween := utils.GetRandomInt(2, 5)
		if time.Now().Unix() < latestTimestamp.Unix()+int64(minMinutesBetween*60) {
			log.Println("Not long enough between last dep finishing. returning")
			return nil
		}
	}

	log.Printf("Running task: encodedArgs=%v, walletId=%v", task.EncodedTransaction, task.WalletRefer)

	// Here you would have the logic to get the task handler based on the task type
	// and run the task. Since this is a mock, we'll simulate it.
	taskHandler, err := utils.GetTaskHandler(task.TaskType)
	if err != nil {
		return err
	}

	result, err := taskHandler.RunTask(task.EncodedTransaction, task.RpcUrl)
	if err != nil {
		return err
	}

	// Update the task status to InProgress
	db := database.DB
	update := db.Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"task_status":    model.InProgress,
		"transaction_id": result,
	})
	if update.Error != nil {
		return update.Error
	}

	log.Printf("row updated: %+v", update)
	return nil
}

func CheckTaskStatus(task *model.Task) (model.TaskStatus, error) {
	taskHandler, err := utils.GetTaskHandler(task.TaskType)
	if err != nil {
		return model.Unknown, err
	}

	// Check if task.TransactionID is not nil to avoid dereferencing a nil pointer
	if task.TransactionID == nil {
		return model.Unknown, errors.New("transaction ID is nil")
	}

	status, err := taskHandler.GetTaskStatus(*task.TransactionID, task.RpcUrl)
	if err != nil {
		return model.Unknown, err
	}

	// no need to update if status is the same
	if status == task.TaskStatus {
		return status, nil
	}

	db := database.DB
	update := db.Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"task_status": status,
	})
	if update.Error != nil {
		return model.Unknown, update.Error
	}

	log.Printf("row updated: %+v", update)
	return status, nil
}
