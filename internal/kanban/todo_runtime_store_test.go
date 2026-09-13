package kanban

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestTodoRuntimeStorePublicReadbackSurvivesCardEdit(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	specPath := filepath.Join(root, ".moai", "specs", "SPEC-FIXTURE-001", "spec.md")
	if err := os.MkdirAll(filepath.Dir(specPath), 0700); err != nil {
		t.Fatal(err)
	}
	specBody := []byte("# Runtime provenance fixture\n")
	if err := os.WriteFile(specPath, specBody, 0600); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("fixture git %v: %v: %s", args, err, out)
		}
		return bytes.TrimSpace(out)
	}
	git("init", "-q")
	git("add", ".moai/specs/SPEC-FIXTURE-001/spec.md")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
	commit := string(git("rev-parse", "HEAD"))
	store := NewBacklogStore(BacklogPathForRoot(root))
	card, _, err := store.Add("runtime before edit")
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = legacy.Close() })
	for _, table := range []string{"runs", "cards"} {
		var count int
		if err := legacy.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("legacy fixture %s count=%d", table, count)
		}
	}
	if err := RecordFactoryRunStart(root, "fixture-run", BackendClaude, "SPEC-FIXTURE-001"); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardAssignment(root, "fixture-run", card.ID, "worker-1", "SPEC-FIXTURE-001"); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardState(root, "fixture-run", card.ID, "worker-2", "SPEC-FIXTURE-001", "completed", "card.completed"); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(specBody)
	check := func(stage string) {
		t.Helper()
		record, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if len(record.Items) != 1 || record.Items[0].ID != card.ID || record.Items[0].State != BacklogStateQueued {
			t.Fatalf("card membership/state changed: %+v", record.Items)
		}
		raw, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		if _, ok := decoded["runtime"]; !ok {
			t.Errorf("%s: successful runtime writes missing from public Todo JSON: %s", stage, raw)
			return
		}
		var runtime struct {
			Runs        []map[string]string `json:"runs"`
			Assignments []map[string]string `json:"assignments"`
		}
		if err := json.Unmarshal(decoded["runtime"], &runtime); err != nil {
			t.Fatal(err)
		}
		if len(runtime.Runs) != 1 || len(runtime.Assignments) != 1 {
			t.Errorf("%s: expected one upserted run/assignment, got %s", stage, decoded["runtime"])
			return
		}
		if runtime.Runs[0]["run_id"] != "fixture-run" || runtime.Runs[0]["backend"] != BackendClaude {
			t.Errorf("run values: %+v", runtime.Runs[0])
		}
		var manifest map[string]string
		if err := json.Unmarshal([]byte(runtime.Runs[0]["manifest_json"]), &manifest); err != nil {
			t.Fatal(err)
		}
		if manifest["spec_id"] != "" || manifest["spec_sha256"] != "" || manifest["git_commit"] != commit {
			t.Errorf("run provenance: %+v", manifest)
		}
		assignment := runtime.Assignments[0]
		for key, want := range map[string]string{"run_id": "fixture-run", "card_id": card.ID, "owner_label": "worker-2", "reported_state": "completed", "event_kind": "card.completed"} {
			if assignment[key] != want {
				t.Errorf("assignment %s=%q want %q", key, assignment[key], want)
			}
		}
		var provenance map[string]string
		if err := json.Unmarshal([]byte(assignment["provenance_json"]), &provenance); err != nil {
			t.Fatal(err)
		}
		if provenance["spec_id"] != "SPEC-FIXTURE-001" || provenance["spec_sha256"] != hex.EncodeToString(sum[:]) || provenance["git_commit"] != commit {
			t.Errorf("assignment provenance: %+v", provenance)
		}
		if _, err := time.Parse(time.RFC3339Nano, provenance["captured_at"]); err != nil {
			t.Errorf("capture timestamp: %v", err)
		}
	}
	check("before edit")
	if err := store.Mutate(func(record *BacklogRecord) error { record.Items[0].Text = "runtime after edit"; return nil }); err != nil {
		t.Fatal(err)
	}
	check("after edit")
	record, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if record.Items[0].Text != "runtime after edit" {
		t.Fatalf("card edit lost: %+v", record.Items)
	}
	for _, table := range []string{"runs", "cards"} {
		var count int
		if err := legacy.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("legacy %s increased from 0 to %d", table, count)
		}
	}
}

