// sets.go — generation and read-only check of the rendered agent copies.
//
// A Layout names the agent source directory and the target sets rendered from
// it. Update writes every set; Check renders the same sets in memory, compares
// them with the files on disk and never writes. Both paths run the same
// renderer, so a file the check accepts is exactly the file Update would write.
package agentemit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Target is one set of rendered files.
type Target struct {
	Name  string   // set name used in diagnostics: claude_template, codex_template, claude_project, codex_project
	Dir   string   // directory receiving the files, slash-separated, relative to the repository root
	Codex bool     // true: Codex TOML files; false: Claude Markdown files
	Only  []string // agent names the set covers; empty means every source
}

// Layout names the agent source directory and the target sets of one run.
type Layout struct {
	SourceDir string // slash-separated, relative to the repository root
	Targets   []Target
}

// Summary is the result of a Check or Update run.
type Summary struct {
	Specs  int      // agent sources rendered
	Copies int      // target sets compared or written
	Diffs  []string // one entry per differing file, sorted; empty when the sets match
}

// String renders the summary line `make agents-emit-check` prints.
func (s Summary) String() string {
	return fmt.Sprintf("specs=%d copies=%d diff=%d", s.Specs, s.Copies, len(s.Diffs))
}

// CarriedFour names the agents whose project copies (the local .claude and
// .codex definitions) are generated too. The project copies of the other
// agents stay hand-managed until the follow-on piece absorbs or retires them.
var CarriedFour = []string{"builder-harness", "e2e-tester", "manager-design", "super-advisor"}

// DefaultLayout is the layout `make agents-emit` and `make agents-emit-check` use:
// the template Markdown (C2) and Codex (C3) sets for every source, and the
// project Markdown (C1) and Codex (C3p) sets for the carried four.
func DefaultLayout() Layout {
	return Layout{
		SourceDir: "internal/template/agentemit/source",
		Targets: []Target{
			{Name: "claude_template", Dir: "internal/template/templates/.claude/agents/moai"},
			{Name: "codex_template", Dir: "internal/template/templates/.codex/agents/moai", Codex: true},
			{Name: "claude_project", Dir: ".claude/agents/moai", Only: CarriedFour},
			{Name: "codex_project", Dir: ".codex/agents/moai", Codex: true, Only: CarriedFour},
		},
	}
}

