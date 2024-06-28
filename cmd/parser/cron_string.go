package parser

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/FMotalleb/crontab-go/config"
)

var envRegex = regexp.MustCompile(`^(?<key>[\w\d_]+)=(?<value>.*)$`)

type CronString struct {
	string
}

func NewCronString(cron string) CronString {
	return CronString{cron}
}

func NewCronFromFile(filePath string) CronString {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0o644)
	if err != nil {
		log.Panicf("can't open cron file: %v", err)
	}
	stat, err := file.Stat()
	if err != nil {
		log.Panicf("can't stat cron file: %v", err)
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		log.Panicf("can't open cron file: %v", err)
	}
	return CronString{string(content)}
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
) []cronSpec {
	envTable := make(map[string]string)
	specs := make([]cronSpec, 0)
	lines := s.sanitize().lines()
	matcher, parser := buildMapper(hasUser, pattern)

	for _, line := range lines {
		l := cronLine{line}
		if env := l.exportEnv(); len(env) > 0 {
			for key, val := range l.exportEnv() {
				if old, ok := envTable[key]; ok {
					log.Printf("env var of key `%s`, value `%s`, is going to be replaced by `%s`\n", key, old, val)
				}
				envTable[key] = val

			}
		} else {
			if spec := l.exportSpec(matcher, envTable, parser); spec != nil {
				specs = append(specs, *spec)
			}
		}
	}
	return specs
}

func (s *CronString) ParseConfig(
	pattern string,
	hasUser bool,
) *config.Config {
	specs := s.parseAsSpec(pattern, hasUser)
	cfg := &config.Config{}
	for _, spec := range specs {
		addSpec(cfg, spec)
	}
	return cfg
}

func buildMapper(hasUser bool, pattern string) (*regexp.Regexp, func([]string, map[string]string) *cronSpec) {
	lineParser := "(?<cmd>.*)"
	if hasUser {
		lineParser = fmt.Sprintf(`(?<user>\w[\w\d]*)\s+%s`, lineParser)
	}
	cronLineMatcher := fmt.Sprintf(`^(?<cron>%s)\s+%s$`, pattern, lineParser)

	matcher, err := regexp.Compile(cronLineMatcher)
	if err != nil {
		log.Panicf("cannot parse cron `%s`", matcher)
	}
	var parser cronSpecParser
	if hasUser {
		parser = withUserParser(matcher)
	} else {
		parser = normalParser(matcher)
	}
	return matcher, parser
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
	job.Events = []config.JobEvent{
		{
			Cron: timing,
		},
	}
	cfg.Jobs = append(cfg.Jobs, job)
}
