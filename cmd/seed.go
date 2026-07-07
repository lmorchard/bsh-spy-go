package cmd

import (
	"context"

	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/lmorchard/bsh-spy-go/internal/store"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Populate the local cache from the current playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := GetLogger()
		cfg := GetConfig()
		ctx := context.Background()

		st, err := store.Open(cfg.Database)
		if err != nil {
			return err
		}
		defer func() { _ = st.Close() }()

		sp := spotify.NewFromRefreshToken(ctx, cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
		ids, err := sp.PlaylistTrackIDs(ctx, cfg.SpotifyPlaylistID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := st.Add(store.Track{SpotifyID: id}); err != nil {
				return err
			}
		}
		log.WithField("count", len(ids)).Info("seeded cache from playlist")
		return nil
	},
}

func init() { rootCmd.AddCommand(seedCmd) }
