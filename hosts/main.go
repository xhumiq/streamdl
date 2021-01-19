package main

import (
	"fmt"
	"github.com/kevinburke/ssh_config"
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
	name, addr, group, comments, path := "", "", "", "", ""
	app.Cmd("set-ssh", func(c *cli.Context) error {
		f, _ := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
		cfg, _ := ssh_config.Decode(f)
		for _, host := range cfg.Hosts {
			fmt.Println("patterns:", host.Patterns)
			for _, node := range host.Nodes {
				// Manipulate the nodes as you see fit, or use a type switch to
				// distinguish between Empty, KV, and Include nodes.
				fmt.Println(node.String())
			}
		}
		// Print the config to stdout:
		fmt.Println(cfg.String())
		return nil
	}, &group, &comments, &path, &name, &addr)
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

	app.Cmd("set-host [-g,--group <group>] [-c,--comments <comments>] [-r,--hostpath <host path>] <name> <addr>", func(c *cli.Context) error {
		sc := app.Config.(*AppConfig)
		cfg := sc.Hosts
		hosts, err := sshutils.ParseHostFile(&cfg.HostPath)
		if err!=nil{
			return err
		}
		group = common.FirstNotEmpty(group, cfg.HostGroup)
		addr = common.FirstNotEmpty(addr, cfg.IpAddr)
		name = common.FirstNotEmpty(name, cfg.HostName)
		comments = strings.Trim(comments, " \t#")
		cfg.Comments = strings.Trim(cfg.Comments, " \t#")
		comments = common.FirstNotEmpty(comments, cfg.Comments)
		path = common.FirstNotEmpty(path, cfg.HostPath)
		//mapHosts, err := MapHostEntries(hosts, err)
		hosts = sshutils.SetHostEntry(hosts, &sshutils.HostEntry{
			Group:     group,
			IpAddress: addr,
			Names:     []string{name},
			Comments:  []string{comments},
		})
		path = "/tmp/hosts-new"
		if len(hosts) > 0{
			sshutils.WriteHostFile(path, hosts...)
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
