package core

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// RouteInfo holds information about a registered route
type RouteInfo struct {
	Pattern  string
	FilePath string
	Method   string
}

// RouterManager manages dynamic route registration and removal
type RouterManager struct {
	mu         sync.RWMutex
	router     *gin.Engine
	routes     map[string]string // pattern -> filePath mapping
	fm         *FileManager
	ctx        *Context
	middleware []gin.HandlerFunc
}

func NewRouterManager() *RouterManager {
	return &RouterManager{
		routes:     make(map[string]string),
		middleware: make([]gin.HandlerFunc, 0),
	}
}

func (rm *RouterManager) AddMiddleware(middleware ...gin.HandlerFunc) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.middleware = append(rm.middleware, middleware...)
}

// creates a handler function for a specific file path
func (rm *RouterManager) makeFileHandler(filePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rm.fm == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		file := rm.fm.GetFile(filePath)
		if file == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Handle redirects
		if file.Metadata.RedirectUrl != "" {
			c.Redirect(http.StatusFound, file.Metadata.RedirectUrl)
			return
		}

		// Set appropriate headers
		mimeType := file.Metadata.MimeType
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		c.Data(http.StatusOK, mimeType, file.Content)
	}
}

// ensures the route starts with / and has no double slashes
func normalizeRoute(route string) (string, error) {
	if route == "" {
		return "", errors.New("route cannot be empty")
	}

	// Clean the path
	route = filepath.Clean("/" + strings.TrimPrefix(route, "/"))

	// filepath.Clean converts "/" to ".", so fix that
	if route == "." {
		route = "/"
	}

	// Validate the route
	if !strings.HasPrefix(route, "/") {
		return "", fmt.Errorf("route must start with '/': %s", route)
	}

	return route, nil
}

// creates and configures the gin router with all current files
func (rm *RouterManager) InitializeRouter(ctx *Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Create new router
	rm.router = gin.New()
	rm.fm = ctx.FileManager
	rm.ctx = ctx

	// Add default middleware
	rm.router.Use(gin.Logger())
	rm.router.Use(gin.Recovery())

	// Add custom middleware
	for _, middleware := range rm.middleware {
		rm.router.Use(middleware)
	}

	// Clear existing routes map
	rm.routes = make(map[string]string)

	// Set up routes for files in content directory
	for _, file := range ctx.FileManager.GetAllFiles() {
		if !strings.HasPrefix(file.Path, "content/") {
			continue
		}

		for _, route := range file.Routes {
			normalizedRoute, err := normalizeRoute(route)
			if err != nil {
				continue // Skip invalid routes
			}

			// Skip duplicates during initialization
			if _, exists := rm.routes[normalizedRoute]; exists {
				continue
			}

			rm.routes[normalizedRoute] = file.Path
			rm.router.GET(normalizedRoute, rm.makeFileHandler(file.Path))
		}
	}

	// Add static file serving for assets
	staticDir := filepath.Join(ctx.Config.SiteDirectory, "assets")
	rm.router.Static("/assets", staticDir)

	return nil
}

func (rm *RouterManager) AddRoute(pattern, filePath string) error {
	normalizedPattern, err := normalizeRoute(pattern)
	if err != nil {
		return fmt.Errorf("invalid route pattern: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if route already exists
	if _, exists := rm.routes[normalizedPattern]; exists {
		return fmt.Errorf("route %s already exists", normalizedPattern)
	}

	// Add route to router
	rm.router.GET(normalizedPattern, rm.makeFileHandler(filePath))
	rm.routes[normalizedPattern] = filePath

	return nil
}

func (rm *RouterManager) RemoveRoute(pattern string) error {
	normalizedPattern, err := normalizeRoute(pattern)
	if err != nil {
		return fmt.Errorf("invalid route pattern: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if route exists
	if _, exists := rm.routes[normalizedPattern]; !exists {
		return fmt.Errorf("route %s not found", normalizedPattern)
	}

	// Remove from our routes map
	delete(rm.routes, normalizedPattern)

	// Note: Gin doesn't support removing routes dynamically.
	// We need to rebuild the router to actually remove routes.
	// For now, we'll just remove from our tracking map.
	// The route will still exist in gin but won't be in our management system.

	return rm.rebuildRouter()
}

// rebuildRouter recreates the router with current routes
func (rm *RouterManager) rebuildRouter() error {
	// Store current routes
	currentRoutes := make(map[string]string)
	for pattern, filePath := range rm.routes {
		currentRoutes[pattern] = filePath
	}

	// Initialize fresh router
	newRouter := gin.New()

	// Add default middleware
	newRouter.Use(gin.Logger())
	newRouter.Use(gin.Recovery())

	// Add custom middleware
	for _, middleware := range rm.middleware {
		newRouter.Use(middleware)
	}

	// Re-add all current routes
	for pattern, filePath := range currentRoutes {
		newRouter.GET(pattern, rm.makeFileHandler(filePath))
	}

	// Add static file serving for assets
	staticDir := filepath.Join(rm.ctx.Config.SiteDirectory, "assets")
	newRouter.Static("/assets", staticDir)

	rm.router = newRouter
	return nil
}

func (rm *RouterManager) GetRouteInfo(pattern string) (*RouteInfo, error) {
	normalizedPattern, err := normalizeRoute(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid route pattern: %w", err)
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	filePath, exists := rm.routes[normalizedPattern]
	if !exists {
		return nil, fmt.Errorf("route %s not found", normalizedPattern)
	}

	return &RouteInfo{
		Pattern:  normalizedPattern,
		FilePath: filePath,
		Method:   "GET",
	}, nil
}

func (rm *RouterManager) GetRouter() *gin.Engine {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.router
}
