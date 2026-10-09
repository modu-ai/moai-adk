// SPEC-UPDATE-MIGRATION-FIX-001 M2-b (REQ-UMF-006, AC-UMF-004): the installer
// must verify that every directory target from the common-asset catalog yields
// at least one file. A directory target that yields none is reported as a
// failure with a reason and is never counted as installed.
package userassets

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestInstaller_RejectsEmptyDirectoryTargets installs a catalog that carries one
// directory entry whose source tree holds no file. The empty entry must surface
// as a reported failure, and the installed count must equal the control install
// that does not carry the entry.
func TestInstaller_RejectsEmptyDirectoryTargets(t *testing.T) {
	control := newFixture(t)
	controlRes, err := control.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("control install: %v", err)
	}

	f := newFixture(t)
	f.cat.Catalog.Core.Skills = append(f.cat.Catalog.Core.Skills, template.Entry{
		Name:    "moai-empty",
		Tier:    template.TierCore,
		Path:    "templates/.claude/skills/moai-empty/",
		Version: "1.0.0",
	})
	// An explicit directory with no children: the entry exists, carries no file.
	f.src[".claude/skills/moai-empty"] = &fstest.MapFile{Mode: fs.ModeDir | 0o755}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("install with an empty directory target must report, not abort: %v", err)
	}

	var reported *FileOutcome
	for i := range res.Failures {
		if res.Failures[i].Path == "moai-empty/" {
			reported = &res.Failures[i]
		}
	}
	if reported == nil {
		t.Fatalf("REQ-UMF-006: the empty directory target was not reported as a failure; failures=%+v", res.Failures)
	}
	if !strings.Contains(reported.Reason, "empty") {
		t.Errorf("REQ-UMF-006: the failure reason must name the empty target, got %q", reported.Reason)
	}
	if res.Installed != controlRes.Installed {
		t.Errorf("REQ-UMF-006: the empty directory target was counted as installed: installed=%d, control=%d",
			res.Installed, controlRes.Installed)
	}
}
