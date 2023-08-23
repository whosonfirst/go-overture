# go-overture

Go package for working Overture Maps data.

## Important

Work in progress. Documentation may still be incomplete.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/to-country-jsonl cmd/to-country-jsonl/main.go
go build -mod readonly -ldflags="-s -w" -o bin/append-wof cmd/append-wof/main.go
```

### append-wof

`to-country-jsonl` iterates through a collection of Overture data records exported as line-separated GeoJSON files and performs a Who's On First point-in-polygon operation on each record and updating it with `wof:parent_id`, `wof:hierarchy` and `wof:placetype` properties before re-exporting it to a new line-separted GeoJSON file (named `overture-{COUNTRYCODE}.geojsonl`).

```
$> ./bin/append-wof -h
  -index-spatial-database
    	Create a point-in-polygon enabled spatial index at runtime. If true then both -iterator-uri and -iterator-source must be set. (default true)
  -iterator-source value
    	Zero or more URIs for the iterator defined by -iterator-uri to process.
  -iterator-uri string
    	A valid whosonfirst/go-whosnfirst-iterate/v2 URI. (default "git:///tmp")
  -source-bucket-uri string
    	A valid GoCloud blob URI where Overture JSONL files are read from. (default "file:///")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial.SpatialDatabase URI. (default "sqlite://?dsn=modernc://mem")
  -target-bucket-uri string
    	A valid GoCloud blob URI where Overture JSONL sorted-by-country files are written to. (default "file:///")
  -whosonfirst-placetype string
    	The Who's On First placetype to assign to each Overture record being processed.
```

For example:

```
$> ./bin/append-wof \
	-target-bucket-uri file:///usr/local/data/overture/places-wof \
	-whosonfirst-placetype venue \
	-spatial-database-uri 'sqlite://?dsn=modernc://mem' \
	-index-spatial-database \
	-iterator-uri git:///tmp \
	-iterator-source https://github.com/whosonfirst-data/whosonfirst-data-admin-ca.git \
	/usr/local/data/overture/places-country/overture-CA.jsonl
	
2023/08/22 13:49:13 time to index paths (1) 1m43.425295458s

processed 2897 records in 1m0.00184275s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 5744 records in 2m0.000731167s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 8602 records in 3m0.0049925s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 11399 records in 4m0.001942959s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)

...time passes

processed 1299042 records in 7h23m30.000535792s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 1300522 records in 7h24m0.000620584s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 1302024 records in 7h24m30.000856417s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
processed 1303501 records in 7h25m0.003486084s (started 2023-08-22 13:49:13.051465 -0700 PDT m=+103.466224126)
```

So, 7.5 hours for 1.3M records on a modern laptop which isn't as fast as I'd like but it gets the job done. If you're planning on doing _all_ the Overture places records it is probably best to fan out all the records (by country) across multiple cloud-based computers (EC2, etc).

The results will look like this:

```
$> less /usr/local/data/overture/places-wof/overture-CA.jsonl

{ "type": "Feature", "properties": { "id": "tmp_8BED04D50D81A8AB362C97894783BE85", "updatetime": "2023-07-24T00:00:00.000", "version": 0, "confidence": 0.40360552072525024, "websites": null, "social": [ "https:\/\/www.facebook.com\/156802458499052" ], "emails": null, "brand": { "names": null, "wikidata": null }, "addresses": [ { "locality": "Edmonton", "postcode": "T6T 1H6", "freeform": "3704 29 St NW", "region": "AB", "country": "CA" } ], "categories": { "main": "hotel", "alternate": [ "accommodation" ] }, "sources": [ { "dataset": "meta", "property": "", "recordid": "156802458499052" } ] ,"wof:placetype":"venue","wof:parent_id":1108972447,"wof:country":"CA","wof:hierarchy":[{"borough_id":1108971193,"continent_id":102191575,"country_id":85633041,"county_id":1511799791,"locality_id":890458485,"macrohood_id":1108970597,"neighbourhood_id":1108972447,"region_id":85682091}]}, "geometry": { "type": "Point", "coordinates": [ -113.38536, 53.47245 ] } }
{ "type": "Feature", "properties": { "id": "tmp_C49D0841DB7C06E033542D3D8C637B4C", "updatetime": "2023-07-24T00:00:00.000", "version": 0, "confidence": 0.31127199530601501, "websites": [ "http:\/\/myanmareno.com" ], "social": [ "https:\/\/www.facebook.com\/249884631753075" ], "emails": null, "brand": { "names": null, "wikidata": null }, "addresses": [ { "locality": "Vancouver", "postcode": "V5S 2E4", "freeform": "2615 Hoylake Ave", "region": "BC", "country": "CA" } ], "categories": { "main": "home_improvement_store", "alternate": [ "car_rental_agency" ] }, "sources": [ { "dataset": "meta", "property": "", "recordid": "249884631753075" } ] ,"wof:placetype":"venue","wof:parent_id":85865083,"wof:country":"CA","wof:hierarchy":[{"continent_id":102191575,"country_id":85633041,"county_id":890457467,"locality_id":101741075,"neighbourhood_id":85865083,"region_id":85682117}]}, "geometry": { "type": "Point", "coordinates": [ -123.0547, 49.21636 ] } }

