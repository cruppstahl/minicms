package core

import (
	"fmt"
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
	configFilePath := fmt.Sprintf("%s/config/site.yaml", ctx.Config.SiteDirectory)
	err = ReadConfigYaml(&ctx.Config, configFilePath)
	if err != nil {
		return err
	}

	// read users.yaml
	authorsFilePath := fmt.Sprintf("%s/config/users.yaml", ctx.Config.SiteDirectory)
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
