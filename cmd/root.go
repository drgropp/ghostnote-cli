// Package cmd wires the ghostnote CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drgropp/ghostnote-cli/internal/config"
	"github.com/drgropp/ghostnote-cli/internal/daemon"
	"github.com/drgropp/ghostnote-cli/internal/note"
	"github.com/spf13/cobra"
)

func Execute() error { return rootCmd.Execute() }

var rootCmd = &cobra.Command{
	Use:   "ghostnote",
	Short: "Ghostnote CLI — manage notes and run the local sync daemon",
	Long:  "ghostnote manages ~/.ghostnote/notes (ghostnote/v1 format) and runs the local daemon the web app and TUI sync through.",
}

func init() {
	rootCmd.AddCommand(daemonCmd, listCmd, showCmd, newCmd, rmCmd, appendCmd, exportCmd, importCmd)
}

func store() *note.Store {
	cfg, _ := config.Load()
	return note.NewStore(cfg.NotesDir)
}

// ---------- daemon ----------

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the local sync daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		d, err := daemon.New(cfg)
		if err != nil {
			return err
		}
		return d.Run()
	},
}

// ---------- list ----------

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "notes"},
	Short:   "List saved notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := store().List()
		if err != nil {
			return err
		}
		if len(idx) == 0 {
			fmt.Println("no saved notes")
			return nil
		}
		for _, it := range idx {
			when := time.UnixMilli(it.Updated).Format("2006-01-02 15:04")
			fmt.Printf("%-24s  %s\n", it.Name, when)
		}
		return nil
	},
}

// ---------- show ----------

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Print a note's text",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := store().Load(args[0])
		if err != nil {
			return fmt.Errorf("not found: %s", args[0])
		}
		fmt.Println(n.Text)
		return nil
	},
}

// ---------- new ----------

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create an empty note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := note.Note{Schema: note.Schema, Name: args[0]}
		n.Normalize()
		if _, err := store().Save(n); err != nil {
			return err
		}
		fmt.Println("created:", args[0])
		return nil
	},
}

// ---------- rm ----------

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"delete", "del"},
	Short:   "Delete a note",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := store().Delete(args[0]); err != nil {
			return fmt.Errorf("not found: %s", args[0])
		}
		fmt.Println("deleted:", args[0])
		return nil
	},
}

// ---------- append ----------

var appendCmd = &cobra.Command{
	Use:   "append <name> <text>",
	Short: "Append text to a note (creates it if missing)",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args[1:], " ")
		if _, err := store().Append(args[0], text); err != nil {
			return err
		}
		fmt.Println("appended to:", args[0])
		return nil
	},
}

// ---------- export ----------

var exportCmd = &cobra.Command{
	Use:   "export <name> [path]",
	Short: "Export a note as a .ghostnote.json file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := store().Load(args[0])
		if err != nil {
			return fmt.Errorf("not found: %s", args[0])
		}
		path := note.Slug(args[0]) + ".ghostnote.json"
		if len(args) > 1 {
			path = args[1]
		}
		data, err := json.MarshalIndent(n, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
		fmt.Println("exported:", path)
		return nil
	},
}

// ---------- import ----------

var importCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import a .ghostnote.json file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var n note.Note
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		n.Normalize()
		if err := n.Validate(); err != nil {
			return err
		}
		if _, err := store().Save(n); err != nil {
			return err
		}
		fmt.Println("imported:", n.Name)
		return nil
	},
}
