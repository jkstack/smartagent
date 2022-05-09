//go:build windows || aix
// +build windows aix

package exec

import (
	"agent/code/conf"
	"os/exec"
)

type osBased struct {
}

func (ex *osBased) init(cfg *conf.Configure) {
}

func (ex *osBased) chown(cmd *exec.Cmd) {
}
