package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	name := flag.String("name", "", "Name of the git author")

	flag.Parse()

	if *name == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Hello, %s are you ready to do standup report based on git history?\n", *name)

	standupCmd := exec.Command("git", "standup")

	standupOut, err := standupCmd.Output()

	if err != nil {
		panic(err)
	}
	fmt.Printf("standup output:\n%s\n", standupOut)
}
