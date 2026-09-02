package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-CMA-005 (CLI surface, SPEC-CODEMAPS-ACCURACY-001): the citations row
// rides the existing layer consumption path — a positive phantom in a codemaps
// doc makes `moai graph check` exit 1 naming the citations layer, and the same
// tree with the phantom removed (blockquote negative-citation form) exits 0 on
// that row. The exit-code contract (0/1/2) is unchanged: the row participates
// in Failed()/OffendingLayers() exactly as the freshness rows do.
func TestGraphCheckCmd_CitationsRowStaleExitsOne(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)

	docPath := filepath.Join(root, ".moai", "project", "codemaps", "modules.md")
	if err := os.WriteFile(docPath,
		[]byte("# modules\n\nCross-check internal/zzz-phantom for the legacy flow.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("positive phantom must exit non-zero via the citations row")
	}
	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 1 {
		t.Fatalf("exit code must be 1 (stale layer), got %v", err)
	}
	if !strings.Contains(errOut.String(), "citations") {
		t.Errorf("stderr must name the citations layer:\n%s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "internal/zzz-phantom") {
		t.Errorf("stderr must name the phantom path (driving-path contract):\n%s", errOut.String())
	}

	// Round-trip GREEN side: the same absence expressed as a blockquote
	// negative citation is exempt — the citations row goes fresh. (The
	// codemaps freshness row is red here because the doc moved past the
	// stamp, so the exit stays 1; the assertion is on the citations row's
	// verdict in the JSON report, the same read AC-CMA-005 prescribes.)
	if err := os.WriteFile(docPath,
		[]byte("# modules\n\n> **[REMOVED]** internal/zzz-phantom 는 제거된 패키지이다.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd2 := newGraphCmd()
	cmd2.SetArgs([]string{"check", "--root", root, "--json"})
	var out2, errOut2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(&errOut2)
	_ = cmd2.Execute()
	body := out2.String()
	idx := strings.Index(body, `"layer": "citations"`)
	if idx < 0 {
		t.Fatalf("JSON report must carry the citations layer:\n%s", body)
	}
	if !strings.Contains(body[idx:], `"verdict": "fresh"`) {
		t.Errorf("blockquote-form absence must leave the citations row fresh:\n%s", body[idx:])
	}
}
