package geojsonl

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aaronland/go-jsonl/walk"
	"gocloud.dev/blob"
)

type WalkCallbackFunc func(context.Context, string, *walk.WalkRecord) error

type WalkOptions struct {
	SourceBucket *blob.Bucket
	Callback     WalkCallbackFunc
}

func Walk(ctx context.Context, opts *WalkOptions, uris ...string) error {

	for _, uri := range uris {

		err := walkURI(ctx, opts, uri)

		if err != nil {
			return fmt.Errorf("Failed to walk %s, %v", uri, err)
		}

	}

	return nil
}

func walkURI(ctx context.Context, opts *WalkOptions, uri string) error {

	uri = strings.TrimLeft(uri, "/")

	r, err := opts.SourceBucket.NewReader(ctx, uri, nil)

	if err != nil {
		return fmt.Errorf("Failed to open reader for '%s', %v", uri, err)
	}

	defer r.Close()

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

				err := opts.Callback(ctx, uri, r)

				if err != nil {

					walk_err = fmt.Errorf("Failed to invoke callback for %s, %w", r.Path, err)
					done_ch <- true
					break
				}
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
