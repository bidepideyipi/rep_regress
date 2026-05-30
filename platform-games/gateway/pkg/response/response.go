package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Response struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Errors    []ErrorItem `json:"errors,omitempty"`
}

type ErrorItem struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

const (
	CodeSuccess             = 200
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeForbidden           = 403
	CodeNotFound            = 404
	CodeTooManyRequests     = 429
	CodeInternalServerError = 500
	CodeServiceUnavailable  = 503

	MsgSuccess           = "Success"
	MsgBadRequest        = "Bad Request"
	MsgUnauthorized      = "Unauthorized"
	MsgForbidden         = "Forbidden"
	MsgNotFound          = "Not Found"
	MsgTooManyRequests   = "Too Many Requests"
	MsgInternalError     = "Internal Server Error"
	MsgServiceUnavailable = "Service Unavailable"

	ErrInvalidSignature    = "Invalid signature"
	ErrExpiredTimestamp    = "Expired timestamp"
	ErrReplayAttack        = "Replay attack detected"
	ErrInvalidToken        = "Invalid token"
	ErrExpiredToken        = "Expired token"
	ErrInvalidMerchant     = "Invalid merchant"
	ErrInvalidGame         = "Invalid game"
	ErrInsufficientBalance = "Insufficient balance"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   MsgSuccess,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
		RequestID: getRequestID(c),
	})
}

func Error(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Timestamp: getCurrentTimestamp(),
		RequestID: getRequestID(c),
	})
}

func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
		RequestID: getRequestID(c),
	})
}

func ErrorWithErrors(c *gin.Context, code int, message string, errors []ErrorItem) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Errors:    errors,
		Timestamp: getCurrentTimestamp(),
		RequestID: getRequestID(c),
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

func TooManyRequests(c *gin.Context, message string) {
	Error(c, CodeTooManyRequests, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, CodeInternalServerError, message)
}

func getHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeInternalServerError:
		return http.StatusInternalServerError
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	if requestID := c.Writer.Header().Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	requestID := uuid.New().String()
	c.Writer.Header().Set("X-Request-ID", requestID)
	return requestID
}
