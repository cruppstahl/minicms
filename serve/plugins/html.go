package plugins

import (
	"log"
	"path/filepath"
	"serve/core"
	"strings"

	"github.com/adrg/frontmatter"
)

type BuiltinHtmlPlugin struct {
	Context *core.Context
}

func (p *BuiltinHtmlPlugin) Name() string {
	return "builtin/html"
}

func (p *BuiltinHtmlPlugin) Priority() int {
	return 100
}

func (p *BuiltinHtmlPlugin) CanProcess(file *core.File) bool {
	// Ignore files in the layout directory
	if strings.HasPrefix(file.Path, "layout/") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(file.Name), ".html") ||
		strings.HasSuffix(strings.ToLower(file.Name), ".htm")
}

func (p *BuiltinHtmlPlugin) Process(ctx *core.PluginContext) *core.PluginResult {
	var body []byte
	var content []byte

	log.Printf("START Processing html file: %s\n", ctx.File.Path)
	defer log.Printf("END   Processing html file: %s\n", ctx.File.Path)

	// Don't attempt to read a file if it is only a redirection
	if ctx.File.Metadata.RedirectUrl != "" {
		panic("RedirectUrl should not be set for HTML files")
	}

	// Read file content
	content = ctx.File.ReadFile(ctx.SiteDirectory)
	if content == nil {
		return &core.PluginResult{
			Success: false,
		}
	}

	// Parse (and skip) frontmatter metadata
	rest, err := frontmatter.Parse(strings.NewReader(string(content)), &ctx.File.Metadata)
	if err == nil {
		content = rest
	}

	var result core.PluginResult

	// fetch the dependency files (header, footer) unless the layout is ignored
	if !ctx.File.Metadata.IgnoreLayout {
		header := ctx.FileManager.GetFile("layout/header.html")
		footer := ctx.FileManager.GetFile("layout/footer.html")
		if header == nil || footer == nil {
			return &core.PluginResult{
				Success: false,
			}
		}

		if header.Content == nil {
			header.Content = header.ReadFile(ctx.SiteDirectory)
		}
		if footer.Content == nil {
			footer.Content = footer.ReadFile(ctx.SiteDirectory)
		}
		if header.Content == nil || footer.Content == nil {
			return &core.PluginResult{
				Success: false,
			}
		}

		result.Dependencies = []*core.File{header, footer}

		body = append(header.Content, content...)
		body = append(body, footer.Content...)
	} else {
		// If the layout is ignored, we still need to read the file content
		body = content
	}

	// A html file has two routes: the path itself (without "/content") and the path without
	// the extension (e.g. "/about.html" becomes "/about")
	// If this file is an index page then we also add the directory name as a route
	filePath := strings.TrimPrefix(ctx.File.Path, "content/")
	result.Routes = []string{"/" + filePath,
		"/" + strings.TrimSuffix(filePath, filepath.Ext(filePath))}
	if filepath.Base(filePath) == "index.html" {
		// If this is an index page, add the directory name as a route
		dir := filepath.Dir(filePath)
		if dir == "." {
			dir = "/"
		}
		result.Routes = append(result.Routes, dir)
	}

	// Build the map with the template variables
	vars := BuildTemplateVars(p.Context, ctx.File, result.Routes)

	// Apply the template to the different files
	body, err = ApplyTemplate(body, ctx.File, &vars)
	if err != nil {
		log.Printf("failed to apply template for %s: %s", ctx.File.Path, err)
		return &core.PluginResult{
			Success: false,
		}
	}

	result.Success = true
	result.Modified = true
	result.NewContent = body
	result.MimeType = "text/html"
	return &result

}
