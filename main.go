package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"strconv"
)

var feeds []*rss.Feed

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

func eventLoop(s tcell.Screen) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventResize:
			printLayout(s)
		}
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

func main() {
	s := initScreen()
	s.ShowCursor(0, 0)
	appendFeed("http://mumei.space:8020")
	appendFeed("https://news.ycombinator.com/rss")
	showItems(s, feeds[0])
	printLayout(s)
	eventLoop(s)
	deinitScreen(s)
}
