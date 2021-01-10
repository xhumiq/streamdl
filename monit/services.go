package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"nex.com/nelib/common"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"nex.com/nelib/microservice"
)

type service struct {
	*microservice.App
	SvcConfig *AppConfig
	name      string
	Apis      []*apiHealthCheck
	lastCheck time.Time
}

func NewService(app *microservice.App) *service {
	services, err := ReadConfigs(app, "")
	checkError(err)
	cc, err := getCheckClients("localhost", services)
	checkError(err)
	return &service{
		App:    app,
		Apis:   cc,
		SvcConfig: app.Config.(*AppConfig),
	}
}

func ReadConfigs(app *microservice.App, file string) ([]*apiHealthCheckInfo, error) {
	config := app.Config.(*AppConfig)
	if file == ""{
		file = config.Monitor.ConfigFile
	}
	if !strings.Contains(file, "/") || !strings.Contains(file, "\\"){
		bn := filepath.Dir(app.Build.ExeBinPath)
		file = filepath.Join(bn, file)
	}
	log.Info().Msgf("Reading Config File:", file)
	dat, err := ioutil.ReadFile(file)
	if err!=nil{
		return nil, err
	}
	checks := []*apiHealthCheckInfo{}
	err = yaml.Unmarshal(dat, &checks)
	return checks, err
}

func (svc *service) WriteConfigs(file string) error {
	services := []*apiHealthCheckInfo{}
	for _, c := range svc.Apis{
		services = append(services, &c.info)
	}
	out, err := yaml.Marshal(services)
	if err!=nil{
		return err
	}
	if file==""{
		file = svc.SvcConfig.Monitor.ConfigFile
	}
	log.Info().Msgf("Created Monitor Config: %s", file)
	err = ioutil.WriteFile(svc.SvcConfig.Monitor.ConfigFile, out, 0644)
	return err
}

func (svc *service) Name() string {
	return appName
}

func (svc *service) Start(sd *common.ShutDownable) error {
	log.Info().
		Msgf("Started Service Monitor")
	sd.Go(func() error {
		var tick *time.Ticker
		for !sd.IsDying() {
			svc.checkService()
			if tick == nil {
				tick = time.NewTicker(10 * time.Second)
			}
			select {
			case <-tick.C:
			case <-sd.Dying():
				return nil
			}
		}
		return nil
	})
	return nil
}

func (svc *service) checkService() {
	if len(svc.Apis) < 1{
		checkError(fmt.Errorf("Services not found in %s", svc.SvcConfig.Monitor.ConfigFile))
	}
	svc.lastCheck = time.Now()
	for _, c := range svc.Apis {
		err := c.CheckService()
		if err != nil {
			log.Error().Msgf("%v", err)
			c.RestartService(svc.App)
		}
	}
}

func (svc *service) Stop() error {
	return nil
}

var (
	ERR_NOT_RUNNING = errors.Errorf("Service is not running")
)
