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
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
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
	appID int,
) (string, error) {
	const op = "auth.Login"

	attrs := []any{
		slog.String("op", op),
		slog.String("email", email),
	}
	resultAttrs := make([]any, len(attrs)+1)
	copy(resultAttrs, attrs)
	a.log.Info("Attempting to login user", attrs...)

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("User not found", attrs...)
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
		a.log.Error("Failed to get user", resultAttrs...)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
		a.log.Info("Invalid password", resultAttrs...)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
			a.log.Info("Invalid app_id", resultAttrs...)

			return "", fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
		a.log.Info("Failed to get app", resultAttrs...)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("User logged in successfully", attrs...)

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("Failed to generate token", prettylogger.Err(err))
		return "", fmt.Errorf("%s: %w", op, a.tokenTTL)
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

	attrs := []any{
		slog.String("op", op),
		slog.String("email", email),
	}
	resultAttrs := make([]any, len(attrs)+1)
	copy(resultAttrs, attrs)
	a.log.Info("Registering user", attrs...)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
		a.log.Error("Failed to generate password hash", resultAttrs...)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userSaver.SaveUser(ctx, email, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("User already exists", attrs...)
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		resultAttrs[len(resultAttrs)-1] = prettylogger.Err(err)
		a.log.Error("Failed to save user", resultAttrs...)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	resultAttrs[len(resultAttrs)-1] = slog.Int64("user_id", userID)
	a.log.Info("User registered", resultAttrs...)

	return userID, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	attrs := []any{
		slog.String("op", op),
		slog.Int64("user_id", userID),
	}

	a.log.Info("Checking if user is admin", attrs...)
	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		a.log.Error("Failed to check if user is admin", attrs...)
		return false, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("Checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
