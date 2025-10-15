package core

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string                          `json:"name"`
	Status      HealthStatus                    `json:"status"`
	Message     string                          `json:"message,omitempty"`
	LastChecked time.Time                       `json:"last_checked"`
	Duration    time.Duration                   `json:"duration"`
	CheckFunc   func(ctx context.Context) error `json:"-"`
}

// HealthChecker manages health checks for the application
type HealthChecker struct {
	mu           sync.RWMutex
	checks       map[string]*HealthCheck
	globalStatus HealthStatus
	lastUpdate   time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:       make(map[string]*HealthCheck),
		globalStatus: HealthStatusUnknown,
		lastUpdate:   time.Now(),
	}
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, checkFunc func(ctx context.Context) error) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = &HealthCheck{
		Name:        name,
		Status:      HealthStatusUnknown,
		CheckFunc:   checkFunc,
		LastChecked: time.Time{},
	}
}

// UnregisterCheck removes a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.checks, name)
}

// RunCheck executes a specific health check
func (hc *HealthChecker) RunCheck(ctx context.Context, name string) error {
	hc.mu.RLock()
	check, exists := hc.checks[name]
	hc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("health check %s not found", name)
	}

	start := time.Now()
	err := check.CheckFunc(ctx)
	duration := time.Since(start)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	check.Duration = duration
	check.LastChecked = time.Now()

	if err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = err.Error()
	} else {
		check.Status = HealthStatusHealthy
		check.Message = ""
	}

	return err
}

// RunAllChecks executes all registered health checks
func (hc *HealthChecker) RunAllChecks(ctx context.Context) map[string]error {
	hc.mu.RLock()
	checkNames := make([]string, 0, len(hc.checks))
	for name := range hc.checks {
		checkNames = append(checkNames, name)
	}
	hc.mu.RUnlock()

	errors := make(map[string]error)
	for _, name := range checkNames {
		if err := hc.RunCheck(ctx, name); err != nil {
			errors[name] = err
		}
	}

	hc.updateGlobalStatus()
	return errors
}

// updateGlobalStatus calculates the overall health status
func (hc *HealthChecker) updateGlobalStatus() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if len(hc.checks) == 0 {
		hc.globalStatus = HealthStatusUnknown
		return
	}

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0
	unknownCount := 0

	for _, check := range hc.checks {
		switch check.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnknown:
			unknownCount++
		}
	}

	// Determine global status
	if unhealthyCount > 0 {
		hc.globalStatus = HealthStatusUnhealthy
	} else if degradedCount > 0 {
		hc.globalStatus = HealthStatusDegraded
	} else if unknownCount > 0 {
		hc.globalStatus = HealthStatusUnknown
	} else if healthyCount > 0 {
		hc.globalStatus = HealthStatusHealthy
	} else {
		hc.globalStatus = HealthStatusUnknown
	}

	hc.lastUpdate = time.Now()
}

// GetStatus returns the current health status
func (hc *HealthChecker) GetStatus() (HealthStatus, map[string]*HealthCheck) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Create a copy of checks to avoid race conditions
	checksCopy := make(map[string]*HealthCheck, len(hc.checks))
	for name, check := range hc.checks {
		checksCopy[name] = &HealthCheck{
			Name:        check.Name,
			Status:      check.Status,
			Message:     check.Message,
			LastChecked: check.LastChecked,
			Duration:    check.Duration,
		}
	}

	return hc.globalStatus, checksCopy
}

// HealthHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		errors := hc.RunAllChecks(ctx)
		globalStatus, checks := hc.GetStatus()

		response := gin.H{
			"status":      globalStatus,
			"timestamp":   time.Now(),
			"checks":      checks,
			"last_update": hc.lastUpdate,
		}

		if len(errors) > 0 {
			response["errors"] = errors
		}

		// Set appropriate HTTP status code
		var httpStatus int
		switch globalStatus {
		case HealthStatusHealthy:
			httpStatus = http.StatusOK
		case HealthStatusDegraded:
			httpStatus = http.StatusOK // 200 but with degraded status
		case HealthStatusUnhealthy:
			httpStatus = http.StatusServiceUnavailable
		case HealthStatusUnknown:
			httpStatus = http.StatusServiceUnavailable
		default:
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, response)
	}
}

