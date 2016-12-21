package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Report indicate each report entity with commit id and comment
type Report struct {
	ID      string
	Comment string
}

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

	standupParts := strings.Split(string(standupOut), "\n")
	commits := []Report{}

	for _, part := range standupParts {
		commitParts := strings.Split(part, " - ")

		if len(commitParts) == 2 {
			commitID := commitParts[0]
			commit := commitParts[1]

			commits = append(commits, Report{commitID, commit})
		}
	}
}
