package cli

// Parity of the gate-output writer's marker detector with core.Redact. For each prefix, a long value is put after
// it: when Redact masks the value, the writer must have marked the line (so that a cut which drops the prefix still
// masks the rest); when Redact leaves the value alone the writer must leave the line unmarked (so a loud gate's last
// line is still shown). closer is what ends the value for the rules that need an end (a JSON closing quote, the @ of
// a URL); it may arrive in a later write.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

type afParityRow struct {
	prefix, closer string
	secret         bool // what Redact is expected to do: a changed table is a visible decision
}

func afParityRows() []afParityRow {
	var rows []afParityRow
	for _, name := range []string{"pass" + "word", "PASS" + "WORD", "db_pass" + "word", "pass" + "wd", "pass" + "phrase",
		"api_key", "Api-Key", "apikey", "x-auth-token", "client_secret", "credential", "Cookie", "Set-Cookie"} {
		for _, sep := range []string{"=", ": ", " = "} {
			rows = append(rows, afParityRow{prefix: name + sep, secret: true})
		}
	}
	// the bare `key` rule takes only a directly attached `=`
	rows = append(rows, afParityRow{prefix: "monkey_key=", secret: true}, afParityRow{prefix: "monkey_key: "})
	// the value on a later line: a bare scheme word, a name with its separator at the end of the line, blank lines
	for _, p := range []string{"Bearer\n", "Authorization: Bearer\n", "pass" + "word=\n", "api_key:\n\n \n", "Cookie:\n"} {
		rows = append(rows, afParityRow{prefix: p, secret: true})
	}
	for _, p := range []string{"Authorization: Bearer ", "Authorization: Basic ", "Authorization: Token ", "Bearer ", "bearer ",
		"gh" + "p_", "gh" + "o_", "gh" + "u_", "gh" + "s_", "gh" + "r_", "github" + "_pat_", "s" + "k-", afPEMBegin + "\n"} {
		rows = append(rows, afParityRow{prefix: p, secret: true})
	}
	for _, name := range []string{"pass" + "word", "api_key", "client_token"} {
		rows = append(rows, afParityRow{prefix: `{"` + name + `": "`, closer: `"`, secret: true})
		rows = append(rows, afParityRow{prefix: `{"` + name + `":"`, closer: `"`, secret: true})
	}
	rows = append(rows, afParityRow{prefix: "https://user:", closer: "@host.example", secret: true},
		afParityRow{prefix: "http://u:", closer: "@h.example", secret: true})
	for _, p := range []string{"status: ", "count=", "path ", "hello ", "keyboard layout: ", "passed: ", "monkey ", "see key ", "ask ",
		"https://example.test/", "name = ", "tokens "} {
		rows = append(rows, afParityRow{prefix: p})
	}
	return append(rows, afParityRow{prefix: `{"name": "`, closer: `"`})
}

func TestAFWatch_TheWritersMarkerDetectorAgreesWithRedact(t *testing.T) {
	v := strings.Repeat("Q", 64)
	for _, r := range afParityRows() {
		t.Run(strings.TrimSpace(r.prefix), func(t *testing.T) {
			masked, _ := core.Redact(r.prefix + v + r.closer)
			redactMasks := !strings.Contains(masked, afQRun)
			if redactMasks != r.secret {
				t.Fatalf("core.Redact masks=%v for %q but the table says secret=%v: update the table on purpose", redactMasks, r.prefix, r.secret)
			}
			// the whole line in one write, then with the closer in a later write
			for _, split := range []bool{false, true} {
				w := mergeq.NewTailWriter(16)
				writes := []string{r.prefix + v + r.closer}
				if split {
					writes = []string{r.prefix + v, r.closer}
				}
				for _, s := range writes {
					if _, err := w.Write([]byte(s)); err != nil {
						t.Fatal(err)
					}
				}
				if marked := w.MaskedPrefix() > 0; marked != redactMasks {
					t.Errorf("split=%v: the writer marks %q = %v but core.Redact masks it = %v", split, r.prefix, marked, redactMasks)
				}
			}
		})
	}
}
