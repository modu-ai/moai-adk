// paths_test.go — SPEC-MOAI-DB-PATHS-001 (card t1827) M1/M2: `moai paths`.
//
// Every test pins an absolute MOAI_HOME under t.TempDir() (home layout) or
// clears it (temporary-origin layout), points CLAUDE_PROJECT_DIR at a committed
// scratch repository, and never resolves the operator's real home. All tests
// are serial: they mutate process-global environment through t.Setenv.
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// runPaths executes a fresh `moai paths` command against args.
func runPaths(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := newPathsCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errBuf.String(), err
}

// pathsDoc runs `paths --json` (plus extra args), requires success, and decodes
// stdout into a generic map.
func pathsDoc(t *testing.T, extra ...string) map[string]any {
	t.Helper()
	out, errOut, err := runPaths(t, append([]string{"--json"}, extra...)...)
	if err != nil {
		t.Fatalf("paths --json %v: %v (stderr %q)", extra, err, errOut)
	}
	return decodePathsDoc(t, out)
}

func decodePathsDoc(t *testing.T, out string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v\n%s", err, out)
	}
	return doc
}

// docGet walks a dotted key path ("todo.path") through a decoded document.
func docGet(t *testing.T, doc map[string]any, dotted string) any {
	t.Helper()
	var cur any = doc
	for _, part := range strings.Split(dotted, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("key %q: %T is not an object", dotted, cur)
		}
		v, present := m[part]
		if !present {
			t.Fatalf("key %q missing from %v", dotted, m)
		}
		cur = v
	}
	return cur
}

func docString(t *testing.T, doc map[string]any, dotted string) string {
	t.Helper()
	s, ok := docGet(t, doc, dotted).(string)
	if !ok {
		t.Fatalf("key %q is not a string: %v", dotted, docGet(t, doc, dotted))
	}
	return s
}

// pathsHomeFixture builds a committed scratch repository whose project state
// lives under an absolute MOAI_HOME inside the test's temp directory.
func pathsHomeFixture(t *testing.T) (root, home string) {
	t.Helper()
	root = t.TempDir()
	home = filepath.Join(t.TempDir(), "moai-home")
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv(config.EnvHome, home)
	initGitRepo(t, root)
	seedGitFlowPrecondition(t, root)
	return root, home
}

// pathsTempOriginFixture builds a committed scratch repository with no
// MOAI_HOME: the temporary-origin layout, where state stays project-local.
func pathsTempOriginFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv(config.EnvHome, "")
	initGitRepo(t, root)
	seedGitFlowPrecondition(t, root)
	return root
}

// pathsTreeSnapshot records every entry below dirs with size and mtime, so a
// before/after comparison sees a created, moved, resized, or touched file.
func pathsTreeSnapshot(t *testing.T, dirs ...string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			out[path] = fmt.Sprintf("%v|%d|%d", info.IsDir(), info.Size(), info.ModTime().UnixNano())
			return nil
		})
		if err != nil {
			t.Fatalf("snapshot %s: %v", dir, err)
		}
	}
	return out
}

func snapshotDiff(before, after map[string]string) []string {
	var diff []string
	for k, v := range after {
		if bv, ok := before[k]; !ok {
			diff = append(diff, "created "+k)
		} else if bv != v {
			diff = append(diff, "changed "+k)
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			diff = append(diff, "removed "+k)
		}
	}
	sort.Strings(diff)
	return diff
}

