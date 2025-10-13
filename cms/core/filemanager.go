package core

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// File represents a file with dependency tracking
type File struct {
	Name    string
	Path    string   // Path is relative to Config.SiteDirectory
	Routes  []string // Routes this file is associated with
	Content []byte
	Parent  *Directory // Reference to parent directory

	// Dependencies: files this file depends on
	Dependencies map[string]*File

	// Dependents: files that depend on this file
	Dependents map[string]*File

	// Additional metadata about the File
	Metadata FileMetadata
}

// Directory represents a directory that can contain files and subdirectories
type Directory struct {
	Name    string
	Path    string                // Full path from root
	Parent  *Directory            // Reference to parent directory (nil for root)
	Subdirs map[string]*Directory // Child directories
	Files   map[string]*File      // Files in this directory

	// Additional metadata about the File
	Metadata DirectoryMetadata
}

// FileManager manages the hierarchical file system with dependencies
type FileManager struct {
	mu            sync.RWMutex // Protects all data structures
	root          *Directory
	Files         map[string]*File // Global file lookup by full path
	SiteDirectory string
	pluginManager *PluginManager // Plugin system for file processing
}

// NewFileManager creates a new file manager with root directory
func NewFileManager(siteDirectory string) *FileManager {
	root := &Directory{
		Name:    "",
		Path:    "",
		Parent:  nil,
		Subdirs: make(map[string]*Directory),
		Files:   make(map[string]*File),
	}

	return &FileManager{
		root:          root,
		Files:         make(map[string]*File),
		pluginManager: NewPluginManager(),
		SiteDirectory: siteDirectory,
	}
}

// Do we need to invoke Plugins on this File?
func (f *File) NeedsUpdate() bool {
	return f.Content == nil
}

// Read the file data from disk, or nil in case of error
func (f *File) ReadFile(siteDirectory string) []byte {
	path := filepath.Join(siteDirectory, f.Path)
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read file %s: %v", path, err)
		return nil
	}
	return body
}

// Adds a dependency relationship to the other file
func (f *File) AddDependency(other *File) {
	f.Dependencies[other.Path] = other
	other.Dependents[f.Path] = f
}

// Marks this file and all its dependents for update (thread-safe)
func (f *File) MarkForUpdate() {
	visited := make(map[string]bool)
	f.markForUpdateRecursive(visited)
}

// Recursively marks file (and its dependencies) for update
func (f *File) markForUpdateRecursive(visited map[string]bool) {
	if visited[f.Path] {
		return
	}

	visited[f.Path] = true
	f.Content = nil // Trigger update

	// Mark all dependents
	for _, dep := range f.Dependents {
		dep.markForUpdateRecursive(visited)
	}
}

// Returns the plugin manager
func (fm *FileManager) GetPluginManager() *PluginManager {
	return fm.pluginManager
}

// Processes all files with their applicable plugins (thread-safe)
func (fm *FileManager) ProcessAllFiles() {
	fm.mu.RLock()
	files := make(map[string]*File, len(fm.Files))
	maps.Copy(files, fm.Files)
	fm.mu.RUnlock()

	// process outside locks (plugin code may be slow)
	for path, file := range files {
		newFile := fm.pluginManager.Process(*file, fm)
		// write back under write lock
		fm.mu.Lock()
		fm.Files[path] = newFile
		fm.mu.Unlock()
	}
}

// Processes all files which need to be updated (e.g. because they were modified)
func (fm *FileManager) ProcessUpdatedFiles() {
	// collect targets under read lock
	type upd struct {
		path string
		file *File
	}
	var toUpdate []upd

	fm.mu.RLock()
	for path, file := range fm.Files {
		if file.NeedsUpdate() {
			// capture path and pointer
			toUpdate = append(toUpdate, upd{path: path, file: file})
		}
	}
	fm.mu.RUnlock()

	// process outside locks (plugin code may be slow)
	for _, u := range toUpdate {
		newFile := fm.pluginManager.Process(*u.file, fm)
		// write back under write lock
		fm.mu.Lock()
		fm.Files[u.path] = newFile
		fm.mu.Unlock()
	}
}

// GetRoot returns the root directory (thread-safe)
func (fm *FileManager) GetRoot() *Directory {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.root
}

func (fm *FileManager) findDirectoryRecursive(dir *Directory, path []string) *Directory {
	if len(path) == 0 || (len(path) == 1 && path[0] == "" || path[0] == ".") {
		return dir
	}

	subdir := dir.Subdirs[path[0]]
	return fm.findDirectoryRecursive(subdir, path[1:])
}

// findDirectory finds a directory by path (assumes lock is held)
func (fm *FileManager) findDirectory(path string) *Directory {
	if path == "" || path == "." {
		return fm.root
	}

	// Clean the path and split it into parts
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	subdir := fm.root.Subdirs[parts[0]]

	return fm.findDirectoryRecursive(subdir, parts[1:])
}

// createDirectory creates a directory and all its parent directories (assumes lock is held)
func (fm *FileManager) createDirectory(path string) *Directory {
	if path == "" || path == "." {
		return fm.root
	}

	// Clean and split the path
	cleanPath := filepath.Clean(path)
	parts := strings.Split(cleanPath, string(filepath.Separator))

	current := fm.root
	currentPath := ""

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		if subdir, exists := current.Subdirs[part]; exists {
			current = subdir
		} else {
			// Create new directory
			newDir := &Directory{
				Name:    part,
				Path:    currentPath,
				Parent:  current,
				Subdirs: make(map[string]*Directory),
				Files:   make(map[string]*File),
			}
			current.Subdirs[part] = newDir
			current = newDir
		}
	}

	return current
}

