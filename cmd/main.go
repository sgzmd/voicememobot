package main

import (
	"context"
	"flag"
	"log"
	bot2 "voicesummary/bot"
	"voicesummary/config"
)

var cfg *config.Config

func main() {
	// Define a string flag for the configuration file path
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")

	// Parse the flags
	flag.Parse()

	// Use the flag value
	var err error
	cfg, err = config.GetConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	processor, err := bot2.NewBotProcessor(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to create bot processor: %v", err)
	}

	processor.ProcessUpdatesLoop()
}
