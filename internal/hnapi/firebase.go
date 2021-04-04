package hnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/reddit/baseplate.go/httpbp"
)

const (
	prefix    = `https://hacker-news.firebaseio.com/v0/`
	item      = "item/%d.json"
	maxItem   = "maxitem.json"
	userAgent = "httpsbot/1.0"
)

var client http.Client

func generateReq(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, prefix+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", userAgent)
	return req.WithContext(ctx), nil
}

// MaxItem returns the current max item number on hn.
func MaxItem(ctx context.Context) (int64, error) {
	req, err := generateReq(ctx, maxItem)
	if err != nil {
		return 0, fmt.Errorf("failed to generate request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request failed: %w", err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return 0, fmt.Errorf("http request failed: %w", err)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return 0, fmt.Errorf("failed to read body: %w", err)
	}
	n, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse body: %w, (%q)", err, body)
	}
	return n, nil
}

// GetItem gets an item from hn by its id.
func GetItem(ctx context.Context, id int64) (*Item, error) {
	req, err := generateReq(ctx, fmt.Sprintf(item, id))
	if err != nil {
		return nil, fmt.Errorf("failed to generate request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	i := new(Item)
	if err := json.NewDecoder(resp.Body).Decode(i); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	return i, nil
}
