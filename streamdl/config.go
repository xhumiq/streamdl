package main

import (
	"bitbucket.org/xhumiq/go-mclib/storage"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
)

type AppConfig struct {
	Smtp   common.SmtpConfig
	Log    common.LogConfig  `json:"LOG" yaml:"log"`
	Ffmpeg struct {
		Bin        string `default:"ffmpeg" env:"FFMPEG_BIN" json:"FFMPEG_PATH" yaml:"bin"`
		OutputPath string `default:"" env:"OUTPUT_PATH" json:"OUTPUT_PATH" yaml:"output"`
		Options    string `default:"" env:"FFMPEG_OPTIONS" json:"FFMPEG_OPTIONS" yaml:"options"`
	} `json:"FFMPEG" yaml:"ffmpeg"`
	Target struct {
		Site   string `default:"" env:"SITE" json:"TARGET_SITE" yaml:"site"`
		Prefix string `default:"TV" env:"PREFIX" json:"TARGET_PREFIX" yaml:"prefix"`
	} `json:"TARGET" yaml:"target"`
	Recorder struct {
		Bin      string `default:"youtube-dl" env:"RECORDER" json:"REC_DEFAULT_HOST" yaml:"bin"`
		Options  string `default:"" env:"REC_OPTIONS" json:"REC_OPTIONS" yaml:"options"`
		Minutes  int    `default:"70" env:"MINUTES" json:"REC_MINUTES" yaml:"minutes"`
		TempPath string `default:"/tmp/streams" env:"TEMP_PATH" json:"TEMP_PATH" yaml:"tempPath"`
	} `json:"RECORDER" yaml:"recorder"`
}

func NewApp(name, display string) *microservice.App {
	app := microservice.NewApp(&build, &secrets, &AppConfig{}, microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
		config := app.Config.(*AppConfig)
		evt = evt.Str("Stream Url", config.Target.Site).
			Str("FFMpeg     Bin", config.Ffmpeg.Bin).
			Str("FFMpeg Options", config.Ffmpeg.Options).
			Str("Output Path", config.Ffmpeg.OutputPath).
			Str("Output Prefix", config.Target.Prefix).
			Str("Rec     Bin", config.Recorder.Bin).
			Str("Rec Options", config.Recorder.Options).
			Int("TTL Mins Video", config.Recorder.Minutes).
			Str("Temp Path", config.Recorder.TempPath)
		for i, c := range build.ConfigFiles {
			evt = evt.Str(fmt.Sprintf("CfgFile %d", i+1), c)
		}
		evt.Msgf("streamdl version: %s", build.Version)
	}))
	app.PreRunApp(func(app *microservice.App) {
		config := app.Config.(*AppConfig)
		config.Target.Site = cleanUrl(config.Target.Site)
		config.Ffmpeg.Bin = common.StringDefault(&config.Ffmpeg.Bin, DEF_FFMPEG_BIN)
		if config.Ffmpeg.Bin != ""{
			config.Ffmpeg.Bin = storage.ConvertPathUNC(config.Ffmpeg.Bin)
		}
		if config.Ffmpeg.OutputPath != ""{
			config.Ffmpeg.OutputPath = storage.ConvertPathUNC(config.Ffmpeg.OutputPath)
		}
		config.Recorder.Bin = common.StringDefault(&config.Recorder.Bin, DEF_REC_BIN)
		if config.Recorder.Bin != ""{
			config.Recorder.Bin = storage.ConvertPathUNC(config.Recorder.Bin)
		}
		config.Recorder.TempPath = common.StringDefault(&config.Recorder.TempPath, DEF_TEMP_PATH)
		if config.Recorder.TempPath != ""{
			config.Recorder.TempPath = storage.ConvertPathUNC(config.Recorder.TempPath)
		}
		config.Ffmpeg.Options = common.StringDefault(&config.Ffmpeg.Options, DEF_FFMPEG_OPTION)
		config.Recorder.Bin = common.StringDefault(&config.Recorder.Bin, DEF_FFMPEG_BIN)
		config.Recorder.Options = common.StringDefault(&config.Recorder.Options, DEF_REC_OPTION)
	})
	return app
}

func cleanUrl(url string) string {
	url = strings.Trim(url, " /")
	lcurl := strings.ToLower(url)
	segs := strings.Split(url, "/")
	lseg := segs[len(segs)-1]
	if lseg != ""{
		if strings.HasPrefix(lcurl, "vimeo") {
			return "https://vimeo.com/event/" + lseg
		}
		if strings.HasPrefix(lcurl, "youtube") || len(segs) < 2 {
			return "https://www.youtube.com/watch?v=" + lseg
		}
	}else if lcurl == ""{
		return ""
	}
	if !strings.HasPrefix(lcurl, "http") {
		url = "https://" + url
	}
	return url
}

var (
	chkError   func(err error)
	checkError func(err error)
	version    = "0.1.0"
	gitHash    = "MISSING"
	buildStamp = "MISSING"
	branch     = ""
	sourceTag  = ""
	cfgFile    = "config.yml"
	commitMsg  = ""
	smtpPwd    = ""
	sqlPwd     = ""
	jwtSecret  = ""
	awsKey     = ""
	build      microservice.BuildInfo
	secrets    microservice.SecretInfo
)

func init() {
	build = *microservice.NewBuildInfo(version, gitHash, buildStamp, branch, sourceTag, cfgFile, commitMsg, appName, "ntc.org/netutils/streamdl")
	secrets = microservice.SecretInfo{sqlPwd, smtpPwd, jwtSecret, awsKey}
	chkError = microservice.CheckError(build.AppName)
	checkError = microservice.CheckError(build.AppName)
}
