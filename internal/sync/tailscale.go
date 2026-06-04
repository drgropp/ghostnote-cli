package sync

import (
	"github.com/drgropp/ghostnote-cli/internal/config"
	"github.com/drgropp/ghostnote-cli/internal/note"
)

// TailscaleProvider exposes the daemon over a private Tailscale mesh so a phone
// can reach this desktop directly — no cloud storage, no token in a browser.
// Requires the desktop to be on and joined to the tailnet.
//
// Implementation approach (for Claude Code to fill in):
//   - This machine (the destination) does NOT need Pull/Ack — drops arrive as
//     normal HTTP PUT/POST to the daemon, which is reachable over the tailnet
//     because the daemon also binds the Tailscale interface (not just
//     127.0.0.1) when this provider is active.
//   - So the "provider" here mostly signals the daemon to listen on the
//     Tailscale IP in addition to localhost, and to widen CORS to the tailnet
//     origin. Pull/Push/Ack are no-ops on the destination side.
//   - Optionally, use the tsnet library (tailscale.com/tsnet) to embed
//     Tailscale directly so users don't need the separate Tailscale daemon.
type TailscaleProvider struct {
	cfg  config.TailscaleConfig
	port int
}

func NewTailscaleProvider(cfg config.TailscaleConfig, port int) *TailscaleProvider {
	return &TailscaleProvider{cfg: cfg, port: port}
}

func (p *TailscaleProvider) Name() string { return "tailscale" }

// Pull is a no-op: with Tailscale the phone pushes directly to this daemon's
// HTTP API over the mesh, so there's nothing to poll.
func (p *TailscaleProvider) Pull() ([]Drop, error) { return nil, nil }

// Push could forward a note to a peer daemon listed in cfg.Peer.
// TODO: HTTP PUT to http://<peer>:<port>/notes/:id over the tailnet.
func (p *TailscaleProvider) Push(n note.Note) error { return nil }

func (p *TailscaleProvider) Ack([]string) error { return nil }

// BindTailscale reports whether the daemon should also listen on the Tailscale
// interface (true when this provider is active). The daemon checks this.
func (p *TailscaleProvider) BindTailscale() bool { return true }