// LivenessHandler returns a simple liveness probe handler
func (hc *HealthChecker) LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	}
}

// ReadinessHandler returns a readiness probe handler
func (hc *HealthChecker) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		globalStatus, _ := hc.GetStatus()

		if globalStatus == HealthStatusHealthy || globalStatus == HealthStatusDegraded {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ready",
				"timestamp": time.Now(),
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not_ready",
				"timestamp": time.Now(),
			})
		}
	}
}

// StartPeriodicChecks starts running health checks periodically
func (hc *HealthChecker) StartPeriodicChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check
	hc.RunAllChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			errors := hc.RunAllChecks(checkCtx)
			cancel()

			// Log any errors
			for name, err := range errors {
				Error("Health check %s failed: %v", name, err)
			}
		}
	}
}

// Predefined health checks for MiniCMS components

// FileManagerHealthCheck checks if the file manager is working properly
func FileManagerHealthCheck(fm *FileManager) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if fm == nil {
			return fmt.Errorf("file manager is nil")
		}

		// Check if we can access the root directory
		root := fm.GetRoot()
		if root == nil {
			return fmt.Errorf("file manager root is nil")
		}

		// Check if we can get all files
		files := fm.GetAllFiles()
		if files == nil {
			return fmt.Errorf("could not retrieve files from file manager")
		}

		return nil
	}
}

// FileWatcherHealthCheck checks if the file watcher is running
func FileWatcherHealthCheck(fw *FileWatcher) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if fw == nil {
			return fmt.Errorf("file watcher is nil")
		}

		if !fw.IsRunning() {
			return fmt.Errorf("file watcher is not running")
		}

		// Check if we have any watched directories
		dirs := fw.GetWatchedDirectories()
		if len(dirs) == 0 {
			return fmt.Errorf("no directories being watched")
		}

		return nil
	}
}

// RouterHealthCheck checks if the router is working
func RouterHealthCheck(rm *RouterManager) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if rm == nil {
			return fmt.Errorf("router manager is nil")
		}

		router := rm.GetRouter()
		if router == nil {
			return fmt.Errorf("gin router is nil")
		}

		// Check if we have any routes
		routeCount := rm.GetRouteCount()
		if routeCount == 0 {
			return fmt.Errorf("no routes registered")
		}

		return nil
	}
}

// PluginManagerHealthCheck checks if plugins are loaded
func PluginManagerHealthCheck(pm *PluginManager) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if pm == nil {
			return fmt.Errorf("plugin manager is nil")
		}

		plugins := pm.ListPlugins()
		if len(plugins) == 0 {
			return fmt.Errorf("no plugins registered")
		}

		return nil
	}
}

// Global health checker instance
var GlobalHealthChecker = NewHealthChecker()

// RegisterDefaultHealthChecks registers standard health checks for MiniCMS
func RegisterDefaultHealthChecks(ctx *Context) {
	if ctx.FileManager != nil {
		GlobalHealthChecker.RegisterCheck("file_manager", FileManagerHealthCheck(ctx.FileManager))
	}

	if ctx.FileWatcher != nil {
		GlobalHealthChecker.RegisterCheck("file_watcher", FileWatcherHealthCheck(ctx.FileWatcher))
	}

	// Register plugin manager check
	GlobalHealthChecker.RegisterCheck("plugin_manager", PluginManagerHealthCheck(&ctx.PluginManager))

	// System-level checks
	GlobalHealthChecker.RegisterCheck("memory", func(ctx context.Context) error {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Alert if memory usage is over 1GB
		if memStats.Alloc > 1024*1024*1024 {
			return fmt.Errorf("high memory usage: %d bytes", memStats.Alloc)
		}

		return nil
	})

	GlobalHealthChecker.RegisterCheck("goroutines", func(ctx context.Context) error {
		count := runtime.NumGoroutine()

		// Alert if we have too many goroutines (possible leak)
		if count > 1000 {
			return fmt.Errorf("high goroutine count: %d", count)
		}

		return nil
	})
}