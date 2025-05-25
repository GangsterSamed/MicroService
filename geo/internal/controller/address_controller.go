package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/client"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/service"
)

type AddressControllerInterface interface {
	AddressSearchHandler(ctx *gin.Context)
	AddressGeocodeHandler(ctx *gin.Context)
}

// AddressController отвечает за обработку запросов, связанных с адресами
type AddressController struct {
	geoClient  client.GeoClient
	authClient *service.AuthClient
}

// NewAddressController создаёт новый экземпляр контроллера
func NewAddressController(geoClient client.GeoClient, authClient *service.AuthClient) *AddressController {
	return &AddressController{
		geoClient:  geoClient,
		authClient: authClient,
	}
}

type contextKey string

const (
	authTokenKey contextKey = "auth_token"
)

// AddressSearchHandler обрабатывает запрос на поиск адресов
//
// @Summary Поиск адреса
// @Description Возвращает список адресов по строке запроса, используя внешний геосервис
// @Tags Геоданные
// @Accept json
// @Produce json
// @Param input body models.RequestAddressSearch true "Параметры поиска"
// @Example {json} Пример запроса:
//
//	{
//	    "query": "Москва, Ленина 1"
//	}
//
// @Success 200 {object} models.ResponseAddress "Успешный ответ"
// @Failure 400 {object} models.ErrorResponse "Невалидный запрос"
// @Failure 403 {object} models.ErrorResponse "Вход закрыт"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /api/address/search [post]
func (ac *AddressController) AddressSearchHandler(ctx *gin.Context) {
	// 1. Проверяем токен
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Token required"})
		return
	}

	valid, err := ac.authClient.ValidateToken(ctx.Request.Context(), token)
	if err != nil || !valid {
		ctx.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Invalid token"})
		return
	}

	// 2. Парсим тело запроса
	var req models.RequestAddressSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query cannot be empty"})
		return
	}

	// 3. Создаём контекст с токеном
	c := context.WithValue(ctx.Request.Context(), authTokenKey, token)

	// 4. Вызываем бизнес-логику
	addresses, err := ac.geoClient.AddressSearch(c, req.Query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Service unavailable"})
		return
	}

	ctx.JSON(http.StatusOK, models.ResponseAddress{Addresses: addresses})
}

// AddressGeocodeHandler выполняет поиск адреса по геоданным
//
// @Summary Обратное геокодирование
// @Description Возвращает адреса по географическим координатам (широта/долгота)
// @Tags Геоданные
// @Accept json
// @Produce json
// @Param input body models.RequestAddressGeocode true "Координаты для поиска"
// @Example {json} Пример запроса:
//
//	{
//	    "lat": "55.7558",
//	    "lng": "37.6173"
//	}
//
// @Success 200 {object} models.ResponseAddress "Успешный ответ"
// @Failure 400 {object} models.ErrorResponse "Невалидные координаты"
// @Failure 403 {object} models.ErrorResponse "Вход закрыт"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /api/address/geocode [post]
func (ac *AddressController) AddressGeocodeHandler(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Token required"})
		return
	}

	valid, err := ac.authClient.ValidateToken(ctx.Request.Context(), token)
	if err != nil || !valid {
		ctx.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Invalid token"})
		return
	}

	var req models.RequestAddressGeocode
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Lat == "" || req.Lng == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng are required"})
		return
	}

	c := context.WithValue(ctx.Request.Context(), authTokenKey, token)

	// 4. Вызываем бизнес-логику
	addresses, err := ac.geoClient.GeoCode(c, req.Lat, req.Lng)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Service unavailable"})
		return
	}

	ctx.JSON(http.StatusOK, models.ResponseAddress{Addresses: addresses})
}
