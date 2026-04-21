---
id: SPEC-DOCS-SB-REMOVE-001
document: spec-compact
version: 1.0.0
---

# Compact View

## Requirements

- REQ-DOCS-KO-001~004: ko 4 파일 정리 (what-is-moai-adk, auto-quality 삭제, harness-engineering, _meta.yaml)
- REQ-DOCS-EN-001: en what-is-moai-adk.md에 /simplify·/batch 0건
- REQ-DOCS-JA-001: ja 동일
- REQ-DOCS-ZH-001: zh 동일
- REQ-DOCS-BUILD-001~002: Hugo 빌드 성공 + 깨진 링크 0건

## Acceptance

- AC-1: 4 locale 그렙 0건
- AC-2: auto-quality.md 삭제
- AC-3: ko auto-quality 참조 0건
- AC-4: Hugo 빌드 성공
- AC-5: 핵심 방법론 설명 보존
- AC-6: _meta.yaml 일관성
- AC-7: Pencil batch_design/batch_get 보존

## Exclusions

- Pencil MCP tool 이름 변경 없음
- utility-commands batch 일반 명사 변경 없음
- 도메인·Vercel 설정 변경 없음
