package main

import (
	"os/exec"
	"os/user"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func OpenURL(url string) {
	cmd := exec.Command("firefox", url)
	err := cmd.Start()
	check(err)
}

func Expand(path string) string {
	usr, err := user.Current()
	check(err)

	return strings.Replace(path, "~", usr.HomeDir, 1)
}
