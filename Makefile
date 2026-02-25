.PHONY: build run test test-cover lint fmt release clean

build:
	go build -o swiftssh ./cmd/swiftssh

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

clean:
	rm -f swiftssh swiftssh.exe coverage.out
