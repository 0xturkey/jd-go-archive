package main

import (
	"log"
	"time"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/model"
	"github.com/0xturkey/jd-go/pkg/handlers"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.ConnectDB()
	db := database.DB

	taskInterval := 120 * time.Second

	for {
		log.Println("Checking for updates to tasks")

		var tasks []model.Task

		// Get tasks that are scheduled and past the schedule time
		result := db.Where("task_status = ?", model.InProgress).Find(&tasks)
		if result.Error != nil {
			log.Printf("Error retrieving tasks: %s\n", result.Error)

			// Sleep for the defined interval before checking for new tasks again
			time.Sleep(taskInterval)

			continue
		}

		for _, task := range tasks {
			log.Printf("Current task: %+v\n", task)

			// Run the task and handle any potential errors
			if _, err := handlers.CheckTaskStatus(&task); err != nil {
				// ignore for now but should figure out how to handle this
				log.Printf("Error checking status of tx %s: %s\n", task.ID, err)
			}
		}

		// Sleep for the defined interval before checking for new tasks again
		time.Sleep(taskInterval)
	}
}
