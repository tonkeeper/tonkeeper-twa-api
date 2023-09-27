.PHONY: all fmt test gen run

all: gen fmt test

fmt:
	gofmt -s -l -w $$(go list -f {{.Dir}} ./... | grep -v /vendor/)

test:
	go test $$(go list ./... | grep -v /vendor/) -race -coverprofile cover.out

gen:
	go generate

run:
	go run cmd/api/main.go

build:
	go build -o bin/twa-api ./cmd/api/

