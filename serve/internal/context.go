package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Context struct {
	Users     Users
	Config    Config
	DataTree  DataTree
	DataCache map[string]File
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

func fetchFileContent(file *File) error {
	// Read content from file.LocalPath and store it in file.Content
	content, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return err
	}
	file.Content = string(content)

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

		err = fetchFileContent(&file)
		if err != nil {
			return File{}, err
		}

		context.DataCache[normalizedPath] = file
	}

	return file, nil
}
