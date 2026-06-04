// Package note defines the ghostnote/v1 format and filesystem storage.
// This is the shared contract with the web app and TUI — keep field names and
// JSON tags identical across all repos (see FORMAT_SPEC.md).
package note

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const Schema = "ghostnote/v1"

// Meta holds presentation hints. All optional; tools apply what they support.
type Meta struct {
	TextColor    string  `json:"textColor,omitempty"`
	TextSize     string  `json:"textSize,omitempty"`
	PenColor     string  `json:"penColor,omitempty"`
	PenWeight    float64 `json:"penWeight,omitempty"`
	EraserWeight float64 `json:"eraserWeight,omitempty"`
	BgAlpha      string  `json:"bgAlpha,omitempty"`
	TextAlpha    string  `json:"textAlpha,omitempty"`
}

// Note is a single ghostnote/v1 note.
type Note struct {
	Schema  string  `json:"schema"`
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Text    string  `json:"text"`
	Drawing *string `json:"drawing"` // web-only; preserve on round-trip
	Meta    Meta    `json:"meta"`
	Created int64   `json:"created"`
	Updated int64   `json:"updated"`
}

func NowMS() int64 { return time.Now().UnixMilli() }

// NewUUID returns a v4 UUID string.
func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

var slugRe = regexp.MustCompile(`[^a-z0-9_-]+`)

// Slug turns a note name into a filesystem-safe base name.
func Slug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugRe.ReplaceAllString(s, "")
	if s == "" {
		s = "note"
	}
	return s
}

// Validate checks the schema version and required fields.
func (n *Note) Validate() error {
	if n.Schema != Schema {
		return fmt.Errorf("unsupported schema %q (want %q)", n.Schema, Schema)
	}
	if n.Name == "" {
		return fmt.Errorf("note has no name")
	}
	return nil
}

// Normalize fills in defaults (schema, id, timestamps) in place.
func (n *Note) Normalize() {
	if n.Schema == "" {
		n.Schema = Schema
	}
	if n.ID == "" {
		n.ID = NewUUID()
	}
	if n.Updated == 0 {
		n.Updated = NowMS()
	}
	if n.Created == 0 {
		n.Created = n.Updated
	}
}

// ---------- Store ----------

// Store reads and writes notes under a directory (default ~/.ghostnote/notes).
type Store struct{ Dir string }

// DefaultDir returns ~/.ghostnote/notes, creating it if needed.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".ghostnote", "notes")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func NewStore(dir string) *Store {
	if dir == "" {
		dir = DefaultDir()
	}
	_ = os.MkdirAll(dir, 0o755)
	return &Store{Dir: dir}
}

func (s *Store) path(name string) string {
	return filepath.Join(s.Dir, Slug(name)+".ghostnote.json")
}

// Save writes a note, normalizing and bumping Updated.
func (s *Store) Save(n Note) (Note, error) {
	n.Normalize()
	n.Updated = NowMS()
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return n, err
	}
	if err := os.WriteFile(s.path(n.Name), data, 0o644); err != nil {
		return n, err
	}
	return n, nil
}

// Load reads a note by name.
func (s *Store) Load(name string) (Note, error) {
	var n Note
	data, err := os.ReadFile(s.path(name))
	if err != nil {
		return n, err
	}
	if err := json.Unmarshal(data, &n); err != nil {
		return n, err
	}
	return n, nil
}

// Delete removes a note by name.
func (s *Store) Delete(name string) error {
	return os.Remove(s.path(name))
}

// Index is a lightweight listing entry (no body).
type Index struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Updated int64  `json:"updated"`
}

// List returns all notes as index entries, newest first.
func (s *Store) List() ([]Index, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var out []Index
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ghostnote.json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.Dir, e.Name()))
		if err != nil {
			continue
		}
		var n Note
		if json.Unmarshal(data, &n) != nil {
			continue
		}
		name := n.Name
		if name == "" {
			name = strings.TrimSuffix(e.Name(), ".ghostnote.json")
		}
		out = append(out, Index{Name: name, ID: n.ID, Updated: n.Updated})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Updated > out[j].Updated })
	return out, nil
}

// Append adds text to an existing note (or creates it). Used by the safe
// remote "append" action and by ghost-hook.
func (s *Store) Append(name, text string) (Note, error) {
	n, err := s.Load(name)
	if err != nil {
		n = Note{Schema: Schema, Name: name}
		n.Normalize()
	}
	if n.Text != "" && !strings.HasSuffix(n.Text, "\n") {
		n.Text += "\n"
	}
	n.Text += text
	return s.Save(n)
}
