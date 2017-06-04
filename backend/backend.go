package backend

import (
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/config"
	"mime/multipart"
	"errors"
	"fmt"
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
func NewBackend(cfg *config.AppConfig) (Backend, error) {
	var backend Backend
	switch cfg.BackendName {
	case "s3":
		b, err := newS3(cfg)
		if err != nil {
			return nil, err
		}
		backend = b
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized storage backend: %s", cfg.BackendName))
	}

	// initialize
	err := backend.Initialize()
	if err != nil {
		log.Error(err.Error())
		return nil, errors.New("failed to initialize backend")
	}

	return backend, nil
}
