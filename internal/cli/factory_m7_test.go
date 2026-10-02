package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-SD-021 — every legacy role spelling is refused where a lane label, role
// token, or leader label is expected, with the REQ-RNC-003/-005/-007 message
// naming the canonical form — on `moai codex` as on `moai cc` and `moai glm`
// (cc and glm share parseLauncherEntry, so its rows cover both launchers) —
// in the lane-identity variable the lane verbs read, and via the negative
// source scan: no user-facing factory string carries `worker` or `agent` as a
// role or `lead` as the leader noun.
func TestSD_AC021_LegacySpellingsRefused(t *testing.T) {
	t.Run("launcher parse (moai cc / moai glm)", func(t *testing.T) {
		rows := []struct {
			args     []string
			wantMsg  []string // substrings the RNC message must carry
			anyError bool     // refused, message shape unchecked (position expects no such label)
		}{
			{args: []string{"-f", "worker"}, wantMsg: []string{"legacy role token", "-f lane"}},
			{args: []string{"-f", "agent"}, wantMsg: []string{"legacy role token", "-f lane"}},
			{args: []string{"-f", "Worker"}, wantMsg: []string{"legacy role token", "-f lane"}},
			{args: []string{"-f", "worker-1"}, wantMsg: []string{"legacy lane label", "lane-1"}},
			{args: []string{"-f", "agent-2"}, wantMsg: []string{"legacy lane label", "lane-2"}},
			{args: []string{"-f", "Worker-1"}, wantMsg: []string{"legacy lane label", "lane-1"}},
			// `lead` at the -f value position: the position expects a lane
			// value, so the leader label is not expected here — the parse
			// still refuses (usage error naming the accepted forms).
			{args: []string{"-f", "lead"}, anyError: true},
			// The leader label IS expected at the --name position.
			{args: []string{"--name", "lead"}, wantMsg: []string{"legacy leader spelling", "leader"}},
			{args: []string{"--name", "lead-3"}, wantMsg: []string{"legacy leader spelling", "leader-3"}},
		}
		for _, row := range rows {
			_, err := parseLauncherEntry(row.args)
			if err == nil {
				t.Errorf("args %v: parse succeeded, want a legacy refusal", row.args)
				continue
			}
			for _, want := range row.wantMsg {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("args %v: error %q does not name %q", row.args, err, want)
				}
			}
			if row.anyError && strings.Contains(err.Error(), "legacy") {
				t.Errorf("args %v: unexpected legacy-message shape at a position that expects no such label: %q", row.args, err)
			}
		}
	})

	t.Run("moai codex", func(t *testing.T) {
		prevDirect, prevSpawn := codexDirectLaunchFn, codexSpawnLaunchFn
		codexDirectLaunchFn = func(*exec.Cmd) error {
			t.Error("the direct launch ran; the legacy refusal must start no child")
			return nil
		}
		codexSpawnLaunchFn = func(string, string, []string, []string) error {
			t.Error("the spawn launch ran; the legacy refusal must start no child")
			return nil
		}
		t.Cleanup(func() { codexDirectLaunchFn, codexSpawnLaunchFn = prevDirect, prevSpawn })

		rows := []struct {
			value   string
			wantMsg []string
		}{
			{value: "worker", wantMsg: []string{"legacy role token", "moai codex -f lane"}},
			{value: "agent", wantMsg: []string{"legacy role token", "moai codex -f lane"}},
			{value: "WORKER", wantMsg: []string{"legacy role token", "moai codex -f lane"}},
			{value: "worker-1", wantMsg: []string{"legacy lane label", "moai codex -f lane"}},
			{value: "agent-1", wantMsg: []string{"legacy lane label", "moai codex -f lane"}},
			{value: "lead", wantMsg: []string{"legacy leader spelling", "moai cc -f"}},
			{value: "lead-1", wantMsg: []string{"legacy leader spelling", "moai cc -f"}},
		}
		for _, row := range rows {
			_, errB, err := runCodexCmd(t, "-f", row.value)
			if err == nil {
				t.Errorf("-f %s: codex accepted a legacy factory token", row.value)
				continue
			}
			if strings.Contains(errB, codexFactoryRefusalDiag) {
				t.Errorf("-f %s: refused with the REQ-SD-004 line; AC-SD-021 wants the REQ-RNC message naming the canonical form:\n%s", row.value, errB)
			}
			for _, want := range row.wantMsg {
				if !strings.Contains(errB, want) {
					t.Errorf("-f %s: stderr %q does not name %q", row.value, errB, want)
				}
			}
			code, ok := ResolveExitCode(err)
			if !ok || code != 1 {
				t.Errorf("-f %s: exit code = (%d, %v), want (1, true); err=%v", row.value, code, ok, err)
			}
		}
	})

	t.Run("lane verbs refuse a legacy lane-identity label", func(t *testing.T) {
		t.Chdir(t.TempDir())
		for _, label := range []string{"worker-1", "agent-1", "worker", "agent", "lead"} {
			for _, verb := range []struct {
				args []string
			}{
				{args: []string{"next"}},
				{args: []string{"stage", "t1", "run"}},
				{args: []string{"complete", "t1"}},
			} {
				sdLaneEnv(t, label, "")
				_, errB, err := runFactory(t, append([]string{verb.args[0]}, verb.args[1:]...)...)
				if err == nil {
					t.Errorf("label %q, %s: verb ran with a legacy lane-identity label", label, verb.args[0])
					continue
				}
				combined := errB + err.Error()
				if !strings.Contains(combined, "legacy") {
					t.Errorf("label %q, %s: refusal %q does not name the legacy spelling", label, verb.args[0], combined)
				}
				if label == "worker-1" || label == "agent-1" {
					if !strings.Contains(combined, "lane-1") {
						t.Errorf("label %q, %s: refusal %q does not name the canonical %q", label, verb.args[0], combined, "lane-1")
					}
				}
			}
		}
	})

	t.Run("verbs take no lane-label argument", func(t *testing.T) {
		t.Chdir(t.TempDir())
		for _, arg := range [][]string{{"next", "worker-1"}, {"status", "agent-1"}} {
			if _, _, err := runFactory(t, arg...); err == nil {
				t.Errorf("args %v: verb accepted a positional legacy token; the verbs take no lane-label argument", arg)
			}
		}
	})

	t.Run("no legacy role nouns in factory user-facing strings", func(t *testing.T) {
		// The scan is defined by the factory files, not the whole package:
		// kanban-mode strings that legitimately use these words outside role
		// semantics are out of scope. The allowlist entries are each a
		// sanctioned non-role literal, and every entry must still match a
		// real literal (a stale entry fails the test, so it cannot rot).
		allow := map[string]string{
			// factory.go refuseLegacyEntryNames: the legacy-leader detection
			// literal (REQ-RNC-009 — legacy tokens survive as detection
			// values); it never reaches a user-facing string as a role.
			"factory.go:lead": "legacy-leader detection literal (TrimPrefix input, REQ-RNC-009)",
			// codex_launcher.go: "agent" names the generated Codex agent
			// config TOMLs, not a factory role.
			"codex_launcher.go:the auth provider, the project wiring, the generated agent TOMLs, and\n": "non-role noun (Codex agent config files)",
		}
		used := map[string]bool{}
		roleNoun := regexp.MustCompile(`\b(worker|agent|lead)\b`)
		files := []string{"codex_launcher.go"}
		for _, pattern := range []string{"factory*.go", "mcp_factory*.go"} {
			matches, err := filepath.Glob(filepath.Join(".", pattern))
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, matches...)
		}
		scanned := 0
		for _, file := range files {
			base := filepath.Base(file)
			if strings.HasSuffix(base, "_test.go") {
				continue
			}
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			fset := token.NewFileSet()
			parsed, err := parser.ParseFile(fset, file, src, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", file, err)
			}
			ast.Inspect(parsed, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					return true
				}
				if !roleNoun.MatchString(value) {
					return true
				}
				scanned++
				key := base + ":" + value
				if _, allowed := allow[key]; allowed {
					used[key] = true
					return true
				}
				t.Errorf("%s:%d: factory string carries a legacy role noun: %q (not in the scan allowlist)",
					file, fset.Position(lit.Pos()).Line, value)
				return true
			})
		}
		if scanned == 0 {
			t.Fatal("the scan matched no literal at all — a zero-result pass asserts nothing (positive control failed)")
		}
		for key, reason := range allow {
			if !used[key] {
				t.Errorf("stale allowlist entry %q (%s): no literal matched it; update the key", key, reason)
			}
		}
	})
}

