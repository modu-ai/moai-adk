//go:build !windows

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	preApprovalLiveCallBound  = 330 * time.Second
	preApprovalDiscWindow     = 900 * time.Second
	preApprovalMaxLiveCalls   = 14
	preApprovalRefusal        = "MCP tool call requires approval, but approval policy is never"
	preApprovalM1aRowsSHA256  = "4e9b04acb14ed10b7122f1740fdcfaec147a91698bdebe2867a43123d6cb73ef"
	preApprovalM1aFilesSHA256 = "c8653794d64014487c9f05d3e54f07e4b5fb61cf783f5e18f6460bb0e6a4b937"
)

type preApprovalArmResult struct {
	MoaiRoleAuditCalls int    `json:"moai_role_audit_calls"`
	Outcome            string `json:"outcome"`
	ErrorText          string `json:"error_text,omitempty"`
	ResultText         string `json:"result_text,omitempty"`
	ChildProcesses     int    `json:"child_processes"`
	LaunchRecords      int    `json:"launch_records"`
	UserTurns          int    `json:"user_turns"`
}

type preApprovalDiscEvidence struct {
	CodexVersion     string                          `json:"codex_version"`
	Invocations      int                             `json:"invocations"`
	Attempt          int                             `json:"attempt"`
	Aborted          bool                            `json:"aborted"`
	ExportComplete   bool                            `json:"export_complete"`
	ArmDiffOK        bool                            `json:"arm_diff_ok"`
	StartupChecks    map[string]string               `json:"startup_checks"`
	Arms             map[string]preApprovalArmResult `json:"arms"`
	RecordedPIDs     []int                           `json:"recorded_pids"`
	KilledPIDs       []int                           `json:"killed_pids"`
	AuthSHA256Before string                          `json:"auth_sha256_before"`
	AuthSHA256After  string                          `json:"auth_sha256_after"`
}

func preApprovalReadLedger(path string) ([]map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func preApprovalWriteLedger(path string, rows []map[string]any) error {
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func preApprovalReadExports(dir string) (preApprovalManifest, []string, error) {
	var manifest preApprovalManifest
	b, err := os.ReadFile(filepath.Join(dir, "export-manifest.json"))
	if err != nil {
		return manifest, nil, err
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return manifest, nil, err
	}
	if manifest.ExportedNS <= 0 || len(manifest.Files) != 14 {
		return manifest, nil, errors.New("discriminator export manifest is incomplete")
	}
	var paths []string
	for _, arm := range []string{"control", "treatment"} {
		for _, name := range []string{"codex-home-config.toml", "project-config.toml", "repo.txt", "moai-build.txt", "argv.txt", "prompt.txt", "root-state.txt"} {
			key := arm + "/" + name
			path := filepath.Join(dir, "inputs", key)
			body, err := os.ReadFile(path)
			if err != nil || len(body) == 0 || preApprovalSHA(body) != manifest.Files[key] {
				return manifest, nil, fmt.Errorf("export mismatch: %s: %v", key, err)
			}
			paths = append(paths, path)
		}
	}
	actualFiles := 0
	err = filepath.WalkDir(filepath.Join(dir, "inputs"), func(_ string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			actualFiles++
		}
		return nil
	})
	if err != nil || actualFiles != 14 {
		return manifest, nil, fmt.Errorf("unexpected input file count %d: %v", actualFiles, err)
	}
	return manifest, paths, nil
}

func preApprovalArmInputs(dir, arm string) (home, project, argv, prompt, rootState []byte, err error) {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, "inputs", arm, name))
	}
	if home, err = read("codex-home-config.toml"); err != nil {
		return
	}
	if project, err = read("project-config.toml"); err != nil {
		return
	}
	if argv, err = read("argv.txt"); err != nil {
		return
	}
	if prompt, err = read("prompt.txt"); err != nil {
		return
	}
	rootState, err = read("root-state.txt")
	return
}

