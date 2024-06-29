package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

type CronString struct {
	string
}

func NewCronString(cron string) CronString {
	return CronString{cron}
}

func (s CronString) replaceAll(regex string, repl string) CronString {
	reg := regexp.MustCompile(regex)
	out := reg.ReplaceAllString(s.string, repl)
	return CronString{out}
}

func (s CronString) sanitizeLineBreaker() CronString {
	return s.replaceAll(
		`\s*\\\s*\n\s*([\n|\n\s])*`,
		" ",
	)
}

func (s CronString) sanitizeEmptyLine() CronString {
	return s.replaceAll(
		`\n\s*\n`,
		"\n",
	)
}

func (s CronString) sanitizeComments() CronString {
	return s.replaceAll(
		`\s*#.*`,
		"",
	)
}

func (s CronString) sanitize() CronString {
	return s.
		replaceAll("\r\n", "\n").
		sanitizeComments().
		sanitizeLineBreaker().
		sanitizeEmptyLine()
}

func (s CronString) lines() []string {
	return strings.Split(s.string, "\n")
}

func (s *CronString) parseAsSpec(
	pattern string,
	hasUser bool,
) ([]cronSpec, error) {
	envTable := make(map[string]string)
	specs := make([]cronSpec, 0)
	lines := s.sanitize().lines()
	matcher, parser, err := buildMapper(hasUser, pattern)
	if err != nil {
		return []cronSpec{}, err
	}
	for _, line := range lines {
		l := cronLine{line}
		if env, err := l.exportEnv(); len(env) > 0 {
			if err != nil {
				return nil, err
			}
			for key, val := range env {
				if old, ok := envTable[key]; ok {
					logrus.Warnf("env var of key `%s`, value `%s`, is going to be replaced by `%s`", key, old, val)
				}
				envTable[key] = val
			}
		} else {
			spec, err := l.exportSpec(matcher, envTable, parser)
			if err != nil {
				return nil, err
			}
			if spec != nil {
				specs = append(specs, *spec)
			}
		}
	}
	return specs, nil
}

func (s *CronString) ParseConfig(
	pattern string,
	hasUser bool,
) (*config.Config, error) {
	specs, err := s.parseAsSpec(pattern, hasUser)
	if err != nil {
		return nil, err
	}
	cfg := &config.Config{}
	for _, spec := range specs {
		addSpec(cfg, spec)
	}
	return cfg, nil
}

func buildMapper(hasUser bool, pattern string) (*regexp.Regexp, cronSpecParser, error) {
	lineParser := "(?<cmd>.*)"
	if hasUser {
		lineParser = fmt.Sprintf(`(?<user>\w[\w\d]*)\s+%s`, lineParser)
	}
	cronLineMatcher := fmt.Sprintf(`^(?<cron>%s)\s+%s$`, pattern, lineParser)

	matcher, err := regexp.Compile(cronLineMatcher)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse cron `%s`", matcher)
	}
	var parser cronSpecParser
	if hasUser {
		parser, err = withUserParser(matcher)
	} else {
		parser, err = normalParser(matcher)
	}
	if err != nil {
		return nil, nil, err
	}
	return matcher, parser, nil
}

func addSpec(cfg *config.Config, spec cronSpec) {
	jobName := fmt.Sprintf("FromCron: %s", spec.timing)
	for _, job := range cfg.Jobs {
		if job.Name == jobName {
			task := config.Task{
				Command:  spec.command,
				UserName: spec.user,
				Env:      spec.environ,
			}
			job.Tasks = append(
				job.Tasks,
				task,
			)
			return
		}
	}
	initJob(jobName, spec.timing, cfg)
	addSpec(cfg, spec)
}

func initJob(jobName string, timing string, cfg *config.Config) {
	job := &config.JobConfig{}
	job.Name = jobName
	job.Description = "Imported from cron file"
	job.Disabled = false
	if strings.Contains(timing, "@reboot") {
		job.Events = []config.JobEvent{
			{
				OnInit: true,
			},
		}
	} else {
		job.Events = []config.JobEvent{
			{
				Cron: timing,
			},
		}
	}
	cfg.Jobs = append(cfg.Jobs, job)
}
