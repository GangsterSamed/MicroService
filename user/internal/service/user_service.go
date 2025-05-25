package service

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/repository"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type UserService interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*pb.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*pb.UserResponse, error)
	GetUserProfile(ctx context.Context, userID string) (*pb.UserProfileResponse, error)
	ListUsers(ctx context.Context, limit, offset int32) (*pb.ListUsersResponse, error)
	UpdateUser(ctx context.Context, userID, email, password string) (*pb.UserProfileResponse, error)
	DeleteUser(ctx context.Context, userID string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, email, passwordHash string) (*pb.UserResponse, error) {
	user, err := s.repo.CreateUser(ctx, email, passwordHash)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		UpdatedAt:    timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*pb.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		UpdatedAt:    timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID string) (*pb.UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &pb.UserProfileResponse{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int32) (*pb.ListUsersResponse, error) {
	users, total, err := s.repo.ListUsers(ctx, int(limit), int(offset))
	if err != nil {
		return nil, err
	}

	response := &pb.ListUsersResponse{
		TotalCount: int32(total),
	}

	for _, user := range users {
		response.Users = append(response.Users, &pb.UserProfileResponse{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return response, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID, email, password string) (*pb.UserProfileResponse, error) {
	// 1. Хеширование пароля
	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	// 2. Обновление в БД
	user, err := s.repo.UpdateUser(ctx, userID, email, passwordHash)
	if err != nil {
		return nil, err
	}

	// 3. Формирование ответа
	return &pb.UserProfileResponse{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteUser(ctx, userID)
}
