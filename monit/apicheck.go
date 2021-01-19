package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"ntc.org/mclib/common"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"ntc.org/mclib/api"
	"ntc.org/mclib/microservice"
)

func getCheckClients(host string, services []*apiHealthCheckInfo) ([]*apiHealthCheck, error) {
	var cc []*apiHealthCheck
	for _, s := range services {
		if host != "" {
			s.Host = host
		}
		ac, err := api.NewApiClient(fmt.Sprintf("%s:%d", s.Host, s.Port), nil)
		if err != nil {
			return nil, err
		}
		chk := &apiHealthCheck{client: ac, info: *s}
		chk.status.CheckAt = time.Now()
		cc = append(cc, chk)
	}
	return cc, nil
}

func (ac *apiHealthCheck) Status() *apiStatus {
	return &ac.status
}
func (ac *apiHealthCheck) setStatus(status interface{}) error {
	switch v := status.(type) {
	case error:
		if v == ERR_NOT_RUNNING {
			ac.Error(v, "NOT RUNNING", "Service Process %s (%s) is not running!", ac.info.ServiceName, ac.info.Exe)
		} else {
			ac.Error(v, "ERROR", "Service Process %s (%s) Health Error: %+v", ac.info.ServiceName, ac.info.Exe, v)
		}
		return v
	case string:
		ac.Info(v, "Service Process %s (%s) Status: %s", ac.info.ServiceName, ac.info.Exe, v)
	}
	return nil
}

func (ac *apiHealthCheck) CheckService() error {
	log.Debug().Msgf("Check Service Process %s (%s)", ac.info.ServiceName, ac.info.Exe)
	ps, err := microservice.FindProcessByName(ac.info.Exe)
	if err != nil {
		return err
	}
	if ps == nil {
		return ac.setStatus(errors.Wrapf(ERR_NOT_RUNNING, "Check Service %s", ac.info.ServiceName))
	}
	err = ac.CheckHealth()
	if err != nil {
		return err
	}
	return ac.setStatus("OK")
}

func (ac *apiHealthCheck) RestartService(p *microservice.App) error {
	log.Debug().Msgf("Service Process %s (%s) Status: %s", ac.info.ServiceName, ac.info.Exe, p.Status(ac.info.ServiceName))
	_ = p.StopService(ac.info.ServiceName)
	return p.StartService(ac.info.ServiceName)
}

func (ac *apiHealthCheck) CheckHealth() error {
	v := NetApiHealthCheckResponse{}
	_, err := ac.client.Get(&v, ac.info.Url)
	if err != nil {
		return ac.setStatus(err)
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Debug().Msgf("Service Process %s (%s) Response: %s", ac.info.ServiceName, ac.info.Exe, string(b))
	return nil
}

func (ac *apiHealthCheck) Error(err error, status, message string, args ...interface{}){
	s := apiStatus{}
	s.Status = strings.ToUpper(status)
	s.Err = err
	s.CheckAt = time.Now()
	s.Message = message
	if len(args) > 0{
		s.Message = fmt.Sprintf(s.Message, args...)
	}
	logErr := ac.status.Message != s.Message || s.CheckAt.Sub(ac.status.CheckAt).Minutes() > 30
	if logErr{
		logErr = ac.lastError==nil || s.CheckAt.Sub(*ac.lastError).Minutes() > 30
	}
	if logErr{
		log.Error().Msgf(s.Message)
	}else{
		log.Warn().Msgf(s.Message)
	}
	if ac.firstRace!=nil && (s.CheckAt.Sub(*ac.firstRace).Minutes() < 3 || (s.CheckAt.Sub(*ac.firstRace).Seconds()/float64(ac.raceCnt)) >= 1){
		ac.raceCnt++
	}else {
		ac.firstRace = common.TimeToTimePtr(s.CheckAt)
		ac.raceCnt = 0
	}
	ac.lastError = common.TimeToTimePtr(s.CheckAt)
	ac.status = s
}

func (ac *apiHealthCheck) Info(status, message string, args ...interface{}){
	s := ac.status
	s.Status = strings.ToUpper(status)
	s.Err = nil
	s.CheckAt = time.Now()
	s.Message = message
	if len(args) > 0{
		s.Message = fmt.Sprintf(s.Message, args...)
	}
	if ac.firstRace!=nil && ac.raceCnt > 3 && (s.CheckAt.Sub(*ac.firstRace).Seconds()/float64(ac.raceCnt)) >= 1{
		if ac.lastRaceLog==nil || s.CheckAt.Sub(*ac.lastRaceLog).Minutes() > 1{
			log.Error().Msgf("Race Condition Found for : %s -- %s", ac.info.ServiceName, ac.info.Exe)
		}
		ac.lastError = common.TimeToTimePtr(s.CheckAt)
		ac.lastRaceLog = common.TimeToTimePtr(s.CheckAt)
	}else{
		ac.lastError = nil
	}
	log.Info().Msgf(s.Message)
}
