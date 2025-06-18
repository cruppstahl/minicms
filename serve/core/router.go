package core

import (
	"log"

	"github.com/gin-gonic/gin"
)

func InitializeRouter(context *Context) (*gin.Engine, error) {
	router := gin.Default()

	// Store context in the router's gin context
	router.Use(func(c *gin.Context) {
		c.Set("context", context)
		c.Next()
	})

	// Go through the Filesystem structure and set up the routes
	for url := range context.Navigation.Filesystem {
		router.GET(url, func(c *gin.Context) {
			// Don't forget type assertion when getting the connection from context.
			context, _ := c.MustGet("context").(*Context)

			file, err := GetFileWithContent(c.Request.URL.Path, context)
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

	// Add a static route
	staticDir := context.Config.SiteDirectory + "/assets"
	router.Static("/assets", staticDir)

	return router, nil
}
