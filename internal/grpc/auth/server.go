package auth

import (
	"auth-api/internal/domain/models"
	"auth-api/internal/services/auth"
	"auth-api/internal/storage"
	"context"
	"errors"

	auth_apiv1 "github.com/deeimos/proto-deimos-app/gen/go/auth-api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Auth interface {
	Login(ctx context.Context, email string, password string) (user *models.UserResponse, err error)
	Register(ctx context.Context, name string, email string, password string) (*models.UserResponse, error)
	Refresh(ctx context.Context, refersh string) (*models.Refresh, error)
	GetUser(ctx context.Context, token string) (*models.UserInfo, error)
}

type serverApi struct {
	auth_apiv1.UnimplementedAuthAPIServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	auth_apiv1.RegisterAuthAPIServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) Login(ctx context.Context, req *auth_apiv1.LoginRequest) (*auth_apiv1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	user, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "Неверный логин или пароль")
		}
		return nil, status.Error(codes.Internal, "internal Error")
	}
	return &auth_apiv1.LoginResponse{
		Id:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		Token:        user.Token,
		RefreshToken: user.RefreshToken,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *auth_apiv1.RegisterRequest) (*auth_apiv1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}
	user, err := s.auth.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.InvalidArgument, "email: Адрес электронной почты уже занят")
		}

		return nil, status.Error(codes.Internal, "internal Error")
	}
	return &auth_apiv1.RegisterResponse{
		Id:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		Token:        user.Token,
		RefreshToken: user.RefreshToken,
	}, nil
}

func (s *serverApi) Refresh(ctx context.Context, req *auth_apiv1.RefreshRequest) (*auth_apiv1.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.Unauthenticated, "Отсутствует токен")
	}
	tokens, err := s.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "Неверный или недействительный токен")
		}
		return nil, status.Error(codes.Internal, "internal Error")
	}
	return &auth_apiv1.RefreshResponse{Token: tokens.Token, RefreshToken: tokens.RefreshToken}, nil
}

func (s *serverApi) GetUser(ctx context.Context, req *auth_apiv1.GetUserRequest) (*auth_apiv1.GetUserResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Отсутствует токен")
	}
	user, err := s.auth.GetUser(ctx, req.GetToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "Неверный или недействительный токен")
		}
		return nil, status.Error(codes.Internal, "internal Error")
	}
	return &auth_apiv1.GetUserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

func validateLogin(req *auth_apiv1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email: Введите email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password: Введите пароль")
	}
	return nil
}

func validateRegister(req *auth_apiv1.RegisterRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name: Введите имя")
	}
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email: Введите email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password: Введите пароль")
	}
	return nil
}
