package main

import (
	"fmt"
	"log"
	"os"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/pkg/router"
	"github.com/0xturkey/jd-go/pkg/turnkey"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	godotenv.Load()
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // comma string format e.g. "localhost, nikschaefer.tech"
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	err := turnkey.Init(
		"api.turnkey.com",
		os.Getenv("TURNKEY_API_PRIVATE_KEY"),
		os.Getenv("TURNKEY_ORGANIZATION_ID"),
	)
	if err != nil {
		log.Fatalf("Unable to initialize Turnkey client: %+v", err)
	}

	var userID string
	userID, err = turnkey.Client.Whoami()
	if err != nil {
		log.Fatalf("Unable to use Turnkey client for whoami request: %+v", err)
	}

	fmt.Printf("User ID: %s\n", userID)

	database.ConnectDB()

	router.Initalize(app)
	log.Fatal(app.Listen(":" + getenv("PORT", "4000")))
}

/*
ENV Variables:
will auto set to 3000 if not set
PORT=3000
this should be a connection string or url
DATABASE_URL="host=localhost port=5432 user=postgres password= dbname= sslmode=disable"
**
Docker Command for Postgres database:
docker run --name database -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:alpine

DB_URL Variable for docker database
DATABASE_URL="host=localhost port=5432 user=postgres password=password dbname=postgres sslmode=disable"
**
Docker build base image in first stage for development
docker build --target build -t base .
**
run dev container
docker run -p 3000:3000 --mount type=bind,source="C:\Users\schaefer\go\src\fiber",target=/go/src/app --name fiber -td base
**
rebuild and run package
docker exec -it web go run main.go
**
stop and remove container
docker stop fiber; docker rm fiber
*/
