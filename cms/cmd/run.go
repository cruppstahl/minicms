package cmd

import (
	"cms/core"
	"fmt"
	"log"
	"strconv"
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

	// Then run the server
	err = rm.GetRouter().Run(":" + strconv.Itoa(ctx.Config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
