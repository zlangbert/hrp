package main

import (
	log "github.com/sirupsen/logrus"
	"github.nike.com/zlangb/hrp/backend"
	"github.nike.com/zlangb/hrp/config"
	"github.nike.com/zlangb/hrp/web"
	"os"
)

func main() {

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

	web.Start(cfg, storageBackend)
}
