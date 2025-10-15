package core

import (
	"path/filepath"
)

type Context struct {
	Users         Users
	Config        Config
	Navigation    Navigation
	FileManager   *FileManager
	PluginManager PluginManager
	FileWatcher   *FileWatcher
	Logger        *Logger
}

func InitializeContext(ctx *Context) error {
	var err error

	// Initialize logger
	ctx.Logger = NewLogger(LogLevelInfo)

	// read config.yaml
	configFilePath := filepath.Join(ctx.Config.SiteDirectory, "config", "site.yaml")
	err = ReadConfigYaml(&ctx.Config, configFilePath)
	if err != nil {
		return err
	}

	// read users.yaml
	authorsFilePath := filepath.Join(ctx.Config.SiteDirectory, "config", "users.yaml")
	ctx.Users, err = ReadUsersYaml(authorsFilePath)
	if err != nil {
		return err
	}

	// Build the Navigation structure
	ctx.Navigation, err = InitializeNavigation(ctx)
	if err != nil {
		return err
	}

	// Register default health checks
	RegisterDefaultHealthChecks(ctx)

	return nil
}
