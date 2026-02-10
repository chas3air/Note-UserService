package users

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"usersservice/internal/models/domain"
	"usersservice/internal/service"
	"usersservice/internal/storage"

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

var emailReg = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var phoneReg = regexp.MustCompile(`^\+375\((29|33|44|25)\)[0-9]{3}-[0-9]{2}-[0-9]{2}$`)

type Service struct {
	log     *zap.Logger
	storage Storage
}

func New(log *zap.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s *Service) GetUsers(ctx context.Context, offset int, limit int) ([]domain.User, error) {
	const op = "service.users.GetUsers"
	log := s.log.With(zap.String("op", op))

	if offset < 0 {
		log.Warn("negative offset, resetting to 0", zap.Int("offset", offset))
		return nil, fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}
	if limit <= 0 {
		log.Warn("non-positive limit, resetting to 10", zap.Int("limit", limit))
		return nil, fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}

	users, err := s.storage.GetUsers(ctx, offset, limit)
	if err != nil {
		log.Error("failed to get users", zap.Error(err))
		return nil, err
	}

	return users, nil
}

func (s *Service) GetUserById(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const op = "service.users.GetUserById"
	log := s.log.With(zap.String("op", op), zap.String("id", id.String()))

	user, err := s.storage.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found")
			return domain.User{}, fmt.Errorf("%s: %w", op, service.ErrNotFound)
		}

		log.Error("failed to get user by id", zap.Error(err))
		return domain.User{}, err
	}

	return user, nil
}

func (s *Service) Insert(ctx context.Context, user domain.User) error {
	const op = "service.users.Insert"
	log := s.log.With(zap.String("op", op), zap.String("email", user.Email))

	if !checkRegexpValid(emailReg, user.Email) {
		log.Warn("invalid email format", zap.String("email", user.Email))
		return fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}

	if !checkRegexpValid(phoneReg, user.Phone) {
		log.Warn("invalid phone format", zap.String("phone", user.Phone))
		return fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}

	err := s.storage.Insert(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Warn("user already exists", zap.String("email", user.Email))
			return fmt.Errorf("%s: %w", op, service.ErrAlreadyExists)
		}

		log.Error("failed to insert user", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, user domain.User) error {
	const op = "service.users.Update"
	log := s.log.With(zap.String("op", op), zap.String("id", id.String()))

	if !checkRegexpValid(emailReg, user.Email) {
		log.Warn("invalid email format", zap.String("email", user.Email))
		return fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}

	if !checkRegexpValid(phoneReg, user.Phone) {
		log.Warn("invalid phone format", zap.String("phone", user.Phone))
		return fmt.Errorf("%s: %w", op, service.ErrInvalidArgument)
	}

	err := s.storage.Update(ctx, id, user)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found for update", zap.String("user_id", id.String()))
			return fmt.Errorf("%s: %w", op, service.ErrNotFound)
		}
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Warn("email already exists for another user", zap.String("email", user.Email))
			return fmt.Errorf("%s: %w", op, service.ErrAlreadyExists)
		}

		log.Error("failed to update user", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	const op = "service.users.ChangePassword"
	log := s.log.With(zap.String("op", op), zap.String("id", id.String()))

	err := s.storage.ChangePassword(ctx, id, newPassword)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found for password change", zap.String("user_id", id.String()))
			return fmt.Errorf("%s: %w", op, service.ErrNotFound)
		}

		log.Error("failed to change password", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "service.users.Delete"
	log := s.log.With(zap.String("op", op), zap.String("id", id.String()))

	err := s.storage.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found for deletion", zap.String("user_id", id.String()))
			return fmt.Errorf("%s: %w", op, service.ErrNotFound)
		}

		log.Error("failed to delete user", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func checkRegexpValid(reg *regexp.Regexp, str string) bool {
	return reg.MatchString(str)
}
