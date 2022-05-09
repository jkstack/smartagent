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
	fmt.Printf("程序版本: %s\n代码版本: %s.%s.%s\n时间: %s\ngo版本: %s\n",
		version,
		gitBranch, gitHash, gitReversion,
		buildTime,
		rt.Version())
}

func main() {
	cf := flag.String("conf", "", "配置文件所在路径")
	ver := flag.Bool("version", false, "查看版本号")
	act := flag.String("action", "", "install、uninstall或reset")
	server := flag.String("server", "", "服务器地址，格式：<ip>:<port>")
	flag.Parse()

	if *ver {
		showVersion()
		return
	}

	if len(*cf) == 0 {
		fmt.Println("缺少-conf参数")
		os.Exit(1)
	}

	if len(*server) > 0 {
		_, _, err := net.SplitHostPort(*server)
		if err != nil {
			fmt.Printf("-server参数格式错误：%s\n", err)
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
		runtime.Assert(svr.Install())
	case "uninstall":
		runtime.Assert(svr.Stop())
		runtime.Assert(svr.Uninstall())
	default:
		runtime.Assert(svr.Run())
	}
}
