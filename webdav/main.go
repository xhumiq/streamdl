package main

import (
	"encoding/json"
	"fmt"
	"ntc.org/mclib/auth/cognito"
	"os"
	"sync"

	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
)

const (
	appName = "WebDav"
)

func main() {
	app := NewApp(appName, "WebDav Service")
	app.Cmd("login", func(c *cli.Context) error {
		config := app.Config.(*AppConfig)
		client, err := cognito.CreateProvider(config.Cognito)
		if err!=nil{
			return err
		}
		res, err := cognito.Login(client, config.Cognito, "", "")
		if err!=nil{
			return err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		println(string(b))
		return nil
	})
	app.Cmd("webdav", func(c *cli.Context) error {
		config := app.Config.(*AppConfig)
		wg := sync.WaitGroup{}
		for i:=0;i<30;i++{
			wg.Add(1)
			go func(){
				defer wg.Done()
				_, err := CheckHealth("https://file-us.ziongjcc.org", config)
				println(fmt.Sprintf("%+v", err))
			}()
		}
		wg.Wait()
		return nil
	})
	app.Cmd("pwd", func(c *cli.Context) error {
		var bpwd []byte
		var err error
		hpwd, ok := os.LookupEnv("HEBRON_PASSWD")
		if !ok {
			return fmt.Errorf("Vars HEBRON_PASSWD and JACOB_PASSWD must be set")
		}
		if bpwd, err = bcrypt.GenerateFromPassword([]byte(hpwd), 5); err != nil {
			return err
		}
		if len(bpwd) > 0 {
			println(string(bpwd))
		}
		if hpwd, ok = os.LookupEnv("JACOB_PASSWD"); !ok {
			return fmt.Errorf("Vars HEBRON_PASSWD and JACOB_PASSWD must be set")
		}
		if bpwd, err = bcrypt.GenerateFromPassword([]byte(hpwd), 5); err != nil {
			return err
		}
		if len(bpwd) > 0 {
			println(string(bpwd))
		}
		return nil
	})
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt = evt.Str("User Hebron User", config.Users.HebronUser).
				Str("User Hebron Path", config.Users.HebronPath).
				Str("User Upload User", config.Users.UploadUser).
				Str("User Upload Path", config.Users.UploadPath).
				Str("AC A RegionId", config.Cognito.RegionID).
				Str("AC B   PoolId", config.Cognito.UserPoolID).
				Str("AC C ClientId", config.Cognito.AppClientID).
				Str("AC D   Secret", common.MaskedSecret(config.Cognito.AppClientSecret))

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
		}),
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			s := NewService(app)
			app.RegisterWebService(NewIdApi(s))
			if s.SvcConfig.Monitor.AppMode != "WEBDAVONLY" {
				app.RegisterService(s)
			}
			return app
		}))
	checkError(err)
}
