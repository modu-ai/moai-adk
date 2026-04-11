package gopls

import (
	"encoding/json"
	"sync"
)

// ─── PendingRegistry ─────────────────────────────────────────────────────────
//
// REQ-GB-033: 요청 id를 키로 응답 채널을 관리하여 요청-응답을 상관시킨다.

// NotificationHandler는 알림 payload를 처리하는 함수 타입이다.
type NotificationHandler func(payload json.RawMessage)

// PendingRegistry는 진행 중인 요청을 추적하는 레지스트리다.
// sync.Map을 사용하여 동시성 안전한 방식으로 id → 응답 채널을 관리한다.
//
// @MX:ANCHOR: [AUTO] 요청-응답 상관 핵심 타입. bridge.go의 readLoop, sendRequest에서 사용한다.
// @MX:REASON: fan_in >= 3 (Register, Dispatch, Unregister 호출 지점)
type PendingRegistry struct {
	// sync.Map은 map[int64]chan json.RawMessage를 동시성 안전하게 관리한다.
	m sync.Map
}

// Register는 주어진 id에 대한 응답 채널을 생성하고 등록한다.
// 반환된 채널은 Dispatch가 호출될 때 payload를 수신한다.
func (r *PendingRegistry) Register(id int64) <-chan json.RawMessage {
	// 버퍼 크기 1: Dispatch가 발신자를 블록하지 않도록 한다.
	ch := make(chan json.RawMessage, 1)
	r.m.Store(id, ch)
	return ch
}

// Dispatch는 id에 해당하는 채널에 payload를 전달하고 채널을 레지스트리에서 제거한다.
// 등록된 id가 없으면 false를 반환한다.
func (r *PendingRegistry) Dispatch(id int64, payload json.RawMessage) bool {
	v, ok := r.m.LoadAndDelete(id)
	if !ok {
		return false
	}
	ch := v.(chan json.RawMessage)
	ch <- payload
	return true
}

// Unregister는 id에 해당하는 채널을 레지스트리에서 제거한다.
// 타임아웃 등으로 더 이상 응답을 기다리지 않을 때 호출한다.
func (r *PendingRegistry) Unregister(id int64) {
	r.m.Delete(id)
}

// ─── NotificationDispatcher ──────────────────────────────────────────────────
//
// REQ-GB-034: id 없는 메시지(알림)를 등록된 핸들러에 라우팅한다.

// NotificationDispatcher는 LSP 알림을 메서드별로 등록된 핸들러에 라우팅한다.
// 뮤텍스로 핸들러 맵을 보호한다.
type NotificationDispatcher struct {
	mu       sync.RWMutex
	handlers map[string]NotificationHandler
}

// NewNotificationDispatcher는 새 NotificationDispatcher를 생성한다.
func NewNotificationDispatcher() *NotificationDispatcher {
	return &NotificationDispatcher{
		handlers: make(map[string]NotificationHandler),
	}
}

// Register는 method에 대한 핸들러를 등록한다.
// 같은 method에 두 번 등록하면 이전 핸들러를 덮어쓴다.
func (d *NotificationDispatcher) Register(method string, handler NotificationHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[method] = handler
}

// Dispatch는 method에 해당하는 핸들러를 동기적으로 호출한다.
// 등록된 핸들러가 없으면 무시한다(패닉 없음).
func (d *NotificationDispatcher) Dispatch(method string, payload json.RawMessage) {
	d.mu.RLock()
	handler, ok := d.handlers[method]
	d.mu.RUnlock()

	if !ok {
		// 알 수 없는 알림은 조용히 무시한다.
		return
	}
	handler(payload)
}
