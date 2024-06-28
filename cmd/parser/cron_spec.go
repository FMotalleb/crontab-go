package parser

import (
	"log"
	"regexp"
)

type (
	cronSpecParser = func([]string, map[string]string) *cronSpec
	cronSpec       struct {
		timing  string
		user    string
		command string
		environ map[string]string
	}
)

func normalParser(regex *regexp.Regexp) cronSpecParser {
	cronIndex := regex.SubexpIndex("cron")
	// userIndex := regex.SubexpIndex("user")
	cmdIndex := regex.SubexpIndex("cmd")
	if cronIndex < 0 || cmdIndex < 0 {
		log.Panicf("cannot find groups (cron,cmd) in regexp: `%v`", regex.SubexpIndex)
	}
	return func(match []string, env map[string]string) *cronSpec {
		return &cronSpec{
			timing:  match[cronIndex],
			user:    "",
			command: match[cmdIndex],
			environ: env,
		}
	}
}

func withUserParser(regex *regexp.Regexp) cronSpecParser {
	cronIndex := regex.SubexpIndex("cron")
	userIndex := regex.SubexpIndex("user")
	cmdIndex := regex.SubexpIndex("cmd")
	if cronIndex < 0 || cmdIndex < 0 || userIndex < 0 {
		log.Panicf("cannot find groups (cron,user,cmd) in regexp: `%v`", regex.SubexpIndex)
	}
	return func(match []string, env map[string]string) *cronSpec {
		return &cronSpec{
			timing:  match[cronIndex],
			user:    match[userIndex],
			command: match[cmdIndex],
			environ: env,
		}
	}
}
