package ui

import (
	"github.com/gdamore/tcell"
)

func processCommand(cmd string) bool {
	switch cmd {
	case "q", "quit":
		return true
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
				return processCommand(command)
			case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
				if len(command) > 0 {
					command = command[:len(command)-1]
					s.printStr(1, h-1, command)
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

func StartCmdLine(s *Screen) bool {
	x, y := s.GetCursor()
	defer s.SetCursor(x, y)
	s.ShowCmdLine()
	s.screen.Show()
	return handleCommands(s)
}
