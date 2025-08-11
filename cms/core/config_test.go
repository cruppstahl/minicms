package core

import (
	"os"
	"reflect"
	"testing"
)

func TestReadConfigYaml(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected Config
		wantErr  bool
	}{
		{
			name: "valid complete config",
			yaml: `
server:
  port: 9090
  hostname: "example.com"
  title: "Test Site"
  description: "A test site"
branding:
  favicon: "/custom/favicon.ico"
  cssfile: "/custom/style.css"
plugins:
  markdown:
    template: "default"
  static:
    enabled: "true"
`,
			expected: Config{
				Server: Server{
					Port:        9090,
					Hostname:    "example.com",
					Title:       "Test Site",
					Description: "A test site",
				},
				Branding: Branding{
					Favicon: "/custom/favicon.ico",
					CssFile: "/custom/style.css",
				},
				Plugins: Plugins{
					"markdown": {
						"template": "default",
					},
					"static": {
						"enabled": "true",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			yaml: `
server:
  port: 8080
`,
			expected: Config{
				Server: Server{
					Port:     8080,
					Hostname: "",
					Title:    "",
				},
			},
			wantErr: false,
		},
		{
			name: "empty config",
			yaml: ``,
			expected: Config{
				Server: Server{},
			},
			wantErr: true,
		},
		{
			name:     "invalid yaml",
			yaml:     `invalid: yaml: content: [`,
			expected: Config{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "config_test_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write test YAML
			if _, err := tmpfile.WriteString(tt.yaml); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpfile.Close()

			// Test ReadConfigYaml
			var config Config
			err = ReadConfigYaml(&config, tmpfile.Name())

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfigYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check FilePath was set
				if config.FilePath != tmpfile.Name() {
					t.Errorf("Expected FilePath to be %s, got %s", tmpfile.Name(), config.FilePath)
				}

				// Check other fields
				if config.Server.Port != tt.expected.Server.Port {
					t.Errorf("Expected Port %d, got %d", tt.expected.Server.Port, config.Server.Port)
				}
				if config.Server.Hostname != tt.expected.Server.Hostname {
					t.Errorf("Expected Hostname %s, got %s", tt.expected.Server.Hostname, config.Server.Hostname)
				}
				if config.Server.Title != tt.expected.Server.Title {
					t.Errorf("Expected Title %s, got %s", tt.expected.Server.Title, config.Server.Title)
				}
				if config.Branding.Favicon != tt.expected.Branding.Favicon {
					t.Errorf("Expected Favicon %s, got %s", tt.expected.Branding.Favicon, config.Branding.Favicon)
				}
				if !reflect.DeepEqual(config.Plugins, tt.expected.Plugins) {
					t.Errorf("Expected Plugins %+v, got %+v", tt.expected.Plugins, config.Plugins)
				}
			}
		})
	}
}

func TestReadConfigYaml_NonexistentFile(t *testing.T) {
	var config Config
	err := ReadConfigYaml(&config, "/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseCommandLineArguments_RunCommand(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "run command with directory",
			args: []string{"program", "run", "/tmp"},
			expected: Config{
				Mode:          "run",
				SiteDirectory: "/tmp",
				Server: Server{
					Port:        8080,
					Hostname:    "localhost",
					Title:       "minicms Server",
					Description: "",
				},
				Branding: Branding{
					Favicon: "/assets/favicon.png",
				},
			},
			wantErr: false,
		},
		{
			name: "run command with custom port and hostname",
			args: []string{"program", "-p", "9000", "--hostname", "example.com", "run", "/tmp"},
			expected: Config{
				Mode:          "run",
				SiteDirectory: "/tmp",
				Server: Server{
					Port:        9000,
					Hostname:    "example.com",
					Title:       "minicms Server",
					Description: "",
				},
				Branding: Branding{
					Favicon: "/assets/favicon.png",
				},
			},
			wantErr: false,
		},
		{
			name: "run command with long flags",
			args: []string{"program", "--port", "3000", "--hostname", "test.local", "run", "/tmp"},
			expected: Config{
				Mode:          "run",
				SiteDirectory: "/tmp",
				Server: Server{
					Port:        3000,
					Hostname:    "test.local",
					Title:       "minicms Server",
					Description: "",
				},
				Branding: Branding{
					Favicon: "/assets/favicon.png",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			config, err := ParseCommandLineArguments()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandLineArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config.Mode != tt.expected.Mode {
					t.Errorf("Expected Mode %s, got %s", tt.expected.Mode, config.Mode)
				}
				if config.SiteDirectory != tt.expected.SiteDirectory {
					t.Errorf("Expected SiteDirectory %s, got %s", tt.expected.SiteDirectory, config.SiteDirectory)
				}
				if config.Server.Port != tt.expected.Server.Port {
					t.Errorf("Expected Port %d, got %d", tt.expected.Server.Port, config.Server.Port)
				}
				if config.Server.Hostname != tt.expected.Server.Hostname {
					t.Errorf("Expected Hostname %s, got %s", tt.expected.Server.Hostname, config.Server.Hostname)
				}
			}
		})
	}
}

