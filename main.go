package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

type UnraidTemplate struct {
	XMLName     xml.Name `xml:"Container"`
	Name        string   `xml:"Name"`
	Repository  string   `xml:"Repository"`
	Network     string   `xml:"Network"`
	WebUI       string   `xml:"WebUI"`
	Category    string   `xml:"Category"`
	Overview    string   `xml:"Overview"`
	Description string   `xml:"Description"`
	Author      string   `xml:"Author"`
	TemplateURL string   `xml:"TemplateURL"`
	Icon        string   `xml:"Icon"`
	Image       string   `xml:"Image"`
	Configs     []Config `xml:"Config"`
}

type Config struct {
	Name        string `xml:"Name,attr"`
	Target      string `xml:"Target,attr"`
	Default     string `xml:"Default,attr"`
	Mode        string `xml:"Mode,attr"`
	Description string `xml:"Description,attr"`
	Type        string `xml:"Type,attr"`
	Display     string `xml:"Display,attr"`
	Required    string `xml:"Required,attr"`
	Mask        string `xml:"Mask,attr"`
	Value       string `xml:",chardata"`
}

var configFile string
var verbose bool
var force bool

func init() {

	flag.BoolVar(&force, "force", false, "force overwrite of existing XML files")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.StringVar(&configFile, "c", "docker-compose.yml", "path to YAML configuration file")
}

func main() {

	// Parse flags
	flag.Parse()

	// Check if a command was provided
	if flag.NArg() < 1 {
		log.Fatal("no command provided")
	}

	// Check if a configuration file was provided
	if configFile == "" {
		log.Fatal("no configuration file provided. Use -c flag to specify a YAML file.")
	}

	// Check if the configuration file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("configuration file %s does not exist", configFile)
	}

	// Get the command
	cmd := flag.Arg(0)

	fmt.Println("Remaining args", flag.Args())

	if force {
		log.Println("WARNING: --force flag is enabled. This will overwrite any existing XML files.")
	} else {
		log.Println("WARNING: --force flag is disabled. This will not overwrite any existing XML files.")
	}

	// Handle commands
	switch cmd {
	case "convert":
		convertCommand(verbose, configFile, force)
	case "validate":
		validateCommand(verbose, configFile, force)
	case "help":
		printHelp()
	default:
		printHelp()
		log.Fatalf("unknown command: %s", cmd)
	}

}

func processProject(verbose bool, configFile string, force bool) (*types.Project, error) {
	if verbose {
		log.Print("processing project")
	}
	ctx := context.Background()
	// Create a new project options
	options, err := cli.NewProjectOptions(
		[]string{configFile},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName("comp2unraid"))

	// Create a new project
	project, err := cli.ProjectFromOptions(ctx, options)
	if err != nil {
		return nil, err
	}

	// Get the services
	services := project.Services

	// Check if any of the XML files that will be created exist
	existingFiles := make(map[string]bool)
	for _, service := range services {
		xmlFile := fmt.Sprintf("%s.xml", service.Name)
		if _, err := os.Stat(xmlFile); err == nil {
			existingFiles[xmlFile] = true
		}
	}

	// If any existing files were found and the force flag is not set, exit
	if len(existingFiles) > 0 && !force {
		return nil, fmt.Errorf("one or more XML files already exist. Use --force to overwrite")
	}

	return project, nil
}

func convertCommand(verbose bool, configFile string, force bool) {
	project, err := processProject(verbose, configFile, force)
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range project.Services {
		template := UnraidTemplate{
			Name:        service.Name,
			Repository:  service.Image,
			Network:     getNetworkMode(&service),
			TemplateURL: fmt.Sprintf("https://raw.githubusercontent.com/username/repo/main/unraid/%s.xml", service.Name),
			Icon:        fmt.Sprintf("https://raw.githubusercontent.com/username/repo/main/unraid/%s-logo.png", service.Name),
			WebUI:       getWebUI(&service),
		}

		template.Configs = append(template.Configs, getConfigs(&service)...)
		template.Configs = append(template.Configs, getEnvironmentConfigs(&service)...)
		template.Configs = append(template.Configs, getVolumeConfigs(&service)...)

		xmlBytes, err := xml.MarshalIndent(template, "", "  ")
		if err != nil {
			log.Fatalf("error marshaling template to XML: %v", err)
		}

		xmlFile, err := os.Create(fmt.Sprintf("%s.xml", service.Name))
		if err != nil {
			log.Fatalf("error creating XML file: %v", err)
		}
		defer xmlFile.Close()

		_, err = xmlFile.Write(xmlBytes)
		if err != nil {
			log.Fatalf("error writing XML to file: %v", err)
		}
	}
}

func getNetworkMode(service *types.ServiceConfig) string {
	if len(service.NetworkMode) == 0 {
		return "bridge"
	}

	return service.NetworkMode
}

func getWebUI(service *types.ServiceConfig) string {
	if len(service.Ports) == 0 {
		return ""
	}
	return fmt.Sprintf("http://[IP]:[PORT:%s]", service.Ports[0].Published)
}

func getConfigs(service *types.ServiceConfig) []Config {
	if len(service.Ports) == 0 {
		return []Config{}
	}
	var port = service.Ports[0].Published

	return []Config{
		{
			Name:        "WebUI",
			Target:      port,
			Default:     port,
			Mode:        "tcp",
			Description: "WebUI Port",
			Type:        "Port",
			Display:     "always",
			Required:    "true",
			Mask:        "false",
			Value:       port,
		},
	}
}

func getEnvironmentConfigs(service *types.ServiceConfig) []Config {
	configs := make([]Config, 0)
	for key, val := range service.Environment {
		configs = append(configs, Config{
			Name:        key,
			Target:      key,
			Default:     *val,
			Mode:        "env",
			Description: "",
			Type:        "Variable",
			Display:     "always",
			Required:    "false",
			Mask:        "false",
			Value:       *val,
		})
	}
	return configs
}

func getVolumeConfigs(service *types.ServiceConfig) []Config {
	configs := make([]Config, 0)
	for _, volume := range service.Volumes {
		configs = append(configs, Config{
			Name:        fmt.Sprintf("Volume for %s", volume.Target),
			Target:      volume.Target,
			Default:     volume.Source,
			Mode:        "rw",
			Description: fmt.Sprintf("Default %s", volume.Source),
			Type:        "Path",
			Display:     "advanced",
			Required:    "true",
			Mask:        "false",
			Value:       volume.Source,
		})
	}
	return configs
}

func validateCommand(verbose bool, configFile string, force bool) {
	_, err := processProject(verbose, configFile, force)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("OK")
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  comp2unraid [command] -c <config_file>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  convert   Convert the docker compose file to an unraid template")
	fmt.Println("  validate  Validate the configuration file")
	fmt.Println("  help      Display this help message")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -c        Path to the YAML configuration file")
	fmt.Println("  -v        Enable verbose output")
	fmt.Println("  --force   Force overwrite of existing XML files")
}
