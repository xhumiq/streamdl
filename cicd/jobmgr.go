package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"bitbucket.org/xhumiq/go-mclib/microservice"
)

const defaultShellPOSIX = "/bin/sh"
const defaultShOpt = "-c"

func NewJobManager(app *microservice.App) *JobManager {
	config := app.Config.(*AppConfig)
	mgr := JobManager{
		App: app,
		SvcConfig:  config,
		Configs: make(map[string]*DeployConfig),
		Jobs:    make(map[string]*DeployContext),
		ErrorCh: make(chan error, 100),
		mutex:   &sync.Mutex{},
	}
	if len(config.Repos) > 0 {
		for name, repo := range config.Repos {
			ns := strings.Split(name, "/")
			cfg := NewDeployConfig(ns[0], ns[1], repo.DeployPath, repo.GitPath, repo.ConfigPath)
			cfg.Commands = NewShellCmds(repo.Commands)
			cfg.Env = repo.Env
			cfg.EnvPrefix = repo.EnvPrefix
			cfg.Repository = repo.Repository
			cfg.CertPath = repo.CertPath
			cfg.Timeout = repo.Timeout
			if config.DefaultRepo.DeployRootPath != "" {
				cfg.DeployPath = path.Join(config.DefaultRepo.DeployRootPath, cfg.DeployPath)
			}
			if config.DefaultRepo.GitRootPath != "" {
				cfg.GitPath = path.Join(config.DefaultRepo.GitRootPath, cfg.GitPath)
			}
			mgr.Configs[strings.ToUpper(cfg.RepoName+"/"+cfg.Branch)] = cfg
		}
	} else if config.DefaultRepo.RepoName == "" {
		checkError(errors.Errorf("Repo name needs to be configured"))
	} else if config.DefaultRepo.Branch == "" {
		checkError(errors.Errorf("Repo Branch needs to be configured"))
	} else {
		branches := strings.Split(config.DefaultRepo.Branch, ",")
		for _, branch := range branches {
			branch = strings.TrimSpace(branch)
			if len(branch) < 1 {
				continue
			}
			cfg := NewDeployConfig(config.DefaultRepo.RepoName, branch, config.DefaultRepo.DeployPath,
				config.DefaultRepo.GitPath, config.DefaultRepo.ConfigPath)
			if config.DefaultRepo.Repository != "" {
				cfg.Repository = config.DefaultRepo.Repository
			}
			if config.DefaultRepo.DeployRootPath != "" {
				cfg.DeployPath = path.Join(config.DefaultRepo.DeployRootPath, cfg.DeployPath)
			}
			if config.DefaultRepo.GitRootPath != "" {
				cfg.GitPath = path.Join(config.DefaultRepo.GitRootPath, cfg.GitPath)
			}
			if config.ConfigRootPath != "" {
				cfg.ConfigPath = path.Join(config.ConfigRootPath, cfg.ConfigPath)
			}
			if config.DefaultRepo.CertPath != "" {
				cfg.CertPath = config.DefaultRepo.CertPath
			}
			if config.DefaultRepo.EnvPrefix != "" {
				cfg.EnvPrefix = config.DefaultRepo.EnvPrefix
			}
			mgr.Configs[strings.ToUpper(cfg.RepoName+"/"+cfg.Branch)] = cfg
		}
	}
	b, _ := json.MarshalIndent(mgr.Configs, "", "  ")
	log.Info().RawJSON("Configs", b).Msgf("Repo Configs Count: %d. Default Repo: %s / %s", len(config.Repos), config.DefaultRepo.RepoName, config.DefaultRepo.Branch)
	return &mgr
}
func (mgr *JobManager) GetStatus(repo string, branch string) (*DeployContext, error) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	djob, _ := mgr.Jobs[strings.ToUpper(strings.TrimSpace(repo)+"/"+strings.TrimSpace(branch))]
	return djob, nil
}

