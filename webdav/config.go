package main

import (
	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
	"ntc.org/mclib/nechi"
)

type AppConfig struct {
	Service microservice.ServiceConfig
	Smtp    common.SmtpConfig
	Log     common.LogConfig
	Http    nechi.Config
	Users   struct{
		HebronUser string `default:"hebron" env:"HEBRON_USER" json:"HEBRON_USER" yaml:"HEBRON_USER"`
		HebronPwd string `default:"rhema" env:"HEBRON_BCRYPT" json:"-" yaml:"-"`
		HebronPath string `default:"/srv/media/" env:"HEBRON_PATH" json:"HEBRON_PATH" yaml:"HEBRON_PATH"`
		UploadUser string `default:"jacob" env:"JACOB_USER" json:"JACOB_USER" yaml:"JACOB_USER"`
		UploadPwd string `default:"rhema" env:"JACOB_BCRYPT" json:"-" yaml:"-"`
		UploadPath string `default:"/srv/upload" env:"JACOB_PATH" json:"JACOB_PATH" yaml:"JACOB_PATH"`
	}
}

func NewApp(name, display string) *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, &microservice.AppConfig{
		Http:           &nechi.Config{
			Port:             80,
		},
	})
	if config.Service.Name == "" {
		config.Service.Name = name
	}
	config.Service.DisplayName = display
	config.Http.Users = []*nechi.UserProfile{
		&nechi.UserProfile{
			Username: config.Users.HebronUser,
			Password: config.Users.HebronPwd,
			Scope:    config.Users.HebronPath,
			Modify:   false,
		},
		&nechi.UserProfile{
			Username: config.Users.UploadUser,
			Password: config.Users.UploadPwd,
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
