package store

import (
	"os"
	"path"
	"time"

	log "github.com/go-pkgz/lgr"
	bolt "go.etcd.io/bbolt"
)

type BoldStore struct {
	DB *bolt.DB
}

func NewBoldStore(dbFile string) (*BoldStore, error) {
	log.Printf("[INFO] bolt (persistent) store, %s", dbFile)
	if err := os.MkdirAll(path.Dir(dbFile), 0700); err != nil {
		return nil, err
	}

	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second}) // nolint
	if err != nil {
		return nil, err
	}

	return &BoldStore{DB: db}, err
}