func (mgr *JobManager) QueueDeployment(repo string, branch string) (*DeployContext, error) {
	repo = strings.TrimSpace(repo)
	branch = strings.TrimSpace(branch)
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	djob, _ := mgr.Jobs[strings.ToUpper(repo+"/"+branch)]
	if djob != nil {
		if djob.Error != nil || djob.Status == "Completed" || djob.Status == "Error" {
			println("Last run errored out - will try again")
			djob = nil
		} else {
			println("Queue currently running. Request will be queued")
		}
	}
	if djob == nil {
		cfg, _ := mgr.Configs[strings.ToUpper(repo+"/"+branch)]
		if cfg == nil {
			log.Info().Msgf("Repo %s / Branch %s not found\n", repo, branch)
			return nil, fmt.Errorf("Repo %s / Branch %s not found", repo, branch)
		}
		djob = NewDeployJob(*cfg)
		mgr.Jobs[strings.ToUpper(repo+"/"+branch)] = djob
		// mgr.ErrorCh = make(chan error, 100)
		go func(job *DeployContext) {
			log.Info().Msgf("Start Queue: %s %s\n", repo, branch)
			_, err := mgr.StartQueue(job)
			if err != nil {
				log.Error().Str("Trace", fmt.Sprintf("%+v", err)).Msgf("Queue Error: %s %s", repo, branch)
				job.Error = err
				job.Status = "Completed"
				if err != nil {
					job.Status = "Error"
				}
				mgr.ErrorCh <- err
			}
		}(djob)
	}
	if len(djob.Requests) >= 3 {
		return nil, fmt.Errorf("Too many requests in queue. Please wait")
	}
	djob.Requests <- &djob.Config
	return djob, nil
}

func NewDeployJob(config DeployConfig) *DeployContext {
	return &DeployContext{
		Config:   config,
		Requests: make(chan *DeployConfig, 100),
		mutex:    sync.Mutex{},
		reqWait:  sync.WaitGroup{},
	}
}

func (mgr *JobManager) StartQueue(job *DeployContext) (*DeployContext, error) {
	for {
		var cfg *DeployConfig
		select {
		case err := <-mgr.ErrorCh:
			{
				return nil, err
			}
		case cfg = <-job.Requests:
			{
				job.Status = "Running"
				var err error
				job, err = mgr.Run(job, cfg)
				if err!=nil{
					return nil, err
				}
			}
		}
	}
	return job, nil
}

func (job *DeployContext) CreateDefaultCommands() []*DeployCommand {
	fmt.Printf("Analyze %s Dir\n", job.Config.DeployPath)
	pkg := path.Join(job.Config.DeployPath, "package.json")
	if _, err := os.Stat(pkg); !os.IsNotExist(err) {
		return []*DeployCommand{
			&DeployCommand{"npm install", nil, ""},
			&DeployCommand{"npm run build", nil, ""},
		}
	}
	pkg = path.Join(job.Config.DeployPath, "CICD_BUILD.sh")
	if _, err := os.Stat(pkg); !os.IsNotExist(err) {
		return []*DeployCommand{
			&DeployCommand{"source CICD_BUILD.sh", nil, ""},
		}
	}
	pkg = path.Join(job.Config.DeployPath, "Makefile")
	if _, err := os.Stat(pkg); !os.IsNotExist(err) {
		return []*DeployCommand{
			&DeployCommand{"source CICD_BUILD.sh", nil, ""},
		}
	}
	return nil
}

func (mgr *JobManager) Run(job *DeployContext, cfg *DeployConfig)(*DeployContext, error) {
	job.Status = "Running"
	log.Info().Msgf("Process Request: %s %s", cfg.Repository, cfg.Branch)

	if _, err := os.Stat(cfg.DeployPath); os.IsNotExist(err) {
		if err = os.MkdirAll(cfg.DeployPath, 0755); err != nil {
			return nil, errors.Wrapf(err, "Unable to create git dir")
		}
	}
	cmds, err := mgr.CreateGitCheckoutCommands(*cfg)
	if err != nil {
		return nil, err
	}
	for idx := 0; idx < len(cmds); idx++ {
		pc := ProcessContext{
			CmdIndex: idx,
			Command:  *cmds[idx],
			Started:  time.Now(),
		}
		err := job.RunNextCommnd(&pc)
		if err != nil {
			fmt.Printf("Command Execute Error: %+v\n", err)
			pc.Error = err
			job.Status = "Completed"
			return nil, err
		}
	}
	cmds = job.Config.Commands
	if len(cmds) < 1 {
		cmds = job.CreateDefaultCommands()
	}
	for idx := 0; idx < len(cmds); idx++ {
		pc := ProcessContext{
			CmdIndex: idx,
			Command:  *cmds[idx],
			Started:  time.Now(),
		}
		err := job.RunNextCommnd(&pc)
		if err != nil {
			fmt.Printf("Command Execute Error: %+v\n", err)
			pc.Error = err
			break
		}
	}
	job.Status = "Completed"
	return job, nil
}

