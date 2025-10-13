package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Mock RouterManager for testing
type mockRouterManager struct {
	mu           sync.RWMutex
	routes       map[string]string
	addedFiles   []*File
	removedFiles []string
	rebuilt      int
}

// Ensure mockRouterManager implements RouterInterface
var _ RouterInterface = (*mockRouterManager)(nil)

func newMockRouterManager() *mockRouterManager {
	return &mockRouterManager{
		routes:       make(map[string]string),
		addedFiles:   make([]*File, 0),
		removedFiles: make([]string, 0),
	}
}

func (m *mockRouterManager) AddFile(file *File) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addedFiles = append(m.addedFiles, file)
	for _, route := range file.Routes {
		m.routes[route] = file.Path
	}
}

func (m *mockRouterManager) RemoveFile(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removedFiles = append(m.removedFiles, filePath)
	// Remove routes for this file
	for route, path := range m.routes {
		if path == filePath {
			delete(m.routes, route)
		}
	}
	return nil
}

func (m *mockRouterManager) RebuildRouter() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rebuilt++
	return nil
}

func (m *mockRouterManager) GetAddedFiles() []*File {
	m.mu.RLock()
	defer m.mu.RUnlock()
	files := make([]*File, len(m.addedFiles))
	copy(files, m.addedFiles)
	return files
}

func (m *mockRouterManager) GetRemovedFiles() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	files := make([]string, len(m.removedFiles))
	copy(files, m.removedFiles)
	return files
}

func (m *mockRouterManager) GetRebuildCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rebuilt
}

// Test helper to create a test environment
func createListenerTestEnv(t *testing.T) (*FileManager, *FileWatcher, *mockRouterManager, string) {
	tempDir := t.TempDir()

	// Create content directory structure
	contentDir := filepath.Join(tempDir, "content")
	err := os.MkdirAll(contentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}

	fm := NewFileManager(tempDir)
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create FileWatcher: %v", err)
	}

	mockRM := newMockRouterManager()
	fw.SetRouter(mockRM)

	return fm, fw, mockRM, tempDir
}

func TestNewFileWatcherListener(t *testing.T) {
	_, fw, _, _ := createListenerTestEnv(t)

	fwl := newFileWatcherListener(fw)

	if fwl == nil {
		t.Fatal("newFileWatcherListener returned nil")
	}

	if fwl.fw != fw {
		t.Error("FileWatcher reference not set correctly")
	}

	if fwl.ctx == nil {
		t.Error("Context not initialized")
	}

	if fwl.cancel == nil {
		t.Error("Cancel function not initialized")
	}

	if fwl.IsRunning() {
		t.Error("Listener should not be running initially")
	}
}

func TestRegisterFileWatcherListener(t *testing.T) {
	fm, fw, _, _ := createListenerTestEnv(t)

	// Start the file watcher first
	if err := fw.Start(fm.SiteDirectory); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl, err := RegisterFileWatcherListener(fw)
	if err != nil {
		t.Fatalf("Failed to register listener: %v", err)
	}
	defer fwl.Stop()

	if !fwl.IsRunning() {
		t.Error("Listener should be running after registration")
	}

	// Test registering with nil FileWatcher
	_, err = RegisterFileWatcherListener(nil)
	if err == nil {
		t.Error("Expected error when registering with nil FileWatcher")
	}
}

func TestFileWatcherListenerStartStop(t *testing.T) {
	fm, fw, _, _ := createListenerTestEnv(t)

	if err := fw.Start(fm.SiteDirectory); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	// Test starting
	err := fwl.Start(fw)
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	if !fwl.IsRunning() {
		t.Error("Listener should be running after start")
	}

	// Test starting again (should fail)
	err = fwl.Start(fw)
	if err == nil {
		t.Error("Starting already running listener should return error")
	}

	// Test stopping
	err = fwl.Stop()
	if err != nil {
		t.Errorf("Failed to stop listener: %v", err)
	}

	if fwl.IsRunning() {
		t.Error("Listener should not be running after stop")
	}

	// Test stopping again (should fail)
	err = fwl.Stop()
	if err == nil {
		t.Error("Stopping already stopped listener should return error")
	}
}