// findFiles returns every regular file named name below root, skipping .git.
func findFiles(t *testing.T, root, name string) []string {
	t.Helper()
	var found []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == name {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return found
}

func sameFile(a, b string) bool { return sameDir(a, b) }

// ---------------------------------------------------------------------------
// AC-DBP-001 — shape, single object, existence/size/companions, listing
// ---------------------------------------------------------------------------

func TestPaths_JSONShape(t *testing.T) {
	pathsHomeFixture(t)
	doc := pathsDoc(t)

	for _, key := range []string{"project_key", "project_dir", "queue_layout"} {
		if s := docString(t, doc, key); s == "" {
			t.Errorf("%s is empty", key)
		}
	}
	if got := docString(t, doc, "queue_layout"); got != "home" {
		t.Errorf("queue_layout = %q, want home", got)
	}
	dbKeys := []string{"path", "exists", "size_bytes", "wal.exists", "wal.size_bytes", "shm.exists", "shm.size_bytes"}
	for _, key := range dbKeys {
		docGet(t, doc, "todo."+key)
		docGet(t, doc, "factory."+key)
	}
	for _, key := range []string{"dir", "queue_state", "legacy_dir"} {
		docGet(t, doc, "todo."+key)
	}
	if v := docGet(t, doc, "todo.legacy_dir"); v != nil {
		t.Errorf("todo.legacy_dir = %v, want null for a canonical queue", v)
	}
	if got := docString(t, doc, "todo.queue_state"); got != "canonical" {
		t.Errorf("todo.queue_state = %q, want canonical", got)
	}
	docGet(t, doc, "broker.path_template")
	docGet(t, doc, "broker.existing_runs")
}

func TestPaths_JSONSingleObject(t *testing.T) {
	pathsHomeFixture(t)
	out, _, err := runPaths(t, "--json")
	if err != nil {
		t.Fatalf("paths --json: %v", err)
	}
	dec := json.NewDecoder(strings.NewReader(out))
	var first map[string]any
	if err := dec.Decode(&first); err != nil {
		t.Fatalf("first JSON value: %v\n%s", err, out)
	}
	if len(first) == 0 {
		t.Fatalf("the JSON object is empty: %s", out)
	}
	var second any
	if err := dec.Decode(&second); !errors.Is(err, io.EOF) {
		t.Errorf("stdout carries more than one JSON value (second decode err = %v)\n%s", err, out)
	}
}

func TestPaths_ExistenceSizeWalShm(t *testing.T) {
	root, _ := pathsHomeFixture(t)
	doc := pathsDoc(t)
	todoDB := docString(t, doc, "todo.path")
	factoryDB := docString(t, doc, "factory.path")
	if docGet(t, doc, "todo.exists") != false || docGet(t, doc, "factory.exists") != false {
		t.Fatalf("a fresh project must report both databases absent: %v", doc)
	}

	write := func(path string, size int) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, bytes.Repeat([]byte("x"), size), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(todoDB, 5)
	write(todoDB+"-wal", 3)
	write(todoDB+"-shm", 2)
	write(factoryDB+"-wal", 7) // a -wal with no database: reported as such, no database inferred
	before := pathsTreeSnapshot(t, root)

	doc = pathsDoc(t)
	want := map[string]any{
		"todo.exists": true, "todo.size_bytes": float64(5),
		"todo.wal.exists": true, "todo.wal.size_bytes": float64(3),
		"todo.shm.exists": true, "todo.shm.size_bytes": float64(2),
		"factory.exists": false, "factory.size_bytes": float64(0),
		"factory.wal.exists": true, "factory.wal.size_bytes": float64(7),
		"factory.shm.exists": false, "factory.shm.size_bytes": float64(0),
	}
	for key, w := range want {
		if got := docGet(t, doc, key); got != w {
			t.Errorf("%s = %v, want %v", key, got, w)
		}
	}
	if diff := snapshotDiff(before, pathsTreeSnapshot(t, root)); len(diff) > 0 {
		t.Errorf("the project tree changed: %v", diff)
	}

	// A zero-byte database at a canonical path is a database, never a ghost.
	write(todoDB, 0)
	doc = pathsDoc(t)
	if docGet(t, doc, "todo.exists") != true || docGet(t, doc, "todo.size_bytes") != float64(0) {
		t.Errorf("a zero-byte canonical database must report exists=true size_bytes=0: %v", doc["todo"])
	}
}

var pathsListingLine = regexp.MustCompile(`^([a-z][a-z-]*): (.*)$`)

func TestPaths_HumanListing(t *testing.T) {
	pathsHomeFixture(t)
	out, _, err := runPaths(t)
	if err != nil {
		t.Fatalf("paths: %v", err)
	}
	if json.Valid([]byte(out)) {
		t.Errorf("the human listing must not be valid JSON:\n%s", out)
	}
	var names []string
	values := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		m := pathsListingLine.FindStringSubmatch(line)
		if m == nil {
			t.Errorf("line %q is not a `name: value` line", line)
			continue
		}
		names = append(names, m[1])
		values[m[1]] = m[2]
	}
	wantFirst := []string{"project-key", "project-dir", "todo-dir", "todo-db", "factory-db"}
	if len(names) < len(wantFirst) {
		t.Fatalf("listing has %d lines, want at least %d:\n%s", len(names), len(wantFirst), out)
	}
	for i, w := range wantFirst {
		if names[i] != w {
			t.Errorf("line %d name = %q, want %q (listing order is part of the contract)", i+1, names[i], w)
		}
	}
	doc := pathsDoc(t)
	for name, key := range map[string]string{
		"project-key": "project_key", "project-dir": "project_dir",
		"todo-dir": "todo.dir", "todo-db": "todo.path", "factory-db": "factory.path",
	} {
		if values[name] != docString(t, doc, key) {
			t.Errorf("listing %s = %q, JSON %s = %q", name, values[name], key, docString(t, doc, key))
		}
	}
}

