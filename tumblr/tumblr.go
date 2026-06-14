// Package tumblr is the library behind the tumblr CLI.
package tumblr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PublicConsumerKey is the Tumblr consumer key embedded in their web/mobile clients.
const PublicConsumerKey = "fuiKNFp9vQFvjLNvx4sUwti4Yb5yGutBN4Xh10LXZhhRKjWlV4"

// Config holds all tunables for a Client.
type Config struct {
	BaseURL     string
	ConsumerKey string
	Rate        time.Duration
	Timeout     time.Duration
	Retries     int
	UserAgent   string
}

// DefaultConfig returns production-ready defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:     "https://api.tumblr.com",
		ConsumerKey: PublicConsumerKey,
		Rate:        200 * time.Millisecond,
		Timeout:     30 * time.Second,
		Retries:     3,
		UserAgent:   "tumblr/dev (+https://github.com/tamnd/tumblr-cli)",
	}
}

// Client talks to the Tumblr API.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

// NewClient builds a Client from cfg.
func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// Tagged returns posts tagged with the given tag.
func (c *Client) Tagged(ctx context.Context, tag string, limit int) ([]*Post, error) {
	if limit <= 0 {
		limit = 20
	}
	u := fmt.Sprintf("%s/v2/tagged?tag=%s&api_key=%s&limit=%d",
		c.cfg.BaseURL, url.QueryEscape(tag), c.cfg.ConsumerKey, limit)
	var resp struct {
		Response []json.RawMessage `json:"response"`
	}
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	out := make([]*Post, 0, len(resp.Response))
	for i, raw := range resp.Response {
		p := parsePost(raw)
		if p != nil {
			p.Rank = i + 1
			out = append(out, p)
		}
	}
	return out, nil
}

// Posts returns posts from a blog.
func (c *Client) Posts(ctx context.Context, blog string, limit int) ([]*Post, error) {
	if limit <= 0 {
		limit = 20
	}
	u := fmt.Sprintf("%s/v2/blog/%s/posts?api_key=%s&limit=%d",
		c.cfg.BaseURL, url.PathEscape(blog), c.cfg.ConsumerKey, limit)
	var resp struct {
		Response struct {
			Posts []json.RawMessage `json:"posts"`
		} `json:"response"`
	}
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	out := make([]*Post, 0, len(resp.Response.Posts))
	for i, raw := range resp.Response.Posts {
		p := parsePost(raw)
		if p != nil {
			p.Rank = i + 1
			out = append(out, p)
		}
	}
	return out, nil
}

// BlogInfo returns metadata about a blog.
func (c *Client) BlogInfo(ctx context.Context, blog string) (*Blog, error) {
	u := fmt.Sprintf("%s/v2/blog/%s/info?api_key=%s",
		c.cfg.BaseURL, url.PathEscape(blog), c.cfg.ConsumerKey)
	var resp struct {
		Response struct {
			Blog struct {
				Name        string `json:"name"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Posts       int    `json:"posts"`
				Updated     int64  `json:"updated"`
				URL         string `json:"url"`
			} `json:"blog"`
		} `json:"response"`
	}
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	b := resp.Response.Blog
	return &Blog{
		Name:        b.Name,
		Title:       b.Title,
		Description: b.Description,
		Posts:       b.Posts,
		Updated:     b.Updated,
		URL:         b.URL,
	}, nil
}

func parsePost(raw json.RawMessage) *Post {
	var p struct {
		IDStr    string   `json:"id_string"`
		Type     string   `json:"type"`
		BlogName string   `json:"blog_name"`
		Summary  string   `json:"summary"`
		Date     string   `json:"date"`
		Notes    int      `json:"note_count"`
		Tags     []string `json:"tags"`
		PostURL  string   `json:"post_url"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil
	}
	return &Post{
		ID:      p.IDStr,
		Type:    p.Type,
		Blog:    p.BlogName,
		Summary: p.Summary,
		Date:    p.Date,
		Notes:   p.Notes,
		Tags:    strings.Join(p.Tags, ","),
		URL:     p.PostURL,
	}
}

type apiError struct {
	Meta struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	} `json:"meta"`
}

func (c *Client) getJSON(ctx context.Context, rawURL string, out any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}

	var apiErr apiError
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Meta.Status != 0 && apiErr.Meta.Status != 200 {
		return fmt.Errorf("tumblr API error %d: %s", apiErr.Meta.Status, apiErr.Meta.Msg)
	}

	return json.Unmarshal(body, out)
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
