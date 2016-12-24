package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
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

	if *name != "" {
		fmt.Printf("Got %s for stand up report. Generating report ...\n", *name)
	} else {
		fmt.Println("Generating report for default git user")
	}

	standupCmd := exec.Command("git", "standup")
	standupOut, err := standupCmd.Output()

	if err != nil {
		panic(err)
	}

	standupParts := strings.Split(string(standupOut), "\n")
	commits := []Report{}

	r, _ := regexp.Compile(`\x1b\[[0-9;]*m`)

	for _, part := range standupParts {
		commitParts := strings.Split(part, " - ")

		if len(commitParts) == 2 {
			commitID := r.ReplaceAllString(commitParts[0], "")
			comment := r.ReplaceAllString(commitParts[1], "")

			commits = append(commits, Report{commitID, comment})
		}
	}

	tmpt, err := ioutil.ReadFile("template.html")

	if err != nil {
		panic(err)
	}

	f, err := os.Create("standup.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t := template.New("Report template")
	t.Parse(string(tmpt))
	if err := t.Execute(f, commits); err != nil {
		panic(err)
	}
}
