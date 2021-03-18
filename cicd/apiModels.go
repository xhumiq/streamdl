package main

import (
	"fmt"
	"time"
)

type SystemStatus struct {
	Version    string                    `json:"version,omitempty" yaml:"version,omitempty"`
	GitHash    string                    `json:"gitHash,omitempty" yaml:"gitHash,omitempty"`
	BuildStamp string                    `json:"buildStamp,omitempty" yaml:"buildStamp,omitempty"`
	Config     AppConfig                 `json:"config,omitempty" yaml:"config,omitempty"`
	Statuses   map[string]*ProcessStatus `json:"statuses,omitempty" yaml:"statuses,omitempty"`
}

type ProcessStatus struct {
	Repository      string                  `json:"repository,omitempty" yaml:"repository,omitempty"`
	RequestCount    int                     `json:"requestCount,omitempty" yaml:"requestCount,omitempty"`
	CommandCount    int                     `json:"commandCount,omitempty" yaml:"commandCount,omitempty"`
	RepoName        string                  `json:"repoName,omitempty" yaml:"repoName,omitempty"`
	Branch          string                  `json:"branch,omitempty" yaml:"branch,omitempty"`
	DeployPath      string                  `json:"deployPath,omitempty" yaml:"deployPath,omitempty"`
	GitPath         string                  `json:"gitPath,omitempty" yaml:"gitPath,omitempty"`
	ConfigPath      string                  `json:"configPath,omitempty" yaml:"configPath,omitempty"`
	EnvPrefix       string                  `json:"envPrefix,omitempty" yaml:"envPrefix,omitempty"`
	Timeout         int                     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RepoError       *ResponseError          `json:"repoError,omitempty" yaml:"repoError,omitempty"`
	Status          string                  `json:"status,omitempty" yaml:"status,omitempty"`
	CommandStatus   *ProcessCommandStatus   `json:"commandStatus,omitempty" yaml:"commandStatus,omitempty"`
	Logs            []*ProcessCommandStatus `json:"logs,omitempty" yaml:"logs,omitempty"`
	LastStarted     *time.Time              `json:"lastStarted,omitempty" yaml:"lastFtarted,omitempty"`
	LastFinished    *time.Time              `json:"lastFinished,omitempty" yaml:"lastFinished,omitempty"`
	LastStatus      string                  `json:"lastStatus,omitempty" yaml:"lastStatus,omitempty"`
	TimeToLastStart string                  `json:"timeToLastStart,omitempty" yaml:"timeToLastStart,omitempty"`
}

type ProcessCommandStatus struct {
	CmdIndex  int            `json:"cmdIndex,omitempty" yaml:"cmdIndex,omitempty"`
	Command   string         `json:"command,omitempty" yaml:"command,omitempty"`
	ExitCode  int            `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`
	Started   time.Time      `json:"started,omitempty" yaml:"started,omitempty"`
	Finished  *time.Time     `json:"finished,omitempty" yaml:"finished,omitempty"`
	Duration  *string        `json:"duration,omitempty" yaml:"duration,omitempty"`
	Error     *ResponseError `json:"error,omitempty" yaml:"error,omitempty"`
	Completed bool           `json:"completed,omitempty" yaml:"completed,omitempty"`
	Output    []string       `json:"output,omitempty" yaml:"output,omitempty"`
}

func NewProcessStatusFromContext(repo *DeployContext) *ProcessStatus {
	s := NewProcessStatus(repo.Config, repo.Current)
	s.RequestCount = len(repo.Requests)
	s.Status = repo.Status
	if repo.Error != nil {
		s.RepoError = &ResponseError{
			Message:    repo.Error.Error(),
			StackTrace: fmt.Sprintf("%+v", repo.Error),
		}
	}
	return s
}

func NewProcessStatus(config DeployConfig, context *ProcessContext) *ProcessStatus {
	status := &ProcessStatus{
		Repository:   config.Repository,
		RepoName:     config.RepoName,
		Branch:       config.Branch,
		DeployPath:   config.DeployPath,
		GitPath:      config.GitPath,
		ConfigPath:   config.ConfigPath,
		EnvPrefix:    config.EnvPrefix,
		CommandCount: len(config.Commands),
		Timeout:      config.Timeout,
	}
	if context != nil {
		status.CommandStatus = NewCommandStatus(context)
		status.LastStarted = new(time.Time)
		*status.LastStarted = status.CommandStatus.Started
		status.LastFinished = status.CommandStatus.Finished
		status.TimeToLastStart = time.Now().Sub(*status.LastStarted).String()
		if status.CommandStatus.Error != nil {
			status.LastStatus = "Error"
		} else if status.CommandStatus.Completed {
			status.LastStatus = "Done"
		} else {
			status.LastStatus = "Pending"
		}
	}
	return status
}

func NewCommandStatus(context *ProcessContext) *ProcessCommandStatus {
	if context != nil {
		status := ProcessCommandStatus{
			CmdIndex:  context.CmdIndex,
			Command:   context.Command.Command,
			Error:     context.ErrorResponse,
			ExitCode:  context.ExitCode,
			Started:   context.Started,
			Finished:  context.Finished,
			Output:    context.Output,
			Completed: context.Completed,
		}
		if status.Finished != nil {
			status.Duration = new(string)
			*status.Duration = status.Finished.Sub(status.Started).String()
		}
		return &status
	}
	return nil
}

func (resp *ProcessStatus) SetError(err error) *ResponseError {
	resp.RepoError = &ResponseError{
		Message:    err.Error(),
		StackTrace: fmt.Sprintf("%+v", err),
	}
	return resp.RepoError
}

type ResponseError struct {
	Message    string `json:"message,omitempty" yaml:"message,omitempty"`
	StackTrace string `json:"stackTrace,omitempty" yaml:"stackTrace,omitempty"`
}
