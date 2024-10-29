package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	fmt.Println("Hello, let's start solving problems!")

	// Create a new engine
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		// Pass in Views Template Engine
		Views: engine,

		// Default global path to search for Views
		ViewsLayout: "layouts/main",

		// Enables/Disables access to `ctx.Locals()` entries in rendered view
		// (defaults to false)
		PassLocalsToViews: false,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	app.Get("/layout", func(c *fiber.Ctx) error {
		// Render index within layouts/main
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	app.Get("/layouts-nested", func(c *fiber.Ctx) error {
		// Render index within layouts/nested/main within layouts/nested/base
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/nested/main", "layouts/nested/base")
	})

	app.Listen(":3000")
}
