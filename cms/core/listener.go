package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileEventHandler defines the interface for handling file system events
type FileEventHandler interface {
	HandleFileCreated(event FileWatchEvent) error
	HandleFileModified(event FileWatchEvent) error
	HandleFileDeleted(event FileWatchEvent) error
	HandleDirectoryCreated(event FileWatchEvent) error
	HandleDirectoryDeleted(event FileWatchEvent) error
}

// FileWatcherListener implements FileEventHandler and provides file system event handling
type FileWatcherListener struct {
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	fw      *FileWatcher
}

// Ensure FileWatcherListener implements FileEventHandler
var _ FileEventHandler = (*FileWatcherListener)(nil)

// Helper function to determine if a path affects routes
func (fwl *FileWatcherListener) affectsRoutes(path string) bool {
	return strings.HasPrefix(path, "content/")
}

// Helper function to determine if a router rebuild is needed
func (fwl *FileWatcherListener) needsRouterRebuild(path string, isDirectory bool) bool {
	if isDirectory {
		// Directory operations in content always need rebuild
		return strings.HasPrefix(path, "content/")
	}
	// File operations in content need rebuild
	return fwl.affectsRoutes(path)
}

// Registration function
func RegisterFileWatcherListener(fw *FileWatcher) (*FileWatcherListener, error) {
	fwl := newFileWatcherListener(fw)

	if err := fwl.Start(fw); err != nil {
		return nil, fmt.Errorf("failed to start file watcher listener: %w", err)
	}

	return fwl, nil
}

// Creates a new file watcher event listener
func newFileWatcherListener(fw *FileWatcher) *FileWatcherListener {
	ctx, cancel := context.WithCancel(context.Background())

	return &FileWatcherListener{
		ctx:    ctx,
		fw:     fw,
		cancel: cancel,
	}
}

// Start begins listening to events from the file watcher
func (fwl *FileWatcherListener) Start(fw *FileWatcher) error {
	if fw == nil {
		return fmt.Errorf("file watcher cannot be nil")
	}

	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	if fwl.running {
		return fmt.Errorf("listener is already running")
	}

	fwl.running = true

	// Get the event channel from the file watcher
	eventChan := fw.GetEventChannel()

	// Start the event processing goroutine
	fwl.wg.Add(1)
	go fwl.processEvents(eventChan)

	log.Printf("Started listening to file watcher events")
	return nil
}

// Stop stops the event listener
func (fwl *FileWatcherListener) Stop() error {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	if !fwl.running {
		return fmt.Errorf("listener is not running")
	}

	fwl.running = false

	// Cancel the context to signal shutdown
	fwl.cancel()

	// Wait for the processing goroutine to finish
	fwl.wg.Wait()

	log.Printf("Stopped listening to file watcher events")
	return nil
}

// IsRunning returns whether the listener is currently active
func (fwl *FileWatcherListener) IsRunning() bool {
	fwl.mu.RLock()
	defer fwl.mu.RUnlock()
	return fwl.running
}

// processEvents is the main event processing loop
func (fwl *FileWatcherListener) processEvents(eventChan <-chan FileWatchEvent) {
	defer fwl.wg.Done()

	log.Printf("Event processing started")

	for {
		select {
		case <-fwl.ctx.Done():
			log.Printf("Event processing stopped (context cancelled)")
			return

		case event, ok := <-eventChan:
			if !ok {
				log.Printf("Event processing stopped (channel closed)")
				return
			}

			switch event.Type {
			case FileCreated:
				if err := fwl.HandleFileCreated(event); err != nil {
					log.Printf("Error handling file creation: %v", err)
				}
			case FileModified:
				if err := fwl.HandleFileModified(event); err != nil {
					log.Printf("Error handling file modification: %v", err)
				}
			case FileDeleted:
				if err := fwl.HandleFileDeleted(event); err != nil {
					log.Printf("Error handling file deletion: %v", err)
				}
			case FileRenamed:
				if err := fwl.HandleFileDeleted(event); err != nil {
					log.Printf("Error handling file rename (deletion): %v", err)
				}
				if err := fwl.HandleFileCreated(event); err != nil {
					log.Printf("Error handling file rename (creation): %v", err)
				}
			case DirCreated:
				if err := fwl.HandleDirectoryCreated(event); err != nil {
					log.Printf("Error handling directory creation: %v", err)
				}
			case DirDeleted:
				if err := fwl.HandleDirectoryDeleted(event); err != nil {
					log.Printf("Error handling directory deletion: %v", err)
				}
			}
		}
	}
}

// HandleFileModified implements FileEventHandler
func (fwl *FileWatcherListener) HandleFileModified(event FileWatchEvent) error {
	log.Printf("Processing file modification: %s", event.Path)

	// Update file in FileManager
	file := fwl.fw.fm.AddFile(event.Path)
	if file == nil {
		err := fmt.Errorf("failed to add modified file to FileManager: %s", event.Path)
		log.Printf("Error: %v", err)
		return err
	}

	// Update all files that need to be reprocessed
	fwl.fw.fm.ProcessUpdatedFiles()

	log.Printf("Successfully processed file modification: %s", event.Path)
	return nil
}

