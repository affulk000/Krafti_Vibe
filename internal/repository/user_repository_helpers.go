package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"context"
)

// Helper methods for user sync service

// FindByZitadelID finds a user by Zitadel ID (alias for GetByZitadelID)
func (r *userRepository) FindByZitadelID(ctx context.Context, zitadelID string) (*models.User, error) {
	return r.GetByZitadelID(ctx, zitadelID)
}

// FindByEmail finds a user by email (alias for GetByEmail)
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.GetByEmail(ctx, email)
}