// renderTargets renders every target set from the agent sources under root and
// returns the files keyed by their slash-separated path relative to root,
// together with the number of sources. Nothing is written.
func renderTargets(root string, lay Layout, man Manifest) (map[string][]byte, int, error) {
	pub, err := EmitAll(os.DirFS(root), lay.SourceDir, man)
	if err != nil {
		return nil, 0, err
	}
	names := make(map[string]bool, len(pub.Markdown))
	for key := range pub.Markdown {
		names[strings.TrimSuffix(path.Base(key), ".md")] = true
	}
	files := map[string][]byte{}
	for _, tg := range lay.Targets {
		for _, only := range tg.Only {
			if !names[only] {
				return nil, 0, fmt.Errorf("agentemit: set %s names agent %q, which has no source under %s", tg.Name, only, lay.SourceDir)
			}
		}
		for key, md := range pub.Markdown {
			if len(tg.Only) > 0 && !contains(tg.Only, strings.TrimSuffix(path.Base(key), ".md")) {
				continue
			}
			if tg.Codex {
				tomlPath := codexTOMLPath(man, key)
				files[path.Join(tg.Dir, path.Base(tomlPath))] = pub.CodexTOML[tomlPath]
			} else {
				files[path.Join(tg.Dir, path.Base(key))] = md
			}
		}
	}
	return files, len(pub.Markdown), nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func sortedPaths(files map[string][]byte) []string {
	out := make([]string, 0, len(files))
	for p := range files {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// Check renders the sets in memory and compares them with the files under
// root. It never writes. Every difference — a missing file, a file whose bytes
// differ from a fresh render, and (for a set that covers every source) a file
// with no agent source — is listed in Summary.Diffs, and the error names the
// first one.
func Check(root string, lay Layout, man Manifest) (Summary, error) {
	files, specs, err := renderTargets(root, lay, man)
	sum := Summary{Specs: specs, Copies: len(lay.Targets)}
	if err != nil {
		return sum, err
	}
	for _, p := range sortedPaths(files) {
		got, rerr := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		switch {
		case errors.Is(rerr, fs.ErrNotExist):
			sum.Diffs = append(sum.Diffs, p+": missing; run `make agents-emit`")
		case rerr != nil:
			sum.Diffs = append(sum.Diffs, p+": unreadable: "+rerr.Error())
		case !bytes.Equal(got, files[p]):
			sum.Diffs = append(sum.Diffs, p+": differs from a fresh render of its agent source; edit the source and run `make agents-emit`")
		}
	}
	for _, tg := range lay.Targets {
		if len(tg.Only) > 0 {
			continue // a project set sits beside hand-managed definitions of other agents
		}
		ext := ".md"
		if tg.Codex {
			ext = ".toml"
		}
		entries, derr := os.ReadDir(filepath.Join(root, filepath.FromSlash(tg.Dir)))
		if derr != nil {
			continue // an absent directory is already reported as missing files
		}
		for _, e := range entries {
			p := path.Join(tg.Dir, e.Name())
			if _, rendered := files[p]; !e.IsDir() && strings.HasSuffix(e.Name(), ext) && !rendered {
				sum.Diffs = append(sum.Diffs, p+": no agent source renders this file")
			}
		}
	}
	sort.Strings(sum.Diffs)
	if len(sum.Diffs) > 0 {
		return sum, fmt.Errorf("agents-emit drift: %d file(s) differ, first: %s", len(sum.Diffs), sum.Diffs[0])
	}
	return sum, nil
}

// Update renders the sets and writes the files under root. Rendering finishes
// before the first write, so a source that cannot be rendered leaves root
// untouched. A file that already holds the rendered bytes is not rewritten.
func Update(root string, lay Layout, man Manifest) (Summary, error) {
	files, specs, err := renderTargets(root, lay, man)
	sum := Summary{Specs: specs, Copies: len(lay.Targets)}
	if err != nil {
		return sum, err
	}
	for _, p := range sortedPaths(files) {
		dst := filepath.Join(root, filepath.FromSlash(p))
		if cur, rerr := os.ReadFile(dst); rerr == nil && bytes.Equal(cur, files[p]) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return sum, fmt.Errorf("agentemit: %w", err)
		}
		if err := os.WriteFile(dst, files[p], 0o644); err != nil {
			return sum, fmt.Errorf("agentemit: %w", err)
		}
	}
	return sum, nil
}

// printf writes a line to a command stream. A failed write to stdout or stderr
// has no recovery path inside the command, so the error is deliberately dropped.
func printf(w io.Writer, format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }

// RunCLI is the body of the agents-emit command: with `-check` it compares
// (read-only) and prints the summary line on success, otherwise it regenerates
// every set. It returns the process exit code.
func RunCLI(root string, lay Layout, man Manifest, args []string, stdout, stderr io.Writer) int {
	check := false
	for _, a := range args {
		if a != "-check" {
			printf(stderr, "agents-emit: unknown argument %q (usage: agents-emit [-check])\n", a)
			return 2
		}
		check = true
	}
	if check {
		sum, err := Check(root, lay, man)
		if err != nil {
			printf(stderr, "%v\n", err)
			for _, d := range sum.Diffs[min(1, len(sum.Diffs)):] {
				printf(stderr, "  %s\n", d)
			}
			return 1
		}
		printf(stdout, "agents-emit-check: %s\n", sum.String())
		return 0
	}
	sum, err := Update(root, lay, man)
	if err != nil {
		printf(stderr, "%v\n", err)
		return 1
	}
	printf(stdout, "agents-emit: specs=%d copies=%d\n", sum.Specs, sum.Copies)
	return 0
}
