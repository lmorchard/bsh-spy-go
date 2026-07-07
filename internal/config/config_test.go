package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadReadsEnvCompatibleKeys(t *testing.T) {
	viper.Reset()
	t.Setenv("SPOTIFY_CLIENT_ID", "cid")
	t.Setenv("SPOTIFY_PLAYLIST_ID", "pid")
	t.Setenv("STREEMLION_JSON_URL", "https://example/status.json")
	t.Setenv("INTERVAL", "30s")
	SetDefaults()
	viper.AutomaticEnv()

	c := Load()
	if c.SpotifyClientID != "cid" {
		t.Fatalf("client id: got %q", c.SpotifyClientID)
	}
	if c.SpotifyPlaylistID != "pid" {
		t.Fatalf("playlist id: got %q", c.SpotifyPlaylistID)
	}
	if c.StreemlionURL != "https://example/status.json" {
		t.Fatalf("url: got %q", c.StreemlionURL)
	}
	if c.Interval != 30*time.Second {
		t.Fatalf("interval: got %v", c.Interval)
	}
}

func TestLoadIntervalDefault(t *testing.T) {
	viper.Reset()
	SetDefaults()
	viper.AutomaticEnv()
	if got := Load().Interval; got != 60*time.Second {
		t.Fatalf("default interval: got %v", got)
	}
}
