package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
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

// Repository is a DTO objec to represent JSON from Github repository end point
type Repository struct {
	SSHUrl string `json:"ssh_url"`
}

func main() {
	// get environment variable by 12 factor app practice
	orgName := os.Getenv("GITHUB_ORGANIZATION_NAME")
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")

	resp, err := http.Get("https://api.github.com/orgs/" + orgName + "/repos?access_token=" + accessToken)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var repositories []Repository
	err2 := decoder.Decode(&repositories)
	if err2 != nil {
		panic(err)
	}

	for _, repository := range repositories {
		fmt.Printf("Cloning %s\n", repository.SSHUrl)
		re, _ := regexp.Compile(`git@github.com:(\S*).git`)
		folder := re.FindStringSubmatch(repository.SSHUrl)[1]
		exists, _ := folderExists(folder)
		if !exists {
			cloneCmd := exec.Command("git", "clone", repository.SSHUrl, folder)
			err := cloneCmd.Run()
			if err != nil {
				panic(err)
			}
		}
	}
	tmpt, err := ioutil.ReadFile("template.html")

	// before running standup, change the directory to organization folder
	os.Chdir(orgName)

	standupCmd := exec.Command("git", "standup", "-f", "-d", "7")
	standupOut, err := standupCmd.Output()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Got standup report:\n%s\n", standupOut)

	standupParts := strings.Split(string(standupOut), "\n")
	commits := []Report{}

	r, err := regexp.Compile(`\x1b\[[0-9;]*m`)

	if err != nil {
		panic(err)
	}

	for _, part := range standupParts {
		commitParts := strings.Split(part, " - ")

		if len(commitParts) == 2 {
			commitID := r.ReplaceAllString(commitParts[0], "")
			comment := r.ReplaceAllString(commitParts[1], "")

			commits = append(commits, Report{commitID, comment})
		}
	}

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

func folderExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
