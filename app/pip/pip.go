package pip

// Walk a JSONL file of Overture (GeoJSON) records and PIP each one outputting the following
// column values to a CSV writer: overture_id,wof_parent_id,wof_repo,wof_country,wof_id
// Note that 'wof_id' will be empty.

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-timings"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-overture/geojsonl"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	spatial_filter "github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"
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

	//

	rdr, _ := reader.NewReader(ctx, "null://")
	var csv_wr *csvdict.Writer

	// Walk Overture records

	walk_cb := func(ctx context.Context, uri string, r *walk.WalkRecord) error {

		// t1 := time.Now()

		defer func() {
			// slog.Info("Time to process", "path", r.Path, "line number", r.LineNumber, "time", time.Since(t1))
			go monitor.Signal(ctx)
		}()

		body, err := sjson.SetBytes(r.Body, "properties.wof:placetype", wof_placetype)

		if err != nil {
			return nil
			return fmt.Errorf("Failed to assign placetype, %w", err)
		}

		// START OF ...

		possible, err := resolver.PointInPolygon(ctx, inputs, body)

		if err != nil {
			return nil
			return err
		}

		parent_spr, err := results_cb(ctx, rdr, body, possible)

		if err != nil {
			return nil
			return err
		}

		if parent_spr == nil {
			slog.Warn("Failed to derive SPR for record")
			return nil
		}

		// slog.Info("PIP DONE", "parent", parent_spr)

		id_rsp := gjson.GetBytes(body, "properties.id")
		parent_id := parent_spr.Id()
		parent_repo := parent_spr.Repo()
		country := parent_spr.Country()
		// belongs_to := parent_spr.BelongsTo()

		out := map[string]string{
			"overture_id":   id_rsp.String(),
			"wof_id":        "",
			"wof_parent_id": parent_id,
			"wof_repo":      parent_repo,
			"wof_country":   country,
		}

		mu.Lock()
		defer mu.Unlock()

		if csv_wr == nil {

			fieldnames := make([]string, 0)

			for k, _ := range out {
				fieldnames = append(fieldnames, k)
			}

			wr, err := csvdict.NewWriter(os.Stdout, fieldnames)

			if err != nil {
				return err
			}

			csv_wr = wr
			csv_wr.WriteHeader()
		}

		csv_wr.WriteRow(out)
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
