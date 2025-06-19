package core

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false // If we can't stat the file, assume it's not a file
	}
	return !info.IsDir() // Return true if it's not a directory
}

func InitializeFsWatcher(context *Context) error {
	var err error

	// Initialize the File System watcher
	context.Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to set up fsnotify: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-context.Watcher.Events:
				if !ok {
					log.Fatal("Watcher events channel closed")
				} else {
					log.Printf("File change detected: %s", event.Name)

					changedPath := strings.TrimPrefix(event.Name, context.Config.SiteDirectory)
					// Force recreation of all html files if the layout has changed
					if strings.HasPrefix(changedPath, "/layout") {
						for url, file := range context.Navigation.Filesystem {
							if file.MimeType == "text/html" {
								file.CachedContent = nil
								context.Navigation.Filesystem[url] = file
							}
						}
						break
					}

					changedPath = strings.TrimPrefix(changedPath, "/content")
					// If a metadata.yaml file has changed then force recreation of the
					// corresponding html file (or whole directory)
					if strings.HasSuffix(changedPath, "metadata.yaml") {
						changedPath = strings.TrimSuffix(changedPath, "metadata.yaml")
						for url, file := range context.Navigation.Filesystem {
							if strings.HasPrefix(url, changedPath) && file.MimeType == "text/html" {
								file.CachedContent = nil
								context.Navigation.Filesystem[url] = file
							}
						}
						break
					}

					// If a file was changed then update the Filesystem
					if isFile(event.Name) {
						// Remove a potential .html suffix
						changedPath = strings.TrimSuffix(changedPath, ".html")

						file, exists := context.Navigation.Filesystem[changedPath]
						if exists {
							if file.MimeType == "text/html" {
								file.CachedContent = nil
								context.Navigation.Filesystem[changedPath] = file
							}
							return
						}
					}

					// Otherwise assume that the event is a directory change, and update
					// everything that is below this directory
					for url, file := range context.Navigation.Filesystem {
						if strings.HasPrefix(url, changedPath) && file.MimeType == "text/html" {
							file.CachedContent = nil
							context.Navigation.Filesystem[url] = file
						}
					}
				}
			case err, ok := <-context.Watcher.Errors:
				if !ok {
					log.Fatal("Watcher events channel closed")
				} else {
					log.Printf("Watcher error %v", err)
				}
			}
		}
	}()

	// Populate the Filesystem structure
	contentRoot := filepath.Join(context.Config.SiteDirectory, "content")

	// Add the /layout to the watcher, to get informed about changes in the html header/footer
	if err := context.Watcher.Add(filepath.Join(context.Config.SiteDirectory, "layout")); err != nil {
		log.Printf("failed to add layout directory to watcher: %v", err)
	}

	// Recursively descend into all directories in /content and watch them
	fsys := os.DirFS(contentRoot)
	fs.WalkDir(fsys, ".", func(path string, dir fs.DirEntry, err error) error {
		if dir.IsDir() {
			if err := context.Watcher.Add(filepath.Join(contentRoot, path)); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}
