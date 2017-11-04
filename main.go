package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"./ui"
	"io/ioutil"
	"os"
)

const CONFIG_DIR = "~/.config/atrss/"

var feeds []*rss.Feed

type layout struct {
	ColumnWidth int `yaml:"column_width"`
}

type confStruct struct {
	Feeds  []string `yaml:"feeds"`
	Layout layout   `yaml:"layout"`
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

func getUnread(pos int, feed *rss.Feed) int {
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
					feed := feeds[ui.FeedIdx]
					idx := getUnread(y, feed)
					item := feed.Items[idx]
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
				x, y := s.GetCursor()
				if x == 0 {
					ui.FeedIdx = y
					s.SetCursor(40, y)
				}
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
				feed := feeds[ui.FeedIdx]
				idx := getUnread(y, feed)
				item := feed.Items[idx]
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

func loadFeeds(s *ui.Screen, db *scribble.Driver, cfg confStruct) {
	for _, url := range cfg.Feeds {
		go func(url string) {
			loadFeed(db, url)
			s.Redraw(feeds)
		}(url)
	}
}

func loadConfig() confStruct {
	cfgDir := Expand(CONFIG_DIR)
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		err := os.MkdirAll(cfgDir, os.ModePerm)
		check(err)
	}

	cfgFile := cfgDir + "atrss.yml"
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			var conf confStruct
			d, err := yaml.Marshal(&conf)
			check(err)
			err = ioutil.WriteFile(cfgFile, d, os.ModePerm)
			check(err)
		} else {
			check(err)
		}
	}

	var conf confStruct
	err = yaml.UnmarshalStrict([]byte(data), &conf)
	check(err)
	return conf
}

func main() {
	cfg := loadConfig()
	db := openDB(CONFIG_DIR)
	s := ui.InitScreen()
	defer s.DeinitScreen()

	s.SetLayout(cfg.Layout)

	s.SetCursor(0, 0)
	s.Redraw(feeds)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
}
