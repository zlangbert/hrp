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
	HelmIndexFilename = "index.yaml"
)

type HelmUtil interface {
	GenerateIndex(baseUrl string, path string) error
	ReadIndex(path string) (io.ReadSeeker, error)
}

type helmUtilImpl struct {
	Debug bool
}

func NewHelmUtil(debug bool) HelmUtil {
	return &helmUtilImpl{
		Debug: debug,
	}
}

func (u *helmUtilImpl) GenerateIndex(baseUrl string, path string) error {

	cmd := exec.Command(
		"helm",
		"repo",
		"index",
		fmt.Sprintf("--url=%s", baseUrl),
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
