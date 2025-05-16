// File: j-ticketing/internal/core/routes/view_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
)

// SetupViewRoutes configures all view-related routes
func SetupViewRoutes(app *fiber.App) {
	// Set the views directory
	viewsDir := "./pkg/views"

	// Checkout page - support both /checkout and /checkout.html paths
	app.Get("/checkout", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/checkout.html")
	})

	app.Get("/checkout.html", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/checkout.html")
	})

	// Success page
	app.Get("/success", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/success.html")
	})

	app.Get("/success.html", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/success.html")
	})

	// Failure page
	app.Get("/failure", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/failure.html")
	})

	app.Get("/failure.html", func(c *fiber.Ctx) error {
		return c.SendFile(viewsDir + "/failure.html")
	})
}
