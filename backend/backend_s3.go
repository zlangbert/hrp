package backend

import (
	"io/ioutil"
	"mime/multipart"
	"path/filepath"

	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	log "github.com/sirupsen/logrus"
	"github.com/zlangbert/hrp/config"
	"github.com/zlangbert/hrp/util"
	"sync"
)

type s3Backend struct {
	config   *config.AppConfig
	svc      s3iface.S3API
	awsUtil  util.AwsUtil
	helmUtil util.HelmUtil

	reindexLock *sync.Mutex
}

func newS3(config *config.AppConfig) (*s3Backend, error) {

	// validate config
	if config.S3.Region == "" {
		return nil, errors.New("s3 config - region missing")
	}
	if config.S3.Bucket == "" {
		return nil, errors.New("s3 config - bucket missing")
	}
	if config.S3.LocalSyncPath == "" {
		return nil, errors.New("s3 config - local sync path missing")
	}

	// create aws session
	awsConfig := &aws.Config{Region: aws.String(config.S3.Region)}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		handleAwsError(err)
		return nil, errors.New("failed to create aws session")
	}

	return &s3Backend{
		svc:      s3.New(awsSession),
		config:   config,
		awsUtil:  util.NewAwsUtil(config.Debug),
		helmUtil: util.NewHelmUtil(config.Debug),

		reindexLock: &sync.Mutex{},
	}, nil
}

/*
 * Initialize backend
 */
func (b *s3Backend) Initialize() error {

	log.Info("initializing...")

	err := b.Reindex()
	if err != nil {
		return handleAwsError(err)
	}

	return nil
}

/*
 * Get index:
 *
 * read index from s3
 */
func (b *s3Backend) GetIndex() ([]byte, error) {

	key := filepath.Join(b.config.S3.Prefix, util.HelmIndexFilename)
	return b.getFile(key)
}

/*
 * Get chart:
 *
 * read chart from s3
 */
func (b *s3Backend) GetChart(name string) ([]byte, error) {

	key := filepath.Join(b.config.S3.Prefix, name)
	return b.getFile(key)
}

func (b *s3Backend) getFile(key string) ([]byte, error) {
	result, err := b.svc.GetObject(&s3.GetObjectInput{
		Bucket: &b.config.S3.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, handleAwsError(err)
	}

	bytes, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Errorf("failed reading file from s3: %s", key)
		return nil, handleAwsError(err)
	}

	return bytes, nil
}

/*
 * Put chart:
 *
 * 1. upload chart to s3
 * 2. reindex
 */
func (b *s3Backend) PutChart(filename string, file multipart.File) error {

	key := filepath.Join(b.config.S3.Prefix, filename)

	_, err := b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.S3.Bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return handleAwsError(err)
	}

	err = b.Reindex()
	if err != nil {
		return handleAwsError(err)
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

	b.reindexLock.Lock()
	defer b.reindexLock.Unlock()

	log.Info("reindexing...")

	source := "s3://" + filepath.Join(b.config.S3.Bucket, b.config.S3.Prefix)
	target := b.config.S3.LocalSyncPath

	// local bucket sync
	err := b.awsUtil.Sync(source, target)
	if err != nil {
		return err
	}

	// helm reindex
	err = b.helmUtil.GenerateIndex(b.config.BaseURL, b.config.S3.LocalSyncPath)
	if err != nil {
		return err
	}

	// read index file
	indexData, err := b.helmUtil.ReadIndex(b.config.S3.LocalSyncPath)
	if err != nil {
		return err
	}

	// upload new index
	key := filepath.Join(b.config.S3.Prefix, util.HelmIndexFilename)
	_, err = b.svc.PutObject(&s3.PutObjectInput{
		Bucket: &b.config.S3.Bucket,
		Key:    &key,
		Body:   indexData,
	})

	if err != nil {
		return handleAwsError(err)
	}

	log.Info("done reindexing")

	return nil
}

/*
 * log details if the error is an aws error
 */
func handleAwsError(err error) error {
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
