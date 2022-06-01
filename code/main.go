package main

import (
	"agent/code/app"
	"agent/code/conf"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	rt "runtime"

	"github.com/lwch/runtime"
)

var (
	version      string = "0.0.0"
	gitBranch    string = "<branch>"
	gitHash      string = "<hash>"
	gitReversion string = "0"
	buildTime    string = "0000-00-00 00:00:00"
)

func showVersion() {
	fmt.Printf("version: %s\ncode version: %s.%s.%s\nbuild time: %s\ngo version: %s\n",
		version,
		gitBranch, gitHash, gitReversion,
		buildTime,
		rt.Version())
}

func main() {
	cf := flag.String("conf", "", "config file dir")
	ver := flag.Bool("version", false, "show version info")
	act := flag.String("action", "", "install, uninstall or reset")
	server := flag.String("server", "", "server addr, only support for action=reset, eg: <ip>:<port>")
	flag.Parse()

	if *ver {
		showVersion()
		return
	}

	if len(*cf) == 0 {
		fmt.Println("missing -conf argument")
		os.Exit(1)
	}

	if len(*server) > 0 {
		_, _, err := net.SplitHostPort(*server)
		if err != nil {
			fmt.Printf("invalid -server argument: %s\n", err)
			os.Exit(1)
		}
	}

	if *act == "reset" {
		conf.Reset(*cf, *server)
		return
	}

	confDir, err := filepath.Abs(*cf)
	runtime.Assert(err)

	var svr app.App
	if *act != "install" && *act != "uninstall" {
		dir, err := os.Executable()
		runtime.Assert(err)
		cfg := conf.Load(*cf, filepath.Join(filepath.Dir(dir), "/../"))
		svr = app.New(cfg, version, confDir)
	} else {
		svr = app.Dummy(confDir)
	}
	runtime.Assert(err)

	switch *act {
	case "install":
		if svr.Install() != nil {
			svr.Uninstall()
			runtime.Assert(svr.Install())
		}
	case "uninstall":
		svr.Stop()
		err := svr.Uninstall()
		if err != nil {
			fmt.Printf("service uninstall failed: %v\n", err)
		}
	default:
		runtime.Assert(svr.Run())
	}
}
