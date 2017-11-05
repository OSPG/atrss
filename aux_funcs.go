package main

import (
	"log"
	"os/exec"
	"os/user"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func OpenURL(browser, url string) {
	log.Println(browser)
	cmd := exec.Command(browser, url)
	err := cmd.Start()
	check(err)
}

func Expand(path string) string {
	usr, err := user.Current()
	check(err)

	return strings.Replace(path, "~", usr.HomeDir, 1)
}
