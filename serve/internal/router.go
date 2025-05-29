package internal

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func addRoutesForDirectory(router *gin.Engine, directory *Directory, context *Context) {
	// Create a route for each file in the directory
	for _, file := range directory.Files {
		routePath := strings.TrimPrefix(file.LocalPath, context.Config.SiteDirectory+"/content/")
		router.GET(routePath, func(c *gin.Context) {
			c.Data(200, file.MimeType, []byte(file.Content))
		})
	}

	// Create a route for each subdirectory
	for _, subDir := range directory.Directories {
		addRoutesForDirectory(router, &subDir, context)
	}
}

func addRouteForNavigationItem(router *gin.Engine, item *NavigationItem, context *Context) {
	router.GET(item.RoutePath, func(c *gin.Context) {
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

		c.Data(200, file.MimeType, []byte(file.Content))
	})

	// Chefk if this item has files and subdirectories in its LocalPath
	addRoutesForDirectory(router, &item.Directory, context)

	// Check if the item has children and add routes for them recursively
	for _, child := range item.Children {
		addRouteForNavigationItem(router, &child, context)
	}
}

func SetupRoutes(router *gin.Engine, context *Context) error {
	dataTree, err := ReadDataTree(context)
	if err != nil {
		return err
	}
	context.DataTree = dataTree

	// Store context in the router's gin context
	router.Use(func(c *gin.Context) {
		c.Set("context", context)
		c.Next()
	})

	// Go through the Navigation structure and set up the routes
	for _, item := range context.Navigation.Main {
		addRouteForNavigationItem(router, &item, context)
	}

	// Walk through the data tree and set up all the routes
	// addRoute(router, &context.DataTree.Directory, 0, context)
	// router.GET("/", handlers.GetRoot)
	// router.GET("/posts", handlers.GetPosts)
	// router.GET("/posts/index", handlers.GetPosts)
	// router.GET("/posts/:post_id", handlers.GetPostByID)

	return nil
}
