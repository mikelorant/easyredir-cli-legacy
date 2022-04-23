package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
)

var (
	cfgFile string

	limit	int
	startingAfter	string
	endingBefore	string
)

var rootCmd = &cobra.Command{
	Use:   "easyredir-cli",
	Short: "A brief description of your application",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.easyredir.yaml)")
	rootCmd.PersistentFlags().IntVar(&limit, "limit", 25, "pagination limit")
	rootCmd.PersistentFlags().StringVar(&startingAfter, "starting-after", "", "starting after")
	rootCmd.PersistentFlags().StringVar(&endingBefore, "ending-before", "", "ending before")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".easyredir")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n\n", viper.ConfigFileUsed())
	}
}
