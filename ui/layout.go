package ui

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	"fmt"
	"strconv"
)

// FeedIdx is the index of current feed
var FeedIdx int

type Screen struct {
	screen tcell.Screen
	curX   int
	curY   int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// InitScreen initalize the screen
func InitScreen() *Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	tmp, err := tcell.NewScreen()
	check(err)
	err = tmp.Init()
	check(err)
	return &Screen{screen: tmp}
}

// DeinitScreen close the screen
func (s *Screen) DeinitScreen() {
	s.screen.Clear()
	s.screen.Fini()
}

// GetCursor returns the cursor position
func (s *Screen) GetCursor() (x, y int) {
	return s.curX, s.curY
}

// SetCursor sets the cursor to a position
func (s *Screen) SetCursor(x, y int) {
	s.curX = x
	s.curY = y
	s.screen.ShowCursor(s.curX, s.curY)
}

func (s *Screen) printRectangle(x, y int, sx, sy int, c rune) {
	for row := 0; row < sy; row++ {
		for col := 0; col < sx; col++ {
			style := tcell.StyleDefault.Foreground(tcell.ColorRed)
			s.screen.SetCell(x+col, y+row, style, c)
		}
	}
}

func (s *Screen) printLine(x, y int, sx, sy int) {
	s.printRectangle(x, y, sx, sy, '│')
}

func (s *Screen) printStr(x, y int, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.screen.SetContent(x, y, c, comb, tcell.StyleDefault)
		x += w
	}
}

func (s *Screen) showFeeds(feeds []*rss.Feed) {
	for n, f := range feeds {
		unread := strconv.FormatUint(uint64(f.Unread), 10)
		str := "(" + unread + ") " + f.Title
		s.printStr(0, n, str)
	}
}

func (s *Screen) showItems(f *rss.Feed) {
	for n, i := range f.Items {
		s.printStr(40, n, i.Title)
	}
}

func (s *Screen) PollEvent() interface{} {
	return s.screen.PollEvent()
}

// Redraw prints all the user interface elements and contents
func (s *Screen) Redraw(feeds []*rss.Feed) {
	s.screen.Clear()
	_, h := s.screen.Size()
	s.printLine(30, 0, 1, h+10)
	s.showFeeds(feeds)
	if s.curX == 0 && s.curY < len(feeds) {
		s.showItems(feeds[s.curY])
	} else if s.curX == 40 {
		s.showItems(feeds[FeedIdx])
	}
	x, y := s.GetCursor()
	fmt.Println(x, " ", y)
	s.screen.Show()
}
