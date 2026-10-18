//go:build !windows

package cli

// migrate_ownership_boundary_table_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, chunk 11 adversarial self-review): the path-boundary
// counterexamples of the helpers the last audit rounds touched, frozen as ONE
// table. Each row is (group, case, expected, observed); the observation is a
// string so the whole table reads off the verbose log (`TABLE|...` lines) and is
// copied from that saved output into progress.md.
//
// Groups: restore (the bookkeeping-name filter of the legacy restore walk),
// boundary (withinRootBoundary and pathLeavesRoot: element-wise containment, never
// a raw string prefix, in both the absolute-root shape the tests use and the
// relative "." root the update command really passes), symlink (what the verdict
// function reads through links), kind (FIFOs and directories, with a bounded wait
// that only detects a hang), and emptyroot.
//
// A FIFO may block a read until a writer appears; every row that plants one runs
// the verdict in a goroutine and pokes the FIFO to end a hung read, so a hang is a
// failure, not a stuck suite. A failed os.Symlink / mkfifo is a failure, not a skip.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

type tableRow struct {
	group, name, expected string
	observe               func(t *testing.T) string
}

func tbWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tbLink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("os.Symlink(%s -> %s) failed: %v", link, target, err)
	}
}

// tbVerdict renders a verdict as the table's observed column. "read=content" means
// the verdict carries exactly want's bytes, "read=none" that it carries none.
func tbVerdict(v ownershipVerdict, want string) string {
	read := "none"
	switch {
	case len(v.data) == 0:
	case string(v.data) == want:
		read = "content"
	default:
		read = "other-bytes"
	}
	return fmt.Sprintf("class=%s keep=%s read=%s", v.class, v.keep, read)
}

func tbEnv(rootResolved string, registered map[string]*manifest.FileEntry) ownershipEnv {
	return ownershipEnv{
		entry:                func(rel string) (*manifest.FileEntry, bool) { e, ok := registered[rel]; return e, ok },
		evidence:             fakeProvenance{},
		counterpartConfirmed: func(string) bool { return true },
		notInCatalog:         func(string) bool { return false },
		removalEnabled:       true,
		rootResolved:         rootResolved,
	}
}

// tbClassify runs the verdict in a goroutine with a bounded wait that only detects
// a hang; on a hang the FIFOs are poked until the read ends.
func tbClassify(t *testing.T, env ownershipEnv, rel, abs string, fifos ...string) ownershipVerdict {
	t.Helper()
	done := make(chan ownershipVerdict, 1)
	go func() { done <- classifyOwnership(env, rel, abs) }()
	select {
	case v := <-done:
		return v
	case <-time.After(30 * time.Second):
		releaseBlockedFifoRead(done, fifos...)
		t.Fatalf("the verdict for %s did not return: it is blocked reading a FIFO", rel)
		return ownershipVerdict{}
	}
}

func tbRestoreRow(rel string, restored bool) tableRow {
	exp := "skipped"
	if restored {
		exp = "restored"
	}
	return tableRow{"restore", rel, exp, func(t *testing.T) string {
		projectRoot, backupDir, configDir := t.TempDir(), t.TempDir(), ""
		configDir = filepath.Join(projectRoot, ".moai", "config")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		tbWrite(t, filepath.Join(backupDir, "system.yaml"), "version: 1\n") // reach control
		tbWrite(t, filepath.Join(backupDir, filepath.FromSlash(rel)), "content of "+rel+"\n")
		if err := backup.RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir); err != nil {
			t.Fatalf("RestoreMoaiConfigLegacy: %v", err)
		}
		if _, err := os.Stat(filepath.Join(configDir, "system.yaml")); err != nil {
			t.Fatalf("control: the restore did not reach its walk: %v", err)
		}
		if _, err := os.Stat(filepath.Join(configDir, filepath.FromSlash(rel))); err != nil {
			return "skipped"
		}
		return "restored"
	}}
}

