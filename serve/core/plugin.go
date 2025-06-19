package core

import (
	"errors"
	"strings"
)

type ContentTypePlugin interface {
	Name() string
	Id() string
	Version() string
	Mimetype() string
	IgnoreLayout() bool
	Extensions() []string

	Convert(context *Context, file *File) error
}

type PluginManager struct {
	contentTypePlugins map[string]ContentTypePlugin
	fileExtensions     map[string]ContentTypePlugin
}

var ErrPluginAlreadyRegistered = errors.New("Plugin already registered")
var ErrExtensionAlreadyRegistered = errors.New("File extension already registered")

func RegisterContentTypePlugin(manager *PluginManager, plugin ContentTypePlugin) error {
	id := plugin.Id()
	if _, exists := manager.contentTypePlugins[id]; exists {
		return ErrPluginAlreadyRegistered
	}
	manager.contentTypePlugins[id] = plugin

	for _, ext := range plugin.Extensions() {
		ext = strings.ToLower(ext)
		if _, exists := manager.fileExtensions[ext]; exists {
			return ErrExtensionAlreadyRegistered
		}
		manager.fileExtensions[ext] = plugin
	}

	return nil
}

func GetContentTypePluginByExtension(mgr *PluginManager, ext string) (ContentTypePlugin, bool) {
	ext = strings.ToLower(ext)
	if plugin, exists := mgr.fileExtensions[ext]; exists {
		return plugin, true
	}
	return nil, false
}

func CreatePluginManager() (PluginManager, error) {
	var mgr PluginManager
	mgr.contentTypePlugins = make(map[string]ContentTypePlugin, 0)
	mgr.fileExtensions = make(map[string]ContentTypePlugin, 0)

	return mgr, nil
}
