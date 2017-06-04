package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unknown"
)

// AppConfig contains app wide configuration
type AppConfig struct {
	BaseURL     string
	BackendName string
	Debug       bool

	S3         S3Config
	Filesystem FilesystemConfig
}

// S3Config contains s3 specific config
type S3Config struct {
	Region        string
	Bucket        string
	Prefix        string
	LocalSyncPath string
}

// FilesystemConfig contains s3 specific config
type FilesystemConfig struct {
	Path string
}

// New returns a new, empty AppConfig
func New() *AppConfig {
	return &AppConfig{
		S3: S3Config{},
		Filesystem: FilesystemConfig{},
	}
}

// Parse parses the command line flags and builds the config
func (cfg *AppConfig) Parse(args []string) error {

	app := kingpin.New("hrp", "hrp is a helm chart repository proxy with pluggable storage backends")
	app.Version(version)

	app.Flag("base-url", "base url for this instance").
		Required().
		PlaceHolder("https://charts.mycompany.com").
		StringVar(&cfg.BaseURL)

	app.Flag("backend", "storage backend to use (s3)").
		Required().
		PlaceHolder("backend").
		EnumVar(&cfg.BackendName, "s3")

	app.Flag("debug", "app debug mode").
		BoolVar(&cfg.Debug)

	// build s3 backend config
	app.Flag("s3-region", "The AWS region the bucket is in").
		PlaceHolder("us-east-1").
		StringVar(&cfg.S3.Bucket)

	app.Flag("s3-bucket", "The AWS S3 bucket to use for storage").
		PlaceHolder("my-bucket").
		StringVar(&cfg.S3.Bucket)

	app.Flag("s3-prefix", "The S3 prefix to save charts to").
		Default("charts/").
		StringVar(&cfg.S3.Prefix)

	app.Flag("s3-local-sync-path", "The local path to sync to when reindexing").
		Default("/tmp/hrp").
		StringVar(&cfg.S3.LocalSyncPath)

	// build filesystem backend config
	app.Flag("filesystem-path", "The local path to store repository").
		Default("/var/lib/hrp").
		StringVar(&cfg.Filesystem.Path)

	_, err := app.Parse(args)
	if err != nil {
		return err
	}

	return nil
}
