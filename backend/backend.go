package backend

import (
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/config"
	"mime/multipart"
)

// A Backend is a generic interface for chart storage
type Backend interface {
	Initialize() error
	GetIndex() ([]byte, error)
	GetChart(string) ([]byte, error)
	PutChart(*multipart.FileHeader) error
	Reindex() error
}

// NewBackend is a factory that returns a new Backend based on the config
func NewBackend(cfg *config.AppConfig) Backend {
	var backend Backend
	switch cfg.BackendName {
	case "s3":
		backend = newS3(cfg)
	default:
		log.Fatalf("unrecognized storage backend: %s", cfg.BackendName)
	}

	// initialize
	err := backend.Initialize()
	if err != nil {
		log.Fatal("failed to initialize backend")
	}

	return backend
}
