package tumblr_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/tumblr-cli/tumblr"
)

func newTestClient(ts *httptest.Server) *tumblr.Client {
	cfg := tumblr.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return tumblr.NewClient(cfg)
}

func TestTagged(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post := map[string]any{
			"id_string": "123456789",
			"type":      "photo",
			"blog_name": "testblog",
			"summary":   "A beautiful sunset",
			"date":      "2024-01-01 12:00:00 GMT",
			"note_count": 42,
			"tags":      []string{"sunset", "photography"},
			"post_url":  "https://testblog.tumblr.com/post/123456789",
		}
		raw, _ := json.Marshal(post)
		resp := map[string]any{
			"meta":     map[string]any{"status": 200, "msg": "OK"},
			"response": []json.RawMessage{raw},
		}
		b, _ := json.Marshal(resp)
		fmt.Fprint(w, string(b))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	posts, err := c.Tagged(context.Background(), "photography", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("want 1 post, got %d", len(posts))
	}
	if posts[0].ID != "123456789" {
		t.Errorf("ID = %q, want 123456789", posts[0].ID)
	}
	if posts[0].Type != "photo" {
		t.Errorf("Type = %q, want photo", posts[0].Type)
	}
	if posts[0].Notes != 42 {
		t.Errorf("Notes = %d, want 42", posts[0].Notes)
	}
	if posts[0].Rank != 1 {
		t.Errorf("Rank = %d, want 1", posts[0].Rank)
	}
}

func TestBlogInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"meta": map[string]any{"status": 200, "msg": "OK"},
			"response": map[string]any{
				"blog": map[string]any{
					"name":        "testblog",
					"title":       "Test Blog",
					"description": "A test blog",
					"posts":       100,
					"updated":     1704067200,
					"url":         "https://testblog.tumblr.com/",
				},
			},
		}
		b, _ := json.Marshal(resp)
		fmt.Fprint(w, string(b))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	blog, err := c.BlogInfo(context.Background(), "testblog")
	if err != nil {
		t.Fatal(err)
	}
	if blog.Name != "testblog" {
		t.Errorf("Name = %q, want testblog", blog.Name)
	}
	if blog.Posts != 100 {
		t.Errorf("Posts = %d, want 100", blog.Posts)
	}
}
