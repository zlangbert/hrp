package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {

	cfg := New()

	assert.NotNil(t, cfg, "cfg not nil")
	assert.NotNil(t, cfg.S3, "cfg.S3 not nil")
}

func TestAppConfig_Parse_MissingFlags(t *testing.T) {

	args := []string{}

	cfg := New()
	err := cfg.Parse(args)

	if assert.Error(t, err, "expected err") {
		assert.Contains(t, err.Error(), "required flag")
	}
}

func TestAppConfig_Parse(t *testing.T) {

	args := []string{
		"--base-url=http://localhost:1323",
		"--backend=s3",
	}

	cfg := New()
	err := cfg.Parse(args)

	assert.Nil(t, err, "expected no error")
	assert.Equal(t, "http://localhost:1323", cfg.BaseURL, "unexpected baseURL")
	assert.Equal(t, "s3", cfg.BackendName, "unexpected backend")
}