// sdSchemaTokens are the schema-DDL markers AC-SD-022 freezes (the CREATE
// TABLE / ALTER TABLE / index statements).
var sdSchemaTokens = []string{"CREATE TABLE", "CREATE INDEX", "CREATE UNIQUE INDEX", "ALTER TABLE"}

// sdContainsSchemaDDL reports whether a string literal carries schema DDL.
func sdContainsSchemaDDL(v string) bool {
	for _, tok := range sdSchemaTokens {
		if strings.Contains(v, tok) {
			return true
		}
	}
	return false
}

// sdExtractSchemaStatements walks a parsed Go file and returns every string
// literal whose value carries schema DDL, unquoted, in source order.
func sdExtractSchemaStatements(t *testing.T, fset *token.FileSet, file string, src []byte) []string {
	t.Helper()
	parsed, err := parser.ParseFile(fset, file, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	var out []string
	ast.Inspect(parsed, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		if sdContainsSchemaDDL(value) {
			out = append(out, value)
		}
		return true
	})
	return out
}

// sdSnapshotSchemaStatements is the schema-statement set pinned from the
// absorption baseline, byte-identical to the production string literals of
// internal/homestate, internal/factorymsg, and internal/kanban on the
// t1240 absorption tree (develop 145c3d98c absorbed; the a7190891d run-start
// pin moved here because the absorbed hold-state and transition-stamp cards
// legitimately evolved the items/archived schema and added the items_new
// rebuild migration). Extracted with sdExtractSchemaStatements against the
// production files, sorted (baseline attribution: verification-claim-integrity
// §2; the regeneration command is recorded in the card's absorption report).
var sdSnapshotSchemaStatements = []string{
	"\n);\nCREATE TABLE IF NOT EXISTS findings (\n  subject_id TEXT  NOT NULL,\n  related_id TEXT  NOT NULL,\n  relation   TEXT  NOT NULL,\n  source     TEXT  NOT NULL,\n  score      REAL  NOT NULL,\n  note       TEXT  NOT NULL DEFAULT '',\n  at         TEXT  NOT NULL\n);\nCREATE TABLE IF NOT EXISTS archived_items (\n  seq      INTEGER PRIMARY KEY,\n  id       TEXT    NOT NULL UNIQUE,\n  text     TEXT    NOT NULL,\n  added_at TEXT    NOT NULL,\n  spec_id  TEXT,\n  state    TEXT    NOT NULL,\n  position INTEGER NOT NULL\n);\nCREATE TABLE IF NOT EXISTS archived_findings (\n  archive_seq INTEGER NOT NULL,\n  position    INTEGER NOT NULL,\n  subject_id  TEXT  NOT NULL,\n  related_id  TEXT  NOT NULL,\n  relation    TEXT  NOT NULL,\n  source      TEXT  NOT NULL,\n  score       REAL  NOT NULL,\n  note        TEXT  NOT NULL DEFAULT '',\n  at          TEXT  NOT NULL\n);\nCREATE INDEX IF NOT EXISTS idx_items_state ON items(state);\n",
	"\nCREATE TABLE IF NOT EXISTS dispatches(project_key TEXT NOT NULL, run_id TEXT NOT NULL, dispatch_id TEXT NOT NULL, card_id TEXT NOT NULL, lane_slot TEXT NOT NULL, attempt INTEGER NOT NULL, assignee_generation INTEGER NOT NULL, state TEXT NOT NULL, result_attempt INTEGER NOT NULL DEFAULT 0, result_digest TEXT NOT NULL DEFAULT '', result_ref TEXT NOT NULL DEFAULT '', updated_at TEXT NOT NULL, PRIMARY KEY(project_key,run_id,dispatch_id));\n",
	"\nCREATE TABLE IF NOT EXISTS gtd_meta (\n  key   TEXT PRIMARY KEY,\n  value TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS gtd_items (\n  item_id          TEXT PRIMARY KEY,\n  content          TEXT NOT NULL,\n  source           TEXT NOT NULL,\n  sensitivity      TEXT NOT NULL,\n  event_id         TEXT NOT NULL UNIQUE,\n  outcome          TEXT NOT NULL DEFAULT '',\n  completion_evidence TEXT NOT NULL DEFAULT '',\n  disposition      TEXT NOT NULL DEFAULT '',\n  class            TEXT NOT NULL DEFAULT '',\n  action_context   TEXT NOT NULL DEFAULT '',\n  review_at        TEXT NOT NULL DEFAULT '',\n  authority        TEXT NOT NULL DEFAULT '',\n  source_trust     TEXT NOT NULL DEFAULT '',\n  status           TEXT NOT NULL,\n  source_revision  INTEGER NOT NULL,\n  card_id          TEXT,\n  cancelled        INTEGER NOT NULL DEFAULT 0 CHECK (cancelled IN (0,1))\n);\nCREATE TABLE IF NOT EXISTS gtd_relations (\n  subject_id       TEXT NOT NULL,\n  object_id        TEXT NOT NULL,\n  kind             TEXT NOT NULL,\n  note             BLOB NOT NULL DEFAULT X'',\n  source           TEXT NOT NULL,\n  assertion_status TEXT NOT NULL,\n  source_revision  INTEGER NOT NULL,\n  policy_version   TEXT NOT NULL,\n  PRIMARY KEY(subject_id, object_id, kind)\n);\nCREATE TABLE IF NOT EXISTS gtd_contracts (\n  mission_id       TEXT PRIMARY KEY,\n  policy_version   TEXT NOT NULL,\n  contract_hash    TEXT NOT NULL,\n  contract_json    BLOB NOT NULL,\n  approved         INTEGER NOT NULL DEFAULT 0 CHECK (approved IN (0,1))\n);\nCREATE TABLE IF NOT EXISTS gtd_missions (\n  mission_id       TEXT PRIMARY KEY,\n  mission_mode     TEXT NOT NULL,\n  state            TEXT NOT NULL,\n  contract_hash    TEXT NOT NULL,\n  snapshot_hash    TEXT NOT NULL,\n  owner_id         TEXT NOT NULL DEFAULT '',\n  lease_version    INTEGER NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS gtd_events (\n  event_id         TEXT PRIMARY KEY,\n  mission_id       TEXT NOT NULL DEFAULT '',\n  kind             TEXT NOT NULL,\n  payload_hash     TEXT NOT NULL,\n  created_at       TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS gtd_operations (\n  operation_id     TEXT PRIMARY KEY,\n  mission_id       TEXT NOT NULL,\n  action           TEXT NOT NULL,\n  target           TEXT NOT NULL,\n  state            TEXT NOT NULL,\n  receipt_json     BLOB NOT NULL,\n  snapshot_hash    TEXT NOT NULL,\n  updated_at       TEXT NOT NULL\n);\nCREATE INDEX IF NOT EXISTS idx_gtd_items_card ON gtd_items(card_id);\nCREATE INDEX IF NOT EXISTS idx_gtd_relations_object ON gtd_relations(object_id);\nCREATE INDEX IF NOT EXISTS idx_gtd_operations_mission ON gtd_operations(mission_id);\n",
	"\nCREATE TABLE IF NOT EXISTS lane_handoff_receipts(id TEXT PRIMARY KEY, handoff_id TEXT NOT NULL UNIQUE, slot TEXT NOT NULL, nonce TEXT NOT NULL, card_id TEXT NOT NULL, spec_id TEXT NOT NULL, old_session TEXT NOT NULL, old_generation INTEGER NOT NULL, session_uuid TEXT NOT NULL, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, created_at TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS lane_dispatch_releases(handoff_id TEXT PRIMARY KEY, slot TEXT NOT NULL, to_session TEXT NOT NULL, to_generation INTEGER NOT NULL, message_count INTEGER NOT NULL, released_at TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS lane_message_releases(message_id TEXT PRIMARY KEY, handoff_id TEXT NOT NULL, from_session TEXT NOT NULL, from_generation INTEGER NOT NULL, to_session TEXT NOT NULL, to_generation INTEGER NOT NULL, claim_token TEXT NOT NULL, released_at TEXT NOT NULL);\n",
	"\nCREATE TABLE IF NOT EXISTS lane_handoffs(id TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, slot TEXT NOT NULL, card_id TEXT NOT NULL, spec_id TEXT NOT NULL, mode TEXT NOT NULL, nonce TEXT NOT NULL UNIQUE, handoff_generation INTEGER NOT NULL, source_backend TEXT NOT NULL, source_role TEXT NOT NULL, source_session TEXT NOT NULL, source_generation INTEGER NOT NULL, source_pid INTEGER NOT NULL, source_process_start TEXT NOT NULL, develop_pin TEXT NOT NULL, target_path TEXT NOT NULL, target_branch TEXT NOT NULL, state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);\nCREATE UNIQUE INDEX IF NOT EXISTS lane_handoffs_one_open ON lane_handoffs(slot) WHERE state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS');\nCREATE TABLE IF NOT EXISTS lane_handoff_events(id INTEGER PRIMARY KEY AUTOINCREMENT, handoff_id TEXT NOT NULL, from_state TEXT NOT NULL, to_state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS lane_handoff_relocations(handoff_id TEXT PRIMARY KEY, nonce TEXT NOT NULL, method TEXT NOT NULL, source_thread_id TEXT NOT NULL, thread_id TEXT NOT NULL, forked_from_id TEXT NOT NULL, thread_started INTEGER NOT NULL, request_cwd TEXT NOT NULL, response_cwd TEXT NOT NULL, readback_cwd TEXT NOT NULL, readback_branch TEXT NOT NULL, readback_head TEXT NOT NULL, recorded_at TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS lane_endpoint_tombstones(slot TEXT NOT NULL, session_uuid TEXT NOT NULL, generation INTEGER NOT NULL, replaced_by_session TEXT NOT NULL, replaced_by_generation INTEGER NOT NULL, handoff_id TEXT NOT NULL, bound_at TEXT NOT NULL, PRIMARY KEY(slot,session_uuid,generation));\n",
	"\nCREATE TABLE IF NOT EXISTS meta (\n  key   TEXT PRIMARY KEY,\n  value TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS items (",
	"\nCREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS workers (\n  label TEXT PRIMARY KEY,\n  pid INTEGER NOT NULL,\n  backend TEXT NOT NULL DEFAULT '',\n  session_id TEXT NOT NULL DEFAULT '',\n  run_id TEXT NOT NULL DEFAULT '',\n  registered_at TEXT NOT NULL,\n  heartbeat_at TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS runs (\n  run_id TEXT PRIMARY KEY,\n  lead_session_id TEXT NOT NULL DEFAULT '',\n  lead_backend TEXT NOT NULL DEFAULT '',\n  status TEXT NOT NULL,\n  manifest_json TEXT NOT NULL DEFAULT '{}',\n  lead_pid INTEGER NOT NULL DEFAULT 0,\n  lead_process_start TEXT NOT NULL DEFAULT '',\n  lane_capacity INTEGER NOT NULL DEFAULT 1,\n  created_at TEXT NOT NULL,\n  updated_at TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS cards (\n  run_id TEXT NOT NULL,\n  card_id TEXT NOT NULL,\n  owner_label TEXT NOT NULL DEFAULT '',\n  state TEXT NOT NULL,\n  version INTEGER NOT NULL DEFAULT 1,\n  evidence_path TEXT NOT NULL DEFAULT '',\n  updated_at TEXT NOT NULL,\n  stage TEXT NOT NULL DEFAULT '',\n  lease_holder TEXT NOT NULL DEFAULT '',\n  lease_expires_at TEXT NOT NULL DEFAULT '',\n  heartbeat_at TEXT NOT NULL DEFAULT '',\n  decision_gate TEXT NOT NULL DEFAULT '',\n  decision_question TEXT NOT NULL DEFAULT '',\n  decision_resume TEXT NOT NULL DEFAULT '',\n  decider TEXT NOT NULL DEFAULT '',\n  decided_at TEXT NOT NULL DEFAULT '',\n  failure_reason TEXT NOT NULL DEFAULT '',\n  hint_prefer TEXT NOT NULL DEFAULT '',\n  hint_after TEXT NOT NULL DEFAULT '',\n  spec_id TEXT NOT NULL DEFAULT '',\n  worktree_path TEXT NOT NULL DEFAULT '',\n  evidence_sha TEXT NOT NULL DEFAULT '',\n  merge_sha TEXT NOT NULL DEFAULT '',\n  merge_tree TEXT NOT NULL DEFAULT '',\n  remeasure_path TEXT NOT NULL DEFAULT '',\n  contract_spec_id TEXT NOT NULL DEFAULT '',\n  contract_sha256 TEXT NOT NULL DEFAULT '',\n  contract_signed_at TEXT NOT NULL DEFAULT '',\n  contract_event TEXT NOT NULL DEFAULT '',\n  PRIMARY KEY(run_id, card_id)\n);\nCREATE TABLE IF NOT EXISTS events (\n  seq INTEGER PRIMARY KEY AUTOINCREMENT,\n  run_id TEXT NOT NULL DEFAULT '',\n  kind TEXT NOT NULL,\n  payload_json TEXT NOT NULL DEFAULT '{}',\n  created_at TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS dead_letters (\n  id INTEGER PRIMARY KEY AUTOINCREMENT,\n  run_id TEXT NOT NULL DEFAULT '',\n  envelope_json TEXT NOT NULL,\n  error TEXT NOT NULL,\n  created_at TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS resume_handoffs (\n  id INTEGER PRIMARY KEY AUTOINCREMENT,\n  status TEXT NOT NULL CHECK(status IN ('pending','claimed','consumed','failed','expired','cleared')),\n  schema_version INTEGER NOT NULL,\n  spec_id TEXT NOT NULL DEFAULT '',\n  phase TEXT NOT NULL DEFAULT '',\n  saved_at TEXT NOT NULL,\n  saved_by_session TEXT NOT NULL DEFAULT '',\n  conversation_language TEXT NOT NULL DEFAULT '',\n  directives_json TEXT NOT NULL DEFAULT '{}',\n  embedded_goal_json TEXT,\n  body TEXT NOT NULL,\n  body_sha256 TEXT NOT NULL,\n  claim_token TEXT NOT NULL DEFAULT '',\n  claimed_at TEXT,\n  claim_expires_at TEXT,\n  claim_owner_pid INTEGER,\n  claim_owner_session TEXT NOT NULL DEFAULT '',\n  claim_owner_fingerprint TEXT NOT NULL DEFAULT '',\n  legacy_recovery INTEGER NOT NULL DEFAULT 0,\n  legacy_recovery_reason TEXT NOT NULL DEFAULT '',\n  consumed_at TEXT,\n  error TEXT NOT NULL DEFAULT ''\n);\nCREATE UNIQUE INDEX IF NOT EXISTS one_pending_resume_handoff\nON resume_handoffs(status) WHERE status='pending';\nCREATE TABLE IF NOT EXISTS memory_handoffs (\n  id INTEGER PRIMARY KEY AUTOINCREMENT,\n  status TEXT NOT NULL CHECK(status IN ('pending','persisted','failed','expired')),\n  sprint TEXT NOT NULL,\n  spec TEXT NOT NULL,\n  result_status TEXT NOT NULL,\n  body TEXT NOT NULL,\n  index_line TEXT NOT NULL,\n  supersedes TEXT NOT NULL DEFAULT '',\n  body_sha256 TEXT NOT NULL,\n  created_at TEXT NOT NULL,\n  persisted_at TEXT,\n  error TEXT NOT NULL DEFAULT ''\n);\nCREATE UNIQUE INDEX IF NOT EXISTS one_pending_memory_handoff\nON memory_handoffs(status) WHERE status='pending';\nCREATE TABLE IF NOT EXISTS handoff_events (\n  seq INTEGER PRIMARY KEY AUTOINCREMENT,\n  flow TEXT NOT NULL,\n  handoff_id INTEGER NOT NULL,\n  from_status TEXT NOT NULL DEFAULT '',\n  to_status TEXT NOT NULL,\n  detail TEXT NOT NULL DEFAULT '',\n  created_at TEXT NOT NULL\n);\n",
	"\nCREATE TABLE IF NOT EXISTS peers(slot TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, backend TEXT NOT NULL, role TEXT NOT NULL, session_uuid TEXT NOT NULL UNIQUE, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, updated_at TEXT NOT NULL);\nCREATE TABLE IF NOT EXISTS messages(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, sender_slot TEXT NOT NULL DEFAULT '', UNIQUE(project_key,run_id,sender_slot,idem_key));\nCREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at);\nCREATE TABLE IF NOT EXISTS dead_letters(id INTEGER PRIMARY KEY AUTOINCREMENT, message_id TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL);\n",
	"\nCREATE TABLE IF NOT EXISTS todo_runtime_runs (\n run_id TEXT PRIMARY KEY, backend TEXT NOT NULL, manifest_json TEXT NOT NULL\n);\nCREATE TABLE IF NOT EXISTS todo_runtime_assignments (\n run_id TEXT NOT NULL, card_id TEXT NOT NULL, owner_label TEXT NOT NULL,\n reported_state TEXT NOT NULL, event_kind TEXT NOT NULL, provenance_json TEXT NOT NULL,\n PRIMARY KEY(run_id, card_id)\n);\nINSERT INTO meta(key,value) VALUES('runtime_schema_version','1')\n ON CONFLICT(key) DO NOTHING;\n",
	"ALTER TABLE %s ADD COLUMN %s TEXT",
	"ALTER TABLE cards ADD COLUMN ",
	"ALTER TABLE items_new RENAME TO items",
	"ALTER TABLE messages_v2 RENAME TO messages",
	"ALTER TABLE profile_leases ADD COLUMN transfer_deadline TEXT NOT NULL DEFAULT ''",
	"ALTER TABLE resume_handoffs ADD COLUMN claim_expires_at TEXT",
	"ALTER TABLE resume_handoffs ADD COLUMN claim_owner_fingerprint TEXT NOT NULL DEFAULT ''",
	"ALTER TABLE resume_handoffs ADD COLUMN claim_owner_pid INTEGER",
	"ALTER TABLE resume_handoffs ADD COLUMN claim_owner_session TEXT NOT NULL DEFAULT ''",
	"ALTER TABLE resume_handoffs ADD COLUMN legacy_recovery INTEGER NOT NULL DEFAULT 0",
	"ALTER TABLE resume_handoffs ADD COLUMN legacy_recovery_reason TEXT NOT NULL DEFAULT ''",
	"ALTER TABLE runs ADD COLUMN lane_capacity INTEGER NOT NULL DEFAULT 1",
	"ALTER TABLE runs ADD COLUMN lead_pid INTEGER NOT NULL DEFAULT 0",
	"ALTER TABLE runs ADD COLUMN lead_process_start TEXT NOT NULL DEFAULT ''",
	"CREATE INDEX IF NOT EXISTS idx_items_state ON items(state)",
	"CREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at)",
	"CREATE INDEX IF NOT EXISTS resume_handoff_claim_expiry ON resume_handoffs(status,claim_expires_at,id)",
	"CREATE INDEX IF NOT EXISTS resume_handoff_claim_expiry ON resume_handoffs(status,claim_expires_at,id)",
	"CREATE TABLE IF NOT EXISTS profile_leases(token TEXT PRIMARY KEY,profile_name TEXT NOT NULL,profile_path TEXT NOT NULL,project_key TEXT NOT NULL DEFAULT '',pid INTEGER NOT NULL,process_fingerprint TEXT NOT NULL DEFAULT '',session_id TEXT NOT NULL DEFAULT '',state TEXT NOT NULL CHECK(state IN ('provisional','transferring','enriched','released')),transfer_deadline TEXT NOT NULL DEFAULT '',created_at TEXT NOT NULL,updated_at TEXT NOT NULL); CREATE INDEX IF NOT EXISTS profile_lease_path_state ON profile_leases(profile_path,state)",
	"CREATE TABLE items_new (%s)",
	"CREATE TABLE messages_v2(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, sender_slot TEXT NOT NULL DEFAULT '', UNIQUE(project_key,run_id,sender_slot,idem_key))",
	"CREATE TABLE todo_identities(\n  entity_kind TEXT NOT NULL,\n  local_id TEXT NOT NULL,\n  uuid TEXT NOT NULL UNIQUE,\n  PRIMARY KEY(entity_kind,local_id)\n)",
}

