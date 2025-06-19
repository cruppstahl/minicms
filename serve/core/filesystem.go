package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/goccy/go-yaml"
)

type File struct {
	LocalPath     string
	Title         string   `yaml:"title"`
	Author        string   `yaml:"author"`
	Tags          []string `yaml:"tags"`
	ImagePath     string   `yaml:"image"`
	CssFile       string
	MimeType      string
	CachedContent []byte
	IgnoreLayout  bool       `yaml:"ignore-layout"`
	Directory     *Directory // The directory this file belongs to
}

type Directory struct {
	LocalPath      string
	Title          string `yaml:"title"`
	CssFile        string `yaml:"cssfile"`
	Subdirectories map[string]Directory
	Files          map[string]File
}

func createFileStruct(filePath string, fileName string, directory *Directory, plugin ContentTypePlugin) (File, error) {
	ext := strings.TrimLeft(strings.ToLower(filepath.Ext(fileName)), ".")

	// Create a File struct and populate its fields
	file := File{
		LocalPath:    filePath,
		Title:        strings.TrimSuffix(fileName, ext),
		MimeType:     plugin.Mimetype(),
		IgnoreLayout: plugin.IgnoreLayout(),
		Directory:    directory,
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
			return file, err
		}
	}

	// Read the file and extract any frontmatter, if available
	body, err := os.ReadFile(file.LocalPath)
	if err != nil {
		log.Printf("failed to read file content for %s: %s", file.LocalPath, err)
		return file, err
	}

	// Simply parse the frontmatter from the file content, and ignore any errors
	frontmatter.Parse(strings.NewReader(string(body)), &file)
	return file, nil
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
			fileName := entry.Name()
			filePath := filepath.Join(localPath, fileName)

			// Ignore the file if the PluginManager does not have a plugin for this file type
			ext := strings.TrimLeft(strings.ToLower(filepath.Ext(fileName)), ".")
			plugin, exists := GetContentTypePluginByExtension(&context.PluginManager, ext)
			if !exists {
				continue
			}

			file, err := createFileStruct(filePath, fileName, &directory, plugin)
			if err != nil {
				log.Printf("Failed to create file struct for %s: %v", filePath, err)
				continue
			}

			// Append the file to the directory's Files slice
			base := filepath.Base(filePath)
			directory.Files[strings.TrimSuffix(base, ext)] = file
		}
	}

	return directory, nil
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
