package cmd

import (
	"cms/core"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func initializeFsWatcher(ctx *core.Context) error {
	// Initialize the file watcher
	watcher, err := core.NewFileWatcher(ctx.FileManager)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx.FileWatcher = watcher

	// Start watching the content directory
	err = watcher.Start(ctx.Config.SiteDirectory)
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	return nil
}

func Run(ctx *core.Context) {
	// The FsWatcher will invalidate cached file contents if the underlying file
	// is changed
	err := initializeFsWatcher(ctx)
	if err != nil {
		log.Fatalf("failed to initialize file watcher: %v", err)
	}
	defer ctx.FileWatcher.Stop()

	// Set up the routes
	rm := core.NewRouterManager()
	err = rm.InitializeRouter(ctx)
	if err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	ctx.FileWatcher.SetRouter(rm)

	// Install the file watcher listener
	listener, err := core.RegisterFileWatcherListener(ctx.FileWatcher)
	if err != nil {
		log.Fatalf("Failed to register file watcher listener: %v", err)
	}
	defer listener.Stop()

	// Start monitoring services
	monitoringCtx, cancelMonitoring := context.WithCancel(context.Background())
	defer cancelMonitoring()

	// Start metrics collection
	go core.GlobalMetrics.StartMetricsCollector(monitoringCtx)

	// Start health checks (every 60 seconds)
	go core.GlobalHealthChecker.StartPeriodicChecks(monitoringCtx, 60*time.Second)

	// Create HTTP server with security settings
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(ctx.Config.Server.Port),
		Handler:      rm.GetRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on :%d", ctx.Config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
