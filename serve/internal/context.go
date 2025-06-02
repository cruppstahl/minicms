package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type Context struct {
	Users      Users
	Config     Config
	Navigation Navigation
	DataCache  map[string]File
}

func InitializeContext() (Context, error) {
	var err error
	var context Context
	context.DataCache = make(map[string]File)

	// parse command line arguments
	context.Config, err = ParseCommandLineArguments()
	if err != nil {
		return context, err
	}

	// read config.yaml
	configFilePath := fmt.Sprintf("%s/config/site.yaml", context.Config.SiteDirectory)
	context, err = ReadConfigYaml(context, configFilePath)
	if err != nil {
		return context, err
	}

	// read users.yaml
	authorsFilePath := fmt.Sprintf("%s/config/users.yaml", context.Config.SiteDirectory)
	context, err = ReadUsersYaml(context, authorsFilePath)
	if err != nil {
		return context, err
	}

	// read navigation.yaml
	navigationFilePath := fmt.Sprintf("%s/config/navigation.yaml", context.Config.SiteDirectory)
	context, err = ReadNavigationYaml(context, navigationFilePath)
	if err != nil {
		return context, err
	}

	// Return the parsed data
	return context, err
}

func normalizePath(path string) string {
	// Convert double slashes to single slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// Also convert double slashes to single slashes
	return strings.ReplaceAll(path, "//", "/")
}

func fetchFileBody(file *File, context *Context) (string, error) {
	header, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/header.html")
	footer, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/footer.html")

	// Read content from file.LocalPath and store it in file.Content
	body, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return "", err
	}
	return string(header) + string(body) + string(footer), nil
}

func applyTemplate(body string, file *File, context *Context) (string, error) {
	tmpl, err := template.New(file.LocalPath).Parse(body)
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.LocalPath, err)
		return "", err
	}
	var output strings.Builder
	var vars = map[string]interface{}{
		"Title":       context.Config.Server.Title,
		"Description": context.Config.Server.Description,
		"Author":      context.Users.Users[0].Name, // Assuming at least one user exists
		"Favicon":     context.Config.Branding.Favicon,
	}
	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.LocalPath, err)
		return "", err
	}

	return output.String(), nil
}

func GetFile(path string, context *Context) (*File, error) {
	normalizedPath := normalizePath(path)

	lookup, ok := context.Navigation.LookupIndex[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", normalizedPath)
	}

	// If the file is not cached then build it
	if lookup.File.CachedContent == "" {
		body, err := fetchFileBody(&lookup.File, context)
		if err != nil {
			return nil, err
		}

		// ignore errors, just display the template as is if it cannot be applied
		body, err = applyTemplate(body, &lookup.File, context)
		if err == nil {
			lookup.File.CachedContent = body
		}
	}

	return &lookup.File, nil
}
