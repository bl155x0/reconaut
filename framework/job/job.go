package job

import (
	"fmt"
	"os/exec"
	"reconaut/iobuffer"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-shellwords"
)

//-------------------------------------------------------------------------------------------------

// JobStatus represents the status of the job
type JobStatus string

// JobResult defines the result of a completed job
type JobResult struct {
	//TODO
}

const VarPrefix = "$$"

const (
	// JobStatusNotStarted represents a job which is not started yet
	JobStatusNotStarted = "NotStarted"

	// JobStatusRunning represents a job which is currently running
	JobStatusRunning = "Running"

	// JobStatusCompleted represents a job which is completed
	JobStatusCompleted = "Completed"

	// JobStatusAborted represents a job which was aborted
	JobStatusAborted = "Aborted"
)

// Job interface defines the methods that a job should implement
type Job interface {

	// Gets the ID of the job
	ID() string

	// Execute is the method that performs the job's task
	// It returns new, resulting, jobs to execute
	Execute() []Job

	// Abort is the method that aborts the job
	Abort()

	// GetStatus is the method that retrieves the status of the job
	GetStatus() JobStatus
}

//-------------------------------------------------------------------------------------------------

// TemplateJob is a job which executes command from a template
type TemplateJob struct {
	status    JobStatus
	id        string
	command   Command
	template  *Template
	variables []Variable
}

// NewTemplateJob creates a new job which can execute the given command
func NewTemplateJob(template *Template, command Command, variables []Variable) *TemplateJob {
	return &TemplateJob{
		status:    JobStatusNotStarted,
		id:        uuid.New().String(),
		command:   command,
		template:  template,
		variables: variables,
	}
}

func (job *TemplateJob) ID() string {
	return job.id
}

func (job *TemplateJob) IsVerbose() bool {
	return iobuffer.GetIOBuffer().Verbose
}

func (job *TemplateJob) Print(str string) {
	prefix := fmt.Sprintf("[JOB %s] [%s]", job.ID(), job.command.Name)
	iobuffer.GetIOBuffer().AddOutput(fmt.Sprintf("%s %s", prefix, str))
}

func (job *TemplateJob) PrintVerbose(str string) {
	prefix := fmt.Sprintf("[JOB %s] [%s]", job.ID(), job.command.Name)
	iobuffer.GetIOBuffer().AddOutputVerbose(fmt.Sprintf("%s %s", prefix, str))
}

// Execute implements the Execute method of the Job interface for ExampleJob
func (job *TemplateJob) Execute() []Job {
	if job.template == nil {
		panic("template cannot be nil")
	}
	job.status = JobStatusRunning

	//Create the OS Command
	var cmd *exec.Cmd
	tokens, err := shellwords.Parse(job.command.Exec)
	if err != nil {
		panic("cannot parse shellwords: " + job.command.Exec)
	}
	if len(tokens) == 1 {
		//Single command with no arguments
		tokens[0] = replaceVariables(tokens[0], job.variables)
		cmd = exec.Command(tokens[0])
	} else {
		//Command with arguments
		arguments := tokens[1:]

		//Replace variable in arguments
		for i, argument := range arguments {
			arguments[i] = replaceVariables(argument, job.variables)
		}
		cmd = exec.Command(tokens[0], arguments...)

		//check if we still have un-replaced variables in the command
		for _, argument := range arguments {
			if strings.Contains(argument, VarPrefix) {
				job.Print(fmt.Sprintf("unknown variable %s in command \"%s\"", argument, cmd.String()))
				return nil
			}
		}
	}

	startTime := time.Now()
	job.Print(fmt.Sprintf("Job started (\"%s\")", cmd.String()))

	//Execute the command
	//var stdout, stderr bytes.Buffer
	//cmd.Stdout = &stdout
	//cmd.Stderr = &stderr
	outputData, err := cmd.CombinedOutput()
	output := string(outputData)
	if err != nil {
		//The command failed
		job.Print(fmt.Sprintf("The command failed (%v):\n%s", err, output))
		job.status = JobStatusCompleted
		return nil
	}

	//Process RsultHandler
	var nextJobs []Job
	for _, resultHandler := range job.command.ResultHandler {
		command := getCommandByName(job.template, resultHandler.RunCommand)
		if command == nil {
			job.Print(fmt.Sprintf("unknown command %s", resultHandler.RunCommand))
			return nil
		}
		var vars []Variable
		//Also append the master HOST variable
		//TODO:  introduce global variables
		//vars = append(vars, job.variables...)
		for _, parameter := range resultHandler.Parameters {
			vars = append(vars,
				Variable{
					Name:  parameter.Name,
					Value: parameter.Value,
				})
		}
		nextJobs = append(nextJobs, NewTemplateJob(job.template, *command, vars))
	}

	job.PrintVerbose(fmt.Sprintf("%d next jobs", len(nextJobs)))
	stopTime := time.Now()
	elapsedTime := stopTime.Sub(startTime)
	if job.IsVerbose() && strings.TrimSpace(output) != "" {
		job.PrintVerbose(fmt.Sprintf("Job finished (\"%s\") (%s)\n%s", cmd.String(), elapsedTime, output))
	} else {
		job.Print(fmt.Sprintf("Job finished (\"%s\") (%s)", cmd.String(), elapsedTime))
	}
	job.status = JobStatusCompleted
	return nextJobs
}

func (job *TemplateJob) Abort() {
	job.status = JobStatusAborted
}

// GetStatus implements the GetStatus method of the Job interface for ExampleJob
func (job *TemplateJob) GetStatus() JobStatus {
	return job.status
}

//-------------------------------------------------------------------------------------------------
