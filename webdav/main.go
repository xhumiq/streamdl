package main

import (
	"github.com/judwhite/go-svc/svc"
	"github.com/rs/zerolog"
	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
)

const (
	appName = "WebDav"
)

func main() {
	app := NewApp(appName, "WebDav Service")
	err := app.Run(
		//export HEBRON_PASSWD="*ChRisTKD~144^PeaCE=!"
		//export JACOB_PASSWD="7=FH_^8hsZpg3yM^@Q==Sm/SN<rkr.7/"
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config := app.Config.(*AppConfig)
			evt.Str("Hebron User", config.Users.HebronUser).
				Str("Hebron Pwd", common.MaskedSecret(config.Users.HebronPwd)).
				Str("Hebron Path", config.Users.HebronPath).
				Str("Upload Pwd", common.MaskedSecret(config.Users.UploadPwd)).
				Str("Upload User", config.Users.UploadUser).
				Str("Upload Path", config.Users.UploadPath).
				Msgf("WebDav: %s", build.Version)
		}),
		microservice.RegisterService(func(app *microservice.App) svc.Service {
			s := NewService(app)
			app.RegisterWebService(NewIdApi(s))
			return app
		}))
	checkError(err)
}
