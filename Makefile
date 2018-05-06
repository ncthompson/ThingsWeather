export GOPATH := $(shell pwd)/..
export GOBIN := $(GOPATH)/bin
export PATH := $(PATH):$(shell pwd)/bin
ifeq ("$(origin COMPILER)", "command line")
COMPILER = $(COMPILER)
endif

install:
	go install ./cmd...


gofmt:
	gofmt -l -s -w .

clean:
	rm bin/*
	rm -r pkg
