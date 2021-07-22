MAKEFLAGS += --no-builtin-rules

define helpMessage
Building Targets:

  windows
  linux
  macos

Testing Targets:

  test: simply run tests.
  coverage: test with coverage report './coverage.out'.

Others Targets:

  generate: generated files like protobuf.
  clean: cleanup all auxiliary files.
  install: install protoc-go-gen

endef
export helpMessage

help:
	@echo "$$helpMessage"

include .mk/build.mk
include .mk/test.mk
include .mk/dev.mk

clean::
	rm -rf ./out

.PHONY:: help clean
