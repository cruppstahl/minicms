package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewFileManager(t *testing.T) {
	siteDir := "/test/site"
	fm := NewFileManager(siteDir)

	if fm == nil {
		t.Fatal("NewFileManager returned nil")
	}

	if fm.SiteDirectory != siteDir {
		t.Errorf("Expected SiteDirectory %s, got %s", siteDir, fm.SiteDirectory)
	}

	if fm.Files == nil {
		t.Error("Files map is nil")
	}

	root := fm.GetRoot()
	if root == nil {
		t.Error("Root directory is nil")
	}

	if root.Name != "" {
		t.Errorf("Expected root name to be empty, got %s", root.Name)
	}

	if root.Parent != nil {
		t.Error("Root should have no parent")
	}
}

func TestFileNeedsUpdate(t *testing.T) {
	file := &File{
		Name:         "test.txt",
		Path:         "test.txt",
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	// File with nil content needs update
	if !file.NeedsUpdate() {
		t.Error("File with nil content should need update")
	}

	// File with content doesn't need update
	file.Content = []byte("content")
	if file.NeedsUpdate() {
		t.Error("File with content should not need update")
	}
}

func TestFileReadFile(t *testing.T) {
	// Create temporary directory and file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file := &File{
		Name: "test.txt",
		Path: "test.txt",
	}

	content := file.ReadFile(tempDir)
	if content == nil {
		t.Error("ReadFile returned nil for existing file")
	}

	if string(content) != string(testContent) {
		t.Errorf("Expected content %s, got %s", testContent, content)
	}

	// Test non-existent file
	file.Path = "nonexistent.txt"
	content = file.ReadFile(tempDir)
	if content != nil {
		t.Error("ReadFile should return nil for non-existent file")
	}
}

