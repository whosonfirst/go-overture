package append

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	_ "time"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-timings"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-overture/geojsonl"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial-hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial-hierarchy/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	spatial_filter "github.com/whosonfirst/go-whosonfirst-spatial/filter"
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

	target_bucket, err := bucket.OpenBucket(ctx, target_bucket_uri)

	if err != nil {
		return fmt.Errorf("Failed to open target bucket, %v", err)
	}

	defer target_bucket.Close()

	// Set up spatial database

	spatial_db, err := database.NewSpatialDatabase(ctx, spatial_database_uri)

	if err != nil {
		return fmt.Errorf("Failed to create spatial database, %w", err)
	}

	// Optionally index spatial database here

	if index_spatial_database {

		iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

			body, err := io.ReadAll(r)

			if err != nil {
				return fmt.Errorf("Failed to read both for %s, %w", path, err)
			}

			err = spatial_db.IndexFeature(ctx, body)

			if err != nil {
				return fmt.Errorf("Failed to index %s, %w", path, err)
			}

			return nil
		}

		iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

		if err != nil {
			return fmt.Errorf("Failed to create new iterator, %w", err)
		}

		err = iter.IterateURIs(ctx, iterator_sources...)

		if err != nil {
			return fmt.Errorf("Failed to iterator sources, %w", err)
		}
	}

	// Set up PIP/hierarchy resolver

	resolver_opts := &hierarchy.PointInPolygonHierarchyResolverOptions{
		Database: spatial_db,
		Roles: []string{
			"common",
		},
	}

	resolver, err := hierarchy.NewPointInPolygonHierarchyResolver(ctx, resolver_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new PIP resolver, %w", err)
	}

	inputs := &spatial_filter.SPRInputs{
		IsCurrent: []int64{1},
	}

	results_cb := hierarchy_filter.FirstButForgivingSPRResultsFunc
	update_cb := hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()

	// Set up writers

	writers := make(map[string]io.WriteCloser)

	mu := new(sync.RWMutex)

	// Set up timer

	monitor, err := timings.NewMonitor(ctx, "counter://PT60S")

	if err != nil {
		return fmt.Errorf("Failed to create new monitor, %w", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	// Walk Overture records

	walk_cb := func(ctx context.Context, uri string, r *walk.WalkRecord) error {

		// t1 := time.Now()

		defer func() {
			// log.Printf("Time to process '%s' %d, %v\n", r.Path, r.LineNumber, time.Since(t1))
			go monitor.Signal(ctx)
		}()

		body, err := sjson.SetBytes(r.Body, "properties.wof:placetype", wof_placetype)

		if err != nil {
			return fmt.Errorf("Failed to assign placetype, %w", err)
		}

		// t1 := time.Now()

		_, body, err = resolver.PointInPolygonAndUpdate(ctx, inputs, results_cb, update_cb, body)

		if err != nil {
			return fmt.Errorf("Failed to update record, %w", err)
		}

		// logger.Printf("Time to PIP ... %v\n", time.Since(t1))

		fname := filepath.Base(uri)

		mu.Lock()
		defer mu.Unlock()

		wr, exists := writers[fname]

		if !exists {

			new_wr, err := target_bucket.NewWriter(ctx, fname, nil)

			if err != nil {
				return fmt.Errorf("Failed to create new writer for %s, %w", fname, err)
			}

			wr = new_wr
			writers[fname] = wr
		}

		_, err = wr.Write(body)

		if err != nil {
			return fmt.Errorf("Failed to write record to %s, %w", fname, err)
		}

		return nil
	}

	walk_opts := &geojsonl.WalkOptions{
		SourceBucket: source_bucket,
		Callback:     walk_cb,
		IsBzipped:    is_bzipped,
	}

	err = geojsonl.Walk(ctx, walk_opts, uris...)

	if err != nil {
		return fmt.Errorf("Failed to wal, %w", err)
	}

	for fname, wr := range writers {

		err := wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close writer for %s, %w", fname, err)
		}
	}

	return nil
}
