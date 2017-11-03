package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"io/ioutil"
)

const CONFIG_DIR = "~/.config/atrss/"

var feeds []*rss.Feed
var curX, curY int
var feedIdx int

type ConfStruct struct {
	Feeds []string `yaml:"feeds"`
}

func appendFeed(url string) {
	feed, err := rss.Fetch(url)
	check(err)
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
				if curX == 40 {
					feed := feeds[feedIdx]
					item := feed.Items[curY]
					OpenURL(item.Title)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				if curX == 40 && curY < len(feeds[feedIdx].Items)-1 {
					curY++
					s.ShowCursor(curX, curY)
				} else if curY < len(feeds)-1 {
					curY++
					s.ShowCursor(curX, curY)
				}
			case tcell.KeyUp:
				if curY > 0 {
					curY--
					s.ShowCursor(curX, curY)
				}
			case tcell.KeyRight:
				curX = 40
				feedIdx = curY
				s.ShowCursor(curX, curY)
			case tcell.KeyLeft:
				curX = 0
				curY = feedIdx
				s.ShowCursor(curX, curY)
			}
			switch ev.Rune() {
			case ' ':
				feed := feeds[feedIdx]
				item := feed.Items[curY]
				if !item.Read {
					item.Read = true
					feed.Unread--
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		printLayout(s)
	}
}

func loadFeeds(s tcell.Screen, db *scribble.Driver, cfg ConfStruct) {
	for _, url := range cfg.Feeds {
		dst := feeds[len(feeds)-1]
		go func(dst *rss.Feed) {
			loadFeed(db, url, dst)
			printLayout(s)
		}(dst)
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
	s := initScreen()
	s.ShowCursor(curX, curY)
	printLayout(s)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
	deinitScreen(s)
}
