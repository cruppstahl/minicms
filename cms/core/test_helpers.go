package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFileBuilder helps create test files with various properties
type TestFileBuilder struct {
	path     string
	content  string
	routes   []string
	mimeType string
	metadata map[string]interface{}
}

// NewTestFileBuilder creates a new test file builder
func NewTestFileBuilder(path string) *TestFileBuilder {
	return &TestFileBuilder{
		path:     path,
		content:  "",
		routes:   make([]string, 0),
		metadata: make(map[string]interface{}),
	}
}

// WithContent sets the file content
func (tfb *TestFileBuilder) WithContent(content string) *TestFileBuilder {
	tfb.content = content
	return tfb
}

// WithRoute adds a route to the file
func (tfb *TestFileBuilder) WithRoute(route string) *TestFileBuilder {
	tfb.routes = append(tfb.routes, route)
	return tfb
}

// WithRoutes sets multiple routes for the file
func (tfb *TestFileBuilder) WithRoutes(routes []string) *TestFileBuilder {
	tfb.routes = routes
	return tfb
}

// WithMimeType sets the MIME type
func (tfb *TestFileBuilder) WithMimeType(mimeType string) *TestFileBuilder {
	tfb.mimeType = mimeType
	return tfb
}

// WithMetadata adds metadata to the file
func (tfb *TestFileBuilder) WithMetadata(key string, value interface{}) *TestFileBuilder {
	tfb.metadata[key] = value
	return tfb
}

// Build creates the File object
func (tfb *TestFileBuilder) Build() *File {
	file := &File{
		Name:         filepath.Base(tfb.path),
		Path:         tfb.path,
		Content:      []byte(tfb.content),
		Routes:       tfb.routes,
		Dependencies: make(map[string]*File),
		Dependents:   make(map[string]*File),
		Metadata: FileMetadata{
			MimeType: tfb.mimeType,
		},
	}

	// Add custom metadata if any
	for key, value := range tfb.metadata {
		switch key {
		case "title":
			if title, ok := value.(string); ok {
				file.Metadata.Title = title
			}
		case "redirectUrl":
			if url, ok := value.(string); ok {
				file.Metadata.RedirectUrl = url
			}
		}
	}

	return file
}

