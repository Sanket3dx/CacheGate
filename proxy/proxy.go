package proxy

import (
	cacheoperation "CacheGate/cache_operation"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

// Config holds proxy configurations.
type Config struct {
	Port              string   `json:"port"`
	RemoteURL         string   `json:"remote_url"`
	TTL               int      `json:"ttl"`
	UrlsToCache       []string `json:"urls_to_cache"`
	ParamsToSkipInKey []string `json:"params_to_skip_in_key"`
}

func LoadConfigFromFile(path string) (Config, error) {
	var config Config
	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
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

// RemoveParams removes specified parameters from the URL.
func RemoveParams(rawURL string, paramsToRemove []string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}
	query := parsedURL.Query()
	for _, param := range paramsToRemove {
		query.Del(param)
	}
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

func ProxyHandler(db *badger.DB, config Config) http.HandlerFunc {
	target, err := url.Parse(config.RemoteURL)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}
	ttl := time.Duration(config.TTL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := CacheKey(RemoveParams(r.URL.String(), config.ParamsToSkipInKey))
		if r.Method == "GET" && ShouldCacheURL(r.URL.Path, config.UrlsToCache) {
			if cachedResponse, err := cacheoperation.Get(db, cacheKey); err == nil {
				fmt.Printf("Cache hit: Serving from cache -> %s\n", RemoveParams(r.URL.String(), config.ParamsToSkipInKey))
				for k, v := range cachedResponse.Headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(http.StatusOK)
				w.Write(cachedResponse.Data)
				return
			}
		}

		proxy.ModifyResponse = func(response *http.Response) error {
			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}
			fmt.Printf("Server Hit: -> %s\n", RemoveParams(r.URL.String(), config.ParamsToSkipInKey))
			response.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

			if r.Method == "GET" && ShouldCacheURL(r.URL.Path, config.UrlsToCache) {
				headers := make(map[string]string)
				for k, v := range response.Header {
					headers[k] = strings.Join(v, ", ")
				}
				cacheItem := cacheoperation.NewCacheItem(bodyBytes, headers, ttl)
				if err := cacheoperation.Set(db, cacheKey, *cacheItem); err != nil {
					log.Printf("Error caching response: %v\n", err)
				}
				fmt.Printf("caching response: -> %s\n", RemoveParams(r.URL.String(), config.ParamsToSkipInKey))
			}
			return nil
		}

		proxy.ServeHTTP(w, r)
	}
}

func StartProxy(db *badger.DB, config Config) {
	http.HandleFunc("/", ProxyHandler(db, config))
	log.Printf("Starting proxy on port %s\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
