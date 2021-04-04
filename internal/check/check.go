package check

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reddit/baseplate.go/httpbp"

	"github.com/fishy/https-bot/similarity"
)

// Common errors
var (
	ErrNotHTTP = errors.New("not an http url")
)

var client http.Client

// Check checks whether there's https url to http url urlStr with similar
// content.
func Check(ctx context.Context, urlStr string, peek int64, headers http.Header) (httpsURL string, sim float64, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse url %q: %w", urlStr, err)
	}
	if u.Scheme != "http" {
		return "", 0, ErrNotHTTP
	}

	oldContent, err := peekResponse(reqFromURL(ctx, u, headers), urlStr, peek)
	if err != nil {
		return "", 0, err
	}

	u.Scheme = "https"
	httpsURL = u.String()
	newContent, err := peekResponse(reqFromURL(ctx, u, headers), httpsURL, peek)
	if err != nil {
		return "", 0, err
	}

	sim = similarity.MinSimilarity(oldContent, newContent)
	return
}

func reqFromURL(ctx context.Context, u *url.URL, headers http.Header) *http.Request {
	req := http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: headers,
	}
	return req.WithContext(ctx)
}

func peekResponse(req *http.Request, url string, peek int64) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed on %q: %w", url, err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return nil, fmt.Errorf("http request failed on %q: %w", url, err)
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, peek))
	if err != nil {
		err = fmt.Errorf("failed to read response for %q: %w", url, err)
	}
	return content, err
}
