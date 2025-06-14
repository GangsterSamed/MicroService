package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// HandleGRPCError преобразует gRPC ошибку в HTTP статус код и сообщение
func HandleGRPCError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			return http.StatusBadRequest, st.Message()
		case codes.Unauthenticated:
			return http.StatusUnauthorized, st.Message()
		case codes.PermissionDenied:
			return http.StatusForbidden, st.Message()
		case codes.NotFound:
			return http.StatusNotFound, st.Message()
		case codes.AlreadyExists:
			return http.StatusConflict, st.Message()
		case codes.Unavailable:
			return http.StatusServiceUnavailable, st.Message()
		default:
			return http.StatusInternalServerError, st.Message()
		}
	}
	return http.StatusInternalServerError, err.Error()
}

// MarshalResponse marshals gRPC response and handles errors
func MarshalResponse(resp proto.Message, err error) ([]byte, int, error) {
	if err != nil {
		statusCode, _ := HandleGRPCError(err)
		return nil, statusCode, err
	}
	data, err := protojson.Marshal(resp)
	return data, http.StatusOK, err
}
