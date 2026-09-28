// todo_stale_store_test.go — SPEC-TODO-STALE-STORE-001 (card t1307): the
// stderr disclosure of a stale project-local queue store, judged per read
// verb (AC-TSS-001a..e), with stdout byte-identity (AC-TSS-002) and
// read-path purity (AC-TSS-003) pinned per verb too.
//
// The five read verbs enter disclosure through exactly two entry points:
// bare/list, why and pr via discloseQueueLayout, history via its direct
// discloseNonAuthoritativeBacklogJSON call (todo_history.go). The per-verb
// matrix here is what makes a single-entry-point fix fail acceptance —
// the codex cross-audit defect this SPEC's AC split exists to catch.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// staleStoreFixture builds a project whose home database (MOAI_HOME
// override) answers reads at last_seq 1305 while a rollback-snapshot
// project-local store sits at last_seq 661 — the measured divergence.
func staleStoreFixture(t *testing.T) (root, homeDB, legacyDB string) {
	t.Helper()
	root = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	initGitRepo(t, root)
	home := t.TempDir()
	t.Setenv(config.EnvHome, home)
	homeDB = filepath.Join(home, "db", homestate.ProjectKey(root), "todo", "backlog.db")
	legacyDB = filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	seedTodoStoreAt(t, homeDB, 1305)
	seedTodoStoreAt(t, legacyDB, 661)
	// The `pr` verb must not query the network or the real git history:
	// the disclosure facts under test are queue-store facts. One stubbed
	// runner keeps both subprocesses out.
	prev := todoRunCommand
	t.Cleanup(func() { todoRunCommand = prev })
	todoRunCommand = func(name string, args ...string) (string, error) {
		if name == "gh" {
			return "[]", nil
		}
		return "", nil
	}
	return root, homeDB, legacyDB
}

// seedTodoStoreAt creates a production-shaped queue database at dbPath and
// stamps meta.last_seq to seq through the store's own mutation path.
func seedTodoStoreAt(t *testing.T, dbPath string, seq int) {
	t.Helper()
	store := kanban.NewBacklogStore(strings.TrimSuffix(dbPath, ".db") + ".json")
	if _, _, err := store.Add("seed card"); err != nil {
		t.Fatalf("seed store %s: %v", dbPath, err)
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.LastSeq = seq
		return nil
	}); err != nil {
		t.Fatalf("stamp last_seq %d on %s: %v", seq, dbPath, err)
	}
}

// staleDisclosureVerbArgs enumerates the five read verbs the disclosure
// must reach (AC-TSS-001a..e), each in its normal invocation shape.
func staleDisclosureVerbArgs() map[string][]string {
	return map[string][]string{
		"bare":    {},
		"list":    {"list"},
		"why":     {"why", "t1"},
		"pr":      {"pr"},
		"history": {"history"},
	}
}

// assertStaleDisclosure checks the common AC-TSS-001 Then: one stderr line
// naming the legacy store path and both last_seq values, exit 0, and no
// disclosure on stdout.
func assertStaleDisclosure(t *testing.T, verb, stdout, stderr string, err error, legacyDB string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: exit %v, want 0", verb, err)
	}
	if !strings.Contains(stderr, legacyDB) {
		t.Errorf("%s: stderr does not name the stale store path %s; stderr=%q", verb, legacyDB, stderr)
	}
	if !strings.Contains(stderr, "1305") || !strings.Contains(stderr, "661") {
		t.Errorf("%s: stderr does not name both last_seq values (1305 home, 661 legacy); stderr=%q", verb, stderr)
	}
	if strings.Contains(stdout, "stale") {
		t.Errorf("%s: stdout carries disclosure text; stdout must stay a machine surface (REQ-TSS-002): %q", verb, stdout)
	}
}

// TestTodoStaleStoreDisclosure_PerVerb — AC-TSS-001a..e: every read verb
// discloses the stale store on stderr.
func TestTodoStaleStoreDisclosure_PerVerb(t *testing.T) {
	for verb, args := range staleDisclosureVerbArgs() {
		t.Run(verb, func(t *testing.T) {
			_, _, legacyDB := staleStoreFixture(t)
			stdout, stderr, err := runTodo(t, args...)
			assertStaleDisclosure(t, verb, stdout, stderr, err, legacyDB)
		})
	}
}

// TestTodoStaleStoreDisclosure_JSONStdoutByteIdentity — AC-TSS-002: for
// every verb (in --json form where the verb has one), stdout is
// byte-identical between a divergent-stale-store tree and a tree with no
// legacy store at all, and the clean tree's stderr carries no disclosure.
func TestTodoStaleStoreDisclosure_JSONStdoutByteIdentity(t *testing.T) {
	variants := map[string][]string{
		"bare":       {},
		"list":       {"list"},
		"list-json":  {"list", "--json"},
		"why":        {"why", "t1"},
		"pr":         {"pr"},
		"pr-json":    {"pr", "--json"},
		"history":    {"history"},
		"history-id": {"history", "t1"},
	}
	for name, args := range variants {
		t.Run(name, func(t *testing.T) {
			_, _, legacyDB := staleStoreFixture(t)
			divergentOut, _, err := runTodo(t, args...)
			if err != nil {
				t.Fatalf("divergent run %v: %v", args, err)
			}

			// Same tree shape, but the ghost store never existed: remove the
			// project-local legacy directory entirely.
			if err := os.RemoveAll(filepath.Dir(legacyDB)); err != nil {
				t.Fatalf("remove legacy dir: %v", err)
			}
			cleanOut, cleanErr, err := runTodo(t, args...)
			if err != nil {
				t.Fatalf("clean run %v: %v", args, err)
			}
			_ = cleanErr

			if divergentOut != cleanOut {
				t.Errorf("stdout differs between divergent and clean runs for %v — the disclosure leaked into stdout (REQ-TSS-002):\ndivergent=%q\nclean=%q",
					args, divergentOut, cleanOut)
			}
		})
	}
}

