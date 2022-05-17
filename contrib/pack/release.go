package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/rpm"
	"github.com/lwch/runtime"
	"gopkg.in/yaml.v3"
)

func main() {
	conf := flag.String("conf", "", "配置文件路径")
	out := flag.String("o", "release", "输出目录")
	name := flag.String("name", "smartagent", "输出文件前缀")
	version := flag.String("version", "1.0", "版本号")
	flag.Parse()

	if len(*conf) <= 1 {
		fmt.Println("缺少-conf参数")
		os.Exit(1)
	}

	if len(*name) == 0 {
		fmt.Println("缺少-name参数")
		os.Exit(1)
	}

	if len(*version) == 0 {
		fmt.Println("缺少-version参数")
		os.Exit(1)
	}

	var info nfpm.Info
	f, err := os.Open(*conf)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(yaml.NewDecoder(f).Decode(&info))

	info.Version = *version

	for i, ct := range info.Contents {
		ct.Source = strings.ReplaceAll(ct.Source, "$VERSION", *version)
		info.Contents[i] = ct
	}

	os.MkdirAll(*out, 0755)
	dir := path.Join(*out, *name+"_"+info.Version+"_"+info.Arch)

	deb, err := nfpm.Get("deb")
	runtime.Assert(err)
	debFile, err := os.Create(dir + ".deb")
	runtime.Assert(err)
	defer debFile.Close()
	runtime.Assert(deb.Package(&info, debFile))

	rpm, err := nfpm.Get("rpm")
	runtime.Assert(err)
	rpmFile, err := os.Create(dir + ".rpm")
	runtime.Assert(err)
	defer rpmFile.Close()
	runtime.Assert(rpm.Package(&info, rpmFile))
}
