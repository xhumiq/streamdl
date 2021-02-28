package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	authvault "ntc.org/mclib/auth/vault"
	"ntc.org/mclib/common"
	"ntc.org/mclib/nechi"
	"sync"
)

func init() {
	nechi.ServicePort = 80
}

var (
	mlock         sync.Mutex
)

func NewWebDavListener(srv *service) *nechi.WebChi {
	sconfig := srv.AppConfig.Http
	app := nechi.NewWebApp(&srv.AppStatus, sconfig)
	if  srv.SvcConfig.Monitor.AppMode != "MONITORONLY" {
		ws, err := app.AddWebDav(srv.SvcConfig.Monitor.DAVPrefix, sconfig)
		if err!=nil{
			panic(err)
		}
		ws.AuthUser(srv.VaultAuthUser)
	}
	app.ApiHealth("/healthcheck", HealthCheck)
	return app
}

func (srv *service) VaultAuthUser(dav *nechi.WebDavService, r *http.Request) (*nechi.UserProfile, error){
	userName, password , ok := r.BasicAuth()
	if !ok{
		return nil, nil
	}
	//  || userName == srv.SvcConfig.Users.UploadUser
	if srv.SvcConfig.Vault.Token == "" || userName == srv.SvcConfig.Users.HebronUser {
		return nechi.DefaultAuthenticateUser(dav, r)
	}
	env := srv.SvcConfig.Vault.Environment
	domain := srv.SvcConfig.Vault.Domain
	res, err := srv.vault.UserPassLogin(env, domain, userName, password)
	if err != nil{
		log.Error().Str("Error", fmt.Sprintf("%+v", err)).Msgf("Error on user Login: %s", userName)
		return nil, common.NewErrorInfo(common.HttpStatusCode(401))
	}
	if res == nil{
		log.Error().Msgf("User not found: %s", userName)
		return nil, common.NewErrorInfo(common.HttpStatusCode(401))
	}
	md, _ := res.Auth.MetaData.(*authvault.AuthIdentity)
	if md==nil{
		log.Error().Msgf("System Error: Auth Identity not found for user %s", userName)
		return nil, common.NewErrorInfo(common.HttpStatusCode(401))
	}
	mlock.Lock()
	defer mlock.Unlock()
	user, _ := dav.Users[md.Scope]
	if user == nil{
		log.Error().Msgf("Scope not found: %s", md.Scope)
		return nil, common.NewErrorInfo(common.HttpStatusCode(401))
	}
	log.Info().Msgf("User logged in %s as %s", userName, md.Scope)
	return user, nil
}

func HealthCheck() ([]string, []error) {
	logs := []string{}
	errs := []error{}
	return logs, errs
}
