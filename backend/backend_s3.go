package backend

import (
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var (
	indexFilename = "index.yaml"
)

type s3Backend struct {
	config *s3Config
	svc    *s3.S3
}

type s3Config struct {
	bucket        string
	prefix        string
	localSyncPath string
}

func NewS3() *s3Backend {

	// create aws session
	awsConfig := &aws.Config{Region: aws.String("us-west-2")}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		log.Fatal("failed to create aws session", err)
	}

	config := &s3Config{
		bucket:        "zlangb",
		prefix:        filepath.Clean("charts"),
		localSyncPath: filepath.Clean("repo"),
	}

	return &s3Backend{
		svc:    s3.New(awsSession),
		config: config,
	}
}

/**
 * Get index:
 *
 * read index from s3
 */
func (b *s3Backend) GetIndex() ([]byte, error) {

	key := filepath.Join(b.config.prefix, indexFilename)
	return b.getFile(key)
}

/**
 * Get chart:
 *
 * read chart from s3
 */
func (b *s3Backend) GetChart(name string) ([]byte, error) {

	key := filepath.Join(b.config.prefix, name)
	return b.getFile(key)
}

func (b *s3Backend) getFile(key string) ([]byte, error) {
	result, err := b.svc.GetObject(&s3.GetObjectInput{
		Bucket: &b.config.bucket,
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

/**
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

	key := filepath.Join(b.config.prefix, header.Filename)

	_, err = b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.bucket,
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

/**
 * Reindex repository:
 *
 * 1. sync bucket locally
 * 2. regenerate index
 * 3. sync index back to s3
 */
func (b *s3Backend) Reindex() error {

	err := b.localSync()
	if err != nil {
		return err
	}

	cmd := exec.Command("helm", "repo", "index", b.config.localSyncPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf("helm repo index failed: %s", err.Error())
		return err
	}

	// read index file
	file, err := os.Open(filepath.Join(b.config.localSyncPath, indexFilename))
	if err != nil {
		log.Errorf("failed to open index file: %s", err.Error())
		return err
	}
	defer file.Close()

	// upload new index
	key := filepath.Join(b.config.prefix, indexFilename)
	_, err = b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.bucket,
		Key:    &key,
		Body:   file,
	})

	return b.handleAwsError(err)
}

/*
 * pull bucket contents to local filesystem to reindex
 */
func (b *s3Backend) localSync() error {

	source := "s3://" + filepath.Join(b.config.bucket, b.config.prefix)
	target := b.config.localSyncPath

	cmd := exec.Command("aws", "s3", "sync", "--delete", source, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Errorf("failed s3 sync: %s", err.Error())
	}

	return err
}

/*
 * log details if the error is an aws error
 */
func (b *s3Backend) handleAwsError(err error) error {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Get error details
			log.Error("Error:", awsErr.Code(), awsErr.Message())

			// Prints out full error message, including original error if there was one.
			log.Error("Error:", awsErr.Error())

			// Get original error
			if origErr := awsErr.OrigErr(); origErr != nil {
				// operate on original error.
				log.Error("Error:", origErr.Error())
			}
		} else {
			log.Error(err.Error())
		}
	}
	return err
}
