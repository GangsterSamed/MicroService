package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type userServer struct {
	pb.UnimplementedUserServiceServer
	service UserService
}

func NewUserServer(service UserService) pb.UserServiceServer {
	return &userServer{
		service: service,
	}
}

func (s *userServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	// Создание пользователя не требует авторизации
	return s.service.CreateUser(ctx, req.Email, req.PasswordHash)
}

func (s *userServer) GetUserByEmail(ctx context.Context, req *pb.EmailRequest) (*pb.UserResponse, error) {
	// Метод используется для аутентификации, поэтому не требует токен
	return s.service.GetUserByEmail(ctx, req.Email)
}

func (s *userServer) GetUserProfile(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfileResponse, error) {
	// Получаем user_id из метаданных
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Используем ID из запроса, если он не указан - используем ID из токена
	targetUserID := req.UserId
	if targetUserID == "" {
		targetUserID = userID
	}

	// Проверяем, что пользователь запрашивает свой профиль
	if targetUserID != userID {
		return nil, status.Error(codes.PermissionDenied, "can't get other users profiles")
	}

	return s.service.GetUserProfile(ctx, targetUserID)
}

func (s *userServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	if _, err := s.getUserIDFromContext(ctx); err != nil {
		return nil, err
	}

	return s.service.ListUsers(ctx, req.Limit, req.Offset)
}

func (s *userServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserProfileResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь обновляет свой профиль
	if req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "can't update other users profiles")
	}

	return s.service.UpdateUser(ctx, req.UserId, req.Email, req.Password)
}

func (s *userServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := s.getUserIDFromContext(ctx)
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

func (s *userServer) getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "user_id is not provided")
	}

	return userIDs[0], nil
}
