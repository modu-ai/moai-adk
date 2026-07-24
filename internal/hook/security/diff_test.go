package security

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLayer2TurnDiff asserts the Layer-2 engine runs over a turn's unified diff
// and flags a dangerous ADDED line while ignoring context lines (AC-SG-007,
// REQ-SG-020).
func TestLayer2TurnDiff(t *testing.T) {
	t.Parallel()

	diff := `diff --git a/app.py b/app.py
index 111..222 100644
--- a/app.py
+++ b/app.py
@@ -1,3 +1,4 @@
 import yaml
-safe = 1
+data = yaml.load(request.body)
 done = True
`
	findings := ScanDiff(diff)
	if len(findings) == 0 {
		t.Fatalf("expected a finding on the added dangerous line, got 0")
	}
	got := false
	for _, f := range findings {
		if f.Class == "unsafe-deserialization" {
			got = true
		}
	}
	if !got {
		t.Errorf("expected unsafe-deserialization from the added line, got %+v", findings)
	}
}

// TestLayer2EmptyDiff asserts a whitespace-only diff no-ops cleanly (§C edge case).
func TestLayer2EmptyDiff(t *testing.T) {
	t.Parallel()
	if fs := ScanDiff("   \n\n"); len(fs) != 0 {
		t.Errorf("empty diff must produce no findings, got %+v", fs)
	}
}

// TestLayer3CrossFile asserts the Layer-3 cross-file review reads changed +
// related files and targets a cross-file class — the IDOR / broken-authorization
// shape: a user-supplied id source in one file, an object-access sink in a
// sibling file, and no authorization check anywhere (AC-SG-010, REQ-SG-030/031).
func TestLayer3CrossFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pkg := filepath.Join(root, "handlers")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	// route.js: user-supplied id source (changed file).
	routeSrc := "app.get('/doc/:id', (req, res) => {\n  const id = req.params.id;\n  render(id);\n});\n"
	if err := os.WriteFile(filepath.Join(pkg, "route.js"), []byte(routeSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	// repo.js: object-access sink, NO authorization check (related sibling file).
	repoSrc := "function render(id) {\n  return db.findById(id);\n}\n"
	if err := os.WriteFile(filepath.Join(pkg, "repo.js"), []byte(repoSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Only route.js is the "changed" file; repo.js is discovered as a related sibling.
	corpus := ReadRelatedFiles(root, []string{filepath.Join("handlers", "route.js")})
	if _, ok := corpus["handlers/route.js"]; !ok {
		t.Fatalf("changed file not read into corpus: %v", keys(corpus))
	}
	if _, ok := corpus["handlers/repo.js"]; !ok {
		t.Fatalf("related sibling file not read into corpus: %v", keys(corpus))
	}

	findings := CrossFileScan(corpus)
	idor := false
	for _, f := range findings {
		if f.Class == "cross-file-idor" {
			idor = true
		}
	}
	if !idor {
		t.Errorf("expected a cross-file-idor finding (source+sink+no-authz across files), got %+v", findings)
	}
}

// TestLayer3CrossFileAuthzSuppresses asserts the IDOR heuristic does NOT fire
// when an authorization check is present in the file set (false-positive guard).
func TestLayer3CrossFileAuthzSuppresses(t *testing.T) {
	t.Parallel()

	corpus := map[string]string{
		"route.js": "const id = req.params.id;",
		"repo.js":  "if (authorize(current_user, id)) { db.findById(id); }",
	}
	for _, f := range CrossFileScan(corpus) {
		if f.Class == "cross-file-idor" {
			t.Errorf("IDOR heuristic must be suppressed when an authz check is present: %+v", f)
		}
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
