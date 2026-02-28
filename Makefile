.PHONY: build build-windows run test test-cover lint fmt release release-windows release-all clean

VERSION ?= dev

build:
	go build -o sssh ./cmd/sssh

build-windows:
	GOOS=windows GOARCH=amd64 go build -o sssh.exe ./cmd/sssh

run: build
	./sssh

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	go vet ./...

fmt:
	go fmt ./...

release:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o sssh ./cmd/sssh

release-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o sssh.exe ./cmd/sssh

release-all: release release-windows

clean:
	rm -f sssh sssh.exe swiftssh swiftssh.exe coverage.out
