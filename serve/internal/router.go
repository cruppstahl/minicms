package internal

import (
	"serve/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, context Context) {
	router.GET("/", handlers.GetRoot)
	router.GET("/posts", handlers.GetPosts)
	router.GET("/posts/index", handlers.GetPosts)
	router.GET("/posts/:post_id", handlers.GetPostByID)
}
