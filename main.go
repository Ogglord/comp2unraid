package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// Set during build
var Commit string
var Branch string
var Version = "1.0.0"

var options commandLineOptions
var tempFiles []string

func init() {
	if Commit == "" {
		// If the Commit is not set during build, try to get it from git at init time
		Commit = getLatestCommit()
	}
	if Branch == "" {
		// If the Commit is not set during build, try to get it from git at init time
		Branch = "unknown"
	}
}

func getLatestCommit() string {
	// Use the "git describe" command to get the latest Git tag.
	// This command returns a string in the format "v1.0.0-123-gabcdef".
	// We'll parse this string to get the tag name and the number of commits since the tag.
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	r := strings.TrimSpace(string(output))

	return r
}
func init() {
	flag.BoolVar(&options.force, "f", false, "overwrite existing XML files")
	flag.BoolVar(&options.verbose, "v", false, "verbose output")
	flag.BoolVar(&options.useEnv, "e", false, "use current environment variables and .env file if available")
	flag.BoolVar(&options.dryRun, "n", false, "dry run - outputs xml to stdout without creating files")

	// modify the default usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "comp2unraid [flags] <config_file>\n")
		fmt.Fprintf(os.Stderr, "Version: %s\n", Version)

		if Branch != "" {
			fmt.Fprintf(os.Stderr, "Branch: %s\n", Branch)
		}
		if Commit != "" {
			fmt.Fprintf(os.Stderr, "Commit: %s\n", Commit)
		}
		fmt.Fprintf(os.Stderr, "\nUsage:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "<config_file> is the path to the configuration file ")
		fmt.Fprintf(os.Stderr, "it may be a URL (https://...) or a local path\n")
	}

}

type UnraidTemplate struct {
	XMLName     xml.Name `xml:"Container"`
	Version     string   `xml:"version,attr"`
	Name        string   `xml:"Name"`
	Repository  string   `xml:"Repository"`
	Registry    string   `xml:"Registry"`
	Network     string   `xml:"Network"`
	WebUI       string   `xml:"WebUI"`
	Category    string   `xml:"Category"`
	Overview    string   `xml:"Overview"`
	Project     string   `xml:"Project"`
	Author      string   `xml:"Author"`
	Support     string   `xml:"Support"`
	TemplateURL string   `xml:"TemplateURL"`
	Icon        string   `xml:"Icon"`
	Shell       string   `xml:"Shell"`
	Privileged  bool     `xml:"Privileged"`
	ExtraParams string   `xml:"ExtraParams"`
	PostArgs    string   `xml:"PostArgs"`
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
	Required    bool   `xml:"Required,attr"`
	Mask        bool   `xml:"Mask,attr"`
	Value       string `xml:",chardata"`
}
type commandLineOptions struct {
	verbose            bool
	force              bool
	useEnv             bool
	dryRun             bool
	configFile         string
	templateRepository string
	resourceRepository string
	Author             string
}

func (c *commandLineOptions) SetRepository(repository string) {
	if strings.Count(repository, "/") != 1 {
		c.resourceRepository = repository
		c.templateRepository = repository
		c.Author = "comp2unraid"
	} else {
		// the repo is in github shorthand format
		c.Author = strings.Split(repository, "/")[0]
		c.resourceRepository = fmt.Sprintf("https://raw.githubusercontent.com/%s/main", repository)
		c.templateRepository = fmt.Sprintf("https://github.com/%s", repository)
	}
}

// getLocalPath returns the path to the local config file
// or downloads the file if it's a URL
// and returns the path to the downloaded file in temp folder
func (options commandLineOptions) getLocalPath() (string, error) {
	url := options.configFile
	var file *os.File
	var err error

	if strings.HasPrefix(url, "https://") {
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		file, err = os.CreateTemp("", "comp2unraid-")
		if err != nil {
			return "", err
		}

		tempFiles = append(tempFiles, file.Name())

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", err
		}
	} else {
		file, err = os.Open(url)
		if err != nil {
			return "", err
		}
		defer file.Close()
	}

	return file.Name(), nil
}

func cleanUpTempFiles() {
	for _, file := range tempFiles {
		os.Remove(file)
	}
}

func main() {
	// Parse flags
	flag.Parse()
	// Get the arguments
	args := flag.Args()

	// Check if at least one command was provided
	if len(args) < 1 {
		if options.verbose {
			log.Printf("arguments: %v", args)
			log.Printf("arguments: %v", os.Args)
		}
		flag.Usage()
		os.Exit(1)
	}

	// Set the config file
	options.configFile = args[0]
	// Get the optional repository argument
	repo := "Ogglord/comp2unraid"
	if len(args) > 1 {
		repo = args[1]
	}
	options.SetRepository(repo)
	defer cleanUpTempFiles()
	convertCommand(options)

}

