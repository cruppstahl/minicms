package plugins

import (
	"cms/core"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func ApplyTemplate(body []byte, file *core.File, vars *map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New(file.Path).Parse(string(body))
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.Path, err)
		return nil, err
	}

	var output strings.Builder
	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.Path, err)
		return nil, err
	}

	return []byte(output.String()), nil
}

func BuildTemplateVars(ctx *core.Context, file *core.File, routes []string) map[string]any {
	vars := map[string]any{
		"SiteTitle":        ctx.Config.Server.Title,
		"SiteDescription":  ctx.Config.Server.Description,
		"Site.Author.Name": ctx.Users.Users[0].Name, // Assuming at least one user exists
		"BrandingFavicon":  ctx.Config.Branding.Favicon,
		"BrandingCssFile":  ctx.Config.Branding.CssFile,
		"PageTitle":        file.Metadata.Title,
		"PageAuthor":       file.Metadata.Author,
		"PageTags":         file.Metadata.Tags,
		"PageCssFile":      file.Metadata.CssFile,
		"PageMimeType":     file.Metadata.MimeType,
	}

	// Date of last modTime is either specified in the metadata or is fetched from the file system
	if file.Metadata.DateOfLastUpdate.IsZero() {
		info, err := os.Stat(filepath.Join(ctx.Config.SiteDirectory, file.Path))
		if err != nil {
			log.Printf("failed to get file info for %s: %s", file.Path, err)
		} else {
			vars["DateOfLastUpdate"] = info.ModTime()
		}
	} else {
		vars["DateOfLastUpdate"] = file.Metadata.DateOfLastUpdate
	}

	if file.Parent != nil { // Can be nil when "dump"ing everything to disk
		vars["Directory"] = map[string]interface{}{
			"Title":   file.Parent.Metadata.Title,
			"CssFile": file.Parent.Metadata.CssFile,
		}
	}

	// Go through all NavigationItems. If their URL matches the current file's URL,
	// then set the "active" variable to true.
	nav := ctx.Navigation
	for i, item := range nav.Children {
		lcurl := strings.ToLower(item.Url)
		item.IsActive = slices.Contains(routes, lcurl)
		nav.Children[i] = item
	}
	vars["Navigation"] = nav

	return vars
}
