package errorhandler

import (
	"context"
	"net/http"
)

func HandlerError(c context.Context, err error) {
	var statusCode int

	switch err.(type) {
	case *NotFoundError:
		statusCode = http.StatusNotFound
	case *BadRequestError:
		statusCode = http.StatusBadRequest
	case *InternalServerError:
		statusCode = http.StatusInternalServerError
	case *UnathorizedError:
		statusCode = http.StatusUnauthorized
	}

	// response := helper.Response(dto.ResponseParam{
	// 	StatusCode: statusCode,
	// 	Message:    err.Error(),
	// })

	c.Value(statusCode)
}
