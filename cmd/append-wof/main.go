package main

/*
 go run -mod vendor cmd/to-wof/main.go -spatial-database-uri 'sqlite://?dsn=modernc:///usr/local/data/us.db' /usr/local/data/overture/places-country/overture-US.geojsonl

> ./bin/to-wof -spatial-database-uri 'sqlite://?dsn=modernc://mem' -index-spatial-database -iterator-uri repo:// -iterator-source /usr/local/data/whosonfirst-data-admin-ca -target-bucket-uri file:///usr/local/data/overture/places-wof/ /usr/local/data/overture/places-country/overture-CA.jsonl

*/

import (
	"context"
	"log"

	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/whosonfirst/go-overture/app/append"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := append.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run application, %v", err)
	}
}
