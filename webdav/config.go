package main

import (
	"bitbucket.org/xhumiq/go-mclib/auth"
	"bitbucket.org/xhumiq/go-mclib/auth/cognito"
	authvault "bitbucket.org/xhumiq/go-mclib/auth/vault"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/nechi"
	"github.com/rs/zerolog/log"
	"fmt"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

type AppConfig struct {
	Service microservice.ServiceConfig
	Smtp    common.SmtpConfig
	Log     common.LogConfig
	Http    nechi.Config
	Cognito cognito.Config
	Vault   struct {
		authvault.VaultConfig
	}
	Monitor struct {
		DAVPrefix  string `default:"" env:"WEBDAV_PREFIX" json:"WEBDAV_PREFIX" yaml:"WEBDAV_PREFIX"`
		DurMins    int    `default:"10" env:"MON_DUR_MINS" json:"MON_DUR_MINS" yaml:"MON_DUR_MINS"`
		AppMode    string `default:"" env:"WEBDAV_MODE" json:"WEBDAV_MODE" yaml:"WEBDAV_MODE"`
		Domains    string `default:"file-jp.ziongjcc.org" env:"MON_DOMAINS" json:"MON_DOMAINS" yaml:"MON_DOMAINS"`
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
	credentials authvault.LocalCredentials
	registered  bool
}

func NewApp(name, display string) *microservice.App {
	app := microservice.NewApp(build, &secrets, &AppConfig{}, microservice.ConfigFlags(
		func(config interface{}) []cli.Flag {
			c := config.(*AppConfig)
			return []cli.Flag{
				&cli.StringFlag{
					Name:        "appmode",
					Usage:       "--appmode [WEBDAVONLY,MONITORONLY]",
					Aliases:     []string{"a"},
					Destination: &c.Monitor.AppMode,
				},
				&cli.StringFlag{
					Name:        "domains",
					Usage:       "--domains [jp,us]",
					Aliases:     []string{"d"},
					Destination: &c.Monitor.Domains,
				},
				&cli.StringFlag{
					Name:        "video",
					Usage:       "--video ./video",
					Aliases:     []string{"w"},
					Destination: &c.Monitor.VideoPath,
				},
				&cli.StringFlag{
					Name:        "audio",
					Usage:       "--audio ./audio",
					Aliases:     []string{"a"},
					Destination: &c.Monitor.AudioPaths,
				},
				&cli.StringFlag{
					Name:        "user",
					Usage:       "--user hebron",
					Aliases:     []string{"u"},
					Destination: &c.Users.HebronUser,
				},
				&cli.StringFlag{
					Name:        "password",
					Usage:       "--password 'rhema'",
					Aliases:     []string{"p"},
					Destination: &c.Users.HebronPwd,
				},
			}
		}))
	app.PreRunApp(func(app *microservice.App) {
		config := app.Config.(*AppConfig)
		if config.Http.Port < 25{
			config.Http.Port = 80
		}
		if config.Service.Name == "" {
			config.Service.Name = name
		}
		config.Service.DisplayName = display
		if config.Vault.Domain == "" {
			config.Vault.Domain = "ziongjcc.org"
		}
		if config.Vault.DefaultPolicy == "" {
			config.Vault.DefaultPolicy = "elzion"
		}
		if config.Vault.HostName == "" {
			config.Vault.HostName = config.Service.Name
		}
		doReg := build.Command!="init"
		resp, err := authvault.InitConfig(authvault.InitConfigParams{
			Env:              config.Log.Environment,
			Name:             config.Vault.HostName,
			Secret:           secrets.JwtSecret,
			Path:             filepath.Dir(build.ExeBinPath),
			Domain:           "ziongjcc.org",
			Policy:           "elzion",
			Config:           &config.Vault.VaultConfig,
			LogConfig:        &config.Log,
			RegisterIfNeeded: doReg,
			SqlConfigs:       []string{"elzion/hebron", "elzion/jacob"},
		})
		if err != nil {
			log.Error().Str("Error", fmt.Sprintf("%+v", err)).Msgf("Unable to Initialize Vault Parameters")
		}
		if resp != nil {
			config.credentials = resp.Credentials
			config.registered = resp.Registered
		}
		if config.credentials!=nil{
			if config.credentials["elzion/hebron"] != nil && config.credentials["elzion/hebron"].Credentials!=nil{
				bc, err := auth.HashPassword(config.credentials["elzion/hebron"].Credentials.Password)
				if err != nil{
					log.Error().Str("Error", fmt.Sprintf("%+v", err)).Msgf("Unable to hash password for hebron")
				}
				config.Users.HebronBCrypt = string(bc)
			}
			if config.credentials["elzion/jacob"] != nil && config.credentials["elzion/jacob"].Credentials!=nil{
				bc, err := auth.HashPassword(config.credentials["elzion/jacob"].Credentials.Password)
				if err != nil{
					log.Error().Str("Error", fmt.Sprintf("%+v", err)).Msgf("Unable to hash password for jacob")
				}
				config.Users.UploadBCrypt = string(bc)
			}
		}
		config.Http.Users = []*nechi.UserProfile{
			&nechi.UserProfile{
				Username: config.Users.HebronUser,
				Password: config.Users.HebronBCrypt,
				Scope:    config.Users.HebronPath,
				Groups:   []string{"hebron"},
				Modify:   false,
			},
			&nechi.UserProfile{
				Username: config.Users.UploadUser,
				Password: config.Users.UploadBCrypt,
				Scope:    config.Users.UploadPath,
				Groups:   []string{"jacob"},
				Modify:   true,
			},
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
