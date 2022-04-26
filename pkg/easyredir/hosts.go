package easyredir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	_ "embed"
)

//go:embed host_print.tmpl
var hostPrintTemplate string

type Host struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name              string `json:"name"`
			DNSStatus         string `json:"dns_status"`
			CertificateStatus string `json:"certificate_status"`
			MatchOptions      struct {
				CaseInsensitive  interface{} `json:"case_insensitive,omitempty"`
				SlashInsensitive interface{} `json:"slash_insensitive,omitempty"`
			} `json:"match_options"`
			Security struct {
				HTTPSUpgrade            interface{} `json:"https_upgrade,omitempty"`
				PreventForeignEmbedding interface{} `json:"prevent_foreign_embedding,omitempty"`
				HstsIncludeSubDomains   interface{} `json:"hsts_include_sub_domains,omitempty"`
				HstsMaxAge              interface{} `json:"hsts_max_age,omitempty"`
				HstsPreload             interface{} `json:"hsts_preload,omitempty"`
			} `json:"security"`
			NotFoundAction struct {
				ForwardParams        interface{} `json:"forward_params,omitempty"`
				ForwardPath          interface{} `json:"forward_path,omitempty"`
				Custom404BodyPresent bool        `json:"custom_404_body_present,omitempty"`
				Custom404Body        string      `json:"custom_404_body,omitempty"`
				ResponseCode         int         `json:"response_code,omitempty"`
				ResponseURL          interface{} `json:"response_url,omitempty"`
			} `json:"not_found_action"`
			AcmeEnabled        bool `json:"acme_enabled"`
			DetectedDNSEntries []struct {
				Type   string   `json:"type"`
				Values []string `json:"values"`
			} `json:"detected_dns_entries"`
			DNSTestedAt        time.Time `json:"dns_tested_at"`
			RequiredDNSEntries struct {
				Recommended struct {
					Type   string   `json:"type"`
					Values []string `json:"values"`
				} `json:"recommended"`
				Alternatives []struct {
					Type   string   `json:"type"`
					Values []string `json:"values"`
				} `json:"alternatives"`
			} `json:"required_dns_entries"`
		} `json:"attributes"`
		Links struct{} `json:"links"`
	} `json:"data"`
}

type Hosts struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name              string `json:"name"`
			DNSStatus         string `json:"dns_status"`
			CertificateStatus string `json:"certificate_status"`
		} `json:"attributes"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Meta  Meta  `json:"meta"`
	Links Links `json:"links"`
}

type HostsOptions struct {
	Limit         int    `json:"limit"`
	StartingAfter string `json:"starting_after"`
	EndingBefore  string `json:"ending_before"`
}

func (c *Client) ListHosts(options *HostsOptions) (hosts Hosts, err error) {
	limit := 100

	var startingAfter string
	var endingBefore string

	for {
		res := Hosts{}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/hosts?limit=%d&starting_after=%s&ending_before=%s", c.baseURL, limit, startingAfter, endingBefore), nil)
		if err != nil {
			return hosts, fmt.Errorf("ListHosts: unable to create request: %w", err)
		}

		if err = c.sendRequest(req, &res); err != nil {
			return hosts, fmt.Errorf("ListHosts: unable to send request: %w", err)
		}

		hosts.Data = append(hosts.Data, res.Data...)

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

	return hosts, nil
}

func (c *Client) GetHost(host *Host) (err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/hosts/%s", c.baseURL, host.Data.ID), nil)
	if err != nil {
		return fmt.Errorf("GetHost: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &host); err != nil {
		return fmt.Errorf("GetHost: unable to send request: %w", err)
	}

	return nil
}

func (c *Client) UpdateHost(h *Host) (host *Host, err error) {
	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(h.Data.Attributes)
	if err != nil {
		return nil, fmt.Errorf("UpdateHost: unable to encode attributes: %w", err)
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/hosts/%s", c.baseURL, h.Data.ID), &buf)
	if err != nil {
		return nil, fmt.Errorf("UpdateHost: unable to create request: %w", err)
	}

	if err = c.sendRequest(req, &host); err != nil {
		return nil, fmt.Errorf("UpdateHost: unable to send request: %w", err)
	}

	return host, nil
}

func (r *Hosts) Print() {
	t := table.NewWriter()

	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.DrawBorder = false
	t.Style().Color = table.ColorOptions{}
	t.Style().Box.PaddingLeft = ""
	t.Style().Box.PaddingRight = "    "
	t.Style().Color.Header = text.Colors{text.Bold}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "NAME", "DNS STATUS", "CERTIFICATE STATUS"})
	for _, h := range r.Data {
		t.AppendRow(table.Row{h.ID, h.Attributes.Name, h.Attributes.DNSStatus, h.Attributes.CertificateStatus})
	}
	t.Render()

	return
}

func (r *Host) Print() {
	fmt.Printf("%s:\n", text.FgMagenta.Sprint("HOST"))
	fmt.Println()

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(hostPrintTemplate))
	t.Execute(&w, r)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")

	fmt.Println()

	return
}
