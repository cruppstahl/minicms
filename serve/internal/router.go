package internal

import (
	"log"

	"github.com/gin-gonic/gin"
)

func addRoute(router *gin.Engine, url string) {
	router.GET(url, func(c *gin.Context) {
		// Don't forget type assertion when getting the connection from context.
		context, _ := c.MustGet("context").(*Context)

		file, err := GetFile(c.Request.URL.Path, context)
		if err != nil {
			log.Printf("Failed to get file for path %s: %v", c.Request.URL.Path, err)
			c.HTML(500, "error.html", gin.H{
				"message": "Internal Server Error",
			})
			return
		}

		c.Data(200, file.MimeType, []byte(file.CachedContent))
	})
}

func SetupRoutes(router *gin.Engine, context *Context) error {
	// Store context in the router's gin context
	router.Use(func(c *gin.Context) {
		c.Set("context", context)
		c.Next()
	})

	// Go through the LookupIndex structure and set up the routes
	for url, _ := range context.Navigation.LookupIndex {
		addRoute(router, url)
	}

	return nil
}
