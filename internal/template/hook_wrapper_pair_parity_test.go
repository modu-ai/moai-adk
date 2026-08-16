// hook_wrapper_pair_parity_test.go
// Pair-drift guard for hook wrapper templates.
//
// The .sh.tmpl file is what `moai update` deploys to user projects; the sibling
// .sh file in the template source is the locally-executed counterpart. When an
// edit lands only in the .sh (e.g. a new guard block), the next `moai update`
// re-renders the stale .tmpl and silently downgrades the deployed wrapper —
// the pair-drift hazard measured on 2026-08-15 where the lifecycle dormant
// guards existed only in the .sh copies.
//
// This guard walks every .sh.tmpl under the template hooks dir and asserts
// byte-identity with its .sh sibling when one exists. A .tmpl without a .sh
// sibling is out of scope (nothing to drift against).
//
// @MX:ANCHOR: [AUTO] hook wrapper pair-drift guard — protects the deployable .tmpl from regressing behind its .sh source.
// @MX:REASON: Without this guard, an edit that lands only in the .sh is silently reverted in deployed user projects on the next moai update.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHookWrapperPairParity asserts that every .sh.tmpl hook wrapper with a
// sibling .sh in the template source is byte-identical to it. Equivalent of
// `for f in *.sh.tmpl; do diff "${f%.tmpl}" "$f"; done` run over the template
// hooks directory — any mismatch fails the test.
func TestHookWrapperPairParity(t *testing.T) {
	root := hocProjectRoot(t)
	hooksDir := filepath.Join(root, "internal", "template", "templates", ".claude", "hooks", "moai")

	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		t.Fatalf("read template hooks dir %s: %v", hooksDir, err)
	}

	checked := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".sh.tmpl") {
			continue
		}
		tmplPath := filepath.Join(hooksDir, name)
		shPath := strings.TrimSuffix(tmplPath, ".tmpl")
		if _, statErr := os.Stat(shPath); os.IsNotExist(statErr) {
			// .tmpl-only wrapper (no local pair) — nothing to drift against.
			continue
		}

		tmplData, readErr := os.ReadFile(tmplPath)
		if readErr != nil {
			t.Fatalf("read %s: %v", name, readErr)
		}
		shData, readErr := os.ReadFile(shPath)
		if readErr != nil {
			t.Fatalf("read %s: %v", filepath.Base(shPath), readErr)
		}

		checked++
		if string(tmplData) != string(shData) {
			t.Errorf("PAIR DRIFT: %s differs from %s — the .tmpl is what moai update deploys, so an edit that landed only in the .sh is reverted in user projects on the next update. Edit BOTH files (or copy the .sh over the .tmpl)",
				name, filepath.Base(shPath))
		}
	}

	// Guard-of-the-guard: a zero-pair run would pass vacuously and hide future
	// regressions. If the hooks dir moves, fail loudly instead.
	if checked == 0 {
		t.Fatalf("no .sh/.sh.tmpl pairs found under %s — guard is not exercising anything; check the directory path", hooksDir)
	}
	t.Logf("verified %d hook wrapper pairs are byte-identical", checked)
}
