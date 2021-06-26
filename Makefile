
all: build

build:
	go build -v ./cmd/btcchecker

clean:
	rm -f btcchecker
	rm -f $(shell find -name *.csv)
