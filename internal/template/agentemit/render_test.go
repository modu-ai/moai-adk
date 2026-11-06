// render_test.go — SPEC-AGENT-CORE4-001 REQ-AC4-002/003/004: both renderers
// read the agent source and keep or drop marked blocks per harness, and a
// source the renderers cannot carry fails emission naming the file and the
// token or marker. Block behavior is exercised with fixture sources because
// none of the shipped agent bodies carries a marker.
package agentemit_test

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

const (
	claudeOnlyLine = "CLAUDE-ONLY-LINE"
	codexOnlyLine  = "CODEX-ONLY-LINE"
)

// markedBody carries one block per harness between shared lines.
const markedBody = "\n# marked\n\nShared line.\n" +
	"<!-- claude-only -->\n" + claudeOnlyLine + "\n<!-- /claude-only -->\n" +
	"<!-- codex-only -->\n" + codexOnlyLine + "\n<!-- /codex-only -->\n" +
	"Tail line.\n"

// markedSource returns a one-agent source set whose body is body.
func markedSource(body string) fstest.MapFS {
	return fstest.MapFS{
		"agents/marked.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"marked", "Marked agent.\n", "Read, Grep, Glob", "inherit", "low", nil, body))},
	}
}

func emitMarked(t *testing.T, body string) *agentemit.Publication {
	t.Helper()
	pub, err := agentemit.EmitAll(markedSource(body), "agents", mustManifest(t))
	if err != nil {
		t.Fatalf("EmitAll: %v", err)
	}
	return pub
}

// codexInstructions decodes the emitted TOML of the marked agent and returns
// its developer_instructions.
func codexInstructions(t *testing.T, pub *agentemit.Publication) string {
	t.Helper()
	data, ok := pub.CodexTOML[".codex/agents/moai/marked.toml"]
	if !ok {
		t.Fatalf("no TOML emitted for the marked agent: %v", keysOf(pub.CodexTOML))
	}
	doc, err := decodeTOML(string(data))
	if err != nil {
		t.Fatalf("decode emitted TOML: %v", err)
	}
	body, _ := doc["developer_instructions"].(string)
	return body
}

func TestClaudeRenderFromSource(t *testing.T) {
	t.Run("claude_only_block_present_in_c2", func(t *testing.T) {
		md := string(emitMarked(t, markedBody).Markdown["agents/marked.md"])
		if !strings.Contains(md, claudeOnlyLine) {
			t.Errorf("claude-only block missing from the Claude render:\n%s", md)
		}
		if strings.Contains(md, "claude-only") || strings.Contains(md, "codex-only") {
			t.Errorf("block marker lines must not survive the render:\n%s", md)
		}
	})
	t.Run("codex_only_block_absent_from_c2", func(t *testing.T) {
		md := string(emitMarked(t, markedBody).Markdown["agents/marked.md"])
		if strings.Contains(md, codexOnlyLine) {
			t.Errorf("codex-only block leaked into the Claude render:\n%s", md)
		}
		wantBody := "\n# marked\n\nShared line.\n" + claudeOnlyLine + "\nTail line.\n"
		if !strings.HasSuffix(md, "---\n"+wantBody) {
			t.Errorf("Claude render body differs from the shared lines plus the claude-only block:\n%s", md)
		}
	})
	t.Run("render_is_deterministic", func(t *testing.T) {
		a := emitMarked(t, markedBody).Markdown["agents/marked.md"]
		b := emitMarked(t, markedBody).Markdown["agents/marked.md"]
		if string(a) != string(b) {
			t.Error("two renders of the same source differ")
		}
	})
}

