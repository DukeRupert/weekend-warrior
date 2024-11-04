// main.go
package main

import (
	"log"

	"github.com/dukerupert/weekend-warrior/pkg/app"
	"github.com/dukerupert/weekend-warrior/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and setup application
	app, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer app.Cleanup()

	// Setup routes and middleware
	app.Setup()

	// Start the server
	log.Printf("Starting server on port %s in %s mode",
		app.Config.Server.Port,
		app.Config.Server.Environment,
	)

	if err := app.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
