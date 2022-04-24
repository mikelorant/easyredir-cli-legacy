package cmd

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	updateID string

	updateRulesForwardParams           bool
	updateRulesForwardPath             bool
	updateRulesResponseType            string
	updateRulesSourceURLs              []string
	updateRulesTargetURL               string
	updateHostsCaseInsensitive         bool
	updateHostsSlashInsensitive        bool
	updateHostsForwardParams           bool
	updateHostsForwardPath             bool
	updateHostsCustom404Body           string
	updateHostsResponseCode            int
	updateHostsResponseURL             string
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
			flagsChanged = getFlagsChanged(cmd)
			doUpdateRules()
		},
	}

	updateHostsCmd = &cobra.Command{
		Use:   "host",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			flagsChanged = getFlagsChanged(cmd)
			doUpdateHosts()
		},
	}
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateRulesCmd)

	updateRulesCmd.Flags().StringVarP(&updateID, "id", "i", defaultID, "ID")
	updateRulesCmd.Flags().BoolVarP(&updateRulesForwardParams, "forward-params", "", defaultForwardParams, "Forward params")
	updateRulesCmd.Flags().BoolVarP(&updateRulesForwardPath, "forward-path", "", defaultForwardPath, "Forward path")
	updateRulesCmd.Flags().StringVarP(&updateRulesResponseType, "response-type", "", defaultResponseType, "Response type")
	updateRulesCmd.Flags().StringSliceVarP(&updateRulesSourceURLs, "source-urls", "", defaultSourceURLS, "Source URLs")
	updateRulesCmd.Flags().StringVarP(&updateRulesTargetURL, "target-url", "", defaultTargetURL, "Target URL")
	updateRulesCmd.MarkFlagRequired("id")

	updateCmd.AddCommand(updateHostsCmd)
	updateHostsCmd.Flags().StringVarP(&updateID, "id", "i", defaultID, "Forward params")
	updateHostsCmd.Flags().BoolVarP(&updateHostsCaseInsensitive, "case-insensitive", "", defaultCaseInsensitive, "Case insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsSlashInsensitive, "slash-insensitive", "", defaultSlashInsensitive, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsForwardParams, "forward-params", "", defaultForwardParams, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsForwardPath, "forward-path", "", defaultForwardPath, "Slash insensitive")
	updateHostsCmd.Flags().StringVarP(&updateHostsCustom404Body, "custom-404-body", "", defaultCustom404Body, "Slash insensitive")
	updateHostsCmd.Flags().IntVarP(&updateHostsResponseCode, "response-code", "", defaultResponseCode, "Slash insensitive")
	updateHostsCmd.Flags().StringVarP(&updateHostsResponseURL, "response-url", "", defaultResponseURL, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsHTTPSUpgrade, "https-upgrade", "", defaultHTTPSUpgrade, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsPreventForeignEmbedding, "prevent-foreign-embedding", "", defaultPreventForeignEmbedding, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsHSTSIncludeSubDomains, "hsts-include-sub-domains", "", defaultHSTSIncludeSubDomains, "Slash insensitive")
	updateHostsCmd.Flags().IntVarP(&updateHostsHSTSMaxAge, "hsts-max-age", "", defaultHSTSMaxAge, "Slash insensitive")
	updateHostsCmd.Flags().BoolVarP(&updateHostsHSTSPreload, "hsts-preload", "", defaultHSTSPreload, "Slash insensitive")
	updateHostsCmd.MarkFlagRequired("id")
}

func doUpdateRules() {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	rule := easyredir.Rule{}
	rule.Data.ID = updateID

	rules, err := c.ListRules(&easyredir.RulesOptions{
		Limit: 1,
	})
	if err != nil {
		log.Error().Err(err).Msg("Unable to list rules.")
	}

	for _, r := range rules.Data {
		if r.ID == rule.Data.ID {
			rule.Data.Attributes.ForwardParams = r.Attributes.ForwardParams
			rule.Data.Attributes.ForwardPath = r.Attributes.ForwardPath
			rule.Data.Attributes.ResponseType = r.Attributes.ResponseType
			rule.Data.Attributes.SourceUrls = r.Attributes.SourceURLs
			rule.Data.Attributes.TargetURL = r.Attributes.TargetURL
			break
		}
	}

	if flagIn("forward-params", flagsChanged) {
		rule.Data.Attributes.ForwardParams = updateRulesForwardParams
	}
	if flagIn("forward-path", flagsChanged) {
		rule.Data.Attributes.ForwardPath = updateRulesForwardPath
	}
	if flagIn("response-type", flagsChanged) {
		rule.Data.Attributes.ResponseType = updateRulesResponseType
	}
	if flagIn("source-urls", flagsChanged) {
		rule.Data.Attributes.SourceUrls = updateRulesSourceURLs
	}
	if flagIn("target-url", flagsChanged) {
		rule.Data.Attributes.TargetURL = updateRulesTargetURL
	}

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

	c.GetHost(&host)

	if flagIn("case-insensitive", flagsChanged) {
		host.Data.Attributes.MatchOptions.CaseInsensitive = updateHostsCaseInsensitive
	}
	if flagIn("slash-insensitive", flagsChanged) {
		host.Data.Attributes.MatchOptions.SlashInsensitive = updateHostsSlashInsensitive
	}
	if flagIn("forward-params", flagsChanged) {
		host.Data.Attributes.NotFoundAction.ForwardParams = updateHostsForwardParams
	}
	if flagIn("forward-path", flagsChanged) {
		host.Data.Attributes.NotFoundAction.ForwardPath = updateHostsForwardPath
	}
	if flagIn("custom-404-body", flagsChanged) {
		host.Data.Attributes.NotFoundAction.Custom404Body = updateHostsCustom404Body
	}
	if flagIn("response-code", flagsChanged) {
		host.Data.Attributes.NotFoundAction.ResponseCode = updateHostsResponseCode
	}
	if flagIn("response-url", flagsChanged) {
		host.Data.Attributes.NotFoundAction.ResponseURL = updateHostsResponseURL
	}
	if flagIn("https-upgrade", flagsChanged) {
		host.Data.Attributes.Security.HTTPSUpgrade = updateHostsHTTPSUpgrade
	}
	if flagIn("prevent-foreign-embedding", flagsChanged) {
		host.Data.Attributes.Security.PreventForeignEmbedding = updateHostsPreventForeignEmbedding
	}
	if flagIn("hsts-include-sub-domains", flagsChanged) {
		host.Data.Attributes.Security.HstsIncludeSubDomains = updateHostsHSTSIncludeSubDomains
	}
	if flagIn("hsts-max-age", flagsChanged) {
		host.Data.Attributes.Security.HstsMaxAge = updateHostsHSTSMaxAge
	}
	if flagIn("hsts-preload", flagsChanged) {
		host.Data.Attributes.Security.HstsPreload = updateHostsHSTSPreload
	}

	res, err := c.UpdateHost(&host)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	res.Print()
}

func getFlagsChanged(cmd *cobra.Command) (flags []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			flags = append(flags, flag.Name)
		}
	})
	return flags
}

func flagIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
