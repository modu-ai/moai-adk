// memory.go — `moai memory` : health check and archival for the Lessons
// Protocol topic-file store (MEMORY.md + feedback_*/project_*/… topic files).
//
// This is a different layer from `moai preference`, which owns the
// user_decisions/ sub-directory under the same memory root. The topic-file
// store is what a session actually loads at start, through MEMORY.md.
//
// The store's failure mode is silent: only MEMORY.md is loaded, so a topic
// file written without its index line is stored and never recalled. Nothing
// reported that until now — the pre-existing audits check a file's own shape
// and the index's line count, neither of which reads the index's links.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
	"github.com/spf13/cobra"
)

// memoryIndexName is the index a session loads.
const memoryIndexName = "MEMORY.md"

// memoryArchiveName holds retired topic files. Archiving never deletes: the
// constitution keeps the audit trail.
const memoryArchiveName = "_archive"

// memoryStore is a resolved topic-file store.
type memoryStore struct {
	Dir    string `json:"dir"`
	Origin string `json:"origin"` // how the path was resolved, for the report
}

// memoryProjectSlug encodes an absolute path the way Claude Code names its
// project directories (/ \ . : → -). It mirrors preference.memorySlug and
// hook.projectSlug; the three must stay character-for-character identical.
func memoryProjectSlug(absPath string) string {
	clean := filepath.Clean(absPath)
	var out []rune
	for _, r := range clean {
		switch r {
		case '/', '\\', '.', ':':
			out = append(out, '-')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// memoryCandidateStores returns the stores this project could be using, in
// the order a session resolves them: the active profile's config dir first,
// then the default ~/.claude root.
//
// Both are returned rather than just the winner because they genuinely
// co-exist — a session launched under a profile and one launched without it
// write to different directories and cannot see each other's memories. A
// health check that reported only one would hide half the store.
func memoryCandidateStores(projectRoot string) ([]memoryStore, error) {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("memory: resolve project root: %w", err)
	}
	slug := memoryProjectSlug(abs)

	var stores []memoryStore
	if cfg := os.Getenv(config.EnvClaudeConfigDir); cfg != "" {
		stores = append(stores, memoryStore{
			Dir:    filepath.Join(cfg, "projects", slug, "memory"),
			Origin: config.EnvClaudeConfigDir,
		})
	}
	home, err := userHomeDir()
	if err != nil {
		return stores, nil
	}
	def := filepath.Join(home, ".claude", "projects", slug, "memory")
	for _, s := range stores {
		if s.Dir == def {
			return stores, nil
		}
	}
	return append(stores, memoryStore{Dir: def, Origin: "default ~/.claude"}), nil
}

// memoryReport is the doctor result for one store.
type memoryReport struct {
	Store      memoryStore             `json:"store"`
	Exists     bool                    `json:"exists"`
	TopicFiles int                     `json:"topic_files"`
	Cap        int                     `json:"cap"`
	IndexLines int                     `json:"index_lines"`
	Findings   []taxonomy.AuditFinding `json:"findings"`
}

// newMemoryCmd builds `moai memory`.
func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Inspect and archive the Lessons Protocol memory store",
		Long: `Inspect and archive the topic-file memory store (MEMORY.md + topic files).

A session loads MEMORY.md, not the directory, so a topic file that was written
without its index line is stored and never recalled. ` + "`doctor`" + ` reports that
(and the other way round: an index line with no file behind it), plus the
per-project topic-file ceiling.

` + "`archive`" + ` retires named topic files into ` + memoryArchiveName + `/ and drops their index
lines. It never deletes — the archive preserves the audit trail. Which memory
is obsolete is a judgement, so the files are named by the operator; nothing is
selected automatically.`,
		GroupID: "tools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newMemoryDoctorCmd(), newMemoryArchiveCmd())
	return cmd
}

// newMemoryDoctorCmd — `moai memory doctor [--json] [--dir PATH]`.
func newMemoryDoctorCmd() *cobra.Command {
	var jsonOutput bool
	var dirOverride string
	var capOverride int

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Report memory-store health (orphans, dangling links, topic-file count)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reports, err := collectMemoryReports(dirOverride, capOverride)
			if err != nil {
				return err
			}
			return renderMemoryReports(cmd.OutOrStdout(), reports, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit the report as JSON on stdout")
	cmd.Flags().StringVar(&dirOverride, "dir", "", "Audit this memory directory instead of the resolved store")
	cmd.Flags().IntVar(&capOverride, "cap", 0,
		fmt.Sprintf("Topic-file ceiling (default %d)", config.DefaultMemoryTopicFileCap))
	return cmd
}

// collectMemoryReports audits every candidate store, or just the override.
func collectMemoryReports(dirOverride string, capOverride int) ([]memoryReport, error) {
	var stores []memoryStore
	if dirOverride != "" {
		abs, err := filepath.Abs(dirOverride)
		if err != nil {
			return nil, fmt.Errorf("memory: resolve --dir: %w", err)
		}
		stores = []memoryStore{{Dir: abs, Origin: "--dir"}}
	} else {
		var err error
		stores, err = memoryCandidateStores(resolveProjectDir())
		if err != nil {
			return nil, err
		}
	}

	capValue := capOverride
	if capValue <= 0 {
		capValue = config.DefaultMemoryTopicFileCap
	}

	reports := make([]memoryReport, 0, len(stores))
	for _, s := range stores {
		rep := memoryReport{Store: s, Cap: capValue}
		info, err := os.Stat(s.Dir)
		if err != nil || !info.IsDir() {
			reports = append(reports, rep)
			continue
		}
		rep.Exists = true

		entries, err := os.ReadDir(s.Dir)
		if err != nil {
			return nil, fmt.Errorf("memory: read %s: %w", s.Dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && e.Name() != memoryIndexName {
				rep.TopicFiles++
			}
		}

		indexPath := filepath.Join(s.Dir, memoryIndexName)
		if data, err := os.ReadFile(indexPath); err == nil {
			rep.IndexLines = len(strings.Split(strings.TrimRight(string(data), "\n"), "\n"))
		}

		linkage, err := taxonomy.AuditLinkage(s.Dir)
		if err != nil {
			return nil, err
		}
		count, err := taxonomy.AuditTopicCount(s.Dir, capValue)
		if err != nil {
			return nil, err
		}
		index, err := taxonomy.AuditIndex(indexPath, 0)
		if err != nil {
			return nil, err
		}
		rep.Findings = append(rep.Findings, linkage...)
		rep.Findings = append(rep.Findings, count...)
		rep.Findings = append(rep.Findings, index...)
		reports = append(reports, rep)
	}
	return reports, nil
}

// renderMemoryReports prints the doctor result. Findings are summarized by
// code rather than listed one per line: an 800-orphan store would otherwise
// bury its own headline under its own detail.
func renderMemoryReports(out io.Writer, reports []memoryReport, jsonOutput bool) error {
	if jsonOutput {
		data, err := json.Marshal(reports)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, string(data))
		return nil
	}

	for i, rep := range reports {
		if i > 0 {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintf(out, "%s  (%s)\n", rep.Store.Dir, rep.Store.Origin)
		if !rep.Exists {
			_, _ = fmt.Fprintln(out, "  not present")
			continue
		}
		_, _ = fmt.Fprintf(out, "  topic files : %d (cap %d)\n", rep.TopicFiles, rep.Cap)
		_, _ = fmt.Fprintf(out, "  index lines : %d\n", rep.IndexLines)

		if len(rep.Findings) == 0 {
			_, _ = fmt.Fprintln(out, "  findings    : none")
			continue
		}
		counts := map[taxonomy.AuditCode]int{}
		order := make([]taxonomy.AuditCode, 0, 4)
		for _, f := range rep.Findings {
			if counts[f.Code] == 0 {
				order = append(order, f.Code)
			}
			counts[f.Code]++
		}
		_, _ = fmt.Fprintln(out, "  findings    :")
		for _, code := range order {
			_, _ = fmt.Fprintf(out, "    %-30s %d\n", code, counts[code])
		}
		for _, f := range rep.Findings {
			if f.Code == taxonomy.WarnTopicCountOverCap || f.Code == taxonomy.WarnIndexOverflow {
				_, _ = fmt.Fprintf(out, "    → %s\n", f.Detail)
			}
		}
	}

	if len(reports) > 1 {
		_, _ = fmt.Fprintf(out, "\nTwo stores resolved. A session launched under a profile and one launched\n"+
			"without it write to different directories and cannot see each other's memories.\n")
	}
	return nil
}

// newMemoryArchiveCmd — `moai memory archive <name>... [--dir PATH]`.
func newMemoryArchiveCmd() *cobra.Command {
	var dirOverride string

	cmd := &cobra.Command{
		Use:   "archive <name>...",
		Short: "Retire named topic files into _archive/ and drop their index lines",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := dirOverride
			if dir == "" {
				stores, err := memoryCandidateStores(resolveProjectDir())
				if err != nil {
					return err
				}
				if len(stores) == 0 {
					return fmt.Errorf("memory archive: no memory store resolved")
				}
				dir = stores[0].Dir
			}
			abs, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("memory archive: resolve dir: %w", err)
			}
			return archiveMemoryFiles(cmd.OutOrStdout(), abs, args)
		},
	}
	cmd.Flags().StringVar(&dirOverride, "dir", "", "Archive within this memory directory instead of the resolved store")
	return cmd
}

// archiveMemoryFiles moves each named topic file into _archive/ and removes
// its line from the index. Every name is validated before anything moves, so
// a typo in the third name does not leave the first two half-archived.
func archiveMemoryFiles(out io.Writer, dir string, names []string) error {
	for _, raw := range names {
		name := filepath.Base(strings.TrimSpace(raw))
		if name == "" || name == "." || name == memoryIndexName {
			return fmt.Errorf("memory archive: %q is not a topic file", raw)
		}
		if !strings.HasSuffix(name, ".md") {
			return fmt.Errorf("memory archive: %q is not a .md file", raw)
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("memory archive: %s: %w", name, err)
		}
	}

	archiveDir := filepath.Join(dir, memoryArchiveName)
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return fmt.Errorf("memory archive: create %s: %w", archiveDir, err)
	}

	archived := make([]string, 0, len(names))
	for _, raw := range names {
		name := filepath.Base(strings.TrimSpace(raw))
		src := filepath.Join(dir, name)
		dst := filepath.Join(archiveDir, name)
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("memory archive: move %s: %w", name, err)
		}
		archived = append(archived, name)
		_, _ = fmt.Fprintf(out, "archived %s\n", name)
	}

	removed, err := dropMemoryIndexLines(filepath.Join(dir, memoryIndexName), archived)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "%d archived, %d index line(s) dropped\n", len(archived), removed)
	return nil
}

