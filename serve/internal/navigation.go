package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type Navigation struct {
	FilePath string
	Main     []NavigationItem `yaml:"main"`
}

type NavigationItem struct {
	LocalPath string           `yaml:"local-path"`
	RoutePath string           `yaml:"url"`
	Children  []NavigationItem `yaml:"children,omitempty"`
	Directory Directory
}

type File struct {
	LocalPath string
	Title     string
	Author    string
	Tags      []string
	ImagePath string
	CssFile   string
	MimeType  string
	Content   string
}

type Directory struct {
	LocalPath   string
	Title       string `yaml:"title"`
	CssFile     string `yaml:"cssfile"`
	Directories map[string]Directory
	Files       map[string]File
}

type DataTree struct {
	Root      string
	Directory Directory
}

func readDirectory(localPath string, context *Context) (Directory, error) {
	var directory Directory
	directory.LocalPath = localPath

	// Construct the path to metadata.yaml
	metadataPath := filepath.Join(localPath, "metadata.yaml")

	// Read and parse metadata.yaml - this file is optional!
	metadataFile, err := os.Open(metadataPath)
	if err != nil {
		// assume that the file does not exist, fill struct with default values
		directory.Title = filepath.Base(localPath)
	} else {
		defer metadataFile.Close()
		decoder := yaml.NewDecoder(metadataFile)
		if err := decoder.Decode(&directory); err != nil {
			log.Printf("Failed to read %s: %v", metadataPath, err)
			// fall through
		}
	}

	// Open the directory
	dirEntries, err := os.ReadDir(localPath)
	if err != nil {
		return Directory{}, err
	}

	directory.Directories = make(map[string]Directory)
	directory.Files = make(map[string]File)

	// Iterate over the directory entries
	for _, entry := range dirEntries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue // Skip hidden directories
			}

			// Set the directory path
			subDirPath := filepath.Join(localPath, entry.Name())

			// Recursively read the subdirectory
			subDir, err := readDirectory(subDirPath, context)
			if err != nil {
				log.Printf("Failed to read subdirectory %s: %v", subDirPath, err)
				continue
			}
			directory.Directories[entry.Name()] = subDir
		} else {
			// Ignore file unless the extension is ".md", ".txt", or ".html"
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".md" && ext != ".txt" && ext != ".html" {
				continue
			}

			// Set the file path
			filePath := filepath.Join(localPath, entry.Name())

			// Create a File struct and populate its fields
			file := File{
				LocalPath: filePath,
				Title:     strings.TrimSuffix(entry.Name(), ext),
				MimeType:  mimeType(strings.TrimLeft(ext, ".")),
			}

			// Check if the file has a corresponding .yaml file for metadata
			metadataFilePath := filePath + ".yaml"
			metadataFile, err := os.Open(metadataFilePath)
			if err == nil {
				defer metadataFile.Close()
				// Decode the metadata file into the File struct
				decoder := yaml.NewDecoder(metadataFile)
				if err := decoder.Decode(&file); err != nil {
					log.Printf("Failed to decode metadata for file %s: %v", metadataFilePath, err)
					continue
				}
			}

			// Append the file to the directory's Files slice
			base := filepath.Base(filePath)
			directory.Files[strings.TrimSuffix(base, ext)] = file
		}
	}

	return directory, nil
}

func mimeType(ext string) string {
	switch ext {
	case "md":
		return "text/html" // Markdown files are served as HTML
	case "txt":
		return "text/plain"
	case "html":
		return "text/html"
	default:
		return "application/octet-stream" // Default MIME type for unknown files
	}
}

func ReadDataTree(context *Context) (DataTree, error) {
	var root = context.Config.SiteDirectory
	var dataTree DataTree
	var err error
	dataTree.Root = root

	// Read the directory and populate the data tree
	dataTree.Directory, err = readDirectory(root+"/content", context)
	return dataTree, err
}

func ReadNavigationYaml(context Context, path string) (Context, error) {
	context.Navigation.FilePath = path

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Context{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse the YAML file
	if err := yaml.Unmarshal(data, &context.Navigation); err != nil {
		return Context{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// We need at least one main navigation item
	if len(context.Navigation.Main) == 0 {
		return Context{}, fmt.Errorf("no main navigation items found in %s", path)
	}

	// Go through each main navigation item and populate its Directory field
	for i, item := range context.Navigation.Main {
		// Set the LocalPath for the item
		localPath := filepath.Join(context.Config.SiteDirectory, "content", item.LocalPath)
		// Read the directory for this item
		item.Directory, err = readDirectory(localPath, &context)
		if err != nil {
			return Context{}, fmt.Errorf("failed to read directory for navigation item %s: %w", item.LocalPath, err)
		}
		// Update the item in the context
		context.Navigation.Main[i] = item
	}

	// Return the parsed data
	return context, nil
}