func boundaryRows() []tableRow {
	inside := func(ok bool) string {
		if ok {
			return "inside"
		}
		return "outside"
	}
	// Pure containment, real directory names that share a string prefix with the root.
	pureRow := func(name, expected string, rootRel, pathRel string) tableRow {
		return tableRow{"boundary", name, expected, func(t *testing.T) string {
			base := t.TempDir()
			return inside(withinRootBoundary(filepath.Join(base, rootRel), filepath.Join(base, pathRel)))
		}}
	}
	// pathLeavesRoot over real files.
	leavesRow := func(name, expected string, build func(t *testing.T, base string) (root, abs string)) tableRow {
		return tableRow{"boundary", name, expected, func(t *testing.T) string {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			root, abs := build(t, base)
			rootResolved, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			return inside(!pathLeavesRoot(rootResolved, abs))
		}}
	}
	return []tableRow{
		pureRow("withinRootBoundary: sibling named root+'-evil' is outside", "outside", "proj-xxx", "proj-xxx-evil/file"),
		pureRow("withinRootBoundary: sibling named root+'2' is outside", "outside", "proj-xxx", "proj-xxx2/file"),
		pureRow("withinRootBoundary: a file under the root is inside", "inside", "proj-xxx", "proj-xxx/sub/file"),
		pureRow("withinRootBoundary: the root itself is inside", "inside", "proj-xxx", "proj-xxx"),
		pureRow("withinRootBoundary: parent of the root is outside", "outside", "proj-xxx", "."),
		pureRow("withinRootBoundary: '..' escape through the root is outside", "outside", "proj-xxx", "proj-xxx/../other/file"),
		pureRow("withinRootBoundary: top-level entry whose name starts with two dots is inside", "inside", "proj-xxx", "proj-xxx/..hidden/file"),
		leavesRow("pathLeavesRoot: real file in the sibling 'proj-xxx-evil'", "outside", func(t *testing.T, base string) (string, string) {
			tbWrite(t, filepath.Join(base, "proj-xxx", "keep.md"), "in\n")
			tbWrite(t, filepath.Join(base, "proj-xxx-evil", "file.md"), "out\n")
			return filepath.Join(base, "proj-xxx"), filepath.Join(base, "proj-xxx-evil", "file.md")
		}),
		leavesRow("pathLeavesRoot: in-project link to the sibling 'proj-xxx2'", "outside", func(t *testing.T, base string) (string, string) {
			tbWrite(t, filepath.Join(base, "proj-xxx", "keep.md"), "in\n")
			tbWrite(t, filepath.Join(base, "proj-xxx2", "file.md"), "out\n")
			tbLink(t, filepath.Join(base, "proj-xxx2", "file.md"), filepath.Join(base, "proj-xxx", "link.md"))
			return filepath.Join(base, "proj-xxx"), filepath.Join(base, "proj-xxx", "link.md")
		}),
		leavesRow("pathLeavesRoot: regular file in the project", "inside", func(t *testing.T, base string) (string, string) {
			tbWrite(t, filepath.Join(base, "proj-xxx", "a", "f.md"), "in\n")
			return filepath.Join(base, "proj-xxx"), filepath.Join(base, "proj-xxx", "a", "f.md")
		}),
		leavesRow("pathLeavesRoot: file in a top-level directory named '..hidden'", "inside", func(t *testing.T, base string) (string, string) {
			tbWrite(t, filepath.Join(base, "proj-xxx", "..hidden", "f.md"), "in\n")
			return filepath.Join(base, "proj-xxx"), filepath.Join(base, "proj-xxx", "..hidden", "f.md")
		}),
		// The relative root the update command passes ("."): the same answers.
		{"boundary", "root '.' (the update command's shape): relative in-project file", "inside", func(t *testing.T) string {
			root := t.TempDir()
			tbWrite(t, filepath.Join(root, ".claude", "skills", "x", "SKILL.md"), "in\n")
			t.Chdir(root)
			return inside(!pathLeavesRoot(".", filepath.Join(".claude", "skills", "x", "SKILL.md")))
		}},
		{"boundary", "root '.': in-project symlink whose target is an ABSOLUTE in-project path", "inside", func(t *testing.T) string {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			tbWrite(t, filepath.Join(root, "shared", "real.md"), "in\n")
			tbLink(t, filepath.Join(root, "shared", "real.md"), filepath.Join(root, ".claude", "skills", "x", "SKILL.md"))
			t.Chdir(root)
			return inside(!pathLeavesRoot(".", filepath.Join(".claude", "skills", "x", "SKILL.md")))
		}},
		{"boundary", "root '.' reached through a symlinked working directory: in-project absolute link target", "inside", func(t *testing.T) string {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			real := filepath.Join(base, "real-proj")
			tbWrite(t, filepath.Join(real, "shared", "real.md"), "in\n")
			tbLink(t, filepath.Join(real, "shared", "real.md"), filepath.Join(real, ".claude", "skills", "x", "SKILL.md"))
			tbLink(t, real, filepath.Join(base, "link-proj"))
			t.Chdir(filepath.Join(base, "link-proj"))
			return inside(!pathLeavesRoot(".", filepath.Join(".claude", "skills", "x", "SKILL.md")))
		}},
		{"boundary", "root '.': symlink to an outside absolute path", "outside", func(t *testing.T) string {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(base, "proj")
			tbWrite(t, filepath.Join(base, "proj-evil", "f.md"), "out\n")
			tbLink(t, filepath.Join(base, "proj-evil", "f.md"), filepath.Join(root, ".claude", "skills", "x", "SKILL.md"))
			t.Chdir(root)
			return inside(!pathLeavesRoot(".", filepath.Join(".claude", "skills", "x", "SKILL.md")))
		}},
		{"boundary", "root '.': symlink to an outside relative path", "outside", func(t *testing.T) string {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(base, "proj")
			tbWrite(t, filepath.Join(base, "proj-evil", "f.md"), "out\n")
			tbLink(t, filepath.Join("..", "..", "..", "..", "proj-evil", "f.md"), filepath.Join(root, ".claude", "skills", "x", "SKILL.md"))
			t.Chdir(root)
			return inside(!pathLeavesRoot(".", filepath.Join(".claude", "skills", "x", "SKILL.md")))
		}},
	}
}

