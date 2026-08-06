package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type IClientHttp interface {
	Get(ctx context.Context, url string, params url.Values) (any, error)
	Post(ctx context.Context, url string, params url.Values, payload string) (any, error)
}

type ClientType string

const (
	gitlab    ClientType = "gitlab"
	sonarqube ClientType = "sonar"
)

type Client struct {
	baseUrl    string
	token      string
	clientFor  ClientType
	httpClient *http.Client
}

func NewClientHttp(baseUrl, token string, cType ClientType) *Client {
	if baseUrl == "" {
		return nil
	}

	return &Client{
		baseUrl:   baseUrl,
		token:     token,
		clientFor: cType,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, out any) error {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("création de la requête %s, %w", path, err)
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
		return fmt.Errorf("appel API %s, %s: %w", c.clientFor, path, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		Body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("%s a repondu %d: %s", path, response.StatusCode, string(Body))
	}

	decoder := json.NewDecoder(response.Body)
	// decoder.DisallowUnknownFields()
	if err5 := decoder.Decode(&out); err5 != nil {
		return err5
	}
	return nil
}

func (c *Client) Post(ctx context.Context, path string, params url.Values, payload string, response any) error {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader([]byte(payload)))
	if err != nil {
		return fmt.Errorf("création de la requête %s: %w", path, err)
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
		return fmt.Errorf("appel API %s, %s : %w", c.clientFor, path, err)
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		Body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s a repondu %d : %s", path, resp.StatusCode, string(Body))
	}

	decoder := json.NewDecoder(resp.Body)
	// decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return err
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, path string, params url.Values, payload string, out any) error {
	u := c.baseUrl + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, bytes.NewReader([]byte(payload)))
	if err != nil {
		return fmt.Errorf("création de la requête %s: %w", path, err)
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
		return fmt.Errorf("appel API %s, %s : %w", c.clientFor, path, err)
	}
	defer resp.Body.Close()

	if !isSuccess(resp.StatusCode) {
		Body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s a répondu %d : %s", path, resp.StatusCode, string(Body))
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}

	return nil
}

func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
