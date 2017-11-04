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

func eventLoop(s tcell.Screen) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC, tcell.KeyCtrlQ:
				return
			case tcell.KeyCtrlO:
				x, y := ui.GetCursor()
				if x == 40 {
					feed := feeds[ui.FeedIdx]
					item := feed.Items[y]
					OpenURL(item.Link)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				x, y := ui.GetCursor()
				if x == 40 && y < len(feeds[ui.FeedIdx].Items)-1 {
					y++
				} else if y < len(feeds)-1 {
					y++
				}
				ui.SetCursor(x, y)
			case tcell.KeyUp:
				x, y := ui.GetCursor()
				if y > 0 {
					y--
					ui.SetCursor(x, y)
				}
			case tcell.KeyRight:
				_, y := ui.GetCursor()
				ui.FeedIdx = y
				ui.SetCursor(40, y)
			case tcell.KeyLeft:
				_, y := ui.GetCursor()
				y = ui.FeedIdx
				ui.SetCursor(0, y)
			case tcell.KeyCtrlR:
				for _, feed := range feeds {
					feed.Update()
				}
			}
			switch ev.Rune() {
			case ' ':
				_, y := ui.GetCursor()
				feed := feeds[ui.FeedIdx]
				item := feed.Items[y]
				if !item.Read {
					item.Read = true
					feed.Unread--
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		ui.Redraw(s, feeds)
	}
}

func loadFeeds(s tcell.Screen, db *scribble.Driver, cfg ConfStruct) {
	for _, url := range cfg.Feeds {
		go func(url string) {
			loadFeed(db, url)
			ui.Redraw(s, feeds)
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
	defer ui.DeinitScreen(s)

	ui.SetCursor(0, 0)
	ui.Redraw(s, feeds)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
}
