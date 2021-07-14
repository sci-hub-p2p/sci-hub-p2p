build: Windows Linux macOS

Windows: ./dist/sci-hub_windows_64.exe
Linux: ./dist/sci-hub_linux_64
macOS: ./dist/sci-hub_macos_64

./dist/sci-hub_macos_64:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $@

./dist/sci-hub_linux_64:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $@

./dist/sci-hub_windows_64.exe:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $@

clean:
	rm dist -rf


