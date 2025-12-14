package handler

import (
	"strconv"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// FileUploadHandler handles file upload HTTP requests
type FileUploadHandler struct {
	fileService service.FileUploadService
}

// NewFileUploadHandler creates a new FileUploadHandler
func NewFileUploadHandler(fileService service.FileUploadService) *FileUploadHandler {
	return &FileUploadHandler{
		fileService: fileService,
	}
}

// UploadFile handles file upload
// @Summary Upload a file
// @Description Upload a new file with metadata
// @Tags Files
// @Accept json
// @Produce json
// @Param request body dto.UploadFileRequest true "File upload request"
// @Success 201 {object} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /files [post]
func (h *FileUploadHandler) UploadFile(c *fiber.Ctx) error {
	// Get authentication context
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	tenantID := authCtx.TenantID
	userID := authCtx.UserID

	// Parse request body
	var req dto.UploadFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	// Upload file
	file, err := h.fileService.UploadFile(c.Context(), tenantID, userID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(file)
}

// GetFile retrieves a file by ID
// @Summary Get file by ID
// @Description Retrieve file details by ID
// @Tags Files
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /files/{id} [get]
func (h *FileUploadHandler) GetFile(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file ID",
			"code":  "INVALID_FILE_ID",
		})
	}

	file, err := h.fileService.GetFile(c.Context(), fileID, tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(file)
}

// UpdateFile updates a file's metadata
// @Summary Update file metadata
// @Description Update file metadata by ID
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Param request body dto.UpdateFileRequest true "Update request"
// @Success 200 {object} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /files/{id} [put]
func (h *FileUploadHandler) UpdateFile(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file ID",
			"code":  "INVALID_FILE_ID",
		})
	}

	var req dto.UpdateFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	file, err := h.fileService.UpdateFile(c.Context(), fileID, tenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(file)
}

// DeleteFile deletes a file
// @Summary Delete file
// @Description Delete a file by ID
// @Tags Files
// @Param id path string true "File ID"
// @Success 204
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /files/{id} [delete]
func (h *FileUploadHandler) DeleteFile(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file ID",
			"code":  "INVALID_FILE_ID",
		})
	}

	if err := h.fileService.DeleteFile(c.Context(), fileID, tenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListFiles lists files with filtering and pagination
// @Summary List files
// @Description List files with optional filters
// @Tags Files
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param file_type query string false "File type filter"
// @Param uploaded_by_id query string false "Uploader ID filter"
// @Param entity_type query string false "Related entity type"
// @Param entity_id query string false "Related entity ID"
// @Param search query string false "Search query"
// @Success 200 {object} dto.FileUploadListResponse
// @Failure 400 {object} fiber.Map
// @Router /files [get]
func (h *FileUploadHandler) ListFiles(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	filter := &dto.FileUploadFilter{
		TenantID:    tenantID,
		Page:        page,
		PageSize:    pageSize,
		SearchQuery: c.Query("search"),
	}

	// Optional filters
	if fileType := c.Query("file_type"); fileType != "" {
		ft := models.FileType(fileType)
		filter.FileType = &ft
	}

	if uploaderID := c.Query("uploaded_by_id"); uploaderID != "" {
		if id, err := uuid.Parse(uploaderID); err == nil {
			filter.UploadedByID = &id
		}
	}

	if entityType := c.Query("entity_type"); entityType != "" {
		filter.RelatedEntityType = &entityType
	}

	if entityID := c.Query("entity_id"); entityID != "" {
		if id, err := uuid.Parse(entityID); err == nil {
			filter.RelatedEntityID = &id
		}
	}

	if minSize := c.Query("min_size"); minSize != "" {
		if size, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			filter.MinSize = &size
		}
	}

	if maxSize := c.Query("max_size"); maxSize != "" {
		if size, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			filter.MaxSize = &size
		}
	}

	files, err := h.fileService.ListFiles(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(files)
}

