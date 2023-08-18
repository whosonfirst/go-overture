package main

import (
	"context"
	"log"

	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/whosonfirst/go-overture/app/reversegeo"
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
