package main

import (
	"bitbucket.org/xhumiq/go-mclib/nechi"
)

func NewApi(store *service) *nechi.WebChi {
	app := nechi.NewWebApp(&store.AppStatus, &store.SvcConfig.Http, nil)
	app.ApiHealth("/healthcheck", HealthCheck)
	return app
}

func HealthCheck() ([]string, []error) {
	logs := []string{}
	errs := []error{}
	return logs, errs
}
