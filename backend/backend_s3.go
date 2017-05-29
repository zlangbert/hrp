package backend

import (
	"fmt"
	"log"
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
)

var (
	indexFilename = "index.yaml"
)

type s3Backend struct {
	log    *log.Logger
	config *s3Config
	svc    *s3.S3
}

type s3Config struct {
	bucket        string
	prefix        string
	localSyncPath string
}

func NewS3() *s3Backend {

	logger := log.New(os.Stdout, "s3Backend", log.LstdFlags)

	// create aws session
	awsConfig := &aws.Config{Region: aws.String("us-west-2")}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		logger.Fatal("failed to create aws session", err)
	}

	config := &s3Config{
		bucket:        "zlangb",
		prefix:        filepath.Clean("charts"),
		localSyncPath: filepath.Clean("repo"),
	}

	return &s3Backend{
		log:    logger,
		svc:    s3.New(awsSession),
		config: config,
	}
}

/**
 * Get index:
 *
 * read s3 or local?
 */
func (b *s3Backend) GetIndex() ([]byte, error) {
	return nil, nil
}

/**
 * Get chart:
 *
 * read s3 or local?
 */
func (b *s3Backend) GetChart(name string) ([]byte, error) {
	return nil, nil
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
		Bucket: aws.String(b.config.bucket),
		Key:    aws.String(key),
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
 * 3. sync files back to s3
 */
func (b *s3Backend) Reindex() error {

	err := b.localSync()
	if err != nil {
		return err
	}

	cmd := exec.Command("helm", "repo", "index", filepath.Clean(b.config.localSyncPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		b.log.Printf("helm repo index failed: %s", err.Error())
		return err
	}

	// read index file
	file, err := os.Open(filepath.Join(b.config.localSyncPath, indexFilename))
	if err != nil {
		b.log.Printf("failed to open index file: %s", err.Error())
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
		b.log.Printf("failed s3 sync: %s", err.Error())
	}

	return err
}

func (b *s3Backend) handleAwsError(err error) error {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Get error details
			b.log.Println("Error:", awsErr.Code(), awsErr.Message())

			// Prints out full error message, including original error if there was one.
			b.log.Println("Error:", awsErr.Error())

			// Get original error
			if origErr := awsErr.OrigErr(); origErr != nil {
				// operate on original error.
				b.log.Println("Error:", origErr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return err
}