func (job *DeployContext) RunNextCommnd(pc *ProcessContext) error {
	job.Current = pc
	if pc.CmdIndex >= len(job.History) {
		job.History = append(job.History, pc)
	} else {
		job.History[pc.CmdIndex] = pc
	}
	if pc.Command.Thunk != nil {
		if err := pc.Command.Thunk(job, pc); err != nil {
			return err
		}
	} else if pc.Command.Command != "" {
		if err := job.RunNextShellCommnd(&pc.Command, pc); err != nil {
			return err
		}
	}
	return nil
}
func (job *DeployContext) RunNextThunk(thunk func(), pc *ProcessContext) error {
	if err := pc.Command.Thunk(job, pc); err != nil {
		return pc.SetErrorf(err, "Unable to run job section: %d", pc.CmdIndex)
	}
	pc.Completed = true
	pc.Finished = new(time.Time)
	*pc.Finished = time.Now()
	return nil
}
func (job *DeployContext) RunNextShellCommnd(shellCmd *DeployCommand, pc *ProcessContext) error {
	cmd, err := job.CreateComand(*shellCmd)
	if err != nil {
		return pc.SetErrorf(err, "Unable to create shell command. %d", pc.CmdIndex)
	}
	pc.CaptureIO(cmd)
	log.Info().Str("Command", shellCmd.Command).
		Str("WorkDir", cmd.Dir).
		Msgf("Start Command")
	if err := cmd.Start(); err != nil {
		println("Error on Command")
		return pc.SetErrorf(err, "Unable to run shell command. %d", pc.CmdIndex)
	}
	mc := make(chan struct{})
	go func() {
		defer close(mc)
		err := cmd.Wait()
		for _, ln := range pc.Output{
			log.Debug().Msgf(ln)
		}
		if err != nil {
			b, _ := cmd.CombinedOutput()
			if len(b) > 0{
				log.Error().Msgf(string(b))
			}
			pc.SetErrorf(err, "Error running shell command. %s", pc.Command.Command)
		}
	}()
	to := job.Config.Timeout
	if to < 1 {
		to = 300
	}
	timeOut := time.Duration(to) * time.Second
	select {
	case <-mc:
	case <-time.After(timeOut):
		return errors.Errorf("Timeout on command")
	}
	pc.Completed = true
	pc.Finished = new(time.Time)
	*pc.Finished = time.Now()
	return pc.Error
}

func (job *DeployContext) CreateComand(shellCmd DeployCommand) (*exec.Cmd, error) {
	args := strings.Split(shellCmd.Command, " ")
	cmdText := strings.Join(args, " ")
	args = append([]string{defaultShOpt}, args...)
	var cmd *exec.Cmd
	if job.Config.Timeout > 0 {
		ctx := context.Background()
		ctx, job.CancelFn = context.WithTimeout(ctx, time.Duration(job.Config.Timeout)*time.Second)
		cmd = exec.CommandContext(ctx, defaultShellPOSIX, defaultShOpt, cmdText) // #nosec
	} else {
		cmd = exec.Command(defaultShellPOSIX, defaultShOpt, cmdText)
	}
	if shellCmd.WorkingDir != "" {
		cmd.Dir = shellCmd.WorkingDir
	} else {
		cmd.Dir = job.Config.DeployPath
	}
	if job.Config.Env != nil && len(job.Config.Env) > 0 {
		for key, value := range job.Config.Env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	return cmd, nil
}

func (pc *ProcessContext) CaptureIO(cmd *exec.Cmd) *ProcessContext {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return pc
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return pc
	}
	pc.StartIOStreams(stdout, stderr)
	return pc
}

func (pc *ProcessContext) StartIOStreams(stdout io.ReadCloser, stderr io.ReadCloser) *ProcessContext {
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			b := scanner.Bytes()
			if len(b) < 1{
				continue
			}
			line := strings.TrimSpace(string(b))
			if len(line) < 1{
				continue
			}
			pc.Output = append(pc.Output, line)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			b := scanner.Bytes()
			if len(b) < 1{
				continue
			}
			line := strings.TrimSpace(string(b))
			if len(line) < 1{
				continue
			}
			log.Error().Msgf(line)
			pc.Output = append(pc.Output, line)
		}
	}()
	return pc
}

func (pc *ProcessContext) SetErrorf(err error, msg string, args ...interface{}) error {
	pc.Error = err
	if err != nil {
		pc.Output = append(pc.Output, fmt.Sprintf(msg, args))
		pc.Output = append(pc.Output, fmt.Sprintf("%+v", err))
		log.Error().Str("Trace", fmt.Sprintf("%+v", err)).Msgf(msg, args...)
	}
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				pc.ExitCode = status.ExitStatus()
			}
		}
	}
	pc.ErrorResponse = &ResponseError{
		Message:    err.Error(),
		StackTrace: fmt.Sprintf("%+v", err),
	}
	pc.Completed = true
	ts := time.Now()
	pc.Finished = &ts
	return err
}
