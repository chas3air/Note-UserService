package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"usersservice/internal/models/domain"
	"usersservice/internal/storage"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Storage struct {
	log *zap.Logger
	db  *sql.DB
}

func New(log *zap.Logger, conn string) (*Storage, func() error, error) {
	const op = "storage.postgres.New"
	tmpLog := log.With(zap.String("op", op))

	db, err := sql.Open("postgres", conn)
	if err != nil {
		tmpLog.Error("failed to open database connection", zap.Error(err))
		return nil, nil, err
	}

	if err := db.Ping(); err != nil {
		tmpLog.Error("failed to ping database", zap.Error(err))
		return nil, nil, err
	}

	wd, _ := os.Getwd()
	migrationPath := filepath.Join(wd, "migrations", "postgres")
	if err := ApplyMigrations(db, migrationPath); err != nil {
		tmpLog.Error("failed to apply migrations", zap.Error(err))
		return nil, nil, err
	}

	return &Storage{
		log: tmpLog,
		db:  db,
	}, db.Close, nil
}

func ApplyMigrations(db *sql.DB, migrationPath string) error {
	const op = "storage.postgres.apply"
	if err := goose.Up(db, migrationPath); err != nil {
		goose.SetLogger(nil)
		return goose.Up(db, migrationPath)
	}

	return nil
}

func (s *Storage) GetUsers(ctx context.Context, offset int, limit int) ([]domain.User, error) {
	const op = "storage.postgres.GetUsers"
	log := s.log.With(zap.String("op", op))

	query := `
		SELECT id, name, email, phone, password FROM users
		OFFSET $1
		LIMIT $2;
	`

	rows, err := s.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		log.Error("failed to execute query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var users = make([]domain.User, 0, limit)
	var tmp domain.User

	for rows.Next() {
		if err := rows.Scan(&tmp.Id, &tmp.Name, &tmp.Email, &tmp.Phone, &tmp.Password); err != nil {
			log.Error("failed to scan row", zap.Error(err))
			return nil, err
		}
		users = append(users, tmp)
	}

	return users, nil
}

func (s *Storage) GetUserById(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const op = "storage.postgres.GetUserById"
	log := s.log.With(zap.String("op", op))

	query := `
		SELECT id, name, email, phone, password FROM users
		WHERE id = $1;
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var user domain.User
	if err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Phone, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to scan row", zap.Error(err))
		return domain.User{}, err
	}

	return user, nil
}

func (s *Storage) Insert(ctx context.Context, user domain.User) error {
	const op = "storage.postgres.Insert"
	log := s.log.With(zap.String("op", op))

	query := `
		INSERT INTO users (id, name, email, phone, password)
		VALUES ($1, $2, $3, $4, $5);
	`

	if _, err := s.db.ExecContext(ctx, query, user.Id, user.Name, user.Email, user.Phone, user.Password); err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				log.Warn("dublicate user", zap.String("email", user.Email))
				return fmt.Errorf("%s: %w", op, storage.ErrAlreadyExists)
			}
		}

		log.Error("failed to execute insert", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	const op = "storage.postgres.ChangePassword"
	log := s.log.With(zap.String("op", op))

	query := `
		UPDATE users
		SET password = $1
		WHERE id = $2;
	`

	res, err := s.db.ExecContext(ctx, query, newPassword, id)
	if err != nil {
		log.Error("failed to execute update", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		log.Warn("user not found for password change", zap.String("user_id", id.String()))
		return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}

	return nil
}

func (s *Storage) Update(ctx context.Context, id uuid.UUID, user domain.User) error {
	const op = "storage.postgres.Update"
	log := s.log.With(zap.String("op", op))

	query := `
		UPDATE users
		SET name = $1, email = $2, phone = $3
		WHERE id = $4;
	`

	res, err := s.db.ExecContext(ctx, query, user.Name, user.Email, user.Phone, id)
	if err != nil {
		log.Error("failed to execute update", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		log.Warn("user not found for update", zap.String("user_id", id.String()))
		return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}

	return nil
}

func (s *Storage) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "storage.postgres.Delete"
	log := s.log.With(zap.String("op", op))

	query := `
		DELETE FROM users
		WHERE id = $1;
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error("failed to execute delete", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		log.Warn("note not found for delete", zap.String("note_id", id.String()))
		return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}

	return nil
}
