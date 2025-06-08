package impl

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

func printHelp() {
	println("Usage: serve [options]")
	println("Options:")
	println("  --port=8080        	Port to run the HTTP server on")
	println("  --hostname=localhost Hostname of the HTTP server on")
	println("  help               	Display help information")
	println("  run <directory>    	Directory to run the server from")
	println("  create <template> --out=<directory>")
	println("					    Create a new project from a template")
	println("  dump	<template> --out=<directory>")
	println("		  	            Generate and dump the full state of the template (for testing)")
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

	config.SiteDirectory = "../site-business-card-01"
	config.OutDirectory = "../site-out"
	config.Mode = "dump"

	help := false

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
				case "out":
					config.OutDirectory = value
				default:
					log.Fatalf("Invalid argument: %s. Use `help` for usage information.", arg)
				}
			} else {
				log.Fatalf("Invalid argument: %s. Use `help` for usage information.", arg)
			}
		} else if arg == "help" {
			help = true
		} else if arg == "run" {
			config.Mode = "run"
		} else if arg == "create" {
			config.Mode = "create"
		} else if arg == "dump" {
			config.Mode = "dump"
		} else {
			config.SiteDirectory = arg
		}
	}

	if help {
		printHelp()
		os.Exit(0)
	}

	if config.Mode == "run" && config.SiteDirectory == "" {
		log.Fatalf("Missing parameter <directory>")
	}
	if config.Mode == "create" {
		if config.SiteDirectory == "" {
			log.Fatalf("Missing parameter <directory>")
		}
		if config.OutDirectory == "" {
			log.Fatalf("Missing parameter --out")
		}
	}
	if config.Mode == "dump" {
		if config.SiteDirectory == "" {
			log.Fatalf("Missing parameter <directory>")
		}
		if config.OutDirectory == "" {
			log.Fatalf("Missing parameter --out")
		}
	}

	return config, nil
}
