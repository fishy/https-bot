package hnapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/reddit/baseplate.go/httpbp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	hnPrefix    = "https://news.ycombinator.com/"
	loginURL    = hnPrefix + "login"
	replyAction = "comment"
	replyURL    = hnPrefix + replyAction

	urlTmpl = hnPrefix + "item?id=%d"

	postFormContentType = "application/x-www-form-urlencoded"
)

// A Session is a logged in hn session that's able to reply.
type Session struct {
	client *http.Client
}

// NewSession logs in.
func NewSession(ctx context.Context, Username, Password string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cookiejar: %w", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	form := make(url.Values)
	form.Set("acct", Username)
	form.Set("pw", Password)
	form.Set("goto", "news")
	req, err := http.NewRequest(
		http.MethodPost,
		loginURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("content-type", postFormContentType)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	if len(jar.Cookies(req.URL)) == 0 {
		return nil, errors.New("login failed")
	}
	return &Session{client: client}, nil
}

// Reply sends a reply to the given id.
func (s *Session) Reply(ctx context.Context, id int64, content string) error {
	form, err := s.getReplyForm(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get parent post: %w", err)
	}
	form.Set("text", content)
	req, err := http.NewRequest(
		http.MethodPost,
		replyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("content-type", postFormContentType)
	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	return nil
}

func (s *Session) getReplyForm(ctx context.Context, id int64) (url.Values, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(urlTmpl, id),
		nil, // body
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpbp.DrainAndClose(resp.Body)
	if err := httpbp.ClientErrorFromResponse(resp); err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}
	formNode := findReplyForm(root)
	if formNode == nil {
		return nil, errors.New("cannot find reply form")
	}
	form := make(url.Values)
	for c := formNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.DataAtom != atom.Input {
			continue
		}
		k := getAttr(c, "name")
		v := getAttr(c, "value")
		if k != "" && v != "" {
			form.Set(k, v)
		}
	}
	return form, nil
}

func findReplyForm(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == atom.Form {
		if getAttr(n, "action") == replyAction && strings.ToUpper(getAttr(n, "method")) == http.MethodPost {
			return n
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findReplyForm(c); found != nil {
			return found
		}
	}

	return nil
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
