package plugin

import (
	"agent/code/utils"
	"os"
	"path/filepath"

	"github.com/lwch/runtime"
)

func (mgr *Mgr) load() {
	files, err := filepath.Glob(filepath.Join(mgr.cfg.PluginDir, "*"))
	runtime.Assert(err)
	for _, file := range files {
		name := filepath.Base(file)
		mgr.loadPlugin(file, name)
	}
}

func (mgr *Mgr) loadPlugin(dir, name string) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	runtime.Assert(err)
	var latest utils.Version
	for _, file := range files {
		ver, err := utils.ParseVersion(filepath.Base(file))
		if err != nil {
			continue
		}
		if !ver.Greater(latest) {
			continue
		}
		os.Remove(filepath.Join(dir, latest.String()))
		latest = ver
		md5, err := utils.MD5Checksum(file)
		runtime.Assert(err)
		mgr.lockClear.Lock()
		mgr.versions[name] = pluginInfo{
			dir:     file,
			name:    name,
			version: ver,
			md5:     md5,
		}
		mgr.lockClear.Unlock()
		mgr.lockMove.Lock()
		mgr.md5[name] = md5[:]
		mgr.lockMove.Unlock()
	}
}
