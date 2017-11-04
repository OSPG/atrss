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
				if ui.CurX == 40 {
					feed := feeds[ui.FeedIdx]
					item := feed.Items[ui.CurY]
					OpenURL(item.Link)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				if ui.CurX == 40 && ui.CurY < len(feeds[ui.FeedIdx].Items)-1 {
					ui.CurY++
					s.ShowCursor(ui.CurX, ui.CurY)
				} else if ui.CurY < len(feeds)-1 {
					ui.CurY++
					s.ShowCursor(ui.CurX, ui.CurY)
				}
			case tcell.KeyUp:
				if ui.CurY > 0 {
					ui.CurY--
					s.ShowCursor(ui.CurX, ui.CurY)
				}
			case tcell.KeyRight:
				ui.CurX = 40
				ui.FeedIdx = ui.CurY
				s.ShowCursor(ui.CurX, ui.CurY)
			case tcell.KeyLeft:
				ui.CurX = 0
				ui.CurY = ui.FeedIdx
				s.ShowCursor(ui.CurX, ui.CurY)
			case tcell.KeyCtrlR:
				for _, feed := range feeds {
					feed.Update()
				}
			}
			switch ev.Rune() {
			case ' ':
				feed := feeds[ui.FeedIdx]
				item := feed.Items[ui.CurY]
				if !item.Read {
					item.Read = true
					feed.Unread--
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		ui.PrintLayout(s, feeds)
	}
}

func loadFeeds(s tcell.Screen, db *scribble.Driver, cfg ConfStruct) {
	for _, url := range cfg.Feeds {
		go func(url string) {
			loadFeed(db, url)
			ui.PrintLayout(s, feeds)
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

	s.ShowCursor(ui.CurX, ui.CurY)
	ui.PrintLayout(s, feeds)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
}
