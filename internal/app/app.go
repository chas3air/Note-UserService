package app

import (
	"context"
	restapp "usersservice/internal/app/rest"
	"usersservice/internal/models/domain"
	"usersservice/internal/service/users"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Storage interface {
	GetUsers(ctx context.Context, offset int, limit int) ([]domain.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (domain.User, error)
	Insert(ctx context.Context, user domain.User) error
	Update(ctx context.Context, id uuid.UUID, user domain.User) error
	ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type App struct {
	RESTServer *restapp.App
}

func New(log *zap.Logger, storage Storage, port int) *App {
	service := users.New(log, storage)

	restServer := restapp.New(log, service, port)

	return &App{
		RESTServer: restServer,
	}
}
