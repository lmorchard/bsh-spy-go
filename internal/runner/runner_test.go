package runner

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/lmorchard/bsh-spy-go/internal/scraper"
	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/lmorchard/bsh-spy-go/internal/store"
	"github.com/sirupsen/logrus"
)

type fakeSpotify struct {
	track    spotify.Track
	found    bool
	addErr   error
	addCalls int
}

func (f *fakeSpotify) SearchTrack(_ context.Context, _, _ string) (spotify.Track, bool, error) {
	return f.track, f.found, nil
}
func (f *fakeSpotify) AddToPlaylist(_ context.Context, _ string, _ []string) error {
	f.addCalls++
	return f.addErr
}
func (f *fakeSpotify) PlaylistTrackIDs(context.Context, string) ([]string, error) { return nil, nil }
func (f *fakeSpotify) Me(context.Context) (string, error)                         { return "", nil }
func (f *fakeSpotify) Playlists(context.Context) ([]spotify.Playlist, error)      { return nil, nil }

type fakeStore struct {
	seen      map[string]bool
	mysteries int
	added     []store.Track
}

func newFakeStore() *fakeStore                   { return &fakeStore{seen: map[string]bool{}} }
func (s *fakeStore) Has(id string) (bool, error) { return s.seen[id], nil }
func (s *fakeStore) Add(t store.Track) error {
	s.added = append(s.added, t)
	s.seen[t.SpotifyID] = true
	return nil
}
func (s *fakeStore) RecordMystery(_, _, _ string) error { s.mysteries++; return nil }

func np() scraper.NowPlaying { return scraper.NowPlaying{Artist: "A", Title: "B"} }
func log() *logrus.Logger    { l := logrus.New(); l.SetOutput(io.Discard); return l }

func TestNewTrackIsAddedAndCached(t *testing.T) {
	sp := &fakeSpotify{track: spotify.Track{ID: "t1", URI: "u1"}, found: true}
	st := newFakeStore()
	d := Deps{Spotify: sp, Store: st, PlaylistID: "pl"}
	if err := ProcessSong(context.Background(), d, np(), log()); err != nil {
		t.Fatal(err)
	}
	if sp.addCalls != 1 || len(st.added) != 1 || st.added[0].SpotifyID != "t1" {
		t.Fatalf("addCalls=%d added=%+v", sp.addCalls, st.added)
	}
}

func TestKnownTrackIsSkipped(t *testing.T) {
	sp := &fakeSpotify{track: spotify.Track{ID: "t1", URI: "u1"}, found: true}
	st := newFakeStore()
	st.seen["t1"] = true
	d := Deps{Spotify: sp, Store: st, PlaylistID: "pl"}
	if err := ProcessSong(context.Background(), d, np(), log()); err != nil {
		t.Fatal(err)
	}
	if sp.addCalls != 0 || len(st.added) != 0 {
		t.Fatalf("should skip: addCalls=%d added=%+v", sp.addCalls, st.added)
	}
}

func TestFailedAddIsNotCached(t *testing.T) {
	sp := &fakeSpotify{track: spotify.Track{ID: "t1", URI: "u1"}, found: true, addErr: errors.New("boom")}
	st := newFakeStore()
	d := Deps{Spotify: sp, Store: st, PlaylistID: "pl"}
	if err := ProcessSong(context.Background(), d, np(), log()); err == nil {
		t.Fatal("expected error from failed add")
	}
	if len(st.added) != 0 {
		t.Fatalf("failed add must not cache: %+v", st.added)
	}
}

func TestNoResultRecordsMystery(t *testing.T) {
	sp := &fakeSpotify{found: false}
	st := newFakeStore()
	d := Deps{Spotify: sp, Store: st, PlaylistID: "pl"}
	if err := ProcessSong(context.Background(), d, np(), log()); err != nil {
		t.Fatal(err)
	}
	if st.mysteries != 1 || sp.addCalls != 0 {
		t.Fatalf("mysteries=%d addCalls=%d", st.mysteries, sp.addCalls)
	}
}
