// launcher_teammode_t2013_test.go — card t2013: removeGLMEnv (the `moai cc`
// settings cleanup) deletes the `teammateMode` override only when the same
// settings.local.json carries GLM env residue. A file without residue is the
// operator's own configuration: it is left byte-identical, never rewritten, and
// the VALUE of teammateMode is never read or judged.
//
// Every fixture is a settings file under t.TempDir(); nothing outside it is
// touched.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func teamModeSettingsPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), ".claude", "settings.local.json")
}

func teamModeWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// (a) A file with GLM env residue and a teammateMode override: the residue and
// the override are both removed, and an unrelated key survives.
func TestT2013RemoveGLMEnvResidueClearsTeammateMode(t *testing.T) {
	for _, mode := range []string{"tmux", "in-process", "glm"} {
		t.Run(mode, func(t *testing.T) {
			path := teamModeSettingsPath(t)
			body := `{"teammateMode":"` + mode + `","env":{"` + config.EnvAnthropicBaseURL + `":"https://glm.invalid","` +
				config.EnvAnthropicDefaultOpusModel + `":"m","CUSTOM_VAR":"keep"},"outputStyle":"x"}`
			teamModeWrite(t, path, body)

			if err := removeGLMEnv(path); err != nil {
				t.Fatalf("removeGLMEnv: %v", err)
			}

			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("parse result: %v\n%s", err, raw)
			}
			if _, present := got["teammateMode"]; present {
				t.Errorf("teammateMode survived a cleanup that had GLM residue to remove: %s", raw)
			}
			env, _ := got["env"].(map[string]any)
			if _, present := env[config.EnvAnthropicBaseURL]; present {
				t.Errorf("GLM residue %s survived: %s", config.EnvAnthropicBaseURL, raw)
			}
			if _, present := env[config.EnvAnthropicDefaultOpusModel]; present {
				t.Errorf("GLM residue %s survived: %s", config.EnvAnthropicDefaultOpusModel, raw)
			}
			if env["CUSTOM_VAR"] != "keep" || got["outputStyle"] != "x" {
				t.Errorf("an unrelated key was lost: %s", raw)
			}
		})
	}
}

// (b) No residue: the file is byte-identical after the call, whatever the value
// of teammateMode is (including values the launcher never writes and
// non-string values) — the value is never judged.
func TestT2013RemoveGLMEnvNoResidueLeavesFileByteIdentical(t *testing.T) {
	bodies := map[string]string{
		"in-process":     `{ "teammateMode" : "in-process" , "env":{"CUSTOM_VAR":"1"} }`,
		"tmux":           `{"teammateMode":"tmux"}`,
		"auto":           `{"teammateMode":"auto","outputStyle":"y"}`,
		"unexpected":     `{"teammateMode":"no-such-mode-zz","env":{}}`,
		"non-string":     `{"teammateMode":42,"env":{"CUSTOM_VAR":"1"}}`,
		"mode-only-user": "{\n\t\"teammateMode\": \"tmux\",\n\t\"env\": {\"CUSTOM_VAR\": \"1\"}\n}\n",
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			path := teamModeSettingsPath(t)
			teamModeWrite(t, path, body)

			if err := removeGLMEnv(path); err != nil {
				t.Fatalf("removeGLMEnv: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, []byte(body)) {
				t.Errorf("a file with no GLM residue was rewritten:\nbefore: %q\nafter:  %q", body, after)
			}
		})
	}
}

// (c) An absent file or an empty one stays as it is: no file, no directory and
// no lock file is created for the absent path, and an empty file stays empty.
func TestT2013RemoveGLMEnvAbsentAndEmptyFile(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		path := teamModeSettingsPath(t)
		if err := removeGLMEnv(path); err != nil {
			t.Fatalf("removeGLMEnv on an absent file: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("settings file exists after the call (stat err: %v), want it absent", err)
		}
		if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
			t.Errorf("settings directory exists after the call (stat err: %v), want it absent", err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		path := teamModeSettingsPath(t)
		teamModeWrite(t, path, "")
		if err := removeGLMEnv(path); err != nil {
			t.Fatalf("removeGLMEnv on an empty file: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 0 {
			t.Errorf("empty file now has %d bytes, want 0", info.Size())
		}
	})
}

// (d) Idempotent: after a cleanup that had residue, a second call changes
// nothing and writes nothing (same bytes, same modification time).
func TestT2013RemoveGLMEnvSecondCallIsANoOp(t *testing.T) {
	path := teamModeSettingsPath(t)
	teamModeWrite(t, path, `{"teammateMode":"tmux","env":{"`+config.EnvAnthropicBaseURL+`":"https://glm.invalid","CUSTOM_VAR":"keep"}}`)

	if err := removeGLMEnv(path); err != nil {
		t.Fatalf("first removeGLMEnv: %v", err)
	}
	firstBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	firstInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := removeGLMEnv(path); err != nil {
		t.Fatalf("second removeGLMEnv: %v", err)
	}
	secondBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	secondInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Errorf("the second call changed the file:\nfirst:  %q\nsecond: %q", firstBytes, secondBytes)
	}
	if !firstInfo.ModTime().Equal(secondInfo.ModTime()) {
		t.Errorf("the second call rewrote the file: mtime %v → %v", firstInfo.ModTime(), secondInfo.ModTime())
	}
}
