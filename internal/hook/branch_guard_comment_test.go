package hook

import "testing"

// TestBranchStatePatterns_ShellCommentIsNotACommand is Arm A of the
// SPEC-GUARD-COMMENT-SCAN-001 two-armed matrix: a shell comment is text the
// shell strips before execution, so branch-state git prose carried inside one
// can never be the command being run. Every row places the branch-state text
// AFTER the `#` (AC-GCS-001), which is what gives the rows their
// discriminating power — and every row elides to a command that must not
// match.
//
// Measured false positive this closes (spec.md §A, HEAD 3dfae918a):
// matched=true suffix="git merge" cmd=# align with git merge --ff-only develop
func TestBranchStatePatterns_ShellCommentIsNotACommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
	}{
		{
			name:    "whole-line comment naming git merge (AC-GCS-001 r1)",
			command: "# align with git merge --ff-only develop",
		},
		{
			name:    "trailing comment after a harmless command (AC-GCS-001 r2)",
			command: "moai todo list  # then git switch main",
		},
		{
			name:    "comment after a separator and a space (AC-GCS-001 r3)",
			command: "ls ; # git reset --hard HEAD",
		},
		{
			// The separator half of REQ-GCS-002's word-start set: `;#` with no
			// intervening space still opens a comment in bash (acceptance.md
			// r4 measurement), so a whitespace-only word-start rule fails here.
			name:    "comment opened by a separator with no space (AC-GCS-001 r4)",
			command: "ls ;# git reset --hard HEAD",
		},
		{
			// Pins where the comment run STARTS (REQ-GCS-004): the run opens at
			// the FIRST word-start `#` — eliding from the last one would leave
			// the prose scannable and re-match `git reset --hard`.
			name:    "second # inside an already-open comment run (AC-GCS-001 r5)",
			command: "ls ; # prose about reset --hard HEAD per #123",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if suffix, matched := matchBranchStateCommand(tc.command); matched {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, true), want false — a shell comment is text the shell never executes", tc.command, suffix)
			}
		})
	}
}

// TestBranchStatePatterns_CommentCollapseDoesNotBlindTheGuard is Arm B of the
// same matrix (AC-GCS-002 … AC-GCS-006): eliding comments must never un-guard
// a real branch-state command. Arm A alone is satisfied equally by a correct
// narrowing and by a guard blunted into matching nothing, so every row here
// asserts its expected suffix — never a bare "no match" — and a blunted guard
// fails loudly. The single no-match row (AC-GCS-005) is the elide-vs-placeholder
// discriminator: eliding must leave `git checkout -b` with NO operand, which
// the checkout pattern cannot match, whereas substituting the operand
// placeholder would present one and deny a command the shell itself rejects.
func TestBranchStatePatterns_CommentCollapseDoesNotBlindTheGuard(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		command   string
		want      string
		wantMatch bool
	}{
		// AC-GCS-002 — real branch-state commands still match, comment or no
		// comment. Rows 3-4 fix the DIRECTION of the elision: the command and
		// the comment share a line, so discarding the whole line fails here.
		{
			name:      "no comment at all: git merge (AC-GCS-002 r1)",
			command:   "git merge --ff-only develop",
			want:      "git merge",
			wantMatch: true,
		},
		{
			name:      "no comment at all: git switch (AC-GCS-002 r2)",
			command:   "git switch main",
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "trailing comment, command survives on the same line (AC-GCS-002 r3)",
			command:   "git merge --ff-only develop  # per lead instruction",
			want:      "git merge",
			wantMatch: true,
		},
		{
			name:      "trailing comment, git switch survives on the same line (AC-GCS-002 r4)",
			command:   "git switch main  # move to main",
			want:      "git switch",
			wantMatch: true,
		},
		// AC-GCS-003 — the elision is bounded to the physical line the comment
		// opens on; a real command on the next line is fully scannable.
		{
			name:      "comment line, then real command on the next line (AC-GCS-003)",
			command:   "# this line is prose about git merge\ngit switch main",
			want:      "git switch",
			wantMatch: true,
		},
		// AC-GCS-004 — a `#` inside a word is a literal hash, not a comment.
		// Rows 1-2 are documentary (the ordinary alphanumeric case); rows 3-6
		// are one falsifying row per non-alphanumeric character spec.md §B.1
		// measured as leaving the `#` mid-word, each carrying a real command on
		// the far side of the hash.
		{
			name:      "alphanumeric-preceded # in a branch name (AC-GCS-004 r1, documentary)",
			command:   "git switch feat#123",
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "alphanumeric-preceded # in a merge topic (AC-GCS-004 r2, documentary)",
			command:   "git merge topic#7",
			want:      "git merge",
			wantMatch: true,
		},
		{
			name:      "slash-preceded mid-word #, command after it survives (AC-GCS-004 r3)",
			command:   "v=bar/#x ; git switch main",
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "dotted version operand, command after it survives (AC-GCS-004 r4)",
			command:   "v=rel-1.0#rc2 ; git switch main",
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "equals-preceded mid-word #, command after it survives (AC-GCS-004 r5)",
			command:   "v=a=#x ; git switch main",
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "hyphen-preceded mid-word #, command after it survives (AC-GCS-004 r6)",
			command:   "v=a-#x ; git switch main",
			want:      "git switch",
			wantMatch: true,
		},
		// AC-GCS-005 — the comment run is elided, NOT replaced by the operand
		// placeholder (REQ-GCS-004, spec.md §B.3).
		{
			name:      "comment after -b elides: no operand presented (AC-GCS-005)",
			command:   "git checkout -b # x",
			want:      "",
			wantMatch: false,
		},
		// AC-GCS-006 — pipeline ordering: the comment step runs LAST, so a `#`
		// inside a quoted span or a heredoc body has already collapsed to the
		// placeholder and opens no comment.
		{
			name:      "quoted # opens no comment (AC-GCS-006)",
			command:   `echo "text # more" ; git switch main`,
			want:      "git switch",
			wantMatch: true,
		},
		{
			name:      "heredoc-body # opens no comment (AC-GCS-006 companion)",
			command:   "cat <<EOF > /tmp/note\n# git merge --no-ff WT-card\nEOF\ngit switch main",
			want:      "git switch",
			wantMatch: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if matched != tc.wantMatch {
				t.Fatalf("matchBranchStateCommand(%q) matched = %v (suffix %q), want match %v", tc.command, matched, suffix, tc.wantMatch)
			}
			if matched && suffix != tc.want {
				t.Fatalf("matchBranchStateCommand(%q) suffix = %q, want %q", tc.command, suffix, tc.want)
			}
		})
	}
}
