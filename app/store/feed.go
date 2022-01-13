package store

import (
	"encoding/json"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/umputun/feed-master/app/models"
)

const bucketNameFeed = "Feeds"

func (b BoldStore) Iterate(cb func(feed models.Feed) error) error {
	err := b.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketNameFeed))
		if bucket == nil {
			return fmt.Errorf("no bucket for %s", bucketNameFeed)
		}
		c := bucket.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			feed := models.Feed{}
			if err := json.Unmarshal(v, &feed); err != nil {
				log.Printf("[WARN] failed to unmarshal, %v", err)
				continue
			}
			cb(feed)
		}
		return nil
	})
	return err
}

func (b BoldStore) Save(feed models.Feed) (bool, error) {
	var created bool

	err := b.DB.Update(func(tx *bolt.Tx) error {
		bucket, e := tx.CreateBucketIfNotExists([]byte(bucketNameFeed))
		if e != nil {
			return e
		}

		key, e := b.keyFeed(feed)
		if e != nil {
			return e
		}

		data, e := json.Marshal(&feed)
		if e != nil {
			return e
		}

		log.Printf("[INFO] save feed: '%s'", feed.URL)
		e = bucket.Put(key, data)
		if e != nil {
			return e
		}

		created = true
		return e
	})

	return created, err
}

func (b BoldStore) keyFeed(f models.Feed) ([]byte, error) {
	return []byte(f.URL), nil
}
