package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test helper to create temporary directory structure
func createTestDir(t *testing.T) string {
	tempDir := t.TempDir()

	// Create some test files and directories
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.txt",
		"subdir/nested/file4.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := ioutil.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	return tempDir
}

func TestNewFileWatcher(t *testing.T) {
	siteDir := "/test/site"
	fm := NewFileManager(siteDir)

	tests := []struct {
		name        string
		fileManager interface{}
		expectError bool
	}{
		{
			name:        "valid file manager",
			fileManager: fm,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fw, err := NewFileWatcher(tt.fileManager.(*FileManager))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if fw == nil {
				t.Error("Expected file watcher but got nil")
				return
			}

			// Check initial state
			if fw.IsRunning() {
				t.Error("File watcher should not be running initially")
			}

			if fw.eventChan == nil {
				t.Error("Event channel should be initialized")
			}

			if fw.watchedDirs == nil {
				t.Error("Watched directories map should be initialized")
			}
		})
	}
}

func TestIgnoreFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		mode     os.FileMode
		expected bool
	}{
		{"regular file", "/path/to/file.txt", 0644, false},
		{"hidden file", "/path/to/.hidden", 0644, true},
		{"backup file", "/path/to/file.bak", 0644, true},
		{"temp file", "/path/to/file.tmp", 0644, true},
		{"vim swap file", "/path/to/file.swp", 0644, true},
		{"tilde backup", "/path/to/file~", 0644, true},
		{"symlink", "/path/to/link", os.ModeSymlink | 0644, true},
		{"nil info", "/path/to/file", 0644, true}, // Will be handled by passing nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var info os.FileInfo
			if tt.name != "nil info" {
				info = &mockFileInfo{name: filepath.Base(tt.path), mode: tt.mode}
			}

			result := IgnoreFile(tt.path, info)
			if result != tt.expected {
				t.Errorf("IgnoreFile(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

type mockFileInfo struct {
	name string
	mode os.FileMode
}

func (mfi *mockFileInfo) Name() string       { return mfi.name }
func (mfi *mockFileInfo) Size() int64        { return 0 }
func (mfi *mockFileInfo) Mode() os.FileMode  { return mfi.mode }
func (mfi *mockFileInfo) ModTime() time.Time { return time.Now() }
func (mfi *mockFileInfo) IsDir() bool        { return mfi.mode.IsDir() }
func (mfi *mockFileInfo) Sys() any           { return nil }

func TestFileWatcherStartStop(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Test starting
	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}

	if !fw.IsRunning() {
		t.Error("File watcher should be running after start")
	}

	// Test starting again (should fail)
	if err := fw.Start(tempDir); err == nil {
		t.Error("Starting already running watcher should return error")
	}

	// Verify directories are being watched
	watchedDirs := fw.GetWatchedDirectories()
	if len(watchedDirs) == 0 {
		t.Error("Expected at least one directory to be watched")
	}

	// Test stopping
	if err := fw.Stop(); err != nil {
		t.Errorf("Failed to stop file watcher: %v", err)
	}

	if fw.IsRunning() {
		t.Error("File watcher should not be running after stop")
	}

	// Test stopping again (should fail)
	if err := fw.Stop(); err == nil {
		t.Error("Stopping already stopped watcher should return error")
	}
}

func TestFileWatcherInvalidStartPaths(t *testing.T) {
	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"non-existent path", "/non/existent/path"},
		{"file instead of directory", "/etc/passwd"}, // Assuming this exists and is a file
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fw.Start(tt.path)
			if err == nil {
				t.Errorf("Starting with %s should return error", tt.name)
				fw.Stop() // Clean up if it somehow started
			}
		})
	}
}

