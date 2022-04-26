package importer

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/tailscale/hujson"
	"io"
	"io/ioutil"
	"os"
	"regexp"
  "text/template"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

  "github.com/alecthomas/chroma/quick"
	"github.com/rs/zerolog/log"
	"github.com/jedib0t/go-pretty/v6/text"

	_ "embed"
)

const (
	manifestStart string = "$controllers = "
)

var (
  //go:embed puppet_print.tmpl
  puppetPrintTemplate string
)

type PuppetRedirects []PuppetRedirect

type PuppetRedirect struct {
	Name          string
	SourceURLs    []string `json:"apache_hosts"`
	TargetURL     string   `json:"apache_redirect"`
  ForwardParams *bool
  ForwardPath   *bool
  ResponseType  *string
}

func (rs *PuppetRedirects) Load(file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	f := bytes.NewReader(content)
	startOffset := scanFile(f, []byte(manifestStart)) + int64(len(manifestStart))
	endOffset := scanBracket(f, startOffset)

  block := getBlock(file, int(startOffset), int(endOffset))
	res := convertPuppet(block)

	rawRedirects := make(map[string]map[string]PuppetRedirect)

	err = hujson.Unmarshal([]byte(res), &rawRedirects)
	if err != nil {
		fmt.Println(err)
	}

	for k, v := range rawRedirects["redirects"] {
		v.Name = k
		*rs = append(*rs, v)
	}
}

func (rs *PuppetRedirects) Defaults() {
	drs := &PuppetRedirects{}

	for _, r := range *rs {
		dr := &PuppetRedirect{}

		if r.ForwardParams == nil {
			r.ForwardParams = &defaultForwardParams
		}
		if r.ForwardPath == nil {
			r.ForwardPath = &defaultForwardPath
		}
		if r.ResponseType == nil {
			r.ResponseType = &defaultResponseType
		}

    dr.Name = r.Name
		dr.ForwardParams = r.ForwardParams
		dr.ForwardPath = r.ForwardPath
		dr.ResponseType = r.ResponseType
    dr.SourceURLs = r.SourceURLs
    dr.TargetURL = r.TargetURL

		*drs = append(*drs, *dr)
	}
	*rs = *drs

	return
}

func (rs *PuppetRedirects) Import(preview bool) {
  c, err := easyredir.NewClient()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

  for _, r := range *rs {
    r.Print()

    if preview == true {
			return
		}

		rule := easyredir.Rule{}
		rule.Data.Attributes.ForwardParams = *r.ForwardParams
		rule.Data.Attributes.ForwardPath = *r.ForwardPath
		rule.Data.Attributes.ResponseType = *r.ResponseType

		for _, v := range r.SourceURLs {
			rule.Data.Attributes.SourceUrls = append(rule.Data.Attributes.SourceUrls, v)
		}
		rule.Data.Attributes.TargetURL = r.TargetURL

		res, err := c.CreateRule(&rule)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		res.Print()
  }
}

func (r *PuppetRedirect) Print() {
  fmt.Printf("%s:\n", text.FgCyan.Sprint("CONFIG"))
	fmt.Println()

	var w bytes.Buffer

	t := template.Must(template.New("").Parse(puppetPrintTemplate))
	t.Execute(&w, r)

	quick.Highlight(os.Stdout, w.String(), "yaml", "terminal256", "pygments")

	fmt.Println()

	return
}

func scanFile(f io.ReadSeeker, search []byte) int64 {
	ix := 0
	r := bufio.NewReader(f)
	offset := int64(0)
	for ix < len(search) {
		b, err := r.ReadByte()
		if err != nil {
			return -1
		}
		if search[ix] == b {
			ix++
		} else {
			ix = 0
		}
		offset++
	}
	return offset - int64(len(search))
}

func scanBracket(f io.ReadSeeker, offset int64) int64 {
	ix := 0
	// bracket match depth is 0
	bm := false
	// found matching close bracket
	cb := false
	r := bufio.NewReader(f)
	f.Seek(offset, 0)
	for !(bm && cb) {
		b, err := r.ReadByte()
		if err != nil {
			return -1
		}
		if '{' == b {
			ix++
			cb = false
		}
		if '}' == b {
			ix--
			cb = true
		}
		if ix == 0 {
			bm = true
		}
		if ix != 0 {
			bm = false
		}
		offset++
	}
	return offset
}

func convertPuppet(content []byte) []byte {
	reRocket := regexp.MustCompile(`([1-9a-z_]+).*=>`)
	removedRocket := reRocket.ReplaceAll(content, []byte("\"$1\":"))

	reSingleQuote := regexp.MustCompile(`'`)
	removedQuote := reSingleQuote.ReplaceAll(removedRocket, []byte("\""))

	return removedQuote
}

func getBlock(filename string, startOffset int, endOffset int) []byte {
  fd, err := os.Open(filename)
  if err != nil {
    log.Error().Err(err).Msg("")
    return nil
  }

  reader := bufio.NewReaderSize(fd, endOffset - startOffset)
  _, _ = reader.Discard(startOffset)
  block, err := reader.Peek(endOffset - startOffset)
  if err != nil {
    fmt.Println(err)
  }

  return block
}
