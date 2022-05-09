package conf

import (
	"agent/code/utils"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	rt "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lwch/kvconf"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const (
	defaultServer      = "127.0.0.1:13081"
	defaultUser        = "nobody"
	defaultLogSize     = utils.Bytes(50 * 1000 * 1000)
	defaultLogRotate   = 7
	defaultCpuLimit    = 100
	defaultMemoryLimit = utils.Bytes(512 * 1000 * 1000)
)

type Configure struct {
	ID          string      `kv:"id"`
	Server      string      `kv:"server"`
	User        string      `kv:"user"`
	PluginDir   string      `kv:"plugin_dir"`
	LogDir      string      `kv:"log_dir"`
	LogSize     utils.Bytes `kv:"log_size"`
	LogRotate   int         `kv:"log_rotate"`
	CpuLimit    int         `kv:"cpu_limit"`
	MemoryLimit utils.Bytes `kv:"memory_limit"`
	dir         string
	UID         uint32 `kv:"-"` // user id of user
	GID         uint32 `kv:"-"` // group id of user
}

func Load(dir, abs string) *Configure {
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()

	var bom [3]byte
	n, err := f.Read(bom[:])
	runtime.Assert(err)
	if n < 3 ||
		bom[0] != 0xef ||
		bom[1] != 0xbb ||
		bom[2] != 0xbf {
		_, err = f.Seek(0, io.SeekStart)
		runtime.Assert(err)
	}

	var ret Configure
	runtime.Assert(kvconf.NewDecoder(f).Decode(&ret))
	ret.check(abs)
	ret.replace()
	u, err := user.Lookup(ret.User)
	if err == nil {
		uid, _ := strconv.ParseUint(u.Uid, 10, 32)
		gid, _ := strconv.ParseUint(u.Gid, 10, 32)
		ret.UID = uint32(uid)
		ret.GID = uint32(gid)
	}
	ret.dir = dir
	return &ret
}

func (cfg *Configure) check(abs string) {
	if len(cfg.Server) == 0 {
		panic("missing server config")
	}
	if len(cfg.User) > 0 && rt.GOOS != "windows" {
		_, err := user.Lookup(cfg.User)
		if err != nil {
			panic(fmt.Sprintf("lookup user(%s) failed: %s", cfg.User, err))
		}
	}
	if len(cfg.PluginDir) == 0 {
		logging.Info("reset conf.plugin_dir to default path: %s", defaultPluginDir)
		cfg.PluginDir = defaultPluginDir
	} else if !filepath.IsAbs(cfg.PluginDir) {
		cfg.PluginDir = filepath.Join(abs, cfg.PluginDir)
	}
	if len(cfg.LogDir) == 0 {
		logging.Info("reset conf.log_dir to default path: %s", defaultLogDir)
		cfg.LogDir = defaultLogDir
	} else if !filepath.IsAbs(cfg.LogDir) {
		cfg.LogDir = filepath.Join(abs, cfg.LogDir)
	}
	if cfg.LogSize == 0 {
		logging.Info("reset conf.log_size to default size: %s", defaultLogSize.String())
		cfg.LogSize = defaultLogSize
	}
	if cfg.LogRotate == 0 {
		logging.Info("reset conf.log_roate to default count: %d", defaultLogRotate)
		cfg.LogRotate = defaultLogRotate
	}
	if cfg.CpuLimit == 0 {
		logging.Info("reset conf.cpu_limit to default count: %d", defaultCpuLimit)
		cfg.CpuLimit = defaultCpuLimit
	}
	if cfg.MemoryLimit == 0 {
		logging.Info("reset conf.memory_limit to default size: %s", defaultMemoryLimit.String())
		cfg.MemoryLimit = defaultMemoryLimit
	}
}

func (cfg *Configure) replace() {
	conn, err := net.DialTimeout("tcp", cfg.Server, 10*time.Second)
	runtime.Assert(err)
	defer conn.Close()
	name, err := os.Hostname()
	runtime.Assert(err)

	cfg.ID = strings.Replace(cfg.ID, "$IP", utils.GetIP(conn).String(), -1)
	cfg.ID = strings.Replace(cfg.ID, "$HOSTNAME", name, -1)
}

func (cfg *Configure) ResetAgentID(id string) {
	if id == cfg.ID {
		return
	}
	f, err := os.Create(cfg.dir)
	if err != nil {
		logging.Error("reset agent id: %v", err)
		return
	}
	defer f.Close()
	cfg.ID = id
	err = kvconf.NewEncoder(f).Encode(cfg)
	if err != nil {
		logging.Error("marshal config file: %v", err)
		return
	}
	logging.Info("agent id reset to %s success", id)
}
