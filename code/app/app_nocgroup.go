//go:build windows || aix
// +build windows aix

package app

import "agent/code/conf"

type osBase struct{}

func (app *osBase) init(cfg *conf.Configure) {
}
