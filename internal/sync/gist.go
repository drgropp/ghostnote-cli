package sync

import (
	"github.com/drgropp/ghostnote-cli/internal/config"
	"github.com/drgropp/ghostnote-cli/internal/note"
)

// GistProvider uses the user's own private GitHub Gist as a drop box. The phone
// writes drops to the Gist; this daemon polls, pulls, applies, and clears them.
// It's the only path that works when the desktop was off when the note was
// sent. Opt-in: notes pass through GitHub's servers, and the token lives in the
// phone's browser. Each user owns their own Gist and data.
//
// Gist layout (suggested, for Claude Code to implement):
//   - One file in the gist, e.g. "ghostnote-drops.json", holding an array of
//     Drop objects awaiting pickup.
//   - Pull: GET the gist, parse the array.
//   - Apply happens in the daemon via sync.Apply (allowlist enforced).
//   - Ack: PATCH the gist, removing the consumed drop ids (last-write-wins;
//     accept the small race — drops are idempotent by id).
type GistProvider struct {
	cfg config.GistConfig
	// http client, base URL (https://api.github.com/gists/<id>), etc.
}

func NewGistProvider(cfg config.GistConfig) *GistProvider {
	if cfg.PollSeconds == 0 {
		cfg.PollSeconds = 20
	}
	return &GistProvider{cfg: cfg}
}

func (p *GistProvider) Name() string { return "gist" }

// PollInterval exposes how often the daemon should call Pull.
func (p *GistProvider) PollSeconds() int { return p.cfg.PollSeconds }

// Pull fetches waiting drops from the gist.
// TODO: GET https://api.github.com/gists/<id> with the user token,
// read the drops file, unmarshal []Drop.
func (p *GistProvider) Pull() ([]Drop, error) {
	return nil, nil
}

// Push publishes a note to the gist (so other devices can pull it).
// TODO: PATCH the gist drops file, appending a Drop with action=save.
func (p *GistProvider) Push(n note.Note) error {
	return nil
}

// Ack removes consumed drop ids from the gist.
// TODO: PATCH the gist drops file, filtering out the given ids.
func (p *GistProvider) Ack(ids []string) error {
	return nil
}
