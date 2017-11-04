package ui

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	"strconv"
)

var CurX, CurY int
var FeedIdx int

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func InitScreen() tcell.Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	check(err)
	err = s.Init()
	check(err)
	return s
}

func DeinitScreen(s tcell.Screen) {
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

func showFeeds(s tcell.Screen, feeds []*rss.Feed) {
	for n, f := range feeds {
		unread := strconv.FormatUint(uint64(f.Unread), 10)
		str := "(" + unread + ") " + f.Title
		printStr(s, 0, n, str)
	}
}

func showItems(s tcell.Screen, f *rss.Feed) {
	for n, i := range f.Items {
		printStr(s, 40, n, i.Title)
	}
}

func PrintLayout(s tcell.Screen, feeds []*rss.Feed) {
	s.Clear()
	_, h := s.Size()
	printLine(s, 30, 0, 1, h+10)
	showFeeds(s, feeds)
	if CurX == 0 && CurY < len(feeds) {
		showItems(s, feeds[CurY])
	} else if CurX == 40 {
		showItems(s, feeds[FeedIdx])
	}
	s.Show()
}
