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
	ID         string
	Author     string
	Date       string
	Comment    string
	Repository string
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

	generateReport(orgName)
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
		} else {
			// should fetch the latest changes
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
		fmt.Printf("Got link: %v %v\n", link[2], link[3])
		url := link[2]
		label := link[3]
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

func generateReport(orgName string) error {
	tmpt, err := ioutil.ReadFile("template.html")
	if err != nil {
		return err
	}
	currentDir, err := os.Getwd()

	if err != nil {
		return err
	}

	// before running standup, change the directory to organization folder
	os.Chdir(orgName)

	standupCmd := exec.Command("git", "standup", "-d", "7", "-a", "all", "-D", "local")
	standupOut, err := standupCmd.Output()

	if err != nil {
		return err
	}

	standupParts := strings.Split(string(standupOut), "\n")
	commits := []Report{}

	// regex to remove terminal color related text
	cleanRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	commitRegex := regexp.MustCompile(`(.+)\s-\s(.+)\s\((.+)\)\s<(.+)>`)
	var repositoryName = ""

	for _, part := range standupParts {
		cleanedPart := cleanRegex.ReplaceAllString(part, "")
		if commitRegex.MatchString(cleanedPart) {
			commitParts := commitRegex.FindStringSubmatch(cleanedPart)
			fmt.Printf("%v\n", commitParts)
			commitID := commitParts[1]
			comment := commitParts[2]
			date := commitParts[3]
			author := commitParts[4]
			commits = append(commits, Report{
				ID:         commitID,
				Author:     author,
				Date:       date,
				Comment:    comment,
				Repository: repositoryName,
			})
		} else {
			repositoryName = strings.Replace(part, currentDir, "", -1)
		}
	}

	f, err := os.Create("standup.html")
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.New("Report template")
	t.Parse(string(tmpt))
	if err := t.Execute(f, commits); err != nil {
		return err
	}

	return nil
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
