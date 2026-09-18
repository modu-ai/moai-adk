package hook

import "testing"

// TestBranchStatePatterns_HeredocBodyIsData is the t807 reproduction: a
// heredoc body is DATA fed to the command's stdin, never a command that runs.
// Quoted arguments are already collapsed (substituteQuotedArguments), but a
// heredoc body carries no quotes, so guarded git prose inside one was scanned
// as if it were the command being run.
//
// Measured incidents (2026-09-10): `moai handoff save --stdin … <<EOF … EOF`
// was denied with `BRANCH_GUARD_VIOLATION: git merge in primary checkout`
// because the saved resume body quoted `git merge --no-ff <sha>`; the lane
// then skipped the save, closing the handoff-record path.
func TestBranchStatePatterns_HeredocBodyIsData(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
	}{
		{
			name:    "handoff save body naming git merge",
			command: "moai handoff save --stdin --spec SPEC-X --phase run <<EOF\nResume: after the window, git merge --no-ff WT-card into develop\nEOF",
		},
		{
			name:    "quoted-delimiter heredoc naming git switch",
			command: "cat <<'EOF' > /tmp/note\ngit switch main\nEOF",
		},
		{
			name:    "heredoc body naming git reset --hard",
			command: "moai handoff save --stdin <<EOF\nnever run git reset --hard HEAD here\nEOF",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if suffix, matched := matchBranchStateCommand(tc.command); matched {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, true), want false — heredoc body is data", tc.command, suffix)
			}
		})
	}
}

// TestBranchStatePatterns_HeredocDoesNotBlindTheGuard is the falsification arm
// of the test above: collapsing heredoc bodies MUST NOT un-guard a real
// branch-state command that merely sits next to one. Without this pin, a fix
// that discarded the whole command from the first `<<` onward would pass the
// true-negative table vacuously.
func TestBranchStatePatterns_HeredocDoesNotBlindTheGuard(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "branch-state command before a heredoc",
			command: "git switch main && cat <<EOF > /tmp/note\nplain text\nEOF",
			want:    "git switch",
		},
		{
			name:    "branch-state command after a heredoc closes",
			command: "cat <<EOF > /tmp/note\nplain text\nEOF\ngit merge --no-ff WT-card",
			want:    "git merge",
		},
		{
			name:    "no heredoc at all (control)",
			command: "git merge --no-ff WT-card",
			want:    "git merge",
		},
		{
			// Security arm: a `<<EOF` sitting INSIDE a quoted argument opens no
			// heredoc — the shell reads it as text. Treating it as an opener
			// would let any command blind the guard for every following line
			// just by quoting the token.
			name:    "quoted <<EOF is text, not an opener",
			command: "moai todo add \"note about <<EOF\"\ngit switch main",
			want:    "git switch",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if !matched {
				t.Fatalf("matchBranchStateCommand(%q) = (_, false), want match %q", tc.command, tc.want)
			}
			if suffix != tc.want {
				t.Fatalf("matchBranchStateCommand(%q) suffix = %q, want %q", tc.command, suffix, tc.want)
			}
		})
	}
}
