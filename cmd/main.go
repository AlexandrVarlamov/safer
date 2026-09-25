package main

import (
	"log"

	"github.com/aleksandrvarlamov/safer/internal/config"
)

func main() {
	config, err := config.Load()
    if err != nil {
        log.Fatalf("config: %v", err)
    }
	log.Printf("Config loaded: %+v", config)
}