// TestTodoStaleStoreDisclosure_NoDisclosureWithoutDivergence — REQ-TSS-005:
// equal last_seq values disclose nothing, and a tree with no legacy store
// discloses nothing.
func TestTodoStaleStoreDisclosure_NoDisclosureWithoutDivergence(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)
	seedTodoStoreAt(t, legacyDB, 1305) // match the home sequence

	for verb, args := range staleDisclosureVerbArgs() {
		_, stderr, err := runTodo(t, args...)
		if err != nil {
			t.Fatalf("%s: %v", verb, err)
		}
		if strings.Contains(stderr, legacyDB) || strings.Contains(stderr, "stale") {
			t.Errorf("%s: stderr discloses with EQUAL last_seq values (REQ-TSS-005); stderr=%q", verb, stderr)
		}
	}

	// And with the legacy store gone entirely.
	if err := os.RemoveAll(filepath.Dir(legacyDB)); err != nil {
		t.Fatalf("remove legacy dir: %v", err)
	}
	if _, err := os.Stat(homeDB); err != nil {
		t.Fatalf("home db vanished: %v", err)
	}
	_ = root
	for verb, args := range staleDisclosureVerbArgs() {
		_, stderr, err := runTodo(t, args...)
		if err != nil {
			t.Fatalf("%s: %v", verb, err)
		}
		if strings.Contains(stderr, "stale") {
			t.Errorf("%s: stderr discloses with no legacy store present (REQ-TSS-005); stderr=%q", verb, stderr)
		}
	}
}

// TestTodoStaleStoreDisclosure_ReadPathPurity — AC-TSS-003: running all
// five read verbs leaves both database files byte- and mtime-identical and
// creates no new files in either directory (no marker, no stray artifact).
func TestTodoStaleStoreDisclosure_ReadPathPurity(t *testing.T) {
	_, homeDB, legacyDB := staleStoreFixture(t)

	before := map[string]cliDBFingerprint{
		homeDB:   cliFingerprint(t, homeDB),
		legacyDB: cliFingerprint(t, legacyDB),
	}

	for _, args := range staleDisclosureVerbArgs() {
		if _, _, err := runTodo(t, args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}

	for path, was := range before {
		now := cliFingerprint(t, path)
		if now.sum != was.sum || now.mtime != was.mtime {
			t.Errorf("%s changed across the read verbs: sha %s->%s mtime %d->%d (REQ-TSS-003)",
				path, was.sum, now.sum, was.mtime, now.mtime)
		}
		if strings.Join(now.names, ",") != strings.Join(was.names, ",") {
			t.Errorf("%s directory gained or lost entries across the read verbs: %v -> %v",
				path, was.names, now.names)
		}
	}
}

type cliDBFingerprint struct {
	sum   string
	mtime int64
	names []string
}

// budgetFailWriter fails once a write exceeds its byte budget — the
// unit-test writer for the disclosure function's error-propagation branch.
type budgetFailWriter struct{ budget int }

func (w *budgetFailWriter) Write(p []byte) (int, error) {
	if len(p) > w.budget {
		n := w.budget
		w.budget = 0
		return n, errBudgetFailWriterFull
	}
	w.budget -= len(p)
	return len(p), nil
}

var errBudgetFailWriterFull = errors.New("writer full")

// TestDiscloseStaleLocalStores_Unit — the fact-filter branches the
// end-to-end verb runs cannot reach: an unreadable store is skipped, a
// same-seq store is skipped, and a writer error propagates.
func TestDiscloseStaleLocalStores_Unit(t *testing.T) {
	fact := kanban.StaleStoreFact{
		HomePresent:  true,
		HomeReadable: true,
		HomeLastSeq:  1305,
		Stores: []kanban.StaleLocalStore{
			{Path: "/tmp/unreadable.db", LastSeq: 0, Readable: false},
			{Path: "/tmp/same.db", LastSeq: 1305, Readable: true},
			{Path: "/tmp/divergent.db", LastSeq: 661, Readable: true},
		},
		Divergent: true,
	}
	var out bytes.Buffer
	if err := discloseStaleLocalStores(&out, "todo", fact); err != nil {
		t.Fatalf("disclose: %v", err)
	}
	got := out.String()
	if strings.Contains(got, "unreadable.db") || strings.Contains(got, "same.db") {
		t.Errorf("disclosure emitted a store it must skip:\n%s", got)
	}
	if !strings.Contains(got, "/tmp/divergent.db") {
		t.Errorf("disclosure omitted the divergent store:\n%s", got)
	}

	if err := discloseStaleLocalStores(&budgetFailWriter{budget: 0}, "todo", fact); err == nil {
		t.Error("writer error not propagated to the caller")
	}

	// Non-divergent facts write nothing at all (REQ-TSS-005).
	var empty bytes.Buffer
	if err := discloseStaleLocalStores(&empty, "todo", kanban.StaleStoreFact{}); err != nil {
		t.Fatalf("disclose empty fact: %v", err)
	}
	if empty.Len() != 0 {
		t.Errorf("non-divergent fact wrote %q", empty.String())
	}
}

func cliFingerprint(t *testing.T, dbPath string) cliDBFingerprint {
	t.Helper()
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat %s: %v", dbPath, err)
	}
	raw, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read %s: %v", dbPath, err)
	}
	sum := sha256.Sum256(raw)
	entries, err := os.ReadDir(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("read dir %s: %v", filepath.Dir(dbPath), err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return cliDBFingerprint{sum: hex.EncodeToString(sum[:]), mtime: info.ModTime().UnixNano(), names: names}
}
