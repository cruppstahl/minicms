package core

import (
	"errors"
	"strings"
)

type Plugin interface {
	Name() string
	Version() string
	Description() string
	Initialize(params map[string]string) error
	Shutdown()
}

type ContentTypePlugin interface {
	Plugin
	Mimetype() string
	IgnoreLayout() bool
	Extensions() []string

	Render(raw string) (string, error)
}

type DataPlugin interface {
	Plugin

	AddFile(file *File) error
}

type PluginManager struct {
	plugins        map[string]Plugin
	contentTypeMap map[string]ContentTypePlugin
}

var ErrPluginAlreadyRegistered = errors.New("Plugin already registered")
var ErrExtensionAlreadyRegistered = errors.New("File extension already registered")

func RegisterPlugin(context *Context, plugin Plugin) error {
	var err error
	manager := &context.PluginManager
	name := plugin.Name()

	if _, exists := manager.plugins[name]; exists {
		return ErrPluginAlreadyRegistered
	}
	manager.plugins[name] = plugin

	// If the plugin is a ContentTypePlugin, register its extensions
	if contentTypePlugin, ok := plugin.(ContentTypePlugin); ok {
		for _, ext := range contentTypePlugin.Extensions() {
			ext = strings.ToLower(ext)
			if _, exists := manager.contentTypeMap[ext]; exists {
				return ErrExtensionAlreadyRegistered
			}
			manager.contentTypeMap[ext] = contentTypePlugin
		}
	}

	// Initialize it with the configuration options in context.Config.Plugins
	if config, exists := context.Config.Plugins[name]; exists {
		if exists {
			err = plugin.Initialize(config)
		} else {
			err = plugin.Initialize(make(map[string]string))
		}
		if err != nil {
			return errors.New("Failed to initialize data plugin: " + name + " - " + err.Error())
		}
	}

	return nil
}

func GetContentTypePluginByExtension(mgr *PluginManager, ext string) (ContentTypePlugin, bool) {
	ext = strings.ToLower(ext)
	if plugin, exists := mgr.contentTypeMap[ext]; exists {
		return plugin, true
	}
	return nil, false
}

func CreatePluginManager() (PluginManager, error) {
	var mgr PluginManager
	mgr.plugins = make(map[string]Plugin, 0)
	mgr.contentTypeMap = make(map[string]ContentTypePlugin, 0)

	return mgr, nil
}

func ShutdownPlugins(mgr *PluginManager) {
	for _, plugin := range mgr.plugins {
		if dataPlugin, ok := plugin.(DataPlugin); ok {
			dataPlugin.Shutdown()
		}
	}
	mgr.plugins = nil
	mgr.contentTypeMap = nil
}
