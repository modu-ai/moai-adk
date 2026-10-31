package cli

// N1 of the merged-tree audit: a gate prints a secret marker followed by a value longer than the 64 KiB the gate
// output writer keeps. The writer drops the start of the output before anything can mask it, so the marker is gone
// and the end of the value is recorded and shown. These tests pin the shapes: the value must not appear in the text
// afGate returns, in the JSON `af land` prints, or in the fail reason the ledger holds; a loud gate with no marker
// must still show its last line; and a value that was masked says so.
//
// The value is a run of Q: it has no digit, so the generic opaque-run rule never hides it, and only the
// marker-based masking can.

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

const (
	afQRun  = "QQQQQQQQ" // eight Q in a row: a piece of the value
	afMaskd = "masked"   // what the reason says when a value was masked for want of its marker
)

// afQ is a shell command that prints n Q and nothing else.
func afQ(n int) string { return `head -c ` + strconv.Itoa(n) + ` /dev/zero | tr '\0' Q` }

// afGateTailOf runs the gate through afGate and returns the output part of the error it fails with.
func afGateTailOf(t *testing.T, gate string) string {
	t.Helper()
	afSkipWithoutSh(t)
	err := afGate(gate, 60*time.Second)(context.Background(), t.TempDir())
	if err == nil {
		t.Fatalf("gate %q exited 0, want a failure", gate)
	}
	return afOutputOf(err.Error())
}

func TestAFGate_ALongValueWhoseMarkerIsOutsideTheKeptWindowIsMasked(t *testing.T) {
	afSkipWithoutSh(t)
	keep := afGateKeep
	cases := []struct {
		name, gate string
		lost       bool // the marker (or its first byte) is outside the kept window: the reason must say a value was masked
	}{
		{"no_newline_after_value", `printf 'password='; ` + afQ(70000) + `; exit 1`, true},
		{"trailing_newline", `printf 'password='; ` + afQ(70000) + `; printf '\n'; exit 1`, true},
		{"other_lines_follow", `printf 'password='; ` + afQ(70000) + `; printf '\nnext line\n'; exit 1`, true},
		{"long_output_follows", `printf 'password='; ` + afQ(70000) + `; printf '\n'; yes 'filler line' | head -n 400; exit 1`, true},
		{"upper_case_marker", `printf 'PASSWORD='; ` + afQ(70000) + `; exit 1`, true},
		{"api_key_colon_space", `printf 'Api_Key: '; ` + afQ(70000) + `; exit 1`, true},
		{"token_marker", `printf 'auth_token='; ` + afQ(70000) + `; exit 1`, true},
		{"cookie_value", `printf 'Cookie: session='; ` + afQ(70000) + `; exit 1`, true},
		{"marker_inside_a_long_prose_line", `printf 'the client sent password='; ` + afQ(70000) + `; exit 1`, true},
		{"marker_ends_the_previous_line", `printf 'password=\n'; ` + afQ(70000) + `; exit 1`, true},
		{"marker_ends_the_previous_line_cookie_style", `printf 'Cookie: a=b; password=\n'; ` + afQ(70000) + `; exit 1`, true},
		{"marker_split_by_BEL", `printf 'pass\007word='; ` + afQ(70000) + `; exit 1`, true},
		{"marker_coloured", `printf '\033[1mpassword\033[0m='; ` + afQ(70000) + `; exit 1`, true},
		{"value_split_by_escape", `printf 'password='; ` + afQ(40000) + `; printf '\033[0m'; ` + afQ(40000) + `; exit 1`, true},
		// around exactly the kept window (the marker `password=` is 9 bytes)
		{"window_inside_by_one", `printf 'password='; ` + afQ(keep-9-1) + `; exit 1`, false},
		{"window_exactly_full", `printf 'password='; ` + afQ(keep-9) + `; exit 1`, false},
		{"window_marker_first_byte_cut", `printf 'password='; ` + afQ(keep-9+1) + `; exit 1`, true},
		{"window_marker_exactly_cut", `printf 'password='; ` + afQ(keep) + `; exit 1`, true},
		{"window_marker_cut_by_nine_more", `printf 'password='; ` + afQ(keep+9) + `; exit 1`, true},
		{"window_value_line_exactly_fills", `printf 'password=\n'; ` + afQ(keep) + `; exit 1`, true},
		{"window_value_line_one_short", `printf 'password=\n'; ` + afQ(keep-1) + `; exit 1`, true},
		{"window_value_line_untruncated", `printf 'password=\n'; ` + afQ(keep-10) + `; exit 1`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tail := afGateTailOf(t, c.gate)
			if strings.Contains(tail, afQRun) {
				t.Errorf("a piece of the value survives (%d bytes of tail): %.80q ... %.80q", len(tail), tail, tail[max(0, len(tail)-80):])
			}
			if c.lost && !strings.Contains(strings.ToLower(tail), afMaskd) {
				t.Errorf("the value was masked for want of its marker but the reason does not say so: %.200q", tail)
			}
			if len(tail) > afGateTail {
				t.Errorf("the tail is %d bytes, over the %d the reason may hold", len(tail), afGateTail)
			}
		})
	}
}

