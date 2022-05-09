package plugin

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

func (mgr *Mgr) Download(msg anet.Msg) (string, error) {
	mgr.lockMove.Lock()
	defer mgr.lockMove.Unlock()
	dir := filepath.Join(mgr.cfg.PluginDir, msg.Plugin.Name)
	os.MkdirAll(dir, 0755)
	dir = filepath.Join(dir, msg.Plugin.Version)
	if runtime.GOOS == "windows" {
		dir += ".exe"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) ||
		!bytes.Equal(mgr.md5[msg.Plugin.Name], msg.Plugin.MD5[:]) {
		logging.Info("download plugin %s...", msg.Plugin.Name)
		err := mgr.download(fmt.Sprintf("http://%s%s", mgr.cfg.Server, msg.Plugin.URI), dir, msg.Plugin.MD5)
		if err != nil {
			return "", err
		}
		mgr.md5[msg.Plugin.Name] = msg.Plugin.MD5[:]
	}
	return dir, nil
}

func (mgr *Mgr) download(url, dir string, enc [md5.Size]byte) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http.get: code=%d, msg=%s", resp.StatusCode, string(data))
	}
	tmpDir := filepath.Join(os.TempDir(), "smartagent", "download")
	os.MkdirAll(tmpDir, 0755)
	f, err := ioutil.TempFile(tmpDir, "plugin")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	f.Close()
	return mgr.mvFile(dir, f.Name())
}

func (mgr *Mgr) mvFile(dst, src string) error {
	err := os.Rename(src, dst)
	if err == nil {
		os.Chmod(dst, 0755)
		return err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		os.Remove(dstFile.Name())
		return err
	}
	err = dstFile.Chmod(0755)
	if err != nil {
		dstFile.Close()
		os.Remove(dstFile.Name())
		return err
	}
	return nil
}
