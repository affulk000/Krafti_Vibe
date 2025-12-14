package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TaskHandler handles HTTP requests for project task operations
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var req dto.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	task, err := h.taskService.CreateTask(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, task, "Task created successfully")
}

// GetTask retrieves a task by ID
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid task ID", err)
	}

	task, err := h.taskService.GetTask(c.Context(), taskID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, task)
}

// UpdateTask updates a task
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid task ID", err)
	}

	var req dto.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	task, err := h.taskService.UpdateTask(c.Context(), taskID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, task, "Task updated successfully")
}

// DeleteTask deletes a task
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid task ID", err)
	}

	if err := h.taskService.DeleteTask(c.Context(), taskID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// GetMilestoneTasks gets all tasks for a milestone
func (h *TaskHandler) GetMilestoneTasks(c *fiber.Ctx) error {
	milestoneID, err := uuid.Parse(c.Params("milestone_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid milestone ID", err)
	}

	tasks, err := h.taskService.ListTasksByMilestone(c.Context(), milestoneID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, tasks)
}

// CompleteTask marks a task as completed
func (h *TaskHandler) CompleteTask(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid task ID", err)
	}

	if err := h.taskService.CompleteTask(c.Context(), taskID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Task completed successfully")
}
