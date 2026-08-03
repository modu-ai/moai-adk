package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// AC-UYP-007 — a key retained by DeepMerge3Way (absent from the new template)
// is reported on stderr as an advisory naming the retained key path.
func TestMergeYAML3Way_ReportsRetainedKey(t *testing.T) {
	newData := []byte("shared: new_val\n")
	oldData := []byte("shared: new_val\nuser_added: custom\n")
	baseData := []byte("shared: base_val\n")

	var advisory bytes.Buffer
	oldSink := retainedKeySink
	retainedKeySink = &advisory
	defer func() { retainedKeySink = oldSink }()

	if _, err := MergeYAML3Way(newData, oldData, baseData); err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if !strings.Contains(advisory.String(), "user_added") {
		t.Errorf("advisory should name user_added (REQ-UYP-007), got: %s", advisory.String())
	}
}

// AC-UYP-008 — the 3-way path is never more destructive than the 2-way
// fallback: every top-level key present in the 2-way output is also present in
// the 3-way output.
func TestMerge3WayNotMoreDestructiveThan2Way(t *testing.T) {
	cases := []struct {
		name    string
		newData string
		oldData string
		base    string
	}{
		{"user-added key absent from new", "shared: v\n", "shared: v\nextra: 1\n", "shared: b\n"},
		{"nested user addition", "a:\n  x: 1\n", "a:\n  x: 1\n  y: 2\n", "a:\n  x: 1\n"},
		{"template-removed key", "kept: v\n", "kept: v\nremoved: old\n", "kept: v\nremoved: old\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			two, err := MergeYAMLDeep([]byte(c.newData), []byte(c.oldData))
			if err != nil {
				t.Fatalf("MergeYAMLDeep: %v", err)
			}
			var sink bytes.Buffer
			oldSink := retainedKeySink
			retainedKeySink = &sink
			defer func() { retainedKeySink = oldSink }()
			three, err := MergeYAML3Way([]byte(c.newData), []byte(c.oldData), []byte(c.base))
			if err != nil {
				t.Fatalf("MergeYAML3Way: %v", err)
			}
			twoRoot, err := decodeDoc(two, "2way")
			if err != nil {
				t.Fatal(err)
			}
			threeRoot, err := decodeDoc(three, "3way")
			if err != nil {
				t.Fatal(err)
			}
			// Every top-level scalar key in the 2-way output MUST appear in 3-way.
			for i := 0; i+1 < len(twoRoot.Content); i += 2 {
				k := twoRoot.Content[i]
				if k == nil || k.Kind != yaml.ScalarNode {
					continue
				}
				if !mappingHasKey(threeRoot, k.Value) {
					t.Errorf("key %q in 2-way output missing from 3-way output (REQ-UYP-008)", k.Value)
				}
			}
		})
	}
}

// AC-UYP-015 — multi-document input errors rather than silently truncating.
func TestMergeYAML3Way_MultiDocumentErrors(t *testing.T) {
	multi := []byte("a: 1\n---\nb: 2\n")
	_, err := MergeYAML3Way(multi, []byte("a: 1\n"), []byte("a: 1\n"))
	if err == nil {
		t.Fatal("expected error for multi-document newData")
	}
	if !strings.Contains(err.Error(), "multi-document") {
		t.Errorf("error should name multi-document, got: %v", err)
	}
}

// AC-UYP-023 — quality.yaml.tmpl's unquoted {{.X}} placeholders make the MAP
// decoder fail but the NODE decoder succeeds (the placeholder lands as scalar
// text). The merge returns a nil error, promoting the file from always-2-way
// to real-3-way.
func TestMergeYAML3Way_QualityTemplateParses(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(sectionTemplatesDir, "quality.yaml.tmpl"))
	if err != nil {
		t.Skipf("quality.yaml.tmpl not found: %v", err)
	}
	// Baseline: the map decoder DOES error on this body (the defect condition).
	mapDecodeErr := yaml.Unmarshal(src, &map[string]any{})

	out, mergeErr := MergeYAML3Way(src, src, src)
	if mergeErr != nil {
		t.Fatalf("node decoder should parse quality.yaml.tmpl (nil error), got: %v", mergeErr)
	}
	if len(out) == 0 {
		t.Error("expected non-empty merged output for quality.yaml.tmpl")
	}
	// The baseline map-decode error is the discriminator: if a future refactor
	// regressed the node decoder to a map decoder, this baseline would stop
	// erroring and the assertion above would silently keep passing.
	if mapDecodeErr == nil {
		t.Error("baseline map decoder should error on quality.yaml.tmpl (defect condition); if it no longer does, the discriminator is stale")
	}
}

// AC-UYP-021 — end-to-end restore merge preserves a customized section:
// comments, the user edit, and quoting all survive. cache.yaml is the
// byte-stable fixture; workflow.yaml is the DIFFER-class fixture (byte-equality
// does not hold, but the property set does).
func TestUpdateEndToEnd_PreservesCustomizedSection(t *testing.T) {
	cacheTpl, err := os.ReadFile(filepath.Join(sectionTemplatesDir, "cache.yaml"))
	if err != nil {
		t.Skipf("cache.yaml template not found: %v", err)
	}
	wfTpl, err := os.ReadFile(filepath.Join(sectionTemplatesDir, "workflow.yaml"))
	if err != nil {
		t.Skipf("workflow.yaml template not found: %v", err)
	}

	t.Run("cache.yaml", func(t *testing.T) {
		// User edit: enabled flipped to false + hand-added user_note.
		userCache := bytes.Replace(cacheTpl, []byte("enabled: true"), []byte("enabled: false"), 1)
		userCache = append(userCache, []byte("\nuser_note: \"keep me\"\n")...)

		var sink bytes.Buffer
		oldSink := retainedKeySink
		retainedKeySink = &sink
		defer func() { retainedKeySink = oldSink }()
		merged, err := MergeYAML3Way(cacheTpl, userCache, cacheTpl)
		if err != nil {
			t.Fatalf("MergeYAML3Way: %v", err)
		}
		out := string(merged)
		if !strings.Contains(out, "enabled: false") {
			t.Errorf("user edit (enabled: false) lost:\n%s", out)
		}
		if !strings.Contains(out, "user_note") {
			t.Errorf("hand-added user_note lost:\n%s", out)
		}
		if !strings.Contains(out, `"1h"`) {
			t.Errorf(`session_ttl quoting lost (should stay "1h")`+":\n%s", out)
		}
		if in, got := countCommentLines(cacheTpl), countCommentLines(merged); in != got {
			t.Errorf("comment line count: template=%d merged=%d", in, got)
		}
	})

	t.Run("workflow.yaml", func(t *testing.T) {
		// DIFFER class: byte-equality does not hold, but the comment-line count
		// (property set) must survive.
		userWf := wfTpl // no-edit: pure round-trip
		var sink bytes.Buffer
		oldSink := retainedKeySink
		retainedKeySink = &sink
		defer func() { retainedKeySink = oldSink }()
		merged, err := MergeYAML3Way(wfTpl, userWf, wfTpl)
		if err != nil {
			t.Fatalf("MergeYAML3Way: %v", err)
		}
		if in, got := countCommentLines(wfTpl), countCommentLines(merged); in != got {
			t.Errorf("workflow.yaml comment line count: template=%d merged=%d", in, got)
		}
	})
}
