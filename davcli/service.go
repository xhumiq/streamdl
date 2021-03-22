package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"bitbucket.org/xhumiq/go-mclib/api"
	"bitbucket.org/xhumiq/go-mclib/common"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/cornelk/hashmap"
	"github.com/pkg/errors"
)

type service struct {
	*microservice.App
	SvcConfig *AppConfig
	lastCheck time.Time
	client    *api.HttpClient
	fileInfos hashmap.HashMap
}

func NewService(app *microservice.App) *service {
	cfg := app.Config.(*AppConfig)
	hc, err := api.NewHttpClient(cfg.Download.Domain, api.AuthBasic(cfg.Download.HebronUser, cfg.Download.HebronPwd))
	checkError(err)
	return &service{
		App:       app,
		SvcConfig: app.Config.(*AppConfig),
		client:    hc,
	}
}

func (s *service) GetLatestFiles(days int, paths ...string) (list []*api.DavFileInfo, err error) {
	if list, err = s.client.NewRequest().ScanDavFileTreePaths(6, paths...); err != nil {
		return
	}
	if days < 0{
		days = -days
	}
	if days > 0{
		list = filterFilesByRecency(list, -time.Duration(days * 24) * time.Hour)
	}
	return
}

func (s *service) ListFiles(days int, paths ...string) (err error) {
	var list []*api.DavFileInfo
	list, err = s.GetLatestFiles(days, paths...)
	PrintFileList(nil, list)
	return
}

func filterFilesByRecency(list []*api.DavFileInfo, lastHours time.Duration) (al []*api.DavFileInfo) {
	ed := time.Now().Add(lastHours)
	for _, l := range list {
		if l.IsDir() {
			childs := filterFilesByRecency(l.Children, lastHours)
			if len(childs) < 1 {
				continue
			}
			l.Children = childs
			l.TotalFiles, l.TotalSize = 0, 0
			for _, e := range l.Children{
				if e.IsDir(){
					l.TotalFiles += e.TotalSize
					l.TotalSize += e.TotalSize
				}else{
					l.TotalFiles ++
					l.TotalSize += e.Size()
				}
			}
		} else if l.ModTime().Before(ed) {
			continue
		}
		al = append(al, l)
	}
	SortContentFiles(al)
	return
}

func SortContentFiles(list []*api.DavFileInfo) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].IsDir() != list[j].IsDir() {
			return list[i].IsDir()
		}
		id := list[i].ModTime().Format("2006-01-02")
		jd := list[j].ModTime().Format("2006-01-02")
		if id != jd {
			return id > jd
		}
		in := strings.ToLower(list[i].Name())
		jn := strings.ToLower(list[j].Name())
		if strings.HasPrefix(in, "zion") || strings.HasPrefix(in, "zs") {
			in = "_" + in
		}
		if strings.HasPrefix(jn, "zion") || strings.HasPrefix(jn, "zs") {
			jn = "_" + jn
		}
		if in != jn {
			return in < jn
		}
		return list[i].ModTime().Before(list[j].ModTime())
	})
}

func PrintFileList(parent *api.DavFileInfo, list []*api.DavFileInfo) {
	if parent == nil && len(list) == 1{
		parent = list[0]
	}
	tf, ts := int64(0), int64(0)
	name := ""
	if parent==nil{
		for _, l := range list {
			if l.IsDir() {
				tf += l.TotalFiles
				ts += l.TotalSize
				continue
			}
			tf++
			ts += l.Size()
		}
	}else{
		tf = parent.TotalFiles
		ts = parent.TotalSize
		name = parent.FullPath() + " "
	}
	fmt.Printf("%sTotal Files: %d Size: %s\n", name, tf, common.AbbrvBytes(ts))
	for _, l := range list {
		ftype := "DIR"
		name := l.Name()
		if strings.HasSuffix(name, "mp4") {
			ftype = "MP4"
		} else if strings.HasSuffix(name, "mp3") {
			ftype = "MP3"
		}else if strings.HasSuffix(name, "zip") {
			ftype = "ZIP"
		}else if strings.HasSuffix(name, "pdf") {
			ftype = "PDF"
		}else if strings.HasSuffix(name, "doc") {
			ftype = "DOC"
		}
		size := common.AbbrvBytesPad(l.Size(), 3)
		if l.IsDir() {
			size = common.AbbrvBytesPad(l.TotalSize, 3)
		}
		fmt.Printf("%s %s %s %s\n", ftype, l.ModTime().Format("Jan 02 15:04"), size, name)
	}
	for _, l := range list {
		if !l.IsDir() || len(l.Children) < 1 {
			continue
		}
		fmt.Printf("\n")
		PrintFileList(l, l.Children)
	}
}

func (s *service) Stop() error {
	return nil
}

var (
	ERR_NOT_RUNNING = errors.Errorf("Service is not running")
)