// WalkDirectory recursively walks a directory and populates the FileManager
func (fm *FileManager) WalkDirectory(rootPath string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	absRootPath := filepath.Join(fm.SiteDirectory, rootPath)
	return filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories (starting with .)
		if IgnoreFile(info.Name(), info) {
			if info.IsDir() {
				return filepath.SkipDir // Skip entire hidden directory
			}
			return nil // Skip hidden file
		}

		// Convert absolute path to relative path from siteDirectory,
		// e.g. "/content/posts/my-post.md")
		relPath, err := filepath.Rel(fm.SiteDirectory, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Create directory structure
			fm.createDirectory(relPath)
		} else {
			// Create parent directory if it doesn't exist
			// TODO simplify this logic
			var parentDir *Directory
			dirPath := filepath.Dir(relPath)
			if dirPath != "." {
				parentDir = fm.createDirectory(dirPath)
			} else {
				parentDir = fm.root
			}

			// Add file to manager
			fileName := filepath.Base(relPath)
			file := &File{
				Name:         fileName,
				Path:         relPath,
				Parent:       parentDir,
				Content:      nil,
				Dependencies: make(map[string]*File),
				Dependents:   make(map[string]*File),
			}

			fm.Files[relPath] = file
			parentDir.Files[fileName] = file
		}

		return nil
	})
}

// Removes all files and directories under the given path
func (fm *FileManager) RemoveDirectory(rootPath string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	rootPath = filepath.Clean(rootPath)

	// Delete files
	for path, file := range fm.Files {
		if strings.HasPrefix(path, rootPath) {
			// Remove from parent directory
			parentDir := file.Parent
			if parentDir != nil {
				delete(parentDir.Files, file.Name)
			}

			// Remove dependencies
			for _, f := range fm.Files {
				delete(f.Dependencies, path)
				delete(f.Dependents, path)
			}

			// Remove from global files map
			delete(fm.Files, path)
		}
	}

	// Delete directories
	if dir := fm.findDirectory(rootPath); dir != nil {
		if parent := dir.Parent; parent != nil {
			delete(parent.Subdirs, dir.Name)
		}
	}
}

// AddFile adds or updates a file in the manager (thread-safe)
// Assumes the directory structure already exists
func (fm *FileManager) AddFile(path string) *File {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Clean the path
	cleanPath := filepath.Clean(path)
	fileName := filepath.Base(cleanPath)
	dirPath := filepath.Dir(cleanPath)

	// Find the parent directory (must already exist)
	var parentDir *Directory
	if dirPath == "." || dirPath == "" {
		parentDir = fm.root
	} else {
		parentDir = fm.findDirectory(dirPath)
		if parentDir == nil {
			// This shouldn't happen if WalkDirectory was used properly
			panic(fmt.Sprintf("parent directory %s does not exist for file %s", dirPath, cleanPath))
		}
	}

	// Check if file already exists
	file, exists := fm.Files[cleanPath]
	if !exists {
		file = &File{
			Name:         fileName,
			Path:         cleanPath,
			Parent:       parentDir,
			Dependencies: make(map[string]*File),
			Dependents:   make(map[string]*File),
		}
		fm.Files[cleanPath] = file
		parentDir.Files[fileName] = file
	}

	file.MarkForUpdate()
	return file
}

// Removes a file from the manager (thread-safe)
func (fm *FileManager) RemoveFile(path string) {
	file := fm.GetFile(path)
	if file == nil {
		return // File doesn't exist
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Clean the path
	cleanPath := filepath.Clean(path)
	fileName := filepath.Base(cleanPath)
	dirPath := filepath.Dir(cleanPath)

	// Find the parent directory (must already exist)
	var parentDir *Directory
	if dirPath == "." || dirPath == "" {
		parentDir = fm.root
	} else {
		parentDir = fm.findDirectory(dirPath)
		if parentDir == nil {
			// This shouldn't happen if WalkDirectory was used properly
			panic(fmt.Sprintf("parent directory %s does not exist for file %s", dirPath, cleanPath))
		}
	}

	// Check if file exists
	_, exists := fm.Files[cleanPath]
	if exists {
		delete(fm.Files, cleanPath)
		delete(parentDir.Files, fileName)
	}

	// Remove this file from dependencies of other files, and mark them all for update
	file.MarkForUpdate()
	for _, f := range fm.Files {
		delete(f.Dependencies, cleanPath)
		delete(f.Dependents, cleanPath)
	}
}

// GetFile returns a file by its full path (thread-safe)
func (fm *FileManager) GetFile(path string) *File {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	cleanPath := filepath.Clean(path)
	return fm.Files[cleanPath]
}

// GetDirectory returns a directory by its full path (thread-safe)
func (fm *FileManager) GetDirectory(path string) *Directory {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if path == "" || path == "." {
		return fm.root
	}

	cleanPath := filepath.Clean(path)
	parts := strings.Split(cleanPath, string(filepath.Separator))

	current := fm.root
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		if subdir, exists := current.Subdirs[part]; exists {
			current = subdir
		} else {
			return nil // Directory doesn't exist
		}
	}

	return current
}

// Returns all files in this directory and subdirectories (thread-safe)
func (fm *FileManager) GetAllFiles() map[string]*File {
	m := make(map[string]*File)

	fm.mu.RLock()
	defer fm.mu.RUnlock()

	maps.Copy(m, fm.Files)
	return m
}
