// Package daemon runs the local HTTP server the web app talks to, plus the
// background poll loop that pulls remote drops via the active sync provider.
// See DAEMON_SPEC.md for the contract.
package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/drgropp/ghostnote-cli/internal/config"
	"github.com/drgropp/ghostnote-cli/internal/note"
	gsync "github.com/drgropp/ghostnote-cli/internal/sync"
)

type Daemon struct {
	cfg      config.Config
	store    *note.Store
	provider gsync.Provider
}

func New(cfg config.Config) (*Daemon, error) {
	prov, err := gsync.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Daemon{
		cfg:      cfg,
		store:    note.NewStore(cfg.NotesDir),
		provider: prov,
	}, nil
}

// Run starts the HTTP server and the sync poll loop, blocking until error.
func (d *Daemon) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", d.handleHealth)
	mux.HandleFunc("/notes", d.handleNotes)    // GET list, POST create
	mux.HandleFunc("/notes/", d.handleNoteByID) // GET/PUT/DELETE one

	handler := d.withCORS(mux)

	go d.pollLoop()

	addr := fmt.Sprintf("127.0.0.1:%d", d.cfg.Port)
	log.Printf("ghostnote daemon listening on http://%s (provider: %s)", addr, d.provider.Name())
	// NOTE: when the tailscale provider is active, also bind the tailnet
	// interface here (see TailscaleProvider.BindTailscale). Left to Claude Code.
	return http.ListenAndServe(addr, handler)
}

// ---------- CORS ----------

func (d *Daemon) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", d.cfg.WebOrigin)
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
		// Chrome Private Network Access preflight ack:
		if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
			h.Set("Access-Control-Allow-Private-Network", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------- handlers ----------

func (d *Daemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "version": note.Schema, "provider": d.provider.Name()})
}

func (d *Daemon) handleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		idx, err := d.store.List()
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"notes": idx})
	case http.MethodPost:
		var n note.Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			writeErr(w, 400, err)
			return
		}
		n.Normalize()
		if err := n.Validate(); err != nil {
			writeErr(w, 400, err)
			return
		}
		saved, err := d.store.Save(n)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"id": saved.ID})
	default:
		writeErr(w, 405, fmt.Errorf("method not allowed"))
	}
}

func (d *Daemon) handleNoteByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/notes/")
	if id == "" {
		writeErr(w, 400, fmt.Errorf("missing note id/name"))
		return
	}
	// NOTE: routing is by name here for simplicity; FORMAT_SPEC uses id as the
	// stable key. A real impl should look up by id across the store. For the
	// skeleton we treat the path segment as the note name.
	switch r.Method {
	case http.MethodGet:
		n, err := d.store.Load(id)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		writeJSON(w, 200, n)
	case http.MethodPut:
		var n note.Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			writeErr(w, 400, err)
			return
		}
		n.Normalize()
		if err := n.Validate(); err != nil {
			writeErr(w, 400, err)
			return
		}
		if _, err := d.store.Save(n); err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	case http.MethodDelete:
		if err := d.store.Delete(id); err != nil {
			writeErr(w, 404, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeErr(w, 405, fmt.Errorf("method not allowed"))
	}
}

// ---------- sync poll loop ----------

func (d *Daemon) pollLoop() {
	interval := 20 * time.Second
	if g, ok := d.provider.(*gsync.GistProvider); ok {
		interval = time.Duration(g.PollSeconds()) * time.Second
	}
	if d.provider.Name() == "none" || d.provider.Name() == "tailscale" || d.provider.Name() == "whisper" {
		// none: nothing to poll. tailscale/whisper: drops arrive out-of-band.
		return
	}
	for {
		drops, err := d.provider.Pull()
		if err != nil {
			log.Printf("sync pull error: %v", err)
		}
		var done []string
		for _, drop := range drops {
			if err := gsync.Apply(d.store, drop); err != nil {
				log.Printf("drop %s rejected: %v", drop.ID, err)
				// still ack rejected drops so they don't loop forever
			}
			done = append(done, drop.ID)
		}
		if len(done) > 0 {
			if err := d.provider.Ack(done); err != nil {
				log.Printf("sync ack error: %v", err)
			}
		}
		time.Sleep(interval)
	}
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"ok": false, "error": err.Error()})
}
