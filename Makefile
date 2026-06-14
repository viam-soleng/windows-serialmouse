
GO_BUILD_ENV := GOOS=windows GOARCH=amd64
GO_BUILD_FLAGS := -tags no_cgo -ldflags="-s -w"
MODULE_BINARY := bin/windows-serialmouse.exe

$(MODULE_BINARY): Makefile go.mod *.go cmd/module/*.go 
	GOOS=$(VIAM_BUILD_OS) GOARCH=$(VIAM_BUILD_ARCH) $(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(MODULE_BINARY) cmd/module/main.go

update:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test ./...

module.tar.gz: meta.json $(MODULE_BINARY)
	tar czf $@ meta.json $(MODULE_BINARY)

module: test module.tar.gz

all: test module.tar.gz

setup:
	go mod tidy

upload:
	echo "https://app.viam.com/module/viam-soleng/windows-serialmouse"
	@echo viam module upload --version \"0.0.2\" --platform \"windows/amd64\" module.tar.gz
