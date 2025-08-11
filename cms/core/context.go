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
}

func InitializeContext(ctx *Context) error {
	var err error

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

	return nil
}
