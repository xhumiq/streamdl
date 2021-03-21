package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/urfave/cli/v2"
	"strings"
)

type AppConfig struct {
	Smtp     common.SmtpConfig
	Log      common.LogConfig
	Download struct {
		Domain     string `default:"file-jp.ziongjcc.org" env:"ZSF_DOMAIN" json:"ZSF_DOMAIN" yaml:"ZSF_DOMAIN"`
		VideoPath  string `default:"./video" env:"ZSF_VIDEO_PATH" json:"ZSF_VIDEO_PATH" yaml:"ZSF_VIDEO_PATH"`
		AudioPath  string `default:"./audio" env:"ZSF_AUDIO_PATH" json:"ZSF_AUDIO_PATH" yaml:"ZSF_AUDIO_PATH"`
		HebronUser string `default:"" env:"HEBRON_USER" json:"HEBRON_USER" yaml:"HEBRON_USER"`
		HebronPwd  string `default:"" env:"HEBRON_PASSWD" json:"-" yaml:"-"`
		VaultFile  string `default:"./davdats" env:"SECRET_VAULT" json:"SECRET_VAULT" yaml:"SECRET_VAULT"`
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
					Usage:       "--video ./video",
					Aliases:     []string{"w"},
					Destination: &c.Download.VideoPath,
				},
				&cli.StringFlag{
					Name:        "audio",
					Usage:       "--audio ./audio",
					Aliases:     []string{"a"},
					Destination: &c.Download.AudioPath,
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
					Name:        "secrets-path",
					Usage:       "--secrets-path ./davdats",
					Aliases:     []string{"s"},
					Destination: &c.Download.VaultFile,
				},
			}
		}))
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
