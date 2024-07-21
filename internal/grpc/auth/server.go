package auth

import (
	"STTAuth/internal/services/auth"
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	ssov1 "github.com/skinkvi/protosSTT/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

type IsAdminRequest struct {
	UserID int64 `validate:"required,min=1,max=1000"`
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=32"`
}

type RegisterRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=32"`
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponce, error) {

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation failed")
	}

	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, auth.ErrTooManyAttempts) {
			return nil, status.Error(codes.ResourceExhausted, "too many failed login attempts")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Parse the token
	claims := &jwt.StandardClaims{}
	tokenParsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})
	if err != nil || !tokenParsed.Valid {
		return nil, status.Error(codes.Unauthenticated, "token expired or invalid")
	}

	return &ssov1.LoginResponce{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponce, error) {
	loginReq := &LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	validate := validator.New()
	err := validate.Struct(loginReq)
	if err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, loginReq.Email, loginReq.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user alreay exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponce{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponce, error) {
	isAdminReq := &IsAdminRequest{
		UserID: req.GetUserId(),
	}

	validare := validator.New()
	err := validare.Struct(isAdminReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Validation falied")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponce{
		IsAdmin: isAdmin,
	}, nil
}