func preApprovalStartupFromFiles(dir, fixture, startupID string) (preApprovalStartupProof, []byte, error) {
	proof := preApprovalStartupProof{Fixture: fixture}
	var err error
	if startupID == "" || filepath.IsAbs(startupID) || strings.Contains(startupID, "..") {
		return proof, nil, errors.New("invalid startup ID")
	}
	base := filepath.Join(dir, "startup", startupID)
	if proof.Stdout, err = os.ReadFile(base + ".jsonl"); err != nil {
		return proof, nil, err
	}
	if proof.Stderr, err = os.ReadFile(base + ".err"); err != nil {
		return proof, nil, err
	}
	counts, err := os.ReadFile(base + ".launches.json")
	if err != nil {
		return proof, nil, err
	}
	if err := json.Unmarshal(counts, &proof.Launches); err != nil {
		return proof, nil, err
	}
	return proof, counts, nil
}

func preApprovalLatestStartupRow(rows []map[string]any, fixture string) map[string]any {
	var latest map[string]any
	for _, row := range rows {
		if row["kind"] == "startup" && row["fixture"] == fixture {
			if latest == nil || preApprovalNumber(row["started_ns"]) > preApprovalNumber(latest["started_ns"]) {
				latest = row
			}
		}
	}
	return latest
}

func preApprovalLatestStartup(rows []map[string]any, fixture string, exportedNS int64, tempRoot, root, exportID, exportHash string, hashes map[string]string, proof preApprovalStartupProof, launches []byte) (string, bool) {
	latest := preApprovalLatestStartupRow(rows, fixture)
	if latest == nil || preApprovalNumber(latest["started_ns"]) <= exportedNS ||
		preApprovalNumber(latest["ended_ns"]) < preApprovalNumber(latest["started_ns"]) ||
		latest["temp_root"] != tempRoot || latest["fixture_root"] != root || latest["fixture_sentinel"] != true ||
		latest["export_id"] != exportID || latest["export_sha256"] != exportHash ||
		latest["passed"] != true || proof.Fixture != fixture || !preApprovalStartupValid(proof) ||
		latest["stdout_sha256"] != preApprovalSHA(proof.Stdout) || latest["stderr_sha256"] != preApprovalSHA(proof.Stderr) ||
		latest["launches_sha256"] != preApprovalSHA(launches) ||
		preApprovalNumber(latest["moai_launches"]) != int64(proof.Launches["moai"]) ||
		preApprovalNumber(latest["decoy_launches"]) != int64(proof.Launches["decoy"]) {
		return "", false
	}
	startupID, ok := latest["startup_id"].(string)
	if !ok || startupID == "" {
		return "", false
	}
	inputHashes, ok := latest["inputs_sha256"].(map[string]any)
	if !ok {
		return "", false
	}
	for key, want := range hashes {
		if inputHashes[key] != want {
			return "", false
		}
	}
	return startupID, true
}

func preApprovalNumber(value any) int64 {
	switch n := value.(type) {
	case json.Number:
		result, _ := n.Int64()
		return result
	case int64:
		return n
	case int:
		return int64(n)
	case string:
		result, _ := strconv.ParseInt(n, 10, 64)
		return result
	default:
		return 0
	}
}

func preApprovalNextAttempt(rows []map[string]any) (int, error) {
	lastEnd := int64(0)
	priorCalls := 0
	invalidatedAt := int64(0)
	for _, row := range rows {
		switch row["kind"] {
		case "stop":
			return 0, errors.New("stopped evidence directory cannot be reused")
		case "live":
			fixture, _ := row["fixture"].(string)
			if strings.HasPrefix(fixture, "disc-") {
				if preApprovalNumber(row["attempt"]) != 1 {
					return 0, errors.New("discriminator retry limit reached")
				}
				priorCalls++
				if end := preApprovalNumber(row["ended_ns"]); end > lastEnd {
					lastEnd = end
				}
			}
		case "invalidate":
			if row["fixture"] == "disc" && preApprovalNumber(row["attempt"]) == 1 && row["recorded_by"] == "lead" {
				invalidatedAt = preApprovalNumber(row["recorded_ns"])
			}
		}
	}
	if priorCalls == 0 {
		if invalidatedAt != 0 {
			return 0, errors.New("orphan discriminator invalidation")
		}
		return 1, nil
	}
	if priorCalls > 2 || invalidatedAt <= lastEnd {
		return 0, errors.New("prior discriminator attempt needs a later lead invalidation")
	}
	return 2, nil
}

