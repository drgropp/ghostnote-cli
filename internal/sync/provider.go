// Package sync defines the SyncProvider interface and its implementations.
// Providers are decoupled behind one interface so the daemon doesn't care which
// is active. See DAEMON_SPEC.md for the design.
package sync

import (
	"fmt"

	"github.com/drgropp/ghostnote-cli/internal/config"
	"github.com/drgropp/ghostnote-cli/internal/note"
)

// Action is the safe set of operations a remote drop may request.
type Action string

const (
	ActionSave   Action = "save"
	ActionAppend Action = "append"
	ActionNew    Action = "new"
)

// AllowedRemote reports whether an action is permitted to arrive over the wire.
// Destructive ops (delete, wipe) and UI/mode ops are never allowed remotely.
func AllowedRemote(a Action) bool {
	switch a {
	case ActionSave, ActionAppend, ActionNew:
		return true
	default:
		return false
	}
}

// Drop is the envelope a remote sender (e.g. the phone web app) writes.
type Drop struct {
	ID     string    `json:"id"`     // unique drop id, for Ack/dedup
	Action Action    `json:"action"` // save | append | new
	Note   note.Note `json:"note"`
	Origin string    `json:"origin"` // e.g. "mobile"
	TS     int64     `json:"ts"`
}

// Provider is the pluggable sync backend.
type Provider interface {
	// Name returns the provider identifier (matches config "provider").
	Name() string
	// Pull fetches drops waiting remotely. Returns nil, nil when none.
	Pull() ([]Drop, error)
	// Push publishes a note outward (optional; may be a no-op).
	Push(n note.Note) error
	// Ack marks the given drop ids as consumed so they aren't pulled again.
	Ack(ids []string) error
}

// New constructs the provider named in the config.
func New(cfg config.Config) (Provider, error) {
	switch cfg.Sync.Provider {
	case "", "none":
		return NoneProvider{}, nil
	case "gist":
		return NewGistProvider(cfg.Sync.Gist), nil
	case "tailscale":
		return NewTailscaleProvider(cfg.Sync.Tailscale, cfg.Port), nil
	case "whisper":
		return NewWhisperProvider(), nil
	default:
		return nil, fmt.Errorf("unknown sync provider %q", cfg.Sync.Provider)
	}
}

// Apply enforces the remote allowlist and applies a drop to the store.
// This is the single choke point that protects the workspace from destructive
// remote actions, regardless of what the drop claims.
func Apply(store *note.Store, d Drop) error {
	if !AllowedRemote(d.Action) {
		return fmt.Errorf("action %q not allowed remotely", d.Action)
	}
	if err := d.Note.Validate(); err != nil {
		return err
	}
	switch d.Action {
	case ActionSave, ActionNew:
		_, err := store.Save(d.Note)
		return err
	case ActionAppend:
		_, err := store.Append(d.Note.Name, d.Note.Text)
		return err
	}
	return nil
}
