package main

import (
	b64 "encoding/base64"
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	scribble "github.com/nanobox-io/golang-scribble"
	"gopkg.in/yaml.v2"

	"io/ioutil"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

const CONFIG_DIR = "~/.config/atrss/"

var feeds []*rss.Feed
var curX, curY int
var feedIdx int

type ConfStruct struct {
	Feeds []string `yaml:"feeds"`
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

func loadFeed(s tcell.Screen, db *scribble.Driver, url string) {
	appendFeed(url)
	feed := feeds[len(feeds)-1]
	f := rss.Feed{}
	encoded_url := b64.StdEncoding.EncodeToString([]byte(url))
	err := db.Read("feed", encoded_url, &f)

	if err == nil {
		err := db.Read("feed", encoded_url, &feed)
		check(err)
	} else {
		err := db.Write("feed", encoded_url, feed)
		check(err)
	}
	printLayout(s)
}

func saveFeeds(db *scribble.Driver) {
	for _, f := range feeds {
		encoded_url := b64.StdEncoding.EncodeToString([]byte(f.UpdateURL))
		err := db.Write("feed", encoded_url, f)
		check(err)
	}
}

func Expand(path string) string {
	usr, err := user.Current()
	check(err)

	return strings.Replace(path, "~", usr.HomeDir, 1)
}

func loadFeeds(s tcell.Screen, db *scribble.Driver, cfg ConfStruct) {
	for _, feed := range cfg.Feeds {
		go loadFeed(s, db, feed)
	}
}

func openDB() *scribble.Driver {
	db, err := scribble.New(Expand(CONFIG_DIR), nil)
	check(err)
	return db
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
	db := openDB()
	s := initScreen()
	s.ShowCursor(curX, curY)
	printLayout(s)
	loadFeeds(s, db, cfg)
	eventLoop(s)
	saveFeeds(db)
	deinitScreen(s)
}
