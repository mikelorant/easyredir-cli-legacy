package cmd

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	describeCmd = &cobra.Command{
		Use:   "describe",
		Short: "A brief description of your command",
	}

	describeHostCmd = &cobra.Command{
		Use:   "host [id]",
		Short: "A brief description of your command",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			doDescribeHost(id)
		},
	}
)

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.AddCommand(describeHostCmd)
}

func doDescribeHost(id string) {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	host := easyredir.Host{}
	host.Data.ID = id
	err = c.GetHost(&host)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	host.Print()
}
