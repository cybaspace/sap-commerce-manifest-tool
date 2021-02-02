package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

const version = "sap-commerce-manifest-tool v1.0"

type Persona string

const (
	Development Persona = "development"
	Staging     Persona = "staging"
	Production  Persona = "production"
	Empty       Persona = ""
)

type Manifest struct {
	CommerceSuiteVersion         string
	UseCloudExtensionPack        bool
	EnableImageProcessingService bool
	Extensions                   []string
	ExtensionPacks               []ExtensionPack
	TroubleshootingModeEnabled   bool
	DisableImageReuse            bool
	UseConfig                    Config
	StorefrontAddons             []Addon
	Properties                   []Property
	Aspects                      []Aspect
	Tests                        []Test
	WebTests                     []Test
}

type ExtensionPack struct {
	Name     string
	Version  string
	Artifact string
}

type Addon struct {
	Addon       string
	Addons      []string
	Storefront  string
	Storefronts []string
	template    string
}

type Config struct {
	Extensions ExtensionConfig
	Properties []PropertyConfig
	Solr       LocationConfig
	Languages  LocationConfig
}

type ExtensionConfig struct {
	Location string
	Exlude   []string
}

type PropertyConfig struct {
	Location string
	Aspect   string
	Persona  Persona
}

type LocationConfig struct {
	Location string
}

type Aspect struct {
	Name       string
	Properties []Property
	Webapps    []Webapp
}

type Property struct {
	Key     string
	Value   string
	Persona Persona
	Secret  bool
}

type Webapp struct {
	Name        string
	ContextPath string
}

type Test struct {
	Extensions       []string
	Annotaions       []string
	Packages         []string
	ExcludedPackages []string
}

// Structure for localextensions.xml
type Hybrisconfig struct {
	XMLName    xml.Name      `xml:"hybrisconfig"`
	Extensions XmlExtensions `xml:"extensions"`
}
type XmlExtensions struct {
	XMLName    xml.Name       `xml:"extensions"`
	Extensions []XmlExtension `xml:"extension"`
}
type XmlExtension struct {
	XMLName xml.Name `xml:"extension"`
	Name    string   `xml:"name,attr"`
}

var latestFlag bool

func main() {
	/*
		flag.StringVar(&configPath, "path", "./", "Path where property files are stored")
		fileListPtr = flag.String("files", "", "Comma separated list of property and csv files")
		flag.StringVar(&system, "system", "", "Name of system for which properties should be considered")
		flag.BoolVar(&verbose, "v", false, "Print information to the console")
		flag.StringVar(&output, "output", localPropertiesFile, "Send output to a file or with '<console>' to console (default='local.properties')")
	*/

	flag.BoolVar(&latestFlag, "latest", false, "Add ':latest' tag, if patch version is missing")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Please specify what to do")
		return
	}

	manifest := parseManifest()

	switch strings.ToLower(flag.Args()[0]) {
	case "version":
		version := manifest.CommerceSuiteVersion
		if latestFlag && !strings.Contains(version, ".") {
			version = version + ":latest"
		}
		fmt.Print(version)
	case "addons":
		fmt.Println(manifest.StorefrontAddons)
	case "extensions":
		extensions := readExtensioins(manifest)
		fmt.Println(strings.Join(extensions, " "))
	}
}

func readExtensioins(manifest Manifest) []string {
	localExtensions := parseLocalextensions(manifest.UseConfig.Extensions.Location)

	var extensionsMap = make(map[string]bool)

	for _, extension := range localExtensions {
		extensionsMap[extension.Name] = true
	}
	for _, extension := range manifest.Extensions {
		extensionsMap[extension] = true
	}

	extensions := make([]string, len(extensionsMap))
	i := 0
	for key := range extensionsMap {
		extensions[i] = key
		i++
	}

	return extensions
}

func parseManifest() Manifest {

	jsonFile, err := ioutil.ReadFile("test-resources/manifest.json")

	if err != nil {
		fmt.Println(err)
	}
	var manifest Manifest
	json.Unmarshal(jsonFile, &manifest)
	//fmt.Printf("Version: %s, use cloud ext pack: %v - %v", com.CommerceSuiteVersion, com.UseCloudExtensionPack, com)

	//fmt.Println("----")
	//fmt.Printf("Properties: %v\n", com.Properties)

	validate(manifest)

	return manifest
}

func parseLocalextensions(location string) []XmlExtension {
	xmlFile, err := ioutil.ReadFile(location)
	if err != nil {
		fmt.Println(err)
	}
	var config Hybrisconfig
	xml.Unmarshal(xmlFile, &config)

	return config.Extensions.Extensions
}

func validate(com Manifest) {
	for _, property := range com.Properties {
		if err := property.Persona.isValid(); err != nil {
			fmt.Println("Error in property [" + property.Key + "]: " + err.Error())
		}
	}
	for _, aspect := range com.Aspects {
		for _, property := range aspect.Properties {
			if err := property.Persona.isValid(); err != nil {
				fmt.Println("Error in aspect '" + aspect.Name + "' property [" + property.Key + "]: " + err.Error())
			}
		}
	}
}

func (persona Persona) isValid() error {
	switch persona {
	case Development, Staging, Production, Empty:
		return nil
	}
	return errors.New("Invalid persona type: " + string(persona))
}

/*
func (persona *Persona) UnmarshalJSON(b []byte) error {
	var s string
	json.Unmarshal(b, &s)
	personaType := Persona(s)
	switch personaType {
	case Development, Staging, Production, "":
		*persona = personaType
		return nil
	}
	panic(errors.New("Invalid persona type"))
}*/
