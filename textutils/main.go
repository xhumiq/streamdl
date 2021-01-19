package main

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"ntc.org/mclib/microservice"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	appName = "textutils"
)

func main() {
	app := NewApp()
	name, addr, group, comments, path := "", "", "", "", ""
	app.Cmd("imghref", func(c *cli.Context) error {
		base := "/ntc/web/zion-org/"
		exts := map[string]string{}
		filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			fpath := path
			lpath := strings.ToLower(path)
			if !strings.HasSuffix(lpath, "htm") && !strings.HasSuffix(lpath, "html") && !strings.HasSuffix(lpath, "css"){
				return nil
			}
			lpath = lpath[len(base):]
			lang := lpath[:2]
			if lang != "zh" && lang != "en"{
				return nil
			}
			path = path[len(base)+3:]
			dir := filepath.Dir(path)
			level := len(strings.Split(dir, "/"))
			if dir == "."{
				level = 0
			}
			file := filepath.Base(path)
			if level  < 1{
				return nil
			}
			if file != "prophet08_branding.css"{
				//return nil
			}
			content, err := ioutil.ReadFile(fpath)
			if err != nil{
				return err
			}
			scnt := string(content)
			ma := REImgHref.FindAllStringIndex(scnt, -1)
			for _, m := range ma{
				sm := scnt[m[0]:m[1]]
				if !strings.Contains(sm, "images/"){
					continue
				}
				smm := REImgHref.FindStringSubmatch(sm)
				smi := smm[1]
				if smi == ""{
					smi = smm[2]
				}
				if strings.HasPrefix(smi, "80); opacity:0.8}"){
					println("Hello")
					println("!!!!", smi)
				}
				//smf := filepath.Base(smi[1])
				//ext := strings.Split(smf,".")[1]
				//exts[smi[1][:10]] = smi[1]
				//println("Match", i, smf, ext)
				pma := REParentDir.FindAllString(smi, -1)
				pl := len(pma)
				if level != pl{
					println(fpath, smi)
				}
				//println("Level", level, smi)
			}
			return nil
		})
		b, _ := json.MarshalIndent(exts, "", "  ")
		println(string(b))
		return nil
	}, &group, &comments, &path, &name, &addr)
	err := app.Run(
		microservice.RegisterShowVersion(func(app *microservice.App, evt *zerolog.Event) {
			config  := app.Config.(*AppConfig)
			evt.Str("HostPath", config.Hosts.HostPath).
				Msgf("Hosts Ver: %s", app.Build.Version)
		}))
	checkError(err)
}

var (
	REImgHref = regexp.MustCompile(`(?i)(?:=|url)(?:["(']?([^"&)']*?\.\w{2,4})["&)'])`)
	REParentDir = regexp.MustCompile(`\.\./`)
)