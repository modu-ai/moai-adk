// hook_wrapper_single_original_test.go — replaces the retired pair-parity
// guard (hook_wrapper_pair_parity_test.go).
//
// History: through 2026-10 the hooks dir carried four .sh/.sh.tmpl twin
// pairs (handle-agent-hook, handle-stop-goal, handle-task-completed,
// handle-teammate-idle — 12,397 bytes shipped twice). The pair model —
// "the .sh.tmpl is what `moai update` deploys; the sibling .sh is the
// locally-executed counterpart" — required a drift guard because two
// originals can diverge: on 2026-08-15 lifecycle guards existed only in the
// .sh copies, and the next `moai update` silently downgraded deployed
// wrappers (the incident TestHookWrapperPairParity was written for).
//
// Card t1540 retired the model instead of guarding it: every hook ships
// exactly ONE original (the .sh.tmpl, rendered at deploy time), so the
// divergence hazard class is removed rather than watched. The former pair
// guard is gone with its pairs — its guard-of-the-guard failed at zero
// pairs by design, and this test takes over the loud-failure duty.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHookWrapperSingleOriginal asserts that no .sh/.sh.tmpl twin pair
// exists under the template hooks dir: each hook carries exactly one
// original form. Equivalent of
// `for f in *.sh; do [ -f "$f.tmpl" ] && echo "$f"; done` plus its
// .sh.tmpl-direction mirror — any twin fails the test.
func TestHookWrapperSingleOriginal(t *testing.T) {
	t.Parallel()

	root := hocProjectRoot(t)
	hooksDir := filepath.Join(root, "internal", "template", "templates", ".claude", "hooks", "moai")

	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		t.Fatalf("read template hooks dir %s: %v", hooksDir, err)
	}

	wrappers := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		switch {
		case strings.HasSuffix(name, ".sh.tmpl"):
			wrappers++
			twin := strings.TrimSuffix(name, ".tmpl")
			if _, statErr := os.Stat(filepath.Join(hooksDir, twin)); statErr == nil {
				t.Errorf("TWIN PAIR: %s and %s both exist — each hook must ship exactly ONE original (the .sh.tmpl, rendered at deploy). The .sh twin double-deploys to the same destination and re-opens the pair-drift hazard retired with card t1540; delete the .sh copy.", name, twin)
			}
		case strings.HasSuffix(name, ".sh"):
			wrappers++
			twin := name + ".tmpl"
			if _, statErr := os.Stat(filepath.Join(hooksDir, twin)); statErr == nil {
				t.Errorf("TWIN PAIR: %s and %s both exist — delete the .sh copy; the .sh.tmpl is the single original (card t1540).", name, twin)
			}
		}
	}

	// Guard-of-the-guard: a moved or emptied hooks dir would pass vacuously
	// and hide future regressions — fail loudly instead.
	if wrappers < 20 {
		t.Fatalf("only %d hook wrappers found under %s — directory path wrong or wrappers removed; check before trusting this guard", wrappers, hooksDir)
	}
}
