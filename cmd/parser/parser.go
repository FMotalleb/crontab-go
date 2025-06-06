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

func run(cmd *cobra.Command, _ []string) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stderr)
	cfg.cronFile = cmd.Flags().Arg(0)

	if trace, err := cmd.Flags().GetBool("verbose"); err == nil && trace {
		log.SetLevel(log.TraceLevel)
	}
	log.Traceln("source file: ", cfg.cronFile)
	cron, err := readInCron(cfg)
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
	fmt.Println("# yaml-language-server: $schema=https://raw.githubusercontent.com/FMotalleb/crontab-go/main/schema.json")
	fmt.Println(result)
	if cfg.output != "" {
		writeOutput(cfg, result)
	}
	log.Println("Done writing output")
	os.Exit(0)
}

func writeOutput(cfg *parserConfig, result string) {
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

func readInCron(cfg *parserConfig) (*CronString, error) {
	if cfg.cronFile == "" {
		return nil, errors.New("please provide a cron file path, usage: `--help`")
	}
	file, err := os.OpenFile(cfg.cronFile, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("can't stat cron file: %w", err)
	}
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, fmt.Errorf("can't open cron file: %w", err)
	}
	str := string(content)
	cron := NewCronString(str)
	return &cron, nil
}

func init() {
	ParserCmd.PersistentFlags().StringVarP(&cfg.output, "output", "o", "", "output file to write configuration to")
	ParserCmd.PersistentFlags().BoolVarP(&cfg.hasUser, "with-user", "u", false, "indicates that whether the given cron file has user field")
	ParserCmd.PersistentFlags().BoolP("verbose", "v", false, "sets the logging level to trace and verbose logging")
	ParserCmd.PersistentFlags().StringVar(&cfg.cronMatcher, "matcher", `(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*|(\*\/\d))\s*){5,7})`, "matcher for cron")
}
