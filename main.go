package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	name := flag.String("name", "", "Name of the git author")

	flag.Parse()

	if *name == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Hello, %v are you ready to do standup report based on git history?", name)
}
