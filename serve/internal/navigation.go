package internal

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Navigation struct {
	FilePath string
	Main     []NavigationItem `yaml:"main"`
}

type NavigationItem struct {
	Name     string           `yaml:"name"`
	URL      string           `yaml:"url"`
	Children []NavigationItem `yaml:"children,omitempty"`
}

func ReadNavigationYaml(context Context, path string) (Context, error) {
	context.Navigation.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Context{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &context.Navigation); err != nil {
		return Context{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// We need at least one main navigation item
	if len(context.Navigation.Main) == 0 {
		return Context{}, fmt.Errorf("no main navigation items found in %s", path)
	}

	// Return the parsed data
	return context, nil
}
