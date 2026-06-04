// Package config loads ~/.ghostnote/config.json.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GistConfig configures the user-owned GitHub Gist sync provider.
type GistConfig struct {
	ID          string `json:"id"`
	Token       string `json:"token"`
	PollSeconds int    `json:"pollSeconds"`
}

// TailscaleConfig configures the Tailscale provider.
type TailscaleConfig struct {
	// Hostname or Tailscale IP of the desktop daemon. Empty = this machine
	// is the daemon and just listens; peers reach it over the tailnet.
	Peer string `json:"peer"`
}

// SyncConfig selects and configures the active sync provider.
type SyncConfig struct {
	Provider  string          `json:"provider"` // none | tailscale | gist | whisper
	Gist      GistConfig      `json:"gist"`
	Tailscale TailscaleConfig `json:"tailscale"`
}

// Config is the top-level config object.
type Config struct {
	Port      int        `json:"port"`      // daemon localhost port
	WebOrigin string     `json:"webOrigin"` // allowed CORS origin
	NotesDir  string     `json:"notesDir"`  // override ~/.ghostnote/notes
	Sync      SyncConfig `json:"sync"`
}

// Default returns sane defaults: local-only, no remote sync.
func Default() Config {
	return Config{
		Port:      7777,
		WebOrigin: "https://drgropp.github.io",
		Sync:      SyncConfig{Provider: "none"},
	}
}

func path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".ghostnote", "config.json")
}

// Load reads the config, falling back to defaults for any missing fields.
func Load() (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config yet — defaults are fine
		}
		return cfg, err
	}
	// Unmarshal over defaults so partial configs keep default values.
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Port == 0 {
		cfg.Port = 7777
	}
	if cfg.Sync.Provider == "" {
		cfg.Sync.Provider = "none"
	}
	return cfg, nil
}

// Save writes the config (used by an eventual `ghostnote config` command).
func Save(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".ghostnote")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o644)
}
