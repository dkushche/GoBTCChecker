
all: build

build:
	go build -v ./cmd/btcchecker

test:
	go test -v -race -timeout 30s ./...
