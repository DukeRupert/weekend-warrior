// pkg/app/app.go
package app

import (
	"fmt"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/logger"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/dukerupert/weekend-warrior/pkg/config"
	"github.com/dukerupert/weekend-warrior/services/calendar"
	"github.com/dukerupert/weekend-warrior/website/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog/log"
)

// App holds all dependencies for our application
type App struct {
	DB       *db.Service
	Fiber    *fiber.App
	Config   *config.Config
	Calendar *calendar.Service
}

// New creates a new instance of App with all dependencies
func New(cfg *config.Config) (*App, error) {
	// Initialize logger
	logger.Setup(cfg.Server.Environment)

	// Initialize DB service
	dbService, err := db.NewService(db.Config{
		URL: cfg.GetDatabaseURL(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize database service")
		return nil, fmt.Errorf("unable to initialize database service: %v", err)
	}

	// Create Fiber instance with config
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Views:             html.New("./website/views", ".html"),
		PassLocalsToViews: false,
	})

	// Add logger middleware
	fiberApp.Use(middleware.Logger())

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

	// Create and register handlers
	a.setupHandlers()
}

// setupHandlers initializes and registers all handlers
func (a *App) setupHandlers() {
	// Initialize auth middleware with defeault options
	authMiddleware, err := middleware.NewAuthMiddleware(a.DB.GetPool(), middleware.DefaultSessionOptions())
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize auth middleware")
		fmt.Errorf("unable to initialize auth middleware: %v", err)
	}

	// Protected routes
	api := a.Fiber.Group("/app")
	api.Use(authMiddleware.Protected()) // Protect all routes under /api

	// Create login handler
	loginHandler := handlers.NewLoginHandler(a.DB, authMiddleware)
	loginHandler.RegisterRoutes(a.Fiber)

	// Create calendar handler
	calendarHandler := handlers.NewCalendarHandler(a.Calendar)

	// Initialize and register facility handler
	facilityHandler := handlers.NewFacilityHandler(a.DB)
	facilityHandler.RegisterRoutes(a.Fiber)

	// Initialize and register controllers handler
	controllersHandler := handlers.NewControllerHandler(a.DB)
	controllersHandler.RegisterRoutes(a.Fiber, authMiddleware)

	// Initialize and register schedule handlers
	scheduleHandler := handlers.NewScheduleHandler(a.DB)
	scheduleHandler.RegisterRoutes(a.Fiber)

	// Setup root route
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
