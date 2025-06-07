package internal

import (
	"fmt"
)

type Context struct {
	Users      Users
	Config     Config
	Navigation Navigation
	DataCache  map[string]File
}

func InitializeContext() (Context, error) {
	var err error
	var context Context
	context.DataCache = make(map[string]File)

	// parse command line arguments
	context.Config, err = ParseCommandLineArguments()
	if err != nil {
		return context, err
	}

	// read config.yaml
	configFilePath := fmt.Sprintf("%s/config/site.yaml", context.Config.SiteDirectory)
	context, err = ReadConfigYaml(context, configFilePath)
	if err != nil {
		return context, err
	}

	// read users.yaml
	authorsFilePath := fmt.Sprintf("%s/config/users.yaml", context.Config.SiteDirectory)
	context, err = ReadUsersYaml(context, authorsFilePath)
	if err != nil {
		return context, err
	}

	// read navigation.yaml
	navigationFilePath := fmt.Sprintf("%s/config/navigation.yaml", context.Config.SiteDirectory)
	context, err = ReadNavigationYaml(context, navigationFilePath)
	if err != nil {
		return context, err
	}

	// Return the parsed data
	return context, err
}
