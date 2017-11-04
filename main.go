package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"./ui"
	"io/ioutil"
)

const CONFIG_DIR = "~/.config/atrss/"

var feeds []*rss.Feed

type ConfStruct struct {
	Feeds []string `yaml:"feeds"`
}

func fetchFeed(url string) *rss.Feed {
	feed, err := rss.Fetch(url)
	check(err)
	return feed
}

func appendFeed(url string) {
	feed := fetchFeed(url)
	feeds = append(feeds, feed)
}

func eventLoop(s *ui.Screen) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC, tcell.KeyCtrlQ:
				return
			case tcell.KeyCtrlO:
				x, y := s.GetCursor()
				if x == 40 {
					var itemIdx int
					feed := feeds[ui.FeedIdx]
					counter := 0
					for n, e := range feed.Items {
						if e.Read {
							continue
						}
						if counter == y {
							itemIdx = n
							break
						}
						counter++
					}
					item := feed.Items[itemIdx]
					OpenURL(item.Link)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				x, y := s.GetCursor()
				if x == 40 {
					f := feeds[ui.FeedIdx]
					counter := 0
					for _, e := range f.Items {
						if !e.Read {
							counter++
						}
					}
					if y < counter-1 && y < len(f.Items)-1 {
						y++
					}
				} else if y < len(feeds)-1 {
					y++
				}
				s.SetCursor(x, y)
			case tcell.KeyUp:
				x, y := s.GetCursor()
				if y > 0 {
					y--
					s.SetCursor(x, y)
				}
			case tcell.KeyRight:
				_, y := s.GetCursor()
				ui.FeedIdx = y
				s.SetCursor(40, y)
			case tcell.KeyLeft:
				_, y := s.GetCursor()
				y = ui.FeedIdx
				s.SetCursor(0, y)
			case tcell.KeyCtrlR:
				for _, feed := range feeds {
					feed.Update()
				}
			}
			switch ev.Rune() {
			case ' ':
				_, y := s.GetCursor()
				var itemIdx int
				feed := feeds[ui.FeedIdx]
				counter := 0
				for n, e := range feed.Items {
					if e.Read {
						continue
					}
					if counter == y {
						itemIdx = n
						break
					}
					counter++
				}
				item := feed.Items[itemIdx]
				if !item.Read {
					item.Read = true
					feed.Unread--
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		s.Redraw(feeds)
	}
}

func loadFeeds(s *ui.Screen, db *scribble.Driver, cfg ConfStruct) {
	for _, url := range cfg.Feeds {
		go func(url string) {
			loadFeed(db, url)
			s.Redraw(feeds)
		}(url)
	}
}

func loadConfig() ConfStruct {
	cfgFile := Expand(CONFIG_DIR) + "atrss.yml"
	data, err := ioutil.ReadFile(cfgFile)
	check(err)

	var conf ConfStruct
	err = yaml.Unmarshal([]byte(data), &conf)
	check(err)
	return conf
}

func main() {
	cfg := loadConfig()
	db := openDB(CONFIG_DIR)
	s := ui.InitScreen()
	defer s.DeinitScreen()

	s.SetCursor(0, 0)
	s.Redraw(feeds)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
}
