package main

import (
	"fmt"
	"sync"
	"time"

	authvault "bitbucket.org/xhumiq/go-mclib/auth/vault"

	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
)

const (
	appName = "WebDav"
)

func main() {
	app := NewApp(appName, "WebDav Service")
	env, domain, policy, token, name := "", "", "", "", ""
	app.Cmd("init [-e,--env <env>] [-d,--domain <domain>] [-p,--policy <policy>] [-t,--token <token>] <name>", func(c *cli.Context) error {
		config := app.Config.(*AppConfig)
		if config.registered {
			return nil
		}
		if name == ""{
			name = config.Vault.HostName
		}
		if name == ""{
			name = config.Service.Name
		}
		env = authvault.GetEnv(&config.Log, env, config.Vault.Environment)
		_, err := authvault.RegisterToken(authvault.InitConfigParams{
			Env:              authvault.GetEnv(&config.Log, env, config.Vault.Environment),
			Name:             name,
			Domain:           domain,
			Policy:           "elzion",
			Secret:           secrets.JwtSecret,
			Config:           &config.Vault,
			SqlConfigs:       []string{"elzion/hebron", "elzion/jacob"},
			LogConfig:        &config.Log,
			RegisterIfNeeded: true,
		})
		return err
	}, &env, &domain, &policy, &token, &name)
	app.Cmd("login", func(c *cli.Context) error {
		svc := NewService(app)
		env := svc.SvcConfig.Log.Environment
		env = authvault.GetEnv(&svc.SvcConfig.Log, env, svc.SvcConfig.Vault.Environment)
		t1 := time.Now()
		//UNNSA~O.980
		for i := 0; i < 10; i++ {
			res, err := svc.vault.UserPassLogin(env, "ziongjcc.org", "UNNSA~O.980", "*ChRisTKD~144^PeaCE=!")
			if err != nil {
				return err
			}
			if res == nil {
				continue
			}
			md, _ := res.Auth.MetaData.(*authvault.AuthIdentity)
			if md != nil {
				println(md.Scope)
			}
			break
		}
		println(time.Now().Sub(t1).String())
		return nil
	})
	app.Cmd("webdav", func(c *cli.Context) error {
		config := app.Config.(*AppConfig)
		wg := sync.WaitGroup{}
		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := CheckHealth("https://file-us.ziongjcc.org", config)
				println(fmt.Sprintf("%+v", err))
			}()
		}
		wg.Wait()
		return nil
	})
	err := app.Run(
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
			if config.Vault.Address != "" {
				evt = evt.Str("Vault  Environ", config.Log.Environment).
					Str("Vault  Address", config.Vault.Address).
					Str("Vault HostName", config.Vault.HostName).
					Str("Vault   Domain", config.Vault.Domain).
					Str("Vault   Secret", common.MaskedSecret(config.Vault.CfgEncSecret)).
					Str("Vault  CfgPath", config.Vault.ConfigPath)
				if config.Vault.Token != "" {
					evt = evt.Str("Vault    Token", common.MaskedSecret(config.Vault.Token))
				}
				if config.Vault.RegToken != "" {
					evt = evt.Str("Vault RegToken", common.MaskedSecret(config.Vault.RegToken))
				}
			}
			evt.Msgf("WebDav: %s Mode: %s", build.Version, mode)
		}),
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			s := NewService(app)
			app.RegisterWebService(NewWebDavListener(s))
			if s.SvcConfig.Monitor.AppMode != "WEBDAVONLY" {
				app.RegisterService(s)
			}
			return app
		}))
	checkError(err)
}
