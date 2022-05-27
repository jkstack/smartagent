.PHONY: all distclean prepare

OUTDIR=$(shell realpath release)

VERSION=2.0.3
TIMESTAMP=`date +%s`

BRANCH=`git rev-parse --abbrev-ref HEAD`
HASH=`git log -n1 --pretty=format:%h`
REVERSION=`git log --oneline|wc -l|tr -d ' '`
BUILD_TIME=`date +'%Y-%m-%d %H:%M:%S'`
LDFLAGS="-X 'main.gitBranch=$(BRANCH)' \
-X 'main.gitHash=$(HASH)' \
-X 'main.gitReversion=$(REVERSION)' \
-X 'main.buildTime=$(BUILD_TIME)' \
-X 'main.version=$(VERSION)'"

all: distclean linux.amd64 linux.386 aix.ppc64 windows.amd64 windows.386 msi.amd64 msi.386
	patch -R -d vendor/github.com/kardianos/service < patch/upstart.patch
	cp CHANGELOG.md $(OUTDIR)/CHANGELOG.md
	rm -fr $(OUTDIR)/$(VERSION)/etc $(OUTDIR)/$(VERSION)/opt
version:
	@echo $(VERSION)
linux.amd64: prepare
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent code/*.go
	cd $(OUTDIR)/$(VERSION) && fakeroot tar -czvf smartagent_$(VERSION)_linux_amd64.tar.gz \
		--warning=no-file-changed opt
	go run contrib/pack/release.go -o $(OUTDIR)/$(VERSION) \
		-conf contrib/pack/amd64.yaml \
		-name smartagent -version $(VERSION) \
		-workdir $(OUTDIR)/$(VERSION)
linux.386: prepare
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent code/*.go
	cd $(OUTDIR)/$(VERSION) && fakeroot tar -czvf smartagent_$(VERSION)_linux_386.tar.gz \
		--warning=no-file-changed opt
	go run contrib/pack/release.go -o $(OUTDIR)/$(VERSION) \
		-conf contrib/pack/i386.yaml \
		-name smartagent -version $(VERSION) \
		-workdir $(OUTDIR)/$(VERSION)
aix.ppc64: prepare
	GOOS=aix GOARCH=ppc64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent code/*.go
	cd $(OUTDIR)/$(VERSION) && fakeroot tar -czvf smartagent_$(VERSION)_aix_ppc64.tar.gz \
		--warning=no-file-changed opt
windows.amd64: prepare
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe code/*.go
	unix2dos conf/client.conf
	makensis -DARCH=amd64 \
		-DPRODUCT_VERSION=$(VERSION) \
		-DBINDIR=$(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe \
		-INPUTCHARSET UTF8 contrib/win.nsi
	mv contrib/smartagent_$(VERSION)_windows_amd64.exe $(OUTDIR)/$(VERSION)/smartagent_$(VERSION)_windows_amd64.exe
windows.386: prepare
	mkdir -p src/agent && cp -R code vendor src/agent
	patch -d src/agent/vendor/github.com/shirou/gopsutil/cpu < patch/strings.replaceall.patch
	patch -d src/agent/vendor/github.com/shirou/gopsutil/v3/host < patch/host.patch
	patch -d src/agent/vendor/github.com/shirou/gopsutil/v3/process < patch/process.patch
	GOPATH=$(shell realpath .) GOOS=windows GOARCH=386 CGO_ENABLED=0 go10 build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe src/agent/code/*.go
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/cpu < patch/strings.replaceall.patch
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/v3/host < patch/host.patch
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/v3/process < patch/process.patch
	unix2dos conf/client.conf
	makensis -DARCH=386 \
		-DPRODUCT_VERSION=$(VERSION) \
		-DBINDIR=$(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe \
		-INPUTCHARSET UTF8 contrib/win.nsi
	mv contrib/smartagent_$(VERSION)_windows_386.exe $(OUTDIR)/$(VERSION)/smartagent_$(VERSION)_windows_386.exe
	rm -fr src
msi.amd64: prepare
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe code/*.go
	unix2dos conf/client.conf
	wixl -D PRODUCT_VERSION=$(VERSION) \
		-D RELEASE_DIR=$(OUTDIR)/$(VERSION)/opt/smartagent \
		-o build.msi \
		-v contrib/win.wxs
	mv build.msi $(OUTDIR)/$(VERSION)/smartagent_$(VERSION)_windows_amd64.msi
msi.386: prepare
	mkdir -p src/agent && cp -R code vendor src/agent
	patch -d src/agent/vendor/github.com/shirou/gopsutil/cpu < patch/strings.replaceall.patch
	patch -d src/agent/vendor/github.com/shirou/gopsutil/v3/host < patch/host.patch
	patch -d src/agent/vendor/github.com/shirou/gopsutil/v3/process < patch/process.patch
	GOPATH=$(shell realpath .) GOOS=windows GOARCH=386 CGO_ENABLED=0 go10 build -ldflags $(LDFLAGS) \
		-o $(OUTDIR)/$(VERSION)/opt/smartagent/bin/smartagent.exe src/agent/code/*.go
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/cpu < patch/strings.replaceall.patch
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/v3/host < patch/host.patch
	patch -R -d src/agent/vendor/github.com/shirou/gopsutil/v3/process < patch/process.patch
	unix2dos conf/client.conf
	wixl -D PRODUCT_VERSION=$(VERSION) \
		-D RELEASE_DIR=$(OUTDIR)/$(VERSION)/opt/smartagent \
		-o build.msi \
		-v contrib/win.wxs
	mv build.msi $(OUTDIR)/$(VERSION)/smartagent_$(VERSION)_windows_386.msi
	rm -fr src
prepare:
	rm -fr $(OUTDIR)/$(VERSION)/opt $(OUTDIR)/$(VERSION)/etc
	mkdir -p $(OUTDIR)/$(VERSION)/opt/smartagent/bin \
		$(OUTDIR)/$(VERSION)/opt/smartagent/conf
	cp conf/client.conf $(OUTDIR)/$(VERSION)/opt/smartagent/conf/client.conf
	echo $(VERSION) > $(OUTDIR)/$(VERSION)/opt/smartagent/.version
	go mod vendor
	patch -d vendor/github.com/kardianos/service < patch/upstart.patch
distclean:
	rm -fr $(OUTDIR) src
docker: distclean linux.amd64
	docker build -t smartagent:debian -f docker/Dockerfile.debian release/$(VERSION)/opt/smartagent
	docker build -t smartagent:ubuntu -f docker/Dockerfile.ubuntu release/$(VERSION)/opt/smartagent
	docker build -t smartagent:redhat -f docker/Dockerfile.redhat release/$(VERSION)/opt/smartagent
	docker build -t smartagent:suse -f docker/Dockerfile.suse release/$(VERSION)/opt/smartagent
	docker build -t smartagent:freebsd -f docker/Dockerfile.freebsd release/$(VERSION)/opt/smartagent