func TestParseCommandLineArguments_StaticCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "static command with required flags",
			args: []string{"program", "-o", "/output", "static", "/tmp"},
			expected: Config{
				Mode:          "static",
				SiteDirectory: "/tmp",
				OutDirectory:  "/output",
				Server: Server{
					Port:        8080,
					Hostname:    "localhost",
					Title:       "minicms Server",
					Description: "",
				},
			},
			wantErr: false,
		},
		{
			name:    "static command missing output directory",
			args:    []string{"program", "static", "/tmp"},
			wantErr: true,
		},
		{
			name:    "static command missing source directory",
			args:    []string{"program", "-o", "/output", "static"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			config, err := ParseCommandLineArguments()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandLineArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config.Mode != tt.expected.Mode {
					t.Errorf("Expected Mode %s, got %s", tt.expected.Mode, config.Mode)
				}
				if config.SiteDirectory != tt.expected.SiteDirectory {
					t.Errorf("Expected SiteDirectory %s, got %s", tt.expected.SiteDirectory, config.SiteDirectory)
				}
				if config.OutDirectory != tt.expected.OutDirectory {
					t.Errorf("Expected OutDirectory %s, got %s", tt.expected.OutDirectory, config.OutDirectory)
				}
			}
		})
	}
}

func TestParseCommandLineArguments_DumpCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "dump command with required flags",
			args: []string{"program", "-o", "/dump", "dump", "/tmp"},
			expected: Config{
				Mode:          "dump",
				SiteDirectory: "/tmp",
				OutDirectory:  "/dump",
				Server: Server{
					Port:        8080,
					Hostname:    "localhost",
					Title:       "minicms Server",
					Description: "",
				},
			},
			wantErr: false,
		},
		{
			name:    "dump command missing output directory",
			args:    []string{"program", "dump", "/tmp"},
			wantErr: true,
		},
		{
			name:    "dump command missing source directory",
			args:    []string{"program", "-o", "/dump", "dump"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			config, err := ParseCommandLineArguments()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandLineArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config.Mode != tt.expected.Mode {
					t.Errorf("Expected Mode %s, got %s", tt.expected.Mode, config.Mode)
				}
				if config.SiteDirectory != tt.expected.SiteDirectory {
					t.Errorf("Expected SiteDirectory %s, got %s", tt.expected.SiteDirectory, config.SiteDirectory)
				}
				if config.OutDirectory != tt.expected.OutDirectory {
					t.Errorf("Expected OutDirectory %s, got %s", tt.expected.OutDirectory, config.OutDirectory)
				}
			}
		})
	}
}

func TestParseCommandLineArguments_VersionCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"program", "version"}

	config, err := ParseCommandLineArguments()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if config.Mode != "version" {
		t.Errorf("Expected Mode 'version', got %s", config.Mode)
	}
}

func TestParseCommandLineArguments_InvalidFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid flag",
			args: []string{"program", "--invalid-flag", "run", "/test"},
		},
		{
			name: "invalid port value",
			args: []string{"program", "-p", "invalid", "run", "/test"},
		},
		{
			name: "no command",
			args: []string{"program"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			_, err := ParseCommandLineArguments()
			if err == nil {
				t.Error("Expected error for invalid arguments")
			}
		})
	}
}

func TestParseCommandLineArguments_DefaultValues(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"program", "run", "/tmp"}

	config, err := ParseCommandLineArguments()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test default values
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
	if config.Server.Hostname != "localhost" {
		t.Errorf("Expected default hostname 'localhost', got %s", config.Server.Hostname)
	}
	if config.Server.Title != "minicms Server" {
		t.Errorf("Expected default title 'minicms Server', got %s", config.Server.Title)
	}
	if config.Branding.Favicon != "/assets/favicon.png" {
		t.Errorf("Expected default favicon '/assets/favicon.png', got %s", config.Branding.Favicon)
	}
}

func TestParseCommandLineArguments_EdgeCases(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "empty directory path",
			args:    []string{"program", "run", ""},
			wantErr: true,
		},
		{
			name:    "port boundary - 0",
			args:    []string{"program", "-p", "0", "run", "/test"},
			wantErr: true,
		},
		{
			name:    "port boundary - 65536",
			args:    []string{"program", "-p", "65536", "run", "/test"},
			wantErr: true,
		},
		{
			name:    "negative port",
			args:    []string{"program", "-p", "-1", "run", "/test"},
			wantErr: true,
		},
		{
			name:    "very long hostname",
			args:    []string{"program", "--hostname", string(make([]byte, 1000)), "run", "/test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			_, err := ParseCommandLineArguments()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandLineArguments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_StructTags(t *testing.T) {
	// Test that YAML tags are correctly set
	configType := reflect.TypeOf(Config{})

	serverField, found := configType.FieldByName("Server")
	if !found {
		t.Error("Server field not found")
	} else {
		tag := serverField.Tag.Get("yaml")
		if tag != "server" {
			t.Errorf("Expected yaml tag 'server', got '%s'", tag)
		}
	}

	brandingField, found := configType.FieldByName("Branding")
	if !found {
		t.Error("Branding field not found")
	} else {
		tag := brandingField.Tag.Get("yaml")
		if tag != "branding" {
			t.Errorf("Expected yaml tag 'branding', got '%s'", tag)
		}
	}

	pluginsField, found := configType.FieldByName("Plugins")
	if !found {
		t.Error("Plugins field not found")
	} else {
		tag := pluginsField.Tag.Get("yaml")
		if tag != "plugins" {
			t.Errorf("Expected yaml tag 'plugins', got '%s'", tag)
		}
	}
}

func TestServer_StructTags(t *testing.T) {
	serverType := reflect.TypeOf(Server{})

	expectedTags := map[string]string{
		"Port":        "port",
		"Hostname":    "hostname",
		"Title":       "title",
		"Description": "description",
	}

	for fieldName, expectedTag := range expectedTags {
		field, found := serverType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found", fieldName)
			continue
		}

		tag := field.Tag.Get("yaml")
		if tag != expectedTag {
			t.Errorf("Field %s: expected yaml tag '%s', got '%s'", fieldName, expectedTag, tag)
		}
	}
}

