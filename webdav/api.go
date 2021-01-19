package main

import (
	"ntc.org/mclib/nechi"
)

func init() {
	nechi.ServicePort = 4100
}

func NewIdApi(srv *service) *nechi.WebChi {
	sconfig := srv.AppConfig.Http
	app := nechi.NewWebApp(&srv.AppStatus, sconfig)
	app.AddWebDav("", sconfig)

	app.ApiHealth("/healthcheck", HealthCheck)
	return app
}

func HealthCheck() ([]string, []error) {
	logs := []string{}
	errs := []error{}
	return logs, errs
}
