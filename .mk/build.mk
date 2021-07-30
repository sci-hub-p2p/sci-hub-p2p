include .mk/args.mk
ProtoFiles = $(shell find . -path "*/.*" -prune -o -name "*.proto" -print)
GoProtoFiles=$(patsubst %.proto,%.pb.go,$(ProtoFiles))

GoSrc = $(shell find . -path "*/.*" -prune -o -name "*.go" -print)
GoSrc += $(GoProtoFiles)

$(GoProtoFiles): $(ProtoFiles)
	protoc --go_out=. $^

windows: Windows
linux: Linux
macos: macOS

# this is used in github ci with `make ${{ runner.os }}`
Windows: dist/client/sci-hub_windows_64.exe dist/daemon/sci-hub_windows_64.exe
Linux: dist/client/sci-hub_linux_64 dist/daemon/sci-hub_linux_64
macOS: dist/client/sci-hub_macos_64 dist/daemon/sci-hub_macos_64

dist/client/sci-hub_windows_64.exe: $(GoSrc)
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/client/

dist/client/sci-hub_linux_64: $(GoSrc)
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/client/

dist/client/sci-hub_macos_64: $(GoSrc)
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/client/


dist/daemon/sci-hub_windows_64.exe: $(GoSrc)
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/daemon/

dist/daemon/sci-hub_linux_64: $(GoSrc)
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/daemon/

dist/daemon/sci-hub_macos_64: $(GoSrc)
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(GoBuildArgs) ./cmd/daemon/

generate: $(GoProtoFiles)

clean::
	rm -rf ./dist

.PHONY:: windows Windows linux Linux macos macOS clean generate
