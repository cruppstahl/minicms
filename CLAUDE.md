# Claude Code Configuration for MiniCMS

This file contains configuration and guidelines for Claude Code when working with the MiniCMS project.

## Project Structure

MiniCMS is a file-based content management system written in Go. The main components are:

- `cms/main.go` - Entry point
- `cms/core/` - Core functionality (config, filemanager, router, etc.)
- `cms/cmd/` - CLI commands
- `cms/plugins/` - Content processing plugins

## Build & Test Commands

```bash
# Build the project
make build

# Run all tests
make test

# Run unit tests only
make unittests

# Run file tests only
make filetest

# Run the server
make run
```

## Go Development Guidelines

### Code Style
- Follow standard Go formatting with `gofmt`
- Use meaningful variable names
- Keep functions focused and small
- Use Go's built-in error handling patterns
- Follow Go naming conventions (exported vs unexported)

### Testing
- Unit tests in `*_test.go` files alongside source
- Use `go test` for running tests
- Follow table-driven test patterns where appropriate
- Mock external dependencies

### Dependencies
- Go version: 1.23+ (toolchain 1.24.3)
- Main web framework: Gin (github.com/gin-gonic/gin)
- Search: Bleve (github.com/blevesearch/bleve)
- Configuration: YAML (gopkg.in/yaml.v2)
- CLI: go-flags (github.com/jessevdk/go-flags)

### Architecture Patterns
- Plugin-based architecture for content processing
- Context pattern for passing shared state
- File watching for automatic content updates
- Template-based theming system

## Common Tasks

### Adding a New Plugin
1. Create plugin in `cms/plugins/`
2. Implement the Plugin interface
3. Register in `cms/main.go`
4. Add tests in `cms/core/plugin_test.go`

### Modifying Core Functionality
- Update relevant files in `cms/core/`
- Ensure tests pass with `make unittests`
- Test file operations with `make filetest`

### Configuration Changes
- Update `cms/core/config.go`
- Update config validation
- Update example templates if needed

## Linting & Quality

```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run tests
make test
```

## Directory Layout

```
cms/
├── main.go           # Application entry point
├── go.mod           # Go module definition
├── cmd/             # CLI command implementations
├── core/            # Core business logic
│   ├── config.go    # Configuration management
│   ├── filemanager.go # File operations
│   ├── router.go    # HTTP routing
│   └── *_test.go    # Unit tests
└── plugins/         # Content processing plugins
```

## Best Practices for Claude

1. Always run `make test` after making changes
2. Follow existing code patterns and naming conventions
3. Update tests when modifying functionality
4. Use the existing error handling patterns
5. Maintain backwards compatibility with existing templates
6. Keep plugin interface simple and focused
7. Use Go's standard library when possible
8. Follow the repository's existing commit message style