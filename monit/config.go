package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/nechi"
)

type AppConfig struct {
	Service microservice.ServiceConfig
	Smtp    common.SmtpConfig
	Log     common.LogConfig
	Http    nechi.Config
	Monitor struct {
		Host       string `default:"localhost" env:"MONIT_DEFAULT_HOST" json:"MONIT_DEFAULT_HOST" yaml:"MONIT_DEFAULT_HOST"`
		ConfigFile string `default:"monit.yml" env:"MONIT_CONFIG" json:"MONIT_CONFIG" yaml:"MONIT_CONFIG"`
	}
}

func NewApp(name, display string) *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, nil)
	if config.Service.Name == "" {
		config.Service.Name = name
	}
	config.Service.DisplayName = display
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
