// doctor_db_ghosts.go — SPEC-MOAI-DB-PATHS-001 (card t1827) REQ-DBP-015: the
// doctor "DB Path Ghosts" check.
//
// An agent that guesses a database path and "just looks" with sqlite3 leaves a
// zero-byte file behind. Instructions are a probabilistic fix; a ghost that
// reappears is the one mechanical signal they failed, so this check lists every
// zero-byte regular *.db file directly inside the directories a guessed path
// lands in, minus the canonical databases `moai paths` reports.
//
// Scanned directories: the moai home, its db/ directory, the project state
// directory with its todo/ and factory/ subdirectories, <root>/.moai,
// <root>/.moai/state, and its todo/ and factory/ subdirectories. A sweep of the
// tree for legitimate *.db writers there found only backlog.db and factory.db
// (canonical, excluded); broker.db, sessions.db, and profile-leases.db live
// below other directories.
//
// The check only reads: it never deletes, moves, or opens a file, and it never
// fails the doctor run — an unreadable directory is skipped.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// dbPathGhostCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const dbPathGhostCheckName = "DB Path Ghosts"

// checkDBPathGhosts reports zero-byte *.db files at unconfirmed paths.
func checkDBPathGhosts(projectRoot string, verbose bool) DiagnosticCheck {
	root := factory.ResolveTodoQueueRoot(projectRoot)

	canonical := map[string]bool{}
	markCanonical := func(p string) {
		canonical[filepath.Clean(p)] = true
		if r, err := filepath.EvalSymlinks(filepath.Dir(p)); err == nil {
			canonical[filepath.Join(r, filepath.Base(p))] = true
		}
	}
	markCanonical(factory.QueueLocationForRoot(root).DBPath)

	dirs := map[string]bool{}
	addTree := func(base string) {
		dirs[base] = true
		dirs[filepath.Join(base, "todo")] = true
		dirs[filepath.Join(base, "factory")] = true
	}
	if home, err := paths.MoaiHome(); err == nil {
		dirs[home] = true
		dirs[filepath.Join(home, "db")] = true
	}
	if dir, err := pathsProjectDir(root); err == nil {
		addTree(dir)
	}
	if p, err := pathsFactoryDBPath(root); err == nil {
		markCanonical(p)
	}
	dirs[filepath.Join(root, ".moai")] = true
	addTree(filepath.Join(root, ".moai", "state"))

	var ghosts []string
	unreadable := 0
	for dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				unreadable++
			}
			continue
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".db") || !e.Type().IsRegular() {
				continue
			}
			full := filepath.Join(dir, e.Name())
			if canonical[full] {
				continue
			}
			if info, err := e.Info(); err == nil && info.Size() == 0 {
				ghosts = append(ghosts, full)
			}
		}
	}
	sort.Strings(ghosts)

	note := ""
	if unreadable > 0 {
		note = fmt.Sprintf(" (%d director(ies) unreadable and skipped)", unreadable)
	}
	if len(ghosts) == 0 {
		return DiagnosticCheck{
			Name:    dbPathGhostCheckName,
			Status:  uikit.CheckOK,
			Message: "no zero-byte database files at unconfirmed paths" + note,
		}
	}
	return DiagnosticCheck{
		Name:   dbPathGhostCheckName,
		Status: uikit.CheckWarn,
		Message: fmt.Sprintf("%d zero-byte database file(s) at unconfirmed paths — a guessed path was opened with sqlite3; run moai paths for the confirmed ones%s",
			len(ghosts), note),
		Detail: strings.Join(ghosts, "; "),
	}
}
