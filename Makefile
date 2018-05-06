
install:
	go install ./cmd...

gofmt:
	gofmt -l -s -w .

getdep:
	dep ensure

clean:
	rm bin/*
	rm -r pkg
