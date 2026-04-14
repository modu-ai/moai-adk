package core

import (
	"context"
	"encoding/json"
	"fmt"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// LSP 쿼리 메서드 상수 (하드코딩 방지).
const (
	methodReferences = "textDocument/references"
	methodDefinition = "textDocument/definition"
)

// GetDiagnostics returns the latest diagnostics for path from the push-model cache (REQ-LC-002b).
//
// 동작 방식:
//   - 서버가 push한 textDocument/publishDiagnostics 알림을 Start()에서 등록한 핸들러가 캐시에 저장
//   - GetDiagnostics는 해당 캐시에서 즉시 반환 (blocking I/O 없음)
//   - URI가 문서 캐시에 없으면 ErrFileNotOpen 반환
//
// @MX:ANCHOR: [AUTO] GetDiagnostics — LSP 진단 쿼리 공개 API
// @MX:REASON: fan_in >= 3 — Ralph 엔진, Quality Gates, LOOP 커맨드, Manager, 통합 테스트에서 호출됨
func (c *client) GetDiagnostics(_ context.Context, path string) ([]lsp.Diagnostic, error) {
	uri := pathToURI(path)

	// ErrFileNotOpen: 문서 캐시에 없으면 반환 (REQ-LC-002b 규칙: 호출자가 먼저 OpenFile 해야 함)
	snap := c.docs.snapshot()
	if _, ok := snap[uri]; !ok {
		return nil, ErrFileNotOpen
	}

	c.diagnosticsMu.RLock()
	diags := c.diagnostics[uri]
	c.diagnosticsMu.RUnlock()

	// nil slice → empty slice (일관성)
	if diags == nil {
		return []lsp.Diagnostic{}, nil
	}
	out := make([]lsp.Diagnostic, len(diags))
	copy(out, diags)
	return out, nil
}

// FindReferences returns all reference locations for the symbol at pos in path.
//
// 사전 조건 검사:
//   - serverCaps.Supports("textDocument/references")가 false이면 ErrCapabilityUnsupported 반환
//
// @MX:ANCHOR: [AUTO] FindReferences — LSP references 쿼리 공개 API
// @MX:REASON: fan_in >= 3 — Ralph 엔진, Quality Gates, LOOP 커맨드, 통합 테스트에서 호출됨
func (c *client) FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error) {
	if !c.serverCaps.Supports(methodReferences) {
		return nil, fmt.Errorf("FindReferences: %w", ErrCapabilityUnsupported)
	}

	uri := pathToURI(path)
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     pos,
		"context":      map[string]any{"includeDeclaration": true},
	}

	var raw json.RawMessage
	if err := transport.CallWithTimeout(ctx, c.tr, methodReferences, params, &raw, c.cfg.Language); err != nil {
		return nil, transport.WrapCallError(methodReferences, uri, c.cfg.Language, err)
	}

	return parseLocations(raw)
}

// GotoDefinition returns definition locations for the symbol at pos in path.
//
// 사전 조건 검사:
//   - serverCaps.Supports("textDocument/definition")가 false이면 ErrCapabilityUnsupported 반환
//
// LSP 응답은 Location, []Location, LocationLink[] 중 하나일 수 있습니다.
// v1 구현에서는 []Location 또는 단일 Location을 처리합니다 (tolerant decoder).
//
// @MX:ANCHOR: [AUTO] GotoDefinition — LSP definition 쿼리 공개 API
// @MX:REASON: fan_in >= 3 — Ralph 엔진, Quality Gates, LOOP 커맨드, 통합 테스트에서 호출됨
func (c *client) GotoDefinition(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error) {
	if !c.serverCaps.Supports(methodDefinition) {
		return nil, fmt.Errorf("GotoDefinition: %w", ErrCapabilityUnsupported)
	}

	uri := pathToURI(path)
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     pos,
	}

	var raw json.RawMessage
	if err := transport.CallWithTimeout(ctx, c.tr, methodDefinition, params, &raw, c.cfg.Language); err != nil {
		return nil, transport.WrapCallError(methodDefinition, uri, c.cfg.Language, err)
	}

	return parseLocations(raw)
}

// parseLocations는 LSP 응답 raw JSON을 []Location으로 디코딩합니다.
//
// 처리 전략 (tolerant decoder):
//  1. 배열 시도: []lsp.Location으로 언마샬
//  2. 실패 시 단일 객체 시도: lsp.Location으로 언마샬 → []lsp.Location{} 래핑
//  3. null 응답: 빈 슬라이스 반환
func parseLocations(raw json.RawMessage) ([]lsp.Location, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []lsp.Location{}, nil
	}

	// 배열 먼저 시도
	var locs []lsp.Location
	if err := json.Unmarshal(raw, &locs); err == nil {
		return locs, nil
	}

	// 단일 객체 시도 (tolerant fallback)
	var single lsp.Location
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("parseLocations: unable to decode as array or single location: %w", err)
	}
	return []lsp.Location{single}, nil
}
