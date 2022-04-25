package importer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/chroma/quick"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type YAMLRedirects []YAMLRedirect

type YAMLRedirect struct {
	Meta          YAMLRedirectMeta     `yaml:"meta"`
	Sources       []YAMLRedirectSource `yaml:"sources"`
	TargetURL     *string              `yaml:"target_url"`
	ForwardParams *bool                `yaml:"forward_params"`
	ForwardPath   *bool                `yaml:"forward_path"`
	ResponseType  *string              `yaml:"response_type"`
}

type YAMLRedirectMeta struct {
	Name        *string    `yaml:"name"`
	Description *string    `yaml:"description"`
	Expires     *time.Time `yaml:"expires"`
}

type YAMLRedirectSource struct {
	URL     *string                   `yaml:"url"`
	Options YAMLRedirectSourceOptions `yaml:"options"`
}

type YAMLRedirectSourceOptions struct {
	MatchOptions struct {
		CaseInsensitive  *bool `yaml:"case_insensitive"`
		SlashInsensitive *bool `yaml:"slash_insensitive"`
	}
	NotFoundAction struct {
		ForwardParams *bool   `yaml:"forward_params"`
		ForwardPath   *bool   `yaml:"forward_path"`
		Custom404Body *string `yaml:"custom_404_body"`
		ResponseCode  *int    `yaml:"response_code"`
		ResponseURL   *string `yaml:"response_url"`
	}
	Security struct {
		HTTPSUpgrade            *bool `yaml:"https_upgrade"`
		PreventForeignEmbedding *bool `yaml:"prevent_foreign_embedding"`
		HSTSIncludeSubDomains   *bool `yaml:"hsts_include_subdomains"`
		HSTSMaxAge              *int  `yaml:"hsts_max_age"`
		HSTSPreload             *bool `yaml:"hsts_preload"`
	}
}

var (
	defaultForwardParams bool   = false
	defaultForwardPath   bool   = false
	defaultResponseType  string = "moved_permanently"

	defaultMatchOptionsCaseInsensitive  bool = false
	defaultMatchOptionsSlashInsensitive bool = false

	defaultNotFoundActionForwardParams bool = false
	defaultNotFoundActionForwardPath   bool = false
	defaultNotFoundActionResponseCode  int  = 302

	defaultSecurityHTTPSUpgrade          bool = false
	defaultSecurityHSTSIncludeSubDomains bool = false
	defaultSecurityHSTSMaxAge            int  = -1
	defaultSecurityHSTSPreload           bool = false
)

func (rs *YAMLRedirects) Load(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = yaml.UnmarshalStrict(data, &rs)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func (rs *YAMLRedirects) Defaults() {
	drs := &YAMLRedirects{}

	for _, r := range *rs {
		dr := &YAMLRedirect{}

		if r.ForwardParams == nil {
			r.ForwardParams = &defaultForwardParams
		}
		if r.ForwardPath == nil {
			r.ForwardPath = &defaultForwardPath
		}
		if r.ResponseType == nil {
			r.ResponseType = &defaultResponseType
		}

		for _, s := range r.Sources {
			if s.Options.MatchOptions.CaseInsensitive == nil {
				s.Options.MatchOptions.CaseInsensitive = &defaultMatchOptionsCaseInsensitive
			}
			if s.Options.MatchOptions.SlashInsensitive == nil {
				s.Options.MatchOptions.SlashInsensitive = &defaultMatchOptionsSlashInsensitive
			}

			if s.Options.NotFoundAction.ForwardParams == nil {
				s.Options.NotFoundAction.ForwardParams = &defaultNotFoundActionForwardParams
			}
			if s.Options.NotFoundAction.ForwardPath == nil {
				s.Options.NotFoundAction.ForwardPath = &defaultNotFoundActionForwardPath
			}
			if s.Options.NotFoundAction.ResponseCode == nil {
				s.Options.NotFoundAction.ResponseCode = &defaultNotFoundActionResponseCode
			}

			if s.Options.Security.HTTPSUpgrade == nil {
				s.Options.Security.HTTPSUpgrade = &defaultSecurityHTTPSUpgrade
			}
			if s.Options.Security.HSTSIncludeSubDomains == nil {
				s.Options.Security.HSTSIncludeSubDomains = &defaultSecurityHSTSIncludeSubDomains
			}
			if s.Options.Security.HSTSMaxAge == nil {
				s.Options.Security.HSTSMaxAge = &defaultSecurityHSTSMaxAge
			}
			if s.Options.Security.HSTSPreload == nil {
				s.Options.Security.HSTSPreload = &defaultSecurityHSTSPreload
			}

			dr.Sources = append(dr.Sources, s)
		}

		dr.Meta = r.Meta
		dr.TargetURL = r.TargetURL
		dr.ForwardParams = r.ForwardParams
		dr.ForwardPath = r.ForwardPath
		dr.ResponseType = r.ResponseType

		*drs = append(*drs, *dr)
	}
	*rs = *drs

	return
}

func (r *YAMLRedirect) Print() {
	tmpl := heredoc.Doc(`
    {{- with .Meta }}
    Meta:
      {{- with .Name }}
      Name: {{ . }}
      {{- end }}
      {{- with .Description }}
      Description: {{ . }}
      {{- end }}
      {{- with .Expires }}
      Expires: {{ . }}
      {{- end }}
    {{- end }}
    Sources:
    {{- range .Sources }}
    - URL: {{ .URL }}
      Options:
        Match Options:
          Case Insensitive: {{ .Options.MatchOptions.CaseInsensitive }}
          Slash Insensitive: {{ .Options.MatchOptions.SlashInsensitive }}
        Not Found Action:
          Forward Params: {{ .Options.NotFoundAction.ForwardParams }}
          Forward Path: {{ .Options.NotFoundAction.ForwardPath }}
          Custom 404 Body Present: {{ .Options.NotFoundAction.Custom404Body }}
          Response Code: {{ .Options.NotFoundAction.ResponseCode }}
          Response URL: {{ .Options.NotFoundAction.ResponseURL }}
        Security:
          HTTPS Upgrade: {{ .Options.Security.HTTPSUpgrade }}
          Prevent Foreign Embedding: {{ .Options.Security.PreventForeignEmbedding }}
          HSTS Include Sub Domains: {{ .Options.Security.HSTSIncludeSubDomains }}
          HSTS Max Age: {{ .Options.Security.HSTSMaxAge }}
          HSTS Preload: {{ .Options.Security.HSTSPreload }}
    {{- end }}
    Target URL: {{ .TargetURL }}
    Forward Params: {{ .ForwardParams }}
    Forward Path: {{ .ForwardPath }}
    Response Type: {{ .ResponseType }}

  `)

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(tmpl))
	t.Execute(&w, r)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")

	return
}

func (rs *YAMLRedirects) Import(preview bool) {
	c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	for _, r := range *rs {
		r.Print()

		rule := easyredir.Rule{}
		rule.Data.Attributes.ForwardParams = *r.ForwardParams
		rule.Data.Attributes.ForwardPath = *r.ForwardPath
		rule.Data.Attributes.ResponseType = *r.ResponseType

		for _, v := range r.Sources {
			rule.Data.Attributes.SourceUrls = append(rule.Data.Attributes.SourceUrls, *v.URL)
		}
		rule.Data.Attributes.TargetURL = *r.TargetURL

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
