package security

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestLayer1Scan asserts the Layer-1 buffer scan produces a finding on a
// dangerous fixture and stays silent on a clean fixture (AC-SG-001, REQ-SG-010).
func TestLayer1Scan(t *testing.T) {
	t.Parallel()

	dangerous := "import yaml\ndata = yaml.load(untrusted_input)\n"
	findings := ScanBuffer(dangerous)
	if len(findings) == 0 {
		t.Fatalf("expected at least one finding on dangerous fixture, got 0")
	}
	found := false
	for _, f := range findings {
		if f.Class == "unsafe-deserialization" {
			found = true
			if f.Line != 2 {
				t.Errorf("expected finding on line 2, got %d", f.Line)
			}
		}
	}
	if !found {
		t.Errorf("expected unsafe-deserialization finding, got %+v", findings)
	}

	clean := "func add(a, b int) int {\n\treturn a + b\n}\n"
	if fs := ScanBuffer(clean); len(fs) != 0 {
		t.Errorf("expected no findings on clean fixture, got %+v", fs)
	}
}

// TestLayer1NeverBlocks asserts ScanBuffer is advisory only — it returns
// findings and never carries a block signal (REQ-SG-014). The GuardianFinding
// type has no "decision"/"block" field by construction, so this test documents
// the contract: a finding is data, not a gate.
func TestLayer1NeverBlocks(t *testing.T) {
	t.Parallel()

	findings := ScanBuffer(`element.innerHTML = userInput`)
	if len(findings) == 0 {
		t.Fatalf("expected a dom-injection-xss finding")
	}
	// A finding must never be interpretable as a block: GuardianFinding carries
	// only advisory fields (class/severity/message/line/match). The advisory
	// posture is enforced at the type level — there is no block field to set.
	for _, f := range findings {
		if f.Class == "" || f.Message == "" {
			t.Errorf("advisory finding must carry class + message: %+v", f)
		}
	}
}

// TestLayer1BinaryAndLargePayload asserts the edge cases: binary content is
// skipped (no finding spam) and a large payload is bounded (§C edge cases).
func TestLayer1BinaryAndLargePayload(t *testing.T) {
	t.Parallel()

	// Binary content (contains a NUL byte) with a would-be match — must be skipped.
	binary := "yaml.load(x)\x00\x01\x02binarygarbage"
	if fs := ScanBuffer(binary); len(fs) != 0 {
		t.Errorf("binary content must be skipped, got %+v", fs)
	}

	// A payload larger than the scan cap must not panic and must remain bounded.
	huge := strings.Repeat("safe line\n", 200000) // ~1.8MB, exceeds the 1MB cap
	_ = ScanBuffer(huge)                            // must not panic
}

// TestLayer1NoLLMOrSubprocess asserts the scan execution path imports no
// os/exec and no net/http and makes no model/tool call (AC-SG-004, REQ-SG-013).
// This is the Go-test half of AC-SG-004; the scoped grep is the shell half. The
// dangerous-pattern string literals legitimately live in patterns.go (excluded
// from the grep), so scan.go and diff.go — the actual scan execution path — are
// asserted clean here.
func TestLayer1NoLLMOrSubprocess(t *testing.T) {
	t.Parallel()

	// Match import lines / call sites, but not comment lines (a leading // is
	// excluded, mirroring the acceptance.md scoped grep).
	forbidden := regexp.MustCompile(`(exec\.Command|os/exec|net/http)`)
	for _, name := range []string{"scan.go", "diff.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue // comment-excluded, matches the AC grep filter
			}
			if forbidden.MatchString(line) {
				t.Errorf("%s:%d references a subprocess/network primitive on the scan path: %q", name, i+1, line)
			}
		}
	}
}
