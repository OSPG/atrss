package main

import (
	b64 "encoding/base64"
	"github.com/SlyMarbo/rss"
	scribble "github.com/nanobox-io/golang-scribble"
)

func openDB(dir string) *scribble.Driver {
	db, err := scribble.New(Expand(dir), nil)
	check(err)
	return db
}

func loadFeed(db *scribble.Driver, url string, feed *rss.Feed) {
	appendFeed(url)
	f := rss.Feed{}
	encoded_url := b64.StdEncoding.EncodeToString([]byte(url))
	err := db.Read("feed", encoded_url, &f)

	if err == nil {
		err := db.Read("feed", encoded_url, &feed)
		check(err)
	} else {
		err := db.Write("feed", encoded_url, feed)
		check(err)
	}
}

func saveFeeds(db *scribble.Driver) {
	for _, f := range feeds {
		encoded_url := b64.StdEncoding.EncodeToString([]byte(f.UpdateURL))
		err := db.Write("feed", encoded_url, f)
		check(err)
	}
}
