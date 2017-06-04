package util

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	// HelmIndexFilename is the filename of the repository index file
	HelmIndexFilename = "index.yaml"
)

// HelmUtil is a wrapper for helm functionality
type HelmUtil interface {
	GenerateIndex(baseURL string, path string) error
	ReadIndex(path string) (io.ReadSeeker, error)
}

type helmUtilImpl struct {
	Debug bool
}

// NewHelmUtil creates a new HelmUtil
func NewHelmUtil(debug bool) HelmUtil {
	return &helmUtilImpl{
		Debug: debug,
	}
}

// GenerateIndex generates a helm repository index at the filesystem path specified
func (u *helmUtilImpl) GenerateIndex(baseURL string, path string) error {

	cmd := exec.Command(
		"helm",
		"repo",
		"index",
		fmt.Sprintf("--url=%s", baseURL),
		path,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("helm repo index failed: %s", err.Error())
		return err
	}

	return nil
}

// ReadIndex reads the repository index in the folder specified by the path
func (u *helmUtilImpl) ReadIndex(path string) (io.ReadSeeker, error) {

	file, err := os.Open(filepath.Join(path, HelmIndexFilename))
	if err != nil {
		log.Errorf("failed to open index file: %s", err.Error())
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("failed reading index fil: %s", err.Error())
		return nil, err
	}

	return bytes.NewReader(data), nil
}
