package main

import (
	"bitbucket.org/xhumiq/go-mclib/api"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/storage"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"strings"
)

const (
	appName = "DavCli"
)

func main() {
	app := NewApp(appName, "WebDav Download Client")
	listpath := ""
	app.Cmd("list <listpath>", func(c *cli.Context) error {
		if listpath == ""{
			listpath = "/"
		}
		app.ShowVersion()
		s := NewService(app)
		days := s.SvcConfig.Download.HistoryDays
		s.ListFiles(days, common.FilterEmptyStrings(strings.Split(listpath, ",")...)...)
		return nil
	}, &listpath)
	app.Cmd("sync <paths>", func(c *cli.Context) error {
		if listpath == ""{
			listpath = "/"
		}
		app.ShowVersion()
		s := NewService(app)
		days := s.SvcConfig.Download.HistoryDays
		files, err := s.GetLatestFiles(days, common.FilterEmptyStrings(strings.Split(listpath, ",")...)...)
		if err!=nil{
			return err
		}
		opts := api.WebDavSaveFileOptions{
			Threads:           s.SvcConfig.Download.DownloadThreads,
			BasePath:          s.SvcConfig.Download.BaseTargetPath,
			ForceOverwrite:    s.SvcConfig.Download.ForceOverwrite,
			SyncSize:          true,
			SyncModDate:       false,
			ReplacePaths: map[string]string{},
			SegmentalDownload: false,
			Downloaded: map[string]*api.DavFileInfo{},
		}
		opts.ReplacePaths["/Video"] = s.SvcConfig.Download.VideoPath
		opts.ReplacePaths["/Audio"] = s.SvcConfig.Download.AudioPath
		opts.ReplacePaths["/LiteralCenter"] = s.SvcConfig.Download.DocsPath
		opts.ReplacePaths["/Materials"] = s.SvcConfig.Download.SchoolPath
		opts.ReplacePaths["/Photos"] = s.SvcConfig.Download.PhotosPath
		opts.ReplacePaths["/Hymns"] = s.SvcConfig.Download.HymnsPath
		return s.client.SaveFilesSD(opts, files...)
	}, &listpath)
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt.Str("WebDav UserName", config.Download.HebronUser).
				Str("Sync    Base Path", storage.ConvertPathUNC(config.Download.BaseTargetPath)).
				Str("Sync   Video Path", storage.ConvertPathUNC(config.Download.VideoPath)).
				Str("Sync   Audio Path", storage.ConvertPathUNC(config.Download.AudioPath)).
				Str("Sync   Hymns Path", storage.ConvertPathUNC(config.Download.HymnsPath)).
				Str("Sync  Photos Path", storage.ConvertPathUNC(config.Download.PhotosPath)).
				Str("Sync  School Path", storage.ConvertPathUNC(config.Download.SchoolPath)).
				Str("Sync     Lit Path", storage.ConvertPathUNC(config.Download.DocsPath)).
				Int("Sync History Days", config.Download.HistoryDays).
				Int("Sync ---- Threads", config.Download.DownloadThreads).
				Bool("Sync ForceReplace", config.Download.ForceOverwrite).
				Str("WebDav   Domain", config.Download.Domain).
				Str("WebDav Password", common.MaskedSecret(config.Download.HebronPwd)).
				Str("Vault Path", storage.ConvertPathUNC(config.Download.VaultFile)).
				Msgf("DavCli: %s", build.Version)
		}))
	checkError(err)
}
