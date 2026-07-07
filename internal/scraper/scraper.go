package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// NowPlaying is a single scraped "now playing" entry.
type NowPlaying struct {
	Artist   string
	Title    string
	PlayedAt time.Time
}

type payload struct {
	Icestats struct {
		Source struct {
			MetadataUpdated    string `json:"metadata_updated"`
			YPCurrentlyPlaying string `json:"yp_currently_playing"`
		} `json:"source"`
	} `json:"icestats"`
}

// Streemlion reports "dd/mm/yyyy hh:mm:ss ±zzzz".
const streemlionLayout = "02/01/2006 15:04:05 -0700"

var parenSuffix = regexp.MustCompile(` \(.*\)`)

// Parse decodes a Streemlion status JSON payload into a NowPlaying.
func Parse(body []byte) (NowPlaying, error) {
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return NowPlaying{}, fmt.Errorf("decode status json: %w", err)
	}
	song := p.Icestats.Source.YPCurrentlyPlaying
	parts := strings.Split(song, " - ")
	artist := html.UnescapeString(strings.TrimSpace(parts[0]))
	title := ""
	if len(parts) > 1 {
		title = html.UnescapeString(strings.TrimSpace(parts[1]))
	}
	artist, title = cleanUpSong(artist, title)

	np := NowPlaying{Artist: artist, Title: title, PlayedAt: time.Now()}
	if t, err := time.Parse(streemlionLayout, p.Icestats.Source.MetadataUpdated); err == nil {
		np.PlayedAt = t
	}
	return np, nil
}

// Fetch GETs the status URL (cache-busted) and parses it.
func Fetch(ctx context.Context, url string) (NowPlaying, error) {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	busted := fmt.Sprintf("%s%s_=%d", url, sep, time.Now().UnixMilli())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, busted, nil)
	if err != nil {
		return NowPlaying{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NowPlaying{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return NowPlaying{}, fmt.Errorf("status %d fetching %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NowPlaying{}, err
	}
	return Parse(body)
}

// cleanUpSong ports the artist/title corrections from the Node implementation.
func cleanUpSong(artist, title string) (string, string) {
	switch artist {
	case "Bj╤årk", "Bjцrk":
		artist = "Björk"
	case "Sinйad O'Connor":
		artist = "Sinead O'Connor"
	case "INKRДKTARE":
		artist = "INKRÄKTARE"
	case "Jуnsi":
		artist = "Jonsi"
	case "Sigur Roґs", "Sigur Rуs":
		artist = "Sigur Ros"
	case "Massive Attack & Azekel":
		artist = "Massive Attack"
	case "The Sisters Of Mercy":
		artist = "Sisters Of Mercy"
	case "Rцyksopp":
		artist = "Royksopp"
	}
	title = strings.ReplaceAll(title, "''", "'")
	title = parenSuffix.ReplaceAllString(title, "")
	title = strings.ReplaceAll(title, "Draems", "Dreams")
	return artist, title
}