...and so on
```

#### Spatial databases (and indices)

By default the application is configured to build a spatial index used for point-in-polygon operations in-memory. This index is derived from data provided by an "iterator" described below. It is also possible to build your own local SQLite spatial database using the tools provided by the [whosonfirst/go-whosonfirst-sqlite-features-index](https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index) package. For example to build a spatial database of Who's On First data for the US you might do something like this:

```
$> ./bin/wof-sqlite-index-features \
	-database-uri modernc:///usr/local/data/us.db \
	-spatial-tables \
	-iterator-uri 'repo://' \
	/usr/local/data/whosonfirst-data-admin-us
```

And then reference it in the `wof-append` command like this:

```
$> bin/wof-append \
	-spatial-database-uri 'sqlite://?dsn=modernc:///usr/local/data/us.db'
```

#### Iterator URIs (and sources)

If you are building a spatial index at runtime the default "iterator" (code that processes one or more Who's On First records) is configured to fetch data from a Git repository. For example:

```
$> ./bin/append-wof \
	-index-spatial-database \
	-iterator-uri git:///tmp \
	-iterator-source https://github.com/whosonfirst-data/whosonfirst-data-admin-ca.git \
```

For details on using other "iterators" for loading data please consult the [whosonfirst/go-whosonfirst-iterate](https://github.com/whosonfirst/go-whosonfirst-iterate) documentation.

### to-country-jsonl

`to-country-jsonl` iterates through a collection of Overture data records exported as line-separated GeoJSON files and re-exports each record to a per-country line-separted GeoJSON file (named `overture-{COUNTRYCODE}.geojsonl`).

```
$> ./bin/to-country-jsonl -h
  -source-bucket-uri string
    	A valid GoCloud blob URI where Overture GeoJSONL files are read from. (default "file:///")
  -target-bucket-uri string
    	A valid GoCloud blob URI where Overture GeoJSONL sorted-by-country files are written to. (default "file:///")
```

For example:

```
$> bin/to-country-jsonl \
	-target-bucket-uri file:///usr/local/data/overture/places-country \
	/usr/local/data/overture/places-geojson/*.geojsonl
	
2023/08/17 10:13:07 Process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_01c760ca-02aa-4387-8b71-b2eaa6c7c700.geojsonl'...
2023/08/17 10:13:16 Time to process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_01c760ca-02aa-4387-8b71-b2eaa6c7c700.geojsonl', 9.069093375s
2023/08/17 10:13:16 Process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_023fe3f2-d72a-40b6-9eb9-4bb1b61664d6.geojsonl'...
2023/08/17 10:13:26 Time to process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_023fe3f2-d72a-40b6-9eb9-4bb1b61664d6.geojsonl', 10.06977625s
2023/08/17 10:13:26 Process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_15b8943c-63b7-45c8-99fd-82a63affb530.geojsonl'...
...
2023/08/17 10:17:46 Process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_d2488fa7-c51b-4fca-b6f4-168af8fbf9fa.geojsonl'...
2023/08/17 10:17:56 Time to process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_d2488fa7-c51b-4fca-b6f4-168af8fbf9fa.geojsonl', 10.100867625s
2023/08/17 10:17:56 Process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_fa803010-a326-4119-8d5a-c4d9173205a7.geojsonl'...
2023/08/17 10:18:06 Time to process 'usr/local/data/overture/places-geojson/20230725_210643_00079_ayc64_fa803010-a326-4119-8d5a-c4d9173205a7.geojsonl', 9.804213208s
2023/08/17 10:18:06 Time to process all files, 4m58.9891885s
```

_For details on how to create a collection of Overture data records exported as line-separated JSON files consult the [Exporting Overture parquet files to line-separated JSON](#exporting-overture-parquet-files-to-line-separated-json) section below._

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

This assumes that you have installed duckdb with the `SPATIAL` extension enabled and that both `/usr/local/data/overture/places` and `/usr/local/data/overture/places-geojson` exist and that the Overture "places" parquet files have been downloaded in to the former. See also: https://github.com/OvertureMaps/data#3-duckdb-sql

```
#!/bin/sh

for f in /usr/local/data/overture/places/*
do
    f=`basename $f`
    echo "process $f"
    duckdb -c "LOAD spatial;COPY (SELECT id, updatetime, version, confidence, JSON(websites) AS websites, JSON(socials) AS social, JSON(emails) AS emails, JSON(brand) AS brand, JSON(addresses) AS addresses, JSON(categories) AS categories, JSON(sources) AS sources, ST_GeomFromWkb(geometry) AS geometry FROM read_parquet('/usr/local/data/overture/places/${f}', filename=true, hive_partitioning=1)) TO '/usr/local/data/overture/places-geojson/${f}.geojsonl' WITH (FORMAT GDAL, DRIVER 'GeoJSONSeq');"
done
```

## See also

* https://github.com/OvertureMaps/data
* https://gocloud.dev/howto/blob/
* https://github.com/aaronland/go-jsonl