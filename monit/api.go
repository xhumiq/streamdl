package main

import (
	"nex.com/nelib/nechi"
)

func NewApi(store *service) *nechi.WebChi {
	sconfig := nechi.NewConfig(store.SvcConfig.Service.Port)
	app := nechi.NewWebApp(&store.AppStatus, sconfig)
	app.ApiHealth("/healthcheck", HealthCheck)
	return app
}

func HealthCheck() ([]string, []error) {
	logs := []string{}
	errs := []error{}
	return logs, errs
}
