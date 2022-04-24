package easyredir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/jedib0t/go-pretty/text"
	"golang.org/x/term"
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
	} `json:"data"`
	Relationships struct {
		RelationshipType struct {
			Data []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
			Links struct {
				Related string `json:"related"`
			} `json:"links"`
		} `json:"[relationship type]"`
	} `json:"relationships"`
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

func (c *Client) ListRules(options *RulesOptions) (rules *Rules, err error) {
	limit := 25

	var sourceURL string
	var targetURL string

	if options != nil {
		limit = options.Limit
		sourceURL = options.SourceURL
		targetURL = options.TargetURL
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rules?limit=%d&sq=%s&tq=%s", c.baseURL, limit, sourceURL, targetURL), nil)
	if err != nil {
		return nil, fmt.Errorf("ListRules: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &rules); err != nil {
		return nil, fmt.Errorf("ListRules: unable to send request: %w", err)
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
	t.Style().Box.PaddingRight = "\t"
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
	tmpl := heredoc.Doc(`
    {{ blue "ID" }}:   {{ .Data.ID | green }}
    {{ blue "Type" }}: {{ .Data.Type | green }}
    {{ blue "Attributes" }}:
      {{ blue "Forward Query" }}: {{ .Data.Attributes.ForwardParams | green }}
      {{ blue "Forward Path" }}:  {{ .Data.Attributes.ForwardPath | green }}
      {{ blue "Response Type" }}: {{ .Data.Attributes.ResponseType | green }}
      {{ blue "Source URLs" }}:
      {{- range .Data.Attributes.SourceUrls }}
      - {{ . | green }}
      {{- end }}
      {{ blue "Target URL" }}:    {{ .Data.Attributes.TargetURL | green }}
  `)

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		text.DisableColors()
	}

	t := template.Must(template.
		New("").
		Funcs(map[string]interface{}{
			"blue": func(v interface{}) string {
				return text.FgHiBlue.Sprint(v)
			},
			"green": func(v interface{}) string {
				return text.FgHiGreen.Sprint(v)
			},
		}).
		Parse(tmpl))

	t.Execute(os.Stdout, r)
}
