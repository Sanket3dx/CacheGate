# CacheGate

CacheGate is a high-performance caching proxy server that accelerates response times and reduces backend load by caching frequently requested data. It intercepts incoming requests, serves cached content when available, and significantly improves overall efficiency. CacheGate is ideal for APIs and web delivery, offering flexible caching and expiration rules for optimized performance.

## Features

- **High-Performance Caching**: Cache frequently accessed content, reducing load on backend servers.
- **Request Interception**: Intercepts requests and serves cached responses if available.
- **Flexible Caching Rules**: Configure specific URL patterns to cache.
- **Custom Expiration (TTL)**: Define custom expiration times for cached content.
- **Simple Configuration**: Set up with a straightforward JSON configuration.

## Use Case

CacheGate is perfect for environments where repeated API requests put a load on the backend servers. By caching responses, CacheGate serves common requests instantly, reducing processing time and improving user experience.

## How It Works

1. CacheGate acts as a proxy server, intercepting HTTP requests.
2. It checks whether a cached version of the requested data is available:
   - If available, CacheGate serves the cached response.
   - If not, it forwards the request to the backend server, caches the response, and serves it to the client.
3. You can define a Time-To-Live (TTL) for how long the content should remain cached.

## Installation

1. Download and install CacheGate.
2. Create a configuration file (JSON) to define caching rules and server settings.

## Config.json sample

```bash
{
    "port": "8091",
    "remote_url": "http://localhost/test",
    "ttl": 5,
    "url_to_cache": [
        "/api/v1/products/*",
        "/api/v1/users/*",
        "/static/*",
        "/assets/images/*",
        "/docs/*"
    ]
}
```

## Example Usage

Run CacheGate with a configuration file:

```bash
CacheGate -config /path/to/config.json

```