func symlinkRows() []tableRow {
	const inContent, outContent = "# inside the project\n", "# outside the project\n"
	row := func(name, expected string, build func(t *testing.T, root, out string) string) tableRow {
		return tableRow{"symlink", name, expected, func(t *testing.T) string {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			out, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			rel := build(t, root, out) // returns the project-relative path to classify
			return tbVerdict(tbClassify(t, tbEnv(root, nil), rel, filepath.Join(root, filepath.FromSlash(rel))), inContent)
		}}
	}
	const p = ".agents/skills/moai-table/"
	return []tableRow{
		row("symlink to an in-project file: judged by its target's bytes", "class=unregistered keep=no-provenance-record read=content", func(t *testing.T, root, _ string) string {
			tbWrite(t, filepath.Join(root, "real.md"), inContent)
			tbLink(t, filepath.Join(root, "real.md"), filepath.Join(root, p+"link.md"))
			return p + "link.md"
		}),
		row("symlink to an outside file: kept unread", "class=unregistered keep=symlink-escape read=none", func(t *testing.T, root, out string) string {
			tbWrite(t, filepath.Join(out, "f.md"), outContent)
			tbLink(t, filepath.Join(out, "f.md"), filepath.Join(root, p+"link.md"))
			return p + "link.md"
		}),
		row("symlinked ancestor directory outside: kept unread", "class=unregistered keep=symlink-escape read=none", func(t *testing.T, root, out string) string {
			tbWrite(t, filepath.Join(out, "d", "f.md"), outContent)
			tbLink(t, filepath.Join(out, "d"), filepath.Join(root, ".agents", "skills", "dirlink"))
			return ".agents/skills/dirlink/f.md"
		}),
		row("symlinked ancestor directory inside: judged by content", "class=unregistered keep=no-provenance-record read=content", func(t *testing.T, root, _ string) string {
			tbWrite(t, filepath.Join(root, "real", "f.md"), inContent)
			tbLink(t, filepath.Join(root, "real"), filepath.Join(root, ".agents", "skills", "dirlink"))
			return ".agents/skills/dirlink/f.md"
		}),
		row("chain of links ending outside: kept unread", "class=unregistered keep=symlink-escape read=none", func(t *testing.T, root, out string) string {
			tbWrite(t, filepath.Join(out, "f.md"), outContent)
			tbLink(t, filepath.Join(out, "f.md"), filepath.Join(root, "hop2"))
			tbLink(t, filepath.Join(root, "hop2"), filepath.Join(root, "hop1"))
			tbLink(t, filepath.Join(root, "hop1"), filepath.Join(root, p+"chain.md"))
			return p + "chain.md"
		}),
		row("chain of links staying inside: judged by content", "class=unregistered keep=no-provenance-record read=content", func(t *testing.T, root, _ string) string {
			tbWrite(t, filepath.Join(root, "real.md"), inContent)
			tbLink(t, filepath.Join(root, "real.md"), filepath.Join(root, "hop1"))
			tbLink(t, filepath.Join(root, "hop1"), filepath.Join(root, p+"chain.md"))
			return p + "chain.md"
		}),
		row("dangling link: read-failed, nothing read", "class=unregistered keep=read-failed read=none", func(t *testing.T, root, _ string) string {
			tbLink(t, filepath.Join(root, "nowhere"), filepath.Join(root, p+"dead.md"))
			return p + "dead.md"
		}),
		row("link loop (ELOOP): read-failed, nothing read", "class=unregistered keep=read-failed read=none", func(t *testing.T, root, _ string) string {
			tbLink(t, filepath.Join(root, "loop-b"), filepath.Join(root, "loop-a"))
			tbLink(t, filepath.Join(root, "loop-a"), filepath.Join(root, "loop-b"))
			tbLink(t, filepath.Join(root, "loop-a"), filepath.Join(root, p+"loop.md"))
			return p + "loop.md"
		}),
		row("a regular file used as a directory component (ENOTDIR): read-failed", "class=unregistered keep=read-failed read=none", func(t *testing.T, root, _ string) string {
			tbWrite(t, filepath.Join(root, p+"afile"), inContent)
			return p + "afile/below.md"
		}),
		row("file that does not exist: read-failed", "class=unregistered keep=read-failed read=none", func(t *testing.T, root, _ string) string {
			if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
				t.Fatal(err)
			}
			return p + "absent.md"
		}),
		row("ordinary file: judged by content", "class=unregistered keep=no-provenance-record read=content", func(t *testing.T, root, _ string) string {
			tbWrite(t, filepath.Join(root, p+"plain.md"), inContent)
			return p + "plain.md"
		}),
	}
}

