package main

import (
	b64 "encoding/base64"
	"github.com/OSPG/atrss/feed"
	scribble "github.com/nanobox-io/golang-scribble"
	"log"
)

func openDB(dir string) *scribble.Driver {
	db, err := scribble.New(Expand(dir), nil)
	if err != nil {
		log.Fatalln("Can not open db: ", err)

	}
	return db
}

func loadFeed(db *scribble.Driver, feedConf confFeed) {
	url := feedConf.Url
	f := feed.Feed{}
	encoded_url := b64.StdEncoding.EncodeToString([]byte(url))
	err := db.Read("feed", encoded_url, &f)

	if err == nil {
		updatedFeed, err := feed.Fetch(url)
		if err == nil {

			for _, item := range updatedFeed.Feed.Items {
				if _, ok := f.Feed.ItemMap[item.ID]; !ok {
					f.Feed.ItemMap[item.ID] = struct{}{}
					f.Feed.Items = append(f.Feed.Items, item)
				}
			}

		} else {
			log.Println("Coud not update the feed: ", err)
		}
		counter := uint32(0)
		for _, e := range f.Feed.Items {
			if !e.Read {
				counter++
			}
		}
		f.Unread = counter
		f.Tags = feedConf.Tags
		feedManager.Append(&f)
	} else {
		newFeed, err := feedManager.New(url)
		if err != nil {
			return
		}

		newFeed.Tags = feedConf.Tags
	}
}

func saveFeeds(db *scribble.Driver) {
	for _, f := range feedManager.Feeds {
		encoded_url := b64.StdEncoding.EncodeToString([]byte(f.Feed.UpdateURL))
		err := db.Write("feed", encoded_url, f)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
