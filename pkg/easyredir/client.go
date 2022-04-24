package easyredir

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/text"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	baseURLV1 = "https://api.easyredir.com/v1"
)

type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	HTTPClient *http.Client
}

type errorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Param    string `json:"param"`
		Code     string `json:"code"`
		Message  string `json:"message"`
	} `json:"errors"`
}

type Meta struct {
	HasMore bool `json:"has_more"`
}

type Links struct {
	Next interface{} `json:"next"`
	Prev interface{} `json:"prev"`
}

func NewClient() (c *Client, err error) {
	if key := viper.GetString("api.key"); key == "" {
		return nil, fmt.Errorf("NewClient: missing api.key")
	}

	if secret := viper.GetString("api.secret"); secret == "" {
		return nil, fmt.Errorf("NewClient: missing api.secret")
	}

	c = &Client{
		baseURL:    baseURLV1,
		apiKey:     viper.GetString("api.key"),
		apiSecret:  viper.GetString("api.secret"),
		HTTPClient: &http.Client{},
	}

	return c, nil
}

func (c *Client) sendRequest(req *http.Request, v interface{}) (err error) {
	cred := authorization(c.apiKey, c.apiSecret)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", cred))

	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		req.Header.Set("Idempotency-Key", uuid.NewString())
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendRequest: unable to send request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		errRes := errorResponse{}
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			errRes.Print()
			return fmt.Errorf("sendRequest: error message: %s type: %s", errRes.Message, errRes.Type)
		}

		return fmt.Errorf("sendRequest: unknown error, status code: %d", res.StatusCode)
	}

	if res.ContentLength == 0 {
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return fmt.Errorf("sendRequest: unable to decode JSON into struct: %w", err)
	}

	if res.Header["x-ratelimit-limit"] != nil {
		log.Warn().Msg(fmt.Sprintf("Rate limit: %s", res.Header["x-ratelimit-limit"]))
		log.Warn().Msg(fmt.Sprintf("Rate limit remaining: %s", res.Header["x-ratelimit-remaining"]))
		log.Warn().Msg(fmt.Sprintf("Rate limit reset: %s", res.Header["x-ratelimit-reset"]))
	}

	return nil
}

func authorization(username string, password string) (credential string) {
	userPass := strings.Join([]string{username, password}, ":")
	userPassBytes := []byte(userPass)

	len := base64.StdEncoding.EncodedLen(len(userPassBytes))
	credBytes := make([]byte, len)

	base64.StdEncoding.Encode(credBytes, userPassBytes)

	return string(credBytes)
}

func (e *errorResponse) Print() {
	tmpl := heredoc.Doc(`
    {{ blue "Type" }}:   {{ .Type| green }}
    {{ blue "Message" }}: {{ .Message | green }}
    {{ blue "Errors" }}:
    {{- range .Errors }}
    - {{ blue "Resource" }}: {{ .Resource | green }}
      {{ blue "Code" }}:  {{ .Code | green }}
      {{ blue "Param" }}: {{ .Param | green }}
      {{ blue "Message" }}: {{ .Message | green }}
    {{- end}}
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

	t.Execute(os.Stdout, e)
}