// flowRows drives the real walk the way the update command does: from the project
// directory with the relative root ".".
func flowRows() []tableRow {
	const link = ".agents/skills/moai-table/abs-link.md"
	return []tableRow{
		{"flow", "update-shaped run (root '.'): in-project symlink with an absolute target", "class=unregistered reason=no-provenance-record removed=1", func(t *testing.T) string {
			root, home := ownershipFixtureBase(t)
			rootResolved, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			tbWrite(t, filepath.Join(rootResolved, "shared", "real.md"), "# in the project\n")
			tbLink(t, filepath.Join(rootResolved, "shared", "real.md"), filepath.Join(rootResolved, filepath.FromSlash(link)))
			t.Chdir(rootResolved)
			rc := &migrationRunContext{userAssetsInstalled: true}
			removed, err := migrateProjectCommonAssets(".", home, true, nil, func(string, ...interface{}) {}, rc)
			if err != nil {
				t.Fatalf("migration: %v", err)
			}
			r, ok := findRemaining(rc, link)
			if !ok {
				return "no remaining row"
			}
			return fmt.Sprintf("class=%s reason=%s removed=%d", r.Class, r.Reason, len(removed))
		}},
	}
}

func kindRows() []tableRow {
	const content = "# a registered file\n"
	reg := func(rel string) map[string]*manifest.FileEntry {
		return map[string]*manifest.FileEntry{rel: {Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes([]byte(content))}}
	}
	const rel = ".agents/skills/moai-table/node"
	rootOf := func(t *testing.T) (string, string) {
		root, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		return root, filepath.Join(root, filepath.FromSlash(rel))
	}
	return []tableRow{
		{"kind", "FIFO at an unregistered project path", "class=unregistered keep=non-regular-file read=none", func(t *testing.T) string {
			root, abs := rootOf(t)
			mkfifo(t, abs)
			return tbVerdict(tbClassify(t, tbEnv(root, nil), rel, abs, abs), content)
		}},
		{"kind", "FIFO at a registered path", "class=user-modified keep=non-regular-file read=none", func(t *testing.T) string {
			root, abs := rootOf(t)
			mkfifo(t, abs)
			return tbVerdict(tbClassify(t, tbEnv(root, reg(rel)), rel, abs, abs), content)
		}},
		{"kind", "link to a FIFO inside the project", "class=unregistered keep=non-regular-file read=none", func(t *testing.T) string {
			root, abs := rootOf(t)
			fifo := filepath.Join(root, "pipe")
			mkfifo(t, fifo)
			tbLink(t, fifo, abs)
			return tbVerdict(tbClassify(t, tbEnv(root, nil), rel, abs, fifo), content)
		}},
		{"kind", "directory where an unregistered file was expected", "class=unregistered keep=read-failed read=none", func(t *testing.T) string {
			root, abs := rootOf(t)
			if err := os.MkdirAll(abs, 0o755); err != nil {
				t.Fatal(err)
			}
			return tbVerdict(tbClassify(t, tbEnv(root, nil), rel, abs), content)
		}},
		{"kind", "directory where a registered file was expected", "class=user-modified keep=read-failed read=none", func(t *testing.T) string {
			root, abs := rootOf(t)
			if err := os.MkdirAll(abs, 0o755); err != nil {
				t.Fatal(err)
			}
			return tbVerdict(tbClassify(t, tbEnv(root, reg(rel)), rel, abs), content)
		}},
		{"kind", "FIFO as the user-root counterpart of a registered managed file", "keep=counterpart-unconfirmed counterpart=unconfirmed removable=false", func(t *testing.T) string {
			root, home := ownershipFixtureBase(t)
			fifo := counterpartFifo(t, home, migIdenticalSkill)
			o := runNonRegularMigration(t, nonRegularWorld{root: root, home: home, fifos: []string{fifo}})
			if o.hung {
				t.Fatalf("the migration did not return: blocked on the counterpart FIFO")
			}
			r, ok := findRemaining(o.rc, migIdenticalSkill)
			if !ok {
				return "no remaining row"
			}
			return fmt.Sprintf("keep=%s counterpart=%s removable=%v", r.Reason, r.Counterpart, len(o.removed) != 0)
		}},
	}
}

