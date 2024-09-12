package main

import "encoding/xml"

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
	namedService       string
	templateRepository string
	resourceRepository string
	Author             string
}
