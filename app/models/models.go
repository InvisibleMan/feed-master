// Package models contains DAO objects
package models

// Feed presents
type Feed struct {
	// Key []byte
	Title string `json:"title"`
	URL   string `json:"url"`
}

// User presents
type User struct {
	// Key []byte
	ID int64 `json:"id"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`

	FeedKeys []string `json:"feed_keys"`
}

// FeedItem presents
type FeedItem struct {
	// Key []byte
	GUID  string `json:"guid"`
	Title string `json:"first_name"`
}

// FeedUsers presents
type FeedUsers struct {
	// Key []byte // same as FeedKey
	UserKeys []string `json:"user_keys"`
}
