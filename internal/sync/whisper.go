package sync

import "github.com/drgropp/ghostnote-cli/internal/note"

// WhisperProvider bridges to ghost-whisper, the same-Wi-Fi peer-to-peer sync
// tool (separate repo, Go, UDP discovery + TCP transfer). No internet, zero
// config. Within ghostnote-cli this is a thin adapter; the real networking
// lives in the ghost-whisper daemon, which writes directly to
// ~/.ghostnote/notes on each machine.
//
// Because ghost-whisper syncs the notes directory itself, this provider mostly
// signals "LAN sync is handled out-of-band" and Pull/Push are no-ops here.
type WhisperProvider struct{}

func NewWhisperProvider() *WhisperProvider { return &WhisperProvider{} }

func (p *WhisperProvider) Name() string          { return "whisper" }
func (p *WhisperProvider) Pull() ([]Drop, error) { return nil, nil }
func (p *WhisperProvider) Push(note.Note) error  { return nil }
func (p *WhisperProvider) Ack([]string) error    { return nil }
