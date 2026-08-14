
# Go HTTP Client

A flexible and feature-rich HTTP client for Go, built on top of `net/http`. It provides a clean API for making HTTP requests with support for retries, caching, multipart forms, query parameters, custom headers, and robust error handling.

## Features

- **All HTTP methods**: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- **Contextual requests** using `context.Context`
- **Configurable base URL** and default headers
- **Query parameters** support
- **Multiple body formats**:
  - JSON (automatic marshalling)
  - Form URL-encoded (`application/x-www-form-urlencoded`)
  - Multipart form data (for file uploads)
- **Retry mechanism** with exponential backoff
- **Optional in-memory caching** for GET requests
- **Typed error handling** with status code and response body
- **Safe for concurrent use** (reuses connections via `http.Client`)

## Installation

```bash
go get github.com/nivekalara237/httpclient
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/yourusername/httpclient"
)

func main() {
    // Create a client with base URL and timeout
    client := httpclient.NewClient(
        httpclient.WithBaseURL("https://api.example.com/"),
        httpclient.WithTimeout(10*time.Second),
    )

    ctx := context.Background()

    // Simple GET request
    resp, err := client.Get(ctx, "users", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Process response...
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

## Configuration Options

You can configure the client using functional options:

| Option | Description |
|--------|-------------|
| `WithBaseURL(rawURL string)` | Sets the base URL for all requests. All paths will be resolved relative to this URL. |
| `WithTimeout(d time.Duration)` | Sets the overall timeout for each HTTP request. |
| `WithDefaultHeader(key, value string)` | Adds a default header to every request. |
| `WithRetry(maxRetries int, backoff time.Duration)` | Configures retry policy. |
| `WithCache(cache Cache)` | Sets a cache implementation for GET requests. An in-memory cache is provided. |

Example:

```go
client := httpclient.NewClient(
    httpclient.WithBaseURL("https://api.example.com/v1/"),
    httpclient.WithTimeout(15*time.Second),
    httpclient.WithDefaultHeader("Authorization", "Bearer YOUR_TOKEN"),
    httpclient.WithDefaultHeader("User-Agent", "MyApp/1.0"),
    httpclient.WithRetry(3, 500*time.Millisecond),
    httpclient.WithCache(httpclient.NewInMemoryCache()),
)
```

## Making Requests

All request methods accept:

- `ctx context.Context`
- `path string` (relative to base URL if set)
- `opts *RequestOptions` (optional, can be `nil`)

`RequestOptions` struct:

```go
type RequestOptions struct {
    QueryParams url.Values      // URL query parameters
    Headers     http.Header     // Additional headers for this request
    Body        any             // JSON body (will be marshalled automatically)
    Form        url.Values      // Form URL-encoded data
    Multipart   *MultipartData  // Multipart form data
}
```

### GET with Query Parameters

```go
resp, err := client.Get(ctx, "search", &httpclient.RequestOptions{
    QueryParams: url.Values{
        "q":    []string{"golang"},
        "page": []string{"2"},
    },
    Headers: http.Header{
        "X-Request-ID": []string{"abc-123"},
    },
})
```

### POST JSON

```go
resp, err := client.Post(ctx, "users", &httpclient.RequestOptions{
    Body: map[string]interface{}{
        "name": "Alice",
        "age":  30,
    },
})
```

### POST Form URL-encoded

```go
resp, err := client.Post(ctx, "login", &httpclient.RequestOptions{
    Form: url.Values{
        "username": []string{"alice"},
        "password": []string{"secret"},
    },
})
```

### POST Multipart (File Upload)

```go
file, _ := os.Open("avatar.png")
defer file.Close()

resp, err := client.Post(ctx, "upload", &httpclient.RequestOptions{
    Multipart: &httpclient.MultipartData{
        Fields: map[string]string{
            "description": "Profile picture",
        },
        Files: map[string]io.Reader{
            "avatar.png": file,
        },
    },
})
```

### Other Methods

```go
client.Put(ctx, "users/1", &httpclient.RequestOptions{Body: user})
client.Patch(ctx, "users/1", &httpclient.RequestOptions{Body: patchData})
client.Delete(ctx, "users/1", nil)
client.Head(ctx, "users/1", nil)
client.Options(ctx, "users", nil)
```

## Error Handling

The client distinguishes between network errors and HTTP error responses. For HTTP status codes >= 400, you can use the helper method `checkResponse` (or manually check). However, the `Do` method returns the raw `*http.Response`; the caller is responsible for checking status codes. For convenience, you can use `checkResponse` as shown below, or implement your own logic.

Example of custom error handling using the provided `HTTPError` type:

```go
resp, err := client.Get(ctx, "users", nil)
if err != nil {
    // Handle network error or context cancellation
    log.Fatal(err)
}
defer resp.Body.Close()

if resp.StatusCode >= 400 {
    body, _ := io.ReadAll(resp.Body)
    httpErr := &httpclient.HTTPError{
        StatusCode: resp.StatusCode,
        Status:     resp.Status,
        Body:       body,
    }
    log.Printf("HTTP error: %v", httpErr)
}
```

Alternatively, you can call `client.checkResponse(resp)` after each request; it will close the body and return an `*HTTPError` if status is >= 400.

## Retry Logic

Retries are configured globally via `WithRetry`. By default, no retries are performed. When enabled, the client will retry requests that:

- Return a network error (connection refused, timeout, etc.)
- Return HTTP status codes: 429, 500, 502, 503, 504

Retries use exponential backoff with the provided base duration. The formula is:

```
wait = backoff * 2^(attempt-1)
```

You can customize which responses should be retried by setting `RetryOn` in `RetryConfig` (not exposed via option, but you can modify the client after creation).

## Caching

Caching is optional and only applies to **GET requests**. The client caches responses with status code 200. The cache key is the full request URL (including query parameters). Cached responses are stored for 5 minutes by default (currently not configurable, but you can modify the `InMemoryCache` or implement your own `Cache` interface).

To enable caching, pass `WithCache(httpclient.NewInMemoryCache())` when creating the client.

**Important**: The simple in-memory cache does not consider request headers (like `Authorization`). It is suitable for single-user or public data. For multi-user scenarios, implement a cache key that includes relevant headers or use a different caching strategy.

## Context and Cancellation

Every request method accepts a `context.Context`. Use it to set deadlines, cancel requests, or pass values.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.Get(ctx, "slow-endpoint", nil)
if err != nil {
    // May be context.DeadlineExceeded
}
```

## Concurrency Safety

The `Client` is safe for concurrent use by multiple goroutines. It internally uses a shared `http.Client` which manages connection pooling. You can create one client and reuse it across your application.

## Custom Cache Implementation

If you want to use Redis, Memcached, or another cache backend, implement the `Cache` interface:

```go
type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte, ttl time.Duration)
}
```

Then pass it via `WithCache`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT
```