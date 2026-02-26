.PHONY: build build-windows run test test-cover lint fmt release release-windows release-all clean

build:
	go build -o swiftssh ./cmd/swiftssh

build-windows:
	GOOS=windows GOARCH=amd64 go build -o swiftssh.exe ./cmd/swiftssh

run: build
	./swiftssh

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
	go build -ldflags="-s -w" -o swiftssh ./cmd/swiftssh

release-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o swiftssh.exe ./cmd/swiftssh

release-all: release release-windows

clean:
	rm -f swiftssh swiftssh.exe coverage.out
