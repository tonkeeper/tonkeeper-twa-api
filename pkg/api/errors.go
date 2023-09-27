package api

import (
	"net/http"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
)

func BadRequest(msg string) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: http.StatusBadRequest,
		Response:   oas.Error{Error: msg},
	}
}

func InternalError(err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response:   oas.Error{Error: err.Error()},
	}
}

func Unauthorized(err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: http.StatusUnauthorized,
		Response:   oas.Error{Error: err.Error()},
	}
}
