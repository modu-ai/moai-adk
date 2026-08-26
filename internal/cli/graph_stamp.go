package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// Repo-relative names carried by the stamp command's user-facing filesystem
// errors — the absolute local path stays out (CR round-2 3855149248): the
// underlying *fs.PathError's path is host information a CLI-boundary error
// need not disclose.
const (
	stampCMDirRel  = ".moai/project/codemaps"
	stampTargetRel = stampCMDirRel + "/provenance.json"
	stampTmpRel    = stampTargetRel + ".tmp"
)

// stampFSDetail renders a filesystem error's path-free detail: a *os.LinkError
// or *fs.PathError contributes its errno only ("not a directory", "file
// exists") — a LinkError's own message names BOTH absolute paths (rename src
// dst), so it needs the same unwrapping — anything else its message.
func stampFSDetail(err error) string {
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err.Error()
	}
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}

// newGraphStampCmd 'moai graph stamp codemaps' stamps the codemaps provenance
// sidecar. The codemaps CONTENT is curated (agent/LLM-written via /moai
// codemaps); this command records WHAT TREE STATE the content describes, so
// the drift gate can judge it. The regeneration pipeline runs it as its last
// step.
//
// @MX:NOTE: [AUTO] graph stamp — content curation stays with /moai codemaps; only the provenance anchor is mechanical
func newGraphStampCmd() *cobra.Command {
	var rootArg string

	cmd := &cobra.Command{
		Use:   "stamp codemaps",
		Short: "Stamp codemaps provenance (run after regenerating codemaps)",
		Long: `Stamp .moai/project/codemaps/provenance.json with the current tree anchor.

Run as the LAST step of a codemaps regeneration: the content is curated, the
provenance is mechanical. The stamped block records the tree root, the HEAD
commit (or dirty + content fingerprint when described sources carry
uncommitted changes), and the described roots — everything 'moai graph check'
needs to judge the layer.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveGraphRoot(rootArg)
			if err != nil {
				return fmt.Errorf("graph stamp: %w", err)
			}

			pv, err := mx.StampCodemaps(projectRoot)
			if err != nil {
				return fmt.Errorf("stamp codemaps provenance: %w", err)
			}

			cmDir := filepath.Join(projectRoot, ".moai", "project", "codemaps")
			if err := os.MkdirAll(cmDir, 0o755); err != nil {
				return fmt.Errorf("graph stamp: create directory %s: %s", stampCMDirRel, stampFSDetail(err))
			}
			data, err := json.MarshalIndent(pv, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal provenance: %w", err)
			}

			target := filepath.Join(cmDir, "provenance.json")
			tmp := target + ".tmp"
			if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
				return fmt.Errorf("graph stamp: write %s: %s", stampTmpRel, stampFSDetail(err))
			}
			// Cleanup contract: temp removal is reported (never silently
			// dropped); once renamed, the temp no longer exists and IsNotExist
			// is the expected healthy outcome. The report names the
			// repo-relative temp — the absolute path stays out of user-facing
			// output (CR round-2 3855149248).
			defer func() {
				if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
					cmd.PrintErrf("graph stamp: temp cleanup %s: %s\n", stampTmpRel, stampFSDetail(err))
				}
			}()
			if err := os.Rename(tmp, target); err != nil {
				return fmt.Errorf("graph stamp: install %s: %s", stampTargetRel, stampFSDetail(err))
			}

			cmd.Printf("OK: stamped %s\n%s\n", target, pv.Describe())
			return nil
		},
	}

	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")
	return cmd
}
