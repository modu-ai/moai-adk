package cli

// navigator-enrich CLI: AST symbol enrichment for /moai codemaps Phase 3
// (SPEC-PROJECT-NAVIGATOR-003). Reads 001's capability-map.md (header-driven),
// walks each implementation-path, extracts tree-sitter symbols, and writes
// capability-symbols.{md,json} atomically under .moai/project/codemaps/.
//
// Fail-open: exit 0 always. Capability gate: when capability-map.md is absent,
// emits an info log and writes no output (REQ-NT-002).

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// navigatorEnrichLogPath is the fail-open warning sink (REQ-NT-008).
const navigatorEnrichLogPath = ".moai/logs/navigator-astx.log"

// newNavigatorEnrichCmd creates the navigator-enrich subcommand.
func newNavigatorEnrichCmd() *cobra.Command {
	var projectRoot string
	var capMapPath string
	var outDir string

	cmd := &cobra.Command{
		Use:    "navigator-enrich",
		Short:  "Enrich capability-map with AST-derived symbols",
		Hidden: true,
		Long: `AST symbol enrichment for /moai codemaps Phase 3.

Reads 001's capability-map.md (header-driven join), walks each row's
implementation-path, extracts tree-sitter symbols, and writes
capability-symbols.{md,json} atomically under .moai/project/codemaps/.

Fail-open: exit 0 always. When capability-map.md is absent, emits an info
log and writes no output (REQ-NT-002).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigatorEnrich(projectRoot, capMapPath, outDir)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")
	cmd.Flags().StringVar(&capMapPath, "capability-map", "",
		"path to capability-map.md (defaults to <root>/.moai/project/navigator/capability-map.md)")
	cmd.Flags().StringVar(&outDir, "out-dir", "",
		"output directory (defaults to <root>/.moai/project/codemaps/)")

	return cmd
}

// runNavigatorEnrich is the fail-open core. It never returns a non-nil error
// to the cobra RunE (errors are logged and swallowed) so /moai codemaps never
// aborts on the enrichment step.
func runNavigatorEnrich(projectRoot, capMapPath, outDir string) error {
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

	if capMapPath == "" {
		capMapPath = filepath.Join(root, ".moai", "project", "navigator", "capability-map.md")
	}
	if outDir == "" {
		outDir = filepath.Join(root, ".moai", "project", "codemaps")
	}

	logPath := filepath.Join(root, navigatorEnrichLogPath)

	// Capability gate (REQ-NT-001 vs REQ-NT-002).
	if _, err := os.Stat(capMapPath); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-astx: capability-map absent at %s; skipping AST enrichment", capMapPath))
		fmt.Fprintln(os.Stderr, "navigator-astx: capability-map.md absent, skipping AST enrichment")
		return nil
	}

	res, err := astx.EnrichRows(astx.EnrichOptions{
		ProjectRoot:       root,
		CapabilityMapPath: capMapPath,
	})
	if err != nil {
		// Fail-open: EnrichRows already swallows errors, but defend anyway.
		appendLog(logPath, fmt.Sprintf("navigator-astx: enrich error: %v", err))
		return nil
	}

	// Render both outputs.
	sourceMap := relOrAbs(root, capMapPath)
	md := astx.RenderMarkdown(res.Provenance, sourceMap, res.Rows)
	js := astx.MarshalCapabilitySymbolsJSON(res.Provenance, sourceMap, res.Rows)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-astx: mkdir error: %v", err))
		return nil
	}

	// Atomic writes: <file>.tmp then mv (REQ-NT-013).
	if err := atomicWrite(filepath.Join(outDir, "capability-symbols.md"), md); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-astx: write md error: %v", err))
		return nil
	}
	if err := atomicWrite(filepath.Join(outDir, "capability-symbols.json"), js); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-astx: write json error: %v", err))
		return nil
	}

	fmt.Fprintf(os.Stderr, "navigator-astx: enriched %d row(s) -> %s\n",
		len(res.Rows), outDir)
	return nil
}

// atomicWrite writes data to <path>.tmp then renames it into place. It honors
// the NAVIGATOR_PRE_RENAME_BARRIER test hook: when set, it writes "ready" to
// the barrier path after creating the .tmp file and blocks (poll loop) until
// the barrier path is removed before the rename lands.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	// Test hook: synchronized barrier for the atomic-rename fixture
	// (mirrors 001's NAVIGATOR_PRE_RENAME_BARRIER pattern). The barrier is a
	// process-global one-shot latch: os.Unsetenv consumes it, so only the
	// first output file blocks and any second concurrent caller would
	// silently skip it (single-caller today, test-only).
	if barrier := os.Getenv("NAVIGATOR_PRE_RENAME_BARRIER"); barrier != "" {
		_ = os.Unsetenv("NAVIGATOR_PRE_RENAME_BARRIER")
		_ = os.WriteFile(barrier, []byte("ready"), 0o644)
		// Bounded wait: sleep per iteration plus an overall ceiling so a
		// test that fails before removing the barrier cannot strand a
		// CPU-burning goroutine (and race t.TempDir cleanup) for the rest
		// of the package run. Passing runs remove the barrier within
		// milliseconds; the 5s ceiling only fires on an already-failing
		// test, after which the rename proceeds and the goroutine exits.
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(barrier); err != nil {
				break // barrier removed → proceed to rename
			}
			time.Sleep(time.Millisecond)
		}
	}

	return os.Rename(tmp, path)
}

// appendLog appends a line to the navigator-astx log (fail-open on error).
func appendLog(path, line string) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintln(f, line)
}

// relOrAbs returns path relative to root when it is under root, else path.
func relOrAbs(root, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return path
	}
	return rel
}
