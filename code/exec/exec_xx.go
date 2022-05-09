//go:build !windows && !aix
// +build !windows,!aix

package exec

import (
	"agent/code/conf"
	"os/exec"
	"syscall"
)

type osBased struct {
}

func (ex *osBased) init(cfg *conf.Configure) {
}

func (ex *Executor) chown(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: ex.cfg.UID,
			Gid: ex.cfg.GID,
		},
	}
}
