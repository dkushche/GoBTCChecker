package main

import (
	"github.com/BurntSushi/toml"

	"flag"
	"log"

	"github.com/dkushche/GoBTCChecker/internal/app/btcchecker"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/btcchecker.toml",
		"path to config file")
}

func main() {
	flag.Parse()

	config := btcchecker.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	if err := btcchecker.Start(config); err != nil {
		log.Fatal(err)
	}
}
