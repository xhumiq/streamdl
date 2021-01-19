package main

import (
	"time"

	"github.com/pkg/errors"
	"ntc.org/mclib/microservice"
)

type service struct {
	*microservice.App
	SvcConfig *AppConfig
	name      string
	lastCheck time.Time
}

func NewService(app *microservice.App) *service {
	return &service{
		App:       app,
		SvcConfig: app.Config.(*AppConfig),
	}
}

func (svc *service) Name() string {
	return appName
}

var (
	ERR_NOT_RUNNING = errors.Errorf("Service is not running")
)
