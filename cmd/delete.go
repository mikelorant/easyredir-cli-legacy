package cmd

import (
  "github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/spf13/cobra"
  "github.com/rs/zerolog/log"
)

var (
  deleteID string

  deleteCmd = &cobra.Command{
  	Use:   "delete",
  	Short: "A brief description of your command",
  }

  deleteRulesCmd = &cobra.Command{
  	Use:   "rule",
  	Short: "A brief description of your command",
  	Run: func(cmd *cobra.Command, args []string) {
      doDeleteRules()
  	},
  }
)

func init() {
	rootCmd.AddCommand(deleteCmd)
  deleteCmd.AddCommand(deleteRulesCmd)
  deleteRulesCmd.Flags().StringVarP(&deleteID, "id", "", "", "ID")
  deleteRulesCmd.MarkFlagRequired("id")
}

func doDeleteRules() {
  c, err := easyredir.NewClient()
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }

  rule := easyredir.Rule{}
  rule.Data.ID = deleteID

  _, err = c.RemoveRule(&rule)
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }
}
