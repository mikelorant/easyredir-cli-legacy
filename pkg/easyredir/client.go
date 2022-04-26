package easyredir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/alecthomas/chroma/quick"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	_ "embed"
)

const (
	baseURLV1 = "https://api.easyredir.com/v1"
)

//go:embed client_error.tmpl
var clientErrorTemplate string

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
	Next string `json:"next"`
	Prev string `json:"prev"`
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
	req.SetBasicAuth(c.apiKey, c.apiSecret)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

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

func (e *errorResponse) Print() {
	fmt.Printf("%s:\n", text.FgRed.Sprint("ERROR"))
	fmt.Println()

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(clientErrorTemplate))
	t.Execute(&w, e)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")

	fmt.Println()

	return
}
