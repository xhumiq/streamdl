package main

import (
	"ntc.org/mclib/auth"
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
	keys        *auth.RsaKeys
}

func NewService(app *microservice.App) *service {
	config := app.Config.(*AppConfig)
	client, err := authvault.NewVaultClient(config.Vault.VaultConfig)
	checkError(err)
	env := authvault.GetEnv(&config.Log, config.Vault.Environment)
	keys, err := client.CheckCurrentRole(env, app.Build.AppName, "elzion", client.Config().Token)
	checkError(err)
	return &service{
		App:         app,
		SvcConfig:   app.Config.(*AppConfig),
		lastResults: make(map[string]*webdavHealth),
		chProc:      make(chan (time.Time), 10),
		vault:       client,
		keys:        keys,
	}
}

func (svc service) Name() string {
	return appName
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
