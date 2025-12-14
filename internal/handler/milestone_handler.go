package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MilestoneHandler handles HTTP requests for project milestone operations
type MilestoneHandler struct {
	milestoneService service.MilestoneService
}

// NewMilestoneHandler creates a new milestone handler
func NewMilestoneHandler(milestoneService service.MilestoneService) *MilestoneHandler {
	return &MilestoneHandler{
		milestoneService: milestoneService,
	}
}

// CreateMilestone creates a new milestone
func (h *MilestoneHandler) CreateMilestone(c *fiber.Ctx) error {
	var req dto.CreateMilestoneRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	milestone, err := h.milestoneService.CreateMilestone(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, milestone, "Milestone created successfully")
}

// GetMilestone retrieves a milestone by ID
func (h *MilestoneHandler) GetMilestone(c *fiber.Ctx) error {
	milestoneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid milestone ID", err)
	}

	milestone, err := h.milestoneService.GetMilestone(c.Context(), milestoneID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, milestone)
}

// UpdateMilestone updates a milestone
func (h *MilestoneHandler) UpdateMilestone(c *fiber.Ctx) error {
	milestoneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid milestone ID", err)
	}

	var req dto.UpdateMilestoneRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	milestone, err := h.milestoneService.UpdateMilestone(c.Context(), milestoneID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, milestone, "Milestone updated successfully")
}

// DeleteMilestone deletes a milestone
func (h *MilestoneHandler) DeleteMilestone(c *fiber.Ctx) error {
	milestoneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid milestone ID", err)
	}

	if err := h.milestoneService.DeleteMilestone(c.Context(), milestoneID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// GetProjectMilestones gets all milestones for a project
func (h *MilestoneHandler) GetProjectMilestones(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	milestones, err := h.milestoneService.ListMilestonesByProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, milestones)
}

// CompleteMilestone marks a milestone as completed
func (h *MilestoneHandler) CompleteMilestone(c *fiber.Ctx) error {
	milestoneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid milestone ID", err)
	}

	if err := h.milestoneService.CompleteMilestone(c.Context(), milestoneID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Milestone completed successfully")
}
