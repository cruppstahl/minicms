package impl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Dump(context *Context) {
	ctxcopy := *context
	outDir := ctxcopy.Config.OutDirectory
	err := os.Mkdir(outDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory %s: %v", outDir, err)
	}

	// LookupIndex has circular references which break the JSON serializer. Remove them,
	// and remove other unsupported types
	ctxcopy.Watcher = nil
	for url, file := range ctxcopy.Navigation.Filesystem {
		file.Directory = nil
		ctxcopy.Navigation.Filesystem[url] = file
	}

	contextJson, err := json.MarshalIndent(ctxcopy, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal context: %v", err)
	}

	outPath := filepath.Join(outDir, "context.json")
	err = os.WriteFile(outPath, contextJson, 0644)
	if err != nil {
		log.Fatalf("Failed to write %s: %v", outPath, err)
	}

	// For each route: create the file
	for url := range ctxcopy.Navigation.Filesystem {
		file, err := GetFileWithContent(url, &ctxcopy)
		if err != nil {
			log.Fatalf("Failed to retrieve file contents: %v", err)
		}
		// split url in path and file name
		path := filepath.Dir(url)
		base := filepath.Base(file.LocalPath)

		// Create the metadata for the file
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatalf("Failed to mkdir %s: %v", path, err)
		}

		// write the metadata
		metadata := fmt.Sprintf("LocalPath: %s\n", file.LocalPath)
		metadata += fmt.Sprintf("Title: %s\n", file.Title)
		metadata += fmt.Sprintf("Author: %s\n", file.Author)
		metadata += fmt.Sprintf("Tags: [%s]\n", strings.Join(file.Tags, ", "))
		metadata += fmt.Sprintf("ImagePath: %s\n", file.ImagePath)
		metadata += fmt.Sprintf("CssFile: %s\n", file.CssFile)
		metadata += fmt.Sprintf("MimeType: %s\n", file.MimeType)

		if file.Directory != nil {
			metadata += fmt.Sprintf("Directory.CssFile: %s\n", file.Directory.CssFile)
			metadata += fmt.Sprintf("Directory.title: %s\n", file.Directory.Title)
		}

		outPath = filepath.Join(outDir, path, base) + ".metadata"
		err = os.WriteFile(outPath, []byte(metadata), 0644)
		if err != nil {
			log.Fatalf("Failed to create %s: %v", outPath, err)
		}

		// Write the cached file content
		outPath = filepath.Join(outDir, path, base)
		err = os.WriteFile(outPath, file.CachedContent, 0644)
		if err != nil {
			log.Fatalf("Failed to create %s: %v", outPath, err)
		}
	}
}
