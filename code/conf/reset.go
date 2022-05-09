package conf

import (
	"agent/code/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lwch/kvconf"
	"github.com/lwch/runtime"
)

func Reset(dir, server string) {
	preReset(dir)

	fmt.Println("开始生成初始化配置...")

	var cfg Configure
	if len(server) > 0 {
		cfg.Server = server
	} else {
		cfg.Server = defaultServer
	}
	fmt.Printf("server=%s\n", cfg.Server)
	cfg.User = defaultUser
	fmt.Printf("user=%s\n", cfg.User)
	cfg.PluginDir = defaultPluginDir
	fmt.Printf("plugin_dir=%s\n", cfg.PluginDir)
	cfg.LogDir = defaultLogDir
	fmt.Printf("log_dir=%s\n", cfg.LogDir)
	cfg.LogSize = defaultLogSize
	fmt.Printf("log_size=%s\n", cfg.LogSize.String())
	cfg.LogRotate = defaultLogRotate
	fmt.Printf("log_rotate=%d\n", cfg.LogRotate)
	cfg.CpuLimit = defaultCpuLimit
	fmt.Printf("cpu_limit=%d\n", cfg.CpuLimit)
	cfg.MemoryLimit = defaultMemoryLimit
	fmt.Printf("memory_limit=%s\n", cfg.MemoryLimit.String())

	f, err := os.Create(dir)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(kvconf.NewEncoder(f).Encode(cfg))
}

func preReset(dir string) {
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	var cfg Configure
	runtime.Assert(kvconf.NewDecoder(f).Decode(&cfg))
	pluginReset(cfg.PluginDir)
	clearPlugin(cfg.PluginDir)
	clearLog(cfg.LogDir)
}

func pluginReset(dir string) {
	fmt.Println("开始重置所有插件...")
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	runtime.Assert(err)
	for _, file := range files {
		name := filepath.Base(file)
		fi, _ := os.Stat(file)
		if !fi.IsDir() {
			continue
		}
		fmt.Printf("开始重置 [%s] 插件...", name)
		if resetPlugin(file) {
			fmt.Println("ok")
		} else {
			fmt.Println("failed")
		}
	}
}

func resetPlugin(dir string) bool {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	runtime.Assert(err)
	ok := true
	for _, ver := range files {
		_, err := utils.ParseVersion(filepath.Base(ver))
		if err != nil {
			continue
		}
		fi, _ := os.Stat(ver)
		if fi.Mode() != 0755 {
			continue
		}
		err = exec.Command(ver, "-reset").Run()
		if err != nil {
			ok = false
		}
	}
	return ok
}

func clearPlugin(dir string) {
	fmt.Println("开始清理插件...")
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	runtime.Assert(err)
	for _, file := range files {
		name := filepath.Base(file)
		rmFile(filepath.Join(file, name+".log"))
		versions, err := filepath.Glob(filepath.Join(file, "*"))
		runtime.Assert(err)
		for _, ver := range versions {
			fi, _ := os.Stat(ver)
			if fi.Mode() != 0755 {
				continue
			}
			_, err = utils.ParseVersion(filepath.Base(ver))
			if err == nil {
				rmFile(ver)
			}
		}
		rmDir(file)
	}
	rmDir(dir)
}

func clearLog(dir string) {
	fmt.Println("开始清理日志...")
	rmFile(filepath.Join(dir, "smartagent.log"))
	files, err := filepath.Glob(filepath.Join(dir, "smartagent.log.*"))
	runtime.Assert(err)
	for _, file := range files {
		name := filepath.Base(file)
		name = strings.TrimPrefix(name, "smartagent.log.")
		_, err := strconv.ParseInt(name, 10, 64)
		if err == nil {
			rmFile(file)
		}
	}
	rmDir(dir)
}

func rmFile(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}
	err := os.Remove(dir)
	if err != nil {
		fmt.Printf("文件 %s 删除失败: %v\n", dir, err)
		return
	}
	fmt.Printf("文件 %s 已删除\n", dir)
}

func rmDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return
	}
	err := os.Remove(dir)
	if err != nil {
		fmt.Printf("删除目录 %s 失败: %v\n", dir, err)
		return
	}
	fmt.Printf("目录 %s 已删除\n", dir)
}
