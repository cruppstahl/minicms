package internal

import (
	"fmt"
)

type Context struct {
	Authors Authors
	Config  Config
}

func InitializeContext() (Context, error) {
	var context Context
	var err error

	// parse command line arguments
	context.Config, err = ParseCommandLineArguments()
	if err != nil {
		return context, err
	}

	// read config.yaml
	configFilePath := fmt.Sprintf("%s/config.yaml", context.Config.DataDirectory)
	context, err = ReadConfigYaml(context, configFilePath)
	if err != nil {
		return context, err
	}

	// read authors.yaml
	authorsFilePath := fmt.Sprintf("%s/authors.yaml", context.Config.DataDirectory)
	context, err = ReadAuthorsYaml(context, authorsFilePath)
	if err != nil {
		return context, err
	}

	// Return the parsed data
	return context, err
}
