package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// LSP 문서 동기화 메서드 상수 (하드코딩 방지).
const (
	methodDidOpen   = "textDocument/didOpen"
	methodDidChange = "textDocument/didChange"
	methodDidClose  = "textDocument/didClose"
	methodDidSave   = "textDocument/didSave"
)

// docEntry는 documentCache의 개별 파일 상태를 나타냅니다.
type docEntry struct {
	languageID   string
	version      int32
	content      string
	lastActivity time.Time
}

// documentCache는 언어 서버에 열린 파일 상태를 추적하여 중복 didOpen 전송을 방지합니다 (REQ-LC-006).
//
// @MX:ANCHOR: [AUTO] documentCache — 모든 문서 동기화 작업의 중앙 상태 저장소
// @MX:REASON: fan_in >= 3 — openOrChange, reapIdle, didSave, OpenFile, DidSave, Manager 모두 이 구조체를 통해 문서 상태를 관리함
type documentCache struct {
	mu      sync.RWMutex
	entries map[string]docEntry
}

// newDocumentCache는 비어 있는 documentCache를 생성합니다.
func newDocumentCache() *documentCache {
	return &documentCache{
		entries: make(map[string]docEntry),
	}
}

// openOrChange는 URI에 따라 textDocument/didOpen 또는 textDocument/didChange를 전송합니다.
//
// 규칙:
//   - 미등록 URI: didOpen (version=1) 전송 후 캐시에 추가
//   - 등록된 URI + 동일 콘텐츠: lastActivity만 업데이트 (no-op)
//   - 등록된 URI + 변경된 콘텐츠: didChange (version 증가, 전체 문서 동기화) 전송 후 캐시 업데이트
func (c *documentCache) openOrChange(ctx context.Context, tr transport.Transport, uri, languageID, content string) error {
	c.mu.Lock()

	entry, exists := c.entries[uri]

	if !exists {
		// 신규 파일: didOpen 전송
		newEntry := docEntry{
			languageID:   languageID,
			version:      1,
			content:      content,
			lastActivity: time.Now(),
		}
		c.entries[uri] = newEntry
		c.mu.Unlock()

		params := map[string]any{
			"textDocument": map[string]any{
				"uri":        uri,
				"languageId": languageID,
				"version":    int32(1),
				"text":       content,
			},
		}
		if err := tr.Notify(ctx, methodDidOpen, params); err != nil {
			return transport.WrapCallError(methodDidOpen, uri, languageID, err)
		}
		return nil
	}

	if entry.content == content {
		// 동일 콘텐츠: lastActivity만 업데이트 (no-op)
		entry.lastActivity = time.Now()
		c.entries[uri] = entry
		c.mu.Unlock()
		return nil
	}

	// 콘텐츠 변경: didChange 전송
	newVersion := entry.version + 1
	entry.version = newVersion
	entry.content = content
	entry.lastActivity = time.Now()
	c.entries[uri] = entry
	c.mu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":     uri,
			"version": newVersion,
		},
		"contentChanges": []map[string]any{
			{"text": content},
		},
	}
	if err := tr.Notify(ctx, methodDidChange, params); err != nil {
		return transport.WrapCallError(methodDidChange, uri, languageID, err)
	}
	return nil
}

// touch는 URI의 lastActivity 타임스탬프를 현재 시각으로 업데이트합니다.
// URI가 캐시에 없으면 no-op입니다.
func (c *documentCache) touch(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entries[uri]; ok {
		entry.lastActivity = time.Now()
		c.entries[uri] = entry
	}
}

// snapshot은 현재 캐시의 복사본을 반환합니다. idle reaper 등에서 사용합니다.
func (c *documentCache) snapshot() map[string]docEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]docEntry, len(c.entries))
	for k, v := range c.entries {
		out[k] = v
	}
	return out
}

// remove는 URI를 캐시에서 삭제합니다.
func (c *documentCache) remove(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, uri)
}

// reapIdle은 idleTimeout을 초과한 항목에 textDocument/didClose를 전송하고 캐시에서 제거합니다 (REQ-LC-022).
// 에러는 로그하지만 reaping을 중단하지 않습니다.
// 제거된 항목 수를 반환합니다.
func (c *documentCache) reapIdle(ctx context.Context, tr transport.Transport, idleTimeout time.Duration) int {
	now := time.Now()

	// 만료된 항목 목록 수집 (락 최소화)
	c.mu.RLock()
	var expired []struct {
		uri        string
		languageID string
	}
	for uri, entry := range c.entries {
		if now.Sub(entry.lastActivity) > idleTimeout {
			expired = append(expired, struct {
				uri        string
				languageID string
			}{uri: uri, languageID: entry.languageID})
		}
	}
	c.mu.RUnlock()

	if len(expired) == 0 {
		return 0
	}

	reaped := 0
	for _, e := range expired {
		params := map[string]any{
			"textDocument": map[string]any{
				"uri": e.uri,
			},
		}
		// 에러는 무시하고 계속 진행 (REQ-LC-022 요구사항)
		_ = tr.Notify(ctx, methodDidClose, params)
		c.remove(e.uri)
		reaped++
	}
	return reaped
}

// didSave는 추적 중인 URI에 textDocument/didSave를 전송합니다 (REQ-LC-023).
// URI가 캐시에 없으면 에러를 반환합니다.
func (c *documentCache) didSave(ctx context.Context, tr transport.Transport, uri string) error {
	c.mu.RLock()
	entry, ok := c.entries[uri]
	c.mu.RUnlock()

	if !ok {
		return transport.WrapCallError(methodDidSave, uri, "", fmt.Errorf("file not tracked: %q", uri))
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}
	if err := tr.Notify(ctx, methodDidSave, params); err != nil {
		return transport.WrapCallError(methodDidSave, uri, entry.languageID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// pathToURI — 파일 경로를 LSP URI로 변환
// ---------------------------------------------------------------------------

// pathToURI는 파일 경로를 LSP file:// URI로 변환합니다.
//
// 변환 규칙:
//   - "file://" 접두어가 있으면 그대로 반환
//   - 절대 경로: "file://" + 슬래시 표준화된 경로
//   - 상대 경로: "file://" + 경로 (as-is)
func pathToURI(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path
	}
	if filepath.IsAbs(path) {
		// Unix: /foo/bar.go → file:///foo/bar.go
		// Windows: C:\foo\bar.go → file:///C:/foo/bar.go
		return "file://" + filepath.ToSlash(path)
	}
	// 상대 경로: 그대로 반환 (테스트 호환성)
	return "file://" + path
}

// resolveLanguageID는 ServerConfig.Language에서 LSP languageId를 결정합니다.
// 알 수 없는 언어는 cfg.Language 그대로를 반환합니다.
func resolveLanguageID(language string) string {
	switch language {
	case "go":
		return "go"
	case "python":
		return "python"
	case "typescript":
		return "typescript"
	case "javascript":
		return "javascript"
	case "rust":
		return "rust"
	case "java":
		return "java"
	default:
		return language
	}
}
