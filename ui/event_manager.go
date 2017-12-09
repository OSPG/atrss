package ui

import (
	"log"
	"os/exec"

	"github.com/gdamore/tcell"

	"github.com/OSPG/atrss/backend"
	"github.com/OSPG/atrss/feed"
)

func openURL(browser, url string) {
	cmd := exec.Command(browser, url)
	err := cmd.Start()
	if err != nil {
		log.Println("Can not execute command: ", err)
	}
}

func StartEventLoop(s *Screen, feedManager *feed.Manager, cfg backend.Config) {
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
					f := feedManager.Get(FeedIdx)
					item := f.GetUnreadItem(y)
					openURL(cfg.Browser, item.Link)
					f.ReadItem(item)
				}
			case tcell.KeyDown:
				_, h := s.GetSize()
				if x == s.ItemsColumn {
					f := feedManager.Get(FeedIdx)
					if f.Unread == 0 {
						y = FeedIdx
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
					FeedIdx = y
					s.SetCursor(s.ItemsColumn, 0)
				}
			case tcell.KeyLeft:
				_, y := s.GetCursor()
				y = FeedIdx
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
					f := feedManager.Get(FeedIdx)

					if f.Unread-1 > uint32(cfg.Layout.BoxHeigh) {
						y = cfg.Layout.BoxHeigh - 1
					} else {
						y = int(f.Unread - 1)
					}

				} else {
					x = 0
					elements := feedManager.Len()
					if elements > y {
						_, h := s.GetSize()
						y = h - 1
					} else {
						y = elements - 1
					}
				}
				s.SetCursor(x, y)
			}
			switch ev.Rune() {
			case ' ':
				if x == s.ItemsColumn {
					f := feedManager.Get(FeedIdx)
					if f.Unread != 0 {
						item := f.GetUnreadItem(y)
						f.ReadItem(item)
					}
				}
			case 'o', 'O':
				if x == s.ItemsColumn {
					f := feedManager.Get(FeedIdx)
					item := f.GetUnreadItem(y)
					openURL(cfg.Browser, item.Link)
				}
			case 'R':
				f := feedManager.Get(FeedIdx)
				f.Update()
			case ':':
				if StartCmdLine(s, feedManager) {
					return
				}

			}
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		s.Redraw(feedManager)
	}
}
