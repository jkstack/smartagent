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
	conf := flag.String("conf", "", "config file dir")
	out := flag.String("o", "release", "output dir")
	name := flag.String("name", "smartagent", "project name")
	version := flag.String("version", "1.0", "version")
	workdir := flag.String("workdir", "", "work dir")
	flag.Parse()

	if len(*conf) <= 1 {
		fmt.Println("missing [-conf] param")
		os.Exit(1)
	}

	if len(*name) == 0 {
		fmt.Println("missing [-name] param")
		os.Exit(1)
	}

	if len(*version) == 0 {
		fmt.Println("missing [-version] param")
		os.Exit(1)
	}

	if len(*workdir) == 0 {
		fmt.Println("missing [-workdir] param")
		os.Exit(1)
	}

	var info nfpm.Info
	f, err := os.Open(*conf)
	runtime.Assert(err)
	defer f.Close()
	runtime.Assert(yaml.NewDecoder(f).Decode(&info))

	info.Version = *version

	for i, ct := range info.Contents {
		ct.Source = strings.ReplaceAll(ct.Source, "$WORKDIR", *workdir)
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
