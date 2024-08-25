package authgrpc

import (
	"context"
	"errors"
	"sso/internal/lib/validators"
	"sso/internal/services/auth"

	ssov1 "github.com/jacute/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int32,
	) (token string, err error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	email := req.GetEmail()
	password := req.GetPassword()
	appID := req.GetAppId()

	validator := validators.ToLoginValidator(email, password, appID)
	if err := validator.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, validators.GetDetailedError(err))
	}

	token, err := s.auth.Login(ctx, email, password, appID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "Invalid credentials")
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	email := req.GetEmail()
	password := req.GetPassword()

	validator := validators.ToRegisterValidator(email, password)
	if err := validator.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, validators.GetDetailedError(err))
	}

	userID, err := s.auth.Register(ctx, email, password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	userID := req.GetUserId()

	validator := validators.ToIsAdminValidator(userID)
	if err := validator.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, validators.GetDetailedError(err))
	}

	isAdmin, err := s.auth.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidAppID) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}
