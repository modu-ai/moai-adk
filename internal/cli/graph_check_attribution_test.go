package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ageCodemapsPastThreshold writes n described-worthy .go files under
// internal/aged, plus noise the predicate must exclude from both the count and
// the listing.
func ageCodemapsPastThreshold(t *testing.T, root string, n int) {
	t.Helper()
	dir := filepath.Join(root, "internal", "aged")
	if err := os.MkdirAll(filepath.Join(dir, "testdata"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := filepath.Join(dir, fmt.Sprintf("f%03d.go", i))
		if err := os.WriteFile(p, []byte("package aged\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "f000_test.go"), []byte("package aged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "testdata", "x.go"), []byte("package testdata\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-GFC-009 (REQ-GFC-008) — a stale codemaps verdict names the described-worthy
// paths driving the count on stderr, truncated to the display bound with an
// explicit overflow indicator, and names the change's own contribution so a
// reader can tell an inherited red from an originated one.
//
// Mutant this kills: a report that computes the attribution but renders only
// the count — the delivered struct carries the fields while the operator, who
// reads stderr and not JSON, learns nothing. Every assertion below fails.
func TestGraphCheckCmd_StaleStderrNamesDrivingPaths(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)
	const aged = 41
	ageCodemapsPastThreshold(t, root, aged)

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err == nil {
		t.Fatal("stale check must return an error (exit 1 path)")
	}
	stderr := errOut.String()

	if !strings.Contains(stderr, "internal/aged/f000.go") {
		t.Fatalf("stderr names no driving path — the count alone was reported:\n%s", stderr)
	}
	listed := strings.Count(stderr, "internal/aged/f")
	if listed != 10 {
		t.Fatalf("stderr lists %d driving paths, want 10 (the display bound):\n%s", listed, stderr)
	}
	if !strings.Contains(stderr, fmt.Sprintf("%d more", aged-10)) {
		t.Fatalf("stderr carries no explicit overflow indicator for the %d omitted paths:\n%s", aged-10, stderr)
	}
	for _, forbidden := range []string{"f000_test.go", "testdata/x.go"} {
		if strings.Contains(stderr, forbidden) {
			t.Fatalf("stderr lists %q, which is not described-worthy:\n%s", forbidden, stderr)
		}
	}
	// This fixture's HEAD is a root commit, so the contribution is absent —
	// and the absence must be stated, not silently rendered as 0.
	if !strings.Contains(stderr, "no first parent") {
		t.Fatalf("stderr does not state why the contribution is unmeasured:\n%s", stderr)
	}
	if strings.Contains(stderr, "contributed 0") {
		t.Fatalf("stderr reports a fabricated contribution of 0 on a checkout with no first parent:\n%s", stderr)
	}
}

// AC-GFC-010 (REQ-GFC-010) — the --json codemaps row carries the per-change
// contribution and the driving-path list as distinct machine-readable fields,
// while every pre-existing field keeps its name and meaning.
func TestGraphCheckCmd_JSONCarriesAttribution(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)
	const aged = 41
	ageCodemapsPastThreshold(t, root, aged)

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"check", "--root", root, "--json"})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err == nil {
		t.Fatal("stale check must return an error (exit 1 path)")
	}

	var doc struct {
		Layers []map[string]any `json:"layers"`
	}
	// Decode only the FIRST JSON value: cobra appends its usage block to the
	// same writer once RunE returns an error, so a whole-buffer Unmarshal
	// fails on trailing text that is not part of the report.
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&doc); err != nil {
		t.Fatalf("decode --json output: %v\n%s", err, out.String())
	}
	var row map[string]any
	for _, l := range doc.Layers {
		if l["layer"] == "codemaps" {
			row = l
		}
	}
	if row == nil {
		t.Fatalf("no codemaps layer in the JSON report:\n%s", out.String())
	}

	// Pre-existing fields, unchanged in name and meaning.
	for _, k := range []string{"layer", "metric", "value", "threshold", "verdict"} {
		if _, ok := row[k]; !ok {
			t.Fatalf("pre-existing field %q vanished from the codemaps row: %v", k, row)
		}
	}
	if got := row["value"].(float64); int(got) != aged {
		t.Fatalf("value = %v, want %d — the predicate must exclude the _test.go and testdata noise", got, aged)
	}
	if row["metric"] != "described-source-diff" || row["verdict"] != "stale" {
		t.Fatalf("pre-existing field meaning changed: metric=%v verdict=%v", row["metric"], row["verdict"])
	}

	// New attribution fields.
	paths, ok := row["driving_paths"].([]any)
	if !ok {
		t.Fatalf("codemaps row carries no driving_paths field: %v", row)
	}
	if len(paths) != 10 {
		t.Fatalf("driving_paths carries %d entries, want 10 (the display bound)", len(paths))
	}
	omitted, ok := row["driving_paths_omitted"].(float64)
	if !ok || int(omitted) != aged-10 {
		t.Fatalf("driving_paths_omitted = %v, want %d", row["driving_paths_omitted"], aged-10)
	}
	// The fixture HEAD is a root commit: contribution must be ABSENT, and the
	// absence must be legible rather than a bare missing key.
	if _, present := row["contribution"]; present {
		t.Fatalf("contribution is present on a checkout with no first parent — a fabricated 0: %v", row["contribution"])
	}
	if reason, _ := row["contribution_absent_reason"].(string); reason == "" {
		t.Fatalf("contribution is absent but carries no contribution_absent_reason: %v", row)
	}
}
