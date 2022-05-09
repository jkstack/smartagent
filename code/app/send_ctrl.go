package app

import (
	"agent/code/app/hostinfo"
	"agent/code/utils"
	"context"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jkstack/anet"
	"github.com/lwch/logging"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (app *app) SendCome(conn *websocket.Conn) {
	var msg anet.Msg
	msg.Type = anet.TypeCome
	ip := utils.GetIP(conn)
	hostName, _ := os.Hostname()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cpus, err := cpu.InfoWithContext(ctx)
	if err != nil {
		cpus = []cpu.InfoStat{
			{ModelName: "unknown", Cores: 0},
		}
	}
	memory, _ := mem.VirtualMemory()
	host, err := hostinfo.Info()
	if err != nil {
		logging.Error("get host info failed, err=%v", err)
		return
	}
	msg.Come = &anet.ComePayload{
		ID:              app.cfg.ID,
		Version:         app.version,
		IP:              ip,
		MAC:             utils.GetMac(ip),
		HostName:        hostName,
		OS:              host.OS,
		Platform:        host.Platform,
		PlatformVersion: host.PlatformVersion,
		KernelVersion:   host.KernelVersion,
		Arch:            host.KernelArch,
		CPU:             cpus[0].ModelName,
		CPUCore:         uint64(cpus[0].Cores),
	}
	if memory != nil {
		msg.Come.Memory = memory.Total
	}
	err = conn.WriteJSON(msg)
	if err != nil {
		logging.Error("send come message: %v", err)
		return
	}
}
