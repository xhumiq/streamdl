package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"fmt"
	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"time"
)

const (
	appName = "streamdl"
)

func main() {
	app := NewApp(appName, "Stream Downloader")
	url, temp, opath, prefix, mins := "", "", "", "", 0
	ffmpeg, rec, fopts, ropts := "", "", "",  ""
	force := false
	app.Cmd("record [-t,--tempPath <file://c:/tmp/streams/>] [-o,--outputPath <file://d:/news/2021>] [-p,--prefix <FTV>] [-m,--minutes <70>] [-f,--ffmpeg-bin <\"C:\\app\\Media\\ffmpeg-3.4.1\\bin\\ffmpeg.exe\">] [-r,--recorder-bin <\"C:\\app\\utils\\youtube-dl.exe\">] [--ffmpeg-options <\"-err_detect ignore_err -c copy\">] [--rec-options <\"-f 96 %s -o -\">] [--force <force>] <url>", func(c *cli.Context) error {
		config := app.Config.(*AppConfig)
		config.Target.Site = common.StringDefault(&url, config.Target.Site)
		config.Target.Site = cleanUrl(config.Target.Site)
		config.Recorder.TempPath = common.StringDefault(&temp, config.Recorder.TempPath)
		config.Ffmpeg.OutputPath = common.StringDefault(&opath, config.Ffmpeg.OutputPath)
		config.Target.Prefix = common.StringDefault(&prefix, config.Target.Prefix)
		config.Recorder.Minutes = common.IntDefault(&mins, config.Recorder.Minutes)
		config.Ffmpeg.Bin = common.StringDefault(&ffmpeg, config.Ffmpeg.Bin)
		config.Recorder.Bin = common.StringDefault(&rec, config.Recorder.Bin)
		config.Ffmpeg.Options = common.StringDefault(&fopts, config.Ffmpeg.Options)
		config.Recorder.Options = common.StringDefault(&ropts, config.Recorder.Options)
		if config.Target.Site == ""{
			return fmt.Errorf("Url is not specified")
		}
		if config.Recorder.TempPath == ""{
			return fmt.Errorf("Temp Path is not specified")
		}
		if config.Ffmpeg.OutputPath == ""{
			return fmt.Errorf("Output Path is not specified")
		}
		if config.Target.Prefix == ""{
			return fmt.Errorf("Prefix is not specified")
		}
		if config.Recorder.Minutes < 1{
			return fmt.Errorf("Recorder Minutes is not specified")
		}
		if config.Ffmpeg.Bin == ""{
			return fmt.Errorf("FFMpeg Bin is not specified")
		}
		if config.Recorder.Bin == ""{
			return fmt.Errorf("Recorder Bin is not specified")
		}
		if config.Ffmpeg.Options == ""{
			return fmt.Errorf("FFMpeg Options is not specified")
		}
		if config.Recorder.Options == ""{
			return fmt.Errorf("Recorder Options is not specified")
		}
		app.ShowVersion()
		tf := filepath.Join(config.Recorder.TempPath, createFileName(config.Target.Prefix, "tmp"))
		cmd, err := createRecCmd(config.Recorder.Bin, config.Target.Site, tf, force)
		checkError(err)
		err = execCommand(cmd, time.Duration(config.Recorder.Minutes) * time.Minute)
		checkError(err)
		if !common.FileExists(tf) && common.FileExists(tf + ".part"){
			println("Rename data")
			err = os.Rename(tf + ".part", tf)
			checkError(err)
		}
		of := filepath.Join(config.Ffmpeg.OutputPath, createFileName(config.Target.Prefix, ""))
		cmd, err = createFFMpegCmd(config.Ffmpeg.Bin, tf, of, true)
		checkError(err)
		err = execCommand(cmd, time.Duration(config.Recorder.Minutes) * time.Minute)
		checkError(err)
		log.Info().Msgf("Completed download of stream -> %s", of)
		return nil
	}, &temp, &opath, &prefix, &mins, &ffmpeg, &rec, &fopts, &ropts, &force, &url)
	err := app.Run(
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			return app
		}))
	checkError(err)
}