// ListFilesByUploader lists files uploaded by a specific user
// @Summary List files by uploader
// @Description Get all files uploaded by a specific user
// @Tags Files
// @Produce json
// @Param uploader_id path string true "Uploader ID"
// @Success 200 {array} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Router /files/uploader/{uploader_id} [get]
func (h *FileUploadHandler) ListFilesByUploader(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	uploaderID, err := uuid.Parse(c.Params("uploader_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid uploader ID",
			"code":  "INVALID_UPLOADER_ID",
		})
	}

	files, err := h.fileService.ListFilesByUploader(c.Context(), uploaderID, tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(files)
}

// ListFilesByEntity lists files related to a specific entity
// @Summary List files by entity
// @Description Get all files related to a specific entity
// @Tags Files
// @Produce json
// @Param entity_type path string true "Entity type"
// @Param entity_id path string true "Entity ID"
// @Success 200 {array} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Router /files/entity/{entity_type}/{entity_id} [get]
func (h *FileUploadHandler) ListFilesByEntity(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	entityType := c.Params("entity_type")
	entityID, err := uuid.Parse(c.Params("entity_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid entity ID",
			"code":  "INVALID_ENTITY_ID",
		})
	}

	files, err := h.fileService.ListFilesByEntity(c.Context(), entityType, entityID, tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(files)
}

// ListFilesByType lists files by type
// @Summary List files by type
// @Description Get all files of a specific type
// @Tags Files
// @Produce json
// @Param file_type path string true "File type" Enums(image, document, video, other)
// @Success 200 {array} dto.FileUploadResponse
// @Failure 400 {object} fiber.Map
// @Router /files/type/{file_type} [get]
func (h *FileUploadHandler) ListFilesByType(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	fileType := models.FileType(c.Params("file_type"))

	files, err := h.fileService.ListFilesByType(c.Context(), tenantID, fileType)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(files)
}

// SearchFiles searches files by query
// @Summary Search files
// @Description Search files by filename
// @Tags Files
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.FileUploadListResponse
// @Failure 400 {object} fiber.Map
// @Router /files/search [get]
func (h *FileUploadHandler) SearchFiles(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search query is required",
			"code":  "MISSING_QUERY",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	files, err := h.fileService.SearchFiles(c.Context(), tenantID, query, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(files)
}

// GetFileStats retrieves file statistics
// @Summary Get file statistics
// @Description Get file upload statistics for the tenant
// @Tags Files
// @Produce json
// @Success 200 {object} dto.FileStatsResponse
// @Failure 400 {object} fiber.Map
// @Router /files/stats [get]
func (h *FileUploadHandler) GetFileStats(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	stats, err := h.fileService.GetFileStats(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(stats)
}

// GetStorageUsage retrieves storage usage
// @Summary Get storage usage
// @Description Get total storage used by the tenant
// @Tags Files
// @Produce json
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /files/storage/usage [get]
func (h *FileUploadHandler) GetStorageUsage(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}
	tenantID := authCtx.TenantID

	usage, err := h.fileService.GetStorageUsage(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"tenant_id":          tenantID,
		"storage_used_bytes": usage,
		"storage_used_mb":    float64(usage) / (1024 * 1024),
		"storage_used_gb":    float64(usage) / (1024 * 1024 * 1024),
	})
}

// CleanupOrphanedFiles cleans up orphaned files
// @Summary Cleanup orphaned files
// @Description Remove files not associated with any entity
// @Tags Files
// @Produce json
// @Param days query int false "Days old" default(30)
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /files/cleanup [post]
func (h *FileUploadHandler) CleanupOrphanedFiles(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))

	if err := h.fileService.CleanupOrphanedFiles(c.Context(), days); err != nil {
		if errors.IsValidationError(err) {
			var appErr *errors.AppError
			if e, ok := err.(*errors.AppError); ok {
				appErr = e
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
		}
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "orphaned files cleaned up successfully",
		"days":    days,
	})
}

// UpdateFileAccess updates file access timestamp
// @Summary Update file access
// @Description Update last accessed timestamp for a file
// @Tags Files
// @Param id path string true "File ID"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /files/{id}/access [post]
func (h *FileUploadHandler) UpdateFileAccess(c *fiber.Ctx) error {
	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file ID",
			"code":  "INVALID_FILE_ID",
		})
	}

	if err := h.fileService.UpdateFileAccess(c.Context(), fileID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "file access updated",
	})
}
