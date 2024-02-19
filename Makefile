.PHONY: install gofmt docker clean vendor

getter:
	go build -mod=vendor ./cmd/getter

gofmt:
	gofmt -l -s -w .

docker:
	docker build --tag getter .

clean:
	rm bin/*
	rm -r pkg

vendor:
	go mod tidy
	go mod vendor

lint:
	golangci-lint run

.PHONY: upgrade-vendor
upgrade-vendor:
	go get -u ./...