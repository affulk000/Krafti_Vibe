package router

import (
	"Krafti_Vibe/internal/handler"
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
	files.Use(r.RequireAuth())

	// Upload routes
	files.Post("/upload", fileHandler.UploadFile)
	files.Get("/:id", fileHandler.GetFile)
	files.Delete("/:id", fileHandler.DeleteFile)
	files.Get("/", fileHandler.ListFiles)
}
