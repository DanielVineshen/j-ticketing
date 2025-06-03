// File: j-ticketing/internal/core/handlers/tag_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/tag"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"

	"github.com/gofiber/fiber/v2"
)

type TagHandler struct {
	tagService *service.TagService
}

func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// GetAllTags handles GET /api/tags
func (h *TagHandler) GetAllTags(c *fiber.Ctx) error {
	tags, err := h.tagService.GetAllTags()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to retrieve tags", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(tags))
}

// CreateTag handles POST /api/tags
func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	var req dto.CreateNewTagRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Create tag
	_, err := h.tagService.CreateTag(&req)
	if err != nil {
		if err.Error() == "tag name already exists" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"Tag name already exits", nil,
			))
		}
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to create tag", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// UpdateTag handles PUT /api/tags
func (h *TagHandler) UpdateTag(c *fiber.Ctx) error {
	var req dto.UpdateTagRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Update tag
	_, err := h.tagService.UpdateTag(&req)
	if err != nil {
		if err.Error() == "tag not found" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"Tag not found", nil,
			))
		}
		if err.Error() == "tag name already exists" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"Tag name already exists", nil,
			))
		}
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to update tag", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// DeleteTag handles DELETE /api/tags
func (h *TagHandler) DeleteTag(c *fiber.Ctx) error {
	var req dto.DeleteTagRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Delete tag
	if err := h.tagService.DeleteTag(&req); err != nil {
		if err.Error() == "tag not found" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"Tag not found", nil,
			))
		}
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to update tag", nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
