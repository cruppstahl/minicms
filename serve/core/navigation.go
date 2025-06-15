package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Navigation struct {
	FilePath       string
	NavigationTree []NavigationItem `yaml:"main"`
	Filesystem     map[string]File
}

type NavigationItem struct {
	LocalPath string           `yaml:"local-path"`
	Url       string           `yaml:"url"`
	Label     string           `yaml:"label"`
	Children  []NavigationItem `yaml:"children,omitempty"`
	IsActive  bool             // helper field for templating
}

func ReadNavigationYaml(path string) (Navigation, error) {
	var navigation Navigation
	navigation.FilePath = path
	navigation.Filesystem = make(map[string]File)

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
	if len(navigation.NavigationTree) == 0 {
		return Navigation{}, fmt.Errorf("no main navigation items found in %s", path)
	}

	// Enforce absolute paths for LocalPath and Url in the NavigationTree
	for _, item := range navigation.NavigationTree {
		if !filepath.IsAbs(item.LocalPath) {
			return Navigation{}, fmt.Errorf("expected absolute path for %s", item.LocalPath)
		}
		if !filepath.IsAbs(item.Url) {
			return Navigation{}, fmt.Errorf("expected absolute url for %s", item.Url)
		}
	}

	return navigation, nil
}
