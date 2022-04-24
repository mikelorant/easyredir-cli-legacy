package cmd

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	createForwardParams bool
	createForwardPath   bool
	createResponseType  string
	createSourceUrls    []string
	createTargetURL     string

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "A brief description of your command",
	}

	createRulesCmd = &cobra.Command{
		Use:   "rule",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			doCreateRule()
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createRulesCmd)
	createRulesCmd.Flags().BoolVarP(&createForwardParams, "forward-params", "", defaultForwardParams, "Forward params")
	createRulesCmd.Flags().BoolVarP(&createForwardPath, "forward-path", "", defaultForwardPath, "Forward path")
	createRulesCmd.Flags().StringVarP(&createResponseType, "response-type", "", defaultResponseType, "Response type")
	createRulesCmd.Flags().StringSliceVarP(&createSourceUrls, "source-urls", "", nil, "Source URLs")
	createRulesCmd.Flags().StringVarP(&createTargetURL, "target-url", "", "", "Target URL")
	createRulesCmd.MarkFlagRequired("source-urls")
	createRulesCmd.MarkFlagRequired("target-url")
}

func doCreateRule() {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	rule := easyredir.Rule{}
	rule.Data.Attributes.ForwardParams = createForwardParams
	rule.Data.Attributes.ForwardPath = createForwardPath
	rule.Data.Attributes.ResponseType = createResponseType
	rule.Data.Attributes.SourceUrls = createSourceUrls
	rule.Data.Attributes.TargetURL = createTargetURL

	res, err := c.CreateRule(&rule)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	res.Print()
}