func emptyRootRows() []tableRow {
	return []tableRow{
		{"emptyroot", "unregistered ordinary file, rootResolved empty: fail closed", "class=unregistered keep=symlink-escape read=none", func(t *testing.T) string {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			abs := filepath.Join(root, ".agents", "skills", "x", "f.md")
			tbWrite(t, abs, "# plain\n")
			return tbVerdict(tbClassify(t, tbEnv("", nil), ".agents/skills/x/f.md", abs), "# plain\n")
		}},
		{"emptyroot", "registered ordinary file, rootResolved empty: never removable", "class=managed keep=symlink-escape read=content", func(t *testing.T) string {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			abs := filepath.Join(root, ".agents", "skills", "x", "f.md")
			tbWrite(t, abs, "# plain\n")
			reg := map[string]*manifest.FileEntry{".agents/skills/x/f.md": {Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes([]byte("# plain\n"))}}
			return tbVerdict(tbClassify(t, tbEnv("", reg), ".agents/skills/x/f.md", abs), "# plain\n")
		}},
	}
}

func TestPathBoundaryCounterexamples(t *testing.T) {
	var rows []tableRow
	for _, c := range []struct {
		rel      string
		restored bool
	}{
		{"manifest.prior.json.tmp-notes/keep.txt", true},
		{"manifest.prior.json.d/x", true},
		{"manifest.prior.json.tmp-notes/manifest.prior.json", true},
		{"manifest.prior.json", false},
		{"manifest.prior.json.tmp-4242", false},
		{"notes/manifest.prior.json", true},
		{"notes/manifest.prior.json.tmp-4242", true},
	} {
		rows = append(rows, tbRestoreRow(c.rel, c.restored))
	}
	rows = append(rows, boundaryRows()...)
	rows = append(rows, symlinkRows()...)
	rows = append(rows, flowRows()...)
	rows = append(rows, kindRows()...)
	rows = append(rows, emptyRootRows()...)
	for _, r := range rows {
		t.Run(r.group+"/"+strings.ReplaceAll(r.name, "/", "|"), func(t *testing.T) {
			got := r.observe(t)
			t.Logf("TABLE|%s|%s|expected: %s|observed: %s", r.group, r.name, r.expected, got)
			if got != r.expected {
				t.Errorf("%s: observed %q, want %q", r.name, got, r.expected)
			}
		})
	}
}
