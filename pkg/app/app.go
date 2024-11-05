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
	Db       *db.Service
	Fiber    *fiber.App
	Config   *config.Config
	Auth     *middleware.AuthMiddleware
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
	app := fiber.New(fiber.Config{
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Views:             html.New("./website/views", ".html"),
		PassLocalsToViews: false,
	})

	// Add logger middleware
	app.Use(middleware.Logger())

	// Initialize auth middleware with default options
	authMiddleware, err := middleware.NewAuthMiddleware(dbService, middleware.DefaultSessionOptions())
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize auth middleware")
		fmt.Errorf("unable to initialize auth middleware: %v", err)
	}

	// Initialize calendar service with the DB pool
	calendarService := calendar.NewService(dbService.GetPool())

	return &App{
		Db:       dbService,
		Fiber:    app,
		Config:   cfg,
		Auth:     authMiddleware,
		Calendar: calendarService,
	}, nil
}

// Setup configures our routes and middleware
func (a *App) Setup() {
	// Store DB pool in context for handlers to use
	a.Fiber.Use(func(c *fiber.Ctx) error {
		c.Locals("db", a.Db.GetPool())
		return c.Next()
	})

	// Create and register handlers
	a.setupHandlers()
}

// setupHandlers initializes and registers all handlers
func (a *App) setupHandlers() {
	loginHandler := handlers.NewLoginHandler(a.Db, a.Auth)
	registerHandler := handlers.NewRegisterHandler(a.Db, a.Auth)
	facilityHandler := handlers.NewFacilityHandler(a.Db)
	// controllersHandler := handlers.NewControllerHandler(a.Db)
	// scheduleHandler := handlers.NewScheduleHandler(a.Db)
	calendarHandler := handlers.NewCalendarHandler(a.Calendar)

	// Unprotected Routes
	loginHandler.RegisterRoutes(a.Fiber)
	registerHandler.RegisterRoutes(a.Fiber)

	admin := a.Fiber.Group("/admin", a.Auth.Protected(), a.Auth.AdminOnly())
	admin.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! 👋. You must be an admin.")
	})

	// Protected
	app := a.Fiber.Group("/app", a.Auth.Protected())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! 👋. You are an authorized user.")
	})
	facilities := app.Group("/facilities")
	// List all facilities
	facilities.Get("/", facilityHandler.ListFacilities)
	// Create new facility endpoint
	facilities.Post("/", facilityHandler.CreateFacility)
	// Create new facility form
	facilities.Get("/create", facilityHandler.ShowCreateForm)
	// Delete facility by ID
	facilities.Delete("/:id", facilityHandler.DeleteFacility)
	// Get controllers at facility
	facilities.Get("/:code/controllers", facilityHandler.GetFacilityControllers)

	// Protected routes
	// controllersHandler.RegisterRoutes(v1)
	// scheduleHandler.RegisterRoutes(v1)

	// Setup root route
	a.Fiber.Get("/", calendarHandler.CalendarHandler)
}

// Start begins listening for requests
func (a *App) Start() error {
	return a.Fiber.Listen(":" + a.Config.Server.Port)
}

// Cleanup handles graceful shutdown
func (a *App) Cleanup() {
	if a.Db != nil {
		a.Db.Close()
	}
}
