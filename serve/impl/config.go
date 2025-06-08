package impl

import (
	"os"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Port        int    `yaml:"port"`
	Hostname    string `yaml:"hostname"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type Branding struct {
	Favicon string `yaml:"favicon"`
	CssFile string `yaml:"cssfile"`
}

type Config struct {
	FilePath      string
	SiteDirectory string
	Mode          string
	OutDirectory  string
	Server        Server   `yaml:"server"`
	Branding      Branding `yaml:"branding"`
}

// Options defines the command-line options structure
type Options struct {
	Port     int    `short:"p" long:"port" description:"Port to run the HTTP server on" default:"8080"`
	Hostname string `short:"h" long:"hostname" description:"Hostname of the HTTP server" default:"localhost"`
	Out      string `short:"o" long:"out" description:"Output directory"`
	Help     bool   `long:"help" description:"Display help information"`
}

// Commands defines the available subcommands
type Commands struct {
	Run  RunCommand  `command:"run" description:"Run the server from a directory"`
	Dump DumpCommand `command:"dump" description:"Generate and dump the full state of the template (for testing)"`
}

type RunCommand struct {
	Args struct {
		Directory string `positional-arg-name:"directory" description:"Directory to run the server from"`
	} `positional-args:"yes" required:"yes"`
}

type DumpCommand struct {
	Args struct {
		Template string `positional-arg-name:"template" description:"Template to dump"`
	} `positional-args:"yes" required:"yes"`
}

func ReadConfigYaml(config *Config, filePath string) error {
	config.FilePath = filePath
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &config)
}

func ParseCommandLineArguments() (Config, error) {
	// Initialize config with defaults
	var config = Config{
		Server: Server{
			Port:        8080,
			Hostname:    "localhost",
			Title:       "Miniblog Server",
			Description: "",
		},
		Branding: Branding{
			Favicon: "/assets/favicon.png",
		},
	}

	var opts Options
	var commands Commands

	parser := flags.NewParser(&opts, flags.Default)
	parser.AddCommand("run", "Run the server from a directory",
		"Run the server from the specified directory", &commands.Run)
	parser.AddCommand("dump", "Generate and dump template state",
		"Generate and dump the full state of the template (for testing)", &commands.Dump)

	_, err := parser.Parse()
	if err != nil {
		// Check if it's a help request
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		return config, err
	}

	// Apply global options to config
	config.Server.Port = opts.Port
	config.Server.Hostname = opts.Hostname
	if opts.Out != "" {
		config.OutDirectory = opts.Out
	}

	// Handle commands
	if parser.Active != nil {
		switch parser.Active.Name {
		case "run":
			config.Mode = "run"
			config.SiteDirectory = commands.Run.Args.Directory
		case "dump":
			config.Mode = "dump"
			config.SiteDirectory = commands.Dump.Args.Template
		}
	}

	// Validation
	if config.Mode == "run" && config.SiteDirectory == "" {
		return config, &flags.Error{
			Type:    flags.ErrRequired,
			Message: "Missing parameter <directory>",
		}
	}

	if config.Mode == "dump" {
		if config.SiteDirectory == "" {
			return config, &flags.Error{
				Type:    flags.ErrRequired,
				Message: "Missing parameter <directory>",
			}
		}
		if config.OutDirectory == "" {
			return config, &flags.Error{
				Type:    flags.ErrRequired,
				Message: "Missing parameter --out",
			}
		}
	}

	return config, nil
}
