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
	var storageBackend backend.Backend = nil
	switch cfg.BackendName {
	case "s3":
		storageBackend = backend.NewS3(&cfg.S3Config)
	default:
		log.Fatalf("unrecognized storage backend: %s", cfg.BackendName)
	}

	// initialize backend
	storageBackend.Initialize()

	// start web server
	web.Start(cfg, storageBackend)
}
