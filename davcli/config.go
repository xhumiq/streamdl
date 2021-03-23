package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/storage"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"strings"
)

type AppConfig struct {
	Smtp     common.SmtpConfig
	Log      common.LogConfig
	HttpClient common.HttpClientConfig
	Download struct {
		Domain     string `default:"file-jp.ziongjcc.org" env:"ZSF_DOMAIN" json:"ZSF_DOMAIN" yaml:"ZSF_DOMAIN"`
		VideoPath  string `default:"$USERPROFILE/Videos/ZSF" env:"ZSF_VIDEO_PATH" json:"ZSF_VIDEO_PATH" yaml:"ZSF_VIDEO_PATH"`
		AudioPath  string `default:"$USERPROFILE/Music/ZSF" env:"ZSF_AUDIO_PATH" json:"ZSF_AUDIO_PATH" yaml:"ZSF_AUDIO_PATH"`
		HymnsPath  string `default:"$USERPROFILE/Music/Hymns" env:"ZSF_HYMNS_PATH" json:"ZSF_HYMNS_PATH" yaml:"ZSF_HYMNS_PATH"`
		DocsPath  string `default:"$USERPROFILE/Documents/LitCenter" env:"ZSF_DOCS_PATH" json:"ZSF_DOCS_PATH" yaml:"ZSF_DOCS_PATH"`
		PhotosPath  string `default:"$USERPROFILE/Pictures" env:"ZSF_PHOTOS_PATH" json:"ZSF_PHOTOS_PATH" yaml:"ZSF_PHOTOS_PATH"`
		SchoolPath  string `default:"$USERPROFILE/Documents/School" env:"ZSF_SCHOOL_PATH" json:"ZSF_SCHOOL_PATH" yaml:"ZSF_SCHOOL_PATH"`
		HebronUser string `default:"" env:"HEBRON_USER" json:"HEBRON_USER" yaml:"HEBRON_USER"`
		HebronPwd  string `default:"" env:"HEBRON_PASSWD" json:"-" yaml:"-"`
		VaultFile  string `default:"./davdats" env:"SECRET_VAULT" json:"SECRET_VAULT" yaml:"SECRET_VAULT"`
		BaseTargetPath string `default:"$USERPROFILE" env:"TARGET_PATH" json:"TARGET_PATH" yaml:"TARGET_PATH"`
		HistoryDays int  `default:"2" env:"HISTORY_DAYS" json:"HISTORY_DAYS" yaml:"HISTORY_DAYS"`
		DownloadThreads int  `default:"6" env:"DOWNLOAD_THREADS" json:"DOWNLOAD_THREADS" yaml:"DOWNLOAD_THREADS"`
		ForceOverwrite bool   `default:"false" env:"FORCE_OVERWRITE" json:"FORCE_OVERWRITE" yaml:"FORCE_OVERWRITE"`
		SyncSize  bool   `default:"true" env:"SYNC_SIZE" json:"SYNC_SIZE" yaml:"SYNC_SIZE"`
	}
}