// AC-SD-022 — every schema statement (CREATE TABLE, ALTER TABLE, index) in
// the production files of internal/homestate, internal/factorymsg, and
// internal/kanban is byte-identical to the absorption baseline pinned above
// (REQ-SD-022). The statement set is pinned with its provenance; the current
// tree is extracted with the same walker and compared as a sorted multiset,
// so an added, modified, or removed statement all fail.
func TestSD_AC022_SchemaStatementsFrozen(t *testing.T) {
	dirs := []string{"../homestate", "../factorymsg", "../kanban"}
	var got []string
	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 0 {
			t.Fatalf("no Go files under %s — the scan would sweep nothing", dir)
		}
		for _, file := range matches {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			got = append(got, sdExtractSchemaStatements(t, token.NewFileSet(), file, src)...)
		}
	}
	if len(got) == 0 {
		t.Fatal("the scan found no schema statement at all — a zero-result pass asserts nothing (positive control failed)")
	}
	want := append([]string(nil), sdSnapshotSchemaStatements...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		gotSet := map[string]int{}
		for _, s := range got {
			gotSet[s]++
		}
		wantSet := map[string]int{}
		for _, s := range want {
			wantSet[s]++
		}
		for s, n := range wantSet {
			if gotSet[s] != n {
				t.Errorf("schema statement drifted from the absorption-baseline pin (want %d, got %d):\n%s", n, gotSet[s], s)
			}
		}
		for s, n := range gotSet {
			if wantSet[s] == 0 {
				t.Errorf("schema statement ADDED since the absorption-baseline pin (%d occurrence(s)):\n%s", n, s)
			}
		}
	}
}

