package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// RouterInterface defines the interface that FileWatcher needs from RouterManager
type RouterInterface interface {
	AddFile(file *File)
	RemoveFile(filePath string) error
	RebuildRouter() error
}

// FileWatcher watches filesystem changes and updates the FileManager accordingly
type FileWatcher struct {
	mu          sync.RWMutex
	rm          RouterInterface
	fm          *FileManager
	watcher     *fsnotify.Watcher
	watchedDirs map[string]bool // Track which directories are being watched
	rootPath    string          // Root path being watched
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	eventChan   chan FileWatchEvent
	wg          sync.WaitGroup
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

// String returns a string representation of the event type
func (t FileWatchEventType) String() string {
	switch t {
	case FileCreated:
		return "FileCreated"
	case FileModified:
		return "FileModified"
	case FileDeleted:
		return "FileDeleted"
	case FileRenamed:
		return "FileRenamed"
	case DirCreated:
		return "DirCreated"
	case DirDeleted:
		return "DirDeleted"
	default:
		return "Unknown"
	}
}

// FileWatchEvent represents a file system change event
type FileWatchEvent struct {
	Type    FileWatchEventType
	Path    string
	OldPath string // For rename events
	IsDir   bool
	Time    time.Time
}

// Creates a new file watcher
func NewFileWatcher(fm *FileManager) (*FileWatcher, error) {
	if fm == nil {
		return nil, fmt.Errorf("file manager cannot be nil")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &FileWatcher{
		fm:          fm,
		watcher:     watcher,
		watchedDirs: make(map[string]bool),
		ctx:         ctx,
		cancel:      cancel,
		eventChan:   make(chan FileWatchEvent, 100),
	}, nil
}

func (fw *FileWatcher) SetRouter(rm RouterInterface) {
	fw.rm = rm
}

// Returns true if a path should be ignored (hidden files, symlinks, etc.)
func IgnoreFile(path string, info os.FileInfo) bool {
	if info == nil {
		return true
	}

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
	tmpSuffixes := []string{".bak", ".tmp", "~", ".swp", ".lock"}
	for _, suffix := range tmpSuffixes {
		if strings.HasSuffix(baseName, suffix) {
			return true
		}
	}

	return false
}

// Converts absolute path to relative path from root
func (fw *FileWatcher) getRelativePath(absPath string) (string, error) {
	if fw.rootPath == "" {
		return "", fmt.Errorf("root path not set")
	}
	return filepath.Rel(fw.rootPath, absPath)
}

// Adds a directory to the watcher recursively
func (fw *FileWatcher) addDirectoryWatch(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue processing other files
			log.Printf("Error walking path %s: %v", path, err)
			return nil
		}

		if info.IsDir() && !IgnoreFile(path, info) {
			if err := fw.watcher.Add(path); err != nil {
				log.Printf("Failed to watch directory %s: %v", path, err)
				return nil // Continue processing other directories
			}

			// Properly lock when updating watchedDirs
			fw.mu.Lock()
			fw.watchedDirs[path] = true
			fw.mu.Unlock()

			log.Printf("Watching directory: %s", path)
		}

		return nil
	})
}

// Removes a directory from the watcher
func (fw *FileWatcher) removeDirectoryWatch(dirPath string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Remove the directory and all subdirectories from watcher
	for watchedDir := range fw.watchedDirs {
		if strings.HasPrefix(watchedDir, dirPath) {
			if err := fw.watcher.Remove(watchedDir); err != nil {
				log.Printf("Failed to remove watcher for %s: %v", watchedDir, err)
			}
			delete(fw.watchedDirs, watchedDir)
			log.Printf("Stopped watching directory: %s", watchedDir)
		}
	}
}

