package importer

import (
	"fmt"
)

type Options struct {
	Format  string
	File    string
	Preview bool
}

func Import(options *Options) {
	switch options.Format {
	case "hiera":
		r := HieraRedirects{}
		r.Load(options.File)
		r.Import(options.Preview)
	case "yaml":
		r := YAMLRedirects{}
		r.Load(options.File)
		r.Defaults()
		r.Import(options.Preview)
	case "puppet":
		r := PuppetRedirects{}
		r.Load(options.File)
		r.Defaults()
		r.Import(options.Preview)
	default:
		fmt.Println("Unknown format.")
	}
}
