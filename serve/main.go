package main

import (
	"log"
	"serve/internal"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	context, err := internal.InitializeContext()
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}
	internal.SetupRoutes(router, context)
	if err := router.Run(":" + strconv.Itoa(context.Config.Server.Port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
