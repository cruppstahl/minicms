package core

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

func ReadUsersYaml(path string) (Users, error) {
	var users Users
	users.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Users{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &users); err != nil {
		return Users{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Check if the authors list is empty
	if len(users.Users) == 0 {
		return Users{}, fmt.Errorf("no users found in %s", path)
	}

	// Return the parsed data
	return users, nil
}
