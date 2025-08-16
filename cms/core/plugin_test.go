package core

import (
	"fmt"
	"sync"
	"testing"
)

// Mock plugin for testing
type mockPlugin struct {
	name         string
	priority     int
	canProcess   bool
	shouldModify bool
	shouldError  bool
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) CanProcess(file *File) bool {
	return m.canProcess
}

func (m *mockPlugin) Process(ctx *PluginContext) *PluginResult {
	if m.shouldError {
		return &PluginResult{
			Success: false,
			Error:   fmt.Errorf("mock error"),
		}
	}

	result := &PluginResult{
		Success: true,
	}

	if m.shouldModify {
		result.Modified = true
		result.NewContent = []byte("modified content")
		result.MimeType = "text/plain"
		result.Routes = []string{"/test"}
	}

	return result
}

func (m *mockPlugin) Priority() int {
	return m.priority
}

func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()
	if pm == nil {
		t.Fatal("NewPluginManager returned nil")
	}
	if pm.plugins == nil {
		t.Fatal("plugins slice not initialized")
	}
}

func TestRegisterPlugin(t *testing.T) {
	pm := NewPluginManager()

	plugin1 := &mockPlugin{name: "plugin1", priority: 10}
	plugin2 := &mockPlugin{name: "plugin2", priority: 5}
	plugin3 := &mockPlugin{name: "plugin3", priority: 15}

	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)
	pm.RegisterPlugin(plugin3)

	// Check that plugins are sorted by priority
	if len(pm.plugins) != 3 {
		t.Fatalf("Expected 3 plugins, got %d", len(pm.plugins))
	}

	if pm.plugins[0].Priority() != 5 {
		t.Errorf("Expected first plugin priority 5, got %d", pm.plugins[0].Priority())
	}
	if pm.plugins[1].Priority() != 10 {
		t.Errorf("Expected second plugin priority 10, got %d", pm.plugins[1].Priority())
	}
	if pm.plugins[2].Priority() != 15 {
		t.Errorf("Expected third plugin priority 15, got %d", pm.plugins[2].Priority())
	}
}

func TestGetPluginsForFile(t *testing.T) {
	pm := NewPluginManager()

	plugin1 := &mockPlugin{name: "plugin1", canProcess: true}
	plugin2 := &mockPlugin{name: "plugin2", canProcess: false}
	plugin3 := &mockPlugin{name: "plugin3", canProcess: true}

	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)
	pm.RegisterPlugin(plugin3)

	file := &File{Path: "test.txt"}
	matchingPlugins := pm.GetPluginsForFile(file)

	if len(matchingPlugins) != 2 {
		t.Fatalf("Expected 2 matching plugins, got %d", len(matchingPlugins))
	}

	// Verify the correct plugins were returned
	names := make(map[string]bool)
	for _, p := range matchingPlugins {
		names[p.Name()] = true
	}

	if !names["plugin1"] || !names["plugin3"] {
		t.Error("Wrong plugins returned")
	}
}

func TestListPlugins(t *testing.T) {
	pm := NewPluginManager()

	plugin1 := &mockPlugin{name: "plugin1", priority: 10}
	plugin2 := &mockPlugin{name: "plugin2", priority: 5}

	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)

	list := pm.ListPlugins()
	if len(list) != 2 {
		t.Fatalf("Expected 2 plugins in list, got %d", len(list))
	}

	// Check format
	expected1 := "plugin2 (priority: 5)"
	expected2 := "plugin1 (priority: 10)"

	if list[0] != expected1 || list[1] != expected2 {
		t.Errorf("Plugin list format incorrect. Got: %v", list)
	}
}

