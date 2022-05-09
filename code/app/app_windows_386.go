package app

import (
	"agent/code/conf"
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/winsvc/eventlog"
	"github.com/btcsuite/winsvc/mgr"
	"github.com/btcsuite/winsvc/svc"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const name = "smartagent"

type Service struct {
	app     *app
	exepath string
}

func New(cfg *conf.Configure, version, confDir string) App {
	exepath, err := os.Executable()
	runtime.Assert(err)
	return &Service{
		app:     new(cfg, version),
		exepath: exepath,
	}
}

func (svr *Service) Stop() error {
	svr.app.stop()
	m, err := mgr.Connect()
	if err != nil {
		logging.Error("mgr connect: %v", err)
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		logging.Error("open service: %v", err)
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(svc.Stop)
	if err != nil {
		logging.Error("control stop: %v", err)
		return fmt.Errorf("could not send control=%d: %v", svc.Stop, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if timeout.Before(time.Now()) {
			logging.Error("timeout")
			return fmt.Errorf("timeout waiting for service to go to state=%d", svc.Stopped)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			logging.Error("query status: %v", err)
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func (svr *Service) Run() error {
	return svc.Run(name, svr)
}

func (svr *Service) Install() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	s, err = m.CreateService(name, svr.exepath, mgr.Config{
		DisplayName: name,
	})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func (svr *Service) Uninstall() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func (svr *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	go svr.app.start()
	for c := range r {
		logging.Info("request: %d => %d", c.CurrentStatus.State, c.Cmd)
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			break
		default:
			logging.Error("unexpected control request #%d", c)
		}
	}
	svr.app.stop()
	changes <- svc.Status{State: svc.StopPending}
	return false, 0
}
