package plugins

import (
	"bytes"
	"cms/core"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type BuiltinMarkdownPlugin struct {
	markdown goldmark.Markdown
	Context  *core.Context
}

func NewMarkdownPlugin(ctx *core.Context) *BuiltinMarkdownPlugin {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"), // or any Chroma style
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true), // optional line numbers
				),
			),
		),
	)
	return &BuiltinMarkdownPlugin{markdown: markdown, Context: ctx}
}

func (p *BuiltinMarkdownPlugin) Name() string {
	return "builtin/markdown"
}

func (p *BuiltinMarkdownPlugin) Priority() int {
	return 100
}

func (p *BuiltinMarkdownPlugin) CanProcess(file *core.File) bool {
	// Ignore files in the layout directory
	if strings.HasPrefix(file.Path, "layout/") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(file.Name), ".md") ||
		strings.HasSuffix(strings.ToLower(file.Name), ".markdown")
}

func (p *BuiltinMarkdownPlugin) Process(ctx *core.PluginContext) *core.PluginResult {
	log.Printf("START Processing markdown file: %s\n", ctx.File.Path)
	defer log.Printf("END   Processing markdown file: %s\n", ctx.File.Path)

	// Don't attempt to read a file if it is only a redirection
	if ctx.File.Metadata.RedirectUrl != "" {
		log.Printf("Error: RedirectUrl should not be set for Markdown files: %s", ctx.File.Path)
		return &core.PluginResult{
			Success: false,
			Error:   fmt.Errorf("RedirectUrl should not be set for Markdown files"),
		}
	}

	content := ctx.File.ReadFile(ctx.SiteDirectory)
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

	var body []byte
	var html bytes.Buffer
	if err := p.markdown.Convert(content, &html); err != nil {
		return &core.PluginResult{
			Success: false,
		}
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

		body = append(header.Content, html.Bytes()...)
		body = append(body, footer.Content...)
	} else {
		// If the layout is ignored, we still need to read the file content
		body = html.Bytes()
	}

	// A markdown file has two routes: the path itself, with ".html" extension, and the path without
	// the extension (e.g. "/about.md" becomes "/about.html" and "/about")
	// If this file is an index page then we also add the directory name as a route
	route := strings.TrimPrefix(ctx.File.Path, "content/")
	route = "/" + strings.TrimLeft(route, "/")
	route = path.Clean(route)
	result.Routes = []string{route,
		strings.TrimSuffix(route, filepath.Ext(route))}
	if filepath.Base(route) == "index.md" {
		// If this is an index page, add the directory name as a route
		dir := filepath.Dir(route)
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
