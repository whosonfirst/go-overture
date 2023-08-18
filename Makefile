GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/to-country-jsonl cmd/to-country-jsonl/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/to-wof cmd/to-wof/main.go
