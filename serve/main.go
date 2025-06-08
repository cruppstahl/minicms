package main

import (
	"log"
	"os/exec"
	"serve/impl"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	var err error
	var context impl.Context

	// parse command line arguments
	context.Config, err = impl.ParseCommandLineArguments()
	if err != nil {
		return
	}

	// Command line mode "create" copies one of the templates to a new directory
	// as a simple way to start a new project.
	if context.Config.Mode == "create" {
		log.Printf("cp -r ../site-" + context.Config.SiteDirectory + " " + context.Config.OutDirectory)
		cmd := exec.Command("cp", "-r", "../site-"+context.Config.SiteDirectory, context.Config.OutDirectory)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to execute command: %v", err)
		}
		return
	}

	// Now read all yaml files and the file tree
	err = impl.InitializeContext(&context)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}

	// Initialize the cached file system
	err = impl.InitializeFilesystem(&context)
	if err != nil {
		log.Fatalf("Failed to initialize lookup index: %v", err)
	}

	// If requested, dump the whole context and the file tree to a directory
	// This is used for testing (the directory can then be compared to
	// a "golden" set of files, and any deviation is a bug)
	if context.Config.Mode == "dump" {
		impl.Dump(&context)
		return
	}

	// From here on we assume that we run the server

	// The FsWatcher will invalidate cached file contents if the underlying file
	// is changed
	err = impl.InitializeFsWatcher(&context)
	if err != nil {
		log.Fatalf("failed to initialize file watcher: %v", err)
	}
	defer context.Watcher.Close()

	// Set up the routes
	router := gin.Default()
	err = impl.SetupRoutes(router, &context)
	if err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	// Then run the server
	err = router.Run(":" + strconv.Itoa(context.Config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
