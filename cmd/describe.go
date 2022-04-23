package cmd

import (
  "github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/spf13/cobra"
  "github.com/rs/zerolog/log"
)

var (
  describeID  string

  describeCmd = &cobra.Command{
  	Use:   "describe",
  	Short: "A brief description of your command",
  }

  describeHostCmd = &cobra.Command{
  	Use:   "host",
  	Short: "A brief description of your command",
  	Run: func(cmd *cobra.Command, args []string) {
      doDescribeHost()
  	},
  }
)

func init() {
	rootCmd.AddCommand(describeCmd)
  describeCmd.AddCommand(describeHostCmd)
  describeHostCmd.Flags().StringVarP(&describeID, "id", "i", "", "Host ID")
  describeHostCmd.MarkFlagRequired("id")
}

func doDescribeHost() {
  c, err := easyredir.NewClient()
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }

  host := easyredir.Host{}
  host.Data.ID = describeID
  err = c.GetHost(&host)
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }
  host.Print()
}
