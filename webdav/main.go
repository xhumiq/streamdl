package main

import (
	"fmt"
	"sync"
	"time"

	authvault "bitbucket.org/xhumiq/go-mclib/auth/vault"

	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/judwhite/go-svc/svc"
	"github.com/urfave/cli/v2"
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
		if name == "" {
			name = config.Vault.HostName
		}
		if name == "" {
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
