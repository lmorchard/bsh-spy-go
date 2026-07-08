package runner

import (
	"context"

	"github.com/lmorchard/bsh-spy-go/internal/scraper"
	"github.com/lmorchard/bsh-spy-go/internal/spotify"
	"github.com/lmorchard/bsh-spy-go/internal/store"
	"github.com/sirupsen/logrus"
)

// StoreIface is the subset of *store.Store the runner needs.
type StoreIface interface {
	Has(spotifyID string) (bool, error)
	Add(t store.Track) error
	RecordMystery(artist, title, query string) error
}

// Deps are the runner's collaborators.
type Deps struct {
	Spotify    spotify.Client
	Store      StoreIface
	PlaylistID string
	DryRun     bool
}

// ProcessSong searches for np, adds it to the playlist if new, and caches it
// only after a successful add. Songs with no match are recorded as mysteries.
func ProcessSong(ctx context.Context, d Deps, np scraper.NowPlaying, log *logrus.Logger) error {
	track, found, err := d.Spotify.SearchTrack(ctx, np.Artist, np.Title)
	if err != nil {
		return err
	}
	if !found {
		if d.DryRun {
			log.WithFields(logrus.Fields{"artist": np.Artist, "title": np.Title}).Info("no match — would record mystery")
			return nil
		}
		log.WithFields(logrus.Fields{"artist": np.Artist, "title": np.Title}).Info("no match — recording mystery")
		return d.Store.RecordMystery(np.Artist, np.Title, spotify.SearchQuery(np.Artist, np.Title))
	}

	seen, err := d.Store.Has(track.ID)
	if err != nil {
		return err
	}
	if seen {
		if d.DryRun {
			log.WithFields(logrus.Fields{"artist": np.Artist, "title": np.Title, "id": track.ID}).Info("already in playlist — would skip")
			return nil
		}
		log.WithField("id", track.ID).Debug("already in cache — skipping")
		return nil
	}

	if d.DryRun {
		log.WithFields(logrus.Fields{"artist": np.Artist, "title": np.Title, "id": track.ID, "uri": track.URI}).Info("would add")
		return nil
	}

	if err := d.Spotify.AddToPlaylist(ctx, d.PlaylistID, []string{track.URI}); err != nil {
		return err // not cached — retried next time it plays
	}
	log.WithFields(logrus.Fields{"artist": np.Artist, "title": np.Title, "id": track.ID}).Info("added to playlist")
	return d.Store.Add(store.Track{SpotifyID: track.ID, Artist: np.Artist, Title: np.Title})
}
