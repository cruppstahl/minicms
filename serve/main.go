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

	err = internal.SetupRoutes(router, &context)
	if err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	// internal.PrettyPrint(context)

	err = router.Run(":" + strconv.Itoa(context.Config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
