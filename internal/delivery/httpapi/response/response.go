package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/rohitashk/golang-rest-api/internal/domain"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err error) {
	requestID, _ := c.Get("request_id")
	rid, _ := requestID.(string)

	status := http.StatusInternalServerError
	body := ErrorBody{
		Code:      "internal",
		Message:   "internal server error",
		RequestID: rid,
	}

	// Treat request binding / JSON / validation errors as 400s.
	// (Gin uses go-playground/validator under the hood for binding tags.)
	var ve validator.ValidationErrors
	var ute *json.UnmarshalTypeError
	var se *json.SyntaxError
	if errors.As(err, &ve) || errors.As(err, &ute) || errors.As(err, &se) {
		status = http.StatusBadRequest
		body.Code = string(domain.ErrKindValidation)
		body.Message = "invalid request"
		c.JSON(status, gin.H{"error": body})
		return
	}

	var derr domain.Error
	if errors.As(err, &derr) {
		body.Code = string(derr.Kind)
		body.Message = derr.Message

		switch derr.Kind {
		case domain.ErrKindValidation:
			status = http.StatusBadRequest
		case domain.ErrKindNotFound:
			status = http.StatusNotFound
		case domain.ErrKindConflict:
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
			body.Message = "internal server error"
		}
	}

	c.JSON(status, gin.H{"error": body})
}