// ---------------------------------------------------------------------------
// AC-DBP-002 — parity with the verbs, pre/post adoption
// ---------------------------------------------------------------------------

func TestPaths_ParityWithVerbs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T) (root, walkRoot string)
	}{
		{"home", func(t *testing.T) (string, string) {
			root, home := pathsHomeFixture(t)
			return root, home
		}},
		{"temp-origin", func(t *testing.T) (string, string) {
			root := pathsTempOriginFixture(t)
			return root, root
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, walkRoot := tc.setup(t)
			if _, _, err := runTodo(t, "add", "parity card"); err != nil {
				t.Fatalf("todo add: %v", err)
			}
			db, err := homestate.OpenFactory(factoryCardRoot())
			if err != nil {
				t.Fatalf("open factory: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("close factory: %v", err)
			}

			doc := pathsDoc(t)
			for _, c := range []struct{ key, file string }{
				{"todo.path", "backlog.db"}, {"factory.path", "factory.db"},
			} {
				reported := docString(t, doc, c.key)
				info, err := os.Stat(reported)
				if err != nil {
					t.Errorf("%s = %q does not exist after the verb created its database: %v", c.key, reported, err)
					continue
				}
				if info.Size() == 0 {
					t.Errorf("%s = %q exists but is empty", c.key, reported)
				}
				created := findFiles(t, walkRoot, c.file)
				if len(created) != 1 || !sameFile(created[0], reported) {
					t.Errorf("%s = %q, but the verbs created %v", c.key, reported, created)
				}
			}
			if docGet(t, doc, "todo.exists") != true || docGet(t, doc, "factory.exists") != true {
				t.Errorf("both databases exist now, the report says todo=%v factory=%v",
					docGet(t, doc, "todo.exists"), docGet(t, doc, "factory.exists"))
			}
			if want := map[string]string{"home": "home", "temp-origin": "project-local"}[tc.name]; docString(t, doc, "queue_layout") != want {
				t.Errorf("queue_layout = %q, want %q", docString(t, doc, "queue_layout"), want)
			}
			_ = root
		})
	}
}

