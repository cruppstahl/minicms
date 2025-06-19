package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/adrg/frontmatter"
)

func normalizePath(path string) string {
	// Convert double slashes to single slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// Also convert double slashes to single slashes
	return strings.ReplaceAll(path, "//", "/")
}

func fetchFileBody(file *File, context *Context) (string, error) {
	bytes, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return "", err
	}
	body := string(bytes)

	if file.IgnoreLayout {
		return body, nil
	}

	// Skip any frontmatter - it was already parsed into the File struct
	rest, err := frontmatter.Parse(strings.NewReader(string(body)), &file)
	if err == nil {
		body = string(rest)
	}

	header, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/header.html")
	footer, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/footer.html")
	return string(header) + body + string(footer), nil
}

func applyTemplate(body string, file *File, context *Context) (string, error) {
	tmpl, err := template.New(file.LocalPath).Parse(body)
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.LocalPath, err)
		return "", err
	}
	var output strings.Builder
	var vars = map[string]interface{}{
		"SiteTitle":        context.Config.Server.Title,
		"SiteDescription":  context.Config.Server.Description,
		"Site.Author.Name": context.Users.Users[0].Name, // Assuming at least one user exists
		"BrandingFavicon":  context.Config.Branding.Favicon,
		"BrandingCssFile":  context.Config.Branding.CssFile,
		"PageTitle":        file.Title,
		"PageAuthor":       file.Author,
		"PageTags":         file.Tags,
		"PageImagePath":    file.ImagePath,
		"PageCssFile":      file.CssFile,
		"PageMimeType":     file.MimeType,
		"ActiveUrl":        "", // This will be set below
	}

	if file.Directory != nil { // Can be nil when "dump"ing everything to disk
		vars["Directory.Title"] = file.Directory.Title
	}

	// Go through all NavigationItems. If their URL matches the current file's URL,
	// then set the "active" variable to true.
	navTree := context.Navigation.NavigationTree
	for i, item := range navTree {
		item.IsActive = strings.HasSuffix(file.LocalPath, item.LocalPath)
		navTree[i] = item
	}
	vars["Navigation"] = navTree

	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.LocalPath, err)
		return "", err
	}

	return output.String(), nil
}

func GetFileWithContent(path string, context *Context) (*File, error) {
	normalizedPath := normalizePath(path)

	file, ok := context.Navigation.Filesystem[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", normalizedPath)
	}

	// If the file is not cached then build it
	var body string
	var err error
	if len(file.CachedContent) == 0 {
		body, err = fetchFileBody(&file, context)
		if err != nil {
			return nil, err
		}

		// ignore errors, just display the template as is if it cannot be applied
		body, err = applyTemplate(body, &file, context)
		if err == nil {
			file.CachedContent = []byte(body)
			context.Navigation.Filesystem[normalizedPath] = file
		}

		// now render the template
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.LocalPath), "."))
		plugin, _ := GetContentTypePluginByExtension(&context.PluginManager, ext)
		plugin.Convert(context, &file)
	}

	return &file, nil
}