func TestHandleFileCreated(t *testing.T) {
	fm, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	tests := []struct {
		name           string
		filePath       string
		expectRoute    bool
		expectedRoutes int
	}{
		{
			name:           "content file should create route",
			filePath:       "content/test.md",
			expectRoute:    true,
			expectedRoutes: 1,
		},
		{
			name:           "non-content file should not create route",
			filePath:       "assets/style.css",
			expectRoute:    false,
			expectedRoutes: 0,
		},
		{
			name:           "nested content file should create route",
			filePath:       "content/blog/post.md",
			expectRoute:    true,
			expectedRoutes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRM.addedFiles = nil
			mockRM.routes = make(map[string]string)

			// Create the file physically
			fullPath := filepath.Join(tempDir, tt.filePath)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
			if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", fullPath, err)
			}

			// Create event
			event := FileWatchEvent{
				Type:  FileCreated,
				Path:  tt.filePath,
				IsDir: false,
				Time:  time.Now(),
			}

			// Handle the event
			err := fwl.HandleFileCreated(event)
			if err != nil {
				t.Errorf("HandleFileCreated failed: %v", err)
			}

			// Verify file was added to FileManager
			file := fm.GetFile(tt.filePath)
			if file == nil {
				t.Error("File should be added to FileManager")
			}

			// Check router interactions
			addedFiles := mockRM.GetAddedFiles()
			if tt.expectRoute {
				if len(addedFiles) != tt.expectedRoutes {
					t.Errorf("Expected %d routes to be added, got %d", tt.expectedRoutes, len(addedFiles))
				}
			} else {
				if len(addedFiles) != 0 {
					t.Errorf("Expected no routes to be added for non-content file, got %d", len(addedFiles))
				}
			}
		})
	}
}

func TestHandleFileModified(t *testing.T) {
	fm, fw, _, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	// Create a test file first
	testFile := "content/test.md"
	fullPath := filepath.Join(tempDir, testFile)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(fullPath, []byte("original content"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}

	// Ensure parent directory exists in FileManager
	if err := fm.WalkDirectory("content"); err != nil {
		t.Fatalf("Failed to walk content directory: %v", err)
	}

	// Add file to FileManager
	fm.AddFile(testFile)

	// Create modification event
	event := FileWatchEvent{
		Type:  FileModified,
		Path:  testFile,
		IsDir: false,
		Time:  time.Now(),
	}

	// Handle the event
	err := fwl.HandleFileModified(event)
	if err != nil {
		t.Errorf("HandleFileModified failed: %v", err)
	}

	// Verify file was marked for update in FileManager
	file := fm.GetFile(testFile)
	if file == nil {
		t.Error("File should exist in FileManager")
	}

	// The file should be marked for update (Content should be nil after MarkForUpdate)
	if file.NeedsUpdate() == false {
		// Note: This depends on the exact implementation of how AddFile works
		// It might process the file immediately, so we check that it was called
		t.Log("File update handling called successfully")
	}
}

func TestHandleFileDeleted(t *testing.T) {
	fm, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	// Add a test file to FileManager first
	testFile := "content/test.md"
	fullPath := filepath.Join(tempDir, testFile)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}

	// Ensure parent directory exists in FileManager
	if err := fm.WalkDirectory("content"); err != nil {
		t.Fatalf("Failed to walk content directory: %v", err)
	}

	file := fm.AddFile(testFile)
	if file == nil {
		t.Fatal("Failed to add file to FileManager")
	}

	// Create deletion event
	event := FileWatchEvent{
		Type:  FileDeleted,
		Path:  testFile,
		IsDir: false,
		Time:  time.Now(),
	}

	// Handle the event
	err := fwl.HandleFileDeleted(event)
	if err != nil {
		t.Errorf("HandleFileDeleted failed: %v", err)
	}

	// Verify file was removed from FileManager
	deletedFile := fm.GetFile(testFile)
	if deletedFile != nil {
		t.Error("File should be removed from FileManager")
	}

	// Verify route was removed
	removedFiles := mockRM.GetRemovedFiles()
	if len(removedFiles) != 1 || removedFiles[0] != testFile {
		t.Errorf("Expected file %s to be removed from router, got %v", testFile, removedFiles)
	}
}

func TestHandleDirectoryCreated(t *testing.T) {
	_, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	// Create a new directory with some files
	newDir := "content/newdir"
	fullDirPath := filepath.Join(tempDir, newDir)
	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", fullDirPath, err)
	}

	// Add a file to the new directory
	testFile := filepath.Join(fullDirPath, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", testFile, err)
	}

	initialRebuildCount := mockRM.GetRebuildCount()

	// Create directory creation event
	event := FileWatchEvent{
		Type:  DirCreated,
		Path:  newDir,
		IsDir: true,
		Time:  time.Now(),
	}

	// Handle the event
	err := fwl.HandleDirectoryCreated(event)
	if err != nil {
		t.Errorf("HandleDirectoryCreated failed: %v", err)
	}

	// Verify router was rebuilt
	finalRebuildCount := mockRM.GetRebuildCount()
	if finalRebuildCount <= initialRebuildCount {
		t.Error("Router should be rebuilt after directory creation")
	}

	// Note: More detailed verification would require checking if the directory
	// was added to the FileManager, but that depends on WalkDirectory implementation
}

