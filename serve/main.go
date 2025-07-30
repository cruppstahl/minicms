package main

import (
	"log"
	"serve/cmd"
	"serve/core"
	"serve/plugins/contenttype"
	"serve/plugins/data"
)

func initializeBuiltinPlugins(context *core.Context) error {
	plugins := []core.Plugin{
		contenttype.NewHtmlPlugin(),
		contenttype.NewTextPlugin(),
		contenttype.NewMarkdownPlugin(),
		data.NewSearchPlugin(),
	}

	for _, plugin := range plugins {
		err := core.RegisterPlugin(context, plugin)
		if err != nil {
			return err
		}
		log.Printf("Registered plugin: %s (version: %s)", plugin.Name(), plugin.Version())
	}
	return nil
}

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
		cmd.Version()
		return
	}

	// Now read all yaml files and the file tree
	err = core.InitializeContext(&context)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}

	// Register all builtin plugins
	err = initializeBuiltinPlugins(&context)
	if err != nil {
		log.Fatalf("Failed to initialize plugin manager: %v", err)
	}

	// Initialize the cached file system
	err = core.InitializeFilesystem(&context)
	if err != nil {
		log.Fatalf("Failed to initialize lookup index: %v", err)
	}

	// Build the Navigation structure
	context.Navigation, err = core.InitializeNavigation(&context)
	if err != nil {
		log.Fatalf("Failed to initialize navigation: %v", err)
	}

	// If requested, dump the whole context and the file tree to a directory
	// This is used for testing (the directory can then be compared to
	// a "golden" set of files, and any deviation is a bug)
	if context.Config.Mode == "static" {
		cmd.Static(&context)
		return
	}

	// If any plugins need to be initialized, do it now
	err = core.InitializeDataPlugins(&context)
	if err != nil {
		log.Fatalf("Failed to initialize plugins: %v", err)
	}

	// From here on we assume that we run the server
	cmd.Run(&context)

	// Shutdown plugins
	core.ShutdownPlugins(&context.PluginManager)
}
