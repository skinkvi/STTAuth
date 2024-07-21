package auth

import (
	"STTAuth/internal/domain/models"
	jwtT "STTAuth/internal/lib/jwt"
	"STTAuth/internal/lib/logger/sl"
	"STTAuth/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvader UserProvider
	appProvader AppProvider
	tokenTTL    time.Duration
}

// Тут мог быть просто один большой интерфейс Storage и так возможно в данном примере могло быть лучше но, я хочу делать все +- на перед и вдруг у меня будет такое что мне нужно будет работать и прикручивать отдельный сервис который будет заниматься юзерпровайдером там та же kafka или может быть что то с кешем связанное. А UserSaver в этом не хочет участвовать и он там будет лишним грузом
// Но у этого подхода тоже есть минус. За место того что бы передать один Storage мы передаем много всего
type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrTooManyAttempts    = errors.New("too many login attempts")
)

// New это конструктор для Auth сервиса
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvader: userProvider,
		log:         log,
		appProvader: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvader.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("falied to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvader.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwtT.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("falied to generate token")

		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		// Если делать такую систему для авторизации то логировать email нельзя ни в коем случае потому что если вся информацию разбредеться по логам а нужно будет удалить кого полностью, искать все это будет очень муторно и не делай так в продакшине))
		slog.String("username", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("falied to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists")

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("falied to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userId),
	)

	isAdmin, err := a.usrProvader.IsAdmin(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("#{op}: #{err}")
	}

	log.Info("checked user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) User(ctx context.Context, email string) (models.User, error) {
	const op = "services.auth.User"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.usrProvader.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))
			return models.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("falied to get user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	log.Info("user found")
	return user, nil
}
