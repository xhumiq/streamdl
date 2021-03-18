package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"bitbucket.org/xhumiq/go-mclib/api"
	"bitbucket.org/xhumiq/go-mclib/common"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type webdavHealth struct {
	Url             string
	LatencyMiliSecs int64
	MinSpeedKBps    float64
	MaxSpeedKBps    float64
	TotalVideo      int
	TotalAudio      int
	Videos          []string
	Audios          []string
	Errors          []*common.ErrorInfo
}

func CheckHealth(url string, config *AppConfig) (res *webdavHealth, err error) {
	log.Info().Msgf("WebDav Check Health: %s", url)
	res = &webdavHealth{
		Url: url,
	}
	hc, err := api.NewHttpClient(url, api.AuthBasic(config.Users.HebronUser, config.Users.HebronPwd))
	if err != nil{
		return nil, err
	}
	t1 := time.Now()
	var list []*api.DavFileInfo
	if list, err = hc.Request().DavListDirFiles(config.Monitor.VideoPath); err != nil {
		return
	}
	for _, f := range list{
		res.Videos = append(res.Videos, f.FullPath())
	}
	res.LatencyMiliSecs = time.Now().Sub(t1).Milliseconds()
	res.TotalVideo = len(res.Videos)
	ap := strings.Split(config.Monitor.AudioPaths, ",")
	if list, err = hc.Request().DavListDirFiles(ap[0]); err != nil {
		return
	}
	la := ""
	ps := int64(0)
	for _, f := range list{
		res.Audios = append(res.Audios, f.FullPath())
		if ps < 1 || (ps > f.Size() && f.Size() > (1024*1024)) || (ps < f.Size() && ps < (1024*1024)){
			ps = f.Size()
			la = f.FullPath()
		}
	}
	res.TotalAudio = len(res.Audios)
	err = CheckSpeed(hc, la, 120, 10)
	return
}

func withSuffix(count int64) string{
	if (count < 1000) {
		return strconv.Itoa(int(count))
	}
	exp := int(math.Log(float64(count)) / math.Log(float64(1024)))
	b := int64(1)
	for i := 0;i<exp-1;i++{
		b *=1024
	}
	st := fmt.Sprintf("%d",count / b)
	d := len(st)-3
	return st[:d] + "." + st[d:] + string("kMGTPE"[exp-1])
}

func CheckSpeed(client *api.HttpClient, url string, maxSecs int, maxTries int) (err error) {
	cnt := int64(0)
	for i:= 0; i < maxTries; i++ {
		t1 := time.Now()
		if cnt, err = client.Request().Get(url).Stream(ioutil.Discard); err != nil {
			return
		}
		println("File", url, withSuffix(cnt), time.Now().Sub(t1).String(), withSuffix(client.Metrics.LastDownloadRate))
	}
	return nil
}

