package quality

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSolutionStyleTsconfigWithComments is the regression the review caught.
//
// tsconfig.json is JSONC and real configs carry comments. Before the strip,
// encoding/json failed on them, isSolutionStyleTsconfig answered false, and the
// config was routed to "run tsc" — a vacuous pass, which is the exact hole the
// solution-style check exists to close.
func TestSolutionStyleTsconfigWithComments(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json": `{"scripts":{"build":"tsc -b"}}`,
		"tsconfig.json": `{
  // Solution file: compiles nothing itself, only references projects.
  "files": [],
  /* the workspaces this solution builds */
  "references": [
    { "path": "./packages/a" },
    { "path": "./packages/b" },
  ],
}`,
	})

	_, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if ok {
		t.Fatal("commented solution-style tsconfig accepted; it passes vacuously")
	}
	if !strings.Contains(reason, "solution-style") {
		t.Errorf("reason %q does not name the solution-style case", reason)
	}
}

// TestCommentedNonSolutionTsconfigStillRuns guards the other direction: comments
// must not turn an ordinary config into a refusal.
func TestCommentedNonSolutionTsconfigStillRuns(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json": `{"scripts":{"build":"tsc"}}`,
		"tsconfig.json": `{
  // strict everywhere
  "compilerOptions": { "strict": true },
}`,
	})

	step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if !ok {
		t.Fatalf("skipped (%s); a commented ordinary tsconfig must still run", reason)
	}
	if step.binary != "npx" {
		t.Errorf("binary = %q, want npx", step.binary)
	}
}

func TestStripJSONC(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want map[string]any
	}{
		{
			name: "line comment",
			in:   "{\n // c\n \"a\": 1\n}",
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "block comment",
			in:   `{/* c */"a": 1}`,
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "trailing comma in object",
			in:   `{"a": 1,}`,
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "trailing comma in array",
			in:   `{"a": [1,2,]}`,
			want: map[string]any{"a": []any{float64(1), float64(2)}},
		},
		{
			// The load-bearing case: a // inside a string is data, not a comment.
			name: "double slash inside a string survives",
			in:   `{"url": "https://example.com/x"}`,
			want: map[string]any{"url": "https://example.com/x"},
		},
		{
			name: "escaped quote does not end the string early",
			in:   `{"a": "say \"hi\" // not a comment"}`,
			want: map[string]any{"a": `say "hi" // not a comment`},
		},
		{
			name: "comma inside a string is not a trailing comma",
			in:   `{"a": "x,"}`,
			want: map[string]any{"a": "x,"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got map[string]any
			if err := json.Unmarshal(stripJSONC([]byte(tc.in)), &got); err != nil {
				t.Fatalf("unmarshal after strip: %v (stripped: %s)", err, stripJSONC([]byte(tc.in)))
			}

			wantJSON, _ := json.Marshal(tc.want)
			gotJSON, _ := json.Marshal(got)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("got %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

// TestStripJSONCLeavesPlainJSONAlone asserts the strip is a no-op on input that
// was already valid JSON — the common case, which must not be perturbed.
func TestStripJSONCLeavesPlainJSONAlone(t *testing.T) {
	t.Parallel()

	const in = `{"compilerOptions":{"strict":true},"include":["src"]}`
	if got := string(stripJSONC([]byte(in))); got != in {
		t.Fatalf("plain JSON was altered:\n got %s\nwant %s", got, in)
	}
}
