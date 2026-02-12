package main

import "log"

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
