package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type cronLine struct {
	string
}

var envRegex = regexp.MustCompile(`^(?<key>[\w\d_]+)=(?<value>.*)$`)

func (l cronLine) exportEnv() (map[string]string, error) {
	match := envRegex.FindStringSubmatch(l.string)
	answer := make(map[string]string)
	switch len(match) {
	case 0:
	case 3:
		answer[match[1]] = match[2]
	default:
		return nil, fmt.Errorf("unexpected response from environment parser for line:\n%s", l.string)
	}
	if len(answer) != 1 && len(strings.Trim(l.string, " \n\t")) != 0 {
		return nil, fmt.Errorf("line cannot be parsed as environment:\n%s", l.string)
	}
	return answer, nil
}

func (l cronLine) exportSpec(regex *regexp.Regexp, env map[string]string, parser cronSpecParser) (*cronSpec, error) {
	match := regex.FindStringSubmatch(l.string)
	if len(match) < 1 {
		if len(strings.Trim(l.string, " \n\t")) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot parse this non-empty line as cron specification: %s", l.string)
	}
	return parser(match, env), nil
}
