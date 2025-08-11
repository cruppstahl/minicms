package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches filesystem changes and updates the FileManager accordingly
type FileWatcher struct {
	mu          sync.RWMutex
	fileManager *FileManager
	watcher     *fsnotify.Watcher
	watchedDirs map[string]bool // Track which directories are being watched
	rootPath    string          // Root path being watched
	running     bool
	stopChan    chan struct{}
	eventChan   chan FileWatchEvent
}

// FileWatchEventType represents the type of file system event
type FileWatchEventType int

const (
	FileCreated FileWatchEventType = iota
	FileModified
	FileDeleted
	FileRenamed
	DirCreated
	DirDeleted
)

// FileWatchEvent represents a file system change event
type FileWatchEvent struct {
	Type    FileWatchEventType
	Path    string
	OldPath string // For rename events
	IsDir   bool
	Time    time.Time
}

// FileWatchSubscriber interface for receiving file watch events
type FileWatchSubscriber interface {
	OnFileEvent(event FileWatchEvent)
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(fm *FileManager) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &FileWatcher{
		fileManager: fm,
		watcher:     watcher,
		watchedDirs: make(map[string]bool),
		stopChan:    make(chan struct{}),
		eventChan:   make(chan FileWatchEvent, 100),
	}, nil
}

// Returns true if a path should be ignored (hidden files, symlinks, etc.)
func IgnoreFile(path string, info os.FileInfo) bool {
	// Get the base name
	baseName := filepath.Base(path)

	// Skip hidden files and directories
	if strings.HasPrefix(baseName, ".") {
		return true
	}

	// Skip symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}

	// Avoid .bak, .tmp, and other temporary files
	if strings.HasSuffix(baseName, ".bak") || strings.HasSuffix(baseName, ".tmp") ||
		strings.HasSuffix(baseName, "~") || strings.HasSuffix(baseName, ".swp") {
		return true
	}

	return false
}

// getRelativePath converts absolute path to relative path from root
func (fw *FileWatcher) getRelativePath(absPath string) (string, error) {
	return filepath.Rel(fw.rootPath, absPath)
}

// addDirectoryWatch adds a directory to the watcher recursively
func (fw *FileWatcher) addDirectoryWatch(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !IgnoreFile(path, info) {
			// TODO review if we require a mutex here!
			err := fw.watcher.Add(path)
			if err != nil {
				log.Printf("Failed to watch directory %s: %v", path, err)
				return err
			}

			fw.mu.Lock()
			fw.watchedDirs[path] = true
			fw.mu.Unlock()

			log.Printf("Watching directory: %s", path)
		}

		return nil
	})
}

// removeDirectoryWatch removes a directory from the watcher
/*
func (fw *FileWatcher) removeDirectoryWatch(dirPath string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Remove the directory and all subdirectories from watcher
	for watchedDir := range fw.watchedDirs {
		if strings.HasPrefix(watchedDir, dirPath) {
			fw.watcher.Remove(watchedDir)
			delete(fw.watchedDirs, watchedDir)
			log.Printf("Stopped watching directory: %s", watchedDir)
		}
	}
}*/

// handleFileCreated handles file creation events
/*
func (fw *FileWatcher) handleFileCreated(path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Failed to stat created file %s: %v", path, err)
		return
	}

	if fw.shouldIgnore(path, info) {
		return
	}

	relPath, err := fw.getRelativePath(path)
	if err != nil {
		log.Printf("Failed to get relative path for %s: %v", path, err)
		return
	}

	if info.IsDir() {
		// Add new directory to watcher
		fw.addDirectoryWatch(path)

		// Notify subscribers
		event := FileWatchEvent{
			Type:  DirCreated,
			Path:  relPath,
			IsDir: true,
			Time:  time.Now(),
		}
		fw.eventChan <- event
	} else {
		file := fw.fileManager.AddFile(relPath)

		// Process the new file with plugins
		fw.fileManager.GetPluginManager().Process(file, fw.fileManager)
		log.Printf("Processed new file %s", relPath)

		// Notify subscribers
		event := FileWatchEvent{
			Type:  FileCreated,
			Path:  relPath,
			IsDir: false,
			Time:  time.Now(),
		}
		fw.eventChan <- event
	}
}*/

// handleFileModified handles file modification events
func (fw *FileWatcher) handleFileModified(path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Failed to stat modified file %s: %v", path, err)
		return
	}

	if IgnoreFile(path, info) || info.IsDir() {
		return
	}

	relPath, err := fw.getRelativePath(path)
	if err != nil {
		log.Printf("Failed to get relative path for %s: %v", path, err)
		return
	}

	// Update file in FileManager
	file := fw.fileManager.AddFile(relPath)

	// Mark file and its dependents for update
	file.MarkForUpdate()

	// Update all files that need to be reprocessed
	fw.fileManager.ProcessUpdatedFiles()

	// Notify subscribers
	event := FileWatchEvent{
		Type:  FileModified,
		Path:  relPath,
		IsDir: false,
		Time:  time.Now(),
	}
	fw.eventChan <- event
}

