package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Navigation struct {
	FilePath string
	Children []NavigationItem `yaml:"main"`
}

type NavigationItem struct {
	Url         string           `yaml:"url"`
	Title       string           `yaml:"title"`
	Children    []NavigationItem `yaml:"children,omitempty"`
	IsActive    bool             // helper field for templating
	IsDirectory bool
}

func readNavigationYaml(path string) (Navigation, error) {
	var navigation Navigation
	navigation.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Navigation{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &navigation); err != nil {
		return Navigation{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// We need at least one main navigation item
	if len(navigation.Children) == 0 {
		return Navigation{}, fmt.Errorf("no main navigation items found in %s", path)
	}

	// Enforce absolute paths
	for _, item := range navigation.Children {
		if !filepath.IsAbs(item.Url) {
			return Navigation{}, fmt.Errorf("expected absolute url for %s", item.Url)
		}
	}

	return navigation, nil
}

func InitializeNavigation(context *Context) (Navigation, error) {
	// Read the navigation.yaml file
	path := fmt.Sprintf("%s/config/navigation.yaml", context.Config.SiteDirectory)
	navigation, err := readNavigationYaml(path)
	if err != nil {
		return Navigation{}, fmt.Errorf("failed to read navigation.yaml: %w", err)
	}
	return navigation, nil
}
