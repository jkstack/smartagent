//go:build !windows && !aix
// +build !windows,!aix

package app

import (
	"agent/code/conf"
	"os"

	"github.com/containerd/cgroups"
	"github.com/lwch/logging"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type osBase struct{}

func (app *osBase) init(cfg *conf.Configure) {
	cpu := int64(cfg.CpuLimit) * 1000
	mem := int64(cfg.MemoryLimit.Bytes())
	cgroup, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/jkstack/smartagent"), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota: &cpu,
		},
		Memory: &specs.LinuxMemory{
			Limit: &mem,
		},
	})
	if err != nil {
		logging.Error("create cgroup: %v", err)
		return
	}
	cgroup.Add(cgroups.Process{
		Pid: os.Getpid(),
	})
}
