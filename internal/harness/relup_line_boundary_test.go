package harness

// relup_line_boundary_test.go — card t1619, codex review round 2 (P2): a line-start anchor is
// a token, not a bare prefix. "### Phase 30" must not count as "### Phase 3".

import "testing"

func TestRelupExactlyOnceLineStartRespectsTokenBoundary(t *testing.T) {
	t.Run("a longer number is not a hit", func(t *testing.T) {
		lines := []string{"### Phase 30 — later", "### Phase 3 — real"}
		at, err := relupExactlyOnceLineStart(lines, "### Phase 3")
		if err != nil || at != 1 {
			t.Fatalf("exactly-once = (%d, %v), want (1, nil): '### Phase 30' is not a '### Phase 3' heading", at, err)
		}
	})

	t.Run("a lone longer number is zero hits", func(t *testing.T) {
		lines := []string{"### Phase 30 — later"}
		if at, err := relupExactlyOnceLineStart(lines, "### Phase 3"); err == nil {
			t.Fatalf("exactly-once accepted '### Phase 30' as '### Phase 3' (at line %d)", at)
		}
	})

	t.Run("end of line and punctuation are boundaries", func(t *testing.T) {
		for _, line := range []string{"### Phase 3", "### Phase 3:", "### Phase 3 —", "### Phase 3."} {
			if at, err := relupExactlyOnceLineStart([]string{line}, "### Phase 3"); err != nil || at != 0 {
				t.Errorf("line %q: exactly-once = (%d, %v), want (0, nil)", line, at, err)
			}
		}
	})
}
