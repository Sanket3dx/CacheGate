package main

import (
	cacheoperation "CacheGate/cache_operation"
	"CacheGate/proxy"
	"flag"
	"fmt"
	"os"
)

func init() {
	fmt.Println("#################---CacheGate---###################")
}

func main() {
	// Define command-line arguments
	configPath := flag.String("config", "", "Path to config file (required)")

	// Parse command-line arguments
	flag.Parse()
	if *configPath == "" {
		fmt.Println("Error: Missing config file path")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration from the provided file path
	config, err := proxy.LoadConfigFromFile(*configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	db := cacheoperation.OpenDB("cacheGate")

	// Start garbage collector in a separate goroutine
	go cacheoperation.CacheGarbageCollector(db)

	// Start the proxy with the loaded config
	proxy.StartProxy(db, config)
}