// HandleFileCreated implements FileEventHandler
func (fwl *FileWatcherListener) HandleFileCreated(event FileWatchEvent) error {
	log.Printf("Processing file creation: %s", event.Path)

	// Check if the file actually exists on disk
	absolutePath := filepath.Join(fwl.fw.rootPath, event.Path)
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		err = fmt.Errorf("file creation event for non-existent file: %s", event.Path)
		log.Printf("Error: %v", err)
		return err
	} else if err != nil {
		err = fmt.Errorf("failed to stat file %s: %v", event.Path, err)
		log.Printf("Error: %v", err)
		return err
	}

	// Ensure parent directory exists in FileManager by walking from root
	dirPath := filepath.Dir(event.Path)
	if dirPath != "." && dirPath != "" {
		if err := fwl.fw.fm.WalkDirectory(dirPath); err != nil {
			log.Printf("Warning: failed to walk directory %s: %v", dirPath, err)
		}
	}

	// Add file to FileManager
	file := fwl.fw.fm.AddFile(event.Path)
	if file == nil {
		err := fmt.Errorf("failed to add created file to FileManager: %s", event.Path)
		log.Printf("Error: %v", err)
		return err
	}

	// Process (a copy of the) new file with plugins
	processedFile := fwl.fw.fm.GetPluginManager().Process(*file, fwl.fw.fm)
	if processedFile == nil {
		log.Printf("Warning: plugin processing returned nil for file: %s", event.Path)
		processedFile = file
	}

	// If it's in the content directory, add a route
	if fwl.affectsRoutes(processedFile.Path) {
		log.Printf("Adding route for content file: %s", processedFile.Path)
		fwl.fw.rm.AddFile(processedFile)
	}

	log.Printf("Successfully processed file creation: %s", event.Path)
	return nil
}

// HandleFileDeleted implements FileEventHandler
func (fwl *FileWatcherListener) HandleFileDeleted(event FileWatchEvent) error {
	path := event.Path
	log.Printf("Processing file deletion: %s", path)

	// Remove file from FileManager
	fwl.fw.fm.RemoveFile(path)
	log.Printf("Removed file from FileManager: %s", path)

	// Also remove from RouterManager if it affects routes
	if fwl.affectsRoutes(path) {
		log.Printf("Removing routes for content file: %s", path)
		if err := fwl.fw.rm.RemoveFile(path); err != nil {
			// Log warning but don't fail - file might not have had routes
			log.Printf("Warning: failed to remove file from router: %s: %v", path, err)
		}
	}

	// Update all files that need to be reprocessed
	fwl.fw.fm.ProcessUpdatedFiles()

	log.Printf("Successfully processed file deletion: %s", path)
	return nil
}

// HandleDirectoryCreated implements FileEventHandler
func (fwl *FileWatcherListener) HandleDirectoryCreated(event FileWatchEvent) error {
	log.Printf("Processing directory creation: %s", event.Path)

	// Add new directory to watcher (convert to absolute path)
	absolutePath := filepath.Join(fwl.fw.rootPath, event.Path)
	log.Printf("Adding directory watch: %s (absolute: %s)", event.Path, absolutePath)
	if err := fwl.fw.addDirectoryWatch(absolutePath); err != nil {
		err = fmt.Errorf("failed to watch new directory %s: %v", absolutePath, err)
		log.Printf("Error: %v", err)
		return err
	}

	// Walk directory to add all files to FileManager
	log.Printf("Walking directory to discover files: %s", event.Path)
	if err := fwl.fw.fm.WalkDirectory(event.Path); err != nil {
		err = fmt.Errorf("failed to walk new directory %s: %v", event.Path, err)
		log.Printf("Error: %v", err)
		return err
	}

	// Update all files that need to be reprocessed
	fwl.fw.fm.ProcessUpdatedFiles()

	// Rebuild router only if the directory affects routes
	if fwl.needsRouterRebuild(event.Path, true) {
		log.Printf("Rebuilding router for directory affecting content routes: %s", event.Path)
		if err := fwl.fw.rm.RebuildRouter(); err != nil {
			err = fmt.Errorf("failed to rebuild router after directory creation %s: %v", event.Path, err)
			log.Printf("Error: %v", err)
			return err
		}
	} else {
		log.Printf("Skipping router rebuild for non-content directory: %s", event.Path)
	}

	log.Printf("Successfully processed directory creation: %s", event.Path)
	return nil
}

// HandleDirectoryDeleted implements FileEventHandler
func (fwl *FileWatcherListener) HandleDirectoryDeleted(event FileWatchEvent) error {
	log.Printf("Processing directory deletion: %s", event.Path)

	// Remove directory from watcher
	log.Printf("Removing directory watch: %s", event.Path)
	fwl.fw.removeDirectoryWatch(event.Path)

	// Remove directory and all its files from FileManager
	log.Printf("Removing directory and all files from FileManager: %s", event.Path)
	fwl.fw.fm.RemoveDirectory(event.Path)

	// Update all files that need to be reprocessed due to dependency changes
	fwl.fw.fm.ProcessUpdatedFiles()

	// Rebuild router only if the directory affects routes
	if fwl.needsRouterRebuild(event.Path, true) {
		log.Printf("Rebuilding router for directory affecting content routes: %s", event.Path)
		if err := fwl.fw.rm.RebuildRouter(); err != nil {
			err = fmt.Errorf("failed to rebuild router after directory deletion %s: %v", event.Path, err)
			log.Printf("Error: %v", err)
			return err
		}
	} else {
		log.Printf("Skipping router rebuild for non-content directory: %s", event.Path)
	}

	log.Printf("Successfully processed directory deletion: %s", event.Path)
	return nil
}
