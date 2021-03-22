package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"strings"
)

const (
	appName = "DavCli"
)

func main() {
	app := NewApp(appName, "DavCli Service")
	days, listpath := 3, ""
	app.Cmd("list [-d,--days <recency>] <listpath>", func(c *cli.Context) error {
		if listpath == ""{
			listpath = "/"
		}
		app.ShowVersion()
		s := NewService(app)
		if days < 1{
			days = 2
		}
		println("List Days", listpath)
		s.ListFiles(days, common.FilterEmptyStrings(strings.Split(listpath, ",")...)...)
		println("End")
		return nil
	}, &days, &listpath)
	app.Cmd("sync [-d,--days <recency>] <paths>", func(c *cli.Context) error {
		if listpath == ""{
			listpath = "/"
		}
		app.ShowVersion()
		s := NewService(app)
		if days < 1{
			days = 2
		}
		files, err := s.GetLatestFiles(days, common.FilterEmptyStrings(strings.Split(listpath, ",")...)...)
		if err!=nil{
			return err
		}
		basePath := "$HOME/Videos/zsf"
		return s.client.SaveFilesSD(6, basePath, files...)
	}, &days, &listpath)
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
