package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

const defaultBaseURL = "https://api.spotify.com/v1"

var spotifyEndpoint = oauth2.Endpoint{
	AuthURL:  "https://accounts.spotify.com/authorize",
	TokenURL: "https://accounts.spotify.com/api/token",
}

// Track is a Spotify track we care about.
type Track struct {
	ID     string
	URI    string
	Artist string
	Title  string
}

// Playlist is a minimal playlist descriptor.
type Playlist struct {
	ID   string
	Name string
}

// Client is the subset of the Spotify API this app uses.
type Client interface {
	SearchTrack(ctx context.Context, artist, title string) (Track, bool, error)
	AddToPlaylist(ctx context.Context, playlistID string, uris []string) error
	PlaylistTrackIDs(ctx context.Context, playlistID string) ([]string, error)
	Me(ctx context.Context) (string, error)
	Playlists(ctx context.Context) ([]Playlist, error)
}

// HTTPClient is the concrete Spotify client.
type HTTPClient struct {
	BaseURL    string
	MaxRetries int
	hc         *http.Client
}

// New wraps an *http.Client (already carrying auth) as a Spotify client.
func New(hc *http.Client) *HTTPClient {
	return &HTTPClient{BaseURL: defaultBaseURL, MaxRetries: 3, hc: hc}
}

// NewFromRefreshToken builds an auto-refreshing client from OAuth credentials.
func NewFromRefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) *HTTPClient {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     spotifyEndpoint,
	}
	tok := &oauth2.Token{RefreshToken: refreshToken}
	return New(conf.Client(ctx, tok))
}

// SearchQuery builds the Spotify search query string used by the app.
func SearchQuery(artist, title string) string {
	return fmt.Sprintf("artist:%q %s", artist, title)
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body any, out any) error {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = b
	}
	var lastErr error
	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.hc.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			wait := 1 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, e := strconv.Atoi(ra); e == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limited on %s", path)
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("status %d on %s", resp.StatusCode, path)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			msg, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status %d on %s: %s", resp.StatusCode, path, msg)
		}
		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	}
	return fmt.Errorf("giving up on %s: %w", path, lastErr)
}

// SearchTrack returns the top track match for artist/title.
func (c *HTTPClient) SearchTrack(ctx context.Context, artist, title string) (Track, bool, error) {
	q := url.Values{}
	q.Set("type", "track")
	q.Set("q", SearchQuery(artist, title))
	q.Set("limit", "1")
	var out struct {
		Tracks struct {
			Items []struct {
				ID      string `json:"id"`
				URI     string `json:"uri"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"items"`
		} `json:"tracks"`
	}
	if err := c.do(ctx, http.MethodGet, "/search?"+q.Encode(), nil, &out); err != nil {
		return Track{}, false, err
	}
	if len(out.Tracks.Items) == 0 {
		return Track{}, false, nil
	}
	it := out.Tracks.Items[0]
	tr := Track{ID: it.ID, URI: it.URI, Title: it.Name}
	if len(it.Artists) > 0 {
		tr.Artist = it.Artists[0].Name
	}
	return tr, true, nil
}

// AddToPlaylist appends track URIs to a playlist.
func (c *HTTPClient) AddToPlaylist(ctx context.Context, playlistID string, uris []string) error {
	body := map[string]any{"uris": uris}
	return c.do(ctx, http.MethodPost, "/playlists/"+playlistID+"/tracks", body, nil)
}

// PlaylistTrackIDs pages the full playlist and returns all track IDs.
func (c *HTTPClient) PlaylistTrackIDs(ctx context.Context, playlistID string) ([]string, error) {
	var ids []string
	offset := 0
	for {
		q := url.Values{}
		q.Set("fields", "items(track(id)),total")
		q.Set("limit", "100")
		q.Set("offset", strconv.Itoa(offset))
		var out struct {
			Items []struct {
				Track struct {
					ID string `json:"id"`
				} `json:"track"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := c.do(ctx, http.MethodGet, "/playlists/"+playlistID+"/tracks?"+q.Encode(), nil, &out); err != nil {
			return nil, err
		}
		for _, it := range out.Items {
			if it.Track.ID != "" {
				ids = append(ids, it.Track.ID)
			}
		}
		offset += len(out.Items)
		if len(out.Items) == 0 || offset >= out.Total {
			break
		}
	}
	return ids, nil
}

// Me returns the authenticated user's display name/id.
func (c *HTTPClient) Me(ctx context.Context) (string, error) {
	var out struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	}
	if err := c.do(ctx, http.MethodGet, "/me", nil, &out); err != nil {
		return "", err
	}
	if out.DisplayName != "" {
		return out.DisplayName, nil
	}
	return out.ID, nil
}

// Playlists lists the authenticated user's playlists (first page).
func (c *HTTPClient) Playlists(ctx context.Context) ([]Playlist, error) {
	var out struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := c.do(ctx, http.MethodGet, "/me/playlists?limit=50", nil, &out); err != nil {
		return nil, err
	}
	pls := make([]Playlist, 0, len(out.Items))
	for _, it := range out.Items {
		pls = append(pls, Playlist{ID: it.ID, Name: it.Name})
	}
	return pls, nil
}
