package main

import (
	"fmt"
	"context"
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/jackc/pgx/v5"
	"github.com/dukerupert/weekend-warrior/config"
	"github.com/dukerupert/weekend-warrior/services"
	"github.com/dukerupert/weekend-warrior/handlers"
)

// Config holds all configuration for our application
type Config struct {
    DatabaseURL string
    Port        string
}

// App holds all dependencies for our application
type App struct {
    DB     *pgx.Conn
    Fiber  *fiber.App
    Config *config.Config
}

func NewApp(cfg *config.Config) (*App, error) {
    // Connect to database using config
    conn, err := pgx.Connect(context.Background(), cfg.GetDatabaseURL())
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %v", err)
    }

    // Create Fiber instance with config
    fiberApp := fiber.New(fiber.Config{
        ReadTimeout:       cfg.Server.ReadTimeout,
        WriteTimeout:      cfg.Server.WriteTimeout,
        Views:            html.New("./views", ".html"),
        ViewsLayout:      "layouts/main",
        PassLocalsToViews: false,
    })

    return &App{
        DB:     conn,
        Fiber:  fiberApp,
        Config: cfg,
    }, nil
}

// Setup configures our routes and middleware
func (a *App) Setup() {
    // Store DB connection in context for handlers to use
    a.Fiber.Use(func(c *fiber.Ctx) error {
        c.Locals("db", a.DB)
        return c.Next()
    })

	// Create calendar service
    calendarService := calendar.NewService(a.DB)
    
    // Create calendar handler
    calendarHandler := handlers.NewCalendarHandler(calendarService)

    // Setup routes
    a.Fiber.Get("/", calendarHandler.CalendarHandler)
}

// Start begins listening for requests
func (a *App) Start() error {
    return a.Fiber.Listen(":" + a.Config.Server.Port)
}

// Cleanup handles graceful shutdown
func (a *App) Cleanup() {
    if err := a.DB.Close(context.Background()); err != nil {
        log.Printf("Error closing DB connection: %v", err)
    }
}

func main() {
	// Load configuration
    cfg, err := config.LoadConfig(".env")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Create new app instance with configuration
    app, err := NewApp(cfg)
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
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
