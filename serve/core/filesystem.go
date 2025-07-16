package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/goccy/go-yaml"
)

type File struct {
	LocalPath        string
	Url              string
	Title            string    `yaml:"title"`
	Author           string    `yaml:"author"`
	Tags             []string  `yaml:"tags"`
	ImagePath        string    `yaml:"image"`
	DateOfLastUpdate time.Time `yaml:"date-of-last-update"`
	CssFile          string
	MimeType         string
	CachedContent    []byte
	RedirectUrl      string     `yaml:"redirect-url"`
	IgnoreLayout     bool       `yaml:"ignore-layout"`
	Directory        *Directory // The directory this file belongs to
}

type Directory struct {
	LocalPath      string
	Url            string
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
		Title:        strings.TrimSuffix(strings.TrimSuffix(fileName, ext), "."),
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

	// Set the date of the last update if not set
	if file.DateOfLastUpdate.IsZero() {
		// get the file's modification time
		info, err := os.Stat(file.LocalPath)
		if err != nil {
			log.Printf("failed to get file info for %s: %s", file.LocalPath, err)
		} else {
			file.DateOfLastUpdate = info.ModTime()
		}
	}

	return file, nil
}

func createDirectoryStruct(localPath string, context *Context) (Directory, error) {
	directory := Directory{
		LocalPath: localPath,
		Title:     filepath.Base(localPath),
	}

	// Construct the path to metadata.yaml
	metadataPath := filepath.Join(localPath, "metadata.yaml")

	// Read and parse metadata.yaml - this file is optional!
	metadataFile, err := os.Open(metadataPath)
	if err == nil {
		defer metadataFile.Close()
		decoder := yaml.NewDecoder(metadataFile)
		decoder.Decode(&directory) // Ignore any errors here, as the file is optional
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
			subDir, err := createDirectoryStruct(subDirPath, context)
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

func populateFilesystem(directory *Directory, url string, context *Context) {
	// Create a lookup item for all files in the current directory
	for _, file := range directory.Files {
		base := filepath.Base(file.LocalPath)
		ext := strings.ToLower(filepath.Ext(base))
		base = strings.TrimSuffix(base, ext)

		// If this is the index file then use it as a default route for the directory
		if base == "index" {
			file.Url = url
			context.Filesystem[url] = file // (e.g. "/doc")
			// Create another alias with a trailing slash ("/doc/")
			if url != "/" {
				context.Filesystem[url+"/"] = file
			}
			// Create another alias for the index ("/doc/index")
			context.Filesystem[filepath.Join(url, "index")] = file
		} else {
			file.Url = filepath.Join(url, base)
			context.Filesystem[file.Url] = file
		}
	}

	// Recursively populate the lookup index for child directories
	for i, dir := range directory.Subdirectories {
		base := filepath.Base(dir.LocalPath)
		dir.Url = filepath.Join(url, base)
		populateFilesystem(&dir, dir.Url, context)
		directory.Subdirectories[i] = dir
	}
}

func InitializeFilesystem(context *Context) error {
	var err error
	contentRoot := filepath.Join(context.Config.SiteDirectory, "content")

	// Read the directory for this item
	context.Root, err = createDirectoryStruct(contentRoot, context)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", contentRoot, err)
	}

	// Create structures for all Files in the Directory
	context.Root.Url = "/"
	populateFilesystem(&context.Root, context.Root.Url, context)

	return nil
}
