//go:build !(windows && 386)
// +build !windows !386

package app

import (
	"agent/code/conf"

	rt "runtime"

	"github.com/kardianos/service"
	"github.com/lwch/runtime"
)

type Service struct {
	app *app
}

func New(cfg *conf.Configure, version, confDir string) App {
	var user string
	var depends []string
	if rt.GOOS != "windows" {
		user = "root"
		depends = append(depends, "After=network.target")
	}
	appCfg := &service.Config{
		Name:         "smartagent",
		DisplayName:  "smartagent",
		Description:  "smartagent",
		UserName:     user,
		Arguments:    []string{"-conf", confDir},
		Dependencies: depends,
	}
	svr, err := service.New(&Service{app: new(cfg, version)}, appCfg)
	runtime.Assert(err)
	return svr
}

func (svr *Service) Start(s service.Service) error {
	go svr.app.start()
	return nil
}

func (svr *Service) Stop(s service.Service) error {
	svr.app.stop()
	return nil
}
