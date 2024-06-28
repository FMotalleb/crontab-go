// Package parser manages holds the logic behind the sub command `parse`
// this package is responsible for parsing a crontab file into valid config yaml file
package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfg       = &parserConfig{}
	ParserCmd = &cobra.Command{
		Use:       "parse <crontab file path>",
		ValidArgs: []string{"crontab file path"},
		Short:     "Parse crontab syntax and converts it into yaml syntax for crontab-go",
		Run:       run,
	}
)

func run(cmd *cobra.Command, args []string) {
	cfg.cronFile = cmd.Flags().Arg(0)

	cron, err := readInCron()
	if err != nil {
		log.Panic(err)
	}
	finalConfig, err := cron.ParseConfig(
		cfg.cronMatcher,
		cfg.hasUser,
	)
	if err != nil {
		log.Panicf("cannot parse given cron file: %v", err)
	}
	result, err := generateYamlFromCfg(finalConfig)
	if err != nil {
		log.Panic(err)
	}
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
}

func readInCron() (*CronString, error) {
	var str string = ""
	if cfg.cronFile == "" {
		return nil, errors.New("please provide a cron file path, usage: `--help`")
	}
	file, err := os.OpenFile(cfg.cronFile, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %v", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("can't stat cron file: %v", err)
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %v", err)
	}
	str = string(content)
	cron := NewCronString(str)
	return &cron, nil
}

func init() {
	ParserCmd.PersistentFlags().StringVarP(&cfg.output, "output", "o", "", "output file to write configuration to")
	ParserCmd.PersistentFlags().BoolVarP(&cfg.hasUser, "with-user", "u", false, "indicates that whether the given cron file has user field")
	ParserCmd.PersistentFlags().StringVar(&cfg.cronMatcher, "matcher", `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*|(\*\/\d)) ?){5,7})`, "matcher for cron")
}
