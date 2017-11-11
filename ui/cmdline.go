package ui

import (
	"github.com/OSPG/atrss/feed"
	"github.com/gdamore/tcell"

	"strings"
)

var feedManager *feed.Manager

func filterTag(tag string) {
	for _, f := range feedManager.Feeds {
		if f.HaveTag(tag) {
			f.Visible = true
		} else {
			f.Visible = false
		}
	}
}

func unsetFilters() {
	for _, f := range feedManager.Feeds {
		f.Visible = true
	}
}

func parseFilterCmd(cmd_args []string) {
	switch cmd_args[0] {
	case "tag":
		filterTag(cmd_args[1])
	case "*":
		unsetFilters()
	}
}

func parseCmd(cmd string) bool {
	cmd_args := strings.Split(cmd, " ")

	switch cmd_args[0] {
	case "q", "quit":
		return true
	case "filter":
		parseFilterCmd(cmd_args[1:])
	}
	return false
}

func handleCommands(s *Screen) bool {
	_, h := s.screen.Size()
	command := ""
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return false
			case tcell.KeyEnter:
				return parseCmd(command)
			case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
				if len(command) > 0 {
					command = command[:len(command)-1]
					s.printStr(1, h-1, command+" ")
					s.SetCursor(len(command)+1, h-1)
					s.screen.Show()
				}
			case tcell.KeyRune:
				command += string(ev.Rune())
				s.printStr(1, h-1, command)
				s.SetCursor(len(command)+1, h-1)
				s.screen.Show()
			}
		}
	}
}

func StartCmdLine(s *Screen, f *feed.Manager) bool {
	feedManager = f
	x, y := s.GetCursor()
	defer s.SetCursor(x, y)
	s.ShowCmdLine()
	s.screen.Show()
	return handleCommands(s)
}
