package core

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Helper function to capture file
func makeFileHandler(fm *FileManager, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file := fm.GetFile(path)
		if file.Metadata.RedirectUrl != "" {
			c.Redirect(302, file.Metadata.RedirectUrl)
		} else {
			c.Data(200, file.Metadata.MimeType, []byte(file.Content))
		}
	}
}

func InitializeRouter(ctx *Context) (*gin.Engine, error) {
	router := gin.Default()

	// Store context in the router's gin context
	router.Use(func(c *gin.Context) {
		c.Set("context", ctx)
		c.Next()
	})

	// Go through the Filesystem structure and set up the routes
	for _, file := range ctx.FileManager.GetAllFiles() {
		// Only create routes for the files in "content/"
		if strings.HasPrefix(file.Path, "content/") {
			for _, route := range file.Routes {
				router.GET(route, makeFileHandler(ctx.FileManager, file.Path))
			}
		}
	}

	// Add a static route for the assets
	staticDir := ctx.Config.SiteDirectory + "/assets"
	router.Static("/assets", staticDir)

	return router, nil
}
