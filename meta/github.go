package meta

import "fmt"

func Project() string {
	return "https://github.com/FMotalleb/crontab-go"
}

func Issues() string {
	return fmt.Sprintf("%s/issues", Project())
}
