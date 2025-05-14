package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler handles errors for Fiber
var ErrorHandler = func(c *fiber.Ctx, err error) error {
	return c.Status(500).JSON(fiber.Map{
		"error": err.Error(),
	})
}