// CreatePhysically creates the file on disk in the given base directory
func (tfb *TestFileBuilder) CreatePhysically(t *testing.T, baseDir string) string {
	fullPath := filepath.Join(baseDir, tfb.path)
	dir := filepath.Dir(fullPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// Write file content
	if err := os.WriteFile(fullPath, []byte(tfb.content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", fullPath, err)
	}

	return fullPath
}

// TestDirectoryStructure helps create complex directory structures for testing
type TestDirectoryStructure struct {
	baseDir string
	files   []*TestFileBuilder
	dirs    []string
}

// NewTestDirectoryStructure creates a new test directory structure builder
func NewTestDirectoryStructure(baseDir string) *TestDirectoryStructure {
	return &TestDirectoryStructure{
		baseDir: baseDir,
		files:   make([]*TestFileBuilder, 0),
		dirs:    make([]string, 0),
	}
}

// WithDirectory adds a directory to be created
func (tds *TestDirectoryStructure) WithDirectory(dirPath string) *TestDirectoryStructure {
	tds.dirs = append(tds.dirs, dirPath)
	return tds
}

// WithFile adds a file to be created
func (tds *TestDirectoryStructure) WithFile(file *TestFileBuilder) *TestDirectoryStructure {
	tds.files = append(tds.files, file)
	return tds
}

// Create creates the entire directory structure on disk
func (tds *TestDirectoryStructure) Create(t *testing.T) []string {
	createdFiles := make([]string, 0)

	// Create directories
	for _, dir := range tds.dirs {
		fullPath := filepath.Join(tds.baseDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", fullPath, err)
		}
	}

	// Create files
	for _, file := range tds.files {
		fullPath := file.CreatePhysically(t, tds.baseDir)
		createdFiles = append(createdFiles, fullPath)
	}

	return createdFiles
}

// EventCollector helps collect and verify file watch events during tests
type EventCollector struct {
	events   []FileWatchEvent
	eventCh  <-chan FileWatchEvent
	stopCh   chan bool
	timeout  time.Duration
	stopping bool
}

// NewEventCollector creates a new event collector
func NewEventCollector(eventCh <-chan FileWatchEvent, timeout time.Duration) *EventCollector {
	return &EventCollector{
		events:  make([]FileWatchEvent, 0),
		eventCh: eventCh,
		stopCh:  make(chan bool),
		timeout: timeout,
	}
}

// Start begins collecting events
func (ec *EventCollector) Start() {
	go ec.collectEvents()
}

// Stop stops collecting events and returns collected events
func (ec *EventCollector) Stop() []FileWatchEvent {
	if !ec.stopping {
		ec.stopping = true
		close(ec.stopCh)
	}
	return ec.events
}

// collectEvents runs in a goroutine to collect events
func (ec *EventCollector) collectEvents() {
	timeoutTimer := time.After(ec.timeout)

	for {
		select {
		case event, ok := <-ec.eventCh:
			if !ok {
				return
			}
			ec.events = append(ec.events, event)
		case <-ec.stopCh:
			return
		case <-timeoutTimer:
			return
		}
	}
}

// WaitForEvents waits for a specific number of events or timeout
func (ec *EventCollector) WaitForEvents(count int) []FileWatchEvent {
	deadline := time.Now().Add(ec.timeout)

	for time.Now().Before(deadline) {
		if len(ec.events) >= count {
			return ec.events[:count]
		}
		time.Sleep(10 * time.Millisecond)
	}

	return ec.events
}

// GetEventsOfType returns events of a specific type
func (ec *EventCollector) GetEventsOfType(eventType FileWatchEventType) []FileWatchEvent {
	filtered := make([]FileWatchEvent, 0)
	for _, event := range ec.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventsForPath returns events for a specific path
func (ec *EventCollector) GetEventsForPath(path string) []FileWatchEvent {
	filtered := make([]FileWatchEvent, 0)
	for _, event := range ec.events {
		if event.Path == path {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// TestEnvironment provides a complete test environment setup
type TestEnvironment struct {
	T           *testing.T
	TempDir     string
	FileManager *FileManager
	FileWatcher *FileWatcher
	Router      *RouterManager
	Listener    *FileWatcherListener
	Context     *Context
}

// NewTestEnvironment creates a complete test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	tempDir := t.TempDir()

	// Create standard directory structure
	standardDirs := []string{
		"content",
		"content/posts",
		"assets",
		"config",
		"layout",
	}

	for _, dir := range standardDirs {
		fullPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Initialize components
	fm := NewFileManager(tempDir)
	fw, err := NewFileWatcher(fm)
	if err != nil {
		t.Fatalf("Failed to create FileWatcher: %v", err)
	}

	rm := NewRouterManager()
	fw.SetRouter(rm)

	// Create basic config
	config := Config{
		SiteDirectory: tempDir,
		Server: Server{
			Port:     8080,
			Hostname: "localhost",
		},
	}

	ctx := &Context{
		Config:      config,
		FileManager: fm,
		FileWatcher: fw,
	}

	if err := rm.InitializeRouter(ctx); err != nil {
		t.Fatalf("Failed to initialize router: %v", err)
	}

	return &TestEnvironment{
		T:           t,
		TempDir:     tempDir,
		FileManager: fm,
		FileWatcher: fw,
		Router:      rm,
		Context:     ctx,
	}
}

// Start starts the file watcher and listener
func (te *TestEnvironment) Start() *TestEnvironment {
	if err := te.FileWatcher.Start(te.TempDir); err != nil {
		te.T.Fatalf("Failed to start FileWatcher: %v", err)
	}

	listener, err := RegisterFileWatcherListener(te.FileWatcher)
	if err != nil {
		te.T.Fatalf("Failed to register listener: %v", err)
	}

	te.Listener = listener
	return te
}

// Stop stops all components
func (te *TestEnvironment) Stop() {
	if te.Listener != nil {
		te.Listener.Stop()
	}
	if te.FileWatcher != nil {
		te.FileWatcher.Stop()
	}
}

// CreateFile creates a file using TestFileBuilder
func (te *TestEnvironment) CreateFile(path string) *TestFileBuilder {
	return NewTestFileBuilder(path)
}

// CreateDirectory creates a directory structure using TestDirectoryStructure
func (te *TestEnvironment) CreateDirectory() *TestDirectoryStructure {
	return NewTestDirectoryStructure(te.TempDir)
}

// CreateEventCollector creates an event collector for the FileWatcher
func (te *TestEnvironment) CreateEventCollector(timeout time.Duration) *EventCollector {
	return NewEventCollector(te.FileWatcher.GetEventChannel(), timeout)
}

// WaitForProcessing waits for file processing to complete
func (te *TestEnvironment) WaitForProcessing(duration time.Duration) {
	time.Sleep(duration)
	te.FileManager.ProcessUpdatedFiles()
}

// AssertFileExists asserts that a file exists in the FileManager
func (te *TestEnvironment) AssertFileExists(filePath string) *File {
	file := te.FileManager.GetFile(filePath)
	if file == nil {
		te.T.Errorf("Expected file %s to exist in FileManager", filePath)
	}
	return file
}

// AssertFileNotExists asserts that a file does not exist in the FileManager
func (te *TestEnvironment) AssertFileNotExists(filePath string) {
	file := te.FileManager.GetFile(filePath)
	if file != nil {
		te.T.Errorf("Expected file %s not to exist in FileManager", filePath)
	}
}

// AssertRouteExists asserts that a route exists in the RouterManager
func (te *TestEnvironment) AssertRouteExists(routePath string) {
	routes := te.Router.GetAllRoutes()
	found := false
	for route := range routes {
		if route == routePath {
			found = true
			break
		}
	}
	if !found {
		te.T.Errorf("Expected route %s to exist", routePath)
	}
}

// AssertRouteNotExists asserts that a route does not exist in the RouterManager
func (te *TestEnvironment) AssertRouteNotExists(routePath string) {
	routes := te.Router.GetAllRoutes()
	for route := range routes {
		if route == routePath {
			te.T.Errorf("Expected route %s not to exist", routePath)
			break
		}
	}
}

// GetRouteForFile returns the route that maps to a specific file
func (te *TestEnvironment) GetRouteForFile(filePath string) string {
	routes := te.Router.GetAllRoutes()
	for route, fp := range routes {
		if fp == filePath {
			return route
		}
	}
	return ""
}

// MockPlugin provides a simple plugin implementation for testing
type MockPlugin struct {
	name          string
	priority      int
	processFunc   func(ctx *PluginContext) *PluginResult
	canProcessFunc func(file *File) bool
	callCount     int
	processedFiles []string
}

// NewMockPlugin creates a new mock plugin
func NewMockPlugin(name string, priority int) *MockPlugin {
	return &MockPlugin{
		name:           name,
		priority:       priority,
		processedFiles: make([]string, 0),
		// Default implementations
		processFunc: func(ctx *PluginContext) *PluginResult {
			// Default: just return success without changes
			return &PluginResult{
				Success: true,
				Error:   nil,
			}
		},
		canProcessFunc: func(file *File) bool {
			// Default: can process all files
			return true
		},
	}
}

// WithProcessFunc sets a custom process function
func (mp *MockPlugin) WithProcessFunc(fn func(ctx *PluginContext) *PluginResult) *MockPlugin {
	mp.processFunc = fn
	return mp
}

// WithCanProcessFunc sets a custom can process function
func (mp *MockPlugin) WithCanProcessFunc(fn func(file *File) bool) *MockPlugin {
	mp.canProcessFunc = fn
	return mp
}

// Plugin interface implementation
func (mp *MockPlugin) Name() string {
	return mp.name
}

func (mp *MockPlugin) Priority() int {
	return mp.priority
}

func (mp *MockPlugin) CanProcess(file *File) bool {
	return mp.canProcessFunc(file)
}

func (mp *MockPlugin) Process(ctx *PluginContext) *PluginResult {
	mp.callCount++
	mp.processedFiles = append(mp.processedFiles, ctx.File.Path)
	return mp.processFunc(ctx)
}

// Test helper methods
func (mp *MockPlugin) GetCallCount() int {
	return mp.callCount
}

func (mp *MockPlugin) GetProcessedFiles() []string {
	return mp.processedFiles
}

func (mp *MockPlugin) Reset() {
	mp.callCount = 0
	mp.processedFiles = make([]string, 0)
}

// FileOperationSequence helps test sequences of file operations
type FileOperationSequence struct {
	env        *TestEnvironment
	operations []func() error
}

// NewFileOperationSequence creates a new file operation sequence
func NewFileOperationSequence(env *TestEnvironment) *FileOperationSequence {
	return &FileOperationSequence{
		env:        env,
		operations: make([]func() error, 0),
	}
}

// CreateFile adds a file creation operation to the sequence
func (fos *FileOperationSequence) CreateFile(filePath, content string) *FileOperationSequence {
	op := func() error {
		fullPath := filepath.Join(fos.env.TempDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0644)
	}
	fos.operations = append(fos.operations, op)
	return fos
}

// ModifyFile adds a file modification operation to the sequence
func (fos *FileOperationSequence) ModifyFile(filePath, newContent string) *FileOperationSequence {
	op := func() error {
		fullPath := filepath.Join(fos.env.TempDir, filePath)
		return os.WriteFile(fullPath, []byte(newContent), 0644)
	}
	fos.operations = append(fos.operations, op)
	return fos
}

// DeleteFile adds a file deletion operation to the sequence
func (fos *FileOperationSequence) DeleteFile(filePath string) *FileOperationSequence {
	op := func() error {
		fullPath := filepath.Join(fos.env.TempDir, filePath)
		return os.Remove(fullPath)
	}
	fos.operations = append(fos.operations, op)
	return fos
}

// CreateDirectory adds a directory creation operation to the sequence
func (fos *FileOperationSequence) CreateDirectory(dirPath string) *FileOperationSequence {
	op := func() error {
		fullPath := filepath.Join(fos.env.TempDir, dirPath)
		return os.MkdirAll(fullPath, 0755)
	}
	fos.operations = append(fos.operations, op)
	return fos
}

// Wait adds a wait operation to the sequence
func (fos *FileOperationSequence) Wait(duration time.Duration) *FileOperationSequence {
	op := func() error {
		time.Sleep(duration)
		return nil
	}
	fos.operations = append(fos.operations, op)
	return fos
}

// Execute runs all operations in sequence
func (fos *FileOperationSequence) Execute() error {
	for i, op := range fos.operations {
		if err := op(); err != nil {
			return err
		}
		// Small delay between operations to ensure file system events are properly ordered
		if i < len(fos.operations)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return nil
}

// ExecuteWithEventCollection runs operations and collects events
func (fos *FileOperationSequence) ExecuteWithEventCollection(timeout time.Duration) ([]FileWatchEvent, error) {
	collector := fos.env.CreateEventCollector(timeout)
	collector.Start()
	defer collector.Stop()

	err := fos.Execute()

	// Wait a bit for all events to be processed
	time.Sleep(200 * time.Millisecond)

	events := collector.Stop()
	return events, err
}