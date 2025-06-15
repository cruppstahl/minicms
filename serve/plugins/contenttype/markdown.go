package contenttype

import (
	"serve/core"
)

type ContentTypeMarkdownPlugin struct {
	core.ContentTypePlugin
}

func NewMarkdownPlugin() core.ContentTypePlugin {
	return ContentTypeMarkdownPlugin{}
}

func (ContentTypeMarkdownPlugin) Name() string { return "builtin/markdown" }
func (ContentTypeMarkdownPlugin) Version() string { return "0.1" }
func (ContentTypeMarkdownPlugin) Mimetype() string { return "text/html" }
func (ContentTypeMarkdownPlugin) Id() string { return "2E9C1AB9-2D58-4BB5-989F-8269C6D2007A" }
func (ContentTypeMarkdownPlugin) Description() string { return "Renders markdown files" }
func (ContentTypeMarkdownPlugin) Extensions() []string { return []string{"md", "markdown"} }

func (ContentTypeMarkdownPlugin) Convert(context *core.Context, file *core.File) error {
	// TODO
	return nil
}