func TestBranding_StructTags(t *testing.T) {
	brandingType := reflect.TypeOf(Branding{})

	expectedTags := map[string]string{
		"Favicon": "favicon",
		"CssFile": "cssfile",
	}

	for fieldName, expectedTag := range expectedTags {
		field, found := brandingType.FieldByName(fieldName)
		if !found {
			t.Errorf("Field %s not found", fieldName)
			continue
		}

		tag := field.Tag.Get("yaml")
		if tag != expectedTag {
			t.Errorf("Field %s: expected yaml tag '%s', got '%s'", fieldName, expectedTag, tag)
		}
	}
}

func TestOptions_FlagTags(t *testing.T) {
	optionsType := reflect.TypeOf(Options{})

	tests := []struct {
		fieldName  string
		shortFlag  string
		longFlag   string
		defaultVal string
	}{
		{"Port", "p", "port", "8080"},
		{"Hostname", "h", "hostname", "localhost"},
		{"Out", "o", "out", ""},
		{"Help", "", "help", ""},
	}

	for _, tt := range tests {
		field, found := optionsType.FieldByName(tt.fieldName)
		if !found {
			t.Errorf("Field %s not found", tt.fieldName)
			continue
		}

		if tt.shortFlag != "" {
			short := field.Tag.Get("short")
			if short != tt.shortFlag {
				t.Errorf("Field %s: expected short flag '%s', got '%s'", tt.fieldName, tt.shortFlag, short)
			}
		}

		if tt.longFlag != "" {
			long := field.Tag.Get("long")
			if long != tt.longFlag {
				t.Errorf("Field %s: expected long flag '%s', got '%s'", tt.fieldName, tt.longFlag, long)
			}
		}

		if tt.defaultVal != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tt.defaultVal {
				t.Errorf("Field %s: expected default '%s', got '%s'", tt.fieldName, tt.defaultVal, defaultTag)
			}
		}
	}
}

func TestReadConfigYaml_FilePermissions(t *testing.T) {
	// Create a temporary file with restricted permissions
	tmpfile, err := os.CreateTemp("", "config_perm_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some content
	tmpfile.WriteString("server:\n  port: 8080\n")
	tmpfile.Close()

	// Make file unreadable (this might not work on all systems/as all users)
	err = os.Chmod(tmpfile.Name(), 0000)
	if err != nil {
		t.Skipf("Cannot change file permissions: %v", err)
	}
	defer os.Chmod(tmpfile.Name(), 0644) // Restore for cleanup

	var config Config
	err = ReadConfigYaml(&config, tmpfile.Name())
	if err == nil {
		t.Error("Expected error reading unreadable file")
	}
}

// Integration tests
func TestIntegration_ConfigFileAndCommandLine(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "integration_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	yamlContent := `
server:
  port: 3000
  hostname: "config.example.com"
  title: "Config File Site"
branding:
  favicon: "/config/favicon.ico"
`
	tmpfile.WriteString(yamlContent)
	tmpfile.Close()

	// Test command line parsing
	os.Args = []string{"program", "-p", "8000", "run", "/tmp"}
	config, err := ParseCommandLineArguments()
	if err != nil {
		t.Fatalf("Failed to parse command line: %v", err)
	}

	// Command line should override defaults
	if config.Server.Port != 8000 {
		t.Errorf("Expected port 8000 from command line, got %d", config.Server.Port)
	}

	// Now read config file (this would typically be done after command line parsing)
	err = ReadConfigYaml(&config, tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Config file should override command line for values it specifies
	if config.Server.Port != 3000 {
		t.Errorf("Expected port 3000 from config file, got %d", config.Server.Port)
	}
	if config.Server.Title != "Config File Site" {
		t.Errorf("Expected title from config file, got %s", config.Server.Title)
	}

	// But command line values for other fields should remain
	if config.Mode != "run" {
		t.Errorf("Expected mode 'run' from command line, got %s", config.Mode)
	}
	if config.SiteDirectory != "/tmp" {
		t.Errorf("Expected site directory from command line, got %s", config.SiteDirectory)
	}
}
