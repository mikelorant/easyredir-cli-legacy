package easyredir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/alecthomas/chroma/quick"
)

type Rule struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			ForwardParams bool     `json:"forward_params"`
			ForwardPath   bool     `json:"forward_path"`
			ResponseType  string   `json:"response_type"`
			SourceUrls    []string `json:"source_urls"`
			TargetURL     string   `json:"target_url"`
		} `json:"attributes"`
		Relationships struct {
			SourceHosts struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"source_hosts"`
		} `json:"relationships"`
	} `json:"data"`
	Included interface{} `json:"included"`
}

type Rules struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			ForwardParams bool     `json:"forward_params"`
			ForwardPath   bool     `json:"forward_path"`
			ResponseType  string   `json:"response_type"`
			SourceURLs    []string `json:"source_urls"`
			TargetURL     string   `json:"target_url"`
		} `json:"attributes"`
		Relationships struct {
			SourceHosts struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"source_hosts"`
		} `json:"relationships"`
	} `json:"data"`
	Meta  Meta  `json:"meta"`
	Links Links `json:"links"`
}

type RulesOptions struct {
	Limit         int    `json:"limit"`
	StartingAfter string `json:"starting_after"`
	EndingBefore  string `json:"ending_before"`
	SourceURL     string `json:"sq"`
	TargetURL     string `json:"tq"`
}

func (c *Client) ListRules(options *RulesOptions) (rules Rules, err error) {
	limit := 100

	var sourceURL string
	var targetURL string

	var startingAfter string
	var endingBefore string

	if options != nil {
		sourceURL = options.SourceURL
		targetURL = options.TargetURL
	}

	for {
		res := Rules{}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rules?limit=%d&sq=%s&tq=%s&starting_after=%s&ending_before=%s", c.baseURL, limit, sourceURL, targetURL, startingAfter, endingBefore), nil)
		if err != nil {
			return rules, fmt.Errorf("ListRules: unable to create request: %w", err)
		}

		if err = c.sendRequest(req, &res); err != nil {
			return rules, fmt.Errorf("ListRules: unable to send request: %w", err)
		}

		rules.Data = append(rules.Data, res.Data...)

		if res.Meta.HasMore == false {
			break
		}

		if res.Links.Next != "" {
			u, err := url.Parse(res.Links.Next)
			if err != nil {
				panic(err)
			}
			startingAfter = u.Query().Get("starting_after")
		}
		// if res.Links.Prev != "" {
		// 	u, err := url.Parse(res.Links.Prev)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	endingBefore = u.Query().Get("ending_before")
		// }
	}

	return rules, nil
}

func (c *Client) CreateRule(r *Rule) (rule *Rule, err error) {
	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(r.Data.Attributes)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rules", c.baseURL), &buf)
	if err != nil {
		return nil, fmt.Errorf("CreateRule: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &rule); err != nil {
		return nil, fmt.Errorf("CreateRule: unable to send request: %w", err)
	}

	return rule, nil
}

func (c *Client) UpdateRule(r *Rule) (rule *Rule, err error) {
	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(r.Data.Attributes)

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/rules/%s", c.baseURL, r.Data.ID), &buf)
	if err != nil {
		return nil, fmt.Errorf("UpdateRule: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &rule); err != nil {
		return nil, fmt.Errorf("UpdateRule: unable to send request: %w", err)
	}

	return rule, nil
}

func (c *Client) RemoveRule(r *Rule) (rule *Rule, err error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/rules/%s", c.baseURL, r.Data.ID), nil)
	if err != nil {
		return nil, fmt.Errorf("RemoveRule: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &rule); err != nil {
		return nil, fmt.Errorf("RemoveRule: unable to send request: %w", err)
	}

	return rule, nil
}

func (r *Rules) Print() {
	t := table.NewWriter()

	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.DrawBorder = false
	t.Style().Color = table.ColorOptions{}
	t.Style().Box.PaddingLeft = ""
	t.Style().Box.PaddingRight = "    "
	t.Style().Color.Header = text.Colors{text.Bold}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "SOURCE URLS", "TARGET URL"})
	for _, h := range r.Data {
		row := []table.Row{}
		for i, s := range h.Attributes.SourceURLs {
			if i == 0 {
				row = append(row, table.Row{h.ID, s, h.Attributes.TargetURL})
				continue
			}
			row = append(row, table.Row{"", s, ""})
		}
		t.AppendRows(row)
	}
	t.Render()
}

func (r *Rule) Print() {
	fmt.Println(text.FgYellow.Sprint("RULE:"))

	tmpl := heredoc.Doc(`
    ID:   {{ .Data.ID }}
    Type: {{ .Data.Type }}
    Attributes:
      Forward Query: {{ .Data.Attributes.ForwardParams }}
      Forward Path:  {{ .Data.Attributes.ForwardPath }}
      Response Type: {{ .Data.Attributes.ResponseType }}
      Source URLs:
      {{- range .Data.Attributes.SourceUrls }}
      - {{ .}}
      {{- end }}
      Target URL:    {{ .Data.Attributes.TargetURL }}
    Relationships:
      Source Hosts:
      {{- range .Data.Relationships.SourceHosts.Data }}
      - {{ .ID }}
      {{- end }}

  `)

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(tmpl))
	t.Execute(&w, r)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")
}
