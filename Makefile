build: windows linux macos

windows: ./dist/sci_hub_p2p_windows_64.exe
linux: ./dist/sci_hub_p2p_linux_64
macos: ./dist/sci_hub_p2p_macos_64

./dist/sci_hub_p2p_macos_64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $@

./dist/sci_hub_p2p_linux_64:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $@

./dist/sci_hub_p2p_windows_64.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $@
