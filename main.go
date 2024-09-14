package main

import (
	cacheoperation "CacheGate/cache_operation"
	"fmt"
)

func init() {
	fmt.Println("#################---CacheGate---###################")
}

func main() {
	db := cacheoperation.OpenDB("cacheGate")

	// garbage collector thread
	go cacheoperation.CacheGarbageCollector(db)

}
