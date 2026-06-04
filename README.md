# ghostnote-cli
command-line notes + local sync daemon for the ghostnote ecosystem.


The command-line core of the **Ghostnote** ecosystem — a fast, local-first note
tool and the sync daemon that connects the Ghostnote web app, TUI, and CLI.

All Ghostnote tools read and write the same notes, stored as plain JSON files in
`~/.ghostnote/notes/` using the shared `ghostnote/v1` format. A note you create
in the CLI shows up in the TUI and the web app, and vice versa.

```
ghostnote new ideas
ghostnote append ideas "new project"
ghostnote list
ghostnote daemon     # lets the web app sync to this machine
```

## Install

Requires [Go](https://go.dev/dl/) 1.22+.

```bash
go install github.com/drgropp/ghostnote-cli@latest
```

This builds `ghostnote` and puts it on your Go bin path (`$(go env GOPATH)/bin`,
usually `~/go/bin`). Make sure that directory is on your `PATH`. Then:

```bash
ghostnote --help
```

### Build from source

```bash
git clone https://github.com/drgropp/ghostnote-cli.git
cd ghostnote-cli
go build -o ghostnote .   # or: go install .
```

## Commands

| Command | Description |
|---|---|
| `ghostnote list` | List saved notes (aliases: `ls`, `notes`) |
| `ghostnote show <name>` | Print a note's text |
| `ghostnote new <name>` | Create an empty note |
| `ghostnote append <name> <text>` | Append text to a note (creates it if missing) |
| `ghostnote rm <name>` | Delete a note (aliases: `delete`, `del`) |
| `ghostnote export <name> [path]` | Write a note to a `.ghostnote.json` file |
| `ghostnote import <path>` | Import a `.ghostnote.json` file |
| `ghostnote daemon` | Run the local sync daemon |

## The sync daemon

`ghostnote daemon` runs a small local HTTP server (default `127.0.0.1:7777`)
that the Ghostnote **web app** talks to. With the daemon running, you can push a
note from the browser straight into `~/.ghostnote/notes/`, where the CLI and TUI
see it instantly.

```bash
ghostnote daemon
# ghostnote daemon listening on http://127.0.0.1:7777 (provider: none)
```

Leave it running while you use the web app. Quit with `Ctrl+C`.

### Verifying it works

In another terminal:

```bash
curl http://127.0.0.1:7777/health
# {"ok":true,"provider":"none","version":"ghostnote/v1"}
```

### Sync providers (optional)

The daemon supports pluggable sync backends, configured in
`~/.ghostnote/config.json`:

| Provider | What it does | Privacy |
|---|---|---|
| `none` (default) | Local only — web ↔ native on the same machine | Fully local |
| `tailscale` | Reach this daemon from your phone over a private mesh | Private, desktop must be on |
| `gist` | Sync via your own private GitHub Gist (works when desktop is off) | Notes pass through your Gist |
| `whisper` | Same-Wi-Fi peer sync (via the `ghost-whisper` tool) | Local, no internet |

Example config:

```json
{
  "port": 7777,
  "sync": { "provider": "none" }
}
```

See `config.example.json`. Note that `gist` and `tailscale` are scaffolded; the
`none` provider (local same-machine sync) is the fully working default. Only `none` is tested.

## The note format (`ghostnote/v1`)

Each note is one self-contained JSON file:

```json
{
  "schema": "ghostnote/v1",
  "id": "uuid",
  "name": "ideas",
  "text": "the note body, plain markdown",
  "drawing": null,
  "meta": {},
  "created": 1730000000000,
  "updated": 1730000000000
}
```

`text` is the universal field every tool reads. `drawing` is an optional PNG
data-URL used by the web app's canvas; the CLI and TUI preserve it untouched.

## The Ghostnote ecosystem

- **ghostnote** — the web app (canvas + text, runs in the browser)
- **ghostnote-tui** — a terminal note editor
- **ghostnote-cli** — this tool (CLI + sync daemon)

All share the `ghostnote/v1` format and the `~/.ghostnote/notes/` directory.

## License

MIT

