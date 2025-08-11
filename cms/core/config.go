package core

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

// Configuration constants
const (
	DefaultPort       = 8080
	DefaultHostname   = "localhost"
	DefaultTitle      = "minicms Server"
	DefaultFavicon    = "/assets/favicon.png"
	MinPort           = 1
	MaxPort           = 65535
	MaxHostnameLength = 253
	MaxTitleLength    = 200
	MaxDescLength     = 500
)

// Validation errors
var (
	ErrInvalidPort       = errors.New("port must be between 1 and 65535")
	ErrInvalidHostname   = errors.New("hostname is invalid")
	ErrEmptyDirectory    = errors.New("directory cannot be empty")
	ErrDirectoryNotExist = errors.New("directory does not exist")
	ErrInvalidPath       = errors.New("path contains invalid characters")
	ErrMissingOutput     = errors.New("output directory is required")
	ErrConfigNotFound    = errors.New("configuration file not found")
	ErrInvalidYAML       = errors.New("invalid YAML configuration")
)

type Server struct {
	Port        int    `yaml:"port"`
	Hostname    string `yaml:"hostname"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func (s *Server) Validate() error {
	if s.Port < MinPort || s.Port > MaxPort {
		return fmt.Errorf("%w: got %d", ErrInvalidPort, s.Port)
	}

	if s.Hostname != "" {
		if len(s.Hostname) > MaxHostnameLength {
			return fmt.Errorf("%w: hostname too long (%d > %d)",
				ErrInvalidHostname, len(s.Hostname), MaxHostnameLength)
		}

		// Basic hostname validation
		if strings.Contains(s.Hostname, " ") || strings.Contains(s.Hostname, "\t") {
			return fmt.Errorf("%w: hostname contains whitespace", ErrInvalidHostname)
		}

		// Check if it's a valid IP or hostname
		if net.ParseIP(s.Hostname) == nil {
			// Not an IP, validate as hostname
			if !isValidHostname(s.Hostname) {
				return fmt.Errorf("%w: invalid hostname format", ErrInvalidHostname)
			}
		}
	}

	if len(s.Title) > MaxTitleLength {
		return fmt.Errorf("title too long: %d > %d", len(s.Title), MaxTitleLength)
	}

	if len(s.Description) > MaxDescLength {
		return fmt.Errorf("description too long: %d > %d", len(s.Description), MaxDescLength)
	}

	return nil
}

// Performs basic hostname validation TODO - improve with more comprehensive checks
func isValidHostname(hostname string) bool {
	if hostname == "" || len(hostname) > MaxHostnameLength {
		return false
	}

	// Hostname cannot start or end with a dot
	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return false
	}

	// Check each label
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Labels cannot start or end with hyphen
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}

		// Labels must contain only alphanumeric characters and hyphens
		for _, char := range label {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '-') {
				return false
			}
		}
	}

	return true
}

type Branding struct {
	Favicon string `yaml:"favicon"`
	CssFile string `yaml:"cssfile"`
}

func (b *Branding) Validate() error {
	// Basic path validation for favicon and CSS file
	if b.Favicon != "" && !isValidPath(b.Favicon) {
		return fmt.Errorf("%w: invalid favicon path", ErrInvalidPath)
	}

	if b.CssFile != "" && !isValidPath(b.CssFile) {
		return fmt.Errorf("%w: invalid CSS file path", ErrInvalidPath)
	}

	return nil
}

type Plugins map[string]map[string]string

func (p Plugins) Validate() error {
	for pluginName, config := range p {
		if pluginName == "" {
			return errors.New("plugin name cannot be empty")
		}

		// Validate plugin configuration keys and values
		for key, _ := range config {
			if key == "" {
				return fmt.Errorf("plugin %s has empty configuration key", pluginName)
			}
		}
	}

	return nil
}

type Config struct {
	FilePath      string
	SiteDirectory string
	Mode          string
	OutDirectory  string
	Server        Server   `yaml:"server"`
	Branding      Branding `yaml:"branding"`
	Plugins       Plugins  `yaml:"plugins"`
}

func (c *Config) Validate() error {
	// Validate server configuration
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server configuration error: %w", err)
	}

	// Validate branding configuration
	if err := c.Branding.Validate(); err != nil {
		return fmt.Errorf("branding configuration error: %w", err)
	}

	// Validate plugins configuration
	if err := c.Plugins.Validate(); err != nil {
		return fmt.Errorf("plugins configuration error: %w", err)
	}

	return nil
}

// Validates the site directory
func (c *Config) validateSiteDirectory() error {
	if c.SiteDirectory == "" {
		return fmt.Errorf("%w: site directory", ErrEmptyDirectory)
	}

	if !isValidPath(c.SiteDirectory) {
		return fmt.Errorf("%w: site directory", ErrInvalidPath)
	}

	// Check if directory exists
	if _, err := os.Stat(c.SiteDirectory); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrDirectoryNotExist, c.SiteDirectory)
	}

	return nil
}

// Validates the output directory
func (c *Config) validateOutDirectory() error {
	if c.OutDirectory == "" {
		return ErrMissingOutput
	}

	if !isValidPath(c.OutDirectory) {
		return fmt.Errorf("%w: output directory", ErrInvalidPath)
	}

	return nil
}

// Validates file system paths
func isValidPath(path string) bool {
	if path == "" {
		return false
	}

	// Check for path traversal attempts
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return false
	}

	// Check for invalid characters (basic check)
	invalidChars := []string{"\x00", "<", ">", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	return true
}

// Options defines the command-line options structure
type Options struct {
	Port     int    `short:"p" long:"port" description:"Port to run the HTTP server on" default:"8080"`
	Hostname string `short:"h" long:"hostname" description:"Hostname of the HTTP server" default:"localhost"`
	Out      string `short:"o" long:"out" description:"Output directory"`
	Help     bool   `long:"help" description:"Display help information"`
}

func (o *Options) Validate() error {
	if o.Port < MinPort || o.Port > MaxPort {
		return fmt.Errorf("%w: got %d", ErrInvalidPort, o.Port)
	}

	if o.Hostname != "" && !isValidHostname(o.Hostname) {
		return fmt.Errorf("%w: %s", ErrInvalidHostname, o.Hostname)
	}

	if o.Out != "" && !isValidPath(o.Out) {
		return fmt.Errorf("%w: output directory", ErrInvalidPath)
	}

	return nil
}

// Commands defines the available subcommands
type Commands struct {
	Run     RunCommand     `command:"run" description:"Run the server from a directory"`
	Static  StaticCommand  `command:"static" description:"Run as static html generator"`
	Dump    DumpCommand    `command:"dump" description:"Dumps the whole state to disk"`
	Version VersionCommand `command:"version" description:"Print the build version"`
}

type RunCommand struct {
	Args struct {
		Directory string `positional-arg-name:"directory" description:"Directory to run the server from"`
	} `positional-args:"yes" required:"yes"`
}

type StaticCommand struct {
	Args struct {
		Directory string `positional-arg-name:"directory" description:"Directory with source files"`
	} `positional-args:"yes" required:"yes"`
}

type DumpCommand struct {
	Args struct {
		Directory string `positional-arg-name:"directory" description:"Directory with source files"`
	} `positional-args:"yes" required:"yes"`
}

type VersionCommand struct {
	Args struct {
	} `positional-args:"no" required:"no"`
}

// Reads and validates a YAML configuration file
func ReadConfigYaml(config *Config, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("%w: empty file path", ErrInvalidPath)
	}

	// Validate file path
	if !isValidPath(filePath) {
		return fmt.Errorf("%w: %s", ErrInvalidPath, filePath)
	}

	// Set the file path
	config.FilePath = filePath

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrConfigNotFound, filePath)
		}
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidYAML, err.Error())
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// Creates a new configuration with default values
func NewDefaultConfig() Config {
	return Config{
		Server: Server{
			Port:        DefaultPort,
			Hostname:    DefaultHostname,
			Title:       DefaultTitle,
			Description: "",
		},
		Branding: Branding{
			Favicon: DefaultFavicon,
		},
		Plugins: make(Plugins),
	}
}

// Parses command line arguments and returns a validated configuration
func ParseCommandLineArguments() (Config, error) {
	// Initialize config with defaults
	config := NewDefaultConfig()

	var opts Options
	var commands Commands

	parser := flags.NewParser(&opts, flags.Default)
	parser.AddCommand("run", "Run the server from a directory",
		"Run the server from the specified directory", &commands.Run)
	parser.AddCommand("static", "Generate static html files",
		"Generate static html files for the specified directory", &commands.Static)
	parser.AddCommand("dump", "Dumps internal state (for testing)",
		"Process the specified directory, then dump the whole state", &commands.Dump)
	parser.AddCommand("version", "Print the build version",
		"Print the build version", &commands.Version)

	_, err := parser.Parse()
	if err != nil {
		// Check if it's a help request
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		return config, fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return config, fmt.Errorf("invalid command line options: %w", err)
	}

	// Apply global options to config
	config.Server.Port = opts.Port
	config.Server.Hostname = opts.Hostname
	if opts.Out != "" {
		config.OutDirectory = opts.Out
	}

	// Handle (and validate) commands
	if parser.Active != nil {
		switch parser.Active.Name {
		case "run":
			config.Mode = "run"
			config.SiteDirectory = commands.Run.Args.Directory
			if err := config.validateSiteDirectory(); err != nil {
				return config, err
			}
		case "static":
			config.Mode = "static"
			config.SiteDirectory = commands.Static.Args.Directory
			if err := config.validateSiteDirectory(); err != nil {
				return config, err
			}
			if err := config.validateOutDirectory(); err != nil {
				return config, err
			}
		case "dump":
			config.Mode = "dump"
			config.SiteDirectory = commands.Dump.Args.Directory
			if err := config.validateSiteDirectory(); err != nil {
				return config, err
			}
			if err := config.validateOutDirectory(); err != nil {
				return config, err
			}
		case "version":
			config.Mode = "version"
		default:
			return config, fmt.Errorf("unknown command: %s", parser.Active.Name)
		}
	} else {
		return config, errors.New("no command specified")
	}

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return config, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}
