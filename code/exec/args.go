package exec

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

func (ex *Executor) buildArgs(data []byte) (string, error) {
	dir := filepath.Join(os.TempDir(), "smartagent", "args")
	os.MkdirAll(dir, 0755)
	f, err := ioutil.TempFile(dir, "arg")
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		err = f.Chmod(0644)
		if err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}
