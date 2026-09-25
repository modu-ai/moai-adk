// mcp_project_root_doc_test.go — t192 (#e2e F-1): the rule file that tells
// agents which MCP tools accept `project_root` drifted from the server, naming
// five tools while six declare the input (`glm_audit` was missing). The
// omission mattered: glm_audit is where the server is told which tree to DIFF
// before sending that diff to z.ai — GLM itself has no filesystem — so a
// reader working in a worktree learned no way to have its tree reviewed.
//
// A doc claim about a machine-readable surface should be checked against that
// surface rather than re-read by hand, which is what this test does — for the
// local rule file and its template mirror alike.
package cli

import (
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// projectRootDocFiles are the two copies of the rule: the one this repository
// loads, and the mirror shipped to every project by `moai init` / `moai update`.
var projectRootDocFiles = []string{
	"../../.claude/rules/moai/core/moai-mcp-tools.md",
	"../../internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md",
}

// The docs site repeats the project_root tool inventory in four languages.
// Keep each list tied to tools/list rather than a hand-maintained count.
var docsSiteProjectRootLine = regexp.MustCompile(`(?m)^## [^\n]*project_root[^\n]*\n\n([^\n]+)`)

func TestDocsSiteProjectRootMatchesServer(t *testing.T) {
	declared := toolsDeclaringProjectRoot(t)
	locales := map[string]string{
		"en": "Thirteen tools",
		"ko": "13개 도구",
		"ja": "13個のツール",
		"zh": "13 个工具",
	}
	for locale, countPhrase := range locales {
		t.Run(locale, func(t *testing.T) {
			path := "../../docs-site/content/" + locale + "/guides/mcp-server.md"
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			match := docsSiteProjectRootLine.FindStringSubmatch(string(body))
			if match == nil {
				t.Fatal("project_root opening paragraph missing")
			}
			line := match[1]
			if !strings.Contains(line, countPhrase) {
				t.Errorf("count phrase %q missing from %s", countPhrase, line)
			}
			_, list, found := strings.Cut(line, ":")
			if !found {
				_, list, found = strings.Cut(line, "：")
			}
			if !found {
				t.Fatal("tool list separator missing")
			}
			list = strings.SplitN(list, ".", 2)[0]
			list = strings.SplitN(list, "。", 2)[0]
			matches := projectRootDocToolName.FindAllStringSubmatch(list, -1)
			var named []string
			for _, match := range matches {
				named = append(named, match[1])
			}
			sort.Strings(named)
			if strings.Join(named, ",") != strings.Join(declared, ",") {
				t.Errorf("docs list %v; tools/list declares %v", named, declared)
			}
		})
	}
}

// projectRootDocSentence captures the enumerating sentence and its count word.
var projectRootDocSentence = regexp.MustCompile(
	`(?m)^(\w+) tools accept an optional ` + "`project_root`" + ` string:((?s).*?)\. It names the tree`)

// projectRootDocToolName pulls each backticked tool name out of that sentence.
var projectRootDocToolName = regexp.MustCompile("`([a-z_]+)`")

// docCountWords maps the spelled-out count the sentence uses. Only the range a
// tool list plausibly occupies is needed; an unmapped word fails loudly.
var docCountWords = map[string]int{
	"Three": 3, "Four": 4, "Five": 5, "Six": 6, "Seven": 7, "Eight": 8, "Nine": 9, "Ten": 10,
	"Eleven": 11, "Twelve": 12, "Thirteen": 13,
}

// TestProjectRootDocMatchesServer asserts the rule file names exactly the tools
// whose live input schema declares project_root, and that its spelled-out count
// matches how many it names.
func TestProjectRootDocMatchesServer(t *testing.T) {
	declared := toolsDeclaringProjectRoot(t)
	if len(declared) == 0 {
		t.Fatal("no tool declares project_root — the probe, not the doc, is broken")
	}

	for _, path := range projectRootDocFiles {
		t.Run(path, func(t *testing.T) {
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			m := projectRootDocSentence.FindStringSubmatch(string(body))
			if m == nil {
				t.Fatalf("%s no longer carries the enumerating sentence this test checks", path)
			}

			countWord, listed := m[1], m[2]
			named := projectRootDocToolName.FindAllStringSubmatch(listed, -1)
			docTools := make([]string, 0, len(named))
			for _, n := range named {
				docTools = append(docTools, n[1])
			}
			sort.Strings(docTools)

			if strings.Join(docTools, ",") != strings.Join(declared, ",") {
				t.Errorf("%s names %v; the server declares project_root on %v", path, docTools, declared)
			}
			want, ok := docCountWords[countWord]
			if !ok {
				t.Fatalf("%s uses count word %q, which this test cannot read", path, countWord)
			}
			if want != len(docTools) {
				t.Errorf("%s says %q but names %d tools", path, countWord, len(docTools))
			}
		})
	}
}

// toolsDeclaringProjectRoot lists, sorted, every registered tool whose input
// schema carries the project_root argument.
func toolsDeclaringProjectRoot(t *testing.T) []string {
	t.Helper()

	srv := newMoaiMCPServer()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)

	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	// The declaration is the properties key, not the text of the schema: a
	// description mentioning project_root would otherwise read as a
	// declaration and fail the doc for a tool that does not take the input.
	var names []string
	for _, tool := range res.Tools {
		if _, declared := tool.InputSchema.Properties[projectRootArg]; declared {
			names = append(names, tool.Name)
		}
	}
	sort.Strings(names)
	return names
}
