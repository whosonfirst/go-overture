package country

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var source_bucket_uri string
var target_bucket_uri string

var is_bzip2 bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("country")

	fs.StringVar(&source_bucket_uri, "source-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL files are read from.")
	fs.StringVar(&target_bucket_uri, "target-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to.")
	fs.BoolVar(&is_bzip2, "bzip2", false, "A boolean flag indicating the input files are bzip2-compressed.")

	return fs
}
