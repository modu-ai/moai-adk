package cli

// init_resume_retrack_test.go — Item E (card t1591): the resume re-records the
// section hashes a healthy init records.
//
// The healthy tail calls retrackSectionFiles after every post-deploy section writer
// (init.go, card t1277), so the project manifest carries the sections' on-disk hashes.
// The resume rewrites the same section family — llm.yaml through ApplyHarness and
// ApplyDeployMode, user.yaml and language.yaml through the profile sync — and never
// re-recorded them, so the next init --force or update reads those files as user edits.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// readSectionBytes returns the bytes of one project-relative section file.
func readSectionBytes(t *testing.T, projectDir, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return data
}

// sectionManifestState reads the project manifest's entry for one section file and the
// file's current on-disk hash. A missing entry fails the test as a setup failure, not
// as the behavior under test.
func sectionManifestState(t *testing.T, projectDir, rel string) (manifest.FileEntry, string) {
	t.Helper()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectDir); err != nil {
		t.Fatalf("load the project manifest: %v", err)
	}
	entry, ok := mgr.GetEntry(rel)
	if !ok {
		t.Fatalf("precondition: the project manifest has no entry for %s", rel)
	}
	hash, err := manifest.HashFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("hash %s on disk: %v", rel, err)
	}
	return *entry, hash
}

// TestInitResumeRetracksRewrittenSections — Item E. The seed run deploys the sections
// and fails at the user-asset ensure, so the manifest records each section at its
// deploy-time hash. The resume then rewrites a section file, and afterwards the manifest
// entry must carry that file's on-disk hash, as a healthy init leaves it.
//
// Each case asserts its preconditions before the behavior: after the seed run the entry
// is template_managed and matches the file on disk, and the resume really rewrote the
// bytes. A failure on the behavior line therefore names the stale hash, not a setup error.
func TestInitResumeRetracksRewrittenSections(t *testing.T) {
	cases := []struct {
		name string
		// rel is the section file the resume rewrites.
		rel string
		// identity makes the resume's wizard answer a user name, so the profile sync
		// rewrites user.yaml. Without it the resume rewrites llm.yaml only.
		identity bool
	}{
		{name: "llm.yaml rewritten by the harness and deploy-mode steps", rel: ".moai/config/sections/llm.yaml"},
		{name: "user.yaml rewritten by the profile sync", rel: ".moai/config/sections/user.yaml", identity: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "retrack-resume-proj")
			resumeSeedRun(t, projectDir, "claude")

			seedEntry, seedHash := sectionManifestState(t, projectDir, tc.rel)
			if seedEntry.Provenance != manifest.TemplateManaged {
				t.Fatalf("precondition: the seed run tracks %s as %q, not template_managed", tc.rel, seedEntry.Provenance)
			}
			if seedEntry.CurrentHash != seedHash {
				t.Fatalf("precondition: after the seed run the manifest does not carry %s's on-disk hash (entry %s, disk %s)",
					tc.rel, seedEntry.CurrentHash, seedHash)
			}
			seedBytes := readSectionBytes(t, projectDir, tc.rel)

			if tc.identity {
				resumeWizardIdentity(t, "Retrack Operator", "")
			}
			if err := resumeRun(t, projectDir, ""); err != nil {
				t.Fatalf("run 2 (the resume) failed: %v", err)
			}
			if bytes.Equal(readSectionBytes(t, projectDir, tc.rel), seedBytes) {
				t.Fatalf("precondition: the resume did not rewrite %s, so there is no rewrite for the manifest to record", tc.rel)
			}

			entry, hash := sectionManifestState(t, projectDir, tc.rel)
			if entry.CurrentHash != hash {
				t.Errorf("the resume rewrote %s, but the manifest entry does not carry the on-disk hash: "+
					"CurrentHash %s, pre-write hash %s, on-disk after the resume %s. "+
					"The resume never re-recorded the section hashes (a healthy init does, via retrackSectionFiles)",
					tc.rel, entry.CurrentHash, seedHash, hash)
			}
		})
	}
}
