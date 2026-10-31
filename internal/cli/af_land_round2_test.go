package cli

// N1 round 2: seven shapes whose marker falls outside the kept window and which round 1 did not recognise as marked
// lines (a bare Bearer credential, a ghp_ and an sk- token, a JSON "password" field, a PEM block with no END line, a
// marker followed by a blank line, a marker split by 600 control bytes).
//
// The value is a run of Q (no digit, so the opaque-run rule never hides it): only marker-based masking can.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

// The PEM lines are assembled so that no whole header sits in this source file.
const (
	afPEMBegin = "-----BEGIN " + "RSA PRIVATE" + " KEY-----"
	afPEMEnd   = "-----END " + "RSA PRIVATE" + " KEY-----"
)

// afBody is a shell command that prints about 70 KB of 64-Q lines.
const afBody = `yes QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ | head -c 70000`

// afRound2Shapes are the leader's seven shapes; the value is about 70 KB and the marker is outside the window.
var afRound2Shapes = []struct{ name, gate string }{
	{"bearer", `printf 'Bearer '; ` + afQ(70000) + `; exit 1`},
	{"ghp_token", `printf 'ghp_'; ` + afQ(70000) + `; exit 1`},
	{"sk_token", `printf 'sk-'; ` + afQ(70000) + `; exit 1`},
	{"json_password", `printf '{"password": "'; ` + afQ(70000) + `; printf '"}'; exit 1`},
	{"pem_without_end", `printf -- '` + afPEMBegin + `\n'; ` + afBody + `; exit 1`},
	{"marker_then_blank_line", `printf 'password=\n\n'; ` + afQ(70000) + `; exit 1`},
	{"marker_split_by_600_BEL", `printf 'pass'; head -c 600 /dev/zero | tr '\0' '\007'; printf 'word='; ` + afQ(70000) + `; exit 1`},
}

func TestAFGate_ARound2ShapesWhoseMarkerFellOffAreMasked(t *testing.T) {
	afSkipWithoutSh(t)
	for _, c := range afRound2Shapes {
		t.Run(c.name, func(t *testing.T) {
			tail := afGateTailOf(t, c.gate)
			if strings.Contains(tail, afQRun) {
				t.Errorf("a piece of the value survives (%d bytes of tail): %.80q", len(tail), tail)
			}
			if !strings.Contains(strings.ToLower(tail), afMaskd) {
				t.Errorf("the value was masked for want of its marker but the reason does not say so: %.200q", tail)
			}
			if len(tail) > afGateTail {
				t.Errorf("the tail is %d bytes, over the %d the reason may hold", len(tail), afGateTail)
			}
		})
	}
	// variations that must keep their other lines
	t.Run("pem_with_end_keeps_the_lines_after_it", func(t *testing.T) {
		tail := afGateTailOf(t, `printf -- '`+afPEMBegin+`\n'; `+afBody+`; printf '\n`+afPEMEnd+`\nafter the key\n'; exit 1`)
		if strings.Contains(tail, afQRun) || !strings.Contains(tail, "after the key") {
			t.Errorf("want the body masked and the line after the END kept: %.200q", tail)
		}
	})
	t.Run("several_blank_lines_after_the_marker", func(t *testing.T) {
		tail := afGateTailOf(t, `printf 'password=\n \n\t\n\n'; `+afQ(70000)+`; exit 1`)
		if strings.Contains(tail, afQRun) {
			t.Errorf("a piece of the value survives: %.80q", tail)
		}
	})
	t.Run("only_the_first_content_line_after_a_marker_is_its_value", func(t *testing.T) {
		tail := afGateTailOf(t, `printf 'password=\nthe value\n'; `+afQ(70000)+`; exit 1`)
		if !strings.Contains(tail, afQRun) || strings.Contains(strings.ToLower(tail), afMaskd) {
			t.Errorf("the second line after the marker is plain output and must stay: %.80q", tail)
		}
	})
}

// The same shapes with the marker and the value in separate writes, the window a small one.
func TestAFFailureTail_ARound2ShapesSplitAcrossWrites(t *testing.T) {
	q := strings.Repeat("Q", 400)
	body := strings.Repeat(strings.Repeat("Q", 64)+"\n", 8)
	cases := []struct {
		name   string
		writes []string
	}{
		{"bearer", []string{"Bea", "rer ", q}},
		{"ghp_token", []string{"gh", "p_", q}},
		{"sk_token", []string{"s", "k-", q}},
		{"json_password", []string{`{"pass`, `word": "`, q, `"}`}},
		{"json_value_closes_in_a_later_write", []string{`{"api_key": "`, q, q, `"`}},
		{"pem_header_split", []string{afPEMBegin[:20], afPEMBegin[20:] + "\n", body}},
		{"pem_body_lines_in_separate_writes", []string{afPEMBegin + "\n", body, body}},
		{"marker_then_blank_lines", []string{"password=", "\n", "\n", " \n", q}},
		{"marker_then_blank_line_in_one_write", []string{"password=\n\n" + q}},
		{"BEL_run_between_the_halves", []string{"pass", strings.Repeat("\a", 600), "word=", q}},
		{"BEL_run_inside_one_write", []string{"pass" + strings.Repeat("\a", 600) + "word=" + q}},
		{"escape_sequence_split_between_writes", []string{"pass\x1b[3", "1mword=", q}},
		{"OSC_sequence_split_between_writes", []string{"pass\x1b]0;ti", "tle\aword=", q}},
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
			if !strings.Contains(strings.ToLower(tail), afMaskd) {
				t.Errorf("the reason does not say a value was masked: %.200q", tail)
			}
		})
	}
}

// The seven shapes through the real verb: stdout JSON and the ledger fail reason must hold no byte of the value.
func TestAFLand_ARound2ShapesAreNotInTheJSONNorTheLedger(t *testing.T) {
	afSkipWithoutSh(t)
	for _, c := range afRound2Shapes {
		t.Run(c.name, func(t *testing.T) {
			db, repo, _, st, _ := afQueuedCard(t)
			outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", c.gate)
			if outcome != "failed" {
				t.Fatalf("outcome %q, want failed", outcome)
			}
			shown, led := afOutputOf(reason), afFailReason(t, st)
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