func TestCodexRenderFromSource(t *testing.T) {
	t.Run("codex_only_block_present_in_c3", func(t *testing.T) {
		body := codexInstructions(t, emitMarked(t, markedBody))
		if !strings.Contains(body, codexOnlyLine) {
			t.Errorf("codex-only block missing from the Codex render:\n%s", body)
		}
		if strings.Contains(body, "claude-only") || strings.Contains(body, "codex-only") {
			t.Errorf("block marker lines must not survive the render:\n%s", body)
		}
	})
	t.Run("claude_only_block_absent_from_c3", func(t *testing.T) {
		body := codexInstructions(t, emitMarked(t, markedBody))
		if strings.Contains(body, claudeOnlyLine) {
			t.Errorf("claude-only block leaked into the Codex render:\n%s", body)
		}
		want := "\n# marked\n\nShared line.\n" + codexOnlyLine + "\nTail line.\n"
		if body != want {
			t.Errorf("Codex body = %q, want the shared lines plus the codex-only block %q", body, want)
		}
	})
	t.Run("c3_rendered_without_c2_bytes", func(t *testing.T) {
		// The claude-only block holds the TOML literal-string delimiter. A Codex
		// render derived from the Claude bytes would carry it and be refused;
		// one rendered from the source drops the block first and succeeds.
		body := "\n# marked\n<!-- claude-only -->\nunrepresentable ''' in a literal string\n<!-- /claude-only -->\nTail.\n"
		pub := emitMarked(t, body)
		if !strings.Contains(string(pub.Markdown["agents/marked.md"]), "'''") {
			t.Error("the Claude render must keep the claude-only block")
		}
		if got := codexInstructions(t, pub); strings.Contains(got, "'''") || got != "\n# marked\nTail.\n" {
			t.Errorf("Codex body = %q, want the claude-only block dropped", got)
		}
	})
	t.Run("contract_rows_unchanged_for_twelve", func(t *testing.T) {
		// Compares the permission-contract verdicts emitted for every shipped
		// source with the in-repo contract table of agents-codex.yaml: each
		// verdict carries exactly its table row. The check follows the table
		// and the source set, not a role count.
		man := mustManifest(t)
		report, err := agentemit.BuildPermissionReport(os.DirFS(repoRoot), sourceRoot, man)
		if err != nil {
			t.Fatalf("BuildPermissionReport: %v", err)
		}
		if len(report.Verdicts) == 0 {
			t.Fatal("permission report is empty")
		}
		roles := map[string]bool{}
		for _, v := range report.Verdicts {
			roles[v.Role] = true
			row, ok := mappingFor(t, man, v.Axis, v.Kind)
			if !ok {
				t.Errorf("%s/%s/%s: verdict with no contract row", v.Role, v.Axis, v.Kind)
				continue
			}
			if v.Mapping != row.Mapping || v.Field != row.Field || v.Basis != row.Basis || v.Reason != row.Reason {
				t.Errorf("%s/%s/%s: verdict %+v differs from contract row %+v", v.Role, v.Axis, v.Kind, v, row)
			}
		}
		entries, err := os.ReadDir(repoRoot + "/" + sourceRoot)
		if err != nil {
			t.Fatalf("read source directory: %v", err)
		}
		sources := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				sources++
			}
		}
		if len(roles) != sources {
			t.Errorf("report covers %d roles for %d sources", len(roles), sources)
		}
	})
}

