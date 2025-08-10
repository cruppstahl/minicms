package core

import (
	"fmt"
	"sort"
	"sync"
)

// PluginContext provides context information to plugins
type PluginContext struct {
	File          *File
	FileManager   *FileManager
	SiteDirectory string // Path to the site root
}

// PluginResult represents the result of plugin execution
type PluginResult struct {
	Success      bool
	Error        error
	Modified     bool              // Whether the file was modified
	NewContent   []byte            // New content if file was modified
	OutputFiles  map[string][]byte // Additional files created
	MimeType     string            // mime type of the file
	Routes       []string          // Routes this file should be associated with
	Dependencies []*File           // Dependencies this file has
}

// Plugin interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// CanProcess determines if this plugin can process the given file
	CanProcess(file *File) bool

	// Process processes the file and returns the result
	Process(ctx *PluginContext) *PluginResult

	// Priority returns the execution priority (lower numbers = higher priority)
	Priority() int
}

// PluginManager manages all registered plugins
type PluginManager struct {
	mu      sync.RWMutex
	plugins []Plugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make([]Plugin, 0),
	}
}

// RegisterPlugin registers a new plugin
func (pm *PluginManager) RegisterPlugin(plugin Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.plugins = append(pm.plugins, plugin)

	// Sort plugins by priority (lower numbers first)
	sort.Slice(pm.plugins, func(i, j int) bool {
		return pm.plugins[i].Priority() < pm.plugins[j].Priority()
	})
}

// GetPluginsForFile returns all plugins that can process the given file
func (pm *PluginManager) GetPluginsForFile(file *File) []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var matchingPlugins []Plugin
	for _, plugin := range pm.plugins {
		if plugin.CanProcess(file) {
			matchingPlugins = append(matchingPlugins, plugin)
		}
	}

	return matchingPlugins
}

// ListPlugins returns information about all registered plugins
func (pm *PluginManager) ListPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var list []string
	for _, plugin := range pm.plugins {
		list = append(list, fmt.Sprintf("%s (priority: %d)", plugin.Name(), plugin.Priority()))
	}

	return list
}

// Processes a file with all applicable plugins. Returns a copy of the modified file.
func (pm *PluginManager) Process(copy File, fm *FileManager) *File {
	plugins := pm.GetPluginsForFile(&copy)

	// TODO we need to lock the file, or swap it atomically

	ctx := &PluginContext{
		File:          &copy,
		FileManager:   fm,
		SiteDirectory: fm.SiteDirectory,
	}

	for _, plugin := range plugins {
		result := plugin.Process(ctx)

		// If plugin modified the file content, update it
		if result.Modified && result.NewContent != nil {
			copy.Content = result.NewContent
		}

		// Handle additional output files
		/*
			if result.OutputFiles != nil {
				for outputPath, content := range result.OutputFiles {
					// Add output files to the file manager
					file := fm.AddFile(outputPath)
					file.Content = content
					// TODO also add dependencies!?
				}
			}
		*/

		// Store dependencies
		for _, dep := range result.Dependencies {
			copy.AddDependency(dep)
		}

		// Merge metadata
		copy.Metadata.MimeType = result.MimeType

		// Collect routes
		if result.Routes != nil {
			copy.Routes = result.Routes
		}
	}

	return &copy
}
