package importer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/chroma/quick"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type HieraRedirects []HieraRedirect

type HieraRedirect struct {
	Name          string
	Aliases       []string `yaml:"aliases"`
	ExtraRewrites []string `yaml:"extra_rewrites"`
	Host          string   `yaml:"host"`
	Redirect      string   `yaml:"redirect"`
	SquashPath    bool     `yaml:"squash_path"`
	Type          int      `yaml:"type"`
	RewriteRules  []HieraRewriteRule
}

type HieraRewriteRules []HieraRewriteRule

type HieraRewriteRule struct {
	Pattern string
	Target  string
	Flags   HieraRewriteRuleFlags
}

type HieraRewriteRuleFlags struct {
	Last               bool // [L]
	NoEscape           bool // [NE]
	QueryStringDiscard bool // [QSD]
	Redirect           int  // [R=x]
}

func (rs *HieraRedirects) Load(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	rawRedirects := make(map[string]map[string]HieraRedirect)

	err = yaml.Unmarshal(data, &rawRedirects)
	if err != nil {
		fmt.Println(err)
	}

	for k, v := range rawRedirects["web_redirects"] {
		v.Name = k
		v.RewriteRules = HieraRewriteRules{}
		v.parseRewrites()
		*rs = append(*rs, v)
	}

	return
}

func (r *HieraRedirect) parseRewrites() {
	for _, rewrite := range r.ExtraRewrites {
		rr := HieraRewriteRule{}

		rs := strings.Split(rewrite, " ")
		rr.Pattern = rs[0]
		rr.Target = rs[1]
		if len(rs) == 3 {
			rr.Flags = HieraRewriteRuleFlags{}
			rr.Flags.parseFlags(rs[2])
		}

		r.RewriteRules = append(r.RewriteRules, rr)
	}

	return
}

func (flags *HieraRewriteRuleFlags) parseFlags(f string) {
	ft := strings.Trim(f, "[]")
	fs := strings.Split(ft, ",")
	for _, v := range fs {
		switch v {
		case "R=301":
			flags.Redirect = 301
		case "R=302":
			flags.Redirect = 302
		case "L":
			flags.Last = true
		case "QSD":
			flags.QueryStringDiscard = true
		case "NE":
			flags.NoEscape = true
		default:
			fmt.Println("Found unknown flag.")
		}
	}

	return
}

func (rs *HieraRedirects) Import(preview bool) {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	for _, r := range *rs {
		r.Print()

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

	return
}

func (r *HieraRedirect) Print() {
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
    {{- end }}
    {{- with .RewriteRules }}
    Rewrite Rules:
    {{- range . }}
    - Pattern: {{ .Pattern }}
      Target: {{ .Target }}
      {{- if or (.Flags.Last) (.Flags.Redirect) (.Flags.QueryStringDiscard) (.Flags.NoEscape) }}
      {{- with .Flags }}
      Flags:
        {{- with .Last }}
        Last: {{ . }}
        {{- end }}
        {{- with .Redirect }}
        Redirect: {{ . }}
        {{- end }}
        {{- with .QueryStringDiscard }}
        Query String Discard: {{ . }}
        {{- end }}
        {{- with .NoEscape }}
        No Escape: {{ . }}
        {{- end }}
      {{- end }}
      {{- end }}
    {{- end }}
    {{- end }}

  `)

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(tmpl))
	t.Execute(&w, r)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")

	return
}
