package core

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

type Context struct {
	Users         Users
	Config        Config
	Navigation    Navigation
	Filesystem    map[string]File
	Root          Directory
	PluginManager PluginManager
	Watcher       *fsnotify.Watcher
}

func InitializeContext(context *Context) error {
	var err error

	// read config.yaml
	configFilePath := fmt.Sprintf("%s/config/site.yaml", context.Config.SiteDirectory)
	err = ReadConfigYaml(&context.Config, configFilePath)
	if err != nil {
		return err
	}

	// read users.yaml
	authorsFilePath := fmt.Sprintf("%s/config/users.yaml", context.Config.SiteDirectory)
	context.Users, err = ReadUsersYaml(authorsFilePath)
	if err != nil {
		return err
	}

	// Initialize the PluginManager
	context.PluginManager, err = CreatePluginManager()
	if err != nil {
		return err
	}

	context.Filesystem = make(map[string]File)

	return nil
}