func TestProcessFile(t *testing.T) {
	pm := NewPluginManager()

	plugin := &mockPlugin{
		name:         "test-plugin",
		priority:     10,
		canProcess:   true,
		shouldModify: true,
	}
	pm.RegisterPlugin(plugin)

	originalFile := &File{
		Path:     "test.txt",
		Content:  []byte("original content"),
		Metadata: FileMetadata{},
	}

	fm := &FileManager{SiteDirectory: "/test"}

	result := pm.Process(*originalFile, fm)

	if result == nil {
		t.Fatal("Process returned nil")
	}

	if string(result.Content) != "modified content" {
		t.Errorf("Expected modified content, got: %s", string(result.Content))
	}

	if result.Metadata.MimeType != "text/plain" {
		t.Errorf("Expected mime type text/plain, got: %s", result.Metadata.MimeType)
	}

	if len(result.Routes) != 1 || result.Routes[0] != "/test" {
		t.Errorf("Expected routes [/test], got: %v", result.Routes)
	}
}

func TestProcessFileWithError(t *testing.T) {
	pm := NewPluginManager()

	plugin := &mockPlugin{
		name:        "error-plugin",
		priority:    10,
		canProcess:  true,
		shouldError: true,
	}
	pm.RegisterPlugin(plugin)

	originalFile := &File{
		Path:    "test.txt",
		Content: []byte("original content"),
	}

	fm := &FileManager{SiteDirectory: "/test"}

	result := pm.Process(*originalFile, fm)

	// Should still return a result even with plugin errors
	if result == nil {
		t.Fatal("Process returned nil")
	}

	// Original content should be unchanged due to error
	if string(result.Content) != "original content" {
		t.Errorf("Content should be unchanged on error, got: %s", string(result.Content))
	}
}

func TestConcurrentAccess(t *testing.T) {
	pm := NewPluginManager()

	// Test concurrent registration and access
	var wg sync.WaitGroup
	const numGoroutines = 10

	// Concurrent registration
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			plugin := &mockPlugin{
				name:       fmt.Sprintf("plugin-%d", id),
				priority:   id,
				canProcess: true,
			}
			pm.RegisterPlugin(plugin)
		}(i)
	}

	// Concurrent reads
	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file := &File{Path: "test.txt"}
			pm.GetPluginsForFile(file)
			pm.ListPlugins()
		}()
	}

	wg.Wait()

	// Verify all plugins were registered
	if len(pm.plugins) != numGoroutines {
		t.Errorf("Expected %d plugins, got %d", numGoroutines, len(pm.plugins))
	}
}

func TestPluginPrioritySorting(t *testing.T) {
	pm := NewPluginManager()

	// Add plugins in random order
	priorities := []int{100, 1, 50, 25, 75}
	for _, p := range priorities {
		plugin := &mockPlugin{
			name:     fmt.Sprintf("plugin-%d", p),
			priority: p,
		}
		pm.RegisterPlugin(plugin)
	}

	// Verify they're sorted correctly
	expectedOrder := []int{1, 25, 50, 75, 100}
	for i, expectedPriority := range expectedOrder {
		if pm.plugins[i].Priority() != expectedPriority {
			t.Errorf("Plugin at index %d has priority %d, expected %d",
				i, pm.plugins[i].Priority(), expectedPriority)
		}
	}
}

func TestProcessMultiplePlugins(t *testing.T) {
	pm := NewPluginManager()

	// Add multiple plugins that will process the same file
	plugin1 := &mockPlugin{
		name:         "plugin1",
		priority:     10,
		canProcess:   true,
		shouldModify: true,
	}
	plugin2 := &mockPlugin{
		name:         "plugin2",
		priority:     20,
		canProcess:   true,
		shouldModify: false, // This one doesn't modify
	}

	pm.RegisterPlugin(plugin1)
	pm.RegisterPlugin(plugin2)

	file := File{
		Path:    "test.txt",
		Content: []byte("original"),
	}

	fm := &FileManager{SiteDirectory: "/test"}
	result := pm.Process(file, fm)

	// Should be modified by plugin1
	if string(result.Content) != "modified content" {
		t.Errorf("Expected 'modified content', got '%s'", string(result.Content))
	}
}
