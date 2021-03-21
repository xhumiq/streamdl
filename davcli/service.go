package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"bitbucket.org/xhumiq/go-mclib/common"

	"bitbucket.org/xhumiq/go-mclib/api"

	"bitbucket.org/xhumiq/go-mclib/microservice"
	"github.com/pkg/errors"
)

type service struct {
	*microservice.App
	SvcConfig *AppConfig
	lastCheck time.Time
	client    *api.HttpClient
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

func (s *service) ListFiles(path string) (err error) {
	//t1 := time.Now()
	var list []*api.DavFileInfo
	if list, err = s.client.NewRequest().MultipleScansDavFileTree(6, path); err != nil {
		return
	}
	list = filterFilesByRecency(list, -48*time.Hour)
	PrintFileList(path, list)
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

func PrintFileList(parentPath string, list []*api.DavFileInfo) {
	tf, ts := int64(0), int64(0)
	for _, l := range list {
		if l.IsDir() {
			tf += l.TotalFiles
			ts += l.TotalSize
			continue
		}
		tf++
		ts += l.Size()
	}
	fmt.Printf("%s Total Files: %d Size: %s\n", parentPath, tf, common.AbbrvBytes(ts))
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
		PrintFileList(l.FullPath(), l.Children)
	}
}

func (s *service) Stop() error {
	return nil
}

var (
	ERR_NOT_RUNNING = errors.Errorf("Service is not running")
)
