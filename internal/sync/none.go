package sync

import "github.com/drgropp/ghostnote-cli/internal/note"

// NoneProvider is the default: fully local, no away-from-home sync.
// Web <-> native still works on the same machine via the localhost daemon.
type NoneProvider struct{}

func (NoneProvider) Name() string          { return "none" }
func (NoneProvider) Pull() ([]Drop, error) { return nil, nil }
func (NoneProvider) Push(note.Note) error  { return nil }
func (NoneProvider) Ack([]string) error    { return nil }
