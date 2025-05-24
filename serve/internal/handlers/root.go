package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRoot handles requests to the root route (/)
func GetRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Go HTTP Server!"})
}