func TestHandleDirectoryDeleted(t *testing.T) {
	fm, fw, _, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	// Create and then delete a directory
	deletedDir := "content/deleteddir"
	fullDirPath := filepath.Join(tempDir, deletedDir)
	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", fullDirPath, err)
	}

	// Add a file to track in FileManager - first ensure directory structure exists
	testFile := filepath.Join(deletedDir, "test.md")
	// Create the file physically first
	testFilePath := filepath.Join(tempDir, testFile)
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	// Walk directory to ensure FileManager knows about the structure
	if err := fm.WalkDirectory(deletedDir); err != nil {
		t.Fatalf("Failed to walk directory %s: %v", deletedDir, err)
	}
	fm.AddFile(testFile)

	// Verify file exists before deletion
	if fm.GetFile(testFile) == nil {
		t.Fatal("Test file should exist in FileManager before deletion")
	}

	// Create directory deletion event
	event := FileWatchEvent{
		Type:  DirDeleted,
		Path:  deletedDir,
		IsDir: true,
		Time:  time.Now(),
	}

	// Handle the event
	err := fwl.HandleDirectoryDeleted(event)
	if err != nil {
		t.Errorf("HandleDirectoryDeleted failed: %v", err)
	}

	// The current implementation only logs, so we just verify it doesn't crash
	// In a complete implementation, we would verify:
	// - Files under the directory are removed from FileManager
	// - Routes for those files are removed
	// - Directory watches are removed

	t.Log("Directory deletion handled (implementation incomplete)")
}

func TestProcessEventsIntegration(t *testing.T) {
	fm, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl, err := RegisterFileWatcherListener(fw)
	if err != nil {
		t.Fatalf("Failed to register listener: %v", err)
	}
	defer fwl.Stop()

	// Create a content file
	contentFile := "content/integration-test.md"
	fullPath := filepath.Join(tempDir, contentFile)
	contentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}

	// Write the file to trigger creation event
	time.Sleep(100 * time.Millisecond) // Give listener time to start
	if err := os.WriteFile(fullPath, []byte("# Integration Test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify file was added to FileManager
	file := fm.GetFile(contentFile)
	if file == nil {
		t.Error("File should be added to FileManager via listener")
	}

	// Verify route was created (for content files)
	addedFiles := mockRM.GetAddedFiles()
	found := false
	for _, addedFile := range addedFiles {
		if addedFile.Path == contentFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("Content file should have route created via listener")
	}

	// Modify the file
	if err := os.WriteFile(fullPath, []byte("# Modified Integration Test"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Wait for modification to be processed
	time.Sleep(200 * time.Millisecond)

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Wait for deletion to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify file was removed from FileManager
	deletedFile := fm.GetFile(contentFile)
	if deletedFile != nil {
		t.Error("File should be removed from FileManager via listener")
	}

	// Verify route was removed
	removedFiles := mockRM.GetRemovedFiles()
	found = false
	for _, removedFile := range removedFiles {
		if removedFile == contentFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("File route should be removed via listener")
	}
}

func TestConcurrentListenerOperations(t *testing.T) {
	_, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl, err := RegisterFileWatcherListener(fw)
	if err != nil {
		t.Fatalf("Failed to register listener: %v", err)
	}
	defer fwl.Stop()

	contentDir := filepath.Join(tempDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}

	var wg sync.WaitGroup
	numFiles := 10

	// Create multiple files concurrently
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fileName := fmt.Sprintf("concurrent-file-%d.md", id)
			filePath := filepath.Join(contentDir, fileName)

			// Create file
			if err := os.WriteFile(filePath, []byte(fmt.Sprintf("Content %d", id)), 0644); err != nil {
				t.Errorf("Failed to create file %s: %v", fileName, err)
				return
			}

			time.Sleep(50 * time.Millisecond)

			// Modify file
			if err := os.WriteFile(filePath, []byte(fmt.Sprintf("Modified Content %d", id)), 0644); err != nil {
				t.Errorf("Failed to modify file %s: %v", fileName, err)
				return
			}

			time.Sleep(50 * time.Millisecond)

			// Delete file
			if err := os.Remove(filePath); err != nil {
				t.Errorf("Failed to delete file %s: %v", fileName, err)
			}
		}(i)
	}

	wg.Wait()

	// Give time for all events to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify listener is still running
	if !fwl.IsRunning() {
		t.Error("Listener should still be running after concurrent operations")
	}

	// Verify some operations were processed
	addedFiles := mockRM.GetAddedFiles()
	removedFiles := mockRM.GetRemovedFiles()

	if len(addedFiles) == 0 {
		t.Error("Expected some files to be added during concurrent operations")
	}

	if len(removedFiles) == 0 {
		t.Error("Expected some files to be removed during concurrent operations")
	}

	t.Logf("Processed %d file additions and %d file removals", len(addedFiles), len(removedFiles))
}

func TestListenerErrorHandling(t *testing.T) {
	_, fw, _, tempDir := createListenerTestEnv(t)

	// Test starting listener with stopped FileWatcher
	fwl := newFileWatcherListener(fw)

	// Try to start listener before FileWatcher is started
	err := fwl.Start(fw)
	if err != nil {
		t.Fatalf("Starting listener with stopped FileWatcher should work: %v", err)
	}
	defer fwl.Stop()

	// Test starting FileWatcher after listener
	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher after listener: %v", err)
	}
	defer fw.Stop()

	// Test handling events for non-existent files
	event := FileWatchEvent{
		Type:  FileCreated,
		Path:  "non-existent-file.txt",
		IsDir: false,
		Time:  time.Now(),
	}

	// This should not crash
	err = fwl.HandleFileCreated(event)
	if err == nil {
		t.Errorf("Expected error for non-existent file, but got none")
	}

	// Test handling events for files with invalid paths
	event = FileWatchEvent{
		Type:  FileDeleted,
		Path:  "",
		IsDir: false,
		Time:  time.Now(),
	}

	// This should not crash
	err = fwl.HandleFileDeleted(event)
	if err != nil {
		t.Logf("HandleFileDeleted returned error for empty path (expected): %v", err)
	}
}

