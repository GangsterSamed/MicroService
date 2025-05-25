package controller

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/service"
	"time"
)

type AuthController interface {
	RegisterHandler(ctx *gin.Context)
	LoginHandler(ctx *gin.Context)
}

type authController struct {
	authService service.AuthService
	logger      *slog.Logger
}

func NewAuthController(authService service.AuthService, logger *slog.Logger) AuthController {
	return &authController{
		authService: authService,
		logger:      logger.With("component", "AuthController"),
	}
}

// RegisterHandler - регистрация пользователя
// @Summary Регистрация
// @Description Создает нового пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body models.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/auth/register [post]
func (c *authController) RegisterHandler(ctx *gin.Context) {
	logger := c.logger.With(
		"handler", "RegisterHandler",
		"client_ip", ctx.ClientIP(),
	)

	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	logger = logger.With("email", req.Email)
	logger.Info("Registration request received")

	response, err := c.authService.RegisterUser(ctx, req.Email, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user already exists" {
			status = http.StatusConflict
			logger.Warn("Registration failed - user already exists")
		} else {
			logger.Error("Registration failed", "error", err)
		}

		ctx.JSON(status, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("Registration successful", "user_id", response.UserId)
	ctx.JSON(http.StatusCreated, models.AuthResponse{
		Token:     response.Token,
		ExpiresAt: time.Unix(response.ExpiresAt, 0),
		UserID:    response.UserId,
	})
}

// LoginHandler - аутентификация пользователя
// @Summary Вход
// @Description Аутентифицирует пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/auth/login [post]
func (c *authController) LoginHandler(ctx *gin.Context) {
	logger := c.logger.With(
		"handler", "LoginHandler",
		"client_ip", ctx.ClientIP(),
	)

	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	logger = logger.With("email", req.Email)
	logger.Info("Login attempt")

	response, err := c.authService.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Login failed - invalid credentials")
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid credentials",
		})
		return
	}

	logger.Info("Login successful", "user_id", response.UserId)
	ctx.JSON(http.StatusOK, models.AuthResponse{
		Token:     response.Token,
		ExpiresAt: time.Unix(response.ExpiresAt, 0),
		UserID:    response.UserId,
	})
}
