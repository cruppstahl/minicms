package internal

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type File struct {
	Path      string
	Title     string
	Author    string
	Tags      []string
	ImagePath string
	CssFile   string
	Format    string
	Content   string
}

type Directory struct {
	Path        string
	Title       string `yaml:"title"`
	CssFile     string `yaml:"cssfile"`
	Directories []Directory
	Files       []File
}

type DataTree struct {
	Root      string
	Directory Directory
}

func readDirectory(path string, context *Context) (Directory, error) {
	var directory Directory
	directory.Path = path

	// Construct the path to metadata.yaml
	metadataPath := filepath.Join(path, "metadata.yaml")

	// Read and parse metadata.yaml - this file is optional!
	metadataFile, err := os.Open(metadataPath)
	if err != nil {
		// assume that the file does not exist, fill struct with default values
		directory.Title = filepath.Base(path)
	} else {
		defer metadataFile.Close()
		decoder := yaml.NewDecoder(metadataFile)
		if err := decoder.Decode(&directory); err != nil {
			log.Printf("Failed to read %s: %v", metadataPath, err)
			// fall through
		}
	}

	// Open the directory
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}

	directory.Directories = []Directory{}
	directory.Files = []File{}

	// Iterate over the directory entries
	for _, entry := range dirEntries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue // Skip hidden directories
			}

			// Set the directory path
			subDirPath := filepath.Join(path, entry.Name())

			// Recursively read the subdirectory
			subDir, err := readDirectory(subDirPath, context)
			if err != nil {
				log.Printf("Failed to read subdirectory %s: %v", subDirPath, err)
				continue
			}
			directory.Directories = append(directory.Directories, subDir)
		} else {
			// Ignore file unless the extension is ".md", ".txt", or ".html"
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".md" && ext != ".txt" && ext != ".html" {
				continue
			}

			// Set the file path
			filePath := filepath.Join(path, entry.Name())

			// Create a File struct and populate its fields
			file := File{
				Path:   filePath,
				Title:  strings.TrimSuffix(entry.Name(), ext),
				Format: strings.TrimLeft(ext, "."),
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
			directory.Files = append(directory.Files, file)
		}
	}

	return directory, nil
}

func addRoute(router *gin.Engine, directory *Directory, level int, context *Context) {
	// Only create routes for /blog, /docs, /shop etc. directories,
	// not for the root directory or nested subdirectories
	if level == 1 {
		// Create a route based on the file's path
		routePath := strings.TrimPrefix(directory.Path, context.Config.SiteDirectory+"/content/")

		// Define the handler function for this route
		handlerFunc := func(c *gin.Context) {
			c.JSON(200, gin.H{
				"title": directory.Title,
			})
		}
		router.GET("/"+routePath, handlerFunc)
	}

	// Go through each file in the directory and add a route for it
	for _, file := range directory.Files {
		// Create a route for the file
		fileRoutePath := strings.TrimPrefix(file.Path, context.Config.SiteDirectory+"/content/")
		fileRoutePath = strings.TrimSuffix(fileRoutePath, filepath.Ext(fileRoutePath)) // Remove the file extension
		fileRoutePath = strings.ReplaceAll(fileRoutePath, "\\", "/")                   // Ensure forward slashes for URLs
		router.GET("/"+fileRoutePath, func(c *gin.Context) {
			// Handler function for the file route
			c.JSON(200, gin.H{
				"path":    c.Request.URL.Path,
				"title":   file.Title,
				"author":  file.Author,
				"tags":    file.Tags,
				"format":  file.Format,
				"content": file.Content,
			})
		})
	}

	// Go through each subdirectory, call this function recursively
	for _, subDir := range directory.Directories {
		addRoute(router, &subDir, level+1, context)
	}
}

func ReadDataTree(context *Context) error {
	var root = context.Config.SiteDirectory
	var err error
	context.DataTree.Root = root

	// Read the directory and populate the data tree
	context.DataTree.Directory, err = readDirectory(root+"/content", context)
	return err
}

func SetupRoutes(router *gin.Engine, context *Context) error {
	err := ReadDataTree(context)
	if err != nil {
		return err
	}

	// Walk through the data tree and set up all the routes
	addRoute(router, &context.DataTree.Directory, 0, context)
	// router.GET("/", handlers.GetRoot)
	// router.GET("/posts", handlers.GetPosts)
	// router.GET("/posts/index", handlers.GetPosts)
	// router.GET("/posts/:post_id", handlers.GetPostByID)

	return nil
}
