package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FMotalleb/crontab-go/config"
)

var (
	cfgFile string
	CFG     *config.Config = &config.Config{}
)

var rootCmd = &cobra.Command{
	Use:   "crontab-go",
	Short: "Crontab replacement for containers",
	Long: `Cronjob-go is a powerful, lightweight, and highly configurable Golang application
designed to replace the traditional crontab in Docker environments.
With its seamless integration and easy-to-use YAML configuration,
Cronjob-go simplifies the process of scheduling and managing recurring tasks
within your containerized applications.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yaml)")
}

func warnOnErr(err error, message string) {
	if err != nil {
		fmt.Printf(message, err)
	}
}

func panicOnErr(err error, message string) {
	if err != nil {
		fmt.Printf(message, err)
		panic(err)
	}
}

func initConfig() {
	warnOnErr(godotenv.Load(), "Cannot initialize .env file: %s")

	viper.SetDefault("log_timestamp_format", "2006-01-02T15:04:05.000Z")
	warnOnErr(
		viper.BindEnv(
			"log_timestamp_format",
			"timestamp_format",
		),
		"Cannot bind log_timestamp_format env variable: %s",
	)

	viper.SetDefault("log_format", "ansi")
	warnOnErr(
		viper.BindEnv(
			"log_format",
			"output_format",
		),
		"Cannot bind log_format env variable: %s",
	)

	warnOnErr(
		viper.BindEnv(
			"log_file",
			"output_file",
		),
		"Cannot bind log_file env variable: %s",
	)

	viper.SetDefault("log_stdout", true)
	warnOnErr(
		viper.BindEnv(
			"log_stdout",
			"print",
		),
		"Cannot bind log_stdout env variable: %s",
	)

	viper.SetDefault("log_level", "info")
	warnOnErr(
		viper.BindEnv(
			"log_level",
			"level",
		),
		"Cannot bind log_level env variable: %s",
	)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	panicOnErr(
		viper.Unmarshal(CFG),
		"Cannot unmarshal the config file: %s",
	)
	panicOnErr(
		CFG.Validate(),
		"Failed to initialize config file: %s",
	)
}
