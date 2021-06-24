
all: build

build:
	go build -v ./cmd/btcchecker

clean:
	rm btcchecker
	rm $(shell find -name *.csv)
