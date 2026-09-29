package httpclient

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapHTTPError(err error, op string) error {
	if err == nil {
		return nil
	}

	// 1) Annulation / deadline propagés depuis le contexte gRPC
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, op+": canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, op+": deadline")
	}

	// 2) Erreurs HTTP typées
	if he, ok := errors.AsType[*ClientError](err); ok {
		code := httpStatusToGRPC(he.StatusCode)
		return status.Errorf(code, "%s: %s", op, he.Error())
	}

	// 3) Erreur réseau / inconnue → Unavailable (retryable côté client)
	return status.Errorf(codes.Unavailable, "%s: %v", op, err)
}

func httpStatusToGRPC(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusBadRequest: // 400
		return codes.InvalidArgument
	case http.StatusUnauthorized: // 401
		return codes.Unauthenticated
	case http.StatusForbidden: // 403
		return codes.PermissionDenied
	case http.StatusNotFound: // 404
		return codes.NotFound
	case http.StatusMethodNotAllowed: // 405
		return codes.Unimplemented
	case http.StatusConflict: // 409
		return codes.AlreadyExists
	case http.StatusGone: // 410
		return codes.NotFound
	case http.StatusPreconditionFailed: // 412
		return codes.FailedPrecondition
	case http.StatusRequestEntityTooLarge: // 413
		return codes.ResourceExhausted
	case http.StatusUnsupportedMediaType: // 415
		return codes.InvalidArgument
	case http.StatusUnprocessableEntity: // 422
		return codes.InvalidArgument
	case http.StatusTooManyRequests: // 429
		return codes.ResourceExhausted
	case http.StatusNotImplemented: // 501
		return codes.Unimplemented
	case http.StatusBadGateway, // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return codes.Unavailable
	case http.StatusInternalServerError: // 500
		return codes.Internal
	}
	if httpStatus >= 500 {
		return codes.Internal
	}
	if httpStatus >= 400 {
		return codes.InvalidArgument
	}
	return codes.Unknown
}
