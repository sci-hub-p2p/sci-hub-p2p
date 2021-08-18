install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install github.com/GeertJohan/go.rice/rice


.PHONY:: install
