package main

import (
	"time"

	"ntc.org/mclib/api"
)

type NetApiHealthCheckResponse struct {
	Version      *NetVersionResponse `json:"version"`
	TmoaRootPath string              `json:"tmoaRootPath"`
}

type NetVersionResponse struct {
	Server       string    `json:"server"`
	Version      string    `json:"version"`
	AssemblyPath string    `json:"assemblyPath"`
	FileVersion  string    `json:"fileVersion"`
	LastModDate  time.Time `json:"lastModDate"`
}

type apiHealthCheck struct {
	client      *api.ApiClient
	info        apiHealthCheckInfo
	status      apiStatus
	lastError   *time.Time
	raceCnt     int
	firstRace   *time.Time
	lastRaceLog *time.Time
}

type apiHealthCheckInfo struct {
	Host        string
	Port        int16
	Url         string
	Exe         string
	ServiceName string
}

type apiStatus struct {
	CheckAt time.Time
	Status  string
	Message string
	Err     error
}

type ApiStatuses map[string]*apiStatus
