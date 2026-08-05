package cli

// navigator-sync CLI: BAS integration layer join step for /moai project
// (SPEC-NAVIGATOR-SYNC-001). Joins 001's capability-map.md, 002's
// audit-report.json, 003's capability-symbols.json, plus the three binding
// token families (@NAV:DEC, @NAV:SYM, @MX:SPEC) into a single
// .moai/project/navigator/nav-graph.json artifact.
//
// Fail-open: exit 0 always. Capability gate: when capability-map.md is
// absent, emits an info log and writes no output (REQ-NS-011).

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// newNavigatorSyncCmd creates the navigator-sync subcommand. Hidden from the
// top-level `moai --help` surface (mirrors `navigator-enrich`).
func newNavigatorSyncCmd() *cobra.Command {
	var projectRoot string
	var capMapPath string
	var capSymPath string
	var auditPath string
	var outPath string

	cmd := &cobra.Command{
		Use:    "navigator-sync",
		Short:  "Join the 3 Navigator chains + binding tokens into nav-graph.json",
		Hidden: true,
		Long: `BAS integration-layer join step.

Reads 001's capability-map.md, 003's capability-symbols.json, 002's
audit-report.json (advisory), scans design docs + code for @NAV:DEC and
@NAV:SYM tokens, and consumes the internal/mx SpecAssociator output for
@MX:SPEC associations. Writes a single nav-graph.json atomically.

Fail-open: exit 0 always. When capability-map.md is absent, emits an info
log and writes no output (REQ-NS-011).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigatorSync(projectRoot, capMapPath, capSymPath, auditPath, outPath)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")
	cmd.Flags().StringVar(&capMapPath, "capability-map", "",
		"path to capability-map.md (defaults to <root>/.moai/project/navigator/capability-map.md)")
	cmd.Flags().StringVar(&capSymPath, "capability-symbols", "",
		"path to capability-symbols.json (defaults to <root>/.moai/project/codemaps/capability-symbols.json)")
	cmd.Flags().StringVar(&auditPath, "audit-report", "",
		"path to audit-report.json (defaults to <root>/.moai/project/navigator/audit-report.json)")
	cmd.Flags().StringVar(&outPath, "out", "",
		"output path (defaults to <root>/.moai/project/navigator/nav-graph.json)")

	return cmd
}

// runNavigatorSync is the fail-open core. It never returns a non-nil error
// to the cobra RunE so /moai project never aborts on the join step.
func runNavigatorSync(projectRoot, capMapPath, capSymPath, auditPath, outPath string) error {
	root := projectRoot
	if root == "" {
		root = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			root = "."
		}
	}
	root, _ = filepath.Abs(root)

	opts := sync.Options{
		ProjectRoot:           root,
		CapabilityMapPath:     capMapPath,
		CapabilitySymbolsPath: capSymPath,
		AuditReportPath:       auditPath,
		OutPath:               outPath,
	}
	if err := sync.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "navigator-sync: %v\n", err)
	}
	return nil
}
