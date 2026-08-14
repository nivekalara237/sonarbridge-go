package httpclient

import (
	"context"
	"net/http"
)

func (c *Client) Get(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, opts)
}
func (c *Client) Post(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, opts)
}
func (c *Client) Put(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, path, opts)
}
func (c *Client) Patch(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodPatch, path, opts)
}
func (c *Client) Delete(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, path, opts)
}
func (c *Client) Head(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodHead, path, opts)
}
func (c *Client) Options(ctx context.Context, path string, opts *RequestOptions) (*http.Response, error) {
	return c.Do(ctx, http.MethodOptions, path, opts)
}
