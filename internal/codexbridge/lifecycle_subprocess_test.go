package codexbridge

import (
	"context"
	"errors"
	"github.com/modu-ai/moai-adk/internal/codexapp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditRealTransportEOFWakesBridge(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip(err)
	}
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fake-codex")
	body := `import sys,json,time
for line in sys.stdin:
 m=json.loads(line)
 if m.get('method')=='initialize': print(json.dumps({'id':m['id'],'result':{'userAgent':'audit'}}),flush=True)
 elif m.get('method')=='thread/start': print(json.dumps({'id':m['id'],'result':{'thread':{'id':'eof-thread'}}}),flush=True)
 elif m.get('method')=='turn/start':
  print(json.dumps({'id':m['id'],'result':{'turn':{'id':'eof-turn'}}}),flush=True)
  time.sleep(0.15)
  sys.exit(0)
`
	if err := os.WriteFile(script, []byte("#!"+py+"\n"+body), 0700); err != nil {
		t.Fatal(err)
	}
	life, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := codexapp.Start(life, codexapp.Config{Binary: script, Home: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err = client.Initialize(life, "audit", "1"); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(dir, "state")
	if err := os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(state)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(life, client, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	start := time.Now()
	seg, err := e.Step(ctx, request("eof-real"))
	t.Logf("Step err=%v transport err=%v elapsed=%v done=%v", err, client.Err(), time.Since(start), seg.Done)
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("transport EOF did not wake bridge; waited for HTTP deadline")
	}
	if err == nil || seg.Done {
		t.Fatal("EOF synthesized success")
	}
}

func TestAuditCancelBeforeRealStartReplyInterruptsTurn(t *testing.T)   { testCanceledStart(t, false) }
func TestCanceledStartBeyondReplyDeadlineStillInterrupts(t *testing.T) { testCanceledStart(t, true) }
func testCanceledStart(t *testing.T, late bool) {
	t.Helper()
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip(err)
	}
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fake-codex")
	body := `import sys,json,time,os
for line in sys.stdin:
 m=json.loads(line)
 if m.get('method')=='initialize': print(json.dumps({'id':m['id'],'result':{'userAgent':'audit'}}),flush=True)
 elif m.get('method')=='thread/start': print(json.dumps({'id':m['id'],'result':{'thread':{'id':'eof-thread'}}}),flush=True)
 elif m.get('method')=='turn/start':
  open(os.path.join(os.environ['CODEX_HOME'],'allocated'),'w').write('yes')
  time.sleep(0.25)
  print(json.dumps({'id':m['id'],'result':{'turn':{'id':'eof-turn'}}}),flush=True)
 elif m.get('method')=='turn/interrupt':
  assert m['params']['threadId']=='eof-thread' and m['params']['turnId']=='eof-turn'
  open(os.path.join(os.environ['CODEX_HOME'],'interrupted'),'a').write('yes')
  print(json.dumps({'id':m['id'],'result':{}}),flush=True)
`
	if late {
		body = strings.Replace(body, "time.sleep(0.25)", "time.sleep(6)\n  for turn in ['eof-turn','eof-turn','unowned-turn']:\n   print(json.dumps({'method':'turn/started','params':{'threadId':'eof-thread','turn':{'id':turn,'status':'inProgress'}}}),flush=True)", 1)
	}

	if err := os.WriteFile(script, []byte("#!"+py+"\n"+body), 0700); err != nil {
		t.Fatal(err)
	}
	life, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := codexapp.Start(life, codexapp.Config{Binary: script, Home: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err = client.Initialize(life, "audit", "1"); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(dir, "state")
	if err := os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(state)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(life, client, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	done := make(chan error, 1)
	go func() { _, err := e.Step(ctx, request("cancel-real")); done <- err }()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		if _, err := os.Stat(filepath.Join(dir, "allocated")); err == nil {
			break
		}
		select {
		case <-deadline.C:
			t.Fatal("allocation never started")
		case <-time.After(10 * time.Millisecond):
		}
	}
	stop()
	err = <-done
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if late {
		time.Sleep(6500 * time.Millisecond)
	} else {
		time.Sleep(450 * time.Millisecond)
	}
	_, interrupted := os.Stat(filepath.Join(dir, "interrupted"))
	t.Logf("Step err=%v transport err=%v interrupt observed=%v", err, client.Err(), interrupted == nil)
	if interrupted != nil {
		t.Fatal("allocated turn was not interrupted after late turn/start reply")
	}
	if raw, err := os.ReadFile(filepath.Join(dir, "interrupted")); err != nil || string(raw) != "yes" {
		t.Fatal("duplicate or missing owned interrupt", string(raw), err)
	}
	if client.Err() != nil {
		t.Fatal("shared transport was stopped", client.Err())
	}

}

func TestAuditCanceledRPCDoesNotExhaustSharedTransport(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip(err)
	}
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fake-codex")
	body := `import sys,json
counter=0
def emit(x): print(json.dumps(x),flush=True)
for line in sys.stdin:
 m=json.loads(line); method=m.get('method');p=m.get('params',{})
 if method=='initialize': emit({'id':m['id'],'result':{'userAgent':'audit'}})
 elif method=='thread/start':
  counter+=1
  emit({'id':m['id'],'result':{'thread':{'id':'thread-'+str(counter)}}})
 elif method=='turn/start':
  th=p['threadId'];tr='turn-'+th
  emit({'id':m['id'],'result':{'turn':{'id':tr}}})
  emit({'id':'rpc-'+th,'method':'item/tool/call','params':{'threadId':th,'turnId':tr,'callId':'call-'+th,'tool':'echo','arguments':{'text':'hello'}}})
 elif method=='turn/interrupt':
  emit({'id':m['id'],'result':{}})
  emit({'method':'turn/completed','params':{'threadId':p['threadId'],'turn':{'id':p['turnId'],'status':'interrupted'}}})
`
	if err := os.WriteFile(script, []byte("#!"+py+"\n"+body), 0700); err != nil {
		t.Fatal(err)
	}
	life, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := codexapp.Start(life, codexapp.Config{Binary: script, Home: dir, QueueSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err = client.Initialize(life, "audit", "1"); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(dir, "state")
	if err := os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(state)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(life, client, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, id := range []string{"a", "b", "c"} {
		q := request(id)
		seg, err := e.Step(life, q)
		if err != nil {
			t.Fatalf("conversation %s failed after earlier canceled calls: %v; transport=%v", id, err, client.Err())
		}
		if seg.Tool == nil {
			t.Fatal("no tool")
		}
		if err = e.Cancel(life, q.Owner); err != nil {
			t.Fatal(err)
		}
	}
}
