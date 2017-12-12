package main

import (
	scribble "github.com/nanobox-io/golang-scribble"

	"github.com/OSPG/atrss/backend"
	"github.com/OSPG/atrss/feed"
	"github.com/OSPG/atrss/ui"
	"log"
	"os"
	"time"
)

const CONFIG_DIR = "~/.config/atrss/"

var feedManager feed.Manager

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func loadFeeds(s *ui.Screen, db *scribble.Driver, cfg backend.Config) {
	for _, f := range cfg.Feeds {
		go func(f backend.ConfFeed) {
			loadFeed(db, f, cfg.UpdateStartup)
			s.Redraw(&feedManager)
		}(f)
	}
}

func initLogger(cfg backend.Config) {
	logFile := Expand(cfg.Log_file)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	log.SetOutput(file)
	log.Println("Logger initalized")
}

func updateFeedsLoop(s *ui.Screen, timer *time.Timer) {
	for {
		<-timer.C
		log.Println("Updateing feeds")
		feedManager.Update()
		s.Redraw(&feedManager)
	}
}

func main() {
	cfg := backend.LoadConfig(Expand(CONFIG_DIR))
	initLogger(*cfg)
	db := openDB(CONFIG_DIR)
	s := ui.InitScreen()
	defer s.DeinitScreen()

	s.SetLayout(cfg.Layout)

	s.SetCursor(0, 0)
	s.Redraw(&feedManager)
	loadFeeds(s, db, *cfg)

	if cfg.UpdateInterval != 0 {
		timer := time.NewTimer(time.Minute * cfg.UpdateInterval)
		defer timer.Stop()
		go updateFeedsLoop(s, timer)
	}

	ui.StartEventLoop(s, &feedManager, *cfg)

	saveFeeds(db)
}
