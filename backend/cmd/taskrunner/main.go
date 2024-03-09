package main

import (
	"log"
	"time"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/model"
	"github.com/0xturkey/jd-go/pkg/handlers"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func handleTaskError(err error, taskID uuid.UUID) {
	// Placeholder for the actual error handling logic
	log.Printf("Error running task %s: %s\n", taskID, err)
}

func main() {
	godotenv.Load()
	database.ConnectDB()
	db := database.DB

	taskInterval := 30 * time.Second

	for {
		log.Println("Checking for new tasks")

		var tasks []model.Task
		now := time.Now()

		log.Printf("Current time: %s\n", now)

		// Get tasks that are scheduled and past the schedule time
		result := db.Where("task_status = ? AND scheduled_at < ?", model.Scheduled, now).Order("scheduled_at desc").Find(&tasks)
		if result.Error != nil {
			log.Printf("Error retrieving tasks: %s\n", result.Error)

			// Sleep for the defined interval before checking for new tasks again
			time.Sleep(taskInterval)

			continue
		}

		for _, task := range tasks {
			log.Printf("Current task: %+v\n", task)

			// Run the task and handle any potential errors
			if err := handlers.Run(&task); err != nil {
				// set the transaction to failed

				handleTaskError(err, task.ID)
			}
		}

		// Sleep for the defined interval before checking for new tasks again
		time.Sleep(taskInterval)
	}
}
