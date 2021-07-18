LDFLAGS = -X 'sci_hub_p2p/pkg/variable.Ref=${REF}'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.Commit=${SHA}'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.Builder=$(shell go version)'
LDFLAGS += -X 'sci_hub_p2p/pkg/variable.BuildTime=${TIME}'

GoBuildArgs = -ldflags "-s -w $(LDFLAGS)" -tags disable_libutp

MAKEFLAGS += --no-builtin-rules
GoSrc =  $(shell find . -path "*/.*" -prune -o -name "*.go" -print)
ifeq ($(OS),Windows_NT)
	DefaultRule := Windows
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		DefaultRule = Linux
	else ifeq ($(UNAME_S),Darwin)
		DefaultRule = macOS
	else
		DefaultRule = None
	endif
endif

# default: current OS only build
$(DefaultRule):

None:
	@echo "Not Support System"
	@echo "try 'make all' to cross compile"

# this is just alias for some lazy person like myself
windows: Windows
linux: Linux
macos: macOS

all: Windows Linux macOS

Windows: dist/sci-hub_windows_64.exe
Linux: dist/sci-hub_linux_64
macOS: dist/sci-hub_macos_64

dist/sci-hub_windows_64.exe: $(GoSrc)
	env GOOS=windows GOARCH=amd64 go build -o $@ $(GoBuildArgs)

dist/sci-hub_linux_64: $(GoSrc)
	env GOOS=linux GOARCH=amd64 go build -o $@ $(GoBuildArgs)

dist/sci-hub_macos_64: $(GoSrc)
	env GOOS=darwin GOARCH=amd64 go build -o $@ $(GoBuildArgs)

testdata/sm_00900000-00999999.torrent:
	bash ./fetch.bash

testdata: testdata/sm_00900000-00999999.torrent

test: testdata
	go test ./...

coverage.out: testdata
	go test -v -covermode=atomic -coverprofile=coverage.out -count=1 ./...

coverage: coverage.out

clean:
	rm dist -rf
	rm -rf ./dist \
		  ./testdata/sm_00900000-00999999.torrent \
		  ./coverage.out \
		  ./out

.PHONY: Windows Linux macOS test coverage clean testdata None windows linux macos all
