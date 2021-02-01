package main

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"ntc.org/mclib/netutils/bitbucket"
	"ntc.org/mclib/netutils/sshutils"
	"os"
	"path/filepath"
	"strings"

	"ntc.org/mclib/common"
	"ntc.org/mclib/microservice"
)

const (
	appName = "hosts"
)

func main() {
	app := NewApp()
	name, addr, group, comments, path, user, key := "", "", "", "", "", "", ""
	port := 0
	app.Cmd("set-ssh [-k,--key <key>] [-p,--port <port>] [-c,--comments <comments>] <file> <name> <user> <addr>", func(c *cli.Context) error {
		sc := app.Config.(*AppConfig)
		cfg := sc.Hosts
		if user == ""{
			user = os.Getenv("USER")
		}
		if addr == ""{
			addr = name
		}
		if path!="" && !common.FileExists(path) {
			sc := filepath.Join(os.Getenv("HOME"), ".ssh", path)
			if common.FileExists(sc) {
				path = sc
			}else {
				sc := filepath.Join(os.Getenv("HOME"), ".ssh", path, "config")
				if common.FileExists(sc) {
					path = sc
				}
			}
		}else if path == "" {
			path = filepath.Join(os.Getenv("HOME"), ".ssh", "config")
		}
		cfile, err := sshutils.ParseSshConfigFile(path)
		if err != nil{
			return err
		}
		comments = strings.Trim(comments, " \t#")
		if key==""{
			key = "id_rsa"
		}
		if !common.FileExists(key){
			bpath := filepath.Join(os.Getenv("HOME"), ".ssh")
			if cfile.File != ""{
				bpath = filepath.Dir(cfile.File)
			}
			sc := filepath.Join(bpath, key)
			if common.FileExists(sc) {
				key = sc
			}else if bpath!= filepath.Join(os.Getenv("HOME"), ".ssh"){
				sc = filepath.Join(filepath.Join(os.Getenv("HOME"), ".ssh"), key)
				if common.FileExists(sc) {
					key = sc
				}
			}
		}
		cfg.Comments = strings.Trim(cfg.Comments, " \t#")
		comments = common.FirstNotEmpty(comments, cfg.Comments)
		cfile.SetEntry(sshutils.SshConfigEntry{
			EntryName:    name,
			Comments:     []string{comments},
			HostName:     addr,
			Port:         port,
			User:         user,
			IdentityFile: key,
		})
		if len(cfile.Entries) > 0{
			sshutils.WriteConfigFile(path, cfile.Entries...)
		}
		return nil
	}, &key, &port, &comments, &path, &name, &user, &addr)
	host := "us.ziongjcc.org"
	app.Cmd("bb-access <host>", func(c *cli.Context) error {
		sc := app.Config.(*AppConfig)
		sd, err := sshutils.RemoteConnect(sshutils.RemoteConfig{
			User:       "mchu",
			Address:    host,
			Port:       9980,
			PrivateKey: "/home/mchu/.ssh/gjcc/gjcc_rsa",
		})
		defer func(){if sd!=nil {_=sd.Close()}}()
		if err!=nil{
			return err
		}
		sd.SetUserAsRoot()
		pub, err := sd.Cat("/root/.ssh/id_rsa.pub")
		if err!=nil{
			return err
		}
		cln, err := bitbucket.NewBBClient("","", sc.Bitbucket)
		if err!=nil{
			return err
		}
		label := "GJCC-" + host[:2]
		_, err = cln.UpsertDeploymentKeys("ChKD144/elzion", bitbucket.BBDeployKeyRequest{
			Label: strings.ToUpper(label),
			Key:   pub,
		})
		if err != nil{
			return err
		}
		return err
	}, host)
	app.Cmd("link-remote", func(c *cli.Context) error {
		err := sshutils.LinkRemoteSsh(sshutils.LinkRemoteSshParams{
			Source:        &sshutils.RemoteConfig{
				User:       "mchu",
				Address:    "jp.ziongjcc.org",
				Port:       9980,
				PrivateKey: "/home/mchu/.ssh/gjcc/gjcc_rsa",
			},
			Destination:        &sshutils.RemoteConfig{
				User:       "mchu",
				Address:    "us.ziongjcc.org",
				Port:       9980,
				PrivateKey: "/home/mchu/.ssh/gjcc/gjcc_rsa",
			},
			SourceUser:    "media",
			DestUser:      "root",
			DestSSHConfig: "sync",
			Comments: "Used to sync media files with japan server",
		})
		println("Completed")
		return err
	}, &group, &comments, &path, &name, &addr)

	app.Cmd("set-host [-g,--group <group>] [-c,--comments <comments>] [-f,--hostpath <host path>] <name> <addr>", func(c *cli.Context) error {
		sc := app.Config.(*AppConfig)
		cfg := sc.Hosts
		path = common.FirstNotEmpty(path, cfg.HostPath)
		hosts, err := sshutils.ParseHostFile(&path)
		if err!=nil{
			return err
		}
		path = hosts.File
		group = common.FirstNotEmpty(group, cfg.HostGroup)
		addr = common.FirstNotEmpty(addr, cfg.IpAddr)
		name = common.FirstNotEmpty(name, cfg.HostName)
		comments = strings.Trim(comments, " \t#")
		cfg.Comments = strings.Trim(cfg.Comments, " \t#")
		comments = common.FirstNotEmpty(comments, cfg.Comments)
		path = common.FirstNotEmpty(path, cfg.HostPath)
		//mapHosts, err := MapHostEntries(hosts, err)
		hosts.Entries = sshutils.SetHostEntry(hosts.Entries, &sshutils.HostEntry{
			Group:     group,
			IpAddress: addr,
			Names:     []string{name},
			Comments:  []string{comments},
		})
		if len(hosts.Entries) > 0{
			sshutils.WriteHostFile(path, hosts.Entries...)
		}
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

func NewHostsUtil(app *microservice.App) *hostsUtil {
	sc := app.Config.(*AppConfig)
	return &hostsUtil{
		SvcConfig: sc,
	}
}

type hostsUtil struct{
	SvcConfig *AppConfig
}
