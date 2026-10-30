package cli

// web_agentic_factory_test.go — the Web Console's Agentic Factory sources
// (Issue #1948).
//
// Defect answer: these tests catch a console that creates the claim store it was
// asked to read, one that shows a broken settings file as "not provided" (or an
// unset one as an error), a relative store directory read from the wrong place,
// a source that writes, and runWeb dropping the sources on the way to web.Run.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
	"github.com/modu-ai/moai-adk/internal/web"
	"gopkg.in/yaml.v3"
)

const (
	afwSHA  = "0123456789abcdef0123456789abcdef01234567"
	afwHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

// afwDoc is a fully valid agentic-factory.yaml whose claim store is dir.
func afwDoc(dir string) map[string]any {
	m := func(kv ...any) map[string]any {
		out := map[string]any{}
		for i := 0; i < len(kv); i += 2 {
			out[kv[i].(string)] = kv[i+1]
		}
		return out
	}
	return m("workflow", m("agentic_factory", m(
		"enabled", true,
		"claim_store", m("dir", dir),
		"provider", m("github", m("backoff", m("min_secondary_wait", "60s", "base", "1s", "max", "5m", "max_attempts", 5))),
		"collector", m("poll", m("min", "10s", "max", "5m", "safety", "15m")),
		"budget", m("reserve", 10, "report_resource", "core"),
		"identity", m("max_age", "10m"),
		"labels", m("per_page", 50),
		"intake", m("duplicate_search_limit", 20, "list_limit", 100),
		"monitor", m("pr_list_limit", 200),
		"cursor", m("max_doc_bytes", 65536, "max_chunks", 4),
		"queue", m("webhook", m("max_doc_bytes", 262144, "seen_limit", 100, "max_id_bytes", 36)),
		"notify", m("inbox", m("max_doc_bytes", 262144, "seen_limit", 50, "max_id_bytes", 32), "renotify_after", "30m"),
		"webhook", m("max_body", 1048576, "provider", "github", "routes", []any{m("repository", "Owner/Name")}),
		"wait", m("max_retries", 3, "escalate_to", "operator"),
		"pr_source", m("repository", "owner/name"),
		"context", m("summary_max_bytes", 4096),
	)))
}

func afwWrite(t *testing.T, root string, doc map[string]any) {
	t.Helper()
	b, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agentic-factory.yaml"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// afwSeedStore opens dir as a claim store and writes one ledger item for issue.
func afwSeedStore(t *testing.T, dir string, issue int) {
	t.Helper()
	s, err := claim.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ctxmodel.Update(s, issue, func(l *ctxmodel.Ledger) error {
		return l.Add(ctxmodel.Item{
			Section: ctxmodel.SectionPlan, Title: "plan", Path: ".moai/specs/SPEC-X/plan.md",
			Commit: afwSHA, SHA256: afwHash, Owner: "lane-18",
		})
	}); err != nil {
		t.Fatal(err)
	}
}

// afwTree lists every path under dir with its size, to prove a read changed nothing.
func afwTree(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		b.WriteString(rel + ":" + info.Mode().String() + ":" + string(rune('0'+info.Size()%10)) + "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestWebAF_NotSetUp_LeavesThePanelsNotProvided(t *testing.T) {
	root := t.TempDir()
	if src := agenticFactoryWebSourcesFor(root); src.IntakeRequests != nil || src.ContextModel != nil {
		t.Errorf("no settings file: sources = %+v, want none (the panels say 'not provided')", src)
	}
	disabled := afwDoc("claims")
	disabled["workflow"].(map[string]any)["agentic_factory"].(map[string]any)["enabled"] = false
	afwWrite(t, root, disabled)
	if src := agenticFactoryWebSourcesFor(root); src.IntakeRequests != nil || src.ContextModel != nil {
		t.Errorf("enabled false: sources = %+v, want none", src)
	}
}

func TestWebAF_BrokenSettings_ShowTheReasonNotNothing(t *testing.T) {
	root := t.TempDir()
	doc := afwDoc("claims")
	delete(doc["workflow"].(map[string]any)["agentic_factory"].(map[string]any), "wait")
	afwWrite(t, root, doc)
	src := agenticFactoryWebSourcesFor(root)
	if src.IntakeRequests == nil || src.ContextModel == nil {
		t.Fatalf("a settings file that is set up but invalid must give sources that say why: %+v", src)
	}
	if _, err := src.ContextModel(context.Background(), 7); err == nil || !strings.Contains(err.Error(), "workflow.agentic_factory.wait") {
		t.Errorf("context model error = %v, want it to name the missing key", err)
	}
	if _, err := src.IntakeRequests(context.Background()); err == nil || !strings.Contains(err.Error(), "workflow.agentic_factory.wait") {
		t.Errorf("intake requests error = %v, want it to name the missing key", err)
	}
}

// The console reads the store; it never makes one. A configured directory that
// is not there is reported, and not created.
func TestWebAF_MissingStoreDir_IsReportedAndNotCreated(t *testing.T) {
	root := t.TempDir()
	afwWrite(t, root, afwDoc("state/claims"))
	src := agenticFactoryWebSourcesFor(root)
	if src.ContextModel == nil {
		t.Fatal("a configured console with a missing store directory must say so")
	}
	if _, err := src.ContextModel(context.Background(), 7); err == nil || !strings.Contains(err.Error(), "claim store directory") {
		t.Errorf("error = %v, want it to name the claim store directory", err)
	}
	if _, err := os.Stat(filepath.Join(root, "state")); !os.IsNotExist(err) {
		t.Errorf("the console created the store directory (stat err %v)", err)
	}
}

func TestWebAF_StoreDirIsAFile_IsReported(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "claims"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	afwWrite(t, root, afwDoc("claims"))
	src := agenticFactoryWebSourcesFor(root)
	if _, err := src.IntakeRequests(context.Background()); err == nil || !strings.Contains(err.Error(), "claim store directory") {
		t.Errorf("error = %v, want it to name the claim store directory", err)
	}
}

func TestWebAF_RelativeAndAbsoluteStoreDir_ReadTheSameStore(t *testing.T) {
	for _, relative := range []bool{true, false} {
		root := t.TempDir()
		store := filepath.Join(root, "state", "claims")
		dir := store
		if relative {
			dir = filepath.Join("state", "claims")
		}
		afwSeedStore(t, store, 7)
		afwWrite(t, root, afwDoc(dir))
		src := agenticFactoryWebSourcesFor(root)
		if src.ContextModel == nil || src.IntakeRequests == nil {
			t.Fatalf("relative=%v: sources = %+v", relative, src)
		}
		raw, err := src.ContextModel(context.Background(), 7)
		if err != nil {
			t.Fatalf("relative=%v: %v", relative, err)
		}
		led, err := ctxmodel.Unmarshal(raw)
		if err != nil || led.Issue != 7 || len(led.Items) != 1 || led.Items[0].Title != "plan" {
			t.Errorf("relative=%v: ledger = %+v, %v; want issue 7 with the stored item", relative, led, err)
		}
		if reqs, err := src.IntakeRequests(context.Background()); err != nil || !json.Valid(reqs) {
			t.Errorf("relative=%v: intake requests = %q, %v", relative, reqs, err)
		}
	}
}

func TestWebAF_Sources_NeverWrite(t *testing.T) {
	root := t.TempDir()
	store := filepath.Join(root, "claims")
	afwSeedStore(t, store, 7)
	afwWrite(t, root, afwDoc("claims"))
	before := afwTree(t, store)
	src := agenticFactoryWebSourcesFor(root)
	for i := 0; i < 3; i++ {
		_, _ = src.ContextModel(context.Background(), 7)
		_, _ = src.ContextModel(context.Background(), 8) // never written: reads as empty, writes nothing
		_, _ = src.IntakeRequests(context.Background())
	}
	if after := afwTree(t, store); after != before {
		t.Errorf("reading through the console changed the store:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// runWeb hands the sources to web.Run, and leaves MonitorRows alone: the monitor
// collector needs a remote reader, and the console builds none.
func TestWebAF_RunWeb_PassesTheSourcesThrough(t *testing.T) {
	run := func(t *testing.T, root string) web.Config {
		t.Helper()
		swapSessionWorktreeSeams(t, swSeams{})
		origFind := findProjectRootFn
		findProjectRootFn = func() (string, error) { return root, nil }
		t.Cleanup(func() { findProjectRootFn = origFind })
		var got web.Config
		origRun := webRunFn
		webRunFn = func(_ context.Context, c web.Config) error { got = c; return nil }
		t.Cleanup(func() { webRunFn = origRun })
		origCheck := checkPortInUse
		checkPortInUse = func(int) bool { return false }
		t.Cleanup(func() { checkPortInUse = origCheck })
		origCwd, cwdErr := os.Getwd()
		t.Cleanup(func() {
			if cwdErr == nil {
				_ = os.Chdir(origCwd)
			}
			webPort, webNoOpen, webNoReuse = 3041, false, false
		})
		cmd := newWebCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("moai web: %v", err)
		}
		return got
	}
	t.Run("configured", func(t *testing.T) {
		root := t.TempDir()
		afwSeedStore(t, filepath.Join(root, "claims"), 7)
		afwWrite(t, root, afwDoc("claims"))
		c := run(t, root)
		if c.ContextModel == nil || c.IntakeRequests == nil {
			t.Errorf("a configured project reached web.Run without its sources: %+v", c)
		}
		if c.MonitorRows != nil {
			t.Error("MonitorRows was wired, but the console has no remote reader to build the collector with")
		}
	})
	t.Run("not set up", func(t *testing.T) {
		c := run(t, t.TempDir())
		if c.ContextModel != nil || c.IntakeRequests != nil || c.MonitorRows != nil {
			t.Errorf("a project without agentic-factory.yaml reached web.Run with sources: %+v", c)
		}
	})
}
