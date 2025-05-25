package controller

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strconv"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"

	"github.com/gin-gonic/gin"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/service"
)

type UserController interface {
	GetUserProfileHandler(ctx *gin.Context)
	ListUsersHandler(ctx *gin.Context)
	AuthMiddleware() gin.HandlerFunc
}

type userController struct {
	userService service.UserService
	authClient  proto.AuthServiceClient
}

func NewUserController(userService service.UserService) (UserController, error) {
	conn, err := grpc.Dial("auth:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("auth service connection failed: %w", err)
	}

	return &userController{
		userService: userService,
		authClient:  proto.NewAuthServiceClient(conn),
	}, nil
}

// @Summary Получить профиль пользователя
// @Description Возвращает профиль пользователя по ID
// @Tags User
// @Security ApiKeyAuth
// @Param user_id query string true "ID пользователя"
// @Success 200 {object} models.UserProfileResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/user/profile [get]
func (c *userController) GetUserProfileHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "user_id not found in context"})
		return
	}

	response, err := c.userService.GetUserProfile(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Список пользователей
// @Description Возвращает список пользователей с пагинацией
// @Tags User
// @Security ApiKeyAuth
// @Param limit query int false "Лимит" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} models.ListUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/user/list [get]
func (c *userController) ListUsersHandler(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid limit value"})
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid offset value"})
		return
	}

	response, err := c.userService.ListUsers(ctx.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// AuthMiddleware для проверки JWT через auth-сервис
func (c *userController) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{Error: "token required"})
			return
		}

		// Проверяем токен через auth-сервис
		conn, err := grpc.Dial("auth:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "auth service unavailable"})
			return
		}
		defer conn.Close()

		authClient := proto.NewAuthServiceClient(conn)
		resp, err := authClient.ValidateToken(ctx.Request.Context(), &proto.TokenRequest{Token: token})
		if err != nil || !resp.Valid {
			ctx.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{Error: "invalid token"})
			return
		}

		ctx.Set("user_id", resp.UserId)
		ctx.Next()
	}
}
