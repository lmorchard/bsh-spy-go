package store

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "nested", "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestAddAndHas(t *testing.T) {
	s := newTestStore(t)
	has, err := s.Has("abc")
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("expected miss on empty store")
	}
	if err := s.Add(Track{SpotifyID: "abc", Artist: "A", Title: "B"}); err != nil {
		t.Fatal(err)
	}
	has, err = s.Has("abc")
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("expected hit after Add")
	}
}

func TestAddIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	tr := Track{SpotifyID: "dup", Artist: "A", Title: "B"}
	if err := s.Add(tr); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(tr); err != nil { // must not error on re-add
		t.Fatalf("second Add errored: %v", err)
	}
}

func TestRecordMystery(t *testing.T) {
	s := newTestStore(t)
	if err := s.RecordMystery("A", "B", `artist:"A" B`); err != nil {
		t.Fatal(err)
	}
}
