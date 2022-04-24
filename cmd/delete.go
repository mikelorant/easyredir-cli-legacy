package cmd

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "A brief description of your command",
	}

	deleteRulesCmd = &cobra.Command{
		Use:   "rule [id]",
		Short: "A brief description of your command",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			doDeleteRules(id)
		},
	}
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteRulesCmd)
}

func doDeleteRules(id string) {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	rule := easyredir.Rule{}
	rule.Data.ID = id

	_, err = c.RemoveRule(&rule)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
}
