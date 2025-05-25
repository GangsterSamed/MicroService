package service

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
	"time"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	service  UserService
	authConn *grpc.ClientConn // Добавляем кеширование соединения
}

func NewUserServer(service UserService) (*UserServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "auth:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &UserServer{
		service:  service,
		authConn: conn,
	}, nil
}

func (s *UserServer) Close() error {
	if s.authConn != nil {
		return s.authConn.Close()
	}
	return nil
}

func (s *UserServer) validateToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	authClient := proto.NewAuthServiceClient(s.authConn)
	resp, err := authClient.ValidateToken(ctx, &proto.TokenRequest{Token: tokens[0]})
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}
	if !resp.Valid {
		return "", status.Error(codes.PermissionDenied, "invalid token")
	}

	return resp.UserId, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	// Создание пользователя не требует авторизации
	return s.service.CreateUser(ctx, req.Email, req.PasswordHash)
}

func (s *UserServer) GetUserByEmail(ctx context.Context, req *pb.EmailRequest) (*pb.UserResponse, error) {
	// Только для внутреннего использования (auth-сервисом)
	if _, err := s.validateToken(ctx); err != nil {
		return nil, err
	}
	return s.service.GetUserByEmail(ctx, req.Email)
}

func (s *UserServer) GetUserProfile(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfileResponse, error) {
	// Проверяем токен через auth-сервис
	userID, err := s.validateToken(ctx)
	if err != nil {
		return nil, err
	}

	// Если user_id не указан, используем из токена
	targetUserID := req.UserId
	if targetUserID == "" {
		targetUserID = userID
	}

	// Проверяем доступ (либо свой профиль, либо админские права)
	if targetUserID != userID {
		return nil, status.Error(codes.PermissionDenied, "can't access other users profiles")
	}

	return s.service.GetUserProfile(ctx, targetUserID)
}

func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	if _, err := s.validateToken(ctx); err != nil {
		return nil, err
	}

	return s.service.ListUsers(ctx, req.Limit, req.Offset)
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserProfileResponse, error) {
	userID, err := s.validateToken(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь обновляет свой профиль
	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "can't update other users profiles")
	}

	return s.service.UpdateUser(ctx, req.UserId, req.Email, req.Password)
}

func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := s.validateToken(ctx)
	if err != nil {
		return nil, err
	}

	// Проверка прав (либо админ, либо удаляет себя)
	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "can't delete other users")
	}

	if err := s.service.DeleteUser(ctx, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}
