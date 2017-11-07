package ui

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/jaytaylor/html2text"
	"github.com/mattn/go-runewidth"

	"github.com/OSPG/atrss/feed"
)

// FeedIdx is the index of current feed
var FeedIdx int

type Layout struct {
	columnWidth int
	itemsMargin int
	boxHeigh    int
}

type Screen struct {
	layout       Layout
	screen       tcell.Screen
	curX, curY   int
	sizeX, sizeY int
	ItemsColumn  int
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
	if err != nil {
		log.Fatalln("Could not create screen: ", err)
	}

	err = s.Init()
	if err != nil {
		log.Fatalln("Could not initalize screen: ", err)
	}
	x, y := s.Size()

	// Default values
	cw := 30
	im := 5
	l := Layout{columnWidth: cw, itemsMargin: im}

	return &Screen{screen: s, sizeX: x, sizeY: y, layout: l, ItemsColumn: cw + im}
}

// DeinitScreen close the screen
func (s *Screen) DeinitScreen() {
	s.screen.Clear()
	s.screen.Fini()
}

func (s *Screen) SetLayout(l interface{}) {
	s.layout.columnWidth = getField(l, "ColumnWidth").(int)
	s.layout.itemsMargin = getField(l, "ItemsMargin").(int)
	s.layout.boxHeigh = getField(l, "BoxHeigh").(int)
	s.ItemsColumn = s.layout.columnWidth + s.layout.itemsMargin
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
		log.Printf("WARNING: Invalid positions %d %d. Max: %d %d\n", x, y, s.sizeX, s.sizeY)
		return
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

func (s *Screen) showFeeds(feeds []*feed.Feed) {
	for n, f := range feeds {
		unread := strconv.FormatUint(uint64(f.Unread), 10)
		title := f.Feed.Title
		str := "(" + unread + ") " + title

		columnWidth := s.layout.columnWidth
		if len(str) > columnWidth {
			str = str[:columnWidth]
		}
		s.printStr(0, n, str)
	}
}

func (s *Screen) showItems(f *feed.Feed) {
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin
	y := 0
	for _, i := range f.Feed.Items {
		if !i.Read {
			s.printStr(cw+im, y, i.Title)
			y++
		}

		// We don't have enough space to show more items
		if y == s.layout.boxHeigh {
			return
		}
	}
}

func (s *Screen) showDescription(content string) {
	w, h := s.screen.Size()
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin
	itemsColumn := cw + im
	w = w - itemsColumn
	x_off := s.layout.boxHeigh
	for _, line := range strings.Split(content, "\n") {
		if x_off >= h {
			return
		}
		length := len(line)
		line_off := 0
		for ; x_off < h && length > w; length -= w {
			n_line := line[line_off : line_off+w]
			line_off += w

			x_off++
			s.printStr(itemsColumn, x_off, n_line)
		}
		if x_off >= h {
			return
		}
		n_line := line[line_off:]
		x_off++
		s.printStr(itemsColumn, x_off, n_line)
	}
}

func (s *Screen) ShowCmdLine() {
	_, h := s.screen.Size()
	s.SetCursor(1, h-1)
	s.printStr(0, h-1, ":")
}

func (s *Screen) PollEvent() interface{} {
	return s.screen.PollEvent()
}

// Redraw prints all the user interface elements and contents
func (s *Screen) Redraw(feedManager *feed.Manager) {
	s.screen.Clear()
	w, h := s.screen.Size()
	cw := s.layout.columnWidth
	im := s.layout.itemsMargin
	bh := s.layout.boxHeigh

	s.printVerticalLine(cw, 0, h+10)
	s.printHorizontalLine(cw+1, bh, w-cw-1)
	s.showFeeds(feedManager.Feeds)
	if s.curX == 0 && s.curY < feedManager.Len() {
		feed := feedManager.Get(s.curY)
		s.showItems(feed)
	} else if s.curX == cw+im {
		feed := feedManager.Get(FeedIdx)
		s.showItems(feed)
		item := feed.GetItem(s.curY)
		content, err := html2text.FromString(item.Summary)
		if err != nil {
			log.Println("Could not show item summary: ", err)
		}

		s.showDescription(content)
	}
	s.screen.Show()
}