func preApprovalHasInitialStartups(rows []map[string]any) bool {
	if len(rows) < 4 {
		return false
	}
	for i, fixture := range []string{"disc-control", "disc-treatment", "car010", "car011"} {
		if rows[i]["kind"] != "startup" || rows[i]["fixture"] != fixture {
			return false
		}
	}
	return true
}

func preApprovalM1aBaseline(evidenceDir string, rows []map[string]any) bool {
	if !preApprovalHasInitialStartups(rows) || !strings.HasSuffix(filepath.ToSlash(evidenceDir), "/.moai/reports/t1172") {
		return false
	}
	var canonical []byte
	for _, row := range rows[:4] {
		body, err := json.Marshal(row)
		if err != nil {
			return false
		}
		canonical = append(canonical, body...)
		canonical = append(canonical, '\n')
	}
	if preApprovalSHA(canonical) != preApprovalM1aRowsSHA256 {
		return false
	}
	var files []byte
	for _, fixture := range []string{"car010", "car011", "disc-control", "disc-treatment"} {
		for _, ext := range []string{".err", ".jsonl", ".launches.json"} {
			name := fixture + ext
			body, err := os.ReadFile(filepath.Join(evidenceDir, "startup", name))
			if err != nil {
				return false
			}
			files = fmt.Appendf(files, "%s  .moai/reports/t1172/startup/%s\n", preApprovalSHA(body), name)
		}
	}
	return preApprovalSHA(files) == preApprovalM1aFilesSHA256
}

func preApprovalPrepareArm(t *testing.T, f *preApprovalFixture, authPath string, arm string, home, project, rootState []byte) error {
	if err := preApprovalCleanFixture(f.root, f.tempRoot, f.token, f.worktrees, f.git); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(f.root, ".codex"), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(f.root, ".codex", "config.toml"), project, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(f.root, ".fixture-sentinel"), []byte(f.token), 0o600); err != nil {
		return err
	}
	if got := preApprovalRootState(t, f.root, f.git); !bytes.Equal(got, rootState) {
		return fmt.Errorf("%s root state changed after clean", arm)
	}
	if err := os.RemoveAll(f.codexHome); err != nil {
		return err
	}
	if err := os.Mkdir(f.codexHome, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(f.codexHome, "config.toml"), home, 0o600); err != nil {
		return err
	}
	return os.Symlink(authPath, filepath.Join(f.codexHome, "auth.json"))
}

func preApprovalFlattenStrings(v any, out *[]string) {
	switch x := v.(type) {
	case string:
		*out = append(*out, x)
	case []any:
		for _, item := range x {
			preApprovalFlattenStrings(item, out)
		}
	case map[string]any:
		for _, item := range x {
			preApprovalFlattenStrings(item, out)
		}
	}
}

func preApprovalClassify(data []byte) preApprovalArmResult {
	r := preApprovalArmResult{Outcome: "NOT MEASURED"}
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		var event map[string]any
		if json.Unmarshal(line, &event) != nil {
			continue
		}
		if event["type"] == "turn.started" {
			r.UserTurns++
		}
		item, ok := event["item"].(map[string]any)
		if !ok || item["type"] != "mcp_tool_call" || item["server"] != "moai" || item["tool"] != "codex_role_audit" {
			continue
		}
		if event["type"] != "item.completed" {
			continue
		}
		r.MoaiRoleAuditCalls++
		var stringsFound []string
		preApprovalFlattenStrings(item["error"], &stringsFound)
		for _, s := range stringsFound {
			if s == preApprovalRefusal {
				r.Outcome, r.ErrorText = "REFUSED", s
				break
			}
		}
		if r.Outcome == "REFUSED" {
			continue
		}
		stringsFound = nil
		preApprovalFlattenStrings(item["result"], &stringsFound)
		for _, s := range stringsFound {
			if strings.HasPrefix(s, "codex_role_audit: codex audit sync-auditor: destination rejected: ") ||
				strings.HasPrefix(s, "codex_role_audit: codex audit sync-auditor: working root rejected: ") {
				r.Outcome, r.ResultText = "RAN", s
				break
			}
		}
	}
	return r
}

