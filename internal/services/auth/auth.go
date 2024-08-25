package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"

	"github.com/jacute/prettylogger"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passwordHash []byte,
	) (userID int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

var (
	ErrUserNotFound       = errors.New("User not found")
	ErrUserExists         = errors.New("User already exists")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrInvalidAppID       = errors.New("Invalid app ID")
)

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login checks if the user with given credentials exists
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int32,
) (string, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("Attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found")
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("Failed to get user", prettylogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Info("Invalid password", prettylogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Info("Invalid app_id", prettylogger.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		log.Info("Failed to get app", prettylogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("Failed to generate token", prettylogger.Err(err))
		return "", fmt.Errorf("%s: %s", op, a.tokenTTL)
	}

	return token, nil
}

// Register registers new user and returns user ID.
func (a *Auth) Register(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.Register"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("Registering user")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate password hash", prettylogger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userSaver.SaveUser(ctx, email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("User already exists", prettylogger.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("Failed to save user", prettylogger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("User registered", slog.Int64("user_id", userID))

	return userID, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("Checking if user is admin")
	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found")
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		log.Error("Failed to check if user is admin")
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
