package router

import (
	"github.com/0xturkey/jd-go/pkg/handlers"
	"github.com/0xturkey/jd-go/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func Initalize(router *fiber.App) {
	router.Use(middleware.Security)

	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello, World!")
	})

	router.Use(middleware.Json)

	users := router.Group("/users")
	users.Post("/", handlers.CreateUser)
	users.Get("/me", middleware.JWTProtected, handlers.GetUserInfo)
	users.Post("/login", handlers.Login)
	users.Get("/exists/:email", handlers.UserExists)
	users.Get("/wallet", middleware.JWTProtected, handlers.GetWallet)
	users.Post("/wallet/export", middleware.JWTProtected, handlers.ExportWallet)

	tasks := router.Group("/tasks")
	// tasks.Post("/", middleware.JWTProtected, handlers.CreateTask)
	tasks.Get("/", middleware.JWTProtected, handlers.GetTasks)

	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Not Found",
		})
	})
}