func TestCodexPreApprovalLivePreflight(t *testing.T) {
	dir := t.TempDir()
	manifest := preApprovalManifest{ExportedNS: time.Now().UnixNano(), Files: map[string]string{}}
	for _, arm := range []string{"control", "treatment"} {
		for _, name := range []string{"codex-home-config.toml", "project-config.toml", "repo.txt", "moai-build.txt", "argv.txt", "prompt.txt", "root-state.txt"} {
			preApprovalExport(t, &manifest, filepath.Join(dir, "inputs"), arm+"/"+name, []byte(arm+" "+name))
		}
	}
	preApprovalWriteJSON(t, filepath.Join(dir, "export-manifest.json"), manifest)
	if _, paths, err := preApprovalReadExports(dir); err != nil || len(paths) != 14 {
		t.Fatalf("valid export refused: paths=%d err=%v", len(paths), err)
	}
	project := filepath.Join(dir, "inputs", "treatment", "project-config.toml")
	if err := os.WriteFile(project, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := preApprovalReadExports(dir); err == nil {
		t.Fatal("mutated export hash accepted")
	}
	preApprovalExport(t, &manifest, filepath.Join(dir, "inputs"), "treatment/project-config.toml", []byte("treatment project-config.toml"))
	if err := os.WriteFile(filepath.Join(dir, "inputs", "control", "extra"), []byte("extra"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := preApprovalReadExports(dir); err == nil {
		t.Fatal("extra exported file accepted")
	}
	if err := os.Remove(filepath.Join(dir, "inputs", "control", "extra")); err != nil {
		t.Fatal(err)
	}
	start := manifest.ExportedNS + 100
	startupProof := preApprovalStartupProof{Fixture: "disc-control", Stdout: []byte(`{"type":"thread.started"}` + "\n"),
		Launches: map[string]int{"moai": 1, "decoy": 0}}
	launches := []byte("{\"moai\":1,\"decoy\":0}\n")
	rows := []map[string]any{{"kind": "startup", "fixture": "disc-control", "started_ns": json.Number(fmt.Sprint(start)),
		"ended_ns": json.Number(fmt.Sprint(start + 10)), "passed": true,
		"export_id": "export-123", "export_sha256": "export-hash",
		"stdout_sha256": preApprovalSHA(startupProof.Stdout), "stderr_sha256": preApprovalSHA(startupProof.Stderr),
		"launches_sha256": preApprovalSHA(launches), "startup_id": "attempt-1/disc-control-123",
		"moai_launches": json.Number("1"), "decoy_launches": json.Number("0"),
		"temp_root": "/tmp/test", "fixture_root": "/tmp/test/fixture", "fixture_sentinel": true,
		"inputs_sha256": map[string]any{"project_config": "same"}}}
	valid := func(exported int64, temp, hash string, proof preApprovalStartupProof, launches []byte) bool {
		_, ok := preApprovalLatestStartup(rows, "disc-control", exported, temp, "/tmp/test/fixture", "export-123", "export-hash", map[string]string{"project_config": hash}, proof, launches)
		return ok
	}
	if !valid(manifest.ExportedNS, "/tmp/test", "same", startupProof, launches) {
		t.Fatal("valid startup ledger rejected")
	}
	if valid(start, "/tmp/test", "same", startupProof, launches) ||
		valid(manifest.ExportedNS, "/tmp/test", "changed", startupProof, launches) ||
		valid(manifest.ExportedNS, "/tmp/other", "same", startupProof, launches) {
		t.Fatal("stale or mismatched startup ledger accepted")
	}
	rows[0]["passed"] = false
	if valid(manifest.ExportedNS, "/tmp/test", "same", startupProof, launches) {
		t.Fatal("failed startup ledger accepted")
	}
	rows[0]["passed"] = true
	corruptProof := startupProof
	corruptProof.Stdout = []byte(`{"type":"thread.started"}` + "\n" + `{"type":"thread.started"}` + "\n")
	if valid(manifest.ExportedNS, "/tmp/test", "same", corruptProof, launches) ||
		valid(manifest.ExportedNS, "/tmp/test", "same", startupProof, []byte("changed")) {
		t.Fatal("altered startup raw output accepted")
	}
	rows[0]["export_sha256"] = "changed"
	if valid(manifest.ExportedNS, "/tmp/test", "same", startupProof, launches) {
		t.Fatal("mismatched startup export accepted")
	}
	rows[0]["export_sha256"] = "export-hash"
	refused := []byte(`{"type":"turn.started"}` + "\n" + `{"type":"item.completed","item":{"type":"mcp_tool_call","server":"moai","tool":"codex_role_audit","error":"MCP tool call requires approval, but approval policy is never"}}` + "\n")
	if got := preApprovalClassify(refused); got.Outcome != "REFUSED" || got.UserTurns != 1 || got.MoaiRoleAuditCalls != 1 {
		t.Fatalf("refusal classification = %+v", got)
	}
	run := []byte(`{"type":"turn.started"}` + "\n" + `{"type":"item.completed","item":{"type":"mcp_tool_call","server":"moai","tool":"codex_role_audit","result":"codex_role_audit: codex audit sync-auditor: destination rejected: AGENTS.md"}}` + "\n")
	if got := preApprovalClassify(run); got.Outcome != "RAN" || got.UserTurns != 1 || got.MoaiRoleAuditCalls != 1 {
		t.Fatalf("tool-run classification = %+v", got)
	}
	unknown := bytes.Replace(run, []byte("destination rejected: AGENTS.md"), []byte("tool call timed out after 60s"), 1)
	if got := preApprovalClassify(unknown); got.Outcome != "NOT MEASURED" {
		t.Fatalf("timeout classified as run: %+v", got)
	}
	if got, err := preApprovalNextAttempt(nil); err != nil || got != 1 {
		t.Fatalf("first attempt = %d, %v", got, err)
	}
	prior := []map[string]any{{"kind": "live", "fixture": "disc-control", "attempt": json.Number("1"), "ended_ns": json.Number("100")},
		{"kind": "live", "fixture": "disc-treatment", "attempt": json.Number("1"), "ended_ns": json.Number("110")}}
	if _, err := preApprovalNextAttempt(prior); err == nil {
		t.Fatal("retry without lead invalidation accepted")
	}
	late := append(append([]map[string]any{}, prior...), map[string]any{"kind": "invalidate", "fixture": "disc", "attempt": json.Number("1"), "recorded_by": "lead", "recorded_ns": json.Number("111")})
	if got, err := preApprovalNextAttempt(late); err != nil || got != 2 {
		t.Fatalf("authorized retry = %d, %v", got, err)
	}
	for _, bad := range []map[string]any{
		{"kind": "invalidate", "fixture": "disc", "attempt": json.Number("1"), "recorded_by": "lead", "recorded_ns": json.Number("110")},
		{"kind": "invalidate", "fixture": "disc", "attempt": json.Number("1"), "recorded_by": "agent", "recorded_ns": json.Number("111")},
	} {
		if _, err := preApprovalNextAttempt(append(append([]map[string]any{}, prior...), bad)); err == nil {
			t.Fatalf("invalid retry accepted: %+v", bad)
		}
	}
	if _, err := preApprovalNextAttempt(append(late, map[string]any{"kind": "live", "fixture": "disc-control", "attempt": json.Number("2"), "ended_ns": json.Number("120")})); err == nil {
		t.Fatal("third attempt accepted")
	}
	if _, err := preApprovalNextAttempt([]map[string]any{{"kind": "stop", "fixture": "disc", "reason": "budget"}}); err == nil {
		t.Fatal("stopped evidence directory accepted")
	}
	initial := []map[string]any{{"kind": "startup", "fixture": "disc-control"}, {"kind": "startup", "fixture": "disc-treatment"},
		{"kind": "startup", "fixture": "car010"}, {"kind": "startup", "fixture": "car011"}}
	if !preApprovalHasInitialStartups(initial) || preApprovalHasInitialStartups(initial[:3]) ||
		preApprovalHasInitialStartups(append([]map[string]any{{"kind": "live", "fixture": "disc-control"}}, initial...)) {
		t.Fatal("initial M1-a ledger boundary misclassified")
	}
	ledgerPath := filepath.Join(t.TempDir(), "ledger.json")
	initial[0]["started_ns"] = json.Number("1780000000000000001")
	if err := preApprovalWriteLedger(ledgerPath, initial); err != nil {
		t.Fatal(err)
	}
	retained, err := preApprovalReadLedger(ledgerPath)
	if err != nil || !preApprovalHasInitialStartups(retained) || preApprovalNumber(retained[0]["started_ns"]) != 1780000000000000001 {
		t.Fatalf("M1-a ledger not preserved exactly: %v, %+v", err, retained)
	}
	if preApprovalMaxStartupCalls != 16 || len(retained)+4 > preApprovalMaxStartupCalls {
		t.Fatal("startup cap or initial ledger count changed")
	}
}

func TestCodexPreApprovalM1aBaseline(t *testing.T) {
	source := filepath.Clean(filepath.Join("..", "..", ".moai", "reports", "t1172"))
	rows, err := preApprovalReadLedger(filepath.Join(source, "ledger.json"))
	if os.IsNotExist(err) {
		t.Skip("M1-a local evidence unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	copyDir := filepath.Join(t.TempDir(), ".moai", "reports", "t1172")
	if err := os.MkdirAll(filepath.Join(copyDir, "startup"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"car010", "car011", "disc-control", "disc-treatment"} {
		for _, ext := range []string{".err", ".jsonl", ".launches.json"} {
			name := fixture + ext
			body, err := os.ReadFile(filepath.Join(source, "startup", name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(copyDir, "startup", name), body, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !preApprovalM1aBaseline(copyDir, rows) {
		t.Fatal("recorded M1-a baseline rejected")
	}
	changed := append([]map[string]any{}, rows...)
	first := make(map[string]any, len(rows[0]))
	for key, value := range rows[0] {
		first[key] = value
	}
	first["exit"] = json.Number("0")
	changed[0] = first
	if preApprovalM1aBaseline(copyDir, changed) {
		t.Fatal("mutated initial ledger row accepted")
	}
	path := filepath.Join(copyDir, "startup", "car010.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("mutated\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if preApprovalM1aBaseline(copyDir, rows) {
		t.Fatal("mutated M1-a raw file accepted")
	}
}

// TestCodexPreApprovalDiscriminatorLive is M1-b. It is intentionally never
// invoked by M1-a's test selection; the lead controls its explicit -run gate.
func TestCodexPreApprovalDiscriminatorLive(t *testing.T) {
	f := preApprovalRunStartup(t, true)
	discDir := filepath.Join(f.evidenceDir, "discriminator")
	manifest, paths, err := preApprovalReadExports(discDir)
	if err != nil {
		t.Fatal(err)
	}
	diffBytes, err := os.ReadFile(filepath.Join(discDir, "arm-diff.txt"))
	if err != nil || !preApprovalOnlyProjectConfigDiff(string(diffBytes)) {
		t.Fatalf("arm diff: %v", err)
	}
	var diffMeta struct {
		WrittenNS int64 `json:"written_ns"`
	}
	meta, err := os.ReadFile(filepath.Join(discDir, "arm-diff.meta.json"))
	if err != nil || json.Unmarshal(meta, &diffMeta) != nil || diffMeta.WrittenNS < manifest.ExportedNS {
		t.Fatal("arm diff meta is stale")
	}
	ledgerPath := filepath.Join(f.evidenceDir, "ledger.json")
	rows, err := preApprovalReadLedger(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	_, authPath := linkedCodexHome(t, f.root)
	authBefore := fileSHA256OrEmpty(authPath)
	if authBefore == "" {
		t.Fatal("operator auth unreadable")
	}
	version := codexRoleVersion(t, f.procs, f.codexBin, []string{"PATH=" + os.Getenv("PATH"), "CODEX_HOME=" + f.codexHome})
	if version != "0.157.0" {
		t.Fatalf("Codex version %q differs from 0.157.0", version)
	}
	evidence := preApprovalDiscEvidence{CodexVersion: version, Attempt: f.attempt, ExportComplete: true,
		ArmDiffOK: true, StartupChecks: map[string]string{}, Arms: map[string]preApprovalArmResult{},
		AuthSHA256Before: authBefore}
	windowStart := time.Now()
	stop := func(reason string) {
		rows = append(rows, map[string]any{"kind": "stop", "fixture": "disc", "reason": reason})
		if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("ABORTED: %s", reason)
	}
	for _, arm := range []string{"control", "treatment"} {
		fixture := "disc-" + arm
		home, project, argvBytes, prompt, rootState, err := preApprovalArmInputs(discDir, arm)
		if err != nil {
			stop(err.Error())
		}
		buildText, err := os.ReadFile(filepath.Join(discDir, "inputs", arm, "moai-build.txt"))
		if err != nil || !bytes.Contains(buildText, []byte("sha256="+f.moaiHash+"\n")) {
			stop("moai binary differs from exported build")
		}
		repoText, err := os.ReadFile(filepath.Join(discDir, "inputs", arm, "repo.txt"))
		fixtureHead, gitErr := f.git("-C", f.root, "rev-parse", "HEAD")
		if err != nil || gitErr != nil || string(repoText) != "is_repo=true\nhead="+strings.TrimSpace(fixtureHead)+"\n" {
			stop("fixture git HEAD differs from exported repo")
		}
		freshManifest, freshPaths, err := preApprovalReadExports(discDir)
		if err != nil || freshManifest.ExportedNS != manifest.ExportedNS {
			stop("export manifest changed before LIVE")
		}
		paths = freshPaths
		exportID := fmt.Sprintf("export-%d", freshManifest.ExportedNS)
		exportSnapshot, err := os.ReadFile(filepath.Join(discDir, "exports", exportID, "export-manifest.json"))
		if err != nil {
			stop(err.Error())
		}
		currentExport, err := os.ReadFile(filepath.Join(discDir, "export-manifest.json"))
		if err != nil || !bytes.Equal(exportSnapshot, currentExport) {
			stop("attempt export differs from current manifest")
		}
		exportHash := preApprovalSHA(exportSnapshot)
		args := preApprovalDiscriminatorArgs(f.root, string(prompt))
		if !preApprovalArgvMatches(f.root, prompt, argvBytes) {
			stop("exported argv differs from invocation")
		}
		if !bytes.Equal(prompt, []byte(preApprovalDiscriminatorPrompt(f.root))) {
			stop("exported prompt differs from invocation")
		}
		hashes := map[string]string{"codex_home_config": preApprovalSHA(home), "project_config": preApprovalSHA(project), "root_state": preApprovalSHA(rootState)}
		latest := preApprovalLatestStartupRow(rows, fixture)
		if latest == nil {
			stop("startup ledger row missing")
		}
		startupID, _ := latest["startup_id"].(string)
		proof, launches, err := preApprovalStartupFromFiles(f.evidenceDir, fixture, startupID)
		if err != nil {
			stop(err.Error())
		}
		if _, ok := preApprovalLatestStartup(rows, fixture, manifest.ExportedNS, f.tempRoot, f.root, exportID, exportHash, hashes, proof, launches); !ok {
			stop("startup ledger, proof, or exports differ")
		}
		evidence.StartupChecks[arm] = "PASS"
		if time.Since(windowStart) >= preApprovalDiscWindow || evidence.Invocations >= 2 {
			stop("discriminator budget exceeded")
		}
		allLive := 0
		for _, row := range rows {
			if row["kind"] == "live" {
				allLive++
			}
		}
		if allLive >= preApprovalMaxLiveCalls {
			stop("absolute LIVE budget exceeded")
		}
		if err := preApprovalPrepareArm(t, f, authPath, arm, home, project, rootState); err != nil {
			stop(err.Error())
		}
		if fileSHA256OrEmpty(authPath) != authBefore {
			stop("auth changed before LIVE")
		}
		callBound := preApprovalLiveCallBound
		if remain := preApprovalDiscWindow - time.Since(windowStart); remain < callBound {
			callBound = remain
		}
		if callBound <= 0 {
			stop("discriminator window exhausted")
		}
		ctx, cancel := context.WithTimeout(context.Background(), callBound)
		cmd := liveCommand(ctx, f.codexBin, args...)
		cmd.Dir = f.root
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "CODEX_HOME=" + f.codexHome, "MOAI_HOME=" + t.TempDir(), "LANG=C"}
		var stdout, stderr []byte
		var runErr error
		start := time.Now().UnixNano()
		gateRows := preApprovalGate(paths, string(diffBytes), proof, func() error {
			stdout, stderr, runErr = preApprovalRunSplit(f.procs, cmd)
			return nil // a tool refusal may make Codex exit nonzero; the call still counts
		})
		end := time.Now().UnixNano()
		cancel()
		if len(gateRows) != 1 || gateRows[0].Kind != "live" {
			stop("LIVE gate refused invocation")
		}
		exit := liveExitCode(runErr)
		row := map[string]any{"kind": "live", "fixture": fixture, "attempt": evidence.Attempt,
			"startup_id": startupID,
			"export_id":  exportID, "export_sha256": exportHash,
			"temp_root": f.tempRoot, "fixture_root": f.root, "fixture_sentinel": true,
			"started_ns": start, "ended_ns": end, "argv": args,
			"exit": exit, "bound_by": "test", "auth_sha256_before": authBefore,
			"auth_sha256_after": fileSHA256OrEmpty(authPath),
			"inputs_sha256": map[string]string{"codex_home_config": preApprovalSHA(home),
				"project_config": preApprovalSHA(project), "argv": preApprovalSHA(argvBytes),
				"prompt": preApprovalSHA(prompt), "root_state": preApprovalSHA(rootState),
				"moai_binary": f.moaiHash}}
		rows = append(rows, row)
		if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
			t.Fatal(err)
		}
		evidence.Invocations++
		if err := os.WriteFile(filepath.Join(discDir, arm+".jsonl"), stdout, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(discDir, arm+".err"), stderr, 0o600); err != nil {
			t.Fatal(err)
		}
		r := preApprovalClassify(stdout)
		// A rejected destination never starts a child. Launch records are the
		// launcher's persistent proof of a spawned child.
		r.LaunchRecords = len(codexAuditRecordNames(f.root))
		r.ChildProcesses = r.LaunchRecords
		evidence.Arms[arm] = r
		if row["auth_sha256_after"] != authBefore {
			stop("auth changed after LIVE")
		}
		if r.UserTurns != 1 || r.MoaiRoleAuditCalls < 1 || r.LaunchRecords != 0 {
			stop("discriminator precondition failed")
		}
		if arm == "control" && r.Outcome != "REFUSED" {
			stop("control refusal was not reproduced")
		}
		if arm == "treatment" && r.Outcome != "REFUSED" && r.Outcome != "RAN" {
			stop("treatment did not produce a classified result")
		}
	}
	evidence.AuthSHA256After = fileSHA256OrEmpty(authPath)
	for _, cmd := range f.procs.cmds {
		if cmd.Process != nil {
			evidence.RecordedPIDs = append(evidence.RecordedPIDs, cmd.Process.Pid)
		}
	}
	evidence.KilledPIDs = f.procs.reap()
	if err := os.WriteFile(filepath.Join(discDir, "outcome.txt"), []byte(evidence.Arms["treatment"].Outcome+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	emitLiveEvidence(t, discDir, "evidence.json", "CPP005", "", evidence)
}
