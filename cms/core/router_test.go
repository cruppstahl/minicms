package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Set gin to test mode to reduce noise in tests
	gin.SetMode(gin.TestMode)
}

type TestContext struct {
	context *Context
	tmpdir  string
}

func createTestContext(t *testing.T) TestContext {
	tempDir := t.TempDir()

	// Create assets directory
	assetsDir := filepath.Join(tempDir, "assets")
	err := os.MkdirAll(assetsDir, 0755)
	require.NoError(t, err)

	// Create a test asset file
	testAsset := filepath.Join(assetsDir, "test.css")
	err = os.WriteFile(testAsset, []byte("body { color: red; }"), 0644)
	require.NoError(t, err)

	fm := NewFileManager(tempDir)

	// Add some test files
	file1 := &File{
		Path:    "content/index.html",
		Content: []byte("<h1>Home Page</h1>"),
		Routes:  []string{"/", "/home"},
		Metadata: FileMetadata{
			MimeType: "text/html",
		},
	}

	file2 := &File{
		Path:    "content/about.html",
		Content: []byte("<h1>About Page</h1>"),
		Routes:  []string{"/about"},
		Metadata: FileMetadata{
			MimeType: "text/html",
		},
	}

	file3 := &File{
		Path:    "content/redirect.html",
		Content: []byte(""),
		Routes:  []string{"/old-page"},
		Metadata: FileMetadata{
			RedirectUrl: "/new-page",
		},
	}

	// File outside content directory - should not have routes
	file4 := &File{
		Path:    "templates/layout.html",
		Content: []byte("<html></html>"),
		Routes:  []string{"/template"},
		Metadata: FileMetadata{
			MimeType: "text/html",
		},
	}

	fm.Files[file1.Path] = file1
	fm.Files[file2.Path] = file2
	fm.Files[file3.Path] = file3
	fm.Files[file4.Path] = file4

	ctx := Context{
		FileManager: fm,
	}

	return TestContext{context: &ctx, tmpdir: tempDir}
}

func newRouterManager(ctx *Context) (*RouterManager, error) {
	rm := NewRouterManager()
	err := rm.InitializeRouter(ctx)
	return rm, err
}

func TestInitializeRouter(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Test home page routes
	testCases := []struct {
		name           string
		route          string
		expectedStatus int
		expectedBody   string
	}{
		{"Home route /", "/", http.StatusOK, "<h1>Home Page</h1>"},
		{"Home route /home", "/home", http.StatusOK, "<h1>Home Page</h1>"},
		{"About route", "/about", http.StatusOK, "<h1>About Page</h1>"},
		{"Not found", "/nonexistent", http.StatusNotFound, ""},
		{"Template route not accessible", "/template", http.StatusNotFound, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.route, nil)
			w := httptest.NewRecorder()
			rm.GetRouter().ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestRedirectHandling(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	req, _ := http.NewRequest("GET", "/old-page", nil)
	w := httptest.NewRecorder()
	rm.GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/new-page", w.Header().Get("Location"))
}

func TestRouterManager(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Test adding a route
	testFile := &File{
		Path:    "content/test.html",
		Content: []byte("<h1>Test</h1>"),
		Routes:  []string{"/test"},
		Metadata: FileMetadata{
			MimeType: "text/html",
		},
	}

	ctx.context.FileManager.Files[testFile.Path] = testFile
	err = rm.AddRoute("/test", testFile.Path)
	assert.NoError(t, err)

	// Test the new route
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	rm.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "<h1>Test</h1>", w.Body.String())

	// Test removing the route
	err = rm.RemoveRoute("/test")
	assert.NoError(t, err)

	// Route should now return 404
	req, _ = http.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	rm.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouteUpdates(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Add a route
	testFile := &File{
		Path:    "content/dynamic.html",
		Content: []byte("<h1>Original</h1>"),
		Routes:  []string{"/dynamic"},
		Metadata: FileMetadata{
			MimeType: "text/html",
		},
	}

	ctx.context.FileManager.Files[testFile.Path] = testFile
	err = rm.AddRoute("/dynamic", testFile.Path)
	require.NoError(t, err)

	// Test original content
	req, _ := http.NewRequest("GET", "/dynamic", nil)
	w := httptest.NewRecorder()
	rm.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "<h1>Original</h1>", w.Body.String())

	// Update file content
	testFile.Content = []byte("<h1>Updated</h1>")

	// Test updated content (should reflect immediately since we use file manager)
	req, _ = http.NewRequest("GET", "/dynamic", nil)
	w = httptest.NewRecorder()
	rm.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "<h1>Updated</h1>", w.Body.String())
}

func TestInvalidRoutes(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Test invalid route patterns
	testCases := []struct {
		name  string
		route string
	}{
		{"Empty route", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := rm.AddRoute(tc.route, "content/test.html")
			assert.Error(t, err)
		})
	}
}

func TestDuplicateRoutes(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Add a route
	err = rm.AddRoute("/duplicate", "content/test1.html")
	assert.NoError(t, err)

	// Try to add the same route again
	err = rm.AddRoute("/duplicate", "content/test2.html")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRemoveRoutes(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Add a route
	err = rm.AddRoute("/duplicate", "content/test1.html")
	assert.NoError(t, err)

	// Remove it
	err = rm.RemoveRoute("/duplicate")
	assert.NoError(t, err)

	// Try to add the same route again - needs to succeed
	err = rm.AddRoute("/duplicate", "content/test2.html")
	assert.NoError(t, err)
}

func TestRemoveNonexistentRoute(t *testing.T) {
	ctx := createTestContext(t)

	rm, err := newRouterManager(ctx.context)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Try to remove a route that doesn't exist
	err = rm.RemoveRoute("/nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
