// Package cli — codex_role_extract_test.go
//
// Card t1260 (t1171 sync-audit F2): a multi-line string value written BEFORE
// developer_instructions — the emitter writes a triple-quoted description
// first —
// may contain lines that look like a table header or like the
// developer_instructions key itself. Those lines are value content, not
// structure, and must not end or hijack the key scan.
package cli

import "testing"

func TestCodexRoleBodyExtractSkipsOtherMultilineValues(t *testing.T) {
	const want = "real body\n"
	cases := map[string]string{
		// The two shapes the audit probed (tomllib yields "real body\n").
		"desc_line_starts_with_bracket": "name = \"r\"\ndescription = '''\n[HARD] a bracketed line\n'''\ndeveloper_instructions = '''\nreal body\n'''\n",
		"desc_line_mimics_key":          "name = \"r\"\ndescription = '''\ndeveloper_instructions = \"decoy\"\n'''\ndeveloper_instructions = '''\nreal body\n'''\n",
		// Same hazards inside a multi-line BASIC string.
		"basic_desc_bracket_and_key": "description = \"\"\"\n[HARD]\ndeveloper_instructions = '''decoy'''\n\"\"\"\ndeveloper_instructions = '''\nreal body\n'''\n",
		// A value that opens and closes on one line must not start a skip.
		"single_line_triple_quoted": "description = '''one line'''\ndeveloper_instructions = '''\nreal body\n'''\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := codexRoleBodyExtractTOML(src)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}

	t.Run("unclosed_other_value_is_an_error", func(t *testing.T) {
		// An unterminated earlier multi-line value swallows the rest of the
		// document; the key is then genuinely absent and must be reported.
		src := "description = '''\nnever closed\ndeveloper_instructions = '''\nbody\n"
		if _, err := codexRoleBodyExtractTOML(src); err == nil {
			t.Fatal("expected an error for a document whose only key sits inside an unclosed value")
		}
	})
}
