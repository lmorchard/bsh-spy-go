package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	authHost string
	authPort int
)

var scopes = []string{
	"playlist-read-collaborative", "playlist-modify-public", "playlist-modify-private",
	"playlist-read-private", "user-library-modify", "user-library-read",
	"user-top-read", "user-read-recently-played",
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Run the one-time Spotify OAuth flow and print a refresh token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		log := GetLogger()
		if cfg.SpotifyClientID == "" || cfg.SpotifyClientSecret == "" {
			return fmt.Errorf("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be set")
		}
		redirect := fmt.Sprintf("http://%s:%d/authorize", authHost, authPort)
		conf := &oauth2.Config{
			ClientID:     cfg.SpotifyClientID,
			ClientSecret: cfg.SpotifyClientSecret,
			RedirectURL:  redirect,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
			},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			authURL := conf.AuthCodeURL("state-noop", oauth2.AccessTypeOffline)
			fmt.Fprintf(w, `<html><body><a href=%q>Authorize!</a></body></html>`, authURL)
		})
		mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if e := q.Get("error"); e != "" {
				http.Error(w, "authorization error: "+e, http.StatusBadRequest)
				return
			}
			tok, err := conf.Exchange(context.Background(), q.Get("code"))
			if err != nil {
				http.Error(w, "token exchange failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			snippet := fmt.Sprintf("SPOTIFY_REFRESH_TOKEN=%s", tok.RefreshToken)
			fmt.Fprintf(w, `<html><body><h2>Add this to your config/.env</h2><textarea style="width:100%%;height:4em">%s</textarea></body></html>`, snippet)
			fmt.Println(snippet)
			log.Info("refresh token issued — printed to stdout and browser")
		})

		addr := fmt.Sprintf("%s:%d", authHost, authPort)
		fmt.Printf("Open http://%s/ and authorize. Redirect URI must be registered in your Spotify app: %s\n", addr, redirect)
		return http.ListenAndServe(addr, mux)
	},
}

func init() {
	authCmd.Flags().StringVar(&authHost, "host", "127.0.0.1", "host for the local auth server")
	authCmd.Flags().IntVar(&authPort, "port", 8675, "port for the local auth server")
	rootCmd.AddCommand(authCmd)
}
