package report

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jkstack/anet"
	"github.com/shirou/gopsutil/v3/process"
)

type counts struct {
	sync.Mutex
	data map[string]uint64
}

func newCounts() *counts {
	return &counts{data: make(map[string]uint64)}
}

func (cts *counts) inc(name string) {
	cts.Lock()
	defer cts.Unlock()
	if _, ok := cts.data[name]; !ok {
		cts.data[name] = 1
		return
	}
	cts.data[name]++
}

func (cts *counts) add(name string, n uint64) {
	cts.Lock()
	defer cts.Unlock()
	if _, ok := cts.data[name]; !ok {
		cts.data[name] = n
		return
	}
	cts.data[name] += n
}

func (cts *counts) all() uint64 {
	cts.Lock()
	defer cts.Unlock()
	var cnt uint64
	for _, n := range cts.data {
		cnt += n
	}
	return cnt
}

func (cts *counts) copyData() map[string]uint64 {
	cts.Lock()
	defer cts.Unlock()
	ret := make(map[string]uint64)
	for k, v := range cts.data {
		ret[k] = v
	}
	return ret
}

type Data struct {
	startup int64
	version string
	// runtime
	inPackets  uint64
	inBytes    uint64
	outPackets uint64
	outBytes   uint64
	p          *process.Process
	// plugin
	pluginRunning    int64
	pluginUses       *counts
	pluginOutPackets *counts
	pluginOutBytes   *counts
}

func New(version string) *Data {
	p, _ := process.NewProcess(int32(os.Getpid()))
	return &Data{
		startup:          time.Now().Unix(),
		version:          version,
		p:                p,
		pluginUses:       newCounts(),
		pluginOutPackets: newCounts(),
		pluginOutBytes:   newCounts(),
	}
}

func (data *Data) Report(ch chan *anet.Msg) {
	var msg anet.Msg
	msg.Type = anet.TypeAgentInfo

	var info anet.AgentInfo
	info.Version = data.version
	data.basicInfo(&info)
	data.pluginInfo(&info)

	msg.AgentInfo = &info
	ch <- &msg
}

func (data *Data) basicInfo(info *anet.AgentInfo) {
	info.GoVersion = runtime.Version()

	cpu, _ := data.p.CPUPercent()
	info.CpuUsage = float32(cpu)
	mem, _ := data.p.MemoryPercent()
	info.MemoryUsage = mem

	n, _ := runtime.ThreadCreateProfile(nil)
	info.Threads = n
	info.Routines = runtime.NumGoroutine()
	info.Startup = data.startup

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	info.HeapInuse = stats.HeapInuse

	var gc debug.GCStats
	gc.PauseQuantiles = make([]time.Duration, 5)
	debug.ReadGCStats(&gc)

	quantiles := make(map[string]float64)
	for idx, pq := range gc.PauseQuantiles[1:] {
		quantiles[fmt.Sprintf("%d", int(float64(idx+1)/float64(len(gc.PauseQuantiles)-1)*100.))] = pq.Seconds()
	}
	quantiles["0"] = gc.PauseQuantiles[0].Seconds()
	info.GC = quantiles

	info.InPackets = data.inPackets
	info.InBytes = data.inBytes
	info.OutPackets = data.outPackets
	info.OutBytes = data.outBytes
}

func (data *Data) pluginInfo(info *anet.AgentInfo) {
	info.PluginExecd = data.pluginUses.all()
	info.PluginRunning = uint64(data.pluginRunning)
	info.PluginUseCount = data.pluginUses.copyData()
	info.PluginOutPackets = data.pluginOutPackets.copyData()
	info.PluginOutBytes = data.pluginOutBytes.copyData()
}

func (data *Data) IncInPackets() {
	atomic.AddUint64(&data.inPackets, 1)
}

func (data *Data) IncInBytes(n uint64) {
	atomic.AddUint64(&data.inBytes, n)
}

func (data *Data) IncOutPackets() {
	atomic.AddUint64(&data.outPackets, 1)
}

func (data *Data) IncOutBytes(n uint64) {
	atomic.AddUint64(&data.outBytes, n)
}

func (data *Data) IncRunning() {
	atomic.AddInt64(&data.pluginRunning, 1)
}

func (data *Data) DecRunning() {
	atomic.AddInt64(&data.pluginRunning, -1)
}

func (data *Data) UsePlugin(name string) {
	data.pluginUses.inc(name)
}

func (data *Data) PluginReply(name string, n uint64) {
	data.pluginOutPackets.inc(name)
	data.pluginOutBytes.add(name, n)
}
