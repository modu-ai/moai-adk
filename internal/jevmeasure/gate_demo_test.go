package jevmeasure

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// gate_demo_test.go — the gate demonstrated end to end for one trial consumer,
// so `SPEC-JEV-CONSUMERS-001` inherits a gate that has been run rather than a
// gate that has been described (SPEC-JEV-OPTIN-MEASURE-001 M3c; AC-JEVO-010,
// AC-JEVO-012).
//
// [HARD] What this demonstrates is the APPARATUS, not a result. The answerer
// is injected, so every figure below is a fixture and the verdict is
// VerdictWithhold by construction. No accuracy figure produced here may be
// cited as a measurement of anything.

// TestGateDemonstration_ArtifactSatisfiesAC010 runs the whole gate and checks
// the artifact carries everything AC-JEVO-010 requires of it: the pinned model
// id, both language arms with per-arm results and the delta, and the
// constant-answer baseline for that question shape.
func TestGateDemonstration_ArtifactSatisfiesAC010(t *testing.T) {
	set := trialSet()

	// A fixture that beats the constant baseline on the English arm, so the
	// artifact exercises a non-zero delta and a non-degenerate best-arm pick.
	answers := map[string]bool{}
	for i, s := range set.Samples {
		correct := s.Label == "dead"
		if i%7 == 0 {
			correct = !correct // a few wrong answers, so accuracy is not 1.0
		}
		answers[s.ID] = correct
	}
	probs := map[string]float64{}
	for i, s := range set.Samples {
		probs[s.ID] = 0.35 + float64(i)*0.03
	}

	rep, err := Run(context.Background(), stubAnswerer{answers: answers, probs: probs}, set, trialBuild)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	th, err := FitThreshold(set.Consumer, set.Shape, rep.BestArm())
	if err != nil {
		t.Fatalf("FitThreshold: %v", err)
	}
	rep.Threshold = &th

	if err := rep.Validate(); err != nil {
		t.Fatalf("the demonstrated artifact does not validate: %v", err)
	}

	// AC-JEVO-010, item by item.
	if rep.ModelID != jev.ModelID {
		t.Errorf("model id = %q, want the pinned %q", rep.ModelID, jev.ModelID)
	}
	if len(rep.Arms) != 2 {
		t.Fatalf("arms = %d, want both language arms", len(rep.Arms))
	}
	for _, a := range rep.Arms {
		if a.N != len(set.Samples) {
			t.Errorf("arm %s measured %d of %d samples", a.Arm, a.N, len(set.Samples))
		}
	}
	if rep.Delta != rep.Arms[1].Accuracy-rep.Arms[0].Accuracy {
		t.Errorf("delta %v is not the English-minus-Korean difference", rep.Delta)
	}
	if rep.Baseline.N != len(set.Samples) || rep.Baseline.Label == "" {
		t.Errorf("the artifact carries no usable constant-answer baseline: %+v", rep.Baseline)
	}
	if rep.Threshold.SampleSize != len(set.Samples) || rep.Threshold.Provenance == "" {
		t.Errorf("the fitted threshold carries no sample size or provenance: %+v", *rep.Threshold)
	}

	// The gate withholds, because the source is injected. This is the
	// demonstration's most important assertion: the apparatus runs end to end
	// AND still refuses to turn a fixture into a shipping decision.
	v, reason := rep.Verdict()
	if v != VerdictWithhold {
		t.Fatalf("the demonstration produced a shipping verdict from an injected answerer: %q (%s)", v, reason)
	}

	out := rep.Render()
	t.Logf("demonstrated artifact (fixture figures, NOT a measurement):\n%s", out)
	if !strings.Contains(out, string(SourceInjected)) {
		t.Error("the rendered artifact does not disclose that its source was injected")
	}
}

// TestNoConsumerCallPathShips is AC-JEVO-012 as this SPEC can hold it: no
// consumer exists yet, so no consumer call path is in the shipped build. The
// gate is built here; the consumers that must pass it belong to the successor
// SPEC.
//
// The assertion is narrow and honest. It does NOT claim a consumer was
// measured and withheld — it claims no consumer is reachable, which is the
// state this SPEC is supposed to leave behind.
func TestNoConsumerCallPathShips(t *testing.T) {
	root := repoRoot(t)

	// The named consumers the successor SPEC owns.
	consumerMarkers := []string{
		"NearDuplicateMark",
		"LaneQuestionRoute",
		"SkillSuggest",
	}
	var found []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path) //nolint:gosec // path comes from the walk
		if err != nil {
			return err
		}
		for _, m := range consumerMarkers {
			if strings.Contains(string(b), m) {
				found = append(found, m+" in "+path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(found) > 0 {
		t.Errorf("a consumer call path is present before its measurement: %v", found)
	}

	// Positive control: the walk actually read this repository's source. A
	// walk that read nothing would report every marker absent.
	sentinel := false
	_ = filepath.WalkDir(filepath.Join(root, "internal", "jevmeasure"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // the control only needs one hit
		}
		if strings.HasSuffix(path, "measure.go") {
			b, readErr := os.ReadFile(path) //nolint:gosec // fixed path under the repo
			if readErr == nil && strings.Contains(string(b), "ConstantBaseline") {
				sentinel = true
			}
		}
		return nil
	})
	if !sentinel {
		t.Error("positive control failed: the walk did not find this package's own source, so the absence above measured nothing")
	}
}

// repoRoot walks up from the package directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("module root not found")
	return ""
}
