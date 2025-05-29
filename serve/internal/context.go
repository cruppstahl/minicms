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
	DataTree   DataTree
	DataCache  map[string]File
	Navigation Navigation
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
	// Normalize the path by removing leading and trailing slashes
	// and converting backslashes to forward slashes
	normalized := path
	if len(normalized) > 0 && normalized[0] == '/' {
		normalized = normalized[1:]
	}
	if len(normalized) > 0 && normalized[len(normalized)-1] == '/' {
		normalized = normalized[:len(normalized)-1]
	}

	// Also convert double slashes to single slashes
	return strings.ReplaceAll(normalized, "//", "/")
}

func fetchMetadata(path string, dataTree DataTree) (File, error) {
	// Split normalizedPath into directories and filename, and look them up
	// in the DataTree
	directory := dataTree.Directory
	dirs := strings.Split(path, "/")
	for _, dir := range dirs[:len(dirs)-1] {
		if subDir, ok := directory.Directories[dir]; ok {
			directory = subDir
		} else {
			return File{}, fmt.Errorf("Directory not found: %s", dir)
		}
	}
	filename := dirs[len(dirs)-1]
	// Check if the file exists in the directory
	if fileData, ok := directory.Files[filename]; ok {
		// If the file exists, use its data
		return fileData, nil
	} else {
		return File{}, fmt.Errorf("File not found: %s", filename)
	}
}

func fetchFileContent(file *File, context *Context) error {
	header, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/header.html")
	footer, _ := os.ReadFile(context.Config.SiteDirectory + "/layout/footer.html")

	// Read content from file.LocalPath and store it in file.Content
	body, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return err
	}
	file.Content = string(header) + string(body) + string(footer)

	return nil
}

func applyTemplate(file *File, context *Context) error {
	tmpl, err := template.New(file.LocalPath).Parse(file.Content)
	if err != nil {
		log.Printf("failed to parse template for %s: %s", file.LocalPath, err)
		return err
	}
	var output strings.Builder
	var vars = map[string]interface{}{
		"Title":       context.Config.Server.Title,
		"Description": context.Config.Server.Description,
		"Author":      context.Users.Users[0].Name, // Assuming at least one user exists
	}
	err = tmpl.Execute(&output, vars)
	if err != nil {
		log.Printf("failed to execute template for %s: %s", file.LocalPath, err)
		return err
	}

	file.Content = output.String()
	return nil
}

func GetFile(path string, context *Context) (File, error) {
	normalizedPath := normalizePath(path)
	// If the file is in the cache then fetch and build it
	var file File
	var exists bool
	var err error
	if file, exists = context.DataCache[normalizedPath]; !exists {
		file, err = fetchMetadata(normalizedPath, context.DataTree)
		if err != nil {
			return File{}, err
		}

		err = fetchFileContent(&file, context)
		if err != nil {
			return File{}, err
		}

		// ignore errors, just display the template as is if it cannot be applied
		_ = applyTemplate(&file, context)

		context.DataCache[normalizedPath] = file
	}

	return file, nil
}
