package job

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reconaut/iobuffer"
	"reconaut/storage"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	ENV_VAR string = "RECONAUT_TEMPLATES"
)

//-------------------------------------------------------------------------------------------------

type Template struct {
	Filename           string
	Name               string                      `yaml:"Name"`
	Description        string                      `yaml:"Description"`
	Commands           []Command                   `yaml:"Commands"`
	Root               Root                        `yaml:"Root"`
	StorageDefinitions []storage.StorageDefinition `yaml:"StorageDefinitions"`
}

type Root struct {
	Commands []CommandReference `yaml:"RootCommands"`
}

type Command struct {
	Name          string          `yaml:"Name"`
	Description   string          `yaml:"Description"`
	Exec          string          `yaml:"Exec"`
	ResultHandler []ResultHandler `yaml:"ResultHandler"`
}

type ResultHandler struct {
	RunCommand string             `yaml:"RunCommand"`
	Parameters []CommandParameter `yaml:"Parameters"`
}

type CommandParameter struct {
	Name  string `yaml:"Name"`
	Value string `yaml:"Value"`
}

type CommandReference struct {
	CommandName string `yaml:"Command"`
}

//-------------------------------------------------------------------------------------------------

type Result struct {
	name  string
	value string
}

type Variable struct {
	Name  string
	Value string
}

//-------------------------------------------------------------------------------------------------

// getVariableValueByName returns the first variable value found by the given name
func getVariableValueByName(vars []Variable, name string) *Variable {
	for _, v := range vars {
		if v.Name == name {
			return &v
		}
	}
	return nil

}

func replaceVariables(input string, variables []Variable) string {
	for _, v := range variables {
		// Construct the placeholder to match (assuming a word is surrounded by spaces)
		//placeholder := " " + v.Name + " "
		placeholder := v.Name
		// Replace all occurrences of the placeholder with the variable value
		//input = strings.ReplaceAll(input, placeholder, " "+v.Value+" ")
		input = strings.ReplaceAll(input, placeholder, v.Value)
	}
	return input

}

// getCommandByName returns the first command found by the given name
func getCommandByName(template *Template, cmdName string) *Command {
	for _, v := range template.Commands {
		if v.Name == cmdName {
			return &v
		}
	}
	return nil
}

//-------------------------------------------------------------------------------------------------

// ListTemplatesFiles will return all yaml file templates in the given directory
func ListTemplatesFiles(directory string) []string {
	iobuffer.GetIOBuffer().AddOutputVerbose(fmt.Sprintf("Searching template file in directory \"%s\"...", directory))
	if strings.HasSuffix(directory, "/") == false {
		directory = directory + "/"
	}
	var templates []string
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != directory {
			//Only files, not recursive
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") {
			templates = append(templates, path)
		}
		return nil
	})
	return templates
}

// GetTemplateFileFromDirectory will get template with the given name
func GetTemplateFileFromDirectory(directory, filename string) string {
	//Get available Templates from the template directory
	templateFiles := ListTemplatesFiles(directory)
	for _, av := range templateFiles {
		templateFile := filepath.Base(av)
		if templateFile == filename || templateFile == (filename+".reconaut.yaml") {
			iobuffer.GetIOBuffer().AddOutputVerbose(fmt.Sprintf("Using template file \"%s\".", av))
			return av
		}

	}
	return ""
}

func GetTemplateFile(filename string) string {
	value, exists := os.LookupEnv(ENV_VAR)
	if exists && value != "" {
		pathes := strings.Split(value, ":")
		for _, path := range pathes {
			templateFile := GetTemplateFileFromDirectory(path, filename)
			if templateFile != "" {
				return templateFile
			}
		}
	}
	return GetTemplateFileFromDirectory(".", filename)
}

// LoadTemplate loads a template from the given file
func NewTemplateFromFile(filename string) (*Template, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	template := &Template{Filename: filename}
	err = yaml.Unmarshal(content, template)
	if err != nil {
		return nil, err
	}
	iobuffer.GetIOBuffer().AddOutput(fmt.Sprintf("Template \"%s\" loaded successfully.", template.Name))
	return template, nil
}

// ProcessTemplate processes the given template with the given WorkerPool
func ProcessTemplate(template *Template, rootVariables map[string]string, workerPool *WorkerPool) error {
	iobuffer.GetIOBuffer().AddOutput(fmt.Sprintf("Processing template \"%s\"...", template.Name))

	// Check for for commands
	if len(template.Root.Commands) == 0 {
		return fmt.Errorf("missing root commands")
	}

	// Geting global root variables from the caller
	var rootVariableValues []Variable
	for name, value := range rootVariables {
		rootVariableValues = append(rootVariableValues,
			Variable{
				Name:  name,
				Value: value,
			})
	}

	//Get all the root commands to execute
	var jobs []Job
	for _, rootCommand := range template.Root.Commands {
		cmd := getCommandByName(template, rootCommand.CommandName)
		if cmd == nil {
			return fmt.Errorf("unable to find command \"%s\"", rootCommand.CommandName)
		}
		jobs = append(jobs, NewTemplateJob(template, *cmd, rootVariableValues))
	}
	workerPool.DoJobs(jobs)
	return nil
}
