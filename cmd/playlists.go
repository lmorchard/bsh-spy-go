package cmd

import (
	"context"
	"fmt"

	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/spf13/cobra"
)

var playlistsCmd = &cobra.Command{
	Use:   "playlists",
	Short: "List the authenticated user's playlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		ctx := context.Background()
		sp := spotify.NewFromRefreshToken(ctx, cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
		pls, err := sp.Playlists(ctx)
		if err != nil {
			return err
		}
		for _, p := range pls {
			fmt.Printf("%s\t%s\n", p.ID, p.Name)
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(playlistsCmd) }
