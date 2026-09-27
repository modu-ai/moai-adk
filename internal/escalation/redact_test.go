package escalation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// fakeSecret is a credential-shaped string that must never reach disk. The
// vendor prefix is concatenated so no scanner-triggering literal is committed.
const fakeSecret = "gh" + "p_" + "FAKEFAKEFAKEFAKE0123456789abcdef"

// fakeSecretCore is the digit-free part of fakeSecret. The class 8
// diagnostic key rewrites digit runs, so a leaked secret can reach disk
// changed; probing for the core still catches it.
const fakeSecretCore = "FAKEFAKEFAKEFAKE"

// leakedSurfaces names every file the detector writes for the card — the
// escalation records, the card log, and the card state file — that contains
// the secret.
func leakedSurfaces(t *testing.T, w *escalationtest.Worktree, secret string) []string {
	t.Helper()
	paths := map[string]string{}
	dir := escalation.RecordDir(w.Root, w.Card)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		paths["record "+e.Name()] = filepath.Join(dir, e.Name())
	}
	files, err := escalation.CardFilesFor(w.Root, w.Card)
	if err != nil {
		t.Fatal(err)
	}
	paths["card log"], paths["state file"] = files.Log, files.State
	var leaked []string
	for name, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if strings.Contains(string(data), secret) {
			leaked = append(leaked, name)
		}
	}
	return leaked
}

// Class 6 and class 8 records name the observed command. A credential in the
// command (URL userinfo, an Authorization header, a token=/password=
// assignment) is masked before anything is written, and the class still trips.
func TestCommandCredentialsAreMasked(t *testing.T) {
	t.Run("class6-url-userinfo", func(t *testing.T) {
		isolateStore(t)
		w := armedWith(t, "t9001", nil, "")
		escalation.Observe(contractSettings(t, w), escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root,
			ToolName: "Bash", Command: "git push https://user:" + fakeSecret + "@github.com/o/r.git main"})
		if rs, _ := recordsOfClass(t, w, escalation.ClassIrreversibleAction); len(rs) != 1 {
			t.Fatalf("class 6 did not trip: %d records", len(rs))
		}
		if l := leakedSurfaces(t, w, fakeSecretCore); len(l) != 0 {
			t.Errorf("secret written to: %v", l)
		}
	})
	cmds := map[string]string{
		"header":   "curl -H 'Authorization: Bearer " + fakeSecret + "' https://api.example.com/x",
		"token":    "deploy --token=" + fakeSecret + " now",
		"password": "PASSWORD=" + fakeSecret + " ./run.sh",
	}
	for name, cmd := range cmds {
		t.Run("class8-"+name, func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			for i := 0; i < 3; i++ {
				// The diagnostic echoes the credential too; the key must stay
				// stable across attempts so the streak still reaches 3.
				postBash(t, w, cmd, true, "error: auth failed for token="+fakeSecret+" after 1"+strings.Repeat("0", i)+"ms")
			}
			rs, _ := recordsOfClass(t, w, escalation.ClassSameDiagnosticRepeat)
			if len(rs) != 1 {
				t.Fatalf("class 8 did not trip: %d records", len(rs))
			}
			if l := leakedSurfaces(t, w, fakeSecretCore); len(l) != 0 {
				t.Errorf("secret written to: %v", l)
			}
		})
	}
}

// Re-audit N2: class 6 classifies the raw command and masks only what it
// renders or hashes. A push-develop-authorized push whose refspec source
// carries a credential word, or whose URL carries a credential, writes no
// record and leaks nothing.
func TestAuthorizedPushNotMisclassifiedByMask(t *testing.T) {
	for _, cmd := range []string{
		"git push origin WT-token-rotation:develop",
		"git push https://user:" + fakeSecret + "@github.com/o/r.git HEAD:develop",
	} {
		t.Run(cmd[:24], func(t *testing.T) {
			isolateStore(t)
			w := armedWith(t, "t9001", nil, "")
			escalation.Observe(contractSettings(t, w), escalation.Event{Hook: escalation.HookPreToolUse, CWD: w.Root,
				ToolName: "Bash", Command: cmd})
			if rs, raw := recordsOfClass(t, w, escalation.ClassIrreversibleAction); len(rs) != 0 {
				t.Errorf("authorized push tripped class 6:\n%s", strings.Join(raw, "\n"))
			}
			if l := leakedSurfaces(t, w, fakeSecretCore); len(l) != 0 {
				t.Errorf("secret written to: %v", l)
			}
		})
	}
}

// Re-audit N4: one case per credential shape the mask missed. Every fake
// value is obviously fake; vendor prefixes are concatenated at compile time.
func TestMaskCommandCoversCredentialShapes(t *testing.T) {
	const pw = "FAKEPASSWORDVALUE"
	const core = "FAKEFAKEFAKEFAKE"
	cases := []struct {
		name, cmd, secret, keep string
	}{
		{"mysql-attached-p", "mysql -uroot -p" + pw + " db", pw, " db"},
		{"curl-u", "curl -u alice:" + pw + " https://api.example.com", pw, "alice"},
		{"docker-login-p", "docker login -p " + pw + " registry.example.com", pw, "registry.example.com"},
		{"json-password", `curl -d '{"password":"` + pw + `"}' https://api.example.com`, pw, `"password"`},
		{"anthropic-prefix", "echo " + "sk" + "-ant-" + core + core, core, "echo"},
		{"aws-access-key-id", "aws s3 ls " + "AK" + "IA" + core, core, "aws s3 ls"},
		{"underscore-key-assignment", "ANTHROPIC_KEY=" + pw + " ./run", pw, "./run"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := escalation.MaskCommand(tc.cmd)
			if strings.Contains(got, tc.secret) || !strings.Contains(got, "***") {
				t.Errorf("MaskCommand(%q) = %q, secret not masked", tc.cmd, got)
			}
			if !strings.Contains(got, tc.keep) {
				t.Errorf("MaskCommand(%q) = %q, lost the non-secret part %q", tc.cmd, got, tc.keep)
			}
		})
	}
}

// Masking is deterministic: two different secrets in the same position give
// the same masked command, so a streak keyed on it is not split.
func TestMaskCommandIsStable(t *testing.T) {
	a := escalation.MaskCommand("git push https://u:" + fakeSecret + "@h/r.git")
	b := escalation.MaskCommand("git push https://u:other-secret-value@h/r.git")
	if a != b || strings.Contains(a, fakeSecret) {
		t.Errorf("masked forms differ or leak: %q vs %q", a, b)
	}
	if got := escalation.MaskCommand("go test ./..."); got != "go test ./..." {
		t.Errorf("a command without credentials changed: %q", got)
	}
}
