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

var wof_placetype string

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("country")

	fs.StringVar(&source_bucket_uri, "source-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL files are read from.")
	fs.StringVar(&target_bucket_uri, "target-bucket-uri", "file:///", "A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to.")

	fs.StringVar(&spatial_database_uri, "spatial-database-uri", "sqlite://?dsn=modernc://mem", "A valid whosonfirst/go-whosonfirst-spatial.SpatialDatabase URI.")
	fs.BoolVar(&index_spatial_database, "index-spatial-database", true, "Create a point-in-polygon enabled spatial index at runtime. If true then both -iterator-uri and -iterator-source must be set.")

	fs.StringVar(&iterator_uri, "iterator-uri", "git:///tmp", "A valid whosonfirst/go-whosnfirst-iterate/v2 URI.")
	fs.Var(&iterator_sources, "iterator-source", "Zero or more URIs for the iterator defined by -iterator-uri to process.")

	fs.StringVar(&wof_placetype, "whosonfirst-placetype", "", "The Who's On First placetype to assign to each Overture record being processed.")
	return fs
}
