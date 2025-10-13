# Test Implementation for Listener-Based File/Route Management

This directory contains comprehensive tests for the listener-based file and route management system in MiniCMS.

## Test Files Overview

### Core Test Files

- **`listener_test.go`** - Unit tests for `FileWatcherListener` functionality
- **`integration_test.go`** - End-to-end integration tests for complete file → route flows
- **`test_helpers.go`** - Reusable test utilities and helper functions
- **`test_helpers_example_test.go`** - Examples demonstrating test helper usage

### Existing Tests (Enhanced)
- **`fswatcher_test.go`** - FileWatcher tests (already comprehensive)
- **`router_test.go`** - RouterManager tests
- **`filemanager_test.go`** - FileManager tests

## Test Categories

### 1. Unit Tests (`listener_test.go`)

Tests individual components in isolation using mocks:

- **FileWatcherListener lifecycle** (start/stop)
- **Event handler functions**:
  - `handleFileCreated()` - File creation handling
  - `handleFileModified()` - File modification handling
  - `handleFileDeleted()` - File deletion handling
  - `handleDirectoryCreated()` - Directory creation handling
  - `handleDirectoryDeleted()` - Directory deletion handling
- **Error handling and edge cases**
- **Concurrent operations**

```go
// Example unit test
func TestHandleFileCreated(t *testing.T) {
    fm, fw, mockRM, tempDir := createListenerTestEnv(t)
    fwl := newFileWatcherListener(fw)

    event := FileWatchEvent{
        Type: FileCreated,
        Path: "content/test.md",
    }

    fwl.handleFileCreated(event)

    // Verify file was added to FileManager and router
}
```

### 2. Integration Tests (`integration_test.go`)

Tests complete end-to-end flows with real components:

- **File creation → Route registration**
- **File modification → Content update**
- **File deletion → Route removal**
- **Directory operations → Bulk route management**
- **HTTP routing verification**
- **Concurrent file operations**
- **Error recovery and resilience**

```go
// Example integration test
func TestFileCreationToRouteFlow(t *testing.T) {
    suite := setupIntegrationTest(t)
    defer suite.teardown()

    // Create file physically
    os.WriteFile(filepath.Join(suite.tempDir, "content/test.md"),
                []byte("# Test"), 0644)

    // Wait for processing
    time.Sleep(300 * time.Millisecond)

    // Verify route was created
    routes := suite.rm.GetAllRoutes()
    // Assert route exists
}
```

### 3. Test Utilities (`test_helpers.go`)

Comprehensive helper functions and builders:

#### TestFileBuilder
Fluent interface for creating test files:

```go
file := env.CreateFile("content/post.md").
    WithContent("# My Post").
    WithRoute("/post").
    WithMimeType("text/html").
    WithMetadata("title", "My Post").
    Build()
```

#### TestDirectoryStructure
Creates complex directory structures:

```go
structure := env.CreateDirectory().
    WithDirectory("content/blog").
    WithFile(env.CreateFile("content/blog/post1.md").WithContent("# Post 1")).
    Create(t)
```

#### EventCollector
Collects and filters FileWatch events:

```go
collector := env.CreateEventCollector(2 * time.Second)
collector.Start()
// ... perform operations ...
events := collector.GetEventsOfType(FileCreated)
```

#### TestEnvironment
Complete test environment setup:

```go
env := NewTestEnvironment(t).Start()
defer env.Stop()

env.AssertFileExists("content/test.md")
env.AssertRouteExists("/test")
```

#### MockPlugin
Plugin testing support:

```go
mockPlugin := NewMockPlugin("test", 100).
    WithProcessFunc(func(file File, fm *FileManager) *File {
        // Custom processing logic
        return &file
    })
```

#### FileOperationSequence
Sequential file operations with event collection:

```go
events, err := NewFileOperationSequence(env).
    CreateFile("content/test.md", "# Test").
    ModifyFile("content/test.md", "# Modified").
    DeleteFile("content/test.md").
    ExecuteWithEventCollection(3 * time.Second)
```

## Running Tests

### Run All Tests
```bash
cd cms/core
go test
```

### Run Specific Test Categories
```bash
# Unit tests only
go test -run TestFileWatcherListener

# Integration tests only
go test -run TestIntegration

# Helper examples
go test -run TestHelperUsage
```

### Run with Verbose Output
```bash
go test -v
```

### Run Benchmarks
```bash
go test -bench=.
```

## Test Strategy

### 1. **Isolation**
- Unit tests use mocks to isolate components
- Integration tests use real components but isolated environments
- Each test uses temporary directories

### 2. **Deterministic**
- Events are collected and verified explicitly
- Timeouts prevent hanging tests
- File system operations use predictable sequences

### 3. **Comprehensive Coverage**
- **Happy paths**: Normal file operations
- **Error conditions**: Invalid paths, permission issues, race conditions
- **Edge cases**: Rapid file creation/deletion, concurrent operations
- **Performance**: Bulk operations, memory usage

### 4. **Maintainable**
- Helper functions reduce code duplication
- Clear test structure and naming
- Comprehensive assertions with good error messages

## Key Testing Patterns

### Event-Driven Testing
```go
collector := env.CreateEventCollector(timeout)
collector.Start()

// Perform operations
createFile("test.md")

// Verify events
events := collector.GetEventsOfType(FileCreated)
assert.Equal(t, 1, len(events))
```

### Fluent Test Building
```go
env.CreateDirectory().
    WithFile(env.CreateFile("content/post.md").WithContent("# Post")).
    Create(t)
```

### Assertion Helpers
```go
env.AssertFileExists("content/post.md")
env.AssertRouteExists("/post")
```

## Debugging Tests

### Enable Verbose Logging
```bash
go test -v -run TestSpecificTest
```

### Debug Event Collection
```go
collector := env.CreateEventCollector(timeout)
events := collector.WaitForEvents(expectedCount)
for _, event := range events {
    t.Logf("Event: %s - %s", event.Type, event.Path)
}
```

### Inspect File Manager State
```go
files := env.FileManager.GetAllFiles()
for path, file := range files {
    t.Logf("File: %s, Routes: %v", path, file.Routes)
}
```

## Coverage

Run with coverage analysis:

```bash
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Expected coverage areas:
- **FileWatcherListener**: >90% (all handler methods)
- **Event processing**: >85% (including error paths)
- **Integration flows**: >80% (file → route workflows)

## Future Enhancements

### Planned Test Additions
1. **Performance tests** for bulk operations
2. **Stress tests** for concurrent file modifications
3. **Plugin integration tests** with real plugins
4. **HTTP endpoint tests** with router verification
5. **Configuration-driven tests** for different site setups

### Test Infrastructure Improvements
1. **Test data generators** for realistic content
2. **Property-based testing** for edge case discovery
3. **Test parallelization** for faster execution
4. **Docker-based testing** for consistent environments