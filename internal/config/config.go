package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration.
type Config struct {
	Database string
	Verbose  bool
	Debug    bool
	LogJSON  bool

	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRefreshToken string
	SpotifyPlaylistID   string
	StreemlionURL       string
	Interval            time.Duration
	LogLevel            string
}

// SetDefaults registers viper defaults. Call before AutomaticEnv / Load.
func SetDefaults() {
	viper.SetDefault("database", "data/bsh.db")
	viper.SetDefault("verbose", false)
	viper.SetDefault("debug", false)
	viper.SetDefault("log_json", false)
	viper.SetDefault("interval", "60s")
	viper.SetDefault("log_level", "info")
}

// Load builds a Config from current viper state.
func Load() *Config {
	interval, err := time.ParseDuration(viper.GetString("interval"))
	if err != nil || interval <= 0 {
		interval = 60 * time.Second
	}
	return &Config{
		Database:            viper.GetString("database"),
		Verbose:             viper.GetBool("verbose"),
		Debug:               viper.GetBool("debug"),
		LogJSON:             viper.GetBool("log_json"),
		SpotifyClientID:     viper.GetString("spotify_client_id"),
		SpotifyClientSecret: viper.GetString("spotify_client_secret"),
		SpotifyRefreshToken: viper.GetString("spotify_refresh_token"),
		SpotifyPlaylistID:   viper.GetString("spotify_playlist_id"),
		StreemlionURL:       viper.GetString("streemlion_json_url"),
		Interval:            interval,
		LogLevel:            viper.GetString("log_level"),
	}
}
