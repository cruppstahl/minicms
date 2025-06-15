package main

import (
	"log"
	"serve/core"
	"serve/cmd"
)

func main() {
	var err error
	var context core.Context

	// parse command line arguments
	context.Config, err = core.ParseCommandLineArguments()
	if err != nil {
		return
	}

	// If requested, print the version and leave
	if context.Config.Mode == "version" {
		cmd.RunVersion()
		return
	}

	// Now read all yaml files and the file tree
	err = core.InitializeContext(&context)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}

	// Initialize the cached file system
	err = core.InitializeFilesystem(&context)
	if err != nil {
		log.Fatalf("Failed to initialize lookup index: %v", err)
	}

	// If requested, dump the whole context and the file tree to a directory
	// This is used for testing (the directory can then be compared to
	// a "golden" set of files, and any deviation is a bug)
	if context.Config.Mode == "dump" {
		cmd.RunDump(&context)
		return
	}

	// From here on we assume that we run the server
	cmd.RunRun(&context)
}
