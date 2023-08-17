# go-overture

Go package for working Overture Maps data.

## Important

Work in progress. Documentation may still be incomplete.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/to-country-jsonl cmd/to-country-jsonl/main.go
```

### to-country-jsonl

```
$> ./bin/to-country-jsonl -h
  -source-bucket-uri string
    	A valid GoCloud blob URI where Overture JSONL files are read from. (default "file:///")
  -target-bucket-uri string
    	A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to. (default "file:///")
```

`to-country-jsonl` iterates through a collection of Overture data records exported as line-separated JSON files and re-exports each record to a per-country line-separted JSON file (named `overture-{COUNTRYCODE}.json`).

For example:

```
$> bin/to-country-jsonl \
	-target-bucket-uri file:///usr/local/data/overture/places-country \
	/usr/local/data/overture/places-geojson/*.geojsonl
	
2023/08/16 18:59:31 Process usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_01c760ca-02aa-4387-8b71-b2eaa6c7c700.geojsonl
2023/08/16 19:00:21 Process usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_023fe3f2-d72a-40b6-9eb9-4bb1b61664d6.geojsonl
2023/08/16 19:01:18 Process usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_15b8943c-63b7-45c8-99fd-82a63affb530.geojsonl
... and so on
```

_For details on how to create a collection of Overture data records exported as line-separated JSON files consult the [Exporting Overture parquet files to line-separated JSON](#) section below.

#### Writing data to remote locations

Under the hood this package uses the [gocloud.dev `Blob`](https://gocloud.dev/howto/blob/) abstraction layer for reading and writing files. By default only [the local filesystem](https://gocloud.dev/howto/blob/#local) is supported. If you need to read or write data from another source you will need to clone the [cmd/to-country-jsonl](cmd/to-country-jsonl/main.go) code and add the relevant driver. For example here is how you would add support to [read and write from S3](https://gocloud.dev/howto/blob/#s3):

```
package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-overture/app/country"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"	
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := country.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run application, %v", err)
	}
}
```

Specifically, the only thing that changes is the addition of the `_ "gocloud.dev/blob/s3blob"` import statement.

## Exporting Overture parquet files to line-separated JSON

```
#!/bin/sh

# This assumes that you have installed duckdb and that both /usr/local/data/overture/places and
# /usr/local/data/overture/places-geojson exist and that the Overture "places" parquet files have
# been downloaded in to the former. See also: https://github.com/OvertureMaps/data#3-duckdb-sql

for f in /usr/local/data/overture/places/*
do
    f=`basename $f`
    echo "process $f"
    duckdb -c "LOAD spatial;COPY (SELECT id, updatetime, version, confidence, JSON(websites) AS websites, JSON(socials) AS social, JSON(emails) AS emails, JSON(brand) AS brand, JSON(addresses) AS addresses, JSON(categories) AS categories, JSON(sources) AS sources, ST_GeomFromWkb(geometry) AS geometry FROM read_parquet('/usr/local/data/overture/places/${f}', filename=true, hive_partitioning=1)) TO '/usr/local/data/overture/places-geojson/${f}.geojsonl' WITH (FORMAT GDAL, DRIVER 'GeoJSONSeq');"
done
```

## See also

* https://github.com/OvertureMaps/data
* https://github.com/aaronland/go-jsonl