// A loud gate with no marker keeps showing its last line, even when that line is the end of a huge one.
func TestAFGate_ALoudGateWithNoMarkerStillShowsItsLastLine(t *testing.T) {
	afSkipWithoutSh(t)
	cases := []struct{ name, gate, want string }{
		{"last_line_after_a_huge_line", afQ(70000) + `; printf '\nREAL-FAILURE-LINE\n'; exit 1`, "REAL-FAILURE-LINE"},
		{"last_line_is_the_end_of_a_huge_line", `printf 'plain text '; ` + afQ(70000) + `; exit 1`, afQRun},
		{"huge_line_with_a_newline_at_the_end", `printf 'plain text '; ` + afQ(70000) + `; printf '\n'; exit 1`, afQRun},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tail := afGateTailOf(t, c.gate)
			if !strings.Contains(tail, c.want) {
				t.Errorf("the gate's last line is missing from the reason (%d bytes): %.120q", len(tail), tail)
			}
			if strings.Contains(strings.ToLower(tail), afMaskd) {
				t.Errorf("nothing was masked, but the reason says so: %.200q", tail)
			}
		})
	}
}

// The marker and the value arrive in separate writes, and the end of what was written is cut by a small window.
func TestAFFailureTail_AMarkerSplitAcrossWritesIsStillSeen(t *testing.T) {
	q := strings.Repeat("Q", 400)
	cases := []struct {
		name   string
		writes []string
		lost   bool
	}{
		{"marker_in_one_write_value_in_the_next", []string{"password=", q}, true},
		{"marker_split_in_the_word", []string{"pa", "ssword=" + q}, true},
		{"marker_split_before_the_separator", []string{"password", "=" + q}, true},
		{"marker_split_by_BEL_across_writes", []string{"pass\a", "word=" + q}, true},
		{"escape_sequence_split_across_writes", []string{"\x1b[3", "1mpassword=" + q}, true},
		{"marker_ends_the_line_value_in_the_next_write", []string{"password=\n", q}, true},
		{"marker_ends_the_line_split_newline", []string{"password=", "\n", q}, true},
		{"value_in_many_small_writes", append([]string{"token: "}, strings.Split(strings.Repeat("Q,", 400), ",")...), true},
		{"marker_after_other_text_on_the_line", []string{"error while reading ", "the password= ", q}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := mergeq.NewTailWriter(128)
			for _, s := range c.writes {
				if _, err := w.Write([]byte(s)); err != nil {
					t.Fatal(err)
				}
			}
			tail := afFailureTail(w)
			if strings.Contains(tail, afQRun) {
				t.Errorf("a piece of the value survives: %.120q", tail)
			}
			if c.lost && !strings.Contains(strings.ToLower(tail), afMaskd) {
				t.Errorf("the reason does not say a value was masked: %.200q", tail)
			}
		})
	}
	t.Run("no_marker_keeps_the_last_line", func(t *testing.T) {
		w := mergeq.NewTailWriter(128)
		for _, s := range []string{"plain text ", q} {
			_, _ = w.Write([]byte(s))
		}
		if tail := afFailureTail(w); !strings.Contains(tail, afQRun) || strings.Contains(strings.ToLower(tail), afMaskd) {
			t.Errorf("a line with no marker must stay as it is: %.120q", tail)
		}
	})
}

// The same shapes through the real verb: the JSON line on stdout and the fail reason in the ledger must agree and
// hold no piece of the value.
func TestAFLand_ALongValueWhoseMarkerFellOffIsNotInTheJSONNorTheLedger(t *testing.T) {
	afSkipWithoutSh(t)
	cases := []struct{ name, gate string }{
		{"no_newline_after_value", `printf 'password='; ` + afQ(70000) + `; exit 1`},
		{"trailing_newline", `printf 'password='; ` + afQ(70000) + `; printf '\n'; exit 1`},
		{"marker_ends_the_previous_line", `printf 'password=\n'; ` + afQ(70000) + `; exit 1`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, repo, _, st, _ := afQueuedCard(t)
			out := runAF(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", c.gate)
			var res struct{ Outcome, Reason string }
			if err := json.Unmarshal([]byte(out), &res); err != nil {
				t.Fatal(err)
			}
			if res.Outcome != "failed" {
				t.Fatalf("outcome %q, want failed", res.Outcome)
			}
			shown, led := afOutputOf(res.Reason), afFailReason(t, st)
			for where, text := range map[string]string{"stdout JSON": shown, "ledger fail reason": led} {
				if strings.Contains(text, afQRun) {
					t.Errorf("%s holds a piece of the value (%d bytes)", where, len(text))
				}
				if !strings.Contains(strings.ToLower(text), afMaskd) {
					t.Errorf("%s does not say a value was masked: %.200q", where, text)
				}
			}
		})
	}
}
