package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"github.com/OSPG/atrss/ui"
	"io/ioutil"
	"log"
	"os"
)

const CONFIG_DIR = "~/.config/atrss/"

var feeds []*rss.Feed

type layout struct {
	ColumnWidth int `yaml:"column_width"`
	ItemsMargin int `yaml:"items_margin"`
	BoxHeigh    int `yaml:"items_box_heigh"`
}

type confStruct struct {
	Browser  string   `yaml:"browser"`
	Log_file string   `yaml:"log_file"`
	Feeds    []string `yaml:"feeds"`
	Layout   layout   `yaml:"layout"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func fetchFeed(url string) (*rss.Feed, error) {
	feed, err := rss.Fetch(url)
	if err != nil {
		log.Printf("Fetch error: %s. Error: %v\n", url, err)
		return nil, err
	}
	return feed, nil
}

func appendFeed(url string) {
	feed, err := fetchFeed(url)
	if err != nil {
		log.Println("Fetch failed: ", err)
		return
	}

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

func eventLoop(s *ui.Screen, cfg confStruct) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC, tcell.KeyCtrlQ:
				return
			case tcell.KeyCtrlO:
				x, y := s.GetCursor()
				if x == s.ItemsColumn {
					feed := feeds[ui.FeedIdx]
					idx := getUnread(y, feed)
					item := feed.Items[idx]
					OpenURL(cfg.Browser, item.Link)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				x, y := s.GetCursor()
				if x == s.ItemsColumn {
					f := feeds[ui.FeedIdx]
					counter := 0
					for _, e := range f.Items {
						if !e.Read {
							counter++
						}
					}
					if y < cfg.Layout.BoxHeigh-1 && y < counter-1 && y < len(f.Items)-1 {
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
					s.SetCursor(s.ItemsColumn, 0)
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
				x, _ := s.GetCursor()
				if x == s.ItemsColumn {
					_, y := s.GetCursor()
					feed := feeds[ui.FeedIdx]
					idx := getUnread(y, feed)
					item := feed.Items[idx]
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case 'o', 'O':
				x, y := s.GetCursor()
				if x == s.ItemsColumn {
					feed := feeds[ui.FeedIdx]
					idx := getUnread(y, feed)
					item := feed.Items[idx]
					OpenURL(cfg.Browser, item.Link)
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

func initLogger(cfg confStruct) {
	logFile := Expand(cfg.Log_file)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	log.SetOutput(file)
	log.Println("Logger initalized")
}

func main() {
	cfg := loadConfig()
	initLogger(cfg)
	db := openDB(CONFIG_DIR)
	s := ui.InitScreen()
	defer s.DeinitScreen()

	s.SetLayout(cfg.Layout)

	s.SetCursor(0, 0)
	s.Redraw(feeds)
	loadFeeds(s, db, cfg)
	eventLoop(s, cfg)
	saveFeeds(db)
}
