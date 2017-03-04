export GOPATH := $(shell pwd)
export GOBIN := $(GOPATH)/bin
export PATH := $(PATH):$(shell pwd)/bin
ifeq ("$(origin COMPILER)", "command line")
COMPILER = $(COMPILER)
endif

install:
	go install -v weathergetter...

gofmt:
	go fmt weathergetter...

clean:
	rm bin/*
	rm -r pkg
