// todo_lock_scope_test.go — SPEC-TCD-LLM-DECIDER-001 M3: the lock-scope
// tests (AC-TLD-005). The LLM judgment is computed BEFORE the queue lock is
// acquired and attached INSIDE the same locked write as a static carrier
// (plan OD-D) — an in-lock placement serializes concurrent adds behind the
// flock for judgment-duration each and holds the queue hostage to a network
// stall. The seams are package vars, so these tests stay NON-parallel.
package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// slowLLMServer stalls stall per request, then answers with one valid
// closed-set judgment.
func slowLLMServer(t *testing.T, stall time.Duration) (*httptest.Server, *llmEndpointRecorder) {
	t.Helper()
	rec := &llmEndpointRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		rec.add(r.Header.Clone(), b)
		time.Sleep(stall)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, llmEnvelope(llmJudgment("high", "parallelizable", false, "llm", "slow but valid")))
	}))
	t.Cleanup(server.Close)
	return server, rec
}

// addAtRoot runs one plain add anchored at root through its own cobra
// command — the shape concurrent adds need (one command cannot execute
// twice concurrently). A nil decider routes through the standing selection,
// exactly like the CLI surface.
func addAtRoot(root string, text string) error {
	cmd := &cobra.Command{}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	_, err := runTodoAddAppendRoot(root, cmd, text, false, nil, nil)
	return err
}

// TestTodoAdd_LLMJudgmentsOverlapOutsideLock — AC-TLD-005: two concurrent
// adds with a 1500 ms stall per judgment overlap the judgments.
//
// D5 boundary (run-phase named, measured on this tree): the OVERLAP
// observation is the arrival GAP between the two judgment requests, bounded
// at 500 ms. A serial-in-lock placement cannot meet it — its lower bound is
// the full first judgment (1500 ms) plus the write, because the second
// judgment's request does not leave until the first locked write finished.
// The measured gap under the overlapped placement is single-digit
// milliseconds.
//
// (A total-wall-time bound was considered and rejected with measurement:
// the write path — card analysis plus the flock-serialized write — costs
// ~1.8 s serial on this tree regardless of placement, so a wall-clock bound
// could not separate the two placements without flaking on the write
// overhead; the arrival gap is the placement's direct, noise-free
// signature.)
func TestTodoAdd_LLMJudgmentsOverlapOutsideLock(t *testing.T) {
	root, store := todoFixture(t)
	t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)
	server, rec := slowLLMServer(t, 1500*time.Millisecond)
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = addAtRoot(root, fmt.Sprintf("concurrent slow judgment card %d", i))
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent add %d errored: %v", i, err)
		}
	}
	if got := rec.count(); got != 2 {
		t.Fatalf("endpoint received %d judgment requests, want 2", got)
	}
	gap := rec.requests[1].arrivedAt.Sub(rec.requests[0].arrivedAt)
	if gap < 0 {
		gap = -gap
	}
	if gap >= 500*time.Millisecond {
		t.Errorf("the two judgment requests arrived %v apart, want under the 500ms overlap bound (a serial-in-lock placement costs the full first judgment, 1500ms, before the second request can leave)", gap)
	}
	both, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	classified := 0
	for _, it := range both.Items {
		c := it.Classification
		if c == nil {
			t.Errorf("card %s admitted with no classification", it.ID)
			continue
		}
		if c.Decider != factory.DeciderIdentityLLM || c.Priority != factory.ClassPriorityHigh {
			t.Errorf("card %s classification = %+v, want the model judgment (decider llm, high)", it.ID, c)
		}
		classified++
	}
	if classified != 2 {
		t.Errorf("queue holds %d classified cards, want 2", classified)
	}
}

// TestTodoAdd_ListReadCompletesDuringLLMJudgment — AC-TLD-005's reader
// clause, as a completion guard: while one add's judgment is in flight, a
// `todo list` read completes with no error and no deadlock.
//
// Observation-design note (D5, measured on this tree): the plain list read
// costs ~3-5 s BY ITSELF (baseline measured 4.87 s), above the 1200 ms
// stall — so a wall-time comparison cannot separate "read blocked by the
// stall" from "read slow by itself". The clause's real blocking risk is the
// queue flock, which the overlap test bounds directly; this test pins that
// the in-flight judgment never deadlocks or errors the reader.
func TestTodoAdd_ListReadCompletesDuringLLMJudgment(t *testing.T) {
	root, _ := todoFixture(t)
	t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)
	server, rec := slowLLMServer(t, 1200*time.Millisecond)
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	done := make(chan error, 1)
	go func() { done <- addAtRoot(root, "in-flight judgment card") }()

	deadline := time.Now().Add(5 * time.Second)
	for rec.count() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if rec.count() == 0 {
		t.Fatal("the add never reached the endpoint, so the reader observation proves nothing")
	}

	readDone := make(chan error, 1)
	go func() {
		_, _, err := runTodo(t, "list", "--json")
		readDone <- err
	}()
	select {
	case err := <-readDone:
		if err != nil {
			t.Fatalf("todo list errored while a judgment was in flight: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Error("todo list did not complete while a judgment was in flight (reader deadlocked behind the judgment)")
	}
	if err := <-done; err != nil {
		t.Fatalf("add: %v", err)
	}
}

// TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice — AC-TLD-003's notice
// clause across BOTH add entry points: an LLM failure still admits the card
// with the fail-safe default and EXACTLY ONE notice line — the pre-classify
// restructure must not double-print (its notice) on top of the seam's.
func TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice(t *testing.T) {
	t.Run("append path", func(t *testing.T) {
		_, store := todoFixture(t)
		t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)
		server, _ := llmServer(t, http.StatusInternalServerError, "boom")
		withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

		out, errOut, err := runTodo(t, "add", "failing endpoint card")
		if err != nil {
			t.Fatalf("add must not fail on an LLM error: %v", err)
		}
		if n := strings.Count(errOut, "classification decider unavailable"); n != 1 {
			t.Errorf("stderr carries %d fallback notices, want exactly 1 (stderr %q)", n, errOut)
		}
		id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range rec.Items {
			if it.ID != id {
				continue
			}
			c := it.Classification
			want := factory.DefaultCardClassification()
			if c == nil || c.Priority != want.Priority || c.Blocked != want.Blocked || c.Mode != want.Mode || c.Decider != want.Decider {
				t.Errorf("failed-LLM classification = %+v, want the fail-safe defaults", c)
			}
		}
	})
	t.Run("pick path", func(t *testing.T) {
		_, store := todoFixture(t)
		t.Setenv(config.EnvTodoDecider, factory.DeciderIdentityLLM)
		server, _ := llmServer(t, http.StatusInternalServerError, "boom")
		withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

		out, errOut, err := runTodo(t, "add", "failing endpoint pick card", "--pick")
		if err != nil {
			t.Fatalf("pick must not fail on an LLM error: %v", err)
		}
		if n := strings.Count(errOut, "classification decider unavailable"); n != 1 {
			t.Errorf("stderr carries %d fallback notices, want exactly 1 (stderr %q)", n, errOut)
		}
		id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range rec.Items {
			if it.ID != id {
				continue
			}
			c := it.Classification
			if c == nil || c.Decider != factory.DeciderIdentityDefault {
				t.Errorf("failed-LLM pick classification = %+v, want the fail-safe default", c)
			}
		}
	})
}
