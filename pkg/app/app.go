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
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog/log"
)

// App holds all dependencies for our application
type App struct {
	Db       *db.Service
	Fiber    *fiber.App
	Config   *config.Config
	Store    *session.Store
	Auth     *middleware.AuthMiddleware
	Calendar *calendar.Service
}

// New creates a new instance of App with all dependencies
func New(cfg *config.Config) (*App, error) {
	// Initialize logger
	logger.Setup(cfg.Server.Environment)

	// Initialize DB service
	db, err := db.NewService(db.Config{
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
	authMiddleware, err := middleware.NewAuthMiddleware(db, middleware.DefaultSessionOptions())
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize auth middleware")
		fmt.Errorf("unable to initialize auth middleware: %v", err)
	}

	// Initialize calendar service with the DB pool
	calendarService := calendar.NewService(db.GetPool())

	return &App{
		Db:       db,
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
	authHandler := handlers.NewAuthHandler(a.Db, a.Auth)
	userHandler := handlers.NewUserHandler(a.Db)
	facilityHandler := handlers.NewFacilityHandler(a.Db)
	// controllersHandler := handlers.NewControllerHandler(a.Db)
	// scheduleHandler := handlers.NewScheduleHandler(a.Db)
	calendarHandler := handlers.NewCalendarHandler(a.Calendar)

	// Unprotected Routes
	a.Fiber.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, 👋. Welcome to Weekend-Warrior")
	})
	a.Fiber.Get("/request-access", func(c *fiber.Ctx) error {
		return c.SendString("Hello, 👋. A request access form is coming soon.")
	})
	a.Fiber.Get("/login", authHandler.LoginForm)
	a.Fiber.Post("/login", authHandler.HandleLogin)
	a.Fiber.Get("/logout", authHandler.LogoutForm)
	a.Fiber.Post("/logout", a.Auth.Logout)

	super := a.Fiber.Group("/super", a.Auth.Protected(), a.Auth.SuperOnly())
	super.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! 👋. You must be an super.")
	})
	super.Get("/facilities", facilityHandler.GetFacilities)
	super.Post("/facilities", facilityHandler.CreateFacility)
	super.Get("/facilities/create", facilityHandler.CreateForm)
	super.Get("/facilities/:id/edit", facilityHandler.EditForm)

	admin := a.Fiber.Group("/admin", a.Auth.Protected(), a.Auth.AdminOnly())
	admin.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! 👋. You must be an admin.")
	})

	controllers := admin.Group("/controllers")
	// List all controllers
	controllers.Get("/", userHandler.List)
	controllers.Post("/", userHandler.Create)
	controllers.Put("/:id", userHandler.Update)
	controllers.Delete("/:id", userHandler.Delete)
	// Create new controller
	controllers.Get("/new", userHandler.CreateForm)
	// Update existing controller
	controllers.Get("/edit/:id", userHandler.UpdateForm)
	// Assign schedule to controller
	controllers.Get("/schedule/:id", userHandler.ScheduleForm)

	// Protected
	app := a.Fiber.Group("/app", a.Auth.Protected())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! 👋. You are an authorized user.")
	})

	// Controller Routes (Protected)
	// View own facility info
	app.Get("/facility", facilityHandler.GetUserFacility)
	// Schedule viewing
	// app.Get("/schedule", GetUserSchedule)
	// app.Post("/schedule", ToggleAvailability)

	facilities := app.Group("/facilities")
	facilities.Get("/", facilityHandler.GetFacilities)
	facilities.Get("/create", facilityHandler.CreateForm)
	facilities.Post("/create", facilityHandler.CreateFacility)
	facilities.Delete("/:code", facilityHandler.DeleteFacility)
	facilities.Put("/:code", func(c *fiber.Ctx) error {
		return c.SendString("Update facility endpoint stub.")
	})

	// Setup root route
	a.Fiber.Get("/calendarExample", calendarHandler.CalendarHandler)
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
