
install:
	go install ./cmd...

gofmt:
	gofmt -l -s -w .

clean:
	rm bin/*
	rm -r pkg