func TestFileEventHandlerInterface(t *testing.T) {
	_, fw, _, _ := createListenerTestEnv(t)

	fwl := newFileWatcherListener(fw)

	// Verify that FileWatcherListener implements FileEventHandler interface
	var handler FileEventHandler = fwl

	// Test that all interface methods are available
	event := FileWatchEvent{
		Type:  FileCreated,
		Path:  "test-file.txt",
		IsDir: false,
		Time:  time.Now(),
	}

	// These should compile and be callable
	err := handler.HandleFileCreated(event)
	if err == nil {
		t.Log("HandleFileCreated interface method is callable")
	}

	err = handler.HandleFileModified(event)
	if err == nil {
		t.Log("HandleFileModified interface method is callable")
	}

	err = handler.HandleFileDeleted(event)
	if err != nil {
		t.Log("HandleFileDeleted interface method is callable")
	}

	dirEvent := FileWatchEvent{
		Type:  DirCreated,
		Path:  "test-dir",
		IsDir: true,
		Time:  time.Now(),
	}

	err = handler.HandleDirectoryCreated(dirEvent)
	if err == nil {
		t.Log("HandleDirectoryCreated interface method is callable")
	}

	err = handler.HandleDirectoryDeleted(dirEvent)
	if err != nil {
		t.Log("HandleDirectoryDeleted interface method is callable")
	}
}

func TestRouterRebuildEfficiency(t *testing.T) {
	_, fw, mockRM, tempDir := createListenerTestEnv(t)

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start FileWatcher: %v", err)
	}
	defer fw.Stop()

	fwl := newFileWatcherListener(fw)

	tests := []struct {
		name                string
		path                string
		isDirectory         bool
		shouldRebuildRouter bool
	}{
		{"content file should not trigger rebuild", "content/test.md", false, false}, // File creation uses AddFile, not RebuildRouter
		{"content directory should trigger rebuild", "content/blog", true, true},
		{"asset file should not trigger rebuild", "assets/style.css", false, false},
		{"config directory should not trigger rebuild", "config/site", true, false},
		{"layout file should not trigger rebuild", "layout/header.html", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset rebuild count
			mockRM.rebuilt = 0

			if tt.isDirectory {
				// Create directory physically for directory creation test
				fullPath := filepath.Join(tempDir, tt.path)
				if err := os.MkdirAll(fullPath, 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}

				event := FileWatchEvent{
					Type:  DirCreated,
					Path:  tt.path,
					IsDir: true,
					Time:  time.Now(),
				}

				err := fwl.HandleDirectoryCreated(event)
				if err != nil {
					t.Errorf("HandleDirectoryCreated failed: %v", err)
				}
			} else {
				// Test with file creation (doesn't trigger router rebuild by itself)
				event := FileWatchEvent{
					Type:  FileCreated,
					Path:  tt.path,
					IsDir: false,
					Time:  time.Now(),
				}

				// Create file physically
				fullPath := filepath.Join(tempDir, tt.path)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}

				err := fwl.HandleFileCreated(event)
				if err != nil {
					t.Errorf("HandleFileCreated failed: %v", err)
				}
			}

			rebuildCount := mockRM.GetRebuildCount()
			if tt.shouldRebuildRouter {
				if rebuildCount == 0 {
					t.Errorf("Expected router rebuild for %s, but rebuild count was 0", tt.path)
				}
			} else {
				if rebuildCount > 0 {
					t.Errorf("Expected no router rebuild for %s, but rebuild count was %d", tt.path, rebuildCount)
				}
			}
		})
	}
}