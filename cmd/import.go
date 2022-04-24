package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/chroma/quick"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	importFile string
	preview		 bool

	importCmd = &cobra.Command{
		Use:   "import",
		Short: "A brief description of your command",
	}

	importRulesCmd = &cobra.Command{
		Use:   "rules",
		Short: "A brief description of your command",
		Run: func(cmd *cobra.Command, args []string) {
			doImportRules()
		},
	}
)

type Redirect struct {
	Name          string
	Aliases       []string `yaml:"aliases"`
	ExtraRewrites []string `yaml:"extra_rewrites"`
	Host          string   `yaml:"host"`
	Redirect      string   `yaml:"redirect"`
	SquashPath    bool     `yaml:"squash_path"`
	Type          int      `yaml:"type"`
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.AddCommand(importRulesCmd)
	importRulesCmd.Flags().BoolVarP(&preview, "preview", "", false, "Preview")
	importRulesCmd.Flags().StringVarP(&importFile, "file", "", "", "Filename")
	importRulesCmd.MarkFlagRequired("file")
}

func doImportRules() {
	data, err := ioutil.ReadFile(importFile)
	if err != nil {
		panic(err)
	}

	fr := make(map[string]map[string]Redirect)

	err = yaml.Unmarshal(data, &fr)
	if err != nil {
		fmt.Println(err)
	}

	redirects := []Redirect{}
	for k, v := range fr["web_redirects"] {
		v.Name = k
		redirects = append(redirects, v)
	}

	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	for _, r := range redirects {
		doImportPrint(r)

		rule := easyredir.Rule{}

		// For now we will always forward params however we need to check the extra_rewrite value in the future.
		rule.Data.Attributes.ForwardParams = true

		// Forward path is the inverse of squash path so we negate it.
		rule.Data.Attributes.ForwardPath = !r.SquashPath

		// Only two types of response type (301 or 302) and unless set to 301 we default to 302.
		if r.Type == 301 {
			rule.Data.Attributes.ResponseType = "found"
		} else {
			rule.Data.Attributes.ResponseType = "moved_permanently"
		}

		// Combine host and aliases to create the complete source URLs.
		rule.Data.Attributes.SourceUrls = append(rule.Data.Attributes.SourceUrls, r.Host)
		for _, v := range r.Aliases {
			rule.Data.Attributes.SourceUrls = append(rule.Data.Attributes.SourceUrls, v)
		}

		// The actual target to redirect to.
		rule.Data.Attributes.TargetURL = r.Redirect

		if preview != true {
			res, err := c.CreateRule(&rule)
			if err != nil {
				log.Error().Err(err).Msg("")
			}

			res.Print()
		}
	}
}

func doImportPrint(redirect Redirect) {
	tmpl := heredoc.Doc(`
    Name: {{ .Name }}
    Host: {{ .Host }}
    Redirect: {{ .Redirect }}
    {{- with .Type }}
    Type: {{ . }}
    {{- end }}
    {{- with .ExtraRewrites }}
    Extra Rewrites:
    {{- range . }}
    - {{ . }}
    {{- end }}
    {{- end }}
    {{- with .SquashPath}}
    Squash Path: {{ . }}
    {{- end }}
    {{- with .Aliases }}
    Aliases:
    {{- range . }}
    - {{ . }}
    {{- end }}
    {{ end }}
  `)

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(tmpl))
	t.Execute(&w, redirect)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")
}
