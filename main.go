package main

import (
	"github.com/SlyMarbo/rss"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	check(err)
	err = s.Init()
	check(err)

	feed, err := rss.Fetch("http://mumei.space:8020")
	check(err)

	emitStr(s, 10, 2, tcell.StyleDefault, feed.Title)
	s.Show()

loop:
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				break loop
			}
		}
	}

	s.Clear()
	s.Fini()

}
