package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "configs/settings.yml", "Path to configuration file")
	flag.Parse()

	log.Printf("Starting SearGo with config: %s", *configPath)
}
