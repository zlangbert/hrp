package backend

import "mime/multipart"

/**
 * Generic storage backend for charts
 */
type Backend interface {
	GetIndex() ([]byte, error)
	GetChart(string) ([]byte, error)
	PutChart(*multipart.FileHeader) error
	Reindex() error
}
