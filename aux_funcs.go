package main

import (
	"log"
	"os/exec"
	"os/user"
	"strings"
)

func OpenURL(browser, url string) {
	cmd := exec.Command(browser, url)
	err := cmd.Start()
	if err != nil {
		log.Println("Can not execute command: ", err)
	}
}

func Expand(path string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Can not get current user: ", err)
	}

	return strings.Replace(path, "~", usr.HomeDir, 1)
}
