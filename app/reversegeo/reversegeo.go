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
	"github.com/whosonfirst/go-whosonfirst-spatial-hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial-hierarchy/filter"		
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	spatial_filter "github.com/whosonfirst/go-whosonfirst-spatial/filter"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
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

	// Set up spatial database
	
	spatial_db, err := database.NewSpatialDatabase(ctx, spatial_database_uri)

	if err != nil {
		return fmt.Errorf("Failed to create spatial database, %w", err)
	}
	
	resolver, err := hierarchy.NewPointInPolygonHierarchyResolver(ctx, spatial_db, nil)

	if err != nil {
		return fmt.Errorf("Failed to create new PIP resolver, %w", err)
	}
	
	inputs := &spatial_filter.SPRInputs{}
	
	results_cb := hierarchy_filter.FirstButForgivingSPRResultsFunc
	update_cb := hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
		
	// Walk Overture records

	walk_cb := func(ctx context.Context, r *walk.WalkRecord) error {

		new_body, err := resolver.PointInPolygonAndUpdate(ctx, inputs, results_cb, update_cb, r.Body)

		if err != nil {
			return fmt.Errorf("Failed to update record, %w", err)
		}
		
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