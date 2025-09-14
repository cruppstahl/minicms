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
		// Get FileManager with read lock to ensure it's not nil
		rm.mu.RLock()
		fm := rm.fm
		rm.mu.RUnlock()

		if fm == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		file := fm.GetFile(filePath)
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

		// Call addFileUnsafe since we already hold the lock
		rm.addFileUnsafe(file)
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

// AddFile is now thread-safe
func (rm *RouterManager) AddFile(file *File) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.addFileUnsafe(file)
}

// addFileUnsafe is the internal implementation that assumes the lock is already held
func (rm *RouterManager) addFileUnsafe(file *File) {
	for _, route := range file.Routes {
		normalizedRoute, err := normalizeRoute(route)
		if err != nil {
			continue // Skip invalid routes
		}

		// Skip duplicates
		if _, exists := rm.routes[normalizedRoute]; exists {
			continue
		}

		rm.routes[normalizedRoute] = file.Path
		rm.router.GET(normalizedRoute, rm.makeFileHandler(file.Path))
	}
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

	return rm.rebuildRouterUnsafe()
}

// RemoveFile removes all routes associated with a file
func (rm *RouterManager) RemoveFile(filePath string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find all routes that map to this file
	var routesToRemove []string
	for pattern, fp := range rm.routes {
		if fp == filePath {
			routesToRemove = append(routesToRemove, pattern)
		}
	}

	if len(routesToRemove) == 0 {
		return fmt.Errorf("no routes found for file %s", filePath)
	}

	// Remove all found routes
	for _, pattern := range routesToRemove {
		delete(rm.routes, pattern)
	}

	return rm.rebuildRouterUnsafe()
}

// GetAllRoutes returns a copy of all current routes (thread-safe)
func (rm *RouterManager) GetAllRoutes() map[string]string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modifications
	routes := make(map[string]string, len(rm.routes))
	for pattern, filePath := range rm.routes {
		routes[pattern] = filePath
	}
	return routes
}

// rebuildRouterUnsafe recreates the router with current routes
// This method assumes the caller already holds the write lock
func (rm *RouterManager) rebuildRouterUnsafe() error {
	// Store current routes (we already have the lock, so this is safe)
	currentRoutes := make(map[string]string, len(rm.routes))
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
	if rm.ctx != nil {
		staticDir := filepath.Join(rm.ctx.Config.SiteDirectory, "assets")
		newRouter.Static("/assets", staticDir)
	}

	rm.router = newRouter
	return nil
}

// rebuildRouter is the public version that acquires the lock
func (rm *RouterManager) rebuildRouter() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.rebuildRouterUnsafe()
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

// GetRouter returns the current router (thread-safe)
func (rm *RouterManager) GetRouter() *gin.Engine {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.router
}

// RouteExists checks if a route pattern exists (thread-safe)
func (rm *RouterManager) RouteExists(pattern string) bool {
	normalizedPattern, err := normalizeRoute(pattern)
	if err != nil {
		return false
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, exists := rm.routes[normalizedPattern]
	return exists
}

// GetRouteCount returns the number of registered routes (thread-safe)
func (rm *RouterManager) GetRouteCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.routes)
}
