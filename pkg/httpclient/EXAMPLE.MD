```go
func main() {
    client := httpclient.NewClient(
        httpclient.WithBaseURL("https://api.example.com/"),
        httpclient.WithTimeout(10*time.Second),
        httpclient.WithDefaultHeader("Authorization", "Bearer token"),
        httpclient.WithRetry(3, 500*time.Millisecond),
        httpclient.WithCache(httpclient.NewInMemoryCache()),
    )

    ctx := context.Background()

    // GET avec query params
    resp, err := client.Get(ctx, "users", &httpclient.RequestOptions{
        QueryParams: url.Values{"page": []string{"2"}},
        Headers: http.Header{"X-Request-ID": []string{"123"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    // Lire le corps...

    // POST JSON
    resp, err = client.Post(ctx, "users", &httpclient.RequestOptions{
        Body: map[string]interface{}{"name": "Alice"},
    })

    // POST form-urlencoded
    resp, err = client.Post(ctx, "login", &httpclient.RequestOptions{
        Form: url.Values{"username": []string{"alice"}, "password": []string{"secret"}},
    })

    // POST multipart
    file, _ := os.Open("avatar.png")
    defer file.Close()
    resp, err = client.Post(ctx, "upload", &httpclient.RequestOptions{
        Multipart: &httpclient.MultipartData{
            Fields: map[string]string{"description": "avatar"},
            Files:  map[string]io.Reader{"avatar.png": file},
        },
    })

    // Gestion d'erreur HTTP
    if httpErr, ok := err.(*httpclient.HTTPError); ok {
        fmt.Printf("Erreur %d: %s\n", httpErr.StatusCode, httpErr.Body)
    }
}
```