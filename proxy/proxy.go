package proxy

import (
	cacheoperation "CacheGate/cache_operation"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

type Config struct {
	Port        string   `json:"port"`
	RemoteURL   string   `json:"remote_url"`
	TTl         int      `json:"ttl"`
	UrlsToCache []string `json:"urls_to_cache"`
}

func LoadConfigFromFile(path string) (Config, error) {
	var config Config
	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}
	return config, nil
}

func matchPattern(url, pattern string) bool {
	// Handle wildcard match (*)
	if strings.HasSuffix(pattern, "/*") {
		basePattern := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(url, basePattern)
	}
	return url == pattern
}

func ShouldCacheURL(url string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(url, pattern) {
			return true
		}
	}
	return false
}

func CacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func ProxyHandler(db *badger.DB, config Config) http.HandlerFunc {
	target, err := url.Parse(config.RemoteURL)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
	}
	ttl := time.Duration(config.TTl)
	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := CacheKey(r.URL.String())
		fmt.Printf("Server hit: Request  🗄️ -> %s \n", r.URL.String())
		if r.Method == "GET" && ShouldCacheURL(r.URL.Path, config.UrlsToCache) {
			// Check if the response is already cached
			if cachedResponse, err := cacheoperation.Get(db, cacheKey); err == nil {
				fmt.Printf("Cache hit: Serving from cache 🗄️ -> %s \n", r.URL.String())
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
			if r.Method == "GET" && ShouldCacheURL(r.URL.Path, config.UrlsToCache) {
				fmt.Println(r.URL.String())
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

func StartProxy(db *badger.DB, config Config) {
	_, err := url.Parse(config.RemoteURL)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
		return
	}
	http.HandleFunc("/", ProxyHandler(db, config))
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