// Handle file modification events
func (fw *FileWatcher) handleFileModified(path string) {
	info, err := os.Stat(path)
	if err != nil {
		// File might have been deleted between event and stat
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

	// Send event to subscribers
	event := FileWatchEvent{
		Type:  FileModified,
		Path:  relPath,
		IsDir: false,
		Time:  time.Now(),
	}

	select {
	case fw.eventChan <- event:
	case <-fw.ctx.Done():
		return
	default:
		log.Printf("Event channel full, dropping event for %s", relPath)
	}
}

// Starts the file watcher
func (fw *FileWatcher) Start(rootPath string) error {
	if rootPath == "" {
		return fmt.Errorf("root path cannot be empty")
	}

	// Check if path exists and is a directory
	if info, err := os.Stat(rootPath); err != nil {
		return fmt.Errorf("failed to access root path %s: %w", rootPath, err)
	} else if !info.IsDir() {
		return fmt.Errorf("root path %s is not a directory", rootPath)
	}

	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return fmt.Errorf("file watcher is already running")
	}
	fw.running = true
	fw.rootPath = rootPath
	fw.mu.Unlock()

	// Add initial directory watches
	if err := fw.addDirectoryWatch(rootPath); err != nil {
		fw.mu.Lock()
		fw.running = false
		fw.mu.Unlock()
		return fmt.Errorf("failed to add initial directory watches: %w", err)
	}

	// Start event processing goroutine
	fw.wg.Add(1)
	go fw.processWatcherEvents()

	log.Printf("FileWatcher started, watching: %s", rootPath)
	return nil
}

// Stops the file watcher
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	if !fw.running {
		fw.mu.Unlock()
		return fmt.Errorf("file watcher is not running")
	}
	fw.running = false
	fw.mu.Unlock()

	// Cancel context to signal shutdown
	fw.cancel()

	// Close the watcher
	err := fw.watcher.Close()

	// Wait for goroutines to finish
	fw.wg.Wait()

	// Close event channel
	close(fw.eventChan)

	log.Printf("FileWatcher stopped")
	return err
}

func (fw *FileWatcher) processWatcherEvents() {
	defer fw.wg.Done()

	for {
		select {
		case <-fw.ctx.Done():
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Handle different event types
			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				fw.handleFileModified(event.Name)
			case event.Op&fsnotify.Create == fsnotify.Create:
				// Handle file/directory creation
				fw.handleFileCreated(event.Name)
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				// Handle file/directory deletion
				fw.handleFileDeleted(event.Name)
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				// Handle rename as deletion for now
				fw.handleFileDeleted(event.Name)
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("FileWatcher error: %v", err)
		}
	}
}

// handles file creation events
func (fw *FileWatcher) handleFileCreated(path string) {
	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Failed to stat created file %s: %v", path, err)
		return
	}

	if IgnoreFile(path, info) {
		return
	}

	relPath, err := fw.getRelativePath(path)
	if err != nil {
		log.Printf("Failed to get relative path for %s: %v", path, err)
		return
	}

	if info.IsDir() {
		// Send event
		event := FileWatchEvent{
			Type:  DirCreated,
			Path:  relPath,
			IsDir: true,
			Time:  time.Now(),
		}
		fw.sendEvent(event)
	} else {
		// Send event
		event := FileWatchEvent{
			Type:  FileCreated,
			Path:  relPath,
			IsDir: false,
			Time:  time.Now(),
		}
		fw.sendEvent(event)
	}
}

// handles file deletion events
func (fw *FileWatcher) handleFileDeleted(path string) {
	relPath, err := fw.getRelativePath(path)
	if err != nil {
		log.Printf("Failed to get relative path for %s: %v", path, err)
		return
	}

	// Check if it was a directory by checking if we were watching it
	fw.mu.RLock()
	wasDir := fw.watchedDirs[path]
	fw.mu.RUnlock()

	if wasDir {
		// Send event
		event := FileWatchEvent{
			Type:  DirDeleted,
			Path:  relPath,
			IsDir: true,
			Time:  time.Now(),
		}
		fw.sendEvent(event)
	} else {
		// Send event
		event := FileWatchEvent{
			Type:  FileDeleted,
			Path:  relPath,
			IsDir: false,
			Time:  time.Now(),
		}
		fw.sendEvent(event)
	}
}

// sends an event to the event channel with proper context handling
func (fw *FileWatcher) sendEvent(event FileWatchEvent) {
	select {
	case fw.eventChan <- event:
	case <-fw.ctx.Done():
		return
	default:
		log.Printf("Event channel full, dropping event for %s", event.Path)
	}
}

// returns whether the file watcher is currently running
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.running
}

// returns the event channel for subscribers
func (fw *FileWatcher) GetEventChannel() <-chan FileWatchEvent {
	return fw.eventChan
}

// returns a copy of currently watched directories
func (fw *FileWatcher) GetWatchedDirectories() []string {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	dirs := make([]string, 0, len(fw.watchedDirs))
	for dir := range fw.watchedDirs {
		dirs = append(dirs, dir)
	}
	return dirs
}
