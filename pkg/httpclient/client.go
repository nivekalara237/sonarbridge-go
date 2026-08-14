package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sonarbridge-go/pkg/arrays"
	stringify "sonarbridge-go/pkg/string"
	"strings"
	"time"
)

type OptionParamItem struct {
	Key   string
	Value any
}

type MultipartData struct {
	Fields map[string]string
	Files  map[string]io.Reader // filename -> reader
}

type RequestOptions struct {
	QueryParams []OptionParamItem
	Headers     []OptionParamItem
	Body        any        // Pour JSON ou form
	Form        url.Values // pour application/x-www-form-urlencoded
	Multipart   *MultipartData
}

type Option func(*Client)

type RetryConfig struct {
	MaxRetries int
	Backoff    time.Duration // durée de base pour l'exponentiel
	RetryOn    func(resp *http.Response, err error) bool
}

type TLSConfig struct {
	CaCertPath    string
	CaCertPemData string
}

type Client struct {
	headers    http.Header
	httpClient *http.Client
	baseURL    *url.URL
	retry      RetryConfig
	cache      Cache
	Tls        TLSConfig
}

func NewClientHttp(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		headers:    make(http.Header),
		retry: RetryConfig{
			MaxRetries: 0,
			Backoff:    time.Second,
		},
		cache: nil,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithBaseURL(rawURL string) Option {
	return func(c *Client) {
		u, err := url.Parse(rawURL)
		if err != nil {
			panic(err)
		}
		c.baseURL = u
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func WithDefaultHeader(key, value string) Option {
	return func(c *Client) {
		c.headers.Set(key, value)
	}
}

func WithRetry(maxRetries int, backoff time.Duration) Option {
	return func(c *Client) {
		c.retry.MaxRetries = maxRetries
		c.retry.Backoff = backoff
	}
}

func WithCache(cache Cache) Option {
	return func(c *Client) {
		c.cache = cache
	}
}

func WithTLSCACertFile(caCertFile string) Option {
	return func(c *Client) {
		ca, err := loadCAPool(caCertFile)
		if err != nil {
			panic(err)
		}
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    ca,
				MinVersion: tls.VersionTLS12,
			},
		}
	}
}

func WithTLSCAPemData(data string) Option {
	return func(c *Client) {
		ca, err := loadPemData(data)
		if err != nil {
			panic(err)
		}
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    ca,
				MinVersion: tls.VersionTLS12,
			},
		}
	}
}

func (c *Client) checkResponse(resp *http.Response) error {
	return checkResponse(resp)
}

func NewRequestOption(key string, value any) OptionParamItem {
	return OptionParamItem{Key: key, Value: value}
}

func (c *Client) Do(ctx context.Context, method, path string, opts *RequestOptions) (*http.Response, error) {
	// Construction of the complete URL
	reqURL, err := c.buildURL(path, opts)
	if err != nil {
		return nil, err
	}

	// If cache is enabled
	if c.cache != nil && method == http.MethodGet {
		cacheKey := reqURL.String() // inclut les query params
		if data, ok := c.cache.Get(cacheKey); ok {
			// Retourner une réponse simulée avec le corps caché
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(data)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/octet-stream")
			return resp, nil
		}
	}

	//request body
	var body io.Reader
	var contentType string
	if opts != nil {
		switch {
		case opts.Multipart != nil:
			body, contentType, err = encodeMultipart(opts.Multipart)
			if err != nil {
				return nil, err
			}
		case opts.Form != nil:
			body = strings.NewReader(opts.Form.Encode())
			contentType = "application/x-www-form-urlencoded"
		case opts.Body != nil:
			jsonData, err := json.Marshal(opts.Body)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(jsonData)
			contentType = "application/json"
		}
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	// Default headers
	maps.Copy(req.Header, c.headers)
	// The specific headers for request
	if opts != nil && len(opts.Headers) > 0 {
		for _, kv := range opts.Headers {
			if arrays.IsArrayOrSlice(kv.Value) {
				ar_v := make([]string, len(kv.Value.([]any)))
				for _, s := range kv.Value.([]any) {
					ar_v = append(ar_v, stringify.ToString(s))
				}
				req.Header[kv.Key] = ar_v
			} else {
				req.Header[kv.Key] = []string{stringify.ToString(kv.Value)}
			}

		}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Execute with retry
	response, ferr := c.doWithRetry(req)

	if c.cache != nil && ferr == nil && method == http.MethodGet && response.StatusCode == http.StatusOK {
		cacheKey := reqURL.String()
		bodyBytes, _ := io.ReadAll(response.Body)
		response.Body.Close()
		response.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		c.cache.Set(cacheKey, bodyBytes, 5*time.Minute) // TTL configurable
	}

	return response, ferr
}

func encodeMultipart(data *MultipartData) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// text field
	for key, val := range data.Fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}

	for filename, reader := range data.Files {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, "", err
		}
		if _, err = io.Copy(part, reader); err != nil {
			return nil, "", err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}
	return &buf, writer.FormDataContentType(), nil
}

func (c *Client) buildURL(path string, opts *RequestOptions) (*url.URL, error) {
	// Base
	var u *url.URL
	if c.baseURL != nil {
		u = c.baseURL.ResolveReference(&url.URL{Path: path})
	} else {
		var err error
		u, err = url.Parse(path)
		if err != nil {
			return nil, err
		}
	}

	// add queries param
	if opts != nil && len(opts.QueryParams) > 0 {
		q := u.Query()
		for _, kv := range opts.QueryParams {
			if arrays.IsArrayOrSlice(kv.Value) {
				for _, v := range kv.Value.([]any) {
					q.Add(kv.Key, stringify.ToString(v))
				}
			} else {
				q.Add(kv.Key, stringify.ToString(kv.Value))
			}
		}
		u.RawQuery = q.Encode()
	}
	return u, nil
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		// Cloner la requête pour réutiliser le corps (important si body non rejouable)
		reqClone := req.Clone(req.Context())
		if req.Body != nil {
			// Réinitialiser le corps en le relisant depuis le buffer d'origine
			// Ici on suppose que le body est un *bytes.Reader ou *strings.Reader (réutilisable)
			// Pour plus de robustesse, vous pouvez stocker le corps original et le réassigner.
		}

		resp, err := c.httpClient.Do(reqClone)
		if err == nil && !c.shouldRetry(resp) {
			return resp, nil
		}
		if err != nil {
			lastErr = err
		} else {
			// Lire et fermer le corps pour éviter les fuites
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("retryable status: %s", resp.Status)
		}

		if attempt < c.retry.MaxRetries {
			// Backoff exponentiel
			wait := c.retry.Backoff * time.Duration(1<<attempt)
			select {
			case <-time.After(wait):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}
	}
	return nil, fmt.Errorf("request failed after %d retries: %w", c.retry.MaxRetries, lastErr)
}

func (c *Client) shouldRetry(resp *http.Response) bool {
	if c.retry.RetryOn != nil {
		return c.retry.RetryOn(resp, nil)
	}
	// By default, retry on 429, 500, 502, 503, 504
	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

func loadCAPool(certPath string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system cert pool: %w", err)
	}

	pemData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	if ok := pool.AppendCertsFromPEM(pemData); !ok {
		return nil, fmt.Errorf("append CA certificate")
	}

	return pool, nil
}

func loadPemData(pemData string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system cert pool: %w", err)
	}

	if ok := pool.AppendCertsFromPEM([]byte(pemData)); !ok {
		return nil, fmt.Errorf("append CA certificate")
	}

	return pool, nil
}
