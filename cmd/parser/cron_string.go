package parser

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

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
	sane := s.
		replaceAll("\r\n", "\n").
		sanitizeComments().
		sanitizeLineBreaker().
		sanitizeEmptyLine()
	log.TraceFn(func() []interface{} {
		return []any{
			"sanitizing input:\n",
			s.string,
			"\nOutput:\n",
			sane.string,
		}
	})
	return sane
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
	log.Tracef("parsing lines using `%s` line matcher", matcher.String())
	if err != nil {
		return []cronSpec{}, err
	}
	for num, line := range lines {
		l := cronLine{line}
		if env, err := l.exportEnv(); len(env) > 0 {
			log.Tracef("line %d(post sanitize) is identified as environment line", num)
			if err != nil {
				return nil, err
			}
			for key, val := range env {
				if old, ok := envTable[key]; ok {
					log.Warnf("env var of key `%s`, value `%s`, is going to be replaced by `%s`", key, old, val)
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
		lineParser = "(?<user>\\w[\\w\\d]*)\\s+" + lineParser
	}

	cronLineMatcher := fmt.Sprintf(`^(?<cron>%s)\s+%s$`, pattern, lineParser)

	matcher, err := regexp.Compile(cronLineMatcher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile cron line parser regexp: `%s`", matcher)
	}
	parser, err := getLineParser(hasUser, matcher)
	if err != nil {
		return nil, nil, err
	}
	return matcher, parser, nil
}

func getLineParser(hasUser bool, matcher *regexp.Regexp) (cronSpecParser, error) {
	if hasUser {
		return withUserParser(matcher)
	}
	return normalParser(matcher)
}

func addSpec(cfg *config.Config, spec cronSpec) {
	jobName := "FromCron: " + spec.timing
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
			job.Concurrency++
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
	job.Concurrency = 1
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
