// pkg/app/app.go
package app

import (
	"fmt"
	"time"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/logger"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/dukerupert/weekend-warrior/pkg/config"
	"github.com/dukerupert/weekend-warrior/services/calendar"
	"github.com/dukerupert/weekend-warrior/website/handlers"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
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

	// Initialize Session middleware with storage
	storage := postgres.New(postgres.Config{
		DB:         db.GetPool(),
		Table:      "fiber_storage",
		Reset:      false,
		GCInterval: 10 * time.Second,
	})
	store := session.New(session.Config{
		Storage: storage,
	})

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
		Store:    store,
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
	a.Fiber.Post("/logout", authHandler.HandleLogout)

	// Testing new fiber middleware
	a.Fiber.Get("/test/protected", func(c *fiber.Ctx) error {
		log.Info().Msg("Starting protected route middleware check")

		// Authentication middleware
		sess, err := a.Store.Get(c)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", c.Path()).
				Msg("Failed to get session")
			return c.Redirect("/")
		}

		log.Debug().
			Str("session_id", sess.ID()).
			Msg("Session retrieved successfully")

		userID := sess.Get("user_id")
		if userID == nil {
			log.Warn().
				Str("session_id", sess.ID()).
				Msg("No user_id found in session")
			return c.Redirect("/")
		}

		// Log all session values for debugging
		log.Debug().
			Interface("user_id", userID).
			Interface("role", sess.Get("role")).
			Interface("facility_id", sess.Get("facility_id")).
			Msg("Session values")

		// Add user info to locals for use in handlers
		c.Locals("user_id", userID)
		c.Locals("role", sess.Get("role"))
		c.Locals("facility_id", sess.Get("facility_id"))

		log.Info().
			Interface("user_id", userID).
			Interface("role", sess.Get("role")).
			Interface("facility_id", sess.Get("facility_id")).
			Msg("Authentication successful, proceeding to handler")

		return c.Next()
	}, func(c *fiber.Ctx) error {
		// Actual route handler
		log.Info().
			Str("path", c.Path()).
			Interface("user_id", c.Locals("user_id")).
			Interface("role", c.Locals("role")).
			Msg("Protected route accessed successfully")

		return c.Render("pages/logout", fiber.Map{
		"title": "Logout",
		"error": c.Query("error"),
	}, "layouts/base")
	})

	a.Fiber.Get("/test/logout", func(c *fiber.Ctx) error {
		return c.Render("pages/logout", fiber.Map{
		"title": "Logout",
		"error": c.Query("error"),
	}, "layouts/base")
	})

	a.Fiber.Post("/test/logout", func(c *fiber.Ctx) error {
		sess, err := a.Store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get session",
			})
		}

		if err := sess.Destroy(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to destroy session",
			})
		}

		return c.Redirect("/test/protected")
	})

	a.Fiber.Get("/test/login", authHandler.LoginForm)

	a.Fiber.Post("/test/login", func(c *fiber.Ctx) error {
		// Get session
		sess, err := a.Store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get session",
			})
		}

		// Parse login request
		var login struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&login); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// ... your database query here ...
		user, err := db.GetLoginResponse(a.Db, login.Email)
		if err != nil {
			return c.Redirect("/test/login?error=Invalid+credentials", fiber.StatusFound)
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Store user information in session
		log.Info().Msg("Starting to set session values")

		sess.Set("user_id", int64(user.ID))
		log.Debug().
			Int("user_id", user.ID).
			Msg("Set user_id in session")

		sess.Set("role", string(user.Role))
		log.Debug().
			Str("role", string(user.Role)).
			Msg("Set role in session")

		sess.Set("facility_id", user.FacilityID)
		log.Debug().
			Int("facility_id", user.FacilityID).
			Msg("Set facility_id in session")

		log.Info().
			Str("session_id", sess.ID()).
			Msg("Session saved successfully")

			// Must save before redirect!
		if err := sess.Save(); err != nil {
			log.Error().Err(err).Msg("Failed to save session")
			return err
		}

		// Redirect based on role
		switch user.Role {
		case "super":
			return c.Redirect("/test/protected")
		case "admin":
			return c.Redirect("/test/protected")
		default:
			return c.Redirect("/test/protected")
		}
	})

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
	//app.Get("/schedule", GetUserSchedule)
	//app.Post("/schedule", ToggleAvailability)

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
