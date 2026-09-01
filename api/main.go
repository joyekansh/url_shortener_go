package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

// Rounting function
// Setting the URL paths
func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)    // Define a route to resolve shortened URLs
	app.Post("/api/v1", routes.ShortenURL) // Define an API endpoint to shorten URLs
}

func main() {
	// Load environment variables from a .env file (if available)
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New() // new fiber application instance

	app.Use(logger.New()) // applying the fiber logger middle ware to log HTTP request

	setupRoutes(app) // Sets routing for the application

	// Starting the fiber application and listen on the port specified in the environment variables
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
