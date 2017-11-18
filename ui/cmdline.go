package ui

import (
	"github.com/OSPG/atrss/feed"
	"github.com/gdamore/tcell"

	"strings"
)

var feedManager *feed.Manager
var screen *Screen

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

func showError(msg string) {
	s := screen
	_, h := s.screen.Size()
	msg = "ERR: " + msg
	style := tcell.StyleDefault.Background(tcell.ColorRed).
		Foreground(tcell.ColorBlack)

	s.printStr(0, h-1, msg, style)
	s.SetCursor(0, h)
	s.screen.Show()

	for {
		ev := s.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			return
		}
	}
}

func parseFilterCmd(cmd_args []string) {
	switch cmd_args[0] {
	case "tag":
		filterTag(cmd_args[1])
	case "*":
		unsetFilters()
	default:
		showError("Filter command not found")
	}
}

func parseCmd(cmd string) bool {
	cmd_args := strings.Split(cmd, " ")

	switch cmd_args[0] {
	case "q", "quit":
		return true
	case "filter":
		parseFilterCmd(cmd_args[1:])
	default:
		showError("Command not found")
	}
	return false
}

func handleCommands() bool {
	s := screen
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
					s.printStrDef(1, h-1, command+" ")
					s.SetCursor(len(command)+1, h-1)
					s.screen.Show()
				}
			case tcell.KeyRune:
				command += string(ev.Rune())
				s.printStrDef(1, h-1, command)
				s.SetCursor(len(command)+1, h-1)
				s.screen.Show()
			}
		}
	}
}

func StartCmdLine(s *Screen, f *feed.Manager) bool {
	feedManager = f
	screen = s
	x, y := s.GetCursor()
	defer s.SetCursor(x, y)
	s.ShowCmdLine()
	s.screen.Show()
	return handleCommands()
}
