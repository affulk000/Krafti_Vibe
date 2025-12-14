package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupFileUploadRoutes sets up file upload routes
func (r *Router) setupFileUploadRoutes(api fiber.Router) {
	// Initialize file upload service
	fileService := service.NewFileUploadService(r.repos, r.config.Logger)

	// Initialize file upload handler
	fileHandler := handler.NewFileUploadHandler(fileService)

	// File upload routes group
	files := api.Group("/files")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// Public routes (with authentication)
	files.Post("", authMiddleware, middleware.RequireScopes(r.scopes.FileWrite), fileHandler.UploadFile)
	files.Get("", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.ListFiles)
	files.Get("/search", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.SearchFiles)
	files.Get("/stats", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.GetFileStats)
	files.Get("/storage/usage", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.GetStorageUsage)

	// File-specific routes
	files.Get("/:id", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.GetFile)
	files.Put("/:id", authMiddleware, middleware.RequireScopes(r.scopes.FileWrite), fileHandler.UpdateFile)
	files.Delete("/:id", authMiddleware, middleware.RequireScopes(r.scopes.FileDelete), fileHandler.DeleteFile)
	files.Post("/:id/access", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.UpdateFileAccess)

	// Query routes
	files.Get("/uploader/:uploader_id", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.ListFilesByUploader)
	files.Get("/entity/:entity_type/:entity_id", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.ListFilesByEntity)
	files.Get("/type/:file_type", authMiddleware, middleware.RequireScopes(r.scopes.FileRead), fileHandler.ListFilesByType)

	// Admin/Management routes
	files.Post("/cleanup", authMiddleware, middleware.RequireScopes(r.scopes.FileManage), fileHandler.CleanupOrphanedFiles)
}
