build:
	go build

bin:
	go build -o multibully cmd/multibully.go

test:
	go test

run:
	go run samples/example.go

fmt:
	gofmt -w -s *.go

.PHONY: build test run
