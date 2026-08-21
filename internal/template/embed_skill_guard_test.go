package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// templateSourceRoot is the on-disk template tree that //go:embed compiles in.
const templateSourceRoot = "templates"

// TestEmbed_TemplateTreeHasNoSymlinks covers AC-CSC-001, both arms.
//
// A directory-pattern //go:embed drops symlinks — files and directories alike —
// with no build error and no warning. Nothing else catches this, which is why
// both arms exist: the first sees a link anywhere in the tree, the second sees
// the consequence (an entry present on disk but missing from the embed).
func TestEmbed_TemplateTreeHasNoSymlinks(t *testing.T) {
	// Arm 1 — no symlink anywhere in the tree. The scope is the whole tree,
	// not just .claude/skills/: a link under .claude/agents/ or .claude/rules/
	// disappears exactly the same way.
	var links []string
	err := filepath.WalkDir(templateSourceRoot, func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			links = append(links, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %q: %v", templateSourceRoot, err)
	}
	if len(links) != 0 {
		t.Errorf("template source tree contains %d symlink(s): %v — go:embed drops these silently, so the files never reach a user", len(links), links)
	}

	// Arm 2 — the on-disk skill set and the embedded skill set are identical.
	//
	// The collection below must NOT filter on entry.IsDir(). fs.DirEntry is
	// Lstat-based, so a symlink to a directory reports IsDir() == false: it
	// would drop out of the disk set exactly as it drops out of the embed, the
	// two sets would still match, and this test would pass in the one state it
	// exists to catch.
	diskDir := filepath.Join(templateSourceRoot, ".claude", "skills")
	diskEntries, err := os.ReadDir(diskDir)
	if err != nil {
		t.Fatalf("read %q: %v", diskDir, err)
	}
	var disk []string
	for _, e := range diskEntries {
		isLink := e.Type()&fs.ModeSymlink != 0
		if e.IsDir() || isLink {
			disk = append(disk, e.Name())
		}
	}
	sort.Strings(disk)

	embedded := embeddedSkillNames(t)
	if !sameStringSlice(disk, embedded) {
		t.Errorf("skill set on disk (%d) differs from the embedded set (%d)\n  disk:     %v\n  embedded: %v", len(disk), len(embedded), disk, embedded)
	}
}

// embeddedSkillNames lists the first-level entry names under .claude/skills/ in
// the embedded filesystem.
func embeddedSkillNames(t *testing.T) []string {
	t.Helper()
	sub, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	entries, err := fs.ReadDir(sub, ".claude/skills")
	if err != nil {
		t.Fatalf("read embedded .claude/skills: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// TestEmbed_SkillNamePrefixInvariant covers AC-CSC-016: every deployed skill
// directory name starts with "moai" — no hyphen.
//
// The names are collected from the directory listing directly, NOT via
// EmbeddedMoaiSkillNames(): that helper filters on "moai-" with a hyphen, and
// the catalog contains a skill named exactly "moai". Using the helper here
// would leave that one skill unexamined by the very guard meant to cover it.
//
// What the invariant protects is the mirror set this deployer produces: the
// clean path that will later retire a stale mirror is prefix-limited, so a
// mirror created under a name outside the prefix would be unreachable by it.
// The agreement is currently a coincidence — catalog.yaml has an empty
// harness-generated skills slot — and this test is what turns the coincidence
// into a decision a human has to make.
func TestEmbed_SkillNamePrefixInvariant(t *testing.T) {
	const requiredPrefix = "moai"

	names := embeddedSkillNames(t)
	if len(names) == 0 {
		t.Fatal("no skills found in the embedded filesystem — fixture assumption broken")
	}

	var offenders []string
	for _, name := range names {
		if !strings.HasPrefix(name, requiredPrefix) {
			offenders = append(offenders, name)
		}
	}
	if len(offenders) != 0 {
		t.Errorf("%d skill name(s) without the %q prefix: %v — a mirror under such a name cannot be reached by the prefix-limited clean path", len(offenders), requiredPrefix, offenders)
	}
}