// handleFileDeleted handles file deletion events
/*
func (fw *FileWatcher) handleFileDeleted(path string) {
	relPath, err := fw.getRelativePath(path)
	if err != nil {
		log.Printf("Failed to get relative path for %s: %v", path, err)
		return
	}

	// Check if it was a file or directory in our FileManager
	file := fw.fileManager.GetFile(relPath)
	dir := fw.fileManager.GetDirectory(relPath)

	if file != nil {
		// Remove file from FileManager
		fw.removeFileFromManager(relPath)

		// Notify subscribers
		event := FileWatchEvent{
			Type:  FileDeleted,
			Path:  relPath,
			IsDir: false,
			Time:  time.Now(),
		}
		fw.eventChan <- event

		log.Printf("Removed deleted file %s from FileManager", relPath)
	} else if dir != nil {
		// Remove directory and all its files from FileManager
		fw.removeDirectoryFromManager(relPath)

		// Remove from file watcher
		fw.removeDirectoryWatch(path)

		// Notify subscribers
		event := FileWatchEvent{
			Type:  DirDeleted,
			Path:  relPath,
			IsDir: true,
			Time:  time.Now(),
		}
		fw.eventChan <- event

		log.Printf("Removed deleted directory %s from FileManager", relPath)
	}
}*/

// removeFileFromManager removes a file from the FileManager
/*
func (fw *FileWatcher) removeFileFromManager(relPath string) {
	fw.fileManager.mu.Lock()
	defer fw.fileManager.mu.Unlock()

	file := fw.fileManager.Files[relPath]
	if file == nil {
		return
	}

	// Remove from parent directory
	if file.Parent != nil {
		delete(file.Parent.Files, file.Name)
	}

	// Remove all dependency relationships
	for depPath := range file.Dependencies {
		if dep := fw.fileManager.Files[depPath]; dep != nil {
			delete(dep.Dependents, relPath)
		}
	}

	for depPath := range file.Dependents {
		if dep := fw.fileManager.Files[depPath]; dep != nil {
			delete(dep.Dependencies, relPath)
		}
	}

	// Remove from global files map
	delete(fw.fileManager.Files, relPath)
}*/

// removeDirectoryFromManager removes a directory and all its contents from FileManager
/*
func (fw *FileWatcher) removeDirectoryFromManager(relPath string) {
	fw.fileManager.mu.Lock()
	defer fw.fileManager.mu.Unlock()

	// Find all files that start with this directory path
	var filesToRemove []string
	for filePath := range fw.fileManager.Files {
		if strings.HasPrefix(filePath, relPath+"/") || filePath == relPath {
			filesToRemove = append(filesToRemove, filePath)
		}
	}

	// Remove all files in the directory
	for _, filePath := range filesToRemove {
		if file := fw.fileManager.Files[filePath]; file != nil {
			// Remove dependency relationships
			for depPath := range file.Dependencies {
				if dep := fw.fileManager.Files[depPath]; dep != nil {
					delete(dep.Dependents, filePath)
				}
			}

			for depPath := range file.Dependents {
				if dep := fw.fileManager.Files[depPath]; dep != nil {
					delete(dep.Dependencies, filePath)
				}
			}

			delete(fw.fileManager.Files, filePath)
		}
	}

	// Remove directory from hierarchy (simplified - would need recursive removal)
	if dir := fw.fileManager.findDirectory(filepath.Dir(relPath)); dir != nil {
		delete(dir.Subdirs, filepath.Base(relPath))
	}
}*/

// Start starts the file watcher
func (fw *FileWatcher) Start(rootPath string) error {
	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return fmt.Errorf("file watcher is already running")
	}
	fw.running = true
	fw.rootPath = rootPath
	fw.mu.Unlock()

	// Add initial directory watches
	err := fw.addDirectoryWatch(rootPath)
	if err != nil {
		return fmt.Errorf("failed to add initial directory watches: %w", err)
	}

	// Start event processing goroutines
	go fw.processWatcherEvents()

	log.Printf("FileWatcher started, watching: %s", rootPath)
	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	if !fw.running {
		fw.mu.Unlock()
		return fmt.Errorf("file watcher is not running")
	}
	fw.running = false
	fw.mu.Unlock()

	// Stop the watcher
	close(fw.stopChan)
	err := fw.watcher.Close()

	log.Printf("FileWatcher stopped")
	return err
}

// processWatcherEvents processes fsnotify events
func (fw *FileWatcher) processWatcherEvents() {
	for {
		select {
		case <-fw.stopChan:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Handle different event types
			switch {
			// case event.Op&fsnotify.Create == fsnotify.Create:
			// fw.handleFileCreated(event.Name)
			case event.Op&fsnotify.Write == fsnotify.Write:
				fw.handleFileModified(event.Name)
				// case event.Op&fsnotify.Remove == fsnotify.Remove:
				// fw.handleFileDeleted(event.Name)
				// case event.Op&fsnotify.Rename == fsnotify.Rename:
				// fw.handleFileDeleted(event.Name) // Treat rename as delete for now
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("FileWatcher error: %v", err)
		}
	}
}

// IsRunning returns whether the file watcher is currently running
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.running
}
