package main

import (
	"cms/cmd"
	"cms/core"
	"cms/plugins"
	"fmt"
	"log"
)

func initializeAndRunPlugins(ctx *core.Context) error {
	fm := ctx.FileManager
	pm := fm.GetPluginManager()
	pm.RegisterPlugin(&plugins.BuiltinHtmlPlugin{Context: ctx})
	pm.RegisterPlugin(&plugins.BuiltinTextPlugin{})
	pm.RegisterPlugin(plugins.NewMarkdownPlugin(ctx))

	if params, exists := ctx.Config.Plugins["builtin/search"]; exists {
		pm.RegisterPlugin(plugins.NewSearchPlugin(params))
	}

	// Print all plugins including their priority
	fmt.Println("Plugins:")
	for _, plugin := range ctx.FileManager.GetPluginManager().ListPlugins() {
		fmt.Printf(" - %s\n", plugin)
	}

	// Then invoke all plugins on the files
	fm.ProcessAllFiles()

	return nil
}

func initializeFileManager(ctx *core.Context) error {
	fm := core.NewFileManager(ctx.Config.SiteDirectory)

	// Load the entire "content" directory structure
	err := fm.WalkDirectory("content")
	if err != nil {
		return err
	}

	// ... and the layout directory
	err = fm.WalkDirectory("layout")
	if err != nil {
		return err
	}

	ctx.FileManager = fm
	return nil
}

func main() {
	var err error
	var ctx core.Context

	// parse command line arguments
	ctx.Config, err = core.ParseCommandLineArguments()
	if err != nil {
		return
	}

	// If requested, print the version and leave
	if ctx.Config.Mode == "version" {
		cmd.Version()
		return
	}

	// Now read all yaml files and the file tree
	err = core.InitializeContext(&ctx)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}

	// Initialize the cached file system
	err = initializeFileManager(&ctx)
	if err != nil {
		log.Fatalf("Failed to initialize lookup index: %v", err)
	}

	// Initialize and run all builtin plugins
	err = initializeAndRunPlugins(&ctx)
	if err != nil {
		log.Fatalf("Failed to initialize plugin manager: %v", err)
	}

	// If requested, dump the whole context and the file tree to a directory
	// This is used for testing (the directory can then be compared to
	// a "golden" set of files, and any deviation is a bug)
	if ctx.Config.Mode == "static" || ctx.Config.Mode == "dump" {
		cmd.Dump(&ctx, ctx.Config.Mode == "dump")
		return
	}

	// From here on we assume that we run the server
	cmd.Run(&ctx)
}
