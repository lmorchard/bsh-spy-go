# bsh-spy-go

A Go CLI that watches [Big Sonic Heaven Radio](https://bigsonicheaven.com/)'s
now-playing status and adds each newly-played track to a Spotify playlist.
It polls the station's status JSON, matches the current track on Spotify,
and appends it to a playlist — skipping anything already added, tracking
tracks it couldn't find ("mystery songs") in a local SQLite cache.

This is a Go rewrite of the original Node.js `bsh-now-playing-scraper`.

## Building

```bash
make build   # builds ./bsh-spy-go (CGO_ENABLED=0, static)
make test    # go test ./...
make lint    # golangci-lint run
make format  # go fmt + gofumpt
```

## Configuration

Configuration comes from a `bsh-spy-go.yaml` file in the working directory,
environment variables, or CLI flags (in increasing priority). Copy
`bsh-spy-go.yaml.example` to `bsh-spy-go.yaml` to get started, or set the
equivalent environment variables (upper-cased key names, e.g.
`SPOTIFY_CLIENT_ID`).

| Key                    | Env var                  | Description                                      |
| ----------------------- | ------------------------- | ------------------------------------------------- |
| `spotify_client_id`     | `SPOTIFY_CLIENT_ID`       | Spotify app client ID                             |
| `spotify_client_secret` | `SPOTIFY_CLIENT_SECRET`   | Spotify app client secret                         |
| `spotify_refresh_token` | `SPOTIFY_REFRESH_TOKEN`   | OAuth refresh token (see `auth` below)            |
| `spotify_playlist_id`   | `SPOTIFY_PLAYLIST_ID`     | Target playlist ID                                |
| `streemlion_json_url`   | `STREEMLION_JSON_URL`     | Station status JSON URL                           |
| `interval`              | `INTERVAL`                | Poll interval, a Go duration (default `60s`)      |
| `database`              | `DATABASE`                | Path to the SQLite cache file (default `data/bsh.db`) |
| `log_level`             | `LOG_LEVEL`                | Log level (default `info`)                        |

## Workflow

1. **`auth`** — run the one-time Spotify OAuth flow and print a refresh
   token:

   ```bash
   go run . auth
   ```

   This opens a small local web server; follow the printed URL, authorize
   in your browser, and copy the `SPOTIFY_REFRESH_TOKEN` it prints into your
   config or `.env`.

2. Set `spotify_refresh_token` (or `SPOTIFY_REFRESH_TOKEN`) from the step
   above.

3. **`seed`** — populate the local cache with the playlist's existing
   tracks, so `run` doesn't re-add them:

   ```bash
   go run . seed
   ```

4. **`run`** — start the polling daemon (or `--once` for a single poll):

   ```bash
   go run . run
   go run . run --once
   ```

Other commands: `whoami` prints the authenticated Spotify user; `playlists`
lists the user's playlists (handy for finding a playlist ID).

## Docker

```bash
docker build -t bsh-spy-go .
mkdir -p data
docker run -d --restart unless-stopped \
  --env-file ./.env \
  -v "$PWD/data:/data" \
  --user "$(id -u):$(id -g)" \
  -e DATABASE=/data/bsh.db \
  --name bsh-spy bsh-spy-go run
```

The image is built from `gcr.io/distroless/static-debian12:nonroot`, so it
ships no shell and runs as a non-root user (UID 65532). `/data` is declared
as a volume for the SQLite cache; mount it (and point `DATABASE` at a path
inside it) so the cache survives container restarts. Because the container
runs as a non-root user, the bind-mounted `data/` directory must be
writable by that user — pass `--user "$(id -u):$(id -g)"` as shown above
(so the container writes as your host user), or alternatively `chown` the
host directory to UID 65532 ahead of time.
