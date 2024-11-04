package main

import (
	"fmt"
	"log"

	"github.com/dukerupert/weekend-warrior/config"
	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/handlers"
	"github.com/dukerupert/weekend-warrior/services/calendar"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

// Config holds all configuration for our application
type Config struct {
	DatabaseURL string
	Port        string
}

// App holds all dependencies for our application
type App struct {
	DB       *db.Service
	Fiber    *fiber.App
	Config   *config.Config
	Calendar *calendar.Service
}

func NewApp(cfg *config.Config) (*App, error) {
	// Initialize DB service
	dbService, err := db.NewService(db.Config{
		URL: cfg.GetDatabaseURL(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize database service: %v", err)
	}

	// Create Fiber instance with config
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Views:             html.New("./views", ".html"),
		PassLocalsToViews: false,
	})

	// Initialize calendar service with the DB pool
	calendarService := calendar.NewService(dbService.GetPool())

	return &App{
		DB:       dbService,
		Fiber:    fiberApp,
		Config:   cfg,
		Calendar: calendarService,
	}, nil
}

// Setup configures our routes and middleware
func (a *App) Setup() {
	// Store DB pool in context for handlers to use
	a.Fiber.Use(func(c *fiber.Ctx) error {
		c.Locals("db", a.DB.GetPool())
		return c.Next()
	})

	// Create calendar handler
	calendarHandler := handlers.NewCalendarHandler(a.Calendar)

	// Initialize and register facility handler
	facilityHandler := handlers.NewFacilityHandler(a.DB)
	facilityHandler.RegisterRoutes(a.Fiber)

	// Initialize and register controllers handler
	controllersHandler := handlers.NewControllerHandler(a.DB)
	controllersHandler.RegisterRoutes(a.Fiber)

	// Setup routes
	a.Fiber.Get("/", calendarHandler.CalendarHandler)
}

// Start begins listening for requests
func (a *App) Start() error {
	return a.Fiber.Listen(":" + a.Config.Server.Port)
}

// Cleanup handles graceful shutdown
func (a *App) Cleanup() {
	if a.DB != nil {
		a.DB.Close()
	}
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".env.local")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create new app
	app, err := NewApp(cfg)
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()

	// Setup routes and middleware
	app.Setup()

	// Start the server
	log.Printf("Starting server on port %s in %s mode",
		cfg.Server.Port,
		cfg.Server.Environment,
	)
	if err := app.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
