package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/mx"
)

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
				return fmt.Errorf("mkdir codemaps: %w", err)
			}
			data, err := json.MarshalIndent(pv, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal provenance: %w", err)
			}

			target := filepath.Join(cmDir, "provenance.json")
			tmp := target + ".tmp"
			if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
				return fmt.Errorf("write provenance: %w", err)
			}
			// Cleanup contract: temp removal is reported (never silently
			// dropped); once renamed, the temp no longer exists and IsNotExist
			// is the expected healthy outcome.
			defer func() {
				if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
					cmd.PrintErrf("graph stamp: temp cleanup: %v\n", err)
				}
			}()
			if err := os.Rename(tmp, target); err != nil {
				return fmt.Errorf("rename provenance: %w", err)
			}

			cmd.Printf("OK: stamped %s\n%s\n", target, pv.Describe())
			return nil
		},
	}

	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")
	return cmd
}
