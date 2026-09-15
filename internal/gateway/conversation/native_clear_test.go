package conversation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeClearRequiresPrivateCommandCompletion(t *testing.T) {
	for _, mutation := range []string{"valid", "pasted", "foreign", "wrong-parent", "sidechain", "missing-completion", "symlink"} {
		t.Run(mutation, func(t *testing.T) {
			root, _ := filepath.EvalSymlinks(t.TempDir())
			if err := os.Chmod(root, 0700); err != nil {
				t.Fatal(err)
			}
			cwd := "/project"
			id := "11111111-1111-4111-8111-111111111111"
			path := filepath.Join(root, "projects", "project", id+".jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			command := map[string]any{"type": "user", "uuid": "command", "sessionId": id, "cwd": cwd, "isSidechain": false, "message": map[string]any{"role": "user", "content": "<command-name>/clear</command-name>\n <command-message>clear</command-message>\n <command-args></command-args>"}}
			completion := map[string]any{"type": "system", "subtype": "local_command", "parentUuid": "command", "sessionId": id, "cwd": cwd, "isSidechain": false, "content": "<local-command-stdout></local-command-stdout>"}
			switch mutation {
			case "pasted":
				command["origin"] = "user"
			case "foreign":
				command["cwd"] = "/other"
			case "wrong-parent":
				completion["parentUuid"] = "other"
			case "sidechain":
				command["isSidechain"] = true
			}
			a, _ := json.Marshal(command)
			b, _ := json.Marshal(completion)
			raw := append(a, '\n')
			if mutation != "missing-completion" {
				raw = append(raw, b...)
				raw = append(raw, '\n')
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if mutation == "symlink" {
				if err := os.Rename(path, path+".real"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path+".real", path); err != nil {
					t.Fatal(err)
				}
			}
			err := ValidateNativeClear(context.Background(), root, cwd, id)
			if (err == nil) != (mutation == "valid") {
				t.Fatalf("mutation %s: %v", mutation, err)
			}
		})
	}
}

func TestNativeClearObservedReadOnly(t *testing.T) {
	config := os.Getenv("MOAI_TEST_NATIVE_CLEAR_CONFIG")
	if config == "" {
		t.Skip("set explicit read-only native clear fixture")
	}
	if err := ValidateNativeClear(context.Background(), config, os.Getenv("MOAI_TEST_NATIVE_CLEAR_CWD"), os.Getenv("MOAI_TEST_NATIVE_CLEAR_ID")); err != nil {
		t.Fatal(err)
	}
}
