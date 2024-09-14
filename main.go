package main

import (
	cacheoperation "CacheGate/cache_operation"
	"CacheGate/proxy"
	"flag"
	"fmt"
	"os"
	"time"
)

func init() {
	fmt.Println("#################---CacheGate---###################")
}

func main() {
	// Define command-line arguments
	// dbName := flag.String("db", "", "Name of the Badger database (required)")
	targetURL := flag.String("url", "", "Target URL to proxy requests to (required)")
	ttl := flag.Duration("ttl", time.Minute*5, "Time-to-live for cached responses (default: 1 minute)")

	// Parse command-line arguments
	flag.Parse()
	dbName := "cacheGate"
	if dbName == "" || *targetURL == "" {
		fmt.Println("Error: Missing required arguments")
		flag.Usage()
		os.Exit(1)
	}

	db := cacheoperation.OpenDB(dbName)

	// Start garbage collector in a separate goroutine
	go cacheoperation.CacheGarbageCollector(db)

	proxy.StartProxy(db, *targetURL, *ttl)
}
