package kanban

import "testing"

// Card t1126 — the kanban frontmatter reader must normalize a YAML-quoted
// `status:` value the same way the spec readers do, so a card's column is
// derived from `completed`, not from `"completed"`. Mismatched quotes are not
// a quoted scalar and are left exactly as written.
func TestParseFrontmatterStatus_QuotedValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, raw, want string
	}{
		{"bare", `completed`, "completed"},
		{"double-quoted", `"completed"`, "completed"},
		{"single-quoted", `'in-progress'`, "in-progress"},
		{"mismatched quotes are not stripped", `"completed'`, `"completed'`},
		{"lone leading quote is not stripped", `"completed`, `"completed`},
	}
	for _, tc := range cases {
		doc := "---\nid: SPEC-QUOTED-001\nstatus: " + tc.raw + "\n---\n"
		got, ok := parseFrontmatterStatus([]byte(doc))
		if !ok {
			t.Fatalf("%s: parseFrontmatterStatus found no status", tc.name)
		}
		if got != tc.want {
			t.Errorf("%s: parseFrontmatterStatus(status: %s) = %q, want %q", tc.name, tc.raw, got, tc.want)
		}
	}
}
