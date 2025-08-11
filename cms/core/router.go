package core

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Helper function to capture file
func makeFileHandler(fm *FileManager, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file := fm.GetFile(path)
		if file == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if file.Metadata.RedirectUrl != "" {
			c.Redirect(http.StatusFound, file.Metadata.RedirectUrl)
			return
		}

		c.Data(http.StatusOK, file.Metadata.MimeType, []byte(file.Content))
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
	staticDir := filepath.Join(ctx.Config.SiteDirectory, "assets")
	router.Static("/assets", staticDir)

	return router, nil
}
