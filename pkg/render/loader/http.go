package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ****** HTTP ******

// HTTP load from a remote URL
type HTTP struct {
	client  *http.Client
	maxSize int64
}

type limitedBody struct {
	io.Reader
	closer io.Closer
}

func NewHTTP(client *http.Client, maxSize int64) *HTTP {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	if maxSize <= 0 {
		maxSize = 8 << 20 // 8Mo
	}
	return &HTTP{
		client:  client,
		maxSize: maxSize,
	}
}

func (l *HTTP) Supports(ref Ref) bool {
	return ref.Scheme() == "http" || ref.Scheme() == "https"
}

func (l *HTTP) Load(ctx context.Context, ref Ref) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	// defer resp.Close()

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("loader: unexpected status %d for %s", resp.StatusCode, ref.String())
	}

	return &limitedBody{Reader: io.LimitReader(resp.Body, l.maxSize), closer: resp.Body}, nil

}

func (b *limitedBody) Close() error {
	return b.closer.Close()
}

func (l *HTTP) Stamp(ctx context.Context, ref Ref) (Stamp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, ref.String(), nil)

	if err != nil {
		return Stamp{}, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return Stamp{}, err
	}
	defer resp.Body.Close()

	return Stamp{
		ETag: resp.Header.Get("ETag"),
		Size: resp.ContentLength,
	}, nil
}
