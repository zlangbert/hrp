package backend

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/zlangbert/hrp/config"
	"errors"
)

func TestNewBackend_InvalidBackendName(t *testing.T) {

	cfg := config.New()

	b, err := NewBackend(cfg, false)

	assert.Nil(t, b, "nil backend")
	if assert.Error(t, err, "expected error") {
		assert.Equal(t, err, errors.New("unrecognized storage backend: "))
	}
}

func TestNewBackend_FailedCreation(t *testing.T) {

	cfg := config.New()
	cfg.BackendName = "s3"

	b, err := NewBackend(cfg, false)

	assert.Nil(t, b, "nil backend")
	if assert.Error(t, err, "expected error") {
		assert.Contains(t, err.Error(), "s3 config")
	}
}

func TestNewBackend_S3(t *testing.T) {

	cfg := config.New()
	cfg.BackendName = "s3"
	cfg.S3.Bucket = "test"
	cfg.S3.LocalSyncPath = "/tmp"

	b, err := NewBackend(cfg, false)

	assert.Nil(t, err, "nil err")
	assert.IsType(t, &s3Backend{}, b, "expected an s3 backend")
}

func TestNewBackend_InitFail(t *testing.T) {

	cfg := config.New()
	cfg.BackendName = "s3"
	cfg.S3.Bucket = "test"
	cfg.S3.LocalSyncPath = "/tmp"

	b, err := NewBackend(cfg, true)

	assert.Nil(t, b, "nil backend")
	if assert.Error(t, err, "expected error") {
		assert.Contains(t, err.Error(), "failed to initialize backend")
	}
}
