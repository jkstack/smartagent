package app

import (
	"agent/code/conf"
	"agent/code/exec"
	"agent/code/report"
	"agent/code/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	rt "runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jkstack/anet"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const channelBuffer = 10000

type App interface {
	Stop() error
	Run() error
	Install() error
	Uninstall() error
}

type app struct {
	osBase
	version  string
	cfg      *conf.Configure
	remote   *websocket.Conn
	executor *exec.Executor
	// runtime
	chWrite    chan *anet.Msg
	reporter   *report.Data
	mWriteLock sync.Mutex
}

func new(cfg *conf.Configure, version string) *app {
	reporter := report.New(version)
	app := &app{
		version:  version,
		cfg:      cfg,
		executor: exec.New(cfg, reporter),
		reporter: reporter,
	}
	app.init(cfg)
	return app
}

func (app *app) start() {
	stdout := true
	if rt.GOOS == "windows" {
		stdout = false
	}
	logging.SetSizeRotate(app.cfg.LogDir, "smartagent",
		int(app.cfg.LogSize.Bytes()), app.cfg.LogRotate, stdout)
	defer logging.Flush()

	defer utils.Recover("service")

	for i := 0; i < 10; i++ {
		app.run()
		time.Sleep(5 * time.Second)
	}
	os.Exit(255)
}

func (app *app) stop() {
	app.executor.Close()
}

var dialer = websocket.Dialer{
	EnableCompression: true,
}

func (app *app) run() {
	defer utils.Recover("run")
	app.remote = app.connect()
	if app.remote == nil {
		return
	}
	defer app.remote.Close()

	app.chWrite = make(chan *anet.Msg, channelBuffer)

	ctx, cancel := context.WithCancel(context.Background())
	go app.read(ctx, cancel)
	go app.write(ctx, cancel)
	go app.print(ctx)
	go app.keepalive(ctx)
	if app.cfg.StatusReport {
		go app.report(ctx)
	}
	<-ctx.Done()
}

func (app *app) connect() *websocket.Conn {
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/ws/agent", app.cfg.Server), nil)
	runtime.Assert(err)

	app.SendCome(conn)
	logging.Info("%s connected", app.cfg.Server)
	redirect, err := app.waitHandshake(conn, time.Minute)
	if err != nil {
		conn.Close()
		logging.Error("wait handshake: %v", err)
		return nil
	}
	// TODO: redirect
	_ = redirect
	return conn
}

func (app *app) read(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		utils.Recover("read")
		cancel()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, data, err := app.remote.ReadMessage()
		if err != nil {
			logging.Error("read message: %v", err)
			return
		}

		app.reporter.IncInPackets()
		app.reporter.IncInBytes(uint64(len(data)))

		var msg anet.Msg
		err = json.Unmarshal(data, &msg)
		if err != nil {
			logging.Error("read json: %v", err)
			return
		}
		if msg.Plugin != nil {
			app.executor.Exec(data, msg)
			continue
		}
	}
}

func (app *app) write(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		utils.Recover("write")
		cancel()
	}()
	for {
		var data []byte
		var err error
		select {
		case <-ctx.Done():
			return
		case msg := <-app.chWrite:
			if msg == nil {
				continue
			}
			data, err = json.Marshal(msg)
			if err == nil {
				app.mWriteLock.Lock()
				err = app.remote.WriteMessage(websocket.TextMessage, data)
				app.mWriteLock.Unlock()
			}
		case data = <-app.executor.ChWrite():
			if len(data) == 0 {
				continue
			}
			app.mWriteLock.Lock()
			err = app.remote.WriteMessage(websocket.TextMessage, data)
			app.mWriteLock.Unlock()
		}
		if err != nil {
			logging.Error("write message: %v", err)
			return
		}

		app.reporter.IncOutPackets()
		app.reporter.IncOutBytes(uint64(len(data)))
	}
}

func (app *app) print(ctx context.Context) {
	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			if len(app.chWrite) == 0 {
				continue
			}
			logging.Info("write channel size: %d", len(app.chWrite))
		}
	}
}

func (app *app) keepalive(ctx context.Context) {
	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			app.mWriteLock.Lock()
			app.remote.WriteControl(websocket.PingMessage, nil, time.Now().Add(2*time.Second))
			app.mWriteLock.Unlock()
		}
	}
}

func (app *app) report(ctx context.Context) {
	tk := time.NewTicker(app.cfg.StatusReportInterval.Duration())
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			app.reporter.Report(app.chWrite)
		}
	}
}

func (app *app) waitHandshake(conn *websocket.Conn, timeout time.Duration) ([]string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})
	var msg anet.Msg
	err := conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	if msg.Type != anet.TypeHandshake {
		return nil, fmt.Errorf("unexpected message type(handshake): %d", msg.Type)
	}
	if !msg.Handshake.OK {
		return nil, errors.New(msg.Handshake.Msg)
	}
	if len(msg.Handshake.ID) > 0 {
		app.cfg.ResetAgentID(msg.Handshake.ID)
	}
	return msg.Handshake.Redirect, nil
}
