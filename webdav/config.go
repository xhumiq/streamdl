package main

import (
	"ntc.org/mclib/auth/cognito"
	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
	"ntc.org/mclib/nechi"
)

type AppConfig struct {
	Service microservice.ServiceConfig
	Smtp    common.SmtpConfig
	Log     common.LogConfig
	Http    nechi.Config
	Cognito cognito.Config
	Monitor struct {
		DAVPrefix  string `default:"" env:"WEBDAV_PREFIX" json:"WEBDAV_PREFIX" yaml:"WEBDAV_PREFIX"`
		DurMins    int    `default:"10" env:"MON_DUR_MINS" json:"MON_DUR_MINS" yaml:"MON_DUR_MINS"`
		AppMode    string `default:"" env:"WEBDAV_MODE" json:"WEBDAV_MODE" yaml:"WEBDAV_MODE"`
		Domains    string `default:"upload-nw.ziongjcc.org" env:"MON_DOMAINS" json:"MON_DOMAINS" yaml:"MON_DOMAINS"`
		VideoPath  string `default:"/Video" env:"MON_VIDEO_PATH" json:"MON_VIDEO_PATH" yaml:"MON_VIDEO_PATH"`
		AudioPaths string `default:"/Audio/mp3mono,/Audio/mp3stereo" env:"MON_AUDIO_PATHS" json:"MON_AUDIO_PATHS" yaml:"MON_AUDIO_PATHS"`
	}
	Users struct {
		HebronUser   string `default:"hebron" env:"HEBRON_USER" json:"HEBRON_USER" yaml:"HEBRON_USER"`
		HebronPwd    string `default:"rhema" env:"HEBRON_PASSWD" json:"-" yaml:"-"`
		HebronBCrypt string `default:"rhema" env:"HEBRON_BCRYPT" json:"-" yaml:"-"`
		HebronPath   string `default:"/srv/media/" env:"HEBRON_PATH" json:"HEBRON_PATH" yaml:"HEBRON_PATH"`
		UploadUser   string `default:"jacob" env:"JACOB_USER" json:"JACOB_USER" yaml:"JACOB_USER"`
		UploadPwd    string `default:"rhema" env:"JACOB_PASSWD" json:"-" yaml:"-"`
		UploadBCrypt string `default:"rhema" env:"JACOB_BCRYPT" json:"-" yaml:"-"`
		UploadPath   string `default:"/srv/upload" env:"JACOB_PATH" json:"JACOB_PATH" yaml:"JACOB_PATH"`
	}
}

func NewApp(name, display string) *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, &microservice.AppConfig{
		Http: &nechi.Config{
			Port: 80,
		},
	})
	if config.Service.Name == "" {
		config.Service.Name = name
	}
	config.Service.DisplayName = display
	config.Http.Users = []*nechi.UserProfile{
		&nechi.UserProfile{
			Username: config.Users.HebronUser,
			Password: config.Users.HebronBCrypt,
			Scope:    config.Users.HebronPath,
			Modify:   false,
		},
		&nechi.UserProfile{
			Username: config.Users.UploadUser,
			Password: config.Users.UploadBCrypt,
			Scope:    config.Users.UploadPath,
			Modify:   true,
		},
	}
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
