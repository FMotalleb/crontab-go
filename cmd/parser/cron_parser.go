package parser

import (
	"github.com/FMotalleb/crontab-go/config"
)

func (cfg parserConfig) parse() config.Config {
	cron := NewCronFromFile(cfg.cronFile)
	return *cron.ParseConfig(
		cfg.cronMatcher,
		cfg.hasUser,
	)
}
