package backend

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/config"
	"sync"
)

var (
	indexFilename = "index.yaml"
)

type s3Backend struct {
	config *config.S3Config
	svc    *s3.S3
	lock   *sync.Mutex
}

func newS3(config *config.S3Config) *s3Backend {

	// validate config
	if config.Bucket == "" {
		log.Fatal("s3 config - bucket missing")
	}
	if config.LocalSyncPath == "" {
		log.Fatal("s3 config - local sync path missing")
	}

	// create aws session
	awsConfig := &aws.Config{Region: aws.String("us-west-2")}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		log.Fatal("failed to create aws session", err)
	}

	return &s3Backend{
		svc:    s3.New(awsSession),
		config: config,
		lock:   &sync.Mutex{},
	}
}

/*
 * Initialize backend
 */
func (b *s3Backend) Initialize() error {

	log.Info("initializing...")

	err := b.Reindex()
	if err != nil {
		return b.handleAwsError(err)
	}

	return nil
}

/*
 * Get index:
 *
 * read index from s3
 */
func (b *s3Backend) GetIndex() ([]byte, error) {

	key := filepath.Join(b.config.Prefix, indexFilename)
	return b.getFile(key)
}

/*
 * Get chart:
 *
 * read chart from s3
 */
func (b *s3Backend) GetChart(name string) ([]byte, error) {

	key := filepath.Join(b.config.Prefix, name)
	return b.getFile(key)
}

func (b *s3Backend) getFile(key string) ([]byte, error) {
	result, err := b.svc.GetObject(&s3.GetObjectInput{
		Bucket: &b.config.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, b.handleAwsError(err)
	}

	bytes, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Errorf("failed reading file from s3: %s", key)
		return nil, b.handleAwsError(err)
	}

	return bytes, nil
}

/*
 * Put chart:
 *
 * 1. upload chart to s3
 * 2. reindex
 */
func (b *s3Backend) PutChart(header *multipart.FileHeader) error {
	src, err := header.Open()
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"failed opening file when uploading chart")
	}
	defer src.Close()

	key := filepath.Join(b.config.Prefix, header.Filename)

	_, err = b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.Bucket,
		Key:    &key,
		Body:   src,
	})
	if err != nil {
		return b.handleAwsError(err)
	}

	err = b.Reindex()
	if err != nil {
		return b.handleAwsError(err)
	}

	return nil
}

/*
 * Reindex repository:
 *
 * 1. sync bucket locally
 * 2. regenerate index
 * 3. sync index back to s3
 */
func (b *s3Backend) Reindex() error {

	b.lock.Lock()
	defer b.lock.Unlock()

	log.Info("reindexing...")

	err := b.localSync()
	if err != nil {
		return err
	}

	cmd := exec.Command("helm", "repo", "index", b.config.LocalSyncPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf("helm repo index failed: %s", err.Error())
		return err
	}

	// read index file
	file, err := os.Open(filepath.Join(b.config.LocalSyncPath, indexFilename))
	if err != nil {
		log.Errorf("failed to open index file: %s", err.Error())
		return err
	}
	defer file.Close()

	// upload new index
	key := filepath.Join(b.config.Prefix, indexFilename)
	_, err = b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.Bucket,
		Key:    &key,
		Body:   file,
	})

	if err != nil {
		return b.handleAwsError(err)
	}

	log.Info("done reindexing")

	return nil
}

/*
 * pull bucket contents to local filesystem to reindex
 */
func (b *s3Backend) localSync() error {

	source := "s3://" + filepath.Join(b.config.Bucket, b.config.Prefix)
	target := b.config.LocalSyncPath

	cmd := exec.Command("aws", "s3", "sync", "--delete", source, target)
	if b.config.Debug {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Error("failed s3 sync")
		return err
	}

	return nil
}

/*
 * log details if the error is an aws error
 */
func (b *s3Backend) handleAwsError(err error) error {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Prints out full error message, including original error if there was one.
			log.Error(awsErr.Error())
		} else {
			log.Error(err.Error())
		}
	}
	return err
}
