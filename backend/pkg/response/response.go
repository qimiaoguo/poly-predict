package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// envelope is the standard JSON response wrapper.
type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *errorBody  `json:"error,omitempty"`
}

// errorBody carries error details in a response.
type errorBody struct {
	Message string `json:"message"`
}

// paginatedEnvelope wraps paginated responses.
type paginatedEnvelope struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination pagination  `json:"pagination"`
}

// pagination holds pagination metadata.
type pagination struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Pages    int   `json:"pages"`
}

// Success responds with HTTP 200 and the given data.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, envelope{
		Success: true,
		Data:    data,
	})
}

// Created responds with HTTP 201 and the given data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, envelope{
		Success: true,
		Data:    data,
	})
}

// Error responds with the given HTTP status code and error message.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, envelope{
		Success: false,
		Error:   &errorBody{Message: message},
	})
}

// ValidationError responds with HTTP 422 and the given validation message.
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, envelope{
		Success: false,
		Error:   &errorBody{Message: message},
	})
}

// Paginated responds with HTTP 200, the given data, and pagination metadata.
func Paginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	c.JSON(http.StatusOK, paginatedEnvelope{
		Success: true,
		Data:    data,
		Pagination: pagination{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Pages:    pages,
		},
	})
}
