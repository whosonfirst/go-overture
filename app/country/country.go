package country

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/tidwall/gjson"
	"gocloud.dev/blob"
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
		return fmt.Errorf("Failed to open target bucket, %w", err)
	}

	defer target_bucket.Close()

	writers := make(map[string]io.WriteCloser)

	mu := new(sync.RWMutex)

	for _, uri := range uris {

		uri = strings.TrimLeft(uri, "/")

		fh, err := source_bucket.NewReader(ctx, uri, nil)

		if err != nil {
			return fmt.Errorf("Failed to open reader for '%s', %v", uri, err)
		}

		defer fh.Close()

		logger.Println("Process", uri)

		err = walkReader(ctx, fh, target_bucket, writers, mu)

		if err != nil {
			return fmt.Errorf("Failed to walk %s, %v", uri, err)
		}

		break
	}

	for country, wr := range writers {

		err := wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close writer for %s, %v", country, err)
		}
	}

	return nil
}

func walkReader(ctx context.Context, r io.Reader, target_bucket *blob.Bucket, writers map[string]io.WriteCloser, mu *sync.RWMutex) error {

	var walk_err error

	record_ch := make(chan *walk.WalkRecord)
	error_ch := make(chan *walk.WalkError)
	done_ch := make(chan bool)

	go func() {

		for {
			select {
			case <-ctx.Done():
				done_ch <- true
				return
			case err := <-error_ch:
				walk_err = err
				done_ch <- true
			case r := <-record_ch:

				country_rsp := gjson.GetBytes(r.Body, "properties.addresses.0.country")

				country := country_rsp.String()

				if country == "" {
					country = "XX"
				}

				fname := fmt.Sprintf("overture-%s.jsonl", country)

				mu.Lock()

				_, exists := writers[country]

				if !exists {

					country_wr, err := target_bucket.NewWriter(ctx, fname, nil)

					if err != nil {
						mu.Unlock()

						walk_err = fmt.Errorf("Failed to create writer for %s, %w", fname, err)
						done_ch <- true
						break
					}

					writers[country] = country_wr
				}

				_, err := writers[country].Write(r.Body)

				if err != nil {
					mu.Unlock()

					walk_err = fmt.Errorf("Failed to write body for %s, %w", fname, err)
					done_ch <- true
					break
				}

				mu.Unlock()
			}
		}
	}()

	workers := runtime.NumCPU() * 2

	walk_opts := &walk.WalkOptions{
		RecordChannel: record_ch,
		ErrorChannel:  error_ch,
		DoneChannel:   done_ch,
		Workers:       workers,
	}

	go walk.WalkReader(ctx, walk_opts, r)

	<-done_ch

	if walk_err != nil && !walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}
