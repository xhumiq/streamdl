package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	authvault "ntc.org/mclib/auth/vault"

	"github.com/rs/zerolog/log"

	"ntc.org/mclib/common"

	"github.com/pkg/errors"
	"ntc.org/mclib/microservice"
)

type service struct {
	*microservice.App
	SvcConfig   *AppConfig
	chProc      chan (time.Time)
	lastResults map[string]*webdavHealth
	lastCheck   time.Time
	vault       *authvault.VaultClient
}

func NewService(app *microservice.App) *service {
	config := app.Config.(*AppConfig)
	client, err := authvault.NewVaultClient(config.Vault.VaultConfig, &config.Log)
	checkError(err)
	return &service{
		App:         app,
		SvcConfig:   app.Config.(*AppConfig),
		lastResults: make(map[string]*webdavHealth),
		chProc:      make(chan (time.Time), 10),
		vault:       client,
	}
}

func (svc service) Name() string {
	return appName
}

func RegisterToken(config *AppConfig, env, domain, name string) (*authvault.CreatedToken, error){
	vc := config.Vault.VaultConfig
	vc.Token = vc.RegToken
	vc.AutoRenew = false
	if vc.Token == ""{
		return nil, errors.Errorf("Registor Token is not specified")
	}
	client, err := authvault.NewVaultClient(vc, &config.Log)
	if err!=nil{
		return nil, err
	}
	env = authvault.GetEnv(config.Log, env, config.Vault.Environment)
	if domain == ""{
		domain = "ziongjcc.org"
	}
	if name == ""{
		name = config.Vault.HostName
	}
	if name == ""{
		name = "webdav"
	}
	token, err := client.CreateServiceToken(env, domain, "elzion", name)
	if err != nil{
		log.Error().Str("Error",fmt.Sprintf("%+v", err)).Msgf("Create Service Token")
		return nil, err
	}
	if token.ClientToken == "" {
		return nil, errors.Errorf("Unable to create token")
	}
	log.Info().Msgf("Created Service Token Env: %s Name: %s Token: %s", env, name, common.MaskedSecret(token.ClientToken))
	vc.Token = token.ClientToken
	config.Vault.Token = token.ClientToken
	key, err := hex.DecodeString(config.Vault.CfgEncSecret)
	if err != nil {
		return nil, err
	}
	println("Save Vault Credentials")
	credentials := authvault.LocalCredentials{}
	credentials["_"] = authvault.LocalCredentialEntry{
		Token: token.ClientToken,
	}
	err = authvault.SaveVaultCredentials(key, config.Vault.ConfigPath, credentials)
	return token, err
}

func (s *service) Start(sd *common.ShutDownable) error {
	if s.SvcConfig.Monitor.AppMode == "WEBDAVONLY" {
		return nil
	}
	dur := s.SvcConfig.Monitor.DurMins
	if dur < 1 {
		dur = 10
	}
	sd.Go(func() error {
		ta := time.NewTimer(time.Duration(dur) * time.Minute)
		domains := common.FilterEmptyStrings(strings.Split(s.SvcConfig.Monitor.Domains, ",")...)
		for {
			lt := time.Now()
			if err := s.CheckHealth(domains...); err != nil {
				return err
			}
			pt := time.Now()
			log.Info().Msgf("Next Run At: %s", lt.Add(time.Duration(dur)*time.Minute).String())
			select {
			case <-sd.Dying():
				return nil
			case <-ta.C:
			case <-s.chProc:
				{
					if time.Now().Sub(pt).Milliseconds() < 100 {
						continue
					}
				}
			}
		}
		return nil
	})
	return nil
}

func (s *service) CheckHealth(domains ...string) (err error) {
	var res *webdavHealth
	for _, d := range domains {
		if d == "localhost" {
			d = "http://" + d
		}
		if !strings.HasPrefix(d, "http") {
			d = "https://" + d
		}
		res, err = CheckHealth(d, s.SvcConfig)
		if err != nil {
			return
		}
		println("MS", res.LatencyMiliSecs)
		s.lastResults[d] = res
	}
	return
}

func (s *service) Stop() error {
	return nil
}

var (
	ERR_NOT_RUNNING = errors.Errorf("Service is not running")
)
