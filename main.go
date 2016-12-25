package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

// Report indicate each report entity with commit id and comment
type Report struct {
	ID      string
	Author  string
	Date    string
	Comment string
}

// Repository is a DTO objec to represent JSON from Github repository end point
type Repository struct {
	SSHUrl string `json:"ssh_url"`
}

func handleError(err error) {
	fmt.Printf("Facing error: %v\n", err)
	panic(err)
}

func main() {
	// get environment variable by 12 factor app practice
	orgName := os.Getenv("GITHUB_ORGANIZATION_NAME")
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")

	err := cloneAllOrganizationRepositories(orgName, accessToken)
	if err != nil {
		handleError(err)
	}
}

func cloneAllOrganizationRepositories(orgName, accessToken string) error {
	repositories, err := getAllRepositories("https://api.github.com/orgs/" + orgName + "/repos?access_token=" + accessToken)
	if err != nil {
		return err
	}

	// for each repository, clone them down
	// TOOD: try to use goroutine here to speed up the clone process
	for _, repository := range repositories {
		fmt.Printf("Cloning %s\n", repository.SSHUrl)
		re := regexp.MustCompile(`git@github.com:(\S*).git`)
		folder := re.FindStringSubmatch(repository.SSHUrl)[1]
		exists, _ := folderExists(folder)
		if !exists {
			cloneCmd := exec.Command("git", "clone", repository.SSHUrl, folder)
			err := cloneCmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getAllRepositories(githubAPIURL string) ([]Repository, error) {
	repositories := []Repository{}

	resp, err := http.Get(githubAPIURL)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	err2 := decoder.Decode(&repositories)
	if err2 != nil {
		return nil, err2
	}

	// check if there is next page
	r := regexp.MustCompile(`(<(\S+)>;\srel="(\S+)",*)+`)
	links := r.FindAllStringSubmatch(resp.Header.Get("Link"), -1)

	for _, link := range links {
		url := link[1]
		label := link[2]
		if label == "next" {
			nextPageRepositories, err := getAllRepositories(url)
			if err != nil {
				return nil, err
			}
			return append(repositories, nextPageRepositories...), nil
		}
	}

	return repositories, nil
}

func generateReport(orgName string) {
	/*
			tmpt, err := ioutil.ReadFile("template.html")
			if err != nil {
				handleError(err)
			}
			// before running standup, change the directory to organization folder
			os.Chdir(orgName)

			standupCmd := exec.Command("git", "standup", "-f", "-d", "7")
			standupOut, err := standupCmd.Output()

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
	*/
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