func TestFileModificationHandling(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Create event listener
	eventReceived := make(chan FileWatchEvent, 1)
	go func() {
		select {
		case event := <-fw.GetEventChannel():
			eventReceived <- event
		case <-time.After(5 * time.Second):
			close(eventReceived)
		}
	}()

	// Modify a file
	testFile := filepath.Join(tempDir, "file1.txt")
	time.Sleep(100 * time.Millisecond) // Give watcher time to set up

	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-eventReceived:
		if !ok {
			t.Error("Timeout waiting for file modification event")
			return
		}

		if event.Type != FileModified {
			t.Errorf("Expected FileModified event, got %v", event.Type)
		}

		if !strings.HasSuffix(event.Path, "file1.txt") {
			t.Errorf("Expected file1.txt in path, got %s", event.Path)
		}

		if event.IsDir {
			t.Error("File modification event should not be marked as directory")
		}

	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file modification event")
	}

	// Verify FileManager was called
	if len(fm.GetAllFiles()) == 0 {
		t.Error("FileManager.AddFile should have been called")
	}
}

func TestFileCreationHandling(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Create event listener
	eventReceived := make(chan FileWatchEvent, 1)
	go func() {
		select {
		case event := <-fw.GetEventChannel():
			eventReceived <- event
		case <-time.After(5 * time.Second):
			close(eventReceived)
		}
	}()

	// Create a new file
	newFile := filepath.Join(tempDir, "newfile.txt")
	time.Sleep(100 * time.Millisecond) // Give watcher time to set up

	if err := os.WriteFile(newFile, []byte("new file content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-eventReceived:
		if !ok {
			t.Error("Timeout waiting for file creation event")
			return
		}

		if event.Type != FileCreated {
			t.Errorf("Expected FileCreated event, got %v", event.Type)
		}

		if !strings.HasSuffix(event.Path, "newfile.txt") {
			t.Errorf("Expected newfile.txt in path, got %s", event.Path)
		}

	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file creation event")
	}
}

func TestDirectoryCreationHandling(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	initialWatchedCount := len(fw.GetWatchedDirectories())

	// Create event listener
	eventReceived := make(chan FileWatchEvent, 1)
	go func() {
		select {
		case event := <-fw.GetEventChannel():
			eventReceived <- event
		case <-time.After(5 * time.Second):
			close(eventReceived)
		}
	}()

	// Create a new directory
	newDir := filepath.Join(tempDir, "newdir")
	time.Sleep(100 * time.Millisecond) // Give watcher time to set up

	if err := os.Mkdir(newDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Wait for event
	select {
	case event, ok := <-eventReceived:
		if !ok {
			t.Error("Timeout waiting for directory creation event")
			return
		}

		if event.Type != DirCreated {
			t.Errorf("Expected DirCreated event, got %v", event.Type)
		}

		if !strings.HasSuffix(event.Path, "newdir") {
			t.Errorf("Expected newdir in path, got %s", event.Path)
		}

		if !event.IsDir {
			t.Error("Directory creation event should be marked as directory")
		}

	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for directory creation event")
	}

	// Verify new directory is being watched
	time.Sleep(100 * time.Millisecond) // Give time for watch to be added
	finalWatchedCount := len(fw.GetWatchedDirectories())
	if finalWatchedCount <= initialWatchedCount {
		t.Error("New directory should be added to watched directories")
	}
}

func TestFileWatchEventType_String(t *testing.T) {
	tests := []struct {
		eventType FileWatchEventType
		expected  string
	}{
		{FileCreated, "FileCreated"},
		{FileModified, "FileModified"},
		{FileDeleted, "FileDeleted"},
		{FileRenamed, "FileRenamed"},
		{DirCreated, "DirCreated"},
		{DirDeleted, "DirDeleted"},
		{FileWatchEventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.eventType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetRelativePath(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Test without starting (no root path set)
	_, err = fw.getRelativePath("/some/path")
	if err == nil {
		t.Error("Expected error when root path not set")
	}

	// Start the watcher to set root path
	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Test with valid paths
	testFile := filepath.Join(tempDir, "subdir", "file3.txt")
	relPath, err := fw.getRelativePath(testFile)
	if err != nil {
		t.Errorf("Unexpected error getting relative path: %v", err)
	}

	expected := filepath.Join("subdir", "file3.txt")
	if relPath != expected {
		t.Errorf("Expected %s, got %s", expected, relPath)
	}
}

func TestConcurrentOperations(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Start multiple goroutines that perform operations
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create and modify files concurrently
			fileName := fmt.Sprintf("concurrent_file_%d.txt", id)
			filePath := filepath.Join(tempDir, fileName)

			// Create file
			if err := os.WriteFile(filePath, []byte("initial"), 0644); err != nil {
				t.Errorf("Failed to create file %s: %v", fileName, err)
				return
			}

			time.Sleep(10 * time.Millisecond)

			// Modify file
			if err := os.WriteFile(filePath, []byte("modified"), 0644); err != nil {
				t.Errorf("Failed to modify file %s: %v", fileName, err)
				return
			}

			// Check running status
			_ = fw.IsRunning()

			// Get watched directories
			_ = fw.GetWatchedDirectories()
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the watcher is still running and functional
	if !fw.IsRunning() {
		t.Error("File watcher should still be running after concurrent operations")
	}
}

func TestEventChannelHandling(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Get the event channel
	eventChan := fw.GetEventChannel()
	if eventChan == nil {
		t.Fatal("Event channel should not be nil")
	}

	// Create multiple files quickly to test channel capacity
	var events []FileWatchEvent
	done := make(chan bool)

	go func() {
		timeout := time.After(3 * time.Second)
		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					done <- true
					return
				}
				events = append(events, event)
				if len(events) >= 5 { // Expect at least 5 events
					done <- true
					return
				}
			case <-timeout:
				done <- true
				return
			}
		}
	}()

	time.Sleep(100 * time.Millisecond) // Give watcher time to set up

	// Create multiple files
	for i := range 5 {
		fileName := fmt.Sprintf("test_file_%d.txt", i)
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Errorf("Failed to create file %s: %v", fileName, err)
		}
		time.Sleep(50 * time.Millisecond) // Small delay between file creations
	}

	<-done

	if len(events) == 0 {
		t.Error("Expected to receive events but got none")
	}

	// Verify event details
	for _, event := range events {
		if event.Time.IsZero() {
			t.Error("Event time should be set")
		}
		if event.Path == "" {
			t.Error("Event path should not be empty")
		}
	}
}

func TestErrorHandling(t *testing.T) {
	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	// Test handling of non-existent directory
	nonExistentDir := "/this/directory/does/not/exist"
	err = fw.Start(nonExistentDir)
	if err == nil {
		t.Error("Expected error when starting with non-existent directory")
	}

	// Test stopping non-running watcher
	err = fw.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running watcher")
	}
}

