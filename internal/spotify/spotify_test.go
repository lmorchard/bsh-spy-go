package spotify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchQuery(t *testing.T) {
	got := SearchQuery("Aphex Twin", "Xtal")
	want := `artist:"Aphex Twin" Xtal`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSearchTrackFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tracks": map[string]any{"items": []map[string]any{
				{"id": "t1", "uri": "spotify:track:t1", "name": "Xtal",
					"artists": []map[string]any{{"name": "Aphex Twin"}}},
			}},
		})
	}))
	defer srv.Close()

	c := New(srv.Client())
	c.BaseURL = srv.URL
	tr, found, err := c.SearchTrack(context.Background(), "Aphex Twin", "Xtal")
	if err != nil {
		t.Fatal(err)
	}
	if !found || tr.ID != "t1" || tr.URI != "spotify:track:t1" {
		t.Fatalf("got %+v found=%v", tr, found)
	}
}

func TestSearchTrackNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"tracks": map[string]any{"items": []any{}}})
	}))
	defer srv.Close()
	c := New(srv.Client())
	c.BaseURL = srv.URL
	_, found, err := c.SearchTrack(context.Background(), "x", "y")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected not found")
	}
}

func TestAddToPlaylistPostsURIs(t *testing.T) {
	var gotBody struct {
		URIs []string `json:"uris"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method %s", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{"snapshot_id": "s1"})
	}))
	defer srv.Close()
	c := New(srv.Client())
	c.BaseURL = srv.URL
	if err := c.AddToPlaylist(context.Background(), "pl", []string{"spotify:track:t1"}); err != nil {
		t.Fatal(err)
	}
	if len(gotBody.URIs) != 1 || gotBody.URIs[0] != "spotify:track:t1" {
		t.Fatalf("posted uris: %+v", gotBody.URIs)
	}
}

func TestAddToPlaylistRetriesWithBody(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if len(bodies) == 1 {
			w.WriteHeader(http.StatusInternalServerError) // force one retry
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"snapshot_id": "s1"})
	}))
	defer srv.Close()
	c := New(srv.Client())
	c.BaseURL = srv.URL
	c.MaxRetries = 3
	if err := c.AddToPlaylist(context.Background(), "pl", []string{"spotify:track:t1"}); err != nil {
		t.Fatal(err)
	}
	if len(bodies) < 2 {
		t.Fatalf("expected a retry (>=2 requests), got %d", len(bodies))
	}
	if !strings.Contains(bodies[1], "spotify:track:t1") {
		t.Fatalf("retried request body lost the uris: %q", bodies[1])
	}
}

func TestAddToPlaylistErrorsOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := New(srv.Client())
	c.BaseURL = srv.URL
	c.MaxRetries = 1 // fail fast in tests
	if err := c.AddToPlaylist(context.Background(), "pl", []string{"x"}); err == nil {
		t.Fatal("expected error on 500")
	}
}
