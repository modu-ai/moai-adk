package cli

// N1 of the final audit of a1ec370195: a failed gate's output is masked before its control characters are removed,
// so a marker or a value split by a control character or an escape sequence is not recognised, and the text that
// remains once the controls are gone reads as a plain `password=<value>`. The Cookie rule also consumes the rest of
// its line, a trailing `password=` included, and the value on the next line loses its marker. Real ledger, real git,
// the real `moai af land` verb; the secret is assembled by the gate from $A and $B, so it appears only in its output.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

const afSecretValue = "hunter2super"

// afOutputOf is what a reason holds of the gate's output: the text after the exit status. The part before it quotes
// the gate command, whose own text ("B=super") is not output and must not be mistaken for a leak.
func afOutputOf(reason string) string {
	if i := strings.Index(reason, "exit status 1: "); i >= 0 {
		return reason[i+len("exit status 1: "):]
	}
	return reason
}

// afFailReason is the output part of the fail reason the ledger holds for the one queued card.
func afFailReason(t *testing.T, st *ledger.Store) string {
	t.Helper()
	rows, err := st.Queue(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	return afOutputOf(rows[0].FailReason)
}

func TestAFLand_ASecretSplitByAControlOrAnEscapeIsStillMasked(t *testing.T) {
	afSkipWithoutSh(t)
	cases := []struct {
		name, printf, leak string
	}{
		{"control_plain_assignment", `password=${A}${B}`, afSecretValue},
		{"value_split_by_BEL", `password=${A}\a${B}`, afSecretValue},
		{"value_split_by_ESC_sequence", `password=${A}\033[0m${B}`, afSecretValue},
		{"value_split_by_CR", `password=${A}\r${B}`, "super"},
		{"value_split_by_FF", `password=${A}\f${B}`, "super"},
		{"value_coloured_whole", `password=\033[31m${A}${B}\033[0m`, afSecretValue},
		{"marker_split_by_BEL", `pass\aword=${A}${B}`, afSecretValue},
		{"marker_split_by_NUL", `pass\000word=${A}${B}`, afSecretValue},
		{"marker_coloured_equals", `\033[1mpassword\033[0m=${A}${B}`, afSecretValue},
		{"marker_coloured_colon", `\033[32mAPI_KEY\033[0m: ${A}${B}`, afSecretValue},
		{"marker_coloured_only_reset_before_sep", `password\033[0m=${A}${B}`, afSecretValue},
		{"authorization_split", `Authorization: Bea\arer ${A}${B}abcdefgh`, afSecretValue},
		{"url_credential_split", `https://user:${A}\a${B}@example.test/x`, afSecretValue},
		{"pem_header_split", `-----BEGIN\a PRIVATE KEY-----\nMIIE${A}${B}\n-----END PRIVATE KEY-----`, afSecretValue},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, repo, _, st, _ := afQueuedCard(t)
			gate := `A=hunter2; B=super; printf "before\n` + c.printf + `\nafter\n"; exit 1`
			out := runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
			var res struct{ Outcome, Reason string }
			if err := json.Unmarshal([]byte(out), &res); err != nil {
				t.Fatal(err)
			}
			led := afFailReason(t, st)
			if shown := afOutputOf(res.Reason); strings.Contains(shown, c.leak) || strings.Contains(led, c.leak) {
				t.Errorf("%q survives: printed %q, ledger %q", c.leak, shown, led)
			}
			if !strings.Contains(led, "after") {
				t.Errorf("masking must keep the rest of the output readable, ledger tail is %q", led)
			}
		})
	}
}

// A line that ends on `Cookie: ... password=` hands the next line's value to nothing: the Cookie rule took the marker.
func TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	gate := `A=hunter2; B=super; printf 'Cookie: %2100s password=\n' x; printf "${A}${B}\n"; exit 1`
	runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	if led := afFailReason(t, st); strings.Contains(led, afSecretValue) {
		t.Errorf("the value on the line after a Cookie line survives: %q", led)
	}
}

// The writer keeps the end of the output. A value that began before the kept part has lost its marker, and the
// first kept line is a fragment: it is not shown.
func TestAFLand_AValueWhoseMarkerFellOffTheKeptTailIsNotShown(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	gate := `printf 'password='; i=0; while [ $i -lt 7000 ]; do printf 'abc.def.gh'; i=$((i+1)); done; printf '\nlast line\n'; exit 1`
	runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	led := afFailReason(t, st)
	if strings.Contains(led, "abc.def.gh") {
		t.Errorf("a 70 KB value whose marker fell off the 64 KiB tail survives (%d bytes)", len(led))
	}
	if !strings.Contains(led, "last line") {
		t.Errorf("the lines after the cut one must stay, ledger tail is %q", led)
	}
}

// Guards: both pass before the repair and must keep passing.
func TestAFLand_AValueStraddlingTheCutAndControlsBeyondItStayMasked(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	gate := `A=hunter2; B=secretvalue; printf "password=${A}${B}"; printf '%2025s' x; exit 1`
	runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	if led := afFailReason(t, st); strings.Contains(led, "value") {
		t.Errorf("the end of a value straddling the 2048-byte cut survives: %q", led[:min(len(led), 80)])
	}
}
