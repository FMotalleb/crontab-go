package cmd

import (
	"fmt"
	"log"
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

func initConfig() {
	godotenv.Load()
	viper.BindEnv(
		"log_timestamp_format",
		"timestamp_format",
	)
	viper.BindEnv(
		"log_format",
		"output_format",
	)
	viper.BindEnv(
		"log_file",
		"output_file",
	)
	viper.BindEnv(
		"log_stdout",
		"print",
	)
	viper.BindEnv(
		"log_level",
		"level",
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

	if err := viper.Unmarshal(CFG); err != nil {
		log.Fatalln("Cannot unmarshal the config file", err)
	}
}
