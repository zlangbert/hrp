package backend

import (
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/config"
	"github.com/zlangbert/hrp/util"
	"io/ioutil"
	"mime/multipart"
	"path/filepath"
	"sync"
)

type filesystemBackend struct {
	config   *config.AppConfig
	helmUtil util.HelmUtil

	reindexLock *sync.Mutex
}

func newFilesystem(config *config.AppConfig) (*filesystemBackend, error) {

	return &filesystemBackend{
		config:   config,
		helmUtil: util.NewHelmUtil(config.Debug),

		reindexLock: &sync.Mutex{},
	}, nil
}

func (b *filesystemBackend) Initialize() error {

	log.Info("Initializing...")

	return b.Reindex()
}

func (b *filesystemBackend) GetIndex() ([]byte, error) {

	index, err := b.helmUtil.ReadIndex(b.config.Filesystem.Path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(index)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *filesystemBackend) GetChart(name string) ([]byte, error) {

	path := filepath.Join(b.config.Filesystem.Path, name)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *filesystemBackend) PutChart(filename string, file multipart.File) error {

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	path := filepath.Join(b.config.Filesystem.Path, filename)

	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (b *filesystemBackend) Reindex() error {

	b.reindexLock.Lock()
	defer b.reindexLock.Unlock()

	err := b.helmUtil.GenerateIndex(b.config.BaseURL, b.config.Filesystem.Path)
	if err != nil {
		return err
	}

	return nil
}
