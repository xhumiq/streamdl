package main

import (
	"fmt"

	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"nex.com/nelib/microservice"
)

const (
	appName = "Monit"
)

func main() {
	app := NewApp(appName, "Monit Watcher")
	app.Cmd("check", func(c *cli.Context) error {
		svc := NewService(app)
		svc.checkService()
		return nil
	})
	file := ""
	app.Cmd("config-write <file>", func(c *cli.Context) error {
		svc := NewService(app)
		err := svc.WriteConfigs(file)
		checkError(err)
		return nil
	}, &file)
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt.Str("ConfigFile", config.Monitor.ConfigFile).
				Str("DefaultHost", config.Monitor.Host).
				Msgf("Monit Ver: %s", app.Build.Version)
			LogConfig(app, *config)
		}),
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			svc := NewService(app)
			app.RegisterService(svc)
			return app
		}))
	checkError(err)
}

func LogConfig(app *microservice.App, config AppConfig) {
	services, err := ReadConfigs(app, "")
	checkError(err)
	evt := log.Info()
	for _, s := range services {
		evt.Str(s.ServiceName, fmt.Sprintf("Host: %s:%d Url: %s %s", s.Host, s.Port, s.Url, s.Exe))
	}
	evt.Msgf("Services Checked")
}
