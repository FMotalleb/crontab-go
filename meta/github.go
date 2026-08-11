// Package meta contains meta data information about this program.
package meta

import "fmt"

const (
	GitHubUser    = "fmotalleb"
	GitHubProject = "crontab-go"
)

func Project() string {
	return fmt.Sprintf("https://github.com/%s/%s", GitHubUser, GitHubProject)
}

func Issues() string {
	return Project() + "/issues"
}
