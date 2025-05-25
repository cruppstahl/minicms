package internal

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Port        int    `yaml:"port"`
	Hostname    string `yaml:"hostname"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type Config struct {
	FilePath      string
	SiteDirectory string
	Server        Server `yaml:"server"`
}

func printHelp() {
	println("Usage: serve [options]")
	println("Options:")
	println("  --port=8080        	Port to run the HTTP server on")
	println("  --hostname=localhost Hostname of the HTTP server on")
	println("  help               	Display help information")
	println("  run <directory>    	Directory to run the server from")
}

func ReadConfigYaml(context Context, filePath string) (Context, error) {
	context.Config.FilePath = filePath
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Context{}, err
	}

	err = yaml.Unmarshal(data, &context.Config)
	if err != nil {
		return Context{}, err
	}

	return context, nil
}

func ParseCommandLineArguments() (Config, error) {
	var config = Config{
		Server: Server{
			Port:        8080,
			Hostname:    "localhost",
			Title:       "Miniblog Server",
			Description: "",
		},
	}

	config.SiteDirectory = "../site"

	help := false
	run := false

	flag.Parse()

	// iterate over all command line arguments. If they start with "--", treat them as flags
	// otherwise treat them as command line commands
	for _, arg := range flag.Args() {
		if strings.HasPrefix(arg, "--") {
			// split arg into key and value
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "--")
				value := parts[1]
				switch key {
				case "port":
					port, err := strconv.Atoi(value)
					if err != nil {
						log.Fatalf("Invalid port value: %s, expected integer.", value)
					}
					config.Server.Port = port
				case "hostname":
					config.Server.Hostname = value
				default:
					log.Fatalf("Invalid argument: %s. Use `help` for usage information.", arg)
				}
			} else {
				log.Fatalf("Invalid argument: %s. Use `help` for usage information.", arg)
			}
		} else if arg == "help" {
			help = true
		} else if arg == "run" {
			run = true
		} else {
			config.SiteDirectory = arg
		}
	}

	if help {
		printHelp()
		os.Exit(0)
	}

	/*
		if !run {
			log.Fatalf("Invalid command. Use `help` for usage information.")
		}*/

	if run && config.SiteDirectory == "" {
		log.Fatalf("Missing parameter <directory>")
	}

	return config, nil
}
