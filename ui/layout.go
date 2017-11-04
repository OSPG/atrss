package ui

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/jaytaylor/html2text"
	"github.com/mattn/go-runewidth"

	"reflect"
	"strconv"
	"strings"
)

// FeedIdx is the index of current feed
var FeedIdx int

type Layout struct {
	columnWidth int
	itemsMargin int
}

type Screen struct {
	layout       Layout
	screen       tcell.Screen
	curX, curY   int
	sizeX, sizeY int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getField(i interface{}, field string) interface{} {
	t := reflect.TypeOf(i)
	count := 0
	for ; count < t.NumField(); count++ {
		if t.Field(count).Name == field {
			break
		}
	}
	//	real_type := t.Field(count).Type
	v := reflect.ValueOf(i)
	return v.Field(count).Interface()
}

// InitScreen initalize the screen
func InitScreen() *Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	check(err)
	err = s.Init()
	check(err)
	x, y := s.Size()

	// Default values
	l := Layout{columnWidth: 30, itemsMargin: 5}

	return &Screen{screen: s, sizeX: x, sizeY: y, layout: l}
}

// DeinitScreen close the screen
func (s *Screen) DeinitScreen() {
	s.screen.Clear()
	s.screen.Fini()
}

func (s *Screen) SetLayout(l interface{}) {
	s.layout.columnWidth = getField(l, "ColumnWidth").(int)
	s.layout.itemsMargin = getField(l, "ItemsMargin").(int)
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

func (s *Screen) printVerticalLine(x, y int, sy int) {
	s.printRectangle(x, y, 1, sy, '│')
}

func (s *Screen) printHorizontalLine(x, y int, sx int) {
	s.printRectangle(x, y, sx, 1, '─')
}

func (s *Screen) printStr(x, y int, str string) {
	if x > s.sizeX || y > s.sizeY {
		return
		panic("Invalid positions")
	}
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
		title := f.Title
		str := "(" + unread + ") " + title

		columnWidth := s.layout.columnWidth
		if len(str) > columnWidth {
			str = str[:columnWidth]
		}
		s.printStr(0, n, str)
	}
}

func (s *Screen) showItems(f *rss.Feed) {
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin
	y := 0
	for _, i := range f.Items {
		if !i.Read {
			s.printStr(cw+im, y, i.Title)
			y++
		}
	}
}

func (s *Screen) showDescription(content string) {
	w, _ := s.screen.Size()
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin
	itemsColumn := cw + im
	w = w - itemsColumn
	x_off := 50
	for _, line := range strings.Split(content, "\n") {
		length := len(line)
		if length > w {
			line_off := 0
			for ; length > w; length -= w {
				n_line := line[line_off : line_off+w]
				line_off += w

				x_off++
				s.printStr(itemsColumn, x_off, n_line)
			}
			n_line := line[line_off:]
			x_off++
			s.printStr(itemsColumn, x_off, n_line)
		} else {
			x_off++
			s.printStr(itemsColumn, x_off, line)
		}
	}
}

func (s *Screen) PollEvent() interface{} {
	return s.screen.PollEvent()
}

// Redraw prints all the user interface elements and contents
func (s *Screen) Redraw(feeds []*rss.Feed) {
	s.screen.Clear()
	w, h := s.screen.Size()
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin

	s.printVerticalLine(cw, 0, h+10)
	s.printHorizontalLine(cw+1, 50, w-cw-1)
	s.showFeeds(feeds)
	if s.curX == 0 && s.curY < len(feeds) {
		s.showItems(feeds[s.curY])
	} else if s.curX == cw+im {
		feed := feeds[FeedIdx]
		s.showItems(feed)
		item := feed.Items[s.curY]
		content, err := html2text.FromString(item.Summary)
		check(err)
		s.showDescription(content)
	}
	s.screen.Show()
}
