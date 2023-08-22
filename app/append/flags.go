package append

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

var source_bucket_uri string
var target_bucket_uri string

var spatial_database_uri string
var index_spatial_database bool

var iterator_uri string
var iterator_sources multi.MultiString

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("country")

	fs.StringVar(&source_bucket_uri, "source-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL files are read from.")
	fs.StringVar(&target_bucket_uri, "target-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to.")

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "", "...")
	fs.BoolVar(&index_spatial_database, "index-spatial-database", false, "...")

	fs.StringVar(&iterator_uri, "iterator-uri", "", "...")
	fs.Var(&iterator_sources, "iterator-source", "...")

	return fs
}
