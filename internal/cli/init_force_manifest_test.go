package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestInitForceCarriesManifestProvenance pins what `moai init --force` owes
// the files a previous deployment left outside .moai/. The force path moves
// .moai/ aside, manifest included; the files it recorded must keep the
// provenance they had rather than all turning user_created:
//   - a template file nobody edited is redeployed and stays template_managed,
//     published-skill paths included (update never refreshes those once they
//     read user_created);
//   - a template file the user edited is left byte-identical and recorded
//     user_modified, so reinitializing never overwrites the edit;
//   - a file recorded user_created is left byte-identical and stays so.
func TestInitForceCarriesManifestProvenance(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	const (
		olderRule  = ".claude/rules/moai/core/moai-constitution.md"
		olderSkill = ".agents/skills/moai-gate/SKILL.md"
		editedRule = ".claude/rules/moai/workflow/mx-tag-protocol.md"
		ownedRule  = ".claude/rules/moai/languages/go.md"
		olderBody  = "older deploy\n"
		editedBody = "the user's own edit\n"
		ownedBody  = "a file the user owns\n"
	)
	root := filepath.Join(t.TempDir(), "proj")
	initProfileAt(t, root, "claude")

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	// An older deployment: content and manifest hashes agree, so the file
	// is still exactly what MoAI wrote.
	h := manifest.HashBytes([]byte(olderBody))
	for _, rel := range []string{olderRule, olderSkill} {
		writeTestFile(t, root, rel, olderBody)
		mgr.Manifest().Files[rel] = manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: h, DeployedHash: h, CurrentHash: h}
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	// A file the user owns at a template path, recorded user_created.
	writeTestFile(t, root, ownedRule, ownedBody)
	mgr.Manifest().Files[ownedRule] = manifest.FileEntry{Provenance: manifest.UserCreated, TemplateHash: h, DeployedHash: manifest.HashBytes([]byte(ownedBody)), CurrentHash: manifest.HashBytes([]byte(ownedBody))}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	// A user edit the manifest never saw: the file no longer matches the
	// hash recorded for it.
	writeTestFile(t, root, editedRule, editedBody)

	cmd := newInitTestCmd()
	for name, val := range map[string]string{"llm": "claude", "force": "true", "non-interactive": "true", "name": "transition", "language": "go", "mode": "tdd"} {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{root}); err != nil {
		t.Fatalf("init --force: %v (stderr: %s)", err, errBuf.String())
	}

	after := manifest.NewManager()
	if _, err := after.Load(root); err != nil {
		t.Fatal(err)
	}
	provenance := func(rel string) manifest.Provenance {
		if e, ok := after.GetEntry(rel); ok && e != nil {
			return e.Provenance
		}
		return ""
	}
	for _, rel := range []string{olderRule, olderSkill} {
		if got := string(readTestFile(t, root, rel)); got == olderBody {
			t.Errorf("%s was not redeployed by init --force", rel)
		}
		if got := provenance(rel); got != manifest.TemplateManaged {
			t.Errorf("%s provenance = %q, want %q", rel, got, manifest.TemplateManaged)
		}
	}
	if got := string(readTestFile(t, root, editedRule)); got != editedBody {
		t.Errorf("%s: the user's edit was overwritten: %q", editedRule, got)
	}
	if got := provenance(editedRule); got != manifest.UserModified {
		t.Errorf("%s provenance = %q, want %q", editedRule, got, manifest.UserModified)
	}
	if got := string(readTestFile(t, root, ownedRule)); got != ownedBody {
		t.Errorf("%s: the user's file was overwritten: %q", ownedRule, got)
	}
	if got := provenance(ownedRule); got != manifest.UserCreated {
		t.Errorf("%s provenance = %q, want %q", ownedRule, got, manifest.UserCreated)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "manifest.json")); err != nil {
		t.Errorf("manifest.json missing after init --force: %v", err)
	}
}
