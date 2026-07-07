package cmd

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmorchard/bsh-spy-go/internal/runner"
	"github.com/lmorchard/bsh-spy-go/internal/scraper"
	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/lmorchard/bsh-spy-go/internal/store"
	"github.com/spf13/cobra"
)

var runOnce bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Poll the station and add newly-played songs to the playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := GetLogger()
		cfg := GetConfig()

		st, err := store.Open(cfg.Database)
		if err != nil {
			return err
		}
		defer st.Close()

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		sp := spotify.NewFromRefreshToken(ctx, cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
		deps := runner.Deps{Spotify: sp, Store: st, PlaylistID: cfg.SpotifyPlaylistID}

		cycle := func() {
			np, err := scraper.Fetch(ctx, cfg.StreemlionURL)
			if err != nil {
				log.WithError(err).Error("scrape failed")
				return
			}
			if err := runner.ProcessSong(ctx, deps, np, log); err != nil {
				log.WithError(err).Error("process song failed")
			}
		}

		cycle()
		if runOnce {
			return nil
		}

		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		log.WithField("interval", cfg.Interval).Info("daemon started")
		for {
			select {
			case <-ctx.Done():
				log.Info("shutting down")
				return nil
			case <-ticker.C:
				cycle()
			}
		}
	},
}

func init() {
	runCmd.Flags().BoolVar(&runOnce, "once", false, "run a single poll and exit")
	rootCmd.AddCommand(runCmd)
}
