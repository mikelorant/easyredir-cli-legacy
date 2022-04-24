package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultCaseInsensitive         bool   = false // verified
	defaultCustom404Body           string = ""    // verified
	defaultForwardParams           bool   = true
	defaultForwardPath             bool   = true
	defaultHSTSIncludeSubDomains   bool   = false // verified
	defaultHSTSMaxAge              int    = 0     // verified
	defaultHSTSPreload             bool   = false // verified
	defaultHTTPSUpgrade            bool   = false // verified
	defaultID                      string = ""    // verified
	defaultPreventForeignEmbedding bool   = false
	defaultResponseCode            int    = 301
	defaultResponseType            string = "moved_permanently"
	defaultResponseURL             string = ""
	defaultSlashInsensitive        bool   = false // verified
	defaultTargetURL               string = ""
)

var (
	cfgFile string

	startingAfter string
	endingBefore  string

	flagsChanged []string

	defaultSourceURLS []string = []string{}
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
	viper.ReadInConfig()
}
