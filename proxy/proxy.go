package proxy

import (
	cacheoperation "CacheGate/cache_operation"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

func CacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func ProxyHandler(db *badger.DB, target *url.URL, ttl time.Duration) http.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := CacheKey(r.URL.String())

		if r.Method == "GET" {
			// Check if the response is already cached
			if cachedResponse, err := cacheoperation.Get(db, cacheKey); err == nil {
				fmt.Println("Cache hit: Serving from cache 🗄️")
				w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
				w.Write(cachedResponse.Data)
				return
			}
		}
		r.Header.Set("Host", target.Host)

		// If not cached, forward the request to the target server
		proxy.ModifyResponse = func(response *http.Response) error {
			bodyBytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}

			// Cache the response if it's a GET request
			if r.Method == "GET" {
				cacheItem := cacheoperation.NewCacheItem(bodyBytes, ttl)
				err = cacheoperation.Set(db, cacheKey, *cacheItem)
				if err != nil {
					fmt.Printf("Error caching response: %v\n", err)
				}
			}

			// Reconstruct the response body after caching
			response.Body = ioutil.NopCloser(strings.NewReader(string(bodyBytes)))
			response.ContentLength = int64(len(bodyBytes))
			return nil
		}

		// Forward the request to the target server
		proxy.ServeHTTP(w, r)
	}
}

func StartProxy(db *badger.DB, targetURL string, ttl time.Duration) {
	target, err := url.Parse(targetURL)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
		return
	}
	http.HandleFunc("/", ProxyHandler(db, target, ttl))
	log.Fatal(http.ListenAndServe(":8091", nil))
}
