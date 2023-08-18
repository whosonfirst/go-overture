package main

// go run -mod readonly cmd/to-wof/main.go -spatial-database-uri 'sqlite://?dsn=modernc://mem' /usr/local/data/overture/places-country/

import (
	"context"
	"log"

	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/whosonfirst/go-overture/app/reversegeo"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := reversegeo.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run application, %v", err)
	}
}
