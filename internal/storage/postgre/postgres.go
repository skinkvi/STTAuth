package postgre

import (
	"STTAuth/internal/domain/models"
	"STTAuth/internal/storage"
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewPostgreStorage(log *slog.Logger, url string) (*Storage, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Error("failed to connect to PostgreSQL", slog.String("url", url))
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Error("failed to ping PostgreSQL", slog.String("url", url))
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	const op = "storage.postgre.Close"

	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, pass_hash []byte) (int64, error) {
	const op = "storage.postgre.SaveUser"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var id int64
	var exists bool

	err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&id, &exists)
	if err != nil {
		if err == sql.ErrNoRows {
			// Пользователь не найден, можно вставить нового пользователя
			stmt, err := tx.Prepare("INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id")
			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}

			err = stmt.QueryRowContext(ctx, email, pass_hash).Scan(&id)
			if err != nil {
				if err == sql.ErrNoRows {
					// Пользователь не найден, не удалось вставить нового пользователя
					return 0, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
				}
				return 0, fmt.Errorf("%s: %w", op, err)
			}

		} else {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	} else if exists {
		// Пользователь уже существует, нельзя вставить нового пользователя
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgre.User"

	var user models.User

	err := s.db.QueryRowContext(ctx, "SELECT id, email, pass_hash FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgre.IsAdmin"

	var isAdmin bool

	err := s.db.QueryRowContext(ctx, "SELECT is_admin FROM users WHERE id = $1", userID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.postgre.App"

	var app models.App

	err := s.db.QueryRowContext(ctx, "SELECT * FROM apps WHERE id = $1", appID).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