func TestIgnoreHiddenFiles(t *testing.T) {
	tempDir := createTestDir(t)
	defer os.RemoveAll(tempDir)

	// Create hidden files
	hiddenFiles := []string{
		".hidden_file.txt",
		".git/config",
		"regular_file.txt",
		"file.bak",
		"file.tmp",
	}

	for _, file := range hiddenFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	fm := NewFileManager("/test/site")
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fw.Start(tempDir); err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Collect events for a short time
	var events []FileWatchEvent
	done := make(chan bool)

	go func() {
		timeout := time.After(1 * time.Second)
		for {
			select {
			case event := <-fw.GetEventChannel():
				events = append(events, event)
			case <-timeout:
				done <- true
				return
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Modify all files
	for _, file := range hiddenFiles {
		fullPath := filepath.Join(tempDir, file)
		if err := os.WriteFile(fullPath, []byte("modified"), 0644); err != nil {
			t.Logf("Failed to modify %s: %v", file, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	<-done

	// Check that only regular files generated events
	regularFileEvents := 0
	for _, event := range events {
		if strings.Contains(event.Path, "regular_file.txt") {
			regularFileEvents++
		}
		if strings.Contains(event.Path, ".hidden") ||
			strings.Contains(event.Path, ".git") ||
			strings.Contains(event.Path, ".bak") ||
			strings.Contains(event.Path, ".tmp") {
			t.Errorf("Should not receive events for ignored file: %s", event.Path)
		}
	}

	if regularFileEvents == 0 {
		t.Error("Should receive events for regular files")
	}
}