func TestPaths_LegacyQueuePreAdoption(t *testing.T) {
	root, home := pathsHomeFixture(t)
	legacy := filepath.Join(root, ".moai", "state", "todo")
	if _, _, err := factory.NewBacklogStore(filepath.Join(legacy, "backlog.json")).Add("legacy card"); err != nil {
		t.Fatalf("seed the legacy queue: %v", err)
	}
	var canonicalBefore string

	t.Run("pre-adoption", func(t *testing.T) {
		before := pathsTreeSnapshot(t, root, home)
		doc := pathsDoc(t)
		canonicalBefore = docString(t, doc, "todo.path")
		if got := docString(t, doc, "todo.queue_state"); got != "pre-adoption" {
			t.Errorf("todo.queue_state = %q, want pre-adoption", got)
		}
		if docGet(t, doc, "todo.exists") != false {
			t.Errorf("todo.exists = %v, want false before the adoption", docGet(t, doc, "todo.exists"))
		}
		if got := docString(t, doc, "todo.legacy_dir"); !sameDir(got, legacy) {
			t.Errorf("todo.legacy_dir = %q, want %q", got, legacy)
		}
		if !strings.HasPrefix(canonicalBefore, home) {
			t.Errorf("todo.path = %q is not below the moai home %q", canonicalBefore, home)
		}
		if diff := snapshotDiff(before, pathsTreeSnapshot(t, root, home)); len(diff) > 0 {
			t.Errorf("reporting moved or created files: %v", diff)
		}
	})

	t.Run("post-adoption", func(t *testing.T) {
		if _, _, err := runTodo(t, "add", "adopting card"); err != nil {
			t.Fatalf("adopting todo add: %v", err)
		}
		doc := pathsDoc(t)
		if got := docString(t, doc, "todo.path"); got != canonicalBefore {
			t.Errorf("todo.path changed across the adoption: %q -> %q", canonicalBefore, got)
		}
		if docGet(t, doc, "todo.exists") != true {
			t.Errorf("todo.exists = %v, want true after the adoption", docGet(t, doc, "todo.exists"))
		}
		if got := docString(t, doc, "todo.queue_state"); got != "canonical" {
			t.Errorf("todo.queue_state = %q, want canonical", got)
		}
		if v := docGet(t, doc, "todo.legacy_dir"); v != nil {
			t.Errorf("todo.legacy_dir = %v, want null after the adoption", v)
		}
		if got, want := docString(t, doc, "todo.dir"), filepath.Dir(docString(t, doc, "todo.path")); got != want {
			t.Errorf("todo.dir = %q, want the directory part of todo.path %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// AC-DBP-003 — read-only by behaviour and by source
// ---------------------------------------------------------------------------

func TestPaths_NeverCreatesFiles(t *testing.T) {
	root, home := pathsHomeFixture(t)
	factoryDir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(factoryDir, "messages", "r1"), 0o700); err != nil {
		t.Fatal(err)
	}

	combos := [][]string{
		{},
		{"--json"},
		{"--field", "project-key"}, {"--field", "project-dir"}, {"--field", "todo-dir"},
		{"--field", "todo-db"}, {"--field", "factory-db"},
		{"--run", "r1"}, {"--run", "r1", "--json"},
	}
	failing := [][]string{{"--run", "../x"}, {"--field", "bogus"}}

	before := pathsTreeSnapshot(t, root, home)
	for _, args := range combos {
		out, _, err := runPaths(t, args...)
		if err != nil {
			t.Errorf("paths %v: %v", args, err)
		}
		if strings.TrimSpace(out) == "" {
			t.Errorf("paths %v wrote nothing to stdout", args)
		}
	}
	for _, args := range failing {
		if _, _, err := runPaths(t, args...); err == nil {
			t.Errorf("paths %v must fail", args)
		}
	}
	if diff := snapshotDiff(before, pathsTreeSnapshot(t, root, home)); len(diff) > 0 {
		t.Errorf("`moai paths` changed the moai home or the project root: %v", diff)
	}

	// Positive control: the same detector must see a layout-ensuring call.
	if err := homestate.EnsureProjectLayout(root); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	if diff := snapshotDiff(before, pathsTreeSnapshot(t, root, home)); len(diff) == 0 {
		t.Errorf("positive control: the detector did not notice homestate.EnsureProjectLayout creating files")
	}
}

// pathsForbiddenSelectors are the identifiers `moai paths` must never reach:
// each opens a database, creates or relocates something, or returns a path other
// than the canonical one (BacklogPathForRoot can name a legacy directory).
var pathsForbiddenSelectors = map[string]bool{
	"OpenFactory": true, "OpenFactoryPath": true, "OpenFactoryBounded": true, "OpenFactoryReadonly": true,
	"EnsureProjectLayout": true, "EnsureHomeLayout": true, "FactoryRegistryPath": true,
	"BacklogPathForRoot": true, "BacklogPathForRootAdopting": true,
	"MkdirAll": true, "Mkdir": true, "WriteFile": true, "Create": true, "OpenFile": true,
	"Rename": true, "Remove": true, "RemoveAll": true,
}

// forbiddenPathsCalls parses src and reports every forbidden selector, plus any
// Open* member of a database-facing package.
func forbiddenPathsCalls(t *testing.T, filename, src string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, _ := sel.X.(*ast.Ident)
		switch {
		case pathsForbiddenSelectors[sel.Sel.Name]:
			found = append(found, sel.Sel.Name)
		case id != nil && sel.Sel.Name == "Join" && (id.Name == "filepath" || id.Name == "path"):
			// the command never joins path segments of its own (REQ-DBP-002)
			found = append(found, id.Name+".Join")
		case id != nil && strings.HasPrefix(sel.Sel.Name, "Open"):
			switch id.Name {
			case "homestate", "factorymsg", "sql", "sqlite3":
				found = append(found, id.Name+"."+sel.Sel.Name)
			}
		}
		return true
	})
	return found
}

func TestPaths_SourceHasNoOpeningCalls(t *testing.T) {
	// Positive control: the detector finds a known-bad call and the list names
	// the two resolvers that would let the command report a legacy directory.
	if got := forbiddenPathsCalls(t, "bad.go", "package x\nfunc f(){ _, _ = homestate.OpenFactory(\"r\") }"); len(got) == 0 {
		t.Fatalf("positive control: the detector missed homestate.OpenFactory")
	}
	if got := forbiddenPathsCalls(t, "bad.go", "package x\nfunc f(){ _ = filepath.Join(\"a\", \"todo\") }"); len(got) == 0 {
		t.Fatalf("positive control: the detector missed filepath.Join")
	}
	if got := forbiddenPathsCalls(t, "ok.go", "package x\nfunc f(){ _ = strings.Join(nil, \",\") }"); len(got) != 0 {
		t.Fatalf("strings.Join joins no path and must not be flagged: %v", got)
	}
	for _, name := range []string{"BacklogPathForRoot", "BacklogPathForRootAdopting", "FactoryRegistryPath", "EnsureProjectLayout"} {
		if !pathsForbiddenSelectors[name] {
			t.Errorf("the forbidden list lacks %s", name)
		}
	}

	src, err := os.ReadFile("paths.go")
	if err != nil {
		t.Fatalf("read paths.go: %v", err)
	}
	if got := forbiddenPathsCalls(t, "paths.go", string(src)); len(got) > 0 {
		t.Errorf("paths.go reaches forbidden calls: %v", got)
	}
}

// ---------------------------------------------------------------------------
// AC-DBP-004 — --field
// ---------------------------------------------------------------------------

func TestPaths_FieldValues(t *testing.T) {
	pathsHomeFixture(t)
	doc := pathsDoc(t)
	for field, key := range map[string]string{
		"project-key": "project_key", "project-dir": "project_dir",
		"todo-dir": "todo.dir", "todo-db": "todo.path", "factory-db": "factory.path",
	} {
		t.Run(field, func(t *testing.T) {
			out, _, err := runPaths(t, "--field", field)
			if err != nil {
				t.Fatalf("paths --field %s: %v", field, err)
			}
			if want := docString(t, doc, key) + "\n"; out != want {
				t.Errorf("stdout = %q, want %q", out, want)
			}
		})
	}
	todoDir, _, _ := runPaths(t, "--field", "todo-dir")
	todoDB, _, _ := runPaths(t, "--field", "todo-db")
	if strings.TrimSpace(todoDir) != filepath.Dir(strings.TrimSpace(todoDB)) {
		t.Errorf("todo-dir %q is not the directory part of todo-db %q", todoDir, todoDB)
	}
}

func TestPaths_FieldUnknown(t *testing.T) {
	pathsHomeFixture(t)
	t.Run("unknown", func(t *testing.T) {
		for _, name := range []string{"todo-path", "broker-db"} {
			out, _, err := runPaths(t, "--field", name)
			if err == nil || out != "" {
				t.Errorf("--field %s: err = %v, stdout = %q, want a failure with empty stdout", name, err, out)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		out, _, err := runPaths(t, "--field", "")
		if err == nil || out != "" {
			t.Errorf("--field \"\": err = %v, stdout = %q, want a failure with empty stdout", err, out)
		}
	})
	t.Run("unresolved", func(t *testing.T) {
		unresolveFactory(t)
		out, _, err := runPaths(t, "--field", "factory-db")
		if err == nil || out != "" {
			t.Errorf("--field factory-db while unresolved: err = %v, stdout = %q, want a failure with empty stdout", err, out)
		}
	})
}

// unresolveFactory makes the home-derived resolvers fail, the way an
// unresolvable moai home does, through the command's resolver seam.
func unresolveFactory(t *testing.T) {
	t.Helper()
	origDir, origDB := pathsProjectDir, pathsFactoryDBPath
	fail := func(string) (string, error) { return "", errors.New("moai home cannot be resolved") }
	pathsProjectDir, pathsFactoryDBPath = fail, fail
	t.Cleanup(func() { pathsProjectDir, pathsFactoryDBPath = origDir, origDB })
}

// ---------------------------------------------------------------------------
// AC-DBP-005 — linked worktree converges on the primary checkout
// ---------------------------------------------------------------------------

func TestPaths_WorktreeConverges(t *testing.T) {
	primary := t.TempDir()
	t.Setenv(config.EnvHome, filepath.Join(t.TempDir(), "moai-home"))
	initGitRepo(t, primary)
	wt := addGitWorktree(t, primary)

	t.Setenv("CLAUDE_PROJECT_DIR", primary)
	fromPrimary, _, err := runPaths(t, "--json")
	if err != nil {
		t.Fatalf("from primary: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	fromWorktree, _, err := runPaths(t, "--json")
	if err != nil {
		t.Fatalf("from worktree: %v", err)
	}
	if fromPrimary != fromWorktree {
		t.Errorf("the report differs between the primary checkout and its worktree:\nprimary:  %s\nworktree: %s", fromPrimary, fromWorktree)
	}
	// Equal emptiness proves nothing: the shared report must carry the primary's key.
	if got, want := docString(t, decodePathsDoc(t, fromWorktree), "project_key"), homestate.ProjectKey(primary); got != want {
		t.Errorf("project_key from the worktree = %q, want the primary checkout's key %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// AC-DBP-006 — layouts and unresolved fields
// ---------------------------------------------------------------------------

func TestPaths_LayoutProjectLocal(t *testing.T) {
	root := pathsTempOriginFixture(t)
	doc := pathsDoc(t)
	if got := docString(t, doc, "queue_layout"); got != "project-local" {
		t.Errorf("queue_layout = %q, want project-local", got)
	}
	queueRoot := factory.ResolveTodoQueueRoot(root)
	if want := newTodoReadStore().EnginePath(); !sameFile(docString(t, doc, "todo.path"), want) {
		t.Errorf("todo.path = %q, the todo verbs use %q", docString(t, doc, "todo.path"), want)
	}
	if want, err := homestate.FactoryDBPath(queueRoot); err != nil || docString(t, doc, "factory.path") != want {
		t.Errorf("factory.path = %q, the factory resolver returns %q (err %v)", docString(t, doc, "factory.path"), want, err)
	}
}

func TestPaths_LayoutUnresolved(t *testing.T) {
	pathsTempOriginFixture(t)
	unresolveFactory(t)

	out, _, err := runPaths(t, "--json")
	if err == nil {
		t.Errorf("an unresolved path must make the command exit non-zero")
	}
	doc := decodePathsDoc(t, out)
	for _, key := range []string{"factory", "project_dir", "broker"} {
		if v, present := doc[key]; !present || v != nil {
			t.Errorf("%s = %v (present %v), want null", key, v, present)
		}
	}
	if got := docString(t, doc, "queue_layout"); got != "project-local" {
		t.Errorf("queue_layout = %q, want project-local", got)
	}
	if docString(t, doc, "todo.path") == "" {
		t.Errorf("the queue still resolves and must be reported")
	}
	if strings.Contains(out, ".moai/state/factory/factory.db") {
		t.Errorf("the output names the fallback path of FactoryRegistryPath:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// AC-DBP-007 — broker path
// ---------------------------------------------------------------------------

func brokerFixture(t *testing.T) {
	t.Helper()
	root, _ := pathsHomeFixture(t)
	factoryDir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"r1", "r2", "r3"} {
		if err := os.MkdirAll(filepath.Join(factoryDir, "messages", id), 0o700); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPaths_BrokerRun(t *testing.T) {
	brokerFixture(t)
	doc := pathsDoc(t, "--run", "r1")
	got := filepath.ToSlash(docString(t, doc, "broker.path"))
	if !strings.HasSuffix(got, "/messages/r1/broker.db") {
		t.Errorf("broker.path = %q, want a path ending /messages/r1/broker.db", got)
	}
	if docGet(t, doc, "broker.exists") != false {
		t.Errorf("broker.exists = %v, want false: no broker database was created", docGet(t, doc, "broker.exists"))
	}
	// A valid id whose directory does not exist: the path is reported, exit 0.
	doc = pathsDoc(t, "--run", "never-ran")
	if !strings.HasSuffix(filepath.ToSlash(docString(t, doc, "broker.path")), "/messages/never-ran/broker.db") {
		t.Errorf("broker.path for an unseen run = %v", docGet(t, doc, "broker.path"))
	}
}

// TestPaths_BrokerInvalidRunUnresolvedHome: an invalid --run id is rejected even
// when the factory path cannot be resolved (AC-DBP-007, sync-audit F2).
func TestPaths_BrokerInvalidRunUnresolvedHome(t *testing.T) {
	pathsTempOriginFixture(t)
	t.Setenv("HOME", "")
	unresolveFactory(t)

	out, _, err := runPaths(t, "--run", "../x", "--field", "todo-dir")
	if err == nil {
		t.Errorf("an invalid --run id must exit non-zero even with an unresolved home")
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("an invalid --run id must print no path, got %q", out)
	}
}

func TestPaths_BrokerInvalidRun(t *testing.T) {
	brokerFixture(t)
	for _, args := range [][]string{{"--run", "../x"}, {"--run", "../x", "--json"}} {
		out, _, err := runPaths(t, args...)
		if err == nil {
			t.Errorf("paths %v must exit non-zero", args)
		}
		if strings.Contains(out, "broker.db") || strings.Contains(out, "messages") {
			t.Errorf("paths %v reported a path for an invalid run id: %q", args, out)
		}
	}
}

func TestPaths_BrokerTemplateCount(t *testing.T) {
	brokerFixture(t)
	doc := pathsDoc(t)
	tmpl := filepath.ToSlash(docString(t, doc, "broker.path_template"))
	if !strings.Contains(tmpl, "/messages/") || !strings.HasSuffix(tmpl, "/broker.db") {
		t.Errorf("broker.path_template = %q, want .../messages/<run>/broker.db", tmpl)
	}
	if got := docGet(t, doc, "broker.existing_runs"); got != float64(3) {
		t.Errorf("broker.existing_runs = %v, want 3", got)
	}
}

// ---------------------------------------------------------------------------
// Edge cases and wiring
// ---------------------------------------------------------------------------

func TestPaths_NoGitMetadata(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv(config.EnvHome, "")
	doc := pathsDoc(t)
	if docString(t, doc, "project_key") == "" || docString(t, doc, "todo.path") == "" {
		t.Errorf("a project without git metadata must still report its paths: %v", doc)
	}
}

func TestPaths_RegisteredOnRoot(t *testing.T) {
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Name() == "paths" {
			found = true
			if c.GroupID != "tools" {
				t.Errorf("GroupID = %q, want tools", c.GroupID)
			}
		}
	}
	if !found {
		t.Errorf("`paths` is not registered on the root command")
	}
}

func TestPaths_TrivialFastPath(t *testing.T) {
	if !trivialCommands["paths"] {
		t.Errorf("paths must be on the deps-free fast path: it reads flags and resolvers only")
	}
}

func TestPaths_FlagsAreExclusive(t *testing.T) {
	pathsHomeFixture(t)
	out, _, err := runPaths(t, "--json", "--field", "todo-db")
	if err == nil || out != "" {
		t.Errorf("--json with --field: err = %v, stdout = %q, want a failure with empty stdout", err, out)
	}
}

// ---------------------------------------------------------------------------
// AC-DBP-008 — help texts point at `moai paths`
// ---------------------------------------------------------------------------

func TestPaths_HelpPointers(t *testing.T) {
	isolateGitGlobalSeams(t)
	var save string
	for _, c := range newHandoffCmd().Commands() {
		if c.Name() == "save" {
			save = c.Long
		}
	}
	for name, text := range map[string]string{
		"moai todo --help":         newTodoCmd().Long,
		"moai factory --help":      newFactoryCommand().Long,
		"moai handoff save --help": save,
	} {
		if !strings.Contains(text, "moai paths") {
			t.Errorf("%s does not point at `moai paths`:\n%s", name, text)
		}
	}
	// The pinned canonical-store string stays in the todo long help.
	// home-guard:allow the help text names the store; the string is asserted, never resolved
	if !strings.Contains(newTodoCmd().Long, "~/.moai/db/<project-key>/todo/backlog.db") {
		t.Errorf("the todo long help lost its canonical-store string")
	}
}
