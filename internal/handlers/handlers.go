package handlers

import (
	"context"

	"usersservice/internal/models/domain"

	"github.com/google/uuid"
)

type Service interface {
	GetUsers(ctx context.Context, offset int, limit int) ([]domain.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (domain.User, error)
	Insert(ctx context.Context, user domain.User) error
	Update(ctx context.Context, id uuid.UUID, user domain.User) error
	ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
