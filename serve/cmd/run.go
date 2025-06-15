package cmd

import (
	"log"
	"strconv"
	"serve/core"
)

func Run(context *core.Context) {
	// The FsWatcher will invalidate cached file contents if the underlying file
	// is changed
	err := core.InitializeFsWatcher(context)
	if err != nil {
		log.Fatalf("failed to initialize file watcher: %v", err)
	}
	defer context.Watcher.Close()

	// Set up the routes
	router, err := core.InitializeRouter(context)
	if err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	// Then run the server
	err = router.Run(":" + strconv.Itoa(context.Config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
