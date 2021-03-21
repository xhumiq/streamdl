package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	authvault "bitbucket.org/xhumiq/go-mclib/auth/vault"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/nechi"
	"sync"
	"time"
)

func init() {
	nechi.ServicePort = 80
}

var (
	mlock         sync.Mutex
)

func NewWebDavListener(srv *service) *nechi.WebChi {
	app := nechi.NewWebApp(&srv.AppStatus, srv.AppConfig.Http, srv.keys)
	cacheCfg := srv.SvcConfig.Caching
	if  srv.SvcConfig.Monitor.AppMode != "MONITORONLY" {
		vc, err := nechi.NewBetterMemCache(1 << 33)
		checkError(err)
		rc, err := nechi.NewBetterMemCache(1 << 24)
		checkError(err)
		ws, err := nechi.NewWebDavService(srv.SvcConfig.Monitor.DAVPrefix, srv.AppConfig.Http,
			nechi.SetShortCacheClient(rc, time.Duration(cacheCfg.ShortTTLMins) * time.Minute),
			nechi.SetAuthUser(srv.VaultAuthUser),
			nechi.AddCachePatternFilter(nechi.CachePatternFilter{
				Method:     "GET",
				UrlPattern: nil,
				UrlPrefix:  "/Video/",
				Duration:   common.ToDurationPtr(2 * 24 * time.Hour),
			}, vc, time.Duration(cacheCfg.VideoTTLMins) * time.Minute),
			nechi.AddCachePatternFilter(nechi.CachePatternFilter{
				Method:     "GET",
				UrlPattern: nil,
				UrlPrefix:  "/Audio/",
				Duration:   common.ToDurationPtr(7 * 24 * time.Hour),
			}, rc, time.Duration(cacheCfg.RecentTTLMins) * time.Minute),
		)
		checkError(err)
		app.Use(ws.WebDavHandler)
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
	if srv.SvcConfig.Vault.Token == "" || userName == srv.SvcConfig.Users.HebronUser || userName == srv.SvcConfig.Users.UploadUser {
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
