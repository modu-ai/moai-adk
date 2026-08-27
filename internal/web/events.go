// SSE 채널 + fsnotify 감시 + 디바운스. Go 표준 라이브러리 + fsnotify 만 쓴다.
//
// 계약: 이벤트 본문에 데이터를 싣지 않는다. "이 영역이 바뀌었다"는 신호만 보내고,
// 브라우저가 해당 조각을 /fragment/... 로 다시 가져온다 (렌더링 진실은 서버 한 곳).
//
// handoff 원본 대비 로컬 보정: eventFor 가 가장 구체적인(=가장 긴) 감시 경로를
// 고르도록 바꿨다. 원본은 map 순회 순서에 의존해서, `.moai/state/goal/x.json`
// 변경이 실행마다 "goal" 로도 "session" 으로도 매칭될 수 있었다 — 디바운스
// 테스트가 비결정적으로 실패하는 원인이다.
package web

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounce = 250 * time.Millisecond

// 이벤트 이름 ↔ 감시 경로. verify 는 키가 수백 개라 디렉터리 단위로만 감시한다.
var watchMap = map[string][]string{
	"spec":    {".moai/specs"},
	"session": {".moai/state"},
	"goal":    {".moai/state/goal"},
	"verify":  {".moai/state/verify"},
	// SSE event KEY stays "kanban" — it is a frontend-visible contract. Only
	// the watched PATH moved with the state-directory rename.
	"kanban": {".moai/state/todo"},
	"config": {".moai/config/sections"},
}

// Hub 는 열린 SSE 연결 집합이다. 값을 나르지 않으므로 상태는 채널뿐이다.
type Hub struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func NewHub() *Hub { return &Hub{clients: map[chan string]struct{}{}} }

func (h *Hub) add() chan string {
	ch := make(chan string, 8)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) remove(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

// Publish 는 열린 모든 연결에 이벤트 이름을 흘린다.
func (h *Hub) Publish(event string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- event:
		default: // 느린 클라이언트는 건너뛴다. 재연결하면 최신 상태를 다시 가져온다.
		}
	}
}

// ServeEvents 는 GET /events 핸들러다. 루프백·동일출처 미들웨어는 app.go 가 이미 씌운다.
func (h *Hub) ServeEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := h.add()
	defer h.remove(ch)

	// 연결 직후 한 번 — 브라우저가 "실시간 켜짐"을 확정할 수 있게.
	// 쓰기 실패는 곧 연결이 끊겼다는 뜻이다. 다음 Flush 와 r.Context().Done()
	// 이 그 사실을 받아 루프를 끝내므로, 여기서는 무시하고 흘려보낸다.
	_, _ = fmt.Fprint(w, "event: ready\ndata: 1\n\n")
	flusher.Flush()

	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: 1\n\n", ev)
			flusher.Flush()
		case <-keepalive.C:
			_, _ = fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// Watch 는 프로젝트 루트 기준 경로들을 감시하고 디바운스해서 Publish 한다.
// 저장 한 번에 파일 이벤트가 여러 개 오므로 묶는 것이 필수다.
func (h *Hub) Watch(root string, stop <-chan struct{}) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	pathEvent := map[string]string{}
	for event, dirs := range watchMap {
		for _, d := range dirs {
			abs := filepath.Join(root, d)
			if err := w.Add(abs); err != nil {
				continue // 아직 없는 디렉터리는 건너뛴다 (프로젝트 초기 상태)
			}
			pathEvent[abs] = event
		}
	}

	pending := map[string]bool{}
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	for {
		select {
		case <-stop:
			return nil
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			if name := eventFor(pathEvent, ev.Name); name != "" {
				pending[name] = true
				timer.Reset(debounce)
			}
		case <-timer.C:
			for name := range pending {
				h.Publish(name)
				delete(pending, name)
			}
		case _, ok := <-w.Errors:
			if !ok {
				return nil
			}
			// 감시 실패는 치명적이지 않다. 브라우저가 폴백 폴링으로 내려간다.
		}
	}
}

// eventFor 는 변경된 경로를 가장 구체적인 감시 경로에 귀속시킨다.
// `.moai/state` 와 `.moai/state/goal` 이 모두 등록돼 있을 때, goal 하위 변경은
// 반드시 "goal" 로 간다 — map 순회 순서에 좌우되지 않는다.
func eventFor(pathEvent map[string]string, changed string) string {
	dir := filepath.Dir(changed)
	best, bestLen := "", -1
	for p, name := range pathEvent {
		if dir != p && filepath.Dir(dir) != p {
			continue
		}
		if len(p) > bestLen {
			best, bestLen = name, len(p)
		}
	}
	return best
}
