package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/netutils/bitbucket"
	"bitbucket.org/xhumiq/go-mclib/netutils/linode"
	"bitbucket.org/xhumiq/go-mclib/netutils/sshutils"
	"bitbucket.org/xhumiq/go-mclib/storage"
)

type AppConfig struct {
	Log       common.LogConfig
	Hosts     sshutils.HostConfig
	Bitbucket bitbucket.BitbucketConfig
	Aws       common.AwsConfig
	Linode    linode.LinodeConfig
}

func NewApp() *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, nil)
	if config.Hosts.SshPrivateKey!=""{
		config.Hosts.SshPrivateKey = storage.ConvertUNCPath(config.Hosts.SshPrivateKey)
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
