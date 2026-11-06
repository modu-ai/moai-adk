package config

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const afsRefChildEnv = "MOAI_AFS_REF_CHILD_ROOT"

// afsWriteRoutesRaw writes a valid document whose webhook.routes value is the
// given raw YAML lines (relative to the routes key's indentation).
func afsWriteRoutesRaw(t *testing.T, entries []string) string {
	t.Helper()
	d := afsValidDoc()
	afsSet(d, "webhook.routes", []any{})
	b, err := yaml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if trimmed != "routes: []" {
			out = append(out, line)
			continue
		}
		indent := line[:len(line)-len(trimmed)]
		out = append(out, indent+"routes:")
		for _, e := range entries {
			out = append(out, indent+"  "+e)
		}
	}
	return afsWriteRaw(t, strings.Join(out, "\n"))
}

// TestAgenticFactorySettings_RefCycleChild is the child half of the
// subprocess test below. Run directly (no env var) it skips.
func TestAgenticFactorySettings_RefCycleChild(t *testing.T) {
	root := os.Getenv(afsRefChildEnv)
	if root == "" {
		t.Skip("child of TestAgenticFactorySettings_RefCycle")
	}
	_, err := LoadAgenticFactorySettings(root)
	if err == nil {
		os.Stdout.WriteString("RESULT:nil\n")
		return
	}
	os.Stdout.WriteString("RESULT:" + err.Error() + "\n")
}

// A self-referencing YAML reference under webhook.routes must be a config
// error, never a stack overflow. A stack overflow is fatal (not recoverable),
// so the load runs in a child process of this test binary.
func TestAgenticFactorySettings_RefCycle(t *testing.T) {
	root := afsWriteRoutesRaw(t, []string{
		"- &cycle",
		"    repository: o/n",
		"    nested: *cycle",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAgenticFactorySettings_RefCycleChild$", "-test.v")
	cmd.Env = append(os.Environ(), afsRefChildEnv+"="+root)
	out, err := cmd.CombinedOutput()
	text := string(out)
	if strings.Contains(text, "stack exceeds") || strings.Contains(text, "fatal error") {
		head := text
		if len(head) > 600 {
			head = head[:600]
		}
		t.Fatalf("loading a self-referencing alias crashed the process (%v):\n%s", err, head)
	}
	if err != nil {
		t.Fatalf("child failed: %v\n%.600s", err, text)
	}
	var result string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "RESULT:") {
			result = strings.TrimPrefix(line, "RESULT:")
		}
	}
	if result == "" || result == "nil" {
		t.Fatalf("expected a config error for the cycle, got %q", result)
	}
	for _, want := range []string{`field "workflow.agentic_factory.webhook.routes[0]"`, "alias", "cycle"} {
		if !strings.Contains(result, want) {
			t.Errorf("error %q does not contain %q", result, want)
		}
	}
	if strings.Contains(result, "o/n") {
		t.Errorf("error echoes a route value: %q", result)
	}
}

// A shared anchor that is not a cycle (the same anchor used twice) is valid.
func TestAgenticFactorySettings_SharedAnchorNotRejected(t *testing.T) {
	root := afsWriteRoutesRaw(t, []string{
		"- repository: o/n",
		"  auth: &shared",
		"    kind: ref",
		"  other: *shared",
		"- repository: o/m",
		"  auth: *shared",
	})
	s := afsWantValid(t, root)
	if len(s.WebhookRoutes) != 2 {
		t.Fatalf("routes = %v", s.WebhookRoutes)
	}
}

// A secret key reached through a (non-cyclic) reference is still refused.
func TestAgenticFactorySettings_SecretThroughAnchorStillRefused(t *testing.T) {
	root := afsWriteRoutesRaw(t, []string{
		"- repository: o/n",
		"  auth: &a",
		"    secret: not-a-real-value",
		"- repository: o/m",
		"  auth: *a",
	})
	_, err := LoadAgenticFactorySettings(root)
	if err == nil || !strings.Contains(err.Error(), "webhook.routes[0].secret") || strings.Contains(err.Error(), "not-a-real-value") {
		t.Fatalf("err = %v", err)
	}
}

func TestAfsHasSecretKey_CycleAndSharing(t *testing.T) {
	parse := func(src string) *yaml.Node {
		t.Helper()
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
			t.Fatal(err)
		}
		return doc.Content[0]
	}
	cases := []struct {
		name        string
		src         string
		found, loop bool
	}{
		{"self reference in a map", "&c\nnested: *c\n", false, true},
		{"self reference in a list", "&c\n- *c\n", false, true},
		{"indirect cycle", "&a\nb: &b\n  back: *a\n", false, true},
		{"shared anchor, no cycle", "x: &s {k: v}\ny: *s\nz: [*s, *s]\n", false, false},
		{"secret through a shared anchor", "x: &s {secret: v}\ny: *s\n", true, false},
		{"plain secret", "secret: v\n", true, false},
		{"no secret", "a: [1, {b: c}]\n", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			found, loop := afsHasSecretKey(parse(tc.src))
			if found != tc.found || loop != tc.loop {
				t.Fatalf("got found=%v cycle=%v, want found=%v cycle=%v", found, loop, tc.found, tc.loop)
			}
		})
	}
}
