package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestGPTProductionAppServerWiring(t *testing.T) {
	home, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	seedGatewayAuthStore(t, filepath.Join(home, "gateway-auth"))
	binDir, protocolLog := installProductionWiringFakeCodex(t)
	verifyProductionWiringFakeCodex(t, binDir, protocolLog)
	startSharedGPTWiringFixture(t, home, protocolLog)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	models := gatewayGPTModels()
	for _, model := range gatewayGPTModels() {
		model := model
		t.Run(model.RouteID, func(t *testing.T) {
			if model.AuthMethod != gateway.AuthAppServer {
				t.Errorf("production GPT route %s auth method=%q, want %q managed App Server wiring", model.RouteID, model.AuthMethod, gateway.AuthAppServer)
			}
		})
	}
	payload, err := marshalGatewayPrivatePayload("private-session", models)
	if err != nil {
		t.Fatal(err)
	}
	h, err := productionGatewayHandlerFactory(payload)
	if err != nil {
		t.Fatalf("production gateway assembly: %v", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = h.(io.Closer).Close()
		}
	}()
	t.Run("concrete-adapter", func(t *testing.T) {
		owned, ok := h.(*gatewayOwnedHandler)
		if !ok {
			t.Fatalf("production handler concrete type=%T, want *gatewayOwnedHandler", h)
		}
		serverValue := reflect.ValueOf(owned.Handler).Elem()
		adapters := serverValue.FieldByName("adapters")
		var concrete string
		iter := adapters.MapRange()
		for iter.Next() {
			value := iter.Value()
			if value.Kind() == reflect.Interface && !value.IsNil() {
				concrete = value.Elem().Type().String()
			}
		}
		if concrete != "*gateway.AppServerAdapter" {
			t.Errorf("production OpenAI adapter concrete type=%q, want %q", concrete, "*gateway.AppServerAdapter")
			return
		}
		raw, err := os.ReadFile(protocolLog)
		if err != nil || !strings.Contains(string(raw), `"method":"initialize"`) && !strings.Contains(string(raw), `"method": "initialize"`) || !strings.Contains(string(raw), `"method":"account/read"`) && !strings.Contains(string(raw), `"method": "account/read"`) {
			t.Errorf("managed App Server initialization log=%q err=%v, want initialize and account/read", string(raw), err)
		}
		if err := h.(io.Closer).Close(); err != nil {
			t.Errorf("managed App Server clean close: %v", err)
			return
		}
		closed = true
	})
	if !closed {
		if err := h.(io.Closer).Close(); err != nil {
			t.Fatalf("production handler close before managed-only probe: %v", err)
		}
		closed = true
	}

	t.Run("managed-without-legacy-store", func(t *testing.T) {
		isolated, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(isolated, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(isolated, "gateway-auth"), []byte("FORBIDDEN-LEGACY-STORE"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("MOAI_HOME", isolated)
		startSharedGPTWiringFixture(t, isolated, protocolLog)
		managed, err := productionGatewayHandlerFactory(payload)
		if err != nil || managed == nil {
			t.Errorf("managed App Server assembly touched absent/poisoned legacy auth store: handler=%T err=%v, want success with zero legacy open/read/refresh", managed, err)
			return
		}
		_ = managed.(io.Closer).Close()
	})
}

func installProductionWiringFakeCodex(t *testing.T) (string, string) {
	t.Helper()
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("protocol fake python3: %v", err)
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatalf("protocol fake bash: %v", err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "protocol.jsonl")
	source := `import sys,json
log_path=` + strconv.Quote(logPath) + `
def emit(x): print(json.dumps(x),flush=True)
for line in sys.stdin:
 with open(log_path,'a') as f: f.write(line)
 m=json.loads(line); method=m.get('method')
 if method=='initialize': emit({'id':m['id'],'result':{'userAgent':'production-wiring-fake'}})
 elif method=='initialized': pass
 elif method=='account/read': emit({'id':m['id'],'result':{'account':{'type':'chatgpt','planType':'test'}}})
 elif method=='model/list': emit({'id':m['id'],'result':{'data':[],'nextCursor':None}})
 else: emit({'id':m['id'],'error':{'code':-32601,'message':'unsupported fixture method'}})
`
	impl := filepath.Join(dir, "fake-codex.py")
	if err := os.WriteFile(impl, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	name := "codex"
	body := "#!" + bash + "\nexec " + strconv.Quote(python) + " " + strconv.Quote(impl) + " \"$@\"\n"
	if runtime.GOOS == "windows" {
		name = "codex.bat"
		body = "@echo off\r\n\"" + python + "\" \"" + impl + "\" %*\r\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return dir, logPath
}

func verifyProductionWiringFakeCodex(t *testing.T, dir, logPath string) {
	t.Helper()
	binary := filepath.Join(dir, "codex")
	if runtime.GOOS == "windows" {
		binary += ".bat"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := codexapp.Start(ctx, codexapp.Config{Binary: binary, Home: dir})
	if err != nil {
		t.Fatalf("protocol fake start: %v", err)
	}
	if _, err := client.Initialize(ctx, "production-wiring-probe", "1"); err != nil {
		_ = client.Close()
		t.Fatalf("protocol fake initialize: %v", err)
	}
	if _, err := client.Account(ctx); err != nil {
		_ = client.Close()
		t.Fatalf("protocol fake account/read: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("protocol fake close: %v", err)
	}
	// A second Start on the same private profile can acquire the lease only
	// after Client.Close has killed and reaped the first process.
	reaped, err := codexapp.Start(ctx, codexapp.Config{Binary: binary, Home: dir})
	if err != nil {
		t.Fatalf("protocol fake process/lease not reaped after Close: %v", err)
	}
	if _, err := reaped.Initialize(ctx, "production-wiring-reap-probe", "1"); err != nil {
		_ = reaped.Close()
		t.Fatalf("protocol fake reacquire initialize: %v", err)
	}
	if err := reaped.Close(); err != nil {
		t.Fatalf("protocol fake reacquire close: %v", err)
	}
	raw, err := os.ReadFile(logPath)
	if err != nil || !strings.Contains(string(raw), `"method":"initialize"`) && !strings.Contains(string(raw), `"method": "initialize"`) || !strings.Contains(string(raw), `"method":"account/read"`) && !strings.Contains(string(raw), `"method": "account/read"`) {
		t.Fatalf("protocol fake probe log=%q err=%v, want initialize/account-read", string(raw), err)
	}
	if err := os.Remove(logPath); err != nil {
		t.Fatal(err)
	}
}

func TestGPTFactoryAndRetiredKanbanContract(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		wantWorker   string
		retired      bool
		wantDispatch bool
	}{
		{name: "lead", args: []string{"-f", "2"}, wantDispatch: true},
		{name: "worker", args: []string{"-f", "lane-2"}, wantWorker: "lane-2"},
		{name: "dispatch", args: []string{"-f"}, wantDispatch: true},
		{name: "agent-prompt-forwarding", args: []string{"-f", "--", "-p", "dispatch one card through Agent and return its tool result"}, wantDispatch: true},
		{name: "alias-effort", args: []string{"-f", "--model", "fable"}, wantDispatch: true},
		{name: "retired-short", args: []string{"-k"}, retired: true},
		{name: "retired-long", args: []string{"--kanban"}, retired: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("MOAI_HOME", t.TempDir())
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Setenv(config.EnvClaudeCodeEffortLevel, "high")
			t.Setenv("MOAI_DISPATCH_BACKEND", "")
			t.Setenv(config.EnvMoaiKanbanBackend, "")
			envBefore := captureFactoryEnvState()
			before := captureGPTFactoryState(t, root)
			launches := 0
			dispatchBackend, legacyBackend, worker, forwarded := "", "", "", []string(nil)
			var stdout, stderr bytes.Buffer
			cmd := newGPTCommand(gptCommandServices{Launch: func(_ string, mode string, args []string) error {
				launches++
				if mode == "gpt" {
					dispatchBackend = os.Getenv("MOAI_DISPATCH_BACKEND")
					legacyBackend = os.Getenv(config.EnvMoaiKanbanBackend)
				}
				worker = os.Getenv(config.EnvMoaiFactoryWorker)
				forwarded = append([]string(nil), args...)
				return nil
			}})
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			err := cmd.RunE(cmd, tc.args)
			envAfter := captureFactoryEnvState()
			after := captureGPTFactoryState(t, root)
			if tc.retired {
				if err == nil {
					t.Errorf("%s exit code = 0, want non-zero before side effects", strings.Join(tc.args, " "))
				}
				combined := err.Error() + stdout.String() + stderr.String()
				if !strings.Contains(combined, "-f") {
					t.Errorf("%s stdout=%q stderr=%q, want retired-mode guidance containing -f", strings.Join(tc.args, " "), stdout.String(), stderr.String())
				}
				if envBefore != envAfter {
					t.Errorf("%s process env changed before=%+v after=%+v, want new/legacy backend env unchanged even without Launch", strings.Join(tc.args, " "), envBefore, envAfter)
				}
				if launches != 0 || before.todoHash != after.todoHash || before.tasksHash != after.tasksHash || before.dispatches != after.dispatches || before.factoryRuns != after.factoryRuns {
					t.Errorf("%s side effects launch=%d dispatch-env=%q legacy-kanban-env=%q todo=%s→%s tasks=%s→%s dispatch=%d→%d factory=%d→%d, want all unchanged/zero/empty", strings.Join(tc.args, " "), launches, dispatchBackend, legacyBackend, before.todoHash, after.todoHash, before.tasksHash, after.tasksHash, before.dispatches, after.dispatches, before.factoryRuns, after.factoryRuns)
				}
				return
			}
			if err != nil {
				t.Fatalf("gpt factory %s: %v", tc.name, err)
			}
			if launches != 1 {
				t.Errorf("gpt factory process/launch count=%d, want 1", launches)
			}
			if dispatchBackend != "gpt" || legacyBackend != "" {
				t.Errorf("gpt Factory backend env dispatch=%q legacy-kanban=%q, want gpt/empty", dispatchBackend, legacyBackend)
			}
			if worker != tc.wantWorker {
				t.Errorf("gpt factory worker: got %q, want %q", worker, tc.wantWorker)
			}
			if tc.wantDispatch && after.factoryRuns <= before.factoryRuns {
				t.Errorf("gpt Factory dispatch runs=%d→%d, want positive change", before.factoryRuns, after.factoryRuns)
			}
			if tc.name == "agent-prompt-forwarding" && !slicesContainPair(forwarded, "-p", "dispatch one card through Agent and return its tool result") {
				t.Errorf("gpt Factory launch prompt args=%v, want -p preserved at launch boundary; Agent/tool execution is an M14-R5 live gate", forwarded)
			}
			if before.tasksHash != after.tasksHash {
				t.Errorf("gpt Factory unit launch changed Tasks snapshot %s→%s, want unchanged; positive Tasks/Dispatch execution is an M14-R5 live gate", before.tasksHash, after.tasksHash)
			}
			if tc.name == "alias-effort" && (!slicesContainPair(forwarded, "--model", "fable") || os.Getenv(config.EnvClaudeCodeEffortLevel) != "high") {
				t.Errorf("gpt Factory alias/effort args=%v effort=%q, want fable/high", forwarded, os.Getenv(config.EnvClaudeCodeEffortLevel))
			}
			combined := stdout.String() + stderr.String()
			if strings.Contains(strings.ToLower(combined), "kanban") {
				t.Errorf("active GPT Factory stdout/stderr contains retired token: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

type factoryEnvState struct {
	DispatchValue, LegacyValue string
	DispatchSet, LegacySet     bool
}

func captureFactoryEnvState() factoryEnvState {
	dispatch, dispatchSet := os.LookupEnv("MOAI_DISPATCH_BACKEND")
	legacy, legacySet := os.LookupEnv(config.EnvMoaiKanbanBackend)
	return factoryEnvState{DispatchValue: dispatch, LegacyValue: legacy, DispatchSet: dispatchSet, LegacySet: legacySet}
}

type gptFactoryState struct {
	todoHash, tasksHash     string
	dispatches, factoryRuns int
}

func captureGPTFactoryState(t *testing.T, root string) gptFactoryState {
	t.Helper()
	state := gptFactoryState{todoHash: hashGPTFactoryTree(t, filepath.Join(root, ".moai", "state", "todo")), tasksHash: hashGPTFactoryTree(t, filepath.Join(root, ".moai", "tasks"))}
	record, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).LoadPure()
	if err == nil {
		state.factoryRuns = len(record.Runtime.Runs)
		state.dispatches = len(record.Runtime.Assignments)
	}
	return state
}

func hashGPTFactoryTree(t *testing.T, root string) string {
	t.Helper()
	hash := sha256.New()
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		_, _ = fmt.Fprintf(hash, "%s\x00", filepath.ToSlash(path))
		_, _ = hash.Write(raw)
		return nil
	})
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func slicesContainPair(items []string, key, value string) bool {
	for i := 0; i+1 < len(items); i++ {
		if items[i] == key && items[i+1] == value {
			return true
		}
	}
	return false
}
