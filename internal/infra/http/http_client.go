package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/logging"
	"sonarbridge-go/pkg/httpclient"
	"strings"
	"time"
)

type IClientHttp interface {
	Get(ctx context.Context, url string, params url.Values) (any, error)
	Post(ctx context.Context, url string, params url.Values, payload string) (any, error)
}

type ClientType string

type OptionParamItem struct {
	Key   string
	Value any
}
type Options struct {
	Params      []OptionParamItem
	QueryParams []OptionParamItem
	Headers     []OptionParamItem
}

const (
	gitlab    ClientType = "gitlab"
	sonarqube ClientType = "sonar"
	bitbucket ClientType = "bb"
)

type Client struct {
	baseUrl    string
	token      string
	clientFor  ClientType
	httpClient *http.Client
}

func NewClientHttp(baseUrl, token string, cType ClientType, caCertPath string) *Client {
	if baseUrl == "" {
		return nil
	}

	var transport *http.Transport

	if caCertPath != "" {
		pool, err := loadCAPool(caCertPath)
		if err != nil {
			panic(err)
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS12,
			},
		}
	}

	return &Client{
		baseUrl:   baseUrl,
		token:     token,
		clientFor: cType,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, out any) *httpclient.ClientError {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		panic(fmt.Errorf("création de la requête %s, %w", path, err))
	}
	if c.clientFor == gitlab {
		req.Header.Set("Content-Type", "application/json")
		if strings.Count(c.token, "ci;") > 0 {
			tt := strings.Split(c.token, ";")
			req.Header.Set("JOB-TOKEN", tt[1])
		} else {
			req.Header.Set("Private-Token", c.token)
		}
	}

	if c.clientFor == sonarqube {
		req.SetBasicAuth(c.token, "")
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("appel API %s, %s: %w", c.clientFor, path, err))
	}
	defer response.Body.Close()

	if !isSuccess(response.StatusCode) {
		Body, _ := io.ReadAll(response.Body)

		// return fmt.Errorf("%s a repondu %d: %s", path, response.StatusCode, string(Body))
		return &httpclient.ClientError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Body:       Body,
		}
	}

	decoder := json.NewDecoder(response.Body)
	// decoder.DisallowUnknownFields()
	if err5 := decoder.Decode(&out); err5 != nil {
		panic(httpx.ErrInternal)
	}
	return nil
}

func (c *Client) Post(ctx context.Context, path string, params url.Values, payload string, response any) *httpclient.ClientError {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader([]byte(payload)))
	if err != nil {
		panic(fmt.Errorf("création de la requête %s: %w", path, err))
	}
	if c.clientFor == gitlab {
		req.Header.Set("Content-Type", "application/json")
		if strings.Count(c.token, "ci;") > 0 {
			tt := strings.Split(c.token, ";")
			req.Header.Set("JOB-TOKEN", tt[1])
		} else {
			req.Header.Set("Private-Token", c.token)
		}
	}

	if c.clientFor == sonarqube {
		req.SetBasicAuth(c.token, "")
	}

	if c.clientFor == bitbucket {
		logging.Warn("no client is implemented for Bitbucket")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("appel API %s, %s : %w", c.clientFor, path, err))
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		Body, _ := io.ReadAll(resp.Body)
		// return fmt.Errorf("%s a repondu %d : %s", path, resp.StatusCode, string(Body))
		return &httpclient.ClientError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       Body,
		}
	}

	decoder := json.NewDecoder(resp.Body)
	// decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		panic(httpx.ErrInternal)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, path string, params url.Values, payload string, out any) *httpclient.ClientError {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, bytes.NewReader([]byte(payload)))
	if err != nil {
		panic(fmt.Errorf("création de la requête %s: %w", path, err))
	}
	if c.clientFor == gitlab {
		req.Header.Set("Content-Type", "application/json")
		if strings.Count(c.token, "ci;") > 0 {
			tt := strings.Split(c.token, ";")
			req.Header.Set("JOB-TOKEN", tt[1])
		} else {
			req.Header.Set("Private-Token", c.token)
		}
	}

	if c.clientFor == sonarqube {
		req.SetBasicAuth(c.token, "")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("appel API %s, %s : %w", c.clientFor, path, err))
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		Body, _ := io.ReadAll(resp.Body)
		// return fmt.Errorf("%s a répondu %d : %s", path, resp.StatusCode, string(Body))
		return &httpclient.ClientError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       Body,
		}
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		panic(httpx.ErrInternal)
	}

	return nil
}

func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
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
