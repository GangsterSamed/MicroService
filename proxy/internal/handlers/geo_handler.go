package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	geoProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/errors"
)

// GeoHandler обрабатывает запросы к geo сервису
type GeoHandler struct {
	GeoClient geoProto.GeoServiceClient
	logger    *slog.Logger
}

// NewGeoHandler создает новый обработчик geo запросов
func NewGeoHandler(geoClient geoProto.GeoServiceClient, logger *slog.Logger) *GeoHandler {
	return &GeoHandler{
		GeoClient: geoClient,
		logger:    logger,
	}
}

// HandleGeoRequest обрабатывает запросы к geo сервису
func (h *GeoHandler) HandleGeoRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	h.logger.Info("Starting handleGeoRequest", "method", method)

	switch method {
	case "/api/address/search":
		h.logger.Info("Searching address")
		var req geoProto.SearchRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := h.GeoClient.AddressSearch(ctx, &req)
		return errors.MarshalResponse(resp, err)

	case "/api/address/geocode":
		h.logger.Info("Geocoding address")
		var req geoProto.GeoRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := h.GeoClient.GeoCode(ctx, &req)
		return errors.MarshalResponse(resp, err)

	default:
		h.logger.Warn("Method not implemented", "method", method)
		return nil, http.StatusNotImplemented, status.Error(codes.Unimplemented, "method not implemented")
	}
}
