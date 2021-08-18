include .mk/args.mk
ProtoFiles = $(shell python ./scripts/wildcard.py proto)
GoProtoFiles=$(patsubst %.proto,%.pb.go,$(ProtoFiles))

GoSrc = $(shell python ./scripts/wildcard.py go)
GoSrc += $(GoProtoFiles)

$(GoProtoFiles): $(ProtoFiles)
	protoc --go_out=. $^

windows: Windows
linux: Linux
macos: macOS

# this is used in github ci with `make ${{ runner.os }}`
Windows: dist/sci-hub_windows_64.exe
Linux: dist/sci-hub_linux_64
macOS: dist/sci-hub_macos_64

dist/sci-hub_windows_64.exe: $(GoSrc) frontend/dist/index.html
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(GoBuildArgs)

dist/sci-hub_linux_64: $(GoSrc) frontend/dist/index.html
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GoBuildArgs)

dist/sci-hub_macos_64: $(GoSrc) frontend/dist/index.html
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(GoBuildArgs)

tmp/dist.zip: scripts/fetch.py
	python ./scripts/fetch.py https://github.com/sci-hub-p2p/sci-hub-p2p-frontend/releases/latest/download/dist.zip tmp/dist.zip

frontend/dist/index.html: tmp/dist.zip
	python scripts/unzip.py tmp/dist.zip frontend

_frontend: frontend/dist/index.html

generate: $(GoProtoFiles) _frontend

clean::
	rm -rf ./dist ./frontend tmp/dist.zip

.PHONY:: windows Windows linux Linux macos macOS clean generate _frontend
