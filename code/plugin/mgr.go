package plugin

import (
	"agent/code/conf"
	"agent/code/utils"
	"crypto/md5"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

type pluginInfo struct {
	dir     string
	name    string
	version utils.Version
	md5     [md5.Size]byte
}

type Mgr struct {
	lockMove  sync.Mutex
	lockClear sync.RWMutex
	cfg       *conf.Configure
	versions  map[string]pluginInfo // name => plugin
	md5       map[string][]byte     // name => md5
	refs      map[string]int        // md5 => count
	chClear   chan pluginInfo
}

func New(cfg *conf.Configure) *Mgr {
	mgr := &Mgr{
		cfg:      cfg,
		versions: make(map[string]pluginInfo),
		md5:      make(map[string][]byte),
		refs:     make(map[string]int),
		chClear:  make(chan pluginInfo),
	}
	mgr.load()
	go mgr.loopClear()
	return mgr
}

func (mgr *Mgr) loopClear() {
	for {
		info := <-mgr.chClear
		mgr.lockClear.RLock()
		cnt := mgr.refs[fmt.Sprintf("%x", info.md5)]
		mgr.lockClear.RUnlock()
		if cnt > 0 {
			continue
		}
		logging.Info("clear old plugin %s, version=%s", info.name, info.version)
		os.Remove(info.dir)
	}
}

func (mgr *Mgr) Add(p *anet.PluginInfo, dir string, ver utils.Version) pluginInfo {
	var old pluginInfo
	mgr.lockClear.Lock()
	defer mgr.lockClear.Unlock()
	if ver.Greater(mgr.versions[p.Name].version) {
		old = mgr.versions[p.Name]
		mgr.versions[p.Name] = pluginInfo{
			dir:     dir,
			name:    p.Name,
			version: ver,
			md5:     p.MD5,
		}
	}
	mgr.refs[fmt.Sprintf("%x", p.MD5)]++
	return old
}

func (mgr *Mgr) Dec(info pluginInfo) {
	enc := fmt.Sprintf("%x", info.md5)
	mgr.lockClear.Lock()
	mgr.refs[enc]--
	cnt := mgr.refs[enc]
	mgr.lockClear.Unlock()
	if len(info.name) > 0 && cnt <= 0 {
		go func() {
			time.Sleep(10 * time.Minute)
			mgr.chClear <- info
		}()
	}
}
