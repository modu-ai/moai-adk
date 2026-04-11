package gopls

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestPendingRegistry_RegisterAndDispatch는 Register로 채널을 등록하고
// Dispatch로 payload를 전달하는 기본 흐름을 검증한다.
func TestPendingRegistry_RegisterAndDispatch(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	ch := r.Register(42)

	payload := json.RawMessage(`{"result":"ok"}`)
	dispatched := r.Dispatch(42, payload)
	if !dispatched {
		t.Fatal("Dispatch() = false, want true")
	}

	select {
	case got := <-ch:
		if string(got) != string(payload) {
			t.Errorf("payload = %s, want %s", got, payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("채널에서 payload를 받지 못함")
	}
}

// TestPendingRegistry_DispatchUnknownID는 등록되지 않은 ID로 Dispatch할 때
// false를 반환하는지 검증한다.
func TestPendingRegistry_DispatchUnknownID(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	dispatched := r.Dispatch(999, json.RawMessage(`{}`))
	if dispatched {
		t.Error("알 수 없는 ID로 Dispatch() = true, want false")
	}
}

// TestPendingRegistry_UnregisterRemovesEntry는 Unregister 후 Dispatch가 false를
// 반환하는지 검증한다.
func TestPendingRegistry_UnregisterRemovesEntry(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	r.Register(10)
	r.Unregister(10)

	dispatched := r.Dispatch(10, json.RawMessage(`{}`))
	if dispatched {
		t.Error("Unregister 후 Dispatch() = true, want false")
	}
}

// TestPendingRegistry_Concurrent는 동시에 여러 고루틴이 Register/Dispatch를 호출할 때
// 데이터 경쟁 없이 동작하는지 경쟁 감지기로 검증한다.
func TestPendingRegistry_Concurrent(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	const numRequests = 50

	var wg sync.WaitGroup
	results := make([]json.RawMessage, numRequests)

	// 모든 채널을 먼저 등록한다.
	channels := make([]<-chan json.RawMessage, numRequests)
	for i := int64(0); i < numRequests; i++ {
		channels[i] = r.Register(i)
	}

	// 수신 고루틴들을 먼저 시작한다.
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			select {
			case got := <-channels[idx]:
				results[idx] = got
			case <-time.After(2 * time.Second):
				t.Errorf("고루틴 %d: 타임아웃", idx)
			}
		}(i)
	}

	// Dispatch 고루틴들을 동시에 실행한다.
	var dispatchWg sync.WaitGroup
	for i := int64(0); i < numRequests; i++ {
		dispatchWg.Add(1)
		go func(id int64) {
			defer dispatchWg.Done()
			payload := json.RawMessage(fmt.Sprintf(`{"id":%d}`, id))
			r.Dispatch(id, payload)
		}(i)
	}

	dispatchWg.Wait()
	wg.Wait()

	// 모든 결과가 수신되었는지 확인한다.
	for i, res := range results {
		if res == nil {
			t.Errorf("결과 %d가 nil", i)
		}
	}
}

// TestNotificationDispatcher_Dispatch는 등록된 핸들러가 올바른 메서드로 호출되는지 검증한다.
func TestNotificationDispatcher_Dispatch(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()

	called := false
	var receivedPayload json.RawMessage
	d.Register("textDocument/publishDiagnostics", func(payload json.RawMessage) {
		called = true
		receivedPayload = payload
	})

	payload := json.RawMessage(`{"uri":"file:///main.go","diagnostics":[]}`)
	d.Dispatch("textDocument/publishDiagnostics", payload)

	if !called {
		t.Fatal("핸들러가 호출되지 않음")
	}
	if string(receivedPayload) != string(payload) {
		t.Errorf("payload = %s, want %s", receivedPayload, payload)
	}
}

// TestNotificationDispatcher_UnknownMethod는 등록되지 않은 메서드에 대해
// 패닉 없이 무시하는지 검증한다.
func TestNotificationDispatcher_UnknownMethod(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()
	// 패닉이 발생하지 않아야 한다.
	d.Dispatch("window/logMessage", json.RawMessage(`{"message":"hello"}`))
}

// TestNotificationDispatcher_MultipleHandlers는 여러 메서드에 대한 핸들러를
// 독립적으로 등록하고 호출하는지 검증한다.
func TestNotificationDispatcher_MultipleHandlers(t *testing.T) {
	t.Parallel()

	d := NewNotificationDispatcher()
	called := map[string]bool{}
	var mu sync.Mutex

	for _, method := range []string{"textDocument/publishDiagnostics", "window/logMessage"} {
		m := method
		d.Register(m, func(payload json.RawMessage) {
			mu.Lock()
			called[m] = true
			mu.Unlock()
		})
	}

	d.Dispatch("textDocument/publishDiagnostics", json.RawMessage(`{}`))
	d.Dispatch("window/logMessage", json.RawMessage(`{}`))

	mu.Lock()
	defer mu.Unlock()
	if !called["textDocument/publishDiagnostics"] {
		t.Error("publishDiagnostics 핸들러가 호출되지 않음")
	}
	if !called["window/logMessage"] {
		t.Error("logMessage 핸들러가 호출되지 않음")
	}
}

// TestPendingRegistry_Timeout은 Dispatch가 호출되지 않을 때 채널이 블록 상태를
// 유지하다가 컨텍스트 취소로 타임아웃되는 패턴을 검증한다.
func TestPendingRegistry_Timeout(t *testing.T) {
	t.Parallel()

	r := &PendingRegistry{}
	ch := r.Register(99)

	// Dispatch 없이 짧은 타임아웃으로 채널이 블록되는지 확인한다.
	select {
	case <-ch:
		t.Fatal("Dispatch 없이 채널에서 데이터를 받음")
	case <-time.After(50 * time.Millisecond):
		// 예상대로 타임아웃
	}
}
