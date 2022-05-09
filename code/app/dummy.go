package app

import (
	rt "runtime"

	"github.com/kardianos/service"
	"github.com/lwch/runtime"
)

type dm struct{}

func Dummy(confDir string) App {
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
	svr, err := service.New(&dm{}, appCfg)
	runtime.Assert(err)
	return svr
}

func (app *dm) Start(s service.Service) error {
	return nil
}

func (app *dm) Stop(s service.Service) error {
	return nil
}
