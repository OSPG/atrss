package main

import (
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"github.com/OSPG/atrss/feed"
	"github.com/OSPG/atrss/ui"
	"io/ioutil"
	"log"
	"os"
)

const CONFIG_DIR = "~/.config/atrss/"

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

var feedManager feed.Manager

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func eventLoop(s *ui.Screen, cfg confStruct) {
	for {
		ev := s.PollEvent()

		x, y := s.GetCursor()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC, tcell.KeyCtrlQ:
				return
			case tcell.KeyCtrlO:
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					item := f.GetUnreadItem(y)
					OpenURL(cfg.Browser, item.Link)
					f.ReadItem(item)
				}
			case tcell.KeyDown:
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					if y < cfg.Layout.BoxHeigh-1 && uint32(y) < f.Unread-1 {
						y++
					}
				} else if y < feedManager.Len()-1 {
					y++
				}
				s.SetCursor(x, y)
			case tcell.KeyUp:
				if y > 0 {
					y--
					s.SetCursor(x, y)
				}
			case tcell.KeyRight:
				if x == 0 {
					ui.FeedIdx = y
					s.SetCursor(s.ItemsColumn, 0)
				}
			case tcell.KeyLeft:
				_, y := s.GetCursor()
				y = ui.FeedIdx
				s.SetCursor(0, y)
			case tcell.KeyCtrlR:
				feedManager.Update()
			case tcell.KeyHome:
				if x == s.ItemsColumn {
					s.SetCursor(s.ItemsColumn, 0)
				} else {
					s.SetCursor(0, 0)
				}
			case tcell.KeyEnd:
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					s.SetCursor(s.ItemsColumn, int(f.Unread-1))
				} else {
					s.SetCursor(0, feedManager.Len()-1)
				}
			}
			switch ev.Rune() {
			case ' ':
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					item := f.GetUnreadItem(y)
					f.ReadItem(item)
				}
			case 'o', 'O':
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					item := f.GetUnreadItem(y)
					OpenURL(cfg.Browser, item.Link)
				}
			case ':':
				if ui.StartCmdLine(s) {
					return
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		s.Redraw(&feedManager)
	}
}

func loadFeeds(s *ui.Screen, db *scribble.Driver, cfg confStruct) {
	for _, url := range cfg.Feeds {
		go func(url string) {
			loadFeed(db, url)
			s.Redraw(&feedManager)
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
	s.Redraw(&feedManager)
	loadFeeds(s, db, cfg)
	eventLoop(s, cfg)
	saveFeeds(db)
}
