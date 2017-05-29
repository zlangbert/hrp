package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/backend"
	"github.com/zlangbert/hrp/config"
	"github.com/zlangbert/hrp/web"
	"os"
)

func main() {

	// build config
	cfg := config.New()
	err := cfg.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("flag parsing error: %v", err)
	}

	// build backend
	b := backend.NewBackend(cfg)

	// start web server
	web.Start(cfg, b)
}
