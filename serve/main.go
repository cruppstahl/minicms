package main

import (
	"log"
	"os/exec"
	"serve/internal"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	var err error
	var context internal.Context

	// parse command line arguments
	context.Config, err = internal.ParseCommandLineArguments()
	if err != nil {
		return
	}

	log.Printf("Running in mode: %s", context.Config.Mode)
	if context.Config.Mode == "create" {
		log.Printf("cp -r ../site-" + context.Config.SiteDirectory + " " + context.Config.OutDirectory)
		cmd := exec.Command("cp", "-r", "../site-"+context.Config.SiteDirectory, context.Config.OutDirectory)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to execute command: %v", err)
		}
		return
	}

	err = internal.InitializeContext(&context)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}

	if context.Config.Mode == "dump" {
		internal.PrettyPrint(context)
		return
	}

	// Assume that context.Config.Mode == "run"
	router := gin.Default()

	err = internal.InitializeFsWatcher(&context)
	if err != nil {
		log.Fatalf("failed to initialize file watcher: %v", err)
	}

	err = internal.InitializeFsLookup(&context)
	if err != nil {
		log.Fatalf("failed to initialize file system lookup: %v", err)
	}

	err = internal.SetupRoutes(router, &context)
	if err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}
	defer context.Watcher.Close()

	err = router.Run(":" + strconv.Itoa(context.Config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
