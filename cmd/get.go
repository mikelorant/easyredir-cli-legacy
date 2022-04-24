package cmd

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	getSourceURL string
	getTargetURL string

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "A brief description of your command",
	}

	getHostsCmd = &cobra.Command{
		Use:   "hosts",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			doGetHosts()
		},
	}
)

var getRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		doGetRules()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getHostsCmd)
	getCmd.AddCommand(getRulesCmd)
	getCmd.PersistentFlags().StringVar(&getSourceURL, "source-url", "", "source url")
	getCmd.PersistentFlags().StringVar(&getTargetURL, "target-url", "", "target url")
}

func doGetHosts() {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	o := easyredir.HostsOptions{}
	hosts, err := c.ListHosts(&o)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	hosts.Print()
}

func doGetRules() {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	o := easyredir.RulesOptions{
		SourceURL:     getSourceURL,
		TargetURL:     getTargetURL,
	}
	rules, err := c.ListRules(&o)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	rules.Print()
}