// AC-SD-018 — the full lane cycle of AC-SD-006 leaves the parent checkout
// untouched: `git status --porcelain` (excluding the worktrees directory the
// cycle itself provisions) is empty, and HEAD and the checked-out branch
// equal their pre-cycle values.
func TestSD_AC018_ParentCheckoutUntouched(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the §B fixture family is POSIX-verified (the TestSD_AC006 sibling)")
	}
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")
	fcGit(t, root, "branch", "develop")
	integWT := filepath.Join(root, sessionWorktreeSubdir, "develop")
	fcGit(t, root, "worktree", "add", "-q", integWT, "develop")
	// The git-flow integration branch is configured by a TRACKED
	// .moai/config/sections/git-strategy.yaml (the real parent carries it as
	// a tracked config file), so the fixture commits it rather than leaving
	// it as parent dirt.
	sdGitFlowDevelop(t, root)
	// -f: todoFixture's git-flow precondition seed keeps .moai/config/ out of
	// git status through .git/info/exclude (card t1453).
	fcGit(t, root, "add", "-f", filepath.Join(".moai", "config", "sections", "git-strategy.yaml"))
	fcGit(t, root, "commit", "-q", "-m", "git-flow config")
	// The gitignore mirrors the real checkout's treatment of the paths the
	// cycle touches: .moai/state/ and .moai/reports/* are ignored there
	// (.gitignore:235/:398), and .moai/db/ never exists in a real checkout —
	// it is the temp-root DB layout, which a real run keeps under the MoAI
	// home, outside the parent. .moai/worktrees/ stays UNignored exactly as
	// in the real repo, which is why the AC's judgement excludes it.
	if err := os.WriteFile(filepath.Join(root, ".gitignore"),
		[]byte(".moai/cache/\n.moai/db/\n.moai/logs/\n.moai/reports/*\n.moai/state/\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", ".gitignore")
	fcGit(t, root, "commit", "-q", "-m", "runtime-state gitignore")
	if remotes := fcGit(t, root, "remote"); remotes != "" {
		t.Fatalf("fixture carries a remote: %s", remotes)
	}

	preHead := fcGit(t, root, "rev-parse", "HEAD")
	preBranch := fcGit(t, root, "branch", "--show-current")

	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next: %v", err)
	}
	card := fcCard(t, root, "t1")
	if card.WorktreePath == "" {
		t.Fatal("next recorded no card worktree")
	}
	if err := os.WriteFile(filepath.Join(card.WorktreePath, "done.txt"), []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, card.WorktreePath, "add", "-A")
	fcGit(t, card.WorktreePath, "commit", "-q", "-m", "card work")
	runSHA := fcGit(t, card.WorktreePath, "rev-parse", "HEAD")

	if _, _, err := runFactory(t, "stage", "t1", "run", "--run", fcRun); err != nil {
		t.Fatalf("stage run: %v", err)
	}
	if _, _, err := runFactory(t, "stage", "t1", "sync", runSHA, "--run", fcRun); err != nil {
		t.Fatalf("stage sync: %v", err)
	}
	if _, _, err := runFactory(t, "stage", "t1", "sync-audit", runSHA+":done.txt", "--run", fcRun); err != nil {
		t.Fatalf("stage sync-audit: %v", err)
	}
	verdictAbs := filepath.Join(card.WorktreePath, ".moai", "reports", "t1", "sync-audit.md")
	if err := os.MkdirAll(filepath.Dir(verdictAbs), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(verdictAbs, []byte("verdict: PASS\naudited_sha: "+runSHA+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runFactory(t, "stage", "t1", "merge-ready", "--run", fcRun); err != nil {
		t.Fatalf("stage merge-ready: %v", err)
	}

	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}

	if c := fcCard(t, root, "t1"); c.State != homestate.CardMergedLocal {
		t.Fatalf("t1 = %s, want merged-local (the cycle must reach its end before the parent is judged)", c.State)
	}

	// The parent judgement: porcelain (untracked files enumerated, so a
	// collapsed `?? .moai/` cannot hide the culprit) excluding the worktrees
	// directory is empty, and HEAD/branch equal the pre-cycle values.
	porcelain := fcGit(t, root, "status", "--porcelain", "--untracked-files=all")
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path := line
		if i := strings.Index(line, " "); i >= 0 {
			path = strings.TrimSpace(line[i+1:])
		}
		if strings.HasPrefix(path, ".moai/worktrees/") {
			continue
		}
		t.Errorf("parent checkout is dirty outside the worktrees directory: %q", line)
	}
	if postHead := fcGit(t, root, "rev-parse", "HEAD"); postHead != preHead {
		t.Errorf("parent HEAD moved during the cycle: pre=%s post=%s", preHead, postHead)
	}
	if postBranch := fcGit(t, root, "branch", "--show-current"); postBranch != preBranch {
		t.Errorf("parent branch changed during the cycle: pre=%s post=%s", preBranch, postBranch)
	}
}
