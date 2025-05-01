package main

import (
	"log"

	"github.com/atnomoverflow/auth-server/config"
	"github.com/atnomoverflow/auth-server/internal/app"
)

func main() {

	cfg, err := config.New("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed on configuration stage %s", err)
	}

	app.Run(cfg)
}
