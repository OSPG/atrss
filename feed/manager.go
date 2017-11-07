package feed

import (
	"github.com/SlyMarbo/rss"

	"log"
)

type Feed struct {
	Feed   *rss.Feed
	Unread uint32
	Read   uint32
}

type Manager struct {
	Feeds []*Feed
}

func Fetch(url string) (*Feed, error) {
	f, err := rss.Fetch(url)
	if err != nil {
		log.Printf("Fetch error: %s. Error: %v\n", url, err)
		return nil, err
	}
	return &Feed{Feed: f, Unread: f.Unread}, nil
}

func (m *Manager) Append(f *Feed) {
	m.Feeds = append(m.Feeds, f)
}

func (m *Manager) New(url string) {
	f, err := Fetch(url)
	if err != nil {
		log.Println("Fetch failed: ", err)
		return
	}

	m.Append(f)
}

func (m *Manager) Len() int {
	return len(m.Feeds)
}

func (m *Manager) Get(idx int) *Feed {
	return m.Feeds[idx]
}

func (m *Manager) Update() {
	for _, f := range m.Feeds {
		f.Update()
	}
}

func getUnreadIdx(pos int, feed *rss.Feed) int {
	counter := 0
	for n, item := range feed.Items {
		if !item.Read {
			if counter == pos {
				return n
			}
			counter++
		}
	}
	panic("Could not find that item")
}

func (f *Feed) GetUnreadItem(pos int) *rss.Item {
	idx := getUnreadIdx(pos, f.Feed)
	return f.GetItem(idx)
}

func (f *Feed) GetItem(pos int) *rss.Item {
	return f.Feed.Items[pos]
}

func (f *Feed) ReadItem(item *rss.Item) {
	if !item.Read {
		item.Read = true
		f.Unread--
		f.Read++
	}
}

func (f *Feed) Update() {
	f.Feed.Update()
}
