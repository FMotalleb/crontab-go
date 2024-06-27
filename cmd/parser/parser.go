package parser

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
)

var (
	cfg       = &parserConfig{}
	ParserCmd = &cobra.Command{
		Use:       "parse <crontab file path>",
		ValidArgs: []string{"crontab file path"},
		Short:     "Parse crontab syntax and converts it into yaml syntax for crontab-go",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.cronFile = cmd.Flags().Arg(0)
			log.SetPrefix("[cron-parser]")
			if cfg.cronFile == "" {
				log.Panicln(errors.New("no crontab file specified, see usage using --help flag"))
			}
			cfg.parse()
		},
	}
)

func init() {
	ParserCmd.PersistentFlags().StringVarP(&cfg.output, "output", "o", "", "output file to write configuration to")
	ParserCmd.PersistentFlags().BoolVarP(&cfg.hasUser, "with-user", "u", false, "indicates that whether the given cron file has user field")
	ParserCmd.PersistentFlags().StringVar(&cfg.matcher, "matcher", `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})`, "matcher for cron")
}
