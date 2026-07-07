package cmd

import (
	"context"
	"fmt"

	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print the authenticated Spotify user",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		ctx := context.Background()
		sp := spotify.NewFromRefreshToken(ctx, cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
		who, err := sp.Me(ctx)
		if err != nil {
			return err
		}
		fmt.Println(who)
		return nil
	},
}

func init() { rootCmd.AddCommand(whoamiCmd) }
