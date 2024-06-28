package parser

import (
	"log"
	"regexp"
	"strings"
)

type cronLine struct {
	string
}

func (l cronLine) exportEnv() map[string]string {
	match := envRegex.FindStringSubmatch(l.string)
	answer := make(map[string]string)
	switch len(match) {
	case 0:
	case 3:
		answer[match[1]] = match[2]
	default:
		log.Panicf("found multiple(%d) env vars in single line\n please attach your crontab file too\n affected line: %s\n parser result: %#v\n", len(match), l.string, match)
	}
	return answer
}

func (l cronLine) exportSpec(regex *regexp.Regexp, env map[string]string, parser cronSpecParser) *cronSpec {
	match := regex.FindStringSubmatch(l.string)
	if len(match) < 1 {
		if len(strings.Trim(l.string, " \n\t")) == 0 {
			return nil
		} else {
			log.Panicf("cannot parse this non-empty line as cron specification: %s", l.string)
		}
	}
	return parser(match, env)
}
