package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unknown"
)

// AppConfig contains app wide configuration
type AppConfig struct {
	BackendName string
	Debug       bool

	S3Config
}

// S3Config contains s3 specific config
type S3Config struct {
	Bucket        string
	Prefix        string
	LocalSyncPath string
	Debug         bool
}

// New returns a new, empty AppConfig
func New() *AppConfig {
	return &AppConfig{
		S3Config: S3Config{},
	}
}

// Parse parses the command line flags and builds the config
func (cfg *AppConfig) Parse(args []string) error {

	app := kingpin.New("hrp", "hrp is a helm chart repository proxy with pluggable storage backends")
	app.Version(version)

	app.Flag("backend", "storage backend to use (s3)").
		Required().
		PlaceHolder("backend").
		EnumVar(&cfg.BackendName, "s3")

	app.Flag("debug", "app debug mode").
		BoolVar(&cfg.Debug)

	// build s3 backend config
	app.Flag("s3-bucket", "The AWS S3 bucket to use for storage").
		PlaceHolder("my-bucket").
		StringVar(&cfg.S3Config.Bucket)

	app.Flag("s3-prefix", "The S3 prefix to save charts to").
		Default("charts/").
		StringVar(&cfg.S3Config.Prefix)

	app.Flag("s3-local-sync-path", "The local path to sync to when reindexing").
		Default("/tmp/hrp").
		StringVar(&cfg.S3Config.LocalSyncPath)

	_, err := app.Parse(args)
	if err != nil {
		return err
	}

	cfg.S3Config.Debug = cfg.Debug

	return nil
}
