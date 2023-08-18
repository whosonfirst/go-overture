package main

// go run -mod vendor cmd/to-wof/main.go -spatial-database-uri 'sqlite://?dsn=modernc:///usr/local/data/us.db' /usr/local/data/overture/places-country/overture-US.geojsonl


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
