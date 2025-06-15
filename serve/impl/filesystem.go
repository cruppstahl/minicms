package impl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type File struct {
	LocalPath     string
	Title         string
	Author        string
	Tags          []string
	ImagePath     string
	CssFile       string
	MimeType      string
	CachedContent []byte
	Directory     *Directory // The directory this file belongs to
}

type Directory struct {
	LocalPath      string
	Title          string `yaml:"title"`
	CssFile        string `yaml:"cssfile"`
	Subdirectories map[string]Directory
	Files          map[string]File
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

	directory.Subdirectories = make(map[string]Directory)
	directory.Files = make(map[string]File)

	// Iterate over the directory entries
	for _, entry := range dirEntries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue // Skip hidden files and directories
		}

		if entry.IsDir() {
			// Set the directory path
			subDirPath := filepath.Join(localPath, entry.Name())

			// Recursively read the subdirectory
			subDir, err := readDirectory(subDirPath, context)
			if err != nil {
				log.Printf("Failed to read subdirectory %s: %v", subDirPath, err)
				continue
			}
			directory.Subdirectories[entry.Name()] = subDir
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
				Directory: &directory,
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
		return "text/html" // Same about text files
	case "html":
		return "text/html"
	default:
		return "application/octet-stream" // Default MIME type for unknown files
	}
}

func addFilesystemEntry(context *Context, url string, file File) {
	// Add the File to the Filesystem
	_, exists := context.Navigation.Filesystem[url]
	if exists {
		log.Fatalf("Duplicate URL found in Filesystem: %s", url)
	}
	context.Navigation.Filesystem[url] = file
}

func populateFilesystem(directory *Directory, url string, context *Context) {
	// Create a lookup item for all files in the current directory
	for _, file := range directory.Files {
		// Create a File structure
		base := filepath.Base(file.LocalPath)
		ext := strings.ToLower(filepath.Ext(base))
		base = strings.TrimSuffix(base, ext)
		addFilesystemEntry(context, filepath.Join(url, base), file)

		// If this is the index file then use it as a default route for the directory
		if base == "index" {
			addFilesystemEntry(context, url, file)
		}
	}

	// Recursively populate the lookup index for child directories
	for _, subDir := range directory.Subdirectories {
		base := filepath.Base(subDir.LocalPath)
		populateFilesystem(&subDir, filepath.Join(url, base), context)
	}
}

func InitializeFilesystem(context *Context) error {
	var err error
	contentRoot := filepath.Join(context.Config.SiteDirectory, "content")

	// Read the directory for this item
	directory, err := readDirectory(contentRoot, context)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", contentRoot, err)
	}

	// Create structures for all Files in the Directory
	populateFilesystem(&directory, "/", context)

	return nil
}
