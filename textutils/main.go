package main

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	_ "golang.org/x/text/encoding"
	_ "golang.org/x/text/encoding/traditionalchinese"
	"io/ioutil"
	"bitbucket.org/xhumiq/go-mclib/microservice"
	"os"
	"path/filepath"
	"strings"
)

const (
	appName = "textutils"
)

func main() {
	app := NewApp()
	name, addr, group, comments, path := "", "", "", "", ""
	app.Cmd("imghref", func(c *cli.Context) error {
		base := "/ntc/web/zion-web/"
		return walkPath(base, false)
	}, &group, &comments, &path, &name, &addr)
	app.Cmd("lowerext", func(c *cli.Context) error {
		base := "/ntc/web/zion-web/"
		filepath.Walk(base, func(fpath string, info os.FileInfo, err error) error {
			if info.IsDir() || strings.Contains(fpath, ".git") {
				return nil
			}
			file := filepath.Base(fpath)
			fe := strings.Split(file,".")
			if len(fe) < 2 || len(fe[len(fe)-1]) > 5{
				return nil
			}
			ext := fe[len(fe)-1]
			//			println(fe[len(fe)-1])
			if strings.ToLower(ext) != ext{
				df := fpath[:len(fpath)-len(ext)]+strings.ToLower(ext)
				if err = os.Rename(fpath, df); err!=nil{
					return err
				}
				println("File:", fpath, df)
			}
			return nil
		})
		return nil
	})
	app.Cmd("big5", func(c *cli.Context) error {
		base := "/ntc/web/zion-web/"
		filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || strings.Contains(path, ".git"){
				return nil
			}
			//rel := path[len(base):]
			se := strings.Split(path,".")
			ext := strings.ToLower(se[len(se)-1])
			if ext != "html" && ext != "htm"{
				if ext == "mno" || ext == "jpg" || ext == "ogv" || ext == "flv" || ext == "mp4" || ext == "css" || ext == "swf" || ext == "htc" || ext == "js" || ext == "png" || ext == "psd" || ext == "gif" || ext == "scc" || ext == "zip" || ext == "pdf" || ext == "doc" || ext == "README"{
					return nil
				}
				//println(ext, path)
				return nil
			}
			content, err := ioutil.ReadFile(path)
			if err!=nil{
				return err
			}
			scnt := string(content)
			if !strings.Contains(scnt, "charset=big5"){
				return nil
			}
			buf, err := Decodebig5(content)
			if err!=nil{
				return err
			}
			scnt = string(buf)
			scnt = strings.Replace(scnt, "charset=big5", "charset=utf-8", 1)
			println("File:>", path, len(scnt))
			ioutil.WriteFile(path, []byte(scnt), 0644)
			return nil
		})
		return nil
	})
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config  := app.Config.(*AppConfig)
			evt.Str("HostPath", config.Hosts.HostPath).
				Msgf("Hosts Ver: %s", app.Build.Version)
		}))
	checkError(err)
}
