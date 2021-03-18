package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/rs/zerolog/log"

	"bitbucket.org/xhumiq/go-mclib/microservice"
)

const (
	appName = "cicd"
)

func main() {
	app := NewApp(appName,"Deployment Server")
	app.Cmd("status", func(c *cli.Context) error {
		jobMgr := NewJobManager(app)
		ss := GetStatus(jobMgr)
		b, _ := json.MarshalIndent(ss, "", "  ")
		println(string(b))
		return nil
	})
	repo := ""
	app.Cmd("run <repo>", func(c *cli.Context) error {
		mgr := NewJobManager(app)
		cfg, _ := mgr.Configs[strings.ToUpper(repo + "/master")]
		if cfg == nil {
			return fmt.Errorf("Repo %s / Branch master not found", repo)
		}
		djob := NewDeployJob(*cfg)
		mgr.Jobs[strings.ToUpper(repo + "/master")] = djob
		log.Info().Msgf("Start Job %s master", repo)
		djob, err := mgr.Run(djob, cfg)
		if err!=nil{
			return err
		}
		return nil
	}, &repo)
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt.Str("Repos Url", config.DefaultRepo.Repository)
			if config.DefaultRepo.RepoName != "" {
				evt = evt.Str("*Name", config.DefaultRepo.RepoName)
			}
			evt.Str("Branch", config.DefaultRepo.Branch).
				Str("DeployPath", config.DefaultRepo.DeployPath).
				Str("EnvPrefix", config.DefaultRepo.EnvPrefix).
				Msgf("CICD: %s", build.Version)
		}),
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			jobMgr := NewJobManager(app)
			app.RegisterWebService(NewApi(jobMgr))
			return jobMgr
		}))
	checkError(err)
}