func TestRenderUnmappedCapabilityFails(t *testing.T) {
	man := mustManifest(t)
	t.Run("unmapped_tools_token_names_file_and_token", func(t *testing.T) {
		set := fstest.MapFS{"agents/bad.md": &fstest.MapFile{Data: []byte(fixtureMD(
			"bad", "Bad.\n", "Read, Bogus", "inherit", "low", nil, "\nbody\n"))}}
		_, err := agentemit.EmitAll(set, "agents", man)
		if err == nil || !strings.Contains(err.Error(), "agents/bad.md") || !strings.Contains(err.Error(), "Bogus") {
			t.Errorf("want an error naming agents/bad.md and Bogus, got %v", err)
		}
	})
	// A claude-only block may sit inside the frontmatter: the Claude view then
	// carries frontmatter the Codex view never sees. Both views are validated, so
	// an unmapped token or broken YAML there fails closed instead of shipping to
	// the Claude publication (REQ-AC4-004, regardless of which view carries it).
	frontmatterWithClaudeOnly := func(claudeLines string) string {
		return "---\nname: marked\ndescription: |\n  Marked.\n" +
			"<!-- claude-only -->\n" + claudeLines + "<!-- /claude-only -->\n" +
			"<!-- codex-only -->\ntools: Read\n<!-- /codex-only -->\n" +
			"model: inherit\neffort: low\n---\n\nbody\n"
	}
	t.Run("claude_view_unmapped_tools_token_names_file_and_token", func(t *testing.T) {
		set := fstest.MapFS{"agents/marked.md": &fstest.MapFile{Data: []byte(frontmatterWithClaudeOnly("tools: Read, Bogus\n"))}}
		pub, err := agentemit.EmitAll(set, "agents", man)
		if err == nil || !strings.Contains(err.Error(), "agents/marked.md") || !strings.Contains(err.Error(), "Bogus") {
			t.Errorf("want an error naming agents/marked.md and Bogus, got %v", err)
		}
		if pub != nil {
			t.Error("a failed emission must return no publication")
		}
	})
	t.Run("claude_view_broken_yaml_names_file", func(t *testing.T) {
		set := fstest.MapFS{"agents/marked.md": &fstest.MapFile{Data: []byte(frontmatterWithClaudeOnly("tools: Read\nskills: [unterminated\n"))}}
		pub, err := agentemit.EmitAll(set, "agents", man)
		if err == nil || !strings.Contains(err.Error(), "agents/marked.md") || !strings.Contains(err.Error(), "frontmatter parse") {
			t.Errorf("want a frontmatter parse error naming agents/marked.md, got %v", err)
		}
		if pub != nil {
			t.Error("a failed emission must return no publication")
		}
	})
	t.Run("claude_view_defect_leaves_no_output_and_check_fails", func(t *testing.T) {
		root, lay := fixtureRoot(t, map[string]string{"marked.md": frontmatterWithClaudeOnly("tools: Read, Bogus\n")})
		lay.Targets = lay.Targets[:2]
		if _, err := agentemit.Update(root, lay, man); err == nil || !strings.Contains(err.Error(), "Bogus") {
			t.Errorf("Update must fail naming the token, got %v", err)
		}
		if _, err := agentemit.Check(root, lay, man); err == nil || !strings.Contains(err.Error(), "Bogus") {
			t.Errorf("Check must fail naming the token, got %v", err)
		}
		for _, tg := range lay.Targets {
			if _, err := os.Stat(root + "/" + tg.Dir); !os.IsNotExist(err) {
				t.Errorf("%s: a failed run left output behind (stat err %v)", tg.Name, err)
			}
		}
	})
	t.Run("unterminated_marker_names_file", func(t *testing.T) {
		_, err := agentemit.EmitAll(markedSource("\nShared.\n<!-- claude-only -->\nnever closed\n"), "agents", man)
		if err == nil || !strings.Contains(err.Error(), "agents/marked.md") || !strings.Contains(err.Error(), "claude-only") {
			t.Errorf("want an error naming agents/marked.md and the claude-only marker, got %v", err)
		}
	})
	t.Run("nested_marker_names_file", func(t *testing.T) {
		body := "\n<!-- claude-only -->\n<!-- codex-only -->\ninner\n<!-- /codex-only -->\n<!-- /claude-only -->\n"
		_, err := agentemit.EmitAll(markedSource(body), "agents", man)
		if err == nil || !strings.Contains(err.Error(), "agents/marked.md") || !strings.Contains(err.Error(), "codex-only") {
			t.Errorf("want an error naming agents/marked.md and the nested codex-only marker, got %v", err)
		}
	})
	t.Run("no_partial_output", func(t *testing.T) {
		set := fstest.MapFS{
			"agents/good.md": &fstest.MapFile{Data: []byte(fixtureMD("good", "Good.\n", "Read", "inherit", "low", nil, "\nbody\n"))},
			"agents/bad.md":  &fstest.MapFile{Data: []byte(fixtureMD("bad", "Bad.\n", "Read, Bogus", "inherit", "low", nil, "\nbody\n"))},
		}
		pub, err := agentemit.EmitAll(set, "agents", man)
		if err == nil || pub != nil {
			t.Fatalf("want (nil, error), got (%v, %v)", pub, err)
		}

		goodOnly := map[string]string{"good.md": fixtureMD("good", "Good.\n", "Read", "inherit", "low", nil, "\nbody\n")}
		// Positive control: with only the good source the same layout does write
		// its directories, so an absent directory below is not an Update that
		// never writes.
		goodRoot, goodLay := fixtureRoot(t, goodOnly)
		goodLay.Targets = goodLay.Targets[:2]
		if _, err := agentemit.Update(goodRoot, goodLay, man); err != nil {
			t.Fatalf("Update over the good source alone: %v", err)
		}
		if _, err := os.Stat(goodRoot + "/out/c2/good.md"); err != nil {
			t.Fatalf("positive control wrote no C2 file: %v", err)
		}

		withBad := map[string]string{
			"good.md": goodOnly["good.md"],
			"bad.md":  fixtureMD("bad", "Bad.\n", "Read, Bogus", "inherit", "low", nil, "\nbody\n"),
		}
		root, lay := fixtureRoot(t, withBad)
		lay.Targets = lay.Targets[:2]
		if _, err := agentemit.Update(root, lay, man); err == nil || !strings.Contains(err.Error(), "Bogus") {
			t.Fatalf("Update must fail on the unmappable source naming its token, got %v", err)
		}
		for _, tg := range lay.Targets {
			if _, err := os.Stat(root + "/" + tg.Dir); !os.IsNotExist(err) {
				t.Errorf("%s: a failed run left output behind (stat err %v)", tg.Name, err)
			}
		}
	})
}

