package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const (
	appName = "DavCli"
)

func main() {
	app := NewApp(appName, "DavCli Service")
	path := ""
	app.Cmd("list", func(c *cli.Context) error {
		if path == ""{
			path = "/"
		}
		app.ShowVersion()
		s := NewService(app)
		return s.ListFiles(path)
	}, &path)
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt = evt.Str("WebDav UserName", config.Download.HebronUser).
				Str("Sync Video Path", config.Download.VideoPath).
				Str("Sync Audio Path", config.Download.AudioPath).
				Str("WebDav   Domain", config.Download.Domain).
				Str("WebDav Password", common.MaskedSecret(config.Download.HebronPwd))
			evt.Msgf("DavCli: %s", build.Version)
		}))
	checkError(err)
}
