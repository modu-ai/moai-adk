package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestMXRawViewRendered verifies SPEC-WEB-CONSOLE-014 M4 (REQ-WC14-030,
// AC-WC14-030b): the GET / response renders the mx raw-view group. mx is a
// raw-only section (0 editable fields — mx stays RouteExcluded), so the render
// path exercises the generic fieldsetSchemaSection with only RawViewBlocks:
// the section header (sec.mx.title), the mx panel container (data-panel="mx"),
// and the two raw-block code chips (mx.danger_categories / mx.test_paths) are
// present in the rendered body.
func TestMXRawViewRendered(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	for _, want := range []string{
		`data-i18n="sec.mx.title"`,
		`data-panel="mx"`,
		`mx.danger_categories`,
		`mx.test_paths`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing mx raw-view marker %q", want)
		}
	}
}

// spec014NewI18nKeys는 SPEC-WEB-CONSOLE-014가 도입한 신규 i18n 키 집합이다
// (M2 read-only/raw honest label + M3 merge_method + M4 mx section). 각 키는
// 4-locale 전부에 존재해야 한다 (REQ-WC14-060).
var spec014NewI18nKeys = []string{
	// M2 — honest label note keys.
	"ro.note.governance", "ro.note.dead_config", "raw.note.informational",
	// M4 — mx raw-only section header.
	"sec.mx.title", "sec.mx.desc",
	// M3 — merge_method options + per-profile title/desc.
	"f.git_strategy.merge_method.opt.merge",
	"f.git_strategy.merge_method.opt.rebase",
	"f.git_strategy.merge_method.opt.squash",
	"f.git_strategy.manual.merge_method.title", "f.git_strategy.manual.merge_method.desc",
	"f.git_strategy.personal.merge_method.title", "f.git_strategy.personal.merge_method.desc",
	"f.git_strategy.team.merge_method.title", "f.git_strategy.team.merge_method.desc",
}

// TestSPEC014I18nKeysFourLocale verifies REQ-WC14-060: every new user-facing
// label introduced by SPEC-WEB-CONSOLE-014 has a key in all four locale blocks
// (en/ko/ja/zh) of i18n.js — not just the EN block that TestDataI18nKeysSubset-
// OfDictionary asserts. It slices the dictionary into per-locale blocks and
// checks each key is present in each block (AC-WC14-060).
func TestSPEC014I18nKeysFourLocale(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")

	// File-order locale block starts (en < ko < ja < zh).
	order := []string{"en", "ko", "ja", "zh"}
	starts := make([]int, len(order))
	for i, loc := range order {
		idx := strings.Index(dict, "\n  "+loc+": {")
		if idx < 0 {
			t.Fatalf("locale block %q not found in i18n.js", loc)
		}
		starts[i] = idx
	}
	blocks := map[string]string{}
	for i, loc := range order {
		end := len(dict)
		if i+1 < len(order) {
			end = starts[i+1]
		}
		blocks[loc] = dict[starts[i]:end]
	}

	for _, loc := range order {
		for _, k := range spec014NewI18nKeys {
			if !strings.Contains(blocks[loc], `"`+k+`":`) {
				t.Errorf("locale %q missing SPEC-WEB-CONSOLE-014 i18n key %q (REQ-WC14-060 4-locale)", loc, k)
			}
		}
	}
}
