package main

import (
	"log"
	"serve/cmd"
	"serve/core"
	"serve/plugins/contenttype"
)

func registerPlugins(context *core.Context) {
	mgr := &context.PluginManager
	core.RegisterContentTypePlugin(mgr, contenttype.NewHtmlPlugin())
	core.RegisterContentTypePlugin(mgr, contenttype.NewTextPlugin())
	core.RegisterContentTypePlugin(mgr, contenttype.NewMarkdownPlugin())
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
	registerPlugins(&context)

	// Initialize the cached file system
	err = core.InitializeFilesystem(&context)
	if err != nil {
		log.Fatalf("Failed to initialize lookup index: %v", err)
	}

	// If requested, dump the whole context and the file tree to a directory
	// This is used for testing (the directory can then be compared to
	// a "golden" set of files, and any deviation is a bug)
	if context.Config.Mode == "static" {
		cmd.Static(&context)
		return
	}

	// From here on we assume that we run the server
	cmd.Run(&context)
}
