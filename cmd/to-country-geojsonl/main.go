package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	_ "os"
	"strings"
	"sync"

	"github.com/aaronland/go-jsonl/walk"
	"github.com/aaronland/gocloud-blob/bucket"
	"github.com/tidwall/gjson"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	bucket_uri := flag.String("bucket-uri", "file:///", "A valid GoCloud blob URI.")
	target_bucket_uri := flag.String("target-bucket-uri", "file:///", "A valid GoCloud blob URI.")

	flag.Parse()

	uris := flag.Args()
	ctx := context.Background()

	source_bucket, err := bucket.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer source_bucket.Close()

	target_bucket, err := bucket.OpenBucket(ctx, *target_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open target bucket, %w", err)
	}

	defer target_bucket.Close()

	writers := make(map[string]io.WriteCloser)

	mu := new(sync.RWMutex)

	for _, uri := range uris {

		uri = strings.TrimLeft(uri, "/")

		fh, err := source_bucket.NewReader(ctx, uri, nil)

		if err != nil {
			log.Fatalf("Failed to open %s, %v", uri, err)
		}

		defer fh.Close()

		log.Println("Process", uri)
		err = walkReader(ctx, fh, target_bucket, writers, mu)

		if err != nil {
			log.Fatalf("Failed to walk %s, %v", uri, err)
		}
	}

	for country, wr := range writers {

		err := wr.Close()

		if err != nil {
			log.Fatalf("Failed to close writer for %s, %v", country, err)
		}
	}
}

func walkReader(ctx context.Context, r io.Reader, target_bucket *blob.Bucket, writers map[string]io.WriteCloser, mu *sync.RWMutex) error {

	// ctx, cancel := context.WithCancel(ctx)
	//defer cancel()

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

				fname := fmt.Sprintf("overture-places-%s.jsonl", country)

				mu.Lock()

				_, exists := writers[country]

				// log.Println(fname, wr)

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

				var f map[string]interface{}

				err := json.Unmarshal(r.Body, &f)

				if err != nil {
					mu.Unlock()

					walk_err = fmt.Errorf("Failed to unmarshal record to %s, %w", fname, err)
					done_ch <- true
					break

				}

				enc, err := json.Marshal(f)

				if err != nil {
					mu.Unlock()

					walk_err = fmt.Errorf("Failed to marshal record to %s, %w", fname, err)
					done_ch <- true
					break
				}

				_, err = writers[country].Write(enc)

				if err != nil {
					mu.Unlock()

					walk_err = fmt.Errorf("Failed to write newline for %s, %w", fname, err)
					done_ch <- true
					break
				}

				writers[country].Write([]byte("\n"))

				mu.Unlock()
				//log.Println("OK", fname, wr)
			}
		}
	}()

	walk_opts := &walk.WalkOptions{
		RecordChannel: record_ch,
		ErrorChannel:  error_ch,
		DoneChannel:   done_ch,
		Workers:       10,
		FormatJSON:    true,
	}

	go walk.WalkReader(ctx, walk_opts, r)

	<-done_ch

	if walk_err != nil && !walk.IsEOFError(walk_err) {
		return fmt.Errorf("Failed to walk document, %v", walk_err)
	}

	return nil
}
