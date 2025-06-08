package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

func addRoute(router *gin.Engine, url string) {
	router.GET(url, func(c *gin.Context) {
		// Don't forget type assertion when getting the connection from context.
		context, _ := c.MustGet("context").(*Context)

		file, err := getFile(c.Request.URL.Path, context)
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

func normalizePath(path string) string {
	// Convert double slashes to single slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// Also convert double slashes to single slashes
	return strings.ReplaceAll(path, "//", "/")
}

func fetchFileBody(file *File, context *Context) (string, error) {
	header, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/header.html")
	footer, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/footer.html")

	// Read content from file.LocalPath and store it in file.Content
	body, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return "", err
	}
	return string(header) + string(body) + string(footer), nil
}

func applyTemplate(body string, file *File, context *Context) (string, error) {
	tmpl, err := template.New(file.LocalPath).Parse(body)
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.LocalPath, err)
		return "", err
	}
	var output strings.Builder
	var vars = map[string]interface{}{
		"SiteTitle":        context.Config.Server.Title,
		"SiteDescription":  context.Config.Server.Description,
		"Site.Author.Name": context.Users.Users[0].Name, // Assuming at least one user exists
		"BrandingFavicon":  context.Config.Branding.Favicon,
		"BrandingCssFile":  context.Config.Branding.CssFile,
		"Directory.Title":  file.Directory.Title,
		"FileTitle":        file.Title,
		"FileAuthor":       file.Author,
		"FileTags":         file.Tags,
		"FileImagePath":    file.ImagePath,
		"FileCssFile":      file.CssFile,
		"FileMimeType":     file.MimeType,
		"FileLocalPath":    file.LocalPath,
		"ActiveUrl":        "", // This will be set below
	}

	// Go through all NavigationItems. If their URL matches the current file's URL,
	// then set the "active" variable to true.
	navTree := context.Navigation.NavigationTree
	for i, item := range navTree {
		item.IsActive = strings.HasSuffix(file.LocalPath, item.LocalPath)
		navTree[i] = item
	}
	vars["Navigation"] = navTree

	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.LocalPath, err)
		return "", err
	}

	return output.String(), nil
}

func getFile(path string, context *Context) (*File, error) {
	normalizedPath := normalizePath(path)

	file, ok := context.Navigation.LookupIndex[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", normalizedPath)
	}

	// If the file is not cached then build it
	if len(file.CachedContent) == 0 {
		body, err := fetchFileBody(&file, context)
		if err != nil {
			return nil, err
		}

		// ignore errors, just display the template as is if it cannot be applied
		body, err = applyTemplate(body, &file, context)
		if err == nil {
			file.CachedContent = []byte(body)
			context.Navigation.LookupIndex[normalizedPath] = file
		}
	}

	return &file, nil
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

	// Add a static route
	staticDir := context.Config.SiteDirectory + "/assets"
	router.Static("/assets", staticDir)

	return nil
}