func convertCommand(args commandLineOptions) {
	project, err := parseYaml(args)
	if err != nil {
		log.Fatalf("error parsing YAML: %v", err)
	}

	if !args.dryRun && !args.force {
		// Check if any of the XML files that will be created exist
		existingFiles := make(map[string]bool)
		for _, service := range project.Services {
			xmlFile := fmt.Sprintf("%s.xml", service.Name)
			if _, err := os.Stat(xmlFile); err == nil {
				existingFiles[xmlFile] = true
			}
		}

		// If any existing files were found and the force flag is not set, exit
		if len(existingFiles) > 0 {
			if args.verbose {
				log.Printf("The following XML files already exist:")
				for file := range existingFiles {
					log.Printf("- %s", file)
				}
			}
			log.Fatalf("unable to proceed with conversion. use -f to force overwrite existing xml files")
		}
	}

	for _, service := range project.Services {

		registry, err := getRegistryURL(service.Image)
		if err != nil {
			log.Fatalf("error in getRegistryURL(...): %v", err)
		}
		template := UnraidTemplate{
			Version:     "2",
			Name:        service.Name,
			Category:    "Other:",
			Repository:  service.Image,
			Registry:    registry,
			Network:     getNetworkMode(&service),
			TemplateURL: fmt.Sprintf("%s/%s.xml", args.resourceRepository, service.Name),
			Icon:        fmt.Sprintf("%s/icons/generic-logo.png", args.resourceRepository),
			Support:     fmt.Sprintf("%s/issues/new/choose", args.templateRepository),
			WebUI:       getWebUI(&service),
			Shell:       "bash",
			Overview:    "This template was created using comp2unraid<br>Convert docker compose templates to unraid templates<br>https://github.com/Ogglord/comp2unraid",
			Author:      args.Author,
			Project:     "",
		}

		template.Configs = append(template.Configs, getConfigs(&service)...)
		template.Configs = append(template.Configs, getEnvironmentConfigs(&service)...)
		template.Configs = append(template.Configs, getVolumeConfigs(&service)...)

		if args.dryRun {
			err = template.writeTemplateToStdout()
			if err != nil {
				log.Fatalf("error printing XML: %v", err)
			}
		} else {
			err = template.writeTemplateToDisk(fmt.Sprintf("%s.xml", service.Name))
			if err != nil {
				log.Fatalf("error writing XML to file: %v", err)
			}
		}
	}
}

func parseYaml(args commandLineOptions) (*types.Project, error) {
	if args.verbose {
		log.Print("processing project")
	}
	ctx := context.Background()
	// Create a new project options

	localPath, err := args.getLocalPath()
	if err != nil {
		return nil, err
	}
	projOptions, err := cli.NewProjectOptions(
		[]string{localPath},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName("comp2unraid"))

	if err != nil {
		return nil, err
	}

	// Create a new project
	project, err := cli.ProjectFromOptions(ctx, projOptions)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (template UnraidTemplate) writeTemplateToDisk(filename string) error {
	xmlBytes, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling template to XML: %v", err)
	}

	// Add the XML header
	xmlHeader := []byte(`<?xml version="1.0"?>` + "\n")
	xmlBytes = append(xmlHeader, xmlBytes...)

	xmlFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating XML file: %v", err)
	}
	defer xmlFile.Close()

	_, err = xmlFile.Write(xmlBytes)
	if err != nil {
		return fmt.Errorf("error writing XML to file: %v", err)
	}

	fmt.Printf("XML file created: %s.xml\n", filename)
	return nil
}

func (template UnraidTemplate) writeTemplateToStdout() error {
	xmlBytes, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling template to XML: %v", err)
	}

	// Add the XML header
	xmlHeader := []byte(`<?xml version="1.0"?>` + "\n")
	xmlBytes = append(xmlHeader, xmlBytes...)

	_, err = os.Stdout.Write(xmlBytes)
	if err != nil {
		return fmt.Errorf("error writing XML to stdout: %v", err)
	}

	return nil
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
		// no ports published, skip this element
		return []Config{}
	}

	// lets assume the first port published is the webui
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
			Required:    true,
			Mask:        false,
			Value:       port,
		},
	}
}

func getEnvironmentConfigs(service *types.ServiceConfig) []Config {
	//<Config Name="Variable: VPN_USER" Target="VPN_USER" Default="" Mode="" Description="Specify your VPN providers username."
	//Type="Variable" Display="always" Required="true" Mask="true">lcuhiYvP8zyhlaC4+pmp</Config>
	configs := make([]Config, 0)
	for key, val := range service.Environment {
		needsMask := strings.Contains(strings.ToUpper(key), "PWD") ||
			strings.Contains(strings.ToUpper(key), "PASS") ||
			strings.Contains(strings.ToUpper(key), "SECRET")
		configs = append(configs, Config{
			Name:        key,
			Target:      key,
			Default:     *val,
			Mode:        "",
			Description: "Specify the value for env: " + key,
			Type:        "Variable",
			Display:     "always",
			Required:    true,
			Mask:        needsMask,
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
			Default:     ".",
			Mode:        "rw",
			Description: fmt.Sprintf("E.g. /mnt/appdata/%s for config and /mnt/data/ for other volumes", service.Name),
			Type:        "Path",
			Display:     "always",
			Required:    true,
			Mask:        false,
			Value:       ".",
		})
	}
	return configs
}

func getRegistryURL(image string) (string, error) {
	// Get the image parts
	// example quay.io/nextcloud/server -> https://quay.io/nextcloud/server/
	imageParts := strings.SplitN(image, "/", 3)
	var registry, repository, imageName string
	if len(imageParts) == 2 {
		registry = ""
		repository = imageParts[0]
		imageName = imageParts[1]
	} else if len(imageParts) == 3 {
		registry = imageParts[0]
		repository = imageParts[1]
		imageName = imageParts[2]
	} else {
		return "", fmt.Errorf("invalid image name: %s", image)
	}

	if len(strings.Split(imageName, ":")) > 1 {
		imageName = strings.Split(imageName, ":")[0]
	}

	switch registry {
	case "quay.io":
		return fmt.Sprintf("https://quay.io/repository/%s/%s", repository, imageName), nil
	case "ghcr.io":
		return fmt.Sprintf("https://github.com/%s/%s", repository, imageName), nil
	case "docker.io":
		return fmt.Sprintf("https://hub.docker.com/r/%s/%s", repository, imageName), nil
	default:
		return fmt.Sprintf("https://hub.docker.com/r/%s/%s", repository, imageName), nil
	}
}