func TestFileAddDependency(t *testing.T) {
	file1 := &File{
		Name:         "file1.txt",
		Path:         "file1.txt",
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file2 := &File{
		Name:         "file2.txt",
		Path:         "file2.txt",
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file1.AddDependency(file2)

	// Check that file1 depends on file2
	if dep, exists := file1.Dependencies[file2.Path]; !exists || dep != file2 {
		t.Error("Dependency not added correctly")
	}

	// Check that file2 has file1 as dependent
	if dep, exists := file2.Dependents[file1.Path]; !exists || dep != file1 {
		t.Error("Dependent not added correctly")
	}
}

func TestFileMarkForUpdate(t *testing.T) {
	// Create a dependency chain: file1 -> file2 -> file3
	file1 := &File{
		Name:         "file1.txt",
		Path:         "file1.txt",
		Content:      []byte("content1"),
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file2 := &File{
		Name:         "file2.txt",
		Path:         "file2.txt",
		Content:      []byte("content2"),
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file3 := &File{
		Name:         "file3.txt",
		Path:         "file3.txt",
		Content:      []byte("content3"),
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file2.AddDependency(file1)
	file3.AddDependency(file2)

	// Mark file1 for update
	file1.MarkForUpdate()

	// All files should be marked for update
	if file1.Content != nil {
		t.Error("file1 should be marked for update")
	}
	if file2.Content != nil {
		t.Error("file2 should be marked for update (dependent)")
	}
	if file3.Content != nil {
		t.Error("file3 should be marked for update (transitive dependent)")
	}
}

func TestFileMarkForUpdateCircularDependency(t *testing.T) {
	// Create circular dependency: file1 -> file2 -> file1
	file1 := &File{
		Name:         "file1.txt",
		Path:         "file1.txt",
		Content:      []byte("content1"),
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file2 := &File{
		Name:         "file2.txt",
		Path:         "file2.txt",
		Content:      []byte("content2"),
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
	}

	file1.AddDependency(file2)
	file2.AddDependency(file1)

	// This should not cause infinite recursion
	file1.MarkForUpdate()

	if file1.Content != nil {
		t.Error("file1 should be marked for update")
	}
	if file2.Content != nil {
		t.Error("file2 should be marked for update")
	}
}

func TestFileManagerAddFile(t *testing.T) {
	fm := NewFileManager("/test")

	// Add file to root
	file := fm.AddFile("test.txt")
	if file == nil {
		t.Fatal("AddFile returned nil")
	}

	if file.Name != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", file.Name)
	}

	if file.Path != "test.txt" {
		t.Errorf("Expected path test.txt, got %s", file.Path)
	}

	// Check file is in Files map
	retrievedFile := fm.GetFile("test.txt")
	if retrievedFile != file {
		t.Error("File not properly stored in Files map")
	}

	// Check file is in parent directory
	root := fm.GetRoot()
	if root.Files["test.txt"] != file {
		t.Error("File not properly stored in parent directory")
	}
}

func TestFileManagerAddFileInSubdirectory(t *testing.T) {
	fm := NewFileManager("/test")

	// Create directory structure first
	fm.createDirectory("subdir")

	// Add file to subdirectory
	file := fm.AddFile("subdir/test.txt")
	if file == nil {
		t.Fatal("AddFile returned nil")
	}

	if file.Name != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", file.Name)
	}

	if file.Path != "subdir/test.txt" {
		t.Errorf("Expected path subdir/test.txt, got %s", file.Path)
	}

	// Check file is in correct directory
	subdir := fm.GetDirectory("subdir")
	if subdir == nil {
		t.Fatal("Subdirectory not found")
	}

	if subdir.Files["test.txt"] != file {
		t.Error("File not properly stored in subdirectory")
	}
}

func TestFileManagerGetFile(t *testing.T) {
	fm := NewFileManager("/test")

	// Test non-existent file
	file := fm.GetFile("nonexistent.txt")
	if file != nil {
		t.Error("GetFile should return nil for non-existent file")
	}

	// Add and retrieve file
	addedFile := fm.AddFile("test.txt")
	retrievedFile := fm.GetFile("test.txt")

	if retrievedFile != addedFile {
		t.Error("GetFile returned different file instance")
	}
}

func TestFileManagerGetDirectory(t *testing.T) {
	fm := NewFileManager("/test")

	// Test root directory
	root := fm.GetDirectory("")
	if root != fm.GetRoot() {
		t.Error("GetDirectory(\"\") should return root")
	}

	root = fm.GetDirectory(".")
	if root != fm.GetRoot() {
		t.Error("GetDirectory(\".\") should return root")
	}

	// Test non-existent directory
	dir := fm.GetDirectory("nonexistent")
	if dir != nil {
		t.Error("GetDirectory should return nil for non-existent directory")
	}

	// Create and retrieve directory
	created := fm.createDirectory("testdir")
	retrieved := fm.GetDirectory("testdir")

	if retrieved != created {
		t.Error("GetDirectory returned different directory instance")
	}
}

func TestFileManagerCreateDirectory(t *testing.T) {
	fm := NewFileManager("/test")

	// Test creating nested directories
	dir := fm.createDirectory("level1/level2/level3")
	if dir == nil {
		t.Fatal("createDirectory returned nil")
	}

	if dir.Name != "level3" {
		t.Errorf("Expected name level3, got %s", dir.Name)
	}

	if dir.Path != "level1/level2/level3" {
		t.Errorf("Expected path level1/level2/level3, got %s", dir.Path)
	}

	// Check parent relationships
	if dir.Parent.Name != "level2" {
		t.Error("Incorrect parent relationship")
	}

	// Check all levels exist
	level1 := fm.GetDirectory("level1")
	level2 := fm.GetDirectory("level1/level2")
	level3 := fm.GetDirectory("level1/level2/level3")

	if level1 == nil || level2 == nil || level3 == nil {
		t.Error("Not all directory levels were created")
	}

	if level3 != dir {
		t.Error("Final directory doesn't match returned directory")
	}
}

func TestFileManagerWalkDirectory(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create subdirectory and files
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create files
	err = os.WriteFile(filepath.Join(tempDir, "root.txt"), []byte("root content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	err = os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("sub content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Create hidden file (should be ignored)
	err = os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden"), 0644)
	if err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	fm := NewFileManager(tempDir)
	err = fm.WalkDirectory(".")
	if err != nil {
		t.Fatalf("WalkDirectory failed: %v", err)
	}

	// Check that files were added
	rootFile := fm.GetFile("root.txt")
	if rootFile == nil {
		t.Error("root.txt not found")
	}

	subFile := fm.GetFile("subdir/sub.txt")
	if subFile == nil {
		t.Error("subdir/sub.txt not found")
	}

	// Check that hidden file was ignored
	hiddenFile := fm.GetFile(".hidden")
	if hiddenFile != nil {
		t.Error("Hidden file should be ignored")
	}

	// Check directory structure
	subdir := fm.GetDirectory("subdir")
	if subdir == nil {
		t.Error("subdir directory not found")
	}

	if subdir.Files["sub.txt"] != subFile {
		t.Error("sub.txt not properly linked to subdirectory")
	}
}

func TestFileManagerConcurrency(t *testing.T) {
	fm := NewFileManager("/test")

	const numGoroutines = 10
	const numFiles = 5

	var wg sync.WaitGroup

	// Test concurrent file additions
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range numFiles {
				fileName := fmt.Sprintf("file_%d_%d.txt", id, j)
				file := fm.AddFile(fileName)
				if file == nil {
					t.Errorf("Failed to add file %s", fileName)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all files were added
	allFiles := fm.GetAllFiles()
	expectedCount := numGoroutines * numFiles
	if len(allFiles) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(allFiles))
	}

	// Test concurrent processing
	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fm.ProcessAllFiles()
		}()
	}

	wg.Wait()

	// All files should be processed
	/*
		for path, file := range fm.GetAllFiles() {
			if file.NeedsUpdate() {
				t.Errorf("File %s still needs update after processing", path)
			}
		}
	*/
}

func TestFileManagerRaceCondition(t *testing.T) {
	fm := NewFileManager("/test")

	// Add initial files
	for i := range 100 {
		fm.AddFile(fmt.Sprintf("file_%d.txt", i))
	}

	var wg sync.WaitGroup
	done := make(chan bool)

	// Start concurrent readers
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = fm.GetAllFiles()
					files := fm.GetAllFiles()
					for _, file := range files {
						_ = file.NeedsUpdate()
					}
					time.Sleep(time.Microsecond)
				}
			}
		}()
	}

	// Start concurrent writers
	for i := range 3 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				select {
				case <-done:
					return
				default:
					fileName := fmt.Sprintf("new_file_%d_%d.txt", id, j)
					fm.AddFile(fileName)
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	// Start file processors
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					fm.ProcessUpdatedFiles()
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Let it run for a short time
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()

	// If we get here without panicking, the race condition test passed
}

func TestGetAllFiles(t *testing.T) {
	fm := NewFileManager("/test")

	// Initially should be empty
	files := fm.GetAllFiles()
	if len(files) != 0 {
		t.Errorf("Expected 0 files initially, got %d", len(files))
	}

	// Add some files
	fm.AddFile("file1.txt")
	fm.AddFile("file2.txt")
	//	fm.AddFile("subdir/file2.txt") TODO panics!

	files = fm.GetAllFiles()
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check that we get copies, not references to internal state
	delete(files, "file1.txt")

	filesAgain := fm.GetAllFiles()
	if len(filesAgain) != 2 {
		t.Error("GetAllFiles should return a copy, not a reference")
	}
}

// Benchmark tests
func BenchmarkFileManagerAddFile(b *testing.B) {
	fm := NewFileManager("/test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileName := fmt.Sprintf("file_%d.txt", i)
		fm.AddFile(fileName)
	}
}

func BenchmarkFileManagerGetFile(b *testing.B) {
	fm := NewFileManager("/test")

	// Pre-populate with files
	for i := range 1000 {
		fileName := fmt.Sprintf("file_%d.txt", i)
		fm.AddFile(fileName)
	}

	b.ResetTimer()
	for i := range b.N {
		fileName := fmt.Sprintf("file_%d.txt", i%1000)
		fm.GetFile(fileName)
	}
}

func BenchmarkFileManagerProcessAllFiles(b *testing.B) {
	fm := NewFileManager("/test")

	// Pre-populate with files
	for i := range 100 {
		fileName := fmt.Sprintf("file_%d.txt", i)
		fm.AddFile(fileName)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.ProcessAllFiles()
	}
}
