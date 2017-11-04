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

func loadFeed(db *scribble.Driver, url string) {
	f := rss.Feed{}
	encoded_url := b64.StdEncoding.EncodeToString([]byte(url))
	err := db.Read("feed", encoded_url, &f)

	if err == nil {
		updatedFeed := fetchFeed(url)
		for _, item := range updatedFeed.Items {
			if _, ok := f.ItemMap[item.ID]; !ok {
				f.ItemMap[item.ID] = struct{}{}
				f.Items = append(f.Items, item)
			}
		}

		counter := uint32(0)
		for _, e := range f.Items {
			if !e.Read {
				counter++
			}
		}
		f.Unread = counter
		feeds = append(feeds, &f)
	} else {
		appendFeed(url)
		feed := feeds[len(feeds)-1]
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
