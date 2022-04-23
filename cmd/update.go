package cmd

import (
  "github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/spf13/cobra"
  "github.com/rs/zerolog/log"
)

var (
  updateID               string

  // rules
  updateRulesForwardParams 	 bool
  updateRulesForwardPath   	 bool
  updateRulesResponseType  	 string
  updateRulesSourceURLs    	 []string
  updateRulesTargetURL     	 string

  // hosts match options
  updateHostsCaseInsensitive  bool
	updateHostsSlashInsensitive bool

  // hosts not found action
  updateHostsForwardParams bool
  updateHostsForwardPath   bool
  updateHostsCustom404Body       string
  updateHostsResponseCode        int
  updateHostsResponseURL         string

  // hosts security
  updateHostsHTTPSUpgrade            bool
  updateHostsPreventForeignEmbedding bool
  updateHostsHSTSIncludeSubDomains   bool
  updateHostsHSTSMaxAge              int
  updateHostsHSTSPreload             bool

  updateCmd = &cobra.Command{
  	Use:   "update",
  	Short: "A brief description of your command",
  }

  updateRulesCmd = &cobra.Command{
  	Use:   "rule",
  	Short: "A brief description of your command",
  	Run: func(cmd *cobra.Command, args []string) {
      doUpdateRules()
  	},
  }

  updateHostsCmd = &cobra.Command{
  	Use:   "host",
  	Short: "A brief description of your command",
  	Run: func(cmd *cobra.Command, args []string) {
      doUpdateHosts()
  	},
  }
)

func init() {
	rootCmd.AddCommand(updateCmd)
  updateCmd.AddCommand(updateRulesCmd)
  updateRulesCmd.Flags().StringVarP(&updateID, "id", "", "", "ID")
  updateRulesCmd.Flags().BoolVarP(&updateRulesForwardParams, "forward-params", "", true, "Forward params")
  updateRulesCmd.Flags().BoolVarP(&updateRulesForwardPath, "forward-path", "", true, "Forward path")
  updateRulesCmd.Flags().StringVarP(&updateRulesResponseType, "response-type", "", "moved_permanently", "Response type")
  updateRulesCmd.Flags().StringSliceVarP(&updateRulesSourceURLs, "source-urls", "", nil, "Source URLs")
  updateRulesCmd.Flags().StringVarP(&updateRulesTargetURL, "target-url", "", "", "Target URL")
  updateRulesCmd.MarkFlagRequired("id")

  updateCmd.AddCommand(updateHostsCmd)
  updateHostsCmd.Flags().StringVarP(&updateID, "id", "", "", "Forward params")
  updateHostsCmd.Flags().BoolVarP(&updateHostsCaseInsensitive, "case-insensitive", "", true, "Case insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsSlashInsensitive, "slash-insensitive", "", true, "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsForwardParams, "forward-params", "", true, "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsForwardPath, "forward-path", "", true, "Slash insensitive")
  updateHostsCmd.Flags().StringVarP(&updateHostsCustom404Body, "custom-404-body", "", "", "Slash insensitive")
  updateHostsCmd.Flags().IntVarP(&updateHostsResponseCode, "response-code", "", 404, "Slash insensitive")
  updateHostsCmd.Flags().StringVarP(&updateHostsResponseURL, "response-url", "", "", "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsHTTPSUpgrade, "https-upgrade", "", true, "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsPreventForeignEmbedding, "prevent-foreign-embedding", "", true, "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsHSTSIncludeSubDomains, "hsts-include-sub-domains", "", true, "Slash insensitive")
  updateHostsCmd.Flags().IntVarP(&updateHostsHSTSMaxAge, "hsts-max-age", "", 0, "Slash insensitive")
  updateHostsCmd.Flags().BoolVarP(&updateHostsHSTSPreload, "hsts-preload", "", true, "Slash insensitive")
  updateHostsCmd.MarkFlagRequired("id")
}

func doUpdateRules() {
  c, err := easyredir.NewClient()
  if err != nil {
    log.Error().Err(err).Msg("")
  }

  rule := easyredir.Rule{}
  rule.Data.ID = updateID
  rule.Data.Attributes.ForwardParams = updateRulesForwardParams
  rule.Data.Attributes.ForwardPath = updateRulesForwardPath
  rule.Data.Attributes.ResponseType = updateRulesResponseType
  rule.Data.Attributes.SourceUrls = updateRulesSourceURLs
  rule.Data.Attributes.TargetURL = updateRulesTargetURL

  res, err := c.UpdateRule(&rule)
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }

  res.Print()
}

func doUpdateHosts() {
  c, err := easyredir.NewClient()
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }

  host := easyredir.Host{}
  host.Data.ID = updateID
  host.Data.Attributes.MatchOptions.CaseInsensitive = updateHostsCaseInsensitive
  host.Data.Attributes.MatchOptions.SlashInsensitive = updateHostsSlashInsensitive
  host.Data.Attributes.NotFoundAction.ForwardParams = updateHostsForwardParams
  host.Data.Attributes.NotFoundAction.ForwardPath = updateHostsForwardPath
  host.Data.Attributes.NotFoundAction.Custom404Body = updateHostsCustom404Body
  host.Data.Attributes.NotFoundAction.ResponseCode = updateHostsResponseCode
  host.Data.Attributes.NotFoundAction.ResponseURL = updateHostsResponseURL
  host.Data.Attributes.Security.HTTPSUpgrade = updateHostsHTTPSUpgrade
  host.Data.Attributes.Security.PreventForeignEmbedding = updateHostsPreventForeignEmbedding
  host.Data.Attributes.Security.HstsIncludeSubDomains = updateHostsHSTSIncludeSubDomains
  host.Data.Attributes.Security.HstsMaxAge = updateHostsHSTSMaxAge
  host.Data.Attributes.Security.HstsPreload = updateHostsHSTSPreload

  res, err := c.UpdateHost(&host)
  if err != nil {
    log.Error().Err(err).Msg("")
    return
  }

  res.Print()
}
