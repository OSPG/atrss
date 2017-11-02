package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mattn/go-runewidth"
	"os/exec"
	"strconv"
	"time"
)

var feeds []*rss.Feed
var curX, curY int
var feedIdx int

type Feed struct {
	Nickname    string // This is not set by the package, but could be helpful.
	Title       string
	Description string
	Link        string // Link to the creator's website.
	UpdateURL   string `gorm:"primary_key"` // URL of the feed itself.
	Items       []*rss.Item
	Refresh     time.Time // Earliest time this feed should next be checked.
	Unread      uint32    // Number of unread items. Used by aggregators.
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func appendFeed(url string) {
	feed, err := rss.Fetch(url)
	check(err)
	feeds = append(feeds, feed)
}

func initScreen() tcell.Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	check(err)
	err = s.Init()
	check(err)
	return s
}

func deinitScreen(s tcell.Screen) {
	s.Clear()
	s.Fini()
}

func printRectangle(s tcell.Screen, x, y int, sx, sy int, c rune) {
	for row := 0; row < sy; row++ {
		for col := 0; col < sx; col++ {
			s.SetCell(x+col, y+row, tcell.StyleDefault.Foreground(tcell.ColorRed), c)
		}
	}
}

func printLine(s tcell.Screen, x, y int, sx, sy int) {
	printRectangle(s, x, y, sx, sy, '│')
}

func openURL(url string) {
	cmd := exec.Command("firefox", url)
	err := cmd.Start()
	check(err)
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
					openURL(item.Title)
					if !item.Read {
						item.Read = true
						feed.Unread--
					}
				}
			case tcell.KeyDown:
				if curX == 40 {
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
			//		case *tcell.EventResize:
			//			printLayout(s)
		}
		if curX == 0 && curY < len(feeds) {
			showItems(s, feeds[curY])
		}
		printLayout(s)
	}
}

func printStr(s tcell.Screen, x, y int, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, tcell.StyleDefault)
		x += w
	}
}

func showFeeds(s tcell.Screen) {
	for n, f := range feeds {
		unread := strconv.FormatUint(uint64(f.Unread), 10)
		str := "(" + unread + ") " + f.Title
		printStr(s, 0, n, str)
	}

	s.Show()
}

func showItems(s tcell.Screen, f *rss.Feed) {
	s.Clear()
	printLayout(s)
	for n, i := range f.Items {
		printStr(s, 40, n, i.Title)
	}
	s.Show()
}

func printLayout(s tcell.Screen) {
	_, h := s.Size()
	printLine(s, 30, 0, 1, h+10)
	showFeeds(s)
}

func loadFeed(s tcell.Screen, db *gorm.DB, url string) {
	appendFeed(url)
	var feed Feed
	db.First(&feed, "update_url = ?", url)
	if feed.UpdateURL != url {
		f := copyFeed(feeds[len(feeds)-1])
		db.Create(&f)
	}
	printLayout(s)
}

func loadFeeds(s tcell.Screen, db *gorm.DB) {
	go loadFeed(s, db, "https://news.ycombinator.com/rss")
	go loadFeed(s, db, "http://mumei.space:8020")
}

func openDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	check(err)
	return db
}

func copyFeed(srcFeed *rss.Feed) Feed {
	var dstFeed Feed
	dstFeed.Title = srcFeed.Title
	dstFeed.Description = srcFeed.Description
	dstFeed.Link = srcFeed.Link
	dstFeed.Nickname = srcFeed.Nickname
	dstFeed.Refresh = srcFeed.Refresh
	dstFeed.Unread = srcFeed.Unread
	dstFeed.UpdateURL = srcFeed.UpdateURL
	dstFeed.Items = srcFeed.Items
	return dstFeed
}

func main() {
	db := openDB()
	db.AutoMigrate(&Feed{})
	s := initScreen()
	s.ShowCursor(curX, curY)
	printLayout(s)
	loadFeeds(s, db)
	eventLoop(s)
	db.Close()
	deinitScreen(s)
}
