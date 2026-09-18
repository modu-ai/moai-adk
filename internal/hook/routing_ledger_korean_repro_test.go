package hook

import "testing"

// TestLiteralSubcommandKoreanParticle is the card t926 reproduction.
//
// literalSubcommand splits on whitespace (strings.Fields) and takes the first
// token. Korean particles attach to the preceding word WITHOUT a space, so a
// prompt written in this project's own conversation_language yields a token
// that is the subcommand plus a particle — "todo를", not "todo".
//
// The live ledger carries two such rows, enumerated exhaustively at
// .moai/state/routing-ledger.jsonl lines 248 and 282:
//
//	project에   (1 row)
//	todo를      (1 row)
//
// Neither value matches any key in .moai/config/sections/delegation.yaml, so
// those rows are excluded from every designation comparison downstream — they
// are orphaned, not merely mislabelled.
//
// Every pre-existing test on this seam uses an ASCII English prompt
// ("/moai plan add auth", "/moai run SPEC-X"), which is why the whole class was
// invisible: the coverage gap and the defect share the same shape.
func TestLiteralSubcommandKoreanParticle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		prompt string
		want   string
	}{
		// The defect, in the two forms the live ledger actually carries.
		{"object particle", "/moai todo를 확인해줘", "todo"},
		{"locative particle", "/moai project에 대해 알려줘", "project"},

		// Controls that must keep their current behaviour, so a fix cannot be
		// mistaken for a regression in the paths that already work.
		{"ascii unchanged", "/moai plan add auth", "plan"},
		{"ascii with spec id", "/moai run SPEC-X", "run"},
		{"bare subcommand", "/moai sync", "sync"},
		{"flag is not a subcommand", "/moai --help", ""},

		// The digit case is not decoration: `e2e` is a real subcommand, and an
		// extraction alphabet written as [a-z-] drops it silently. That exact
		// omission was hit four separate times while this batch was measured,
		// so it is pinned here rather than left to review.
		{"digit-bearing subcommand", "/moai e2e web", "e2e"},
		{"digit-bearing bare", "/moai e2e", "e2e"},

		// A hyphen is legal INSIDE a token and meaningless trailing it.
		{"internal hyphen kept", "/moai some-sub arg", "some-sub"},
		{"trailing hyphen trimmed", "/moai e2e-", "e2e"},
		{"only hyphens is not a subcommand", "/moai ---", ""},

		// A deliberate behaviour CHANGE, pinned so it is visible rather than
		// silent: an uppercase token used to be recorded verbatim and is now
		// dropped. No subcommand is uppercase, so the old value could only ever
		// be a label nothing downstream matches — and recording nothing is what
		// this seam already does for every input it cannot attribute.
		{"uppercase is not a subcommand", "/moai Plan add auth", ""},

		// The control that establishes the EMPTY value is by design, not a
		// matching failure: a natural-language prompt records nothing because
		// literalSubcommand requires the literal prefix.
		{"natural language records nothing", "오늘 뭐 하면 좋을까?", ""},
		{"prose mentioning moai records nothing", "moai 로 계획 세워줘", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := literalSubcommand(tc.prompt); got != tc.want {
				t.Errorf("literalSubcommand(%q) = %q, want %q", tc.prompt, got, tc.want)
			}
		})
	}
}