// dropMemoryIndexLines removes index lines linking any of names and reports
// how many went. A missing index is not an error: archiving a file the index
// never carried is exactly the orphan case doctor reports.
func dropMemoryIndexLines(indexPath string, names []string) (int, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("memory archive: read index: %w", err)
	}

	targeted := make(map[string]bool, len(names))
	for _, n := range names {
		targeted[n] = true
	}

	lines := strings.Split(string(data), "\n")
	kept := make([]string, 0, len(lines))
	removed := 0
	for _, line := range lines {
		drop := false
		for _, m := range markdownLinkTargetCLI.FindAllStringSubmatch(line, -1) {
			if targeted[filepath.Base(m[1])] {
				drop = true
				break
			}
		}
		if drop {
			removed++
			continue
		}
		kept = append(kept, line)
	}
	if removed == 0 {
		return 0, nil
	}

	if err := os.WriteFile(indexPath, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		return 0, fmt.Errorf("memory archive: write index: %w", err)
	}
	return removed, nil
}

// markdownLinkTargetCLI captures a markdown link target. Kept local to the
// CLI so the archive path does not depend on the audit package's unexported
// regex; the two patterns describe the same index format.
var markdownLinkTargetCLI = regexp.MustCompile(`\]\(([^)]+\.md)\)`)

func init() {
	rootCmd.AddCommand(newMemoryCmd())
}
