# Phase 3 — 콘텐츠 마이그레이션 로그

> 실행일: 2026-04-20
> 담당: expert-refactoring (SPEC-DOCS-SITE-001 Phase 3)
> 도구: `scripts/convert-nextra-to-hextra/main.go`

---

## 1. 사전 작업 — 누락 파일 복사

Phase 2 squash merge에서 누락된 ko locale 10개 파일 및 _meta.ts 2개를 원본 레포에서 복사:

| 파일 | 사유 |
|------|------|
| `ko/advanced/hooks-reference.mdx` | squash merge 누락 |
| `ko/agency/gan-loop.mdx` | squash merge 누락 |
| `ko/agency/usage-guide.mdx` | squash merge 누락 |
| `ko/contributing/index.mdx` | contributing/ 디렉토리 전체 누락 |
| `ko/core-concepts/auto-quality.mdx` | squash merge 누락 |
| `ko/core-concepts/harness-engineering.mdx` | squash merge 누락 |
| `ko/getting-started/windows-guide.mdx` | squash merge 누락 |
| `ko/multi-llm/cg-mode.mdx` | multi-llm/ 디렉토리 전체 누락 |
| `ko/multi-llm/index.mdx` | multi-llm/ 디렉토리 전체 누락 |
| `ko/multi-llm/model-policy.mdx` | multi-llm/ 디렉토리 전체 누락 |
| `ko/contributing/_meta.ts` | contributing/ _meta.ts 누락 |
| `ko/multi-llm/_meta.ts` | multi-llm/ _meta.ts 누락 |

복사 후 확인: ko 53 → 63 MDX, _meta.ts 36 → 38

---

## 2. Go 변환 도구 실행 결과

도구: `scripts/convert-nextra-to-hextra/main.go`

### Dry-run 결과 (사전 검증)

```
처리된 MDX 파일:    219
Callout 변환:       735
_meta.ts 변환:      38
Frontmatter 주입:   219
오류:               0건
```

### 실제 실행 결과

```
처리된 MDX 파일:    219
Callout 변환:       735
_meta.ts 변환:      38
Frontmatter 주입:   219
파일명 변경 (.md):  219
오류:               0건
```

---

## 3. 변환 상세

### T1 — Callout 변환 (735건)

- `import { Callout } from 'nextra/components'` 4개 변형 모두 제거
- `<Callout type="info">` → `{{< callout type="info" >}}`
- `<Callout type="warning">` → `{{< callout type="warning" >}}`
- `<Callout type="error">` → `{{< callout type="error" >}}`
- `<Callout type="tip">` → `{{< callout type="info" >}}` (tip → info 매핑, Hextra 지원 타입)
- `<Callout type="success">` → `{{< callout type="info" >}}` (success → info 매핑)
- `</Callout>` → `{{< /callout >}}`

Callout 타입별 분포:
- info: 309건
- tip→info: 270건
- warning: 150건
- error: 4건
- success→info: 2건
- 합계: 735건

### T2 — _meta.ts → _meta.yaml (38개)

- 파서 방식: brace-depth 카운팅 (정규식 본문 추출 대신)
- 중첩 객체 처리: `index: { title: "...", display: "hidden" }` 정확 파싱
- weight 매핑: 키 순서 × 10 (10, 20, 30, ...)
- display: "hidden" → `display: "hidden"` 보존

### T3 — YAML frontmatter 주입 (219페이지)

- `title`: 파일 내 첫 H1 헤더 텍스트 추출
- `weight`: _meta.yaml 키 순서 기반 (10단위)
- `draft: false`

weight 초기 오류(99) 발생 후 별도 스크립트로 수정:
- 원인: 첫 번째 실행 시 _meta.yaml 파서 버그(정규식 기반 본문 추출)로 weight 맵 구성 실패
- 수정: brace-depth 카운팅 파서로 교체 + weight 재주입 스크립트 실행
- 수정된 파일 수: 194

### T4 — .mdx → .md (219개)

- 모든 .mdx 파일 → .md 이름 변경 완료

---

## 4. Hugo 호환성 조정

Phase 3 중 Hugo 빌드를 위한 추가 수정:

1. **hugo.yaml contentDir 추가**: 각 언어에 `contentDir: content/{locale}` 설정
   - 원인: Hugo multilingual 모드에서 locale 기준 콘텐츠 디렉토리 명시 필요
   - 효과: EN/JA/ZH가 KO 콘텐츠를 포함하던 문제 해결

2. **index.md → _index.md 변환**: locale 루트 4개 + 섹션 레벨 34개 = 총 38개
   - 원인: Hugo에서 section landing page는 leaf bundle(index.md) 아닌 branch bundle(_index.md) 필요
   - 효과: 섹션 구조 정상 빌드

---

## 5. G2 Gate 검증 결과

| AC | 항목 | 결과 |
|----|------|------|
| G2-01 | Hugo 빌드 성공 (exit 0) | PASS |
| G2-02 | 파일 수 (ko 63, en/ja/zh 52) | PASS |
| G2-03 | Nextra import 잔재 0건 | PASS |
| G2-04 | Callout shortcode 정확히 735건 | PASS |
| G2-05 | frontmatter 219개 전량 주입 | PASS |
| G2-06 | Mermaid 블록 정확히 569건 | PASS |
| G2-07 | _meta.yaml 38개 | PASS |
| G2-08 | _meta.ts 잔재 0건 | PASS |
| G2-09 | .mdx 잔재 0건 | PASS |

빌드 경고: 3건 (Hextra 내부 deprecated API 사용, 내용 무관)

---

## 6. 실패 파일 목록

없음 (0건)

---

## 7. 다음 단계

Phase 4: 커스텀 컴포넌트 재현 (LanguageSelector, ClientNavbar, structured-data partial, SEO 메타)
