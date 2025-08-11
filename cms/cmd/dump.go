package cmd

import (
	"cms/core"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Dump(ctx *core.Context, everything bool) {
	ctxcopy := *ctx
	outDir := ctxcopy.Config.OutDirectory
	err := os.Mkdir(outDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory %s: %v", outDir, err)
	}

	// For each route: create the file
	for url, file := range ctxcopy.FileManager.GetAllFiles() {
		// split url in path and file name
		path := filepath.Join(outDir, filepath.Dir(url))
		base := filepath.Base(file.Path)

		if everything {
			// Create the metadata for the file
			err = os.MkdirAll(path, 0755)
			if err != nil {
				log.Fatalf("Failed to mkdir %s: %v", path, err)
			}

			// write the metadata
			metadata := fmt.Sprintf("Path: %s\n", file.Path)
			metadata += fmt.Sprintf("Title: %s\n", file.Metadata.Title)
			metadata += fmt.Sprintf("Author: %s\n", file.Metadata.Author)
			metadata += fmt.Sprintf("Tags: [%s]\n", strings.Join(file.Metadata.Tags, ", "))
			metadata += fmt.Sprintf("MimeType: %s\n", file.Metadata.MimeType)
			metadata += fmt.Sprintf("IgnoreLayout: %t\n", file.Metadata.IgnoreLayout)
			metadata += fmt.Sprintf("RedirectUrl: %s\n", file.Metadata.RedirectUrl)

			if file.Parent != nil {
				metadata += fmt.Sprintf("Directory.CssFile: %s\n", file.Parent.Metadata.CssFile)
				metadata += fmt.Sprintf("Directory.Title: %s\n", file.Parent.Metadata.Title)
			}

			outPath := filepath.Join(path, base) + ".yaml"
			err = os.WriteFile(outPath, []byte(metadata), 0644)
			if err != nil {
				log.Fatalf("Failed to create %s: %v", outPath, err)
			}
		}

		// Write the cached file content
		outPath := filepath.Join(path, base)
		err = os.WriteFile(outPath, file.Content, 0644)
		if err != nil {
			log.Fatalf("Failed to create %s: %v", outPath, err)
		}
	}

	if everything {
		// Filesystem has circular references which break the JSON serializer. Remove them,
		// and remove other unsupported types
		ctxcopy.FileWatcher = nil
		for _, file := range ctxcopy.FileManager.GetAllFiles() {
			file.Parent = nil
			file.Dependencies = nil
			file.Dependents = nil
			file.Content = nil
		}

		contextJson, err := json.MarshalIndent(&ctxcopy, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal context: %v", err)
		}

		outPath := filepath.Join(outDir, "context.json")
		err = os.WriteFile(outPath, contextJson, 0644)
		if err != nil {
			log.Fatalf("Failed to write %s: %v", outPath, err)
		}
	}
}
