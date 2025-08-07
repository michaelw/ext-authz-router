package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealthzResponse represents the health check response
type GetHealthzResponse struct {
	Status string `json:"status"`
}

// GetHealthzHandler handles the health check endpoint
func (h *AuthzHandler) GetHealthzHandler(c *gin.Context) {
	response := GetHealthzResponse{
		Status: "UP",
	}
	c.JSON(http.StatusOK, response)
}

// GetOpenAPIJSONHandler serves the OpenAPI specification as JSON
func (h *AuthzHandler) GetOpenAPIJSONHandler(c *gin.Context) {
	if h.Swagger == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAPI spec not available"})
		return
	}

	c.JSON(http.StatusOK, h.Swagger)
}

// RegisterRoutes registers internal server routes
func (h *AuthzHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/ready", h.GetHealthzHandler)    // ready to serve requests
	router.GET("/readyz", h.GetHealthzHandler)   // alias
	router.GET("/health", h.GetHealthzHandler)   // live, but may not be ready
	router.GET("/healthz", h.GetHealthzHandler)  // alias
	router.GET("/startupz", h.GetHealthzHandler) // startup check
}
