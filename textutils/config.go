package main

import (
	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
	"ntc.org/mclib/netutils/bitbucket"
	"ntc.org/mclib/netutils/sshutils"
)

type AppConfig struct {
	Log       common.LogConfig
	Hosts     sshutils.HostConfig
	Bitbucket bitbucket.BitbucketConfig
}

func NewApp() *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, nil)
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
