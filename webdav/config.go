package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"path/filepath"

	"bitbucket.org/xhumiq/go-mclib/auth"
	"bitbucket.org/xhumiq/go-mclib/auth/cognito"
	authvault "bitbucket.org/xhumiq/go-mclib/auth/vault"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/nechi"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type AppConfig struct {
	Service microservice.ServiceConfig
	Smtp    common.SmtpConfig
	Log     common.LogConfig
	Http    nechi.Config
	Caching struct {
		VideoFilterDays  int    `default:"2" env:"VIDEO_FILTER_DAYS" json:"VIDEO_FILTER_DAYS" yaml:"VIDEO_FILTER_DAYS"`
		RecentFilterDays int    `default:"7" env:"RECENT_FILTER_DAYS" json:"RECENT_FILTER_DAYS" yaml:"RECENT_FILTER_DAYS"`
		VideoTTLMins     int    `default:"30" env:"VIDEO_TTL_MINS" json:"VIDEO_TTL_MINS" yaml:"VIDEO_TTL_MINS"`
		RecentTTLMins    int    `default:"30" env:"RECENT_TTL_MINS" json:"RECENT_TTL_MINS" yaml:"RECENT_TTL_MINS"`
		ShortTTLMins     int    `default:"5" env:"SHORT_TTL_SECS" json:"SHORT_TTL_SECS" yaml:"SHORT_TTL_SECS"`
		VideoMaxBytes    string `default:"1>>33" env:"VIDEO_MAX_BYTES" json:"VIDEO_MAX_BYTES" yaml:"VIDEO_MAX_BYTES"`
		RecentMaxBytes   string `default:"1>>23" env:"RECENT_MAX_BYTES" json:"RECENT_MAX_BYTES" yaml:"RECENT_MAX_BYTES"`
	}
	Cognito cognito.Config
	Vault   authvault.VaultConfig
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
			cf := microservice.CreateFlagOption
			return microservice.CreateFlagsCheckErr(checkError,
				cf("[-m,--appmode <WEBDAVONLY,MONITORONLY>]", &c.Monitor.AppMode),
				cf("[-d,--domains <jp,us>]", &c.Monitor.Domains),
				cf("[-w,--video <./Video>]", &c.Monitor.VideoPath),
				cf("[-a,--audio <./Audio>]", &c.Monitor.AudioPaths),
				cf("[-u,--user <hebron>]", &c.Users.HebronUser),
				cf("[-s,--password <rhema>]", &c.Users.HebronPwd),
			)
		}),
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt = evt.Str("User Hebron Usr", config.Users.HebronUser).
				Str("User Hebron Pth", config.Users.HebronPath).
				Str("User Upload Usr", config.Users.UploadUser).
				Str("User Upload Pth", config.Users.UploadPath).
				Str("MaxCost Recent", config.Caching.RecentMaxBytes).
				Str("MaxCost  Video", config.Caching.VideoMaxBytes).
				Int("TTL Mins  Short", config.Caching.ShortTTLMins).
				Int("TTL Mins Recent", config.Caching.RecentTTLMins).
				Int("TTL Mins  Video", config.Caching.VideoTTLMins).
				Int("CacheVideo  Dys", config.Caching.VideoFilterDays).
				Int("CacheRecent Dys", config.Caching.RecentFilterDays)

			if config.Monitor.AppMode != "WEBDAVONLY" {
				evt = evt.Int("Mon    Dur Mins", config.Monitor.DurMins).
					Str("Mon     Domains", config.Monitor.Domains).
					Str("Mon  Video Path", config.Monitor.VideoPath).
					Str("Mon  Audio Path", config.Monitor.AudioPaths)
			}
			if config.Monitor.AppMode != "MONITORONLY" {
				evt = evt.Str("DAV Prefix", config.Monitor.DAVPrefix)
			}
			if config.Users.HebronPwd != "" {
				evt = evt.Str("User Hebron Pwd", common.MaskedSecret(config.Users.HebronPwd)).
					Str("User Upload Pwd", common.MaskedSecret(config.Users.UploadPwd))
			}
			if config.Users.HebronBCrypt != "" {
				evt = evt.Str("User Hebron Hsh", common.MaskedSecret(config.Users.HebronBCrypt)).
					Str("User Upload Hsh", common.MaskedSecret(config.Users.UploadBCrypt))
			}
			mode := config.Monitor.AppMode
			if mode == "" {
				mode = "WebDav and Monitor"
			}
			evt.Msgf("WebDav: %s Mode: %s", build.Version, mode)
		}))
	app.PreRunApp(func(app *microservice.App) {
		config := app.Config.(*AppConfig)
		if config.Http.Port < 25 {
			config.Http.Port = 80
		}
		config.Service.Name = common.StringDefault(&config.Service.Name, name)
		config.Service.DisplayName = display
		config.Vault.Domain = common.StringDefault(&config.Vault.Domain, "ziongjcc.org")
		config.Vault.DefaultPolicy = common.StringDefault(&config.Vault.DefaultPolicy, "elzion")
		config.Vault.HostName = common.StringDefault(&config.Vault.HostName, config.Service.Name)
		doReg := build.Command != "init"
		resp, err := authvault.InitConfig(authvault.InitConfigParams{
			Env:              config.Log.Environment,
			Name:             config.Vault.HostName,
			Secret:           secrets.JwtSecret,
			Path:             filepath.Dir(build.ExeBinPath),
			Domain:           "ziongjcc.org",
			Policy:           "elzion",
			Config:           &config.Vault,
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
		if config.credentials != nil {
			if config.credentials["elzion/hebron"] != nil && config.credentials["elzion/hebron"].Credentials != nil {
				bc, err := auth.HashPassword(config.credentials["elzion/hebron"].Credentials.Password)
				if err != nil {
					log.Error().Str("Error", fmt.Sprintf("%+v", err)).Msgf("Unable to hash password for hebron")
				}
				config.Users.HebronBCrypt = string(bc)
			}
			if config.credentials["elzion/jacob"] != nil && config.credentials["elzion/jacob"].Credentials != nil {
				bc, err := auth.HashPassword(config.credentials["elzion/jacob"].Credentials.Password)
				if err != nil {
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