// This is a preservation guard, not evidence of the new runtime storage.
func TestTodoRuntimeStoreFutureSchemaPreservesBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backlog.db")
	eng, err := openBacklogEngine(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := eng.db.Exec(`UPDATE meta SET value = '999' WHERE key = ?`, backlogMetaKeySchemaVersion); err != nil {
		_ = eng.close()
		t.Fatal(err)
	}
	if err := eng.close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := openBacklogEngine(path)
	if err == nil {
		_ = opened.close()
		t.Fatal("future schema was accepted")
	}
	if !IsBacklogCorrupt(err) {
		t.Fatalf("unexpected refusal: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("future schema refusal changed persistent DB bytes")
	}
}

// A same-length content change must not satisfy the preservation assertion.
// This counterexample checks the assertion, not a production corruption claim.
func TestTodoRuntimeStoreByteGuardRejectsSameLengthChange(t *testing.T) {
	before := []byte("schema-version=999")
	after := bytes.Clone(before)
	after[len(after)-1] = '8'
	if len(before) != len(after) {
		t.Fatal("counterexample length changed")
	}
	if bytes.Equal(before, after) {
		t.Fatal("byte guard accepted the changed content")
	}
}

func TestTodoRuntimeStoreNonGitWithoutSpecSeedsRun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	store := NewBacklogStore(BacklogPathForRoot(root))
	card, _, err := store.Add("plain folder task")
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardAssignment(root, "plain-run", card.ID, "direct", ""); err != nil {
		t.Fatal(err)
	}
	record, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Items) != 1 || record.Items[0].ID != card.ID || record.Items[0].State != BacklogStateQueued {
		t.Fatalf("card changed: %+v", record.Items)
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Runtime *struct {
			Runs        []map[string]string `json:"runs"`
			Assignments []map[string]string `json:"assignments"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Runtime == nil {
		t.Fatalf("successful no-Git/no-SPEC assignment missing Todo runtime: %s", raw)
	}
	if len(decoded.Runtime.Runs) != 1 || len(decoded.Runtime.Assignments) != 1 {
		t.Fatalf("implicit run/assignment not paired: %s", raw)
	}
	if decoded.Runtime.Runs[0]["run_id"] != "plain-run" || decoded.Runtime.Assignments[0]["card_id"] != card.ID {
		t.Fatalf("implicit values differ: %s", raw)
	}
	var provenance map[string]string
	if err := json.Unmarshal([]byte(decoded.Runtime.Assignments[0]["provenance_json"]), &provenance); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"spec_id", "spec_path", "spec_sha256", "git_commit"} {
		if provenance[key] != "" {
			t.Errorf("non-Git/no-SPEC %s=%q", key, provenance[key])
		}
	}
	if _, err := time.Parse(time.RFC3339Nano, provenance["captured_at"]); err != nil {
		t.Errorf("capture timestamp: %v", err)
	}
}

func TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("legacy card"); err != nil {
		t.Fatal(err)
	}
	path := store.EnginePath()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		record, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		var runtime map[string]json.RawMessage
		if value, ok := decoded["runtime"]; !ok {
			t.Errorf("read %d: legacy public JSON lacks empty runtime: %s", i, raw)
		} else {
			if err := json.Unmarshal(value, &runtime); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"runs", "assignments"} {
				if string(runtime[key]) != "[]" {
					t.Errorf("legacy %s=%s want []", key, runtime[key])
				}
			}
		}
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("pure legacy reads changed DB bytes")
	}
}

func TestTodoRuntimeStoreWriterRejectsFutureTodoVersions(t *testing.T) {
	for _, key := range []string{backlogMetaKeySchemaVersion, "runtime_schema_version"} {
		t.Run(key, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("MOAI_HOME", t.TempDir())
			store := NewBacklogStore(BacklogPathForRoot(root))
			card, _, err := store.Add("must survive rejected runtime write")
			if err != nil {
				t.Fatal(err)
			}
			eng, err := openBacklogEngine(store.EnginePath())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := eng.db.Exec(`INSERT INTO meta(key,value) VALUES(?, '999') ON CONFLICT(key) DO UPDATE SET value='999'`, key); err != nil {
				_ = eng.close()
				t.Fatal(err)
			}
			if err := eng.close(); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(store.EnginePath())
			if err != nil {
				t.Fatal(err)
			}
			if err := RecordFactoryRunStart(root, "rejected-run", BackendClaude, ""); err == nil {
				t.Errorf("runtime run writer accepted future Todo %s=999", key)
			}
			if err := RecordFactoryCardAssignment(root, "rejected-assignment", card.ID, "direct", ""); err == nil {
				t.Errorf("runtime assignment writer accepted future Todo %s=999", key)
			}
			after, err := os.ReadFile(store.EnginePath())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Error("rejected runtime write changed future Todo DB bytes")
			}
		})
	}
}

func TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("partial extension"); err != nil {
		t.Fatal(err)
	}
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := eng.db.Exec(`CREATE TABLE todo_runtime_runs (run_id TEXT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if err := eng.close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.LoadPure(); !IsBacklogCorrupt(err) {
		t.Errorf("partial extension pure read must refuse corruption, got %v", err)
	}
	if err := RecordFactoryRunStart(root, "partial", BackendClaude, ""); !IsBacklogCorrupt(err) {
		t.Errorf("partial extension writer must refuse corruption, got %v", err)
	}
	after, err := os.ReadFile(store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("partial schema refusal changed DB bytes")
	}
}
