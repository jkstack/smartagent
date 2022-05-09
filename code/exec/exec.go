package exec

import (
	"agent/code/conf"
	"agent/code/plugin"
	"agent/code/utils"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

const writeBufferSize = 100

type Executor struct {
	osBased
	lockLogger sync.Mutex
	lockPids   sync.RWMutex
	cfg        *conf.Configure
	logger     map[string]logging.Logger // name => logger
	pids       map[int]string            // pid => name
	chWrite    chan []byte
	mgr        *plugin.Mgr
}

func New(cfg *conf.Configure) *Executor {
	ex := &Executor{
		cfg:     cfg,
		logger:  make(map[string]logging.Logger),
		pids:    make(map[int]string),
		chWrite: make(chan []byte, writeBufferSize),
		mgr:     plugin.New(cfg),
	}
	ex.init(cfg)
	return ex
}

func (ex *Executor) Close() {
	pids := make(map[int]string, len(ex.pids))
	ex.lockPids.RLock()
	for pid, name := range ex.pids {
		pids[pid] = name
	}
	ex.lockPids.RUnlock()
	for pid, name := range pids {
		p, err := os.FindProcess(pid)
		if err != nil {
			logging.Error("process %d of plugin %s: %v", pid, name, err)
			continue
		}
		p.Kill()
		logging.Info("killed process %d of plugin %s", pid, name)
	}
}

func (ex *Executor) ChWrite() <-chan []byte {
	return ex.chWrite
}

func (ex *Executor) Exec(data []byte, msg anet.Msg) {
	go func() {
		defer utils.Recover("exec")
		ex.run(data, msg)
	}()
}

func (ex *Executor) run(data []byte, msg anet.Msg) {
	logging.Info("run plugin %s...", msg.Plugin.Name)
	execDir, err := ex.mgr.Download(msg)
	if err != nil {
		ex.err(msg, "download", err)
		return
	}
	argsDir, err := ex.buildArgs(data)
	if err != nil {
		ex.err(msg, "build args", err)
		return
	}
	defer os.Remove(argsDir)
	ex.lockLogger.Lock()
	logger, ok := ex.logger[msg.Plugin.Name]
	if !ok {
		logger = logging.NewRotateSizeLogger(filepath.Join(ex.cfg.PluginDir, msg.Plugin.Name),
			msg.Plugin.Name, int(ex.cfg.LogSize.Bytes()), int(ex.cfg.LogSize), false)
		ex.logger[msg.Plugin.Name] = logger
	}
	ex.lockLogger.Unlock()
	ver, _ := utils.ParseVersion(msg.Plugin.Version)
	clear := ex.mgr.Add(msg.Plugin, execDir, ver)
	defer ex.mgr.Dec(clear)
	ex.exec(execDir, argsDir, msg, logger)
}

func (ex *Executor) err(msg anet.Msg, action string, err error) {
	logging.Error("%s for plugin %s, version=%s, err=%v", action, msg.Plugin.Name,
		msg.Plugin.Version, err)
	ex.sendError(msg.TaskID, fmt.Sprintf("%s, err=%v", action, err))
}

func (ex *Executor) exec(dir, args string, msg anet.Msg, logger logging.Logger) {
	logging.Info("run plugin [%s], task_id=%s...", msg.Plugin.Name, msg.TaskID)
	cmd := exec.Command(dir, "-args", args, "-server", ex.cfg.Server)
	cmd.Dir = filepath.Dir(dir)
	cmd.Env = os.Environ()
	ex.chown(cmd)
	rstdout, wstdout := io.Pipe()
	defer rstdout.Close()
	defer wstdout.Close()
	rstderr, wstderr := io.Pipe()
	defer rstderr.Close()
	defer wstderr.Close()
	cmd.Stdout = wstdout
	cmd.Stderr = wstderr
	err := cmd.Start()
	if err != nil {
		ex.err(msg, "start", err)
		return
	}
	go log(rstderr, logger)
	go send(rstdout, ex.chWrite)

	pid := cmd.Process.Pid

	ex.lockPids.Lock()
	ex.pids[pid] = msg.Plugin.Name
	ex.lockPids.Unlock()

	var code int
	if err := cmd.Wait(); err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			code = ex.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			logging.Error("is not ExitError: %T", err)
		}
		ex.err(msg, "run", err)
	}
	ex.lockPids.Lock()
	delete(ex.pids, pid)
	ex.lockPids.Unlock()
	logging.Info("run plugin [%s] done, task_id=%s, code=%d", msg.Plugin.Name, msg.TaskID, code)
}
