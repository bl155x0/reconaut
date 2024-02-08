package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reconaut/framework/job"
	"reconaut/iobuffer"
	"reconaut/storage"
	"strings"
	"syscall"
	"time"
)

// start starts everything
func start(projectName, templateFileName string, verbose bool, variables []string) error {
	//first things first: init verbosity
	iobuffer.GetIOBuffer().Verbose = verbose

	//Setup signal handling
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, os.Interrupt, syscall.SIGTERM)

	//create a worker pool
	workerPool := job.NewWorkerPool(30, job.NewJobQueue())

	//load default template
	templateFile := job.GetTemplateFile(templateFileName)
	if templateFile == "" {
		return fmt.Errorf("cannot load template \"%s\": file not found", templateFileName)
	}
	template, err := job.NewTemplateFromFile(templateFile)
	if err != nil {
		return fmt.Errorf("cannot load template \"%s\": %v\n", templateFile, err)
	}

	//parse the cmd line
	paramMap, err := parse(variables)
	if err != nil {
		return fmt.Errorf("cannot parsing the commandline arguments: %v\n", err)
	}
	paramMap = applyDefaultVariables(projectName, paramMap)

	//Process the template provided and start
	if err := job.ProcessTemplate(template, paramMap, workerPool); err != nil {
		return fmt.Errorf("cannot process template \"%s\": %v\n", template.Filename, err)
	}
	//Start the worker pool
	workerPool.Start()
	run(workerPool, signalChanel)
	return nil
}

func applyDefaultVariables(projectName string, variables map[string]string) map[string]string {
	if isProjectRelatedDir() {
		if _, hostExists := variables[job.VarPrefix+"HOST"]; hostExists == false {
			variables[job.VarPrefix+"HOST"] = getProjectName(projectName)
		}
	}
	return variables
}

func isProjectRelatedDir() bool {
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	//if the current directory contains "." we asume that it is a recon domain directory
	//-> we use this directory name as our project name
	if strings.Contains(filepath.Base(dir), ".") {
		return true
	}
	return false
}

func getProjectName(projectName string) string {
	if projectName != "" {
		return projectName
	}

	dir, err := os.Getwd()
	if err != nil {
		return "reconaut"
	}

	//if the current directory contains "." we asume that it is a recon domain directory
	//-> we use this directory name as our project name
	if strings.Contains(filepath.Base(dir), ".") {
		return filepath.Base(dir)
	}
	return "reconaut"
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func list(projectName, query string, verbose bool) error {
	//first things first: init verbosity
	iobuffer.GetIOBuffer().Verbose = verbose

	projectName = getProjectName(projectName)
	nameSet := storage.SetFileStorageFileName(projectName)
	iobuffer.GetIOBuffer().AddOutput(fmt.Sprintf("Using project file \"%s\" for project \"%s\"...", nameSet, projectName))
	if fileExists(nameSet) == false {
		return fmt.Errorf("project file \"%s\" does not exist", nameSet)
	}
	flushIOBuffer()
	err := storage.GetSQLiteStorage().ListTable(query, func(col string, value any) {
		fmt.Printf("%v\n", value)
	})
	flushIOBuffer()
	return err
}

func parse(args []string) (map[string]string, error) {
	// Create a map to store variables
	variables := make(map[string]string)

	// Iterate over the variables
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		if len(parts) == 2 {
			variables[job.VarPrefix+parts[0]] = parts[1]
		} else {
			return nil, fmt.Errorf("invalid format for variable: %s", arg)
		}
	}

	return variables, nil
}

func printNextIOBuffer() {
	output := iobuffer.GetIOBuffer().GetOutput()
	if output != nil {
		printOutput(*output)
	}
}

func printOutput(output string) {
	fmt.Printf("[*] %s\n", output)
}

func flushIOBuffer() {
	for {
		output := iobuffer.GetIOBuffer().GetOutput()
		if output != nil {
			printOutput(*output)
		} else {
			break
		}
	}
}

func run(workerPool *job.WorkerPool, signalChanel chan os.Signal) {
	for {
		//In the loop: always look for new STDOUT to print
		printNextIOBuffer()
		time.Sleep(1 * time.Millisecond)

		//Also look for a termination signal
		select {
		case signal := <-signalChanel:
			fmt.Printf("Signal %v. Shutting down...", signal)
			workerPool.Stop()
			fmt.Printf("bye bye\n")
			return
		default:
		}

		if workerPool.Finished() {
			flushIOBuffer()
			fmt.Printf("[✔] ^‿^ Done!\n")
			return
		}
	}
}