func NewApp(name, display string) *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, microservice.ConfigFlags(
		func(config interface{}) []cli.Flag {
			c := config.(*AppConfig)
			return []cli.Flag{
				&cli.StringFlag{
					Name:        "domain",
					Usage:       "--domain [jp,us]",
					Aliases:     []string{"d"},
					Destination: &c.Download.Domain,
				},
				&cli.StringFlag{
					Name:        "video",
					Usage:       "--video $USERPROFILE/Videos",
					Aliases:     []string{"w"},
					Destination: &c.Download.VideoPath,
				},
				&cli.StringFlag{
					Name:        "audio",
					Usage:       "--audio $USERPROFILE/Music",
					Aliases:     []string{"a"},
					Destination: &c.Download.AudioPath,
				},
				&cli.StringFlag{
					Name:        "hymns",
					Usage:       "--hymns $USERPROFILE/Music",
					Destination: &c.Download.HymnsPath,
				},
				&cli.StringFlag{
					Name:        "litcenter",
					Usage:       "--litcenter $USERPROFILE/Documents",
					Destination: &c.Download.DocsPath,
				},
				&cli.StringFlag{
					Name:        "photos",
					Usage:       "--photos $USERPROFILE/Pictures",
					Destination: &c.Download.PhotosPath,
				},
				&cli.StringFlag{
					Name:        "school",
					Usage:       "--school $USERPROFILE/Documents",
					Destination: &c.Download.SchoolPath,
				},
				&cli.StringFlag{
					Name:        "user",
					Usage:       "--user hebron",
					Aliases:     []string{"u"},
					Destination: &c.Download.HebronUser,
				},
				&cli.StringFlag{
					Name:        "password",
					Usage:       "--password 'rhema'",
					Aliases:     []string{"p"},
					Destination: &c.Download.HebronPwd,
				},
				&cli.StringFlag{
					Name:        "base-path",
					Usage:       "--base-path $USERPROFILE/Video",
					Aliases:     []string{"b"},
					Destination: &c.Download.BaseTargetPath,
				},
				&cli.StringFlag{
					Name:        "secrets-path",
					Usage:       "--secrets-path ./davdats",
					Aliases:     []string{"s"},
					Destination: &c.Download.VaultFile,
				},
				&cli.IntFlag{
					Name:        "history-days",
					Usage:       "--history-days 2",
					Aliases:     []string{"y"},
					Destination: &c.Download.HistoryDays,
				},
				&cli.IntFlag{
					Name:        "download-threads",
					Usage:       "--download-threads 6",
					Aliases:     []string{"t"},
					Destination: &c.Download.DownloadThreads,
				},
				&cli.BoolFlag{
					Name:        "force-overwrite",
					Usage:       "--force-overwrite",
					Aliases:     []string{"f"},
					Destination: &c.Download.ForceOverwrite,
				},
			}
		}))
	app.Cli.Description = display
	cfg := app.Config.(*AppConfig)
	if cfg.HttpClient.MaxRetryCount < 4{
		cfg.HttpClient.MaxRetryCount = 200
	}
	app.PreRunApp(func(app *microservice.App) {
		config := app.Config.(*AppConfig)
		domain := config.Download.Domain
		if strings.HasPrefix(domain, "us") || strings.HasPrefix(domain, "jp") {
			domain = "file-" + domain
		}
		if (strings.HasPrefix(domain, "file-us") || strings.HasPrefix(domain, "file-jp")) && !strings.HasSuffix(domain, ".ziongjcc.org") {
			domain = domain + ".ziongjcc.org"
		}
		if !strings.HasPrefix(domain, "http") {
			domain = "https://" + domain
		}
		if config.Download.BaseTargetPath == ""{
			config.Download.BaseTargetPath = "$USERPROFILE/Videos/zsf"
		}
		config.Download.BaseTargetPath = storage.ConvertUNCPath(config.Download.BaseTargetPath)
		if config.Download.VideoPath == ""{
			config.Download.VideoPath = config.Download.BaseTargetPath
		}else{
			config.Download.VideoPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.VideoPath),config.Download.BaseTargetPath)
		}
		if config.Download.AudioPath == ""{
			config.Download.AudioPath = config.Download.BaseTargetPath
		}else{
			config.Download.AudioPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.AudioPath),config.Download.BaseTargetPath)
		}
		if config.Download.HymnsPath == ""{
			config.Download.HymnsPath = config.Download.BaseTargetPath
		}else{
			config.Download.HymnsPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.HymnsPath),config.Download.BaseTargetPath)
		}
		if config.Download.PhotosPath == ""{
			config.Download.PhotosPath = config.Download.BaseTargetPath
		}else{
			config.Download.PhotosPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.PhotosPath),config.Download.BaseTargetPath)
		}
		if config.Download.DocsPath == ""{
			config.Download.DocsPath = config.Download.BaseTargetPath
		}else{
			config.Download.DocsPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.DocsPath),config.Download.BaseTargetPath)
		}
		if config.Download.SchoolPath == ""{
			config.Download.SchoolPath = config.Download.BaseTargetPath
		}else{
			config.Download.SchoolPath = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.SchoolPath),config.Download.BaseTargetPath)
		}

		dir := filepath.Dir(build.ExeBinPath)
		config.Download.VaultFile = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.VaultFile),dir)

		if config.Download.HistoryDays == 0{
			config.Download.HistoryDays = 2
		}
		if config.Download.HistoryDays < 0{
			config.Download.HistoryDays = -config.Download.HistoryDays
		}
		config.Download.Domain = domain
		if config.Download.HebronUser == "" {
			config.Download.HebronUser = sqlPwd
		}
		if config.Download.HebronPwd == "" {
			config.Download.HebronPwd = awsKey
		}
	})
	return app
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
	build = *microservice.NewBuildInfo(version, gitHash, buildStamp, branch, sourceTag, cfgFile, commitMsg, appName)
	secrets = microservice.SecretInfo{sqlPwd, smtpPwd, jwtSecret, awsKey}
	chkError = microservice.CheckError(build.AppName)
	checkError = microservice.CheckError(build.AppName)
}
