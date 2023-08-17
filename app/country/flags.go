package country

import (
	"flag"
	
	"github.com/sfomuseum/go-flags/flagset"
)

var source_bucket_uri string
var target_bucket_uri string

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("country")

	fs.StringVar(&source_bucket_uri, "source-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL files are read from.")
	fs.StringVar(&target_bucket_uri, "target-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to.")

	return fs
}
