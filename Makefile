.PHONY: install gofmt docker

install:
	go build ./cmd...

gofmt:
	gofmt -l -s -w .

docker:
	docker build --tag getter .

clean:
	rm bin/*
	rm -r pkg
