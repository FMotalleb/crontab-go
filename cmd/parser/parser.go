package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	cfg       = &parserConfig{}
	ParserCmd = &cobra.Command{
		Use:       "parse <crontab file path>",
		ValidArgs: []string{"crontab file path"},
		Short:     "Parse crontab syntax and converts it into yaml syntax for crontab-go",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.cronFile = cmd.Flags().Arg(0)
			log.SetPrefix("[Cron Parser]")
			if cfg.cronFile == "" {
				log.Panicln(errors.New("no crontab file specified, see usage using --help flag"))
			}
			finalConfig := cfg.parse()
			str, err := json.Marshal(finalConfig)
			if err != nil {
				log.Panicf("failed to marshal final config: %v", err)
			}
			hashMap := make(map[string]any)
			json.Unmarshal(str, &hashMap)
			ans, _ := yaml.Marshal(hashMap)
			result := string(ans)
			log.Printf("output:\n%s", result)
			if cfg.output != "" {
				outputFile, err := os.OpenFile(cfg.output, os.O_WRONLY|os.O_CREATE, 0o644)
				if err != nil {
					log.Panicf("failed to open output file: %v", err)
				}
				buf := bytes.NewBufferString(result)
				_, err = io.Copy(outputFile, buf)
				if err != nil {
					log.Panicf("failed to write output file: %v", err)
				}
			}
			log.Println("Done writing output")
			os.Exit(0)
		},
	}
)

func init() {
	ParserCmd.PersistentFlags().StringVarP(&cfg.output, "output", "o", "", "output file to write configuration to")
	ParserCmd.PersistentFlags().BoolVarP(&cfg.hasUser, "with-user", "u", false, "indicates that whether the given cron file has user field")
	ParserCmd.PersistentFlags().StringVar(&cfg.cronMatcher, "matcher", `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*|(\*\/\d)) ?){5,7})`, "matcher for cron")
}
