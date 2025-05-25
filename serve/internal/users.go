package internal

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Users struct {
	FilePath string
	Users    []User `yaml:"users"`
}

type User struct {
	Name     string `yaml:"name"`
	FullName string `yaml:"fullname"`
}

func ReadUsersYaml(context Context, path string) (Context, error) {
	context.Users.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Context{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &context.Users); err != nil {
		return Context{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Check if the authors list is empty
	if len(context.Users.Users) == 0 {
		return Context{}, fmt.Errorf("no users found in %s", path)
	}

	// Return the parsed data
	return context, nil
}
