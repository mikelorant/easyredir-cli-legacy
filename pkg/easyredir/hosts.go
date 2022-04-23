package easyredir

import (
  "net/http"
  "fmt"
  "text/tabwriter"
  "text/template"
  "os"
  "time"
  "bytes"
  "encoding/json"

  "golang.org/x/term"
  "github.com/jedib0t/go-pretty/text"
  "github.com/MakeNowJust/heredoc/v2"
)

type Host struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name              string `json:"name"`
			DNSStatus         string `json:"dns_status"`
			CertificateStatus string `json:"certificate_status"`
			MatchOptions      struct {
				CaseInsensitive  interface{} `json:"case_insensitive"`
				SlashInsensitive interface{} `json:"slash_insensitive"`
			} `json:"match_options"`
			Security struct {
				HTTPSUpgrade            interface{} `json:"https_upgrade"`
				PreventForeignEmbedding interface{} `json:"prevent_foreign_embedding"`
				HstsIncludeSubDomains   interface{} `json:"hsts_include_sub_domains"`
				HstsMaxAge              interface{} `json:"hsts_max_age"`
				HstsPreload             interface{} `json:"hsts_preload"`
			} `json:"security"`
			NotFoundAction struct {
				ForwardParams        interface{} `json:"forward_params"`
				ForwardPath          interface{} `json:"forward_path"`
				Custom404BodyPresent bool        `json:"custom_404_body_present"`
        Custom404Body string        `json:"custom_404_body"`
				ResponseCode         int         `json:"response_code"`
				ResponseURL          interface{} `json:"response_url"`
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
		Links struct {
		} `json:"links"`
	} `json:"data"`
}

type Hosts struct {
	Data  []struct {
    ID         string          `json:"id"`
  	Type       string          `json:"type"`
  	Attributes struct {
      Name              string `json:"name"`
      DNSStatus         string `json:"dns_status"`
      CertificateStatus string `json:"certificate_status"`
      } `json:"attributes"`
  	Links      struct {
      Self     string          `json:"self"`
    } `json:"links"`
  } `json:"data"`
	Meta  Meta        `json:"meta"`
	Links Links       `json:"links"`
}

type HostsOptions struct {
  Limit          int    `json:"limit"`
  StartingAfter  string `json:"starting_after"`
  EndingBefore   string `json:"ending_before"`
}

func (c *Client) ListHosts(options *HostsOptions) (hosts *Hosts, err error) {
  limit := 25
  var startingAfter string
  var endingBefore  string

  if options != nil {
    limit = options.Limit
    startingAfter = options.StartingAfter
    endingBefore = options.EndingBefore
  }

  req, err := http.NewRequest("GET", fmt.Sprintf("%s/hosts?limit=%d&starting_after=%s&ending_before=%s", c.baseURL, limit, startingAfter, endingBefore), nil)
  if err != nil {
    return nil, fmt.Errorf("ListHosts: unable to create request: %w", err)
  }

  if err = c.sendRequest(req, &hosts); err != nil {
    return nil, fmt.Errorf("ListHosts: unable to send request: %w", err)
  }

  return hosts, nil
}

func (c *Client) GetHost(host *Host) (err error) {
  req, err := http.NewRequest("GET", fmt.Sprintf("%s/hosts/%s", c.baseURL, host.Data.ID), nil)
  if err != nil {
    return fmt.Errorf("GetHost: unable to create request: %w", err)
  }

  if err = c.sendRequest(req, &host); err != nil {
    return  fmt.Errorf("GetHost: unable to send request: %w", err)
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
  w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)

  const fs = "%s\t%s\t%s\t%s\n"
  fmt.Fprintf(w, fs, "ID", "NAME", "DNS STATUS", "CERTIFICATE STATUS")

  for _, h := range r.Data {
    fmt.Fprintf(w, fs, h.ID, h.Attributes.Name, h.Attributes.DNSStatus, h.Attributes.CertificateStatus)
  }

  w.Flush()
}

func (r *Host) Print() {
  tmpl := heredoc.Doc(`
    {{ blue "ID" }}:   {{ .Data.ID | green }}
    {{ blue "Type" }}: {{ .Data.Type | green }}
    {{ blue "Attributes" }}:
      {{ blue "Name" }}:               {{ .Data.Attributes.Name | green }}
      {{ blue "DNS Status" }}:         {{ .Data.Attributes.DNSStatus | green }}
      {{ blue "Certificate Status" }}: {{ .Data.Attributes.CertificateStatus | green }}
      {{ blue "Match Options" }}:
        {{ blue "Case Insensitive" }}:  {{ .Data.Attributes.MatchOptions.CaseInsensitive | green }}
        {{ blue "Slash Insensitive" }}: {{ .Data.Attributes.MatchOptions.SlashInsensitive | green }}
      {{ blue "Security" }}:
        {{ blue "HTTPS Upgrade" }}:             {{ .Data.Attributes.Security.HTTPSUpgrade | green }}
        {{ blue "Prevent Foreign Embedding" }}: {{ .Data.Attributes.Security.PreventForeignEmbedding | green }}
        {{ blue "HSTS Include Sub Domains" }}:  {{ .Data.Attributes.Security.HstsIncludeSubDomains | green }}
        {{ blue "HSTS Max Age" }}:              {{ .Data.Attributes.Security.HstsMaxAge | green }}
        {{ blue "HSTS Preload" }}:              {{ .Data.Attributes.Security.HstsPreload | green }}
      {{ blue "Not Found Action" }}:
        {{ blue "Forward Params" }}:          {{ .Data.Attributes.NotFoundAction.ForwardParams | green }}
        {{ blue "Forward Path" }}:            {{ .Data.Attributes.NotFoundAction.ForwardPath | green }}
        {{ blue "Custom 404 Body Present" }}: {{ .Data.Attributes.NotFoundAction.Custom404BodyPresent | green }}
        {{ blue "Response Code" }}:           {{ .Data.Attributes.NotFoundAction.ResponseCode | green }}
        {{ blue "Response URL" }}:            {{ .Data.Attributes.NotFoundAction.ResponseURL | green }}
      {{ blue "ACME Enabled" }}: {{ .Data.Attributes.AcmeEnabled | green }}
      {{ blue "Detected DNS Entries" }}:
      {{- range .Data.Attributes.DetectedDNSEntries }}
      - {{ blue "Type" }}: {{ .Type | green }}
        {{ blue "Values" }}:
        {{- range .Values }}
        - {{ . | green }}
        {{- end }}
      {{- end }}
      {{ blue "DNS Tested At" }}: {{ .Data.Attributes.DNSTestedAt | green }}
      {{ blue "Required DNS Entries" }}:
        {{ blue "Recommended" }}:
        - {{ blue "Type" }}: {{ .Data.Attributes.RequiredDNSEntries.Recommended.Type | green }}
          {{ blue "Values" }}:
          {{- range .Data.Attributes.RequiredDNSEntries.Recommended.Values }}
          - {{ . | green}}
          {{- end }}
        {{ blue "Alternatives" }}:
        {{- range .Data.Attributes.RequiredDNSEntries.Alternatives }}
        - {{ blue "Type" }}: {{ .Type | green }}
          {{ blue "Values" }}:
          {{- range .Values }}
          - {{ . | green }}
          {{- end }}
        {{- end }}
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
