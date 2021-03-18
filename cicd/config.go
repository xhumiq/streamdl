package main

import (
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/rs/zerolog/log"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/nechi"
)

type AppConfig struct {
	ConfigRootPath    string `default:"./config" env:"CONFIG_ROOT_PATH" json:"CONFIG_ROOT_PATH" yaml:"CONFIG_ROOT_PATH"`
	DefaultRepo struct{
		Repository string `env:"REPO_URL" json:"REPO_URL" yaml:"REPO_URL"`
		RepoName   string `env:"REPO_NAME" json:"REPO_NAME" yaml:"REPO_NAME"`
		Branch     string `default:"master" env:"REPO_BRANCH" json:"REPO_BRANCH" yaml:"REPO_BRANCH"`
		DeployPath string `env:"DEPLOY_PATH" json:"DEPLOY_PATH" yaml:"DEPLOY_PATH"`
		ConfigPath string `env:"CONFIG_PATH" json:"CONFIG_PATH" yaml:"CONFIG_PATH"`
		GitPath    string `env:"GIT_PATH" json:"GIT_PATH" yaml:"GIT_PATH"`
		CertPath   string `default:"~/.ssh/id_rsa" env:"GIT_SSH_CERT_PATH" json:"GIT_SSH_CERT_PATH" yaml:"GIT_SSH_CERT_PATH"`
		DeployRootPath    string `default:"./deploy" env:"DEPLOY_ROOT_PATH" json:"DEPLOY_ROOT_PATH" yaml:"DEPLOY_ROOT_PATH"`
		GitRootPath       string `default:"./repos" env:"GIT_ROOT_PATH" json:"GIT_ROOT_PATH" yaml:"GIT_ROOT_PATH"`
		EnvPrefix  string `env:"ENV_VAR_PREFIX" json:"ENV_VAR_PREFIX" yaml:"ENV_VAR_PREFIX"`
	} `json:"default" yaml:"default"`
	Repos map[string]struct {
		DeployPath string            `json:"deployPath" yaml:"deployPath"`
		GitPath    string            `json:"gitPath" yaml:"gitPath"`
		ConfigPath string            `json:"configPath" yaml:"configPath"`
		CertPath   string            `json:"certPath" yaml:"certPath"`
		EnvPrefix  string            `json:"envPrefix" yaml:"envPrefix"`
		Repository string            `json:"repository" yaml:"repository"`
		Commands   []string          `json:"commands" yaml:"commands"`
		Env        map[string]string `json:"env" yaml:"env"`
		Timeout    int               `json:"timeout" yaml:"timeout"`
	} `json:"repos" yaml:"repos"`
	Http     nechi.Config
	Service  microservice.ServiceConfig
	Smtp     common.SmtpConfig
	Log      common.LogConfig
}

func LogServiceInfo(config *AppConfig) {
	e := log.Info().
		Str("Repos Url", config.DefaultRepo.Repository)
	if config.DefaultRepo.RepoName != "" {
		e = e.Str("*Name", config.DefaultRepo.RepoName)
	}
	e = e.Str("Branch", config.DefaultRepo.Branch).
		Str("DeployPath", config.DefaultRepo.DeployPath)
	e.Msgf("EnvPrefix", config.DefaultRepo.EnvPrefix)
}

func NewApp(name, display string) *microservice.App {
	config := AppConfig{}
	app := microservice.NewApp(build, &secrets, &config, nil)
	config.Service.Name = name
	config.Service.DisplayName = display
	return app
}

var (
	checkError func(err error)
	version    = "0.1.0"
	gitHash    = "MISSING"
	buildStamp = "MISSING"
	branch     = ""
	sourceTag  = ""
	cfgFile    = "config.yml"
	commitMsg  = ""
	sqlPwd     = ""
	smtpPwd    = ""
	jwtSecret  = ""
	awsSecret  = ""
	build      microservice.BuildInfo
	secrets    microservice.SecretInfo
)

func init() {
	build = *microservice.NewBuildInfo(version, gitHash, buildStamp, branch, sourceTag, cfgFile, commitMsg, appName)
	secrets = microservice.SecretInfo{sqlPwd, smtpPwd, jwtSecret, awsSecret}
	checkError = microservice.CheckError(build.AppName)
}
