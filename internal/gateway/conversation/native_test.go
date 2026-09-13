package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeCompletedTranscriptIsDiscoveredOnExactResume(t *testing.T) {
	m := newTestManager(t)
	cwd := filepath.Join(m.root, "project")
	d, err := m.New(context.Background(), NewRequest{CWD: cwd, Project: cwd})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(d.ConfigDir, "projects", "encoded-project", d.UUID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	raw := fmt.Sprintf("{\"type\":\"last-prompt\",\"sessionId\":%q}\n{\"type\":\"file-history-snapshot\",\"messageId\":\"u1\",\"snapshot\":{}}\n{\"type\":\"user\",\"sessionId\":%q,\"cwd\":%q,\"message\":{\"role\":\"user\",\"content\":\"hello\"}}\n{\"type\":\"assistant\",\"sessionId\":%q,\"cwd\":%q,\"message\":{\"role\":\"assistant\",\"model\":\"gpt-6-astra\",\"stop_reason\":\"end_turn\",\"content\":[{\"type\":\"text\",\"text\":\"answer\"}]}}\n", d.UUID, d.UUID, cwd, d.UUID, cwd)
	if err = os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	resumed, err := m.Resume(context.Background(), d.UUID)
	if err != nil {
		t.Fatal(err)
	}
	fork, err := m.Fork(context.Background(), d.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if fork.Model != "gpt-6-astra" {
		t.Fatal("fork lost parent selected model")
	}
	if resumed.Model != "gpt-6-astra" {
		t.Fatal("native model not retained")
	}
	if resumed.Transcript == "" {
		t.Fatal("native transcript not recorded")
	}
	reopened, err := Open(m.root)
	if err != nil {
		t.Fatal(err)
	}
	continued, err := reopened.Continue(context.Background(), cwd)
	if err != nil || continued.UUID != d.UUID {
		t.Fatal("new process lost native conversation", err)
	}
}

func TestNativeIncompleteAndForeignRowsNeverRegister(t *testing.T) {
	for _, mutation := range []string{"error", "foreign", "unfinished", "later-user"} {
		t.Run(mutation, func(t *testing.T) {
			m := newTestManager(t)
			cwd := filepath.Join(m.root, "project")
			d, err := m.New(context.Background(), NewRequest{CWD: cwd, Project: cwd})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(d.ConfigDir, "projects", "encoded", d.UUID+".jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			sid := d.UUID
			stop := "end_turn"
			failure := mutation == "error"

			if mutation == "foreign" {
				sid = "foreign"
			}
			if mutation == "unfinished" {
				stop = "tool_use"
			}
			raw := fmt.Sprintf("{\"type\":\"assistant\",\"sessionId\":%q,\"cwd\":%q,\"isApiErrorMessage\":%t,\"message\":{\"role\":\"assistant\",\"model\":\"gpt-6-astra\",\"stop_reason\":%q}}\n", sid, cwd, failure, stop)
			if mutation == "later-user" {
				raw += fmt.Sprintf("{\"type\":\"user\",\"sessionId\":%q,\"cwd\":%q}\n", sid, cwd)
			}
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = m.Resume(context.Background(), d.UUID); err == nil {
				t.Fatal("unverified native completion resumed")
			}
		})
	}
}

func TestNativeTranscriptRejectsMalformedIdentityAndPaths(t *testing.T) {
	m := newTestManager(t)
	cwd := filepath.Join(m.root, "project")
	d, err := m.New(context.Background(), NewRequest{CWD: cwd, Project: cwd})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(d.ConfigDir, "native.jsonl")
	r := record{UUID: d.UUID, CWD: cwd, ConfigDir: d.ConfigDir, Project: cwd, Completion: 1, Transcript: "native.jsonl"}
	for _, raw := range []string{`{`, `{}`, fmt.Sprintf(`{"type":"assistant","sessionId":%q,"cwd":"foreign"}`, d.UUID), fmt.Sprintf(`{"type":"assistant","sessionId":%q,"session_id":"foreign","cwd":%q}`, d.UUID, cwd), fmt.Sprintf(`{"type":"assistant","sessionId":%q,"cwd":%q,"message":false}`, d.UUID, cwd), fmt.Sprintf(`{"type":"assistant","sessionId":%q,"cwd":%q,"message":{"role":"user","model":"gpt-6-astra"}}`, d.UUID, cwd), fmt.Sprintf(`{"type":"user","sessionId":%q}`, d.UUID)} {
		if err = os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = transcriptModel(r); err == nil {
			t.Error("malformed native transcript accepted")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err = transcriptModel(r); err == nil {
		t.Fatal("missing transcript accepted")
	}
	external := filepath.Join(m.root, "outside.jsonl")
	if err := os.WriteFile(external, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(external, path); err != nil {
		t.Fatal(err)
	}
	if _, err = transcriptModel(r); err == nil {
		t.Fatal("symlink transcript accepted")
	}
	if err = m.refreshNative(context.Background(), "unknown"); err != ErrMissing {
		t.Fatal("unknown UUID searched", err)
	}
}

func TestNativeDiscoveryRejectsDuplicateAndSymlinkCandidates(t *testing.T) {
	for _, kind := range []string{"duplicate", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			m := newTestManager(t)
			cwd := filepath.Join(m.root, "project")
			d, err := m.New(context.Background(), NewRequest{CWD: cwd, Project: cwd})
			if err != nil {
				t.Fatal(err)
			}
			a := filepath.Join(d.ConfigDir, "projects", "a")
			b := filepath.Join(d.ConfigDir, "projects", "b")
			if err := os.MkdirAll(a, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(b, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(a, d.UUID+".jsonl"), []byte(`{}`), 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "duplicate" {
				if err := os.WriteFile(filepath.Join(b, d.UUID+".jsonl"), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err = os.Symlink(a, filepath.Join(b, "link")); err != nil {
					t.Fatal(err)
				}
			}
			if err = m.refreshNative(context.Background(), d.UUID); err == nil {
				t.Fatal("ambiguous/linked native input accepted")
			}
		})
	}
}

func TestNativeResumeAfterTerminalAPIError(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		prefix   []string
		mutation string
		want     bool
	}{
		{"completed-then-failed-request", []string{"success", "user", "api-error"}, "", true},
		{"local-command-after-api-error", []string{"success", "user", "api-error", "meta", "local-command", "local-output", "meta"}, "", true},
		{"typed-local-command-spoof", []string{"success", "user", "api-error", "typed-command"}, "", false},
		{"repeated-api-error", []string{"success", "user", "api-error", "api-error"}, "", true},
		{"error-only", []string{"user", "api-error"}, "", false},
		{"unfinished-tool", []string{"success", "user", "tool", "user", "api-error"}, "", false},
		{"unfinished-assistant", []string{"success", "user", "partial", "api-error"}, "", false},
		{"arbitrary-synthetic", []string{"success", "user", "api-error"}, "unflagged", false},
		{"foreign-error", []string{"success", "user", "api-error"}, "foreign", false},
		{"wrong-role", []string{"success", "user", "api-error"}, "role", false},
		{"error-with-tool-content", []string{"success", "user", "api-error"}, "tool-content", false},
		{"later-incomplete-request", []string{"success", "user", "api-error", "user"}, "", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			m := newTestManager(t)
			cwd := filepath.Join(m.root, "project")
			d, err := m.New(context.Background(), NewRequest{CWD: cwd, Project: cwd})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(d.ConfigDir, "projects", "encoded", d.UUID+".jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			var raw []byte
			for _, kind := range scenario.prefix {
				message := map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn", "content": []any{map[string]any{"type": "text", "text": "answer"}}}
				row := map[string]any{"type": "assistant", "sessionId": d.UUID, "cwd": cwd, "message": message}
				switch kind {
				case "user":
					row["type"] = "user"
					message["role"] = "user"
				case "meta":
					row["type"] = "user"
					row["isMeta"] = true
					message["role"] = "user"
					message["content"] = "native local metadata"
				case "local-command", "typed-command", "local-output":
					row["type"] = "user"
					message["role"] = "user"
					message["content"] = "<command-name>/context</command-name>\n<command-message>context</command-message>\n<command-args></command-args>"
					if kind == "local-output" {
						message["content"] = "<local-command-stdout>context statistics</local-command-stdout>"
					}
					if kind == "typed-command" {
						row["promptSource"] = "typed"
					}
				case "tool":
					message["stop_reason"] = "tool_use"
				case "partial":
					message["stop_reason"] = nil
				case "api-error":
					row["isApiErrorMessage"] = true
					row["error"] = "unknown"
					message["model"] = "<synthetic>"
					message["stop_reason"] = "stop_sequence"
					switch scenario.mutation {
					case "unflagged":
						delete(row, "isApiErrorMessage")
					case "foreign":
						row["sessionId"] = "foreign"
					case "role":
						message["role"] = "user"
					case "tool-content":
						message["content"] = []any{map[string]any{"type": "tool_use", "id": "call_1"}}
					}
				}
				b, err := json.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				raw = append(raw, b...)
				raw = append(raw, '\n')
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			resumed, err := m.Resume(context.Background(), d.UUID)
			if scenario.want {
				if err != nil {
					t.Fatalf("completed conversation could not recover after API failure: %v", err)
				}
				if resumed.UUID != d.UUID || resumed.Model != "gpt-5.6-sol" {
					t.Fatalf("lost successful identity/model: %+v", resumed)
				}
			} else if err == nil {
				t.Fatal("unsafe incomplete conversation resumed")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(raw) {
				t.Fatal("resume rewrote native transcript")
			}
		})
	}
}
