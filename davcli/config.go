package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"bitbucket.org/xhumiq/go-mclib/storage"
	"github.com/rs/zerolog"
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
	app := microservice.NewApp(&build, &secrets, &config, microservice.ConfigFlags(
		func(config interface{}) []cli.Flag {
			c := config.(*AppConfig)
			cf := microservice.CreateFlagOption
			return microservice.CreateFlagsCheckErr(checkError,
				cf("[-d,--domains <jp,us>]", &c.Download.Domain),
				cf("[-w,--video $USERPROFILE/Videos/zsf]", &c.Download.VideoPath),
				cf("[-a,--audio $USERPROFILE/Music/zsf]", &c.Download.AudioPath),
				cf("[--hymns $USERPROFILE/Music]", &c.Download.HymnsPath),
				cf("[--litcenter $USERPROFILE/Documents]", &c.Download.DocsPath),
				cf("[--photos $USERPROFILE/Pictures]", &c.Download.PhotosPath),
				cf("[--school $USERPROFILE/Documents]", &c.Download.SchoolPath),
				cf("[-b,--base-path $USERPROFILE]", &c.Download.BaseTargetPath),
				cf("[--secrets-path ./davdats", &c.Download.VaultFile),
				cf("[-y,--history-days 2", &c.Download.HistoryDays),
				cf("[-t,--download-threads 6]", &c.Download.DownloadThreads),
				cf("[-f,--force-overwrite]", &c.Download.ForceOverwrite),
			)
		}), microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
		config := app.Config.(*AppConfig)
		evt.Str("WebDav UserName", config.Download.HebronUser).
			Str("Sync    Base Path", storage.ConvertPathUNC(config.Download.BaseTargetPath)).
			Str("Sync   Video Path", storage.ConvertPathUNC(config.Download.VideoPath)).
			Str("Sync   Audio Path", storage.ConvertPathUNC(config.Download.AudioPath)).
			Str("Sync   Hymns Path", storage.ConvertPathUNC(config.Download.HymnsPath)).
			Str("Sync  Photos Path", storage.ConvertPathUNC(config.Download.PhotosPath)).
			Str("Sync  School Path", storage.ConvertPathUNC(config.Download.SchoolPath)).
			Str("Sync     Lit Path", storage.ConvertPathUNC(config.Download.DocsPath)).
			Int("Sync History Days", config.Download.HistoryDays).
			Int("Sync ---- Threads", config.Download.DownloadThreads).
			Bool("Sync ForceReplace", config.Download.ForceOverwrite).
			Str("WebDav   Domain", config.Download.Domain).
			Str("WebDav Password", common.MaskedSecret(config.Download.HebronPwd)).
			Str("Vault Path", storage.ConvertPathUNC(config.Download.VaultFile)).
			Msgf("DavCli: %s", build.Version)
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
		config.Download.Domain = domain
		config.Download.BaseTargetPath = common.StringDefault(&config.Download.BaseTargetPath, "$USERPROFILE/Videos/zsf")
		config.Download.BaseTargetPath = storage.ConvertUNCPath(config.Download.BaseTargetPath)
		config.Download.VideoPath = storage.ConvertToAbsPath(config.Download.VideoPath,config.Download.BaseTargetPath)
  	config.Download.AudioPath = storage.ConvertToAbsPath(config.Download.AudioPath,config.Download.BaseTargetPath)
		config.Download.HymnsPath = storage.ConvertToAbsPath(config.Download.HymnsPath,config.Download.BaseTargetPath)
		config.Download.PhotosPath = storage.ConvertToAbsPath(config.Download.PhotosPath,config.Download.BaseTargetPath)
		config.Download.DocsPath = storage.ConvertToAbsPath(config.Download.DocsPath,config.Download.BaseTargetPath)
		config.Download.SchoolPath = storage.ConvertToAbsPath(config.Download.SchoolPath,config.Download.BaseTargetPath)
		dir := filepath.Dir(build.ExeBinPath)
		config.Download.VaultFile = storage.ConvertToAbsPath(storage.ConvertUNCPath(config.Download.VaultFile),dir)
		if config.Download.HistoryDays < 0{
			config.Download.HistoryDays = -config.Download.HistoryDays
		}
		config.Download.HistoryDays = common.IntDefault(&config.Download.HistoryDays, 2)
		config.Download.HebronUser = common.StringDefault(&config.Download.HebronUser, sqlPwd)
		config.Download.HebronPwd = common.StringDefault(&config.Download.HebronPwd, awsKey)
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
	build = *microservice.NewBuildInfo(version, gitHash, buildStamp, branch, sourceTag, cfgFile, commitMsg, appName, "ntc.org/netutils/davcli")
	secrets = microservice.SecretInfo{sqlPwd, smtpPwd, jwtSecret, awsKey}
	chkError = microservice.CheckError(build.AppName)
	checkError = microservice.CheckError(build.AppName)
}
