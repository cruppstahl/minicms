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

func fetchFileBodies(file *File, context *Context) ([]string, error) {
	bytes, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return nil, err
	}

	ret := make([]string, 0, 3)
	if file.IgnoreLayout {
		ret = append(ret, string(bytes))
		return ret, nil
	}

	header, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/header.html")
	ret = append(ret, string(header))

	// Skip any frontmatter - it was already parsed into the File struct
	rest, err := frontmatter.Parse(strings.NewReader(string(bytes)), &file)
	if err == nil {
		ret = append(ret, string(rest))
	} else {
		ret = append(ret, string(bytes))
	}

	footer, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/footer.html")
	ret = append(ret, string(footer))

	return ret, nil
}

func buildTemplateVars(file *File, context *Context) map[string]interface{} {
	vars := map[string]interface{}{
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
	}

	if file.Directory != nil { // Can be nil when "dump"ing everything to disk
		vars["Directory"] = map[string]interface{}{
			"Title":   file.Directory.Title,
			"CssFile": file.Directory.CssFile,
		}
	}

	// Go through all NavigationItems. If their URL matches the current file's URL,
	// then set the "active" variable to true.
	root := context.Navigation.Root
	for i, item := range root.Children {
		item.IsActive = file.Url == item.Url
		root.Children[i] = item
	}
	vars["Navigation"] = root

	return vars
}

func applyTemplate(body string, file *File, vars *map[string]interface{}) (string, error) {
	tmpl, err := template.New(file.LocalPath).Parse(body)
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.LocalPath, err)
		return "", err
	}

	var output strings.Builder
	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.LocalPath, err)
		return "", err
	}

	return output.String(), nil
}

func GetFileWithContent(path string, context *Context) (*File, error) {
	normalizedPath := normalizePath(path)

	file, ok := context.Filesystem[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", normalizedPath)
	}

	// If the file is cached then use the cache
	if len(file.CachedContent) > 0 {
		return &file, nil
	}

	// Read all files that are required to build the final output
	bodies, err := fetchFileBodies(&file, context)
	if err != nil {
		return nil, err
	}

	// Build the structure with the template variables
	vars := buildTemplateVars(&file, context)

	// Apply the template to the different files
	for i, body := range bodies {
		body, err = applyTemplate(body, &file, &vars)
		if err != nil {
			log.Printf("failed to apply template for %s: %s", file.LocalPath, err)
			continue // Ignore errors and continue
		}
		bodies[i] = body
	}

	// Apply the content type plugin to the content file (but not to header and footer!)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.LocalPath), "."))
	plugin, exists := GetContentTypePluginByExtension(&context.PluginManager, ext)
	if exists {
		var i int
		if file.IgnoreLayout {
			i = 0
		} else {
			i = 1 // Skip header
		}

		body, err := plugin.Convert(bodies[i])
		if err != nil {
			// Print error, but then continue as usual
			log.Printf("failed to convert content for %s using plugin %s: %s",
				file.LocalPath, plugin.Name(), err)
		}
		bodies[i] = body
	}

	// Concatenate all bodies into one
	file.CachedContent = []byte(strings.Join(bodies, "\n"))
	return &file, nil

}
