package plugins

import (
	"serve/core"
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

	return &core.PluginResult{
		Success:    true,
		MimeType:   "text/plain",
		NewContent: content,
		Routes:     []string{route},
	}
}
