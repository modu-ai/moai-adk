package feedback

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// homeAlias is what an absolute home prefix collapses to.
const homeAlias = "~"

// resolveHome returns the home directory used for path collapse: the explicit
// override when set, otherwise paths.Home(), which reads HOME first so a
// t.Setenv override is honoured on every platform.
//
// A resolution failure collapses no paths and fails nothing else — a report
// that names no home path is still worth masking for secrets.
func resolveHome(opt Options) string {
	if opt.Home != "" {
		return opt.Home
	}
	home, err := paths.Home()
	if err != nil {
		slog.Warn("feedback: home directory unresolved, skipping path collapse", "error", err)
		return ""
	}
	return home
}

// collapseHome rewrites an absolute home prefix to "~". It reports how many
// occurrences it replaced.
func collapseHome(s, home string) (string, int) {
	home = strings.TrimRight(home, `/\`)
	if s == "" || home == "" {
		return s, 0
	}

	count := 0
	for _, sep := range separators() {
		prefix := home + sep
		n := strings.Count(s, prefix)
		if n == 0 {
			continue
		}
		s = strings.ReplaceAll(s, prefix, homeAlias+sep)
		count += n
	}
	return s, count
}

// separators returns the path separators a report may use. A report written on
// Windows can carry either form, so both are collapsed.
func separators() []string {
	seps := []string{"/"}
	if native := string(filepath.Separator); native != "/" {
		seps = append(seps, native)
	}
	return seps
}

// ResolveProjectRoot walks up from start looking for the .moai marker
// directory and returns the directory that holds it.
//
// It reports false rather than guessing. The on-disk artefacts (the mask log
// and the retry queue) are written under the resolved root, and writing them
// into an unrelated directory is worse than not writing them at all — the
// scrub itself never depends on the root being found.
func ResolveProjectRoot(start string) (string, bool) {
	dir := start
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		dir = wd
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}

	for {
		if info, err := os.Stat(filepath.Join(dir, defs.MoAIDir)); err == nil && info.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
