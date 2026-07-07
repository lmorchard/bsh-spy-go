package scraper

import "testing"

func TestParseSplitsAndUnescapes(t *testing.T) {
	body := []byte(`{"icestats":{"source":{"metadata_updated":"06/07/2026 12:34:56 +0000","yp_currently_playing":"Massive Attack &amp; Azekel - Ritual Spirit (Live)"}}}`)
	np, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if np.Artist != "Massive Attack" { // fixup collapses "& Azekel"
		t.Fatalf("artist: got %q", np.Artist)
	}
	if np.Title != "Ritual Spirit" { // trailing " (...)" trimmed
		t.Fatalf("title: got %q", np.Title)
	}
}

func TestCleanUpSongFixups(t *testing.T) {
	cases := []struct{ inArtist, inTitle, wantArtist, wantTitle string }{
		{"Bj╤årk", "It''s Oh So Quiet", "Björk", "It's Oh So Quiet"},
		{"Sigur Roґs", "Svefn", "Sigur Ros", "Svefn"},
		{"The Sisters Of Mercy", "Draems", "Sisters Of Mercy", "Dreams"},
	}
	for _, c := range cases {
		a, ti := cleanUpSong(c.inArtist, c.inTitle)
		if a != c.wantArtist || ti != c.wantTitle {
			t.Errorf("cleanUpSong(%q,%q) = %q,%q; want %q,%q", c.inArtist, c.inTitle, a, ti, c.wantArtist, c.wantTitle)
		}
	}
}

func TestParseDropsTrailingDashSegments(t *testing.T) {
	body := []byte(`{"icestats":{"source":{"metadata_updated":"06/07/2026 12:34:56 +0000","yp_currently_playing":"Artist - Title - Radio Edit"}}}`)
	np, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if np.Artist != "Artist" || np.Title != "Title" {
		t.Fatalf("got artist=%q title=%q; want Artist/Title (trailing segment dropped)", np.Artist, np.Title)
	}
}

func TestParseDate(t *testing.T) {
	np, err := Parse([]byte(`{"icestats":{"source":{"metadata_updated":"06/07/2026 12:34:56 +0000","yp_currently_playing":"A - B"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if np.PlayedAt.UTC().Format("2006-01-02 15:04:05") != "2026-07-06 12:34:56" {
		t.Fatalf("playedAt: got %v", np.PlayedAt.UTC())
	}
}
