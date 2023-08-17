package reversegeo

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-overture/geojsonl"
)

func Run(ctx context.Context, logger *log.Logger) error {

	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	flagset.Parse(fs)

	uris := fs.Args()

	source_bucket, err := bucket.OpenBucket(ctx, source_bucket_uri)

	if err != nil {
		return fmt.Errorf("Failed to open source bucket, %v", err)
	}

	defer source_bucket.Close()

	// Build PIP stuff here

	// Walk Overture records

	walk_cb := func(ctx context.Context, r *walk.WalkRecord) error {

		rsp := gjson.GetBytes(r.Body, "geometry.coordinates")
		logger.Println(rsp.String())
		return nil
	}

	walk_opts := &geojsonl.WalkOptions{
		SourceBucket: source_bucket,
		Callback:     walk_cb,
	}

	err = geojsonl.Walk(ctx, walk_opts, uris...)

	if err != nil {
		return fmt.Errorf("Failed to wal, %w", err)
	}

	return nil
}
