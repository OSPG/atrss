package feed

import (
	"github.com/SlyMarbo/rss"

	"log"
)

type Feed struct {
	Feed    *rss.Feed
	Unread  uint32
	Read    uint32
	Visible bool
	Tags    []string
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
	return &Feed{Feed: f, Unread: f.Unread, Visible: true}, nil
}

func (m *Manager) Append(f *Feed) {
	m.Feeds = append(m.Feeds, f)
}

func (m *Manager) New(url string) (*Feed, error) {
	f, err := Fetch(url)
	if err != nil {
		log.Println("Fetch failed: ", err)
		return nil, err
	}

	m.Append(f)
	return f, nil
}

func (m *Manager) Len() int {
	return len(m.Feeds)
}

func (m *Manager) LenVisible() int {
	counter := 0
	for _, f := range m.Feeds {
		if f.Visible {
			counter++
		}
	}
	return counter
}

func (m *Manager) Get(idx int) *Feed {
	return m.Feeds[idx]
}

func (m *Manager) Update() {
	for _, f := range m.Feeds {
		f.Update()
	}
}

func getVisibleIdx(pos int, feeds []*Feed) int {
	counter := 0
	for n, item := range feeds {
		if item.Visible {
			if counter == pos {
				return n
			}
			counter++
		}
	}
	panic("Could not find that feed")
}

func (m *Manager) GetVisibleFeed(pos int) *Feed {
	idx := getVisibleIdx(pos, m.Feeds)
	return m.Feeds[idx]
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

func (f *Feed) HaveTag(tag string) bool {
	for _, t := range f.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (f *Feed) Update() {
	err := f.Feed.Update()
	if err != nil {
		log.Println("Can not update feed: ", err)
		return
	}

	f.Unread = 0
	for _, i := range f.Feed.Items {
		if !i.Read {
			f.Unread++
		}
	}
}