// TestRenderViewMarkerRules pins the marker grammar: markers sit on their own
// lines, blocks neither nest nor straddle, and a source without markers is
// carried byte for byte into both views.
func TestRenderViewMarkerRules(t *testing.T) {
	plain := "---\nname: x\n---\nline one\n  indented <!-- claude-only --> inline mention stays text\n"
	for _, v := range []agentemit.View{agentemit.ViewClaude, agentemit.ViewCodex} {
		got, err := agentemit.RenderView("x.md", []byte(plain), v)
		if err != nil || string(got) != plain {
			t.Errorf("view %d: marker-free source must render to itself, got %q, %v", v, got, err)
		}
	}

	bad := map[string]struct{ src, want string }{
		"stray close":           {"a\n<!-- /claude-only -->\n", "without an open block"},
		"mismatched close":      {"<!-- claude-only -->\nx\n<!-- /codex-only -->\n", "does not match"},
		"unterminated codex":    {"a\n<!-- codex-only -->\nx\n", "unterminated"},
		"nested same":           {"<!-- codex-only -->\n<!-- codex-only -->\n", "nested"},
		"nested different":      {"<!-- claude-only -->\n<!-- codex-only -->\n", "nested"},
		"indented marker works": {"  <!-- claude-only -->\n", "unterminated"},
	}
	for label, tc := range bad {
		for _, v := range []agentemit.View{agentemit.ViewClaude, agentemit.ViewCodex} {
			_, err := agentemit.RenderView("bad.md", []byte(tc.src), v)
			if err == nil || !strings.Contains(err.Error(), "bad.md") || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%s (view %d): want an error naming bad.md and %q, got %v", label, v, tc.want, err)
			}
		}
	}
}
