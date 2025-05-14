package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"j-ticketing/internal/core/services"
	"j-ticketing/internal/db/repositories"
)

// RegisterUserRoutes registers user routes
func RegisterUserRoutes(router fiber.Router, userService *services.UserService) {
	userRouter := router.Group("/users")

	userRouter.Get("/", getAllUsers(userService))
	userRouter.Get("/:id", getUserByID(userService))
	userRouter.Post("/", createUser(userService))
	userRouter.Put("/:id", updateUser(userService))
	userRouter.Delete("/:id", deleteUser(userService))
}

func getAllUsers(service *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		users, err := service.GetAllUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(users)
	}
}

func getUserByID(service *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid ID",
			})
		}

		user, err := service.GetUserByID(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if user == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.JSON(user)
	}
}

func createUser(service *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := new(repositories.User)

		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := service.CreateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	}
}

func updateUser(service *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid ID",
			})
		}

		user := new(repositories.User)
		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		user.ID = id

		if err := service.UpdateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(user)
	}
}

func deleteUser(service *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid ID",
			})
		}

		if err := service.DeleteUser(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusNoContent).Send(nil)
	}
}
