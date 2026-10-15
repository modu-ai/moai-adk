// SPEC-UPDATE-REPAIR-RC30-001 / M4
// AC-RC30-09 (U-08, REQ-RC30-011, REQ-RC30-017): the archive summary line is
// built from the measured archived count only. The former literal
// "0 user customizations modified" was never measured by the archive step.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestArchiveSummaryCarriesMeasuredCountOnly archives two legacy skills and
// checks the summary line reports the measured count and no hardcoded claim.
func TestArchiveSummaryCarriesMeasuredCountOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, id := range legacySkillIDs[:2] {
		makeSkillDir(t, root, id, "# "+id)
	}

	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out, false)
	if err != nil {
		t.Fatalf("archiveLegacySkills: %v", err)
	}
	if archived != 2 {
		t.Fatalf("archived = %d, want 2", archived)
	}

	output := out.String()
	if !strings.Contains(output, "total: 2 skills archived") {
		t.Errorf("summary line must carry the measured archived count, got:\n%s", output)
	}
	if strings.Contains(output, "customizations") {
		t.Errorf("summary line must not carry an unmeasured customization count, got:\n%s", output)
	}
}
