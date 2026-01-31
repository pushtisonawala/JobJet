package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	apiURL       string
	outputFormat string
	cfgFile      string
)

// rootCmd is the base command for jobjet CLI
var rootCmd = &cobra.Command{
	Use:   "jobjet",
	Short: "JobJet CLI - Manage background jobs",
	Long:  `JobJet CLI is a tool to submit, monitor, and manage background jobs for the JobJet system.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8000", "JobJet API base URL")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table|json|yaml")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jobjet.yaml)")

	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigName(".jobjet")
			viper.SetConfigType("yaml")
		}
	}
	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // ignore error if config not found
}
