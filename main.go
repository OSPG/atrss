package main

import (
	"github.com/gdamore/tcell"
	scribble "github.com/nanobox-io/golang-scribble"

	"github.com/OSPG/atrss/backend"
	"github.com/OSPG/atrss/feed"
	"github.com/OSPG/atrss/ui"
	"log"
	"os"
)

const CONFIG_DIR = "~/.config/atrss/"

var feedManager feed.Manager

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func eventLoop(s *ui.Screen, cfg backend.Config) {
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
				_, h := s.GetSize()
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					if f.Unread == 0 {
						y = ui.FeedIdx
						x = 0
					} else if y < cfg.Layout.BoxHeigh-1 && uint32(y) < f.Unread-1 {
						y++
					}
				} else if y < h-1 && y < feedManager.LenVisible()-1 {
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

					if f.Unread-1 > uint32(y) {
						y = cfg.Layout.BoxHeigh - 1
					} else {
						y = int(f.Unread - 1)
					}

				} else {
					x = 0
					y = feedManager.Len() - 1
				}
				s.SetCursor(x, y)
			}
			switch ev.Rune() {
			case ' ':
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					if f.Unread != 0 {
						item := f.GetUnreadItem(y)
						f.ReadItem(item)
					}
				}
			case 'o', 'O':
				if x == s.ItemsColumn {
					f := feedManager.Get(ui.FeedIdx)
					item := f.GetUnreadItem(y)
					OpenURL(cfg.Browser, item.Link)
				}
			case 'R':
				f := feedManager.Get(ui.FeedIdx)
				f.Update()
			case ':':
				if ui.StartCmdLine(s, &feedManager) {
					return
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		s.Redraw(&feedManager)
	}
}

func loadFeeds(s *ui.Screen, db *scribble.Driver, cfg backend.Config) {
	for _, f := range cfg.Feeds {
		//go func(f backend.ConfFeed) {
		//	loadFeed(db, f, cfg.UpdateStartup)
		//	s.Redraw(&feedManager)
		//}(f)
		loadFeed(db, f, cfg.UpdateStartup)
		s.Redraw(&feedManager)
	}
}

func initLogger(cfg backend.Config) {
	logFile := Expand(cfg.Log_file)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	log.SetOutput(file)
	log.Println("Logger initalized")
}

func main() {
	cfg := backend.LoadConfig(Expand(CONFIG_DIR))
	initLogger(*cfg)
	db := openDB(CONFIG_DIR)
	s := ui.InitScreen()
	defer s.DeinitScreen()

	s.SetLayout(cfg.Layout)

	s.SetCursor(0, 0)
	s.Redraw(&feedManager)
	loadFeeds(s, db, *cfg)
	eventLoop(s, *cfg)
	saveFeeds(db)
}
