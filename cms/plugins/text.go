package plugins

import (
	"cms/core"
	"path"
	"strings"
)

type BuiltinTextPlugin struct{}

func (p *BuiltinTextPlugin) Name() string {
	return "builtin/text"
}

func (p *BuiltinTextPlugin) Priority() int {
	return 100
}

func (p *BuiltinTextPlugin) CanProcess(file *core.File) bool {
	return strings.HasSuffix(strings.ToLower(file.Name), ".txt")
}

func (p *BuiltinTextPlugin) Process(ctx *core.PluginContext) *core.PluginResult {
	content := ctx.File.ReadFile(ctx.SiteDirectory)
	if content == nil {
		return &core.PluginResult{
			Success: false,
		}
	}

	route := strings.TrimPrefix(ctx.File.Path, "content/")
	route = "/" + strings.TrimLeft(route, "/")
	route = path.Clean(route)

	return &core.PluginResult{
		Success:    true,
		MimeType:   "text/plain; charset=utf-8",
		NewContent: content,
		Routes:     []string{route},
	}
}
