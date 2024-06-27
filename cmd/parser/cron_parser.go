package parser

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/FMotalleb/crontab-go/config"
)

type cronString struct {
	string
}

func (c cronString) sanitizeLineBreaker() cronString {
	reg, _ := regexp.Compile(`\s*\\\s*[\r\n|\n]\s*([\r\n|\n\s])*`)
	out := reg.ReplaceAllString(c.string, " ")
	return cronString{out}
}

func (c cronString) sanitizeComments() cronString {
	reg, _ := regexp.Compile(`\s*#.*`)
	out := reg.ReplaceAllString(c.string, "")
	return cronString{out}
}

func (cfg parserConfig) parse() config.Config {
	file, err := os.OpenFile(cfg.cronFile, os.O_RDONLY, 0o644)
	if err != nil {
		log.Panicf("can't open cron file: %v", err)
	}
	stat, err := file.Stat()
	content := make([]byte, stat.Size())
	file.Read(content)
	strContent := cronString{string(content)}
	strContent = strContent.
		sanitizeComments()
	fmt.Println(
		strContent.sanitizeComments(),
		strContent.sanitizeComments().sanitizeLineBreaker().sanitizeLineBreaker(),
	)

	c := config.Config{}
	return c
}
