package cli

import (
	"bytes"
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func commandSurface(cmd interface {
	Commands() []*cobra.Command
}) []string {
	var out []string
	for _, child := range cmd.Commands() {
		out = append(out, child.Name())
	}
	sort.Strings(out)
	return out
}

func flagSurface(cmd *cobra.Command) []string {
	var out []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		out = append(out, strings.Join([]string{flag.Name, flag.Shorthand, flag.Value.Type(), flag.DefValue, flag.NoOptDefVal}, "\x00"))
	})
	sort.Strings(out)
	return out
}

func compatibilityHelp(t *testing.T, root *cobra.Command, verb string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{verb, "--help"})
	err := root.Execute()
	text := strings.ReplaceAll(stdout.String(), "moai gtd", "moai queue")
	text = strings.ReplaceAll(text, "moai todo", "moai queue")
	text = strings.ReplaceAll(text, "gtd "+verb, "queue "+verb)
	text = strings.ReplaceAll(text, "todo "+verb, "queue "+verb)
	return text, stderr.String(), err
}

func TestGTDAllTodoVerbsParity(t *testing.T) {
	todoWant := []string{"add", "analyze", "auto-done", "done", "drop", "edit", "export-json", "history", "landed", "list", "move", "next", "pr", "relate", "undone", "undrop", "unpick", "unrelate", "why"}
	sort.Strings(todoWant)
	gtdWant := append(slices.Clone(todoWant), "capture", "clarify", "organize", "reflect", "engage", "answer")
	sort.Strings(gtdWant)

	gtd := NewGTDCommand()
	todo := newTodoCmd()
	if gtd.Name() != "gtd" || todo.Name() != "todo" {
		t.Fatalf("command names = %q/%q, want gtd/todo", gtd.Name(), todo.Name())
	}
	if got := commandSurface(gtd); !slices.Equal(got, gtdWant) {
		t.Fatalf("gtd verbs = %v, want %v", got, gtdWant)
	}
	if got := commandSurface(todo); !slices.Equal(got, todoWant) {
		t.Fatalf("todo verbs = %v, want %v", got, todoWant)
	}
	for _, verb := range todoWant {
		gtdChild, _, err := gtd.Find([]string{verb})
		if err != nil {
			t.Fatalf("gtd %s lookup: %v", verb, err)
		}
		todoChild, _, err := todo.Find([]string{verb})
		if err != nil {
			t.Fatalf("todo %s lookup: %v", verb, err)
		}
		if gtdChild.Use != todoChild.Use || gtdChild.Short != todoChild.Short || !slices.Equal(flagSurface(gtdChild), flagSurface(todoChild)) || (gtdChild.Args == nil) != (todoChild.Args == nil) {
			t.Fatalf("%s handler/flag contract diverged: gtd=%q/%v todo=%q/%v", verb, gtdChild.Use, flagSurface(gtdChild), todoChild.Use, flagSurface(todoChild))
		}
		gtdOut, gtdErr, gtdExit := compatibilityHelp(t, NewGTDCommand(), verb)
		todoOut, todoErr, todoExit := compatibilityHelp(t, newTodoCmd(), verb)
		if gtdExit != nil || todoExit != nil || gtdErr != todoErr || gtdOut != todoOut {
			t.Fatalf("%s help exit/stdout/stderr diverged: gtdErr=%v todoErr=%v", verb, gtdExit, todoExit)
		}
	}

	// One stateful compatibility cell proves both roots reach the same queue
	// constructor and preserve the existing issued identity.
	_, _ = todoFixture(t)
	var out bytes.Buffer
	gtd.SetOut(&out)
	gtd.SetErr(&out)
	gtd.SetArgs([]string{"add", "canonical compatibility card"})
	if err := gtd.Execute(); err != nil {
		t.Fatalf("gtd add: %v", err)
	}
	if got := out.String(); !strings.HasSuffix(got, "t1 1\n") {
		t.Fatalf("gtd add output = %q, want suffix t1 1", got)
	}
	out.Reset()
	todo.SetOut(&out)
	todo.SetErr(&out)
	todo.SetArgs([]string{"list", "--json"})
	if err := todo.Execute(); err != nil {
		t.Fatalf("todo list after gtd add: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"id":"t1"`)) {
		t.Fatalf("todo did not observe gtd-issued t1: %s", out.String())
	}
}

func runGTD(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewGTDCommand()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestGTDFiveStageCLIUsesSameSQLite(t *testing.T) {
	root, store := todoFixture(t)
	out, err := runGTD(t, "capture", "untrusted inbox note", "--event", "cli-capture-1", "--source", "user", "--sensitivity", "private", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var captured struct {
		ItemID string `json:"item_id"`
	}
	if err := json.Unmarshal([]byte(out), &captured); err != nil || captured.ItemID == "" {
		t.Fatalf("capture output=%q err=%v", out, err)
	}
	if record, err := store.Load(); err != nil || len(record.Items) != 0 {
		t.Fatalf("capture published a queue card: %+v err=%v", record, err)
	}
	if _, err := runGTD(t, "clarify", captured.ItemID, "--disposition", "action", "--outcome", "landed", "--evidence", "CI", "--authority", "queue,dispatch", "--trusted"); err != nil {
		t.Fatal(err)
	}
	if _, err := runGTD(t, "organize", captured.ItemID, "--class", "action", "--context", "computer"); err != nil {
		t.Fatal(err)
	}
	if out, err := runGTD(t, "reflect", "--json"); err != nil || !bytes.Contains([]byte(out), []byte(captured.ItemID)) {
		t.Fatalf("reflect output=%q err=%v", out, err)
	}
	if out, err := runGTD(t, "reflect", "--rebuild-projection", "--json"); err != nil || !bytes.Contains([]byte(out), []byte(`"projection"`)) {
		t.Fatalf("projection output=%q err=%v", out, err)
	}
	if out, err := runGTD(t, "engage", captured.ItemID, "--approve", "--fresh", "--dependencies-ready", "--lane", "lane-10", "--resources", "--pick", "--dispatch", "--run-id", "mission-cli-1", "--json"); err != nil || !bytes.Contains([]byte(out), []byte(`"card_id":"t1"`)) {
		t.Fatalf("engage output=%q err=%v", out, err)
	}
	record, err := store.Load()
	if err != nil || len(record.Items) != 1 || record.Items[0].State != "picked" {
		t.Fatalf("engage queue=%+v err=%v", record, err)
	}
	if _, err := runGTD(t, "capture", "reserved word check", "--event", "cli-capture-2"); err != nil {
		t.Fatal(err)
	}
	_ = root
}

func TestGTDCLIInputRefusalsDoNotFallThrough(t *testing.T) {
	_, store := todoFixture(t)
	for _, args := range [][]string{{"capture", "x"}, {"capture", "x", "--event", "e", "--source", "feed"}, {"clarify", "../../escape", "--disposition", "reference"}, {"organize", "not-an-id", "--class", "action"}, {"engage", "not-an-id"}} {
		if _, err := runGTD(t, args...); err == nil {
			t.Fatalf("unsafe args allowed: %v", args)
		}
	}
	record, err := store.Load()
	if err != nil || len(record.Items) != 0 {
		t.Fatalf("refusal created cards: %+v err=%v", record, err)
	}
}
