// mcp_jev_catalog_doc_test.go — SPEC-JEV-GOAL-DIST-001 M8a (AC-JEVG-007).
// The tool-catalogue rule exists in TWO copies (the loaded rule and its
// template mirror) and each carries TWO tool-count figures; a tool added to
// the registry must move all four in lockstep, and the copies must stay
// byte-identical. This test is the only mechanical enforcement for that file's
// mirror parity — internal/template/rule_template_mirror_test.go does not
// enumerate it (measured 2026-09-22) — so it byte-compares BOTH catalogue
// files, not just the stub.
package cli

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// jevCatalogueDocPairs maps each catalogue document to its template mirror.
var jevCatalogueDocPairs = [][2]string{
	{
		"../../.claude/rules/moai/core/moai-mcp-tools.md",
		"../../internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md",
	},
	{
		"../../.claude/rules/moai/core/moai-mcp-tools-catalogue.md",
		"../../internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md",
	},
}

// TestMCPToolCatalogueDocsStayMirrorIdentical — a one-sided edit is the
// documented failure mode for these files; the byte compare is what catches it.
func TestMCPToolCatalogueDocsStayMirrorIdentical(t *testing.T) {
	for _, pair := range jevCatalogueDocPairs {
		local, err := os.ReadFile(pair[0])
		if err != nil {
			t.Fatalf("read %s: %v", pair[0], err)
		}
		mirror, err := os.ReadFile(pair[1])
		if err != nil {
			t.Fatalf("read %s: %v", pair[1], err)
		}
		if !bytes.Equal(local, mirror) {
			t.Errorf("%s and its template mirror %s diverged — a one-sided doc edit changes what a user project reads", pair[0], pair[1])
		}
	}
}

// TestMCPToolCatalogueFiguresMatchRegistry — all four figures agree with the
// registered tool set. The totals are read from the docs and compared against
// len(MoaiMCPToolNames()), which the registration-equality guard already binds
// to the live tools/list; the family-coverage figure is checked against the
// registry's own family arithmetic (total minus the session-messaging family,
// which the stub documents as following below the table).
func TestMCPToolCatalogueFiguresMatchRegistry(t *testing.T) {
	names := mcpcat.MoaiMCPToolNames()
	total := len(names)
	sessionMsg := 0
	for _, n := range names {
		if strings.HasPrefix(n, "session_msg_") {
			sessionMsg++
		}
	}
	if sessionMsg == 0 {
		t.Fatal("no session_msg_* tools in the registry — the family arithmetic this test checks is unverifiable")
	}
	familyCovered := total - sessionMsg

	reTotal := regexp.MustCompile(`(\d+) tools exposed by the self-hosted`)
	reFamily := regexp.MustCompile(`Tool families \((\d+) of the (\d+) tools`)

	for _, pair := range jevCatalogueDocPairs[:1] {
		body, err := os.ReadFile(pair[0])
		if err != nil {
			t.Fatalf("read %s: %v", pair[0], err)
		}
		doc := string(body)

		m := reTotal.FindStringSubmatch(doc)
		if m == nil {
			t.Fatalf("%s no longer carries the total-count sentence this test reads", pair[0])
		}
		if got, _ := strconv.Atoi(m[1]); got != total {
			t.Errorf("%s says %q tools exposed; the registered set holds %d", pair[0], m[1], total)
		}

		m = reFamily.FindStringSubmatch(doc)
		if m == nil {
			t.Fatalf("%s no longer carries the family-coverage header this test reads", pair[0])
		}
		if got, _ := strconv.Atoi(m[1]); got != familyCovered {
			t.Errorf("%s family header covers %s tools; the registry's table-covered set is %d (total %d minus session-messaging %d)", pair[0], m[1], familyCovered, total, sessionMsg)
		}
		if got, _ := strconv.Atoi(m[2]); got != total {
			t.Errorf("%s family header says %q of its total; the registered set holds %d", pair[0], m[2], total)
		}
	}

	// The catalogue companion states its total in several places (description,
	// intro, section header). EVERY count figure it carries must equal the
	// registered total — "three of four correct reads as done" is the
	// documented failure mode for these files.
	reToolCount := regexp.MustCompile(`(\d+)-tool|\((\d+) tools\)|of the (\d+) tools`)
	for _, pair := range jevCatalogueDocPairs[1:] {
		body, err := os.ReadFile(pair[0])
		if err != nil {
			t.Fatalf("read %s: %v", pair[0], err)
		}
		matches := reToolCount.FindAllStringSubmatch(string(body), -1)
		if len(matches) == 0 {
			t.Fatalf("%s carries no count figure — the probe, not the doc, is broken", pair[0])
		}
		for _, m := range matches {
			for _, capIdx := range []int{1, 2, 3} {
				if m[capIdx] == "" {
					continue
				}
				if got, _ := strconv.Atoi(m[capIdx]); got != total {
					t.Errorf("%s carries count figure %q; the registered set holds %d", pair[0], m[capIdx], total)
				}
			}
		}
	}
}
