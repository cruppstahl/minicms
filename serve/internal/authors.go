package internal

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Authors struct {
	FilePath string
	Authors  []Author `yaml:"authors"`
}

type Author struct {
	Name     string `yaml:"name"`
	FullName string `yaml:"fullname"`
}

func ReadAuthorsYaml(context Context, path string) (Context, error) {
	context.Authors.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Context{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &context.Authors); err != nil {
		return Context{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Check if the authors list is empty
	if len(context.Authors.Authors) == 0 {
		return Context{}, fmt.Errorf("no authors found in %s", path)
	}

	// Return the parsed data
	return context, nil
}
