# t273 sync-audit F1 해소 기록 (lead addendum)

2026-08-26 · 레인 t273 오케스트레이터 (감사자 지정 재감사 범위 내 수행)

## F1 (blocking) 해소

- 지적: AC-001 en 규약 헤딩 토큰 `Card Classes` 부재 — 실제 헤딩 `Card classes` (sentence case, :251)
- 수정: `## Card classes — not every card needs every column` → `## Card Classes — not every card needs every column` (한 단어 케이스 수정)
- 근거 커밋: 이 addendum과 같은 커밋에 포함

## 좁은 재검증 (감사자 지정 범위: "AC-001's 4 locale greps only")

| 로케일 | 명령 | 관측 |
|---|---|---|
| ko | `grep -c "카드 클래스" docs-site/content/ko/advanced/kanban-mode.md` | 1 |
| en | `grep -c "Card Classes" docs-site/content/en/advanced/kanban-mode.md` | 1 (수정 후) |
| ja | `grep -c "カードクラス" docs-site/content/ja/advanced/kanban-mode.md` | 1 |
| zh | `grep -c "卡片类别" docs-site/content/zh/advanced/kanban-mode.md` | 1 |
| en (보강) | `grep -c "Class A" .../en/advanced/kanban-mode.md` | 1 (기존 유지 확인) |

전 측정은 수정 직후 현재 트리에서 실행 (this run, this tree).

## F2 반영

progress.md §E.2 AC-001 행을 헤딩 토큰 관측을 포함한 형태로 교체 — 이전 행이 Class A/B/C grep만 기록하고 헤딩 토큰 grep을 관측하지 않았던 과대 표현(감사 F2)을 시정. CHANGELOG의 "11 PASS, 0 FAIL"은 F1 착지 후 사실이 되므로 그대로 유효.

## 판정

- Functionality must-pass 방화벽 해소: AC-001 4로케일 전부 규약 토큰 충족
- 나머지 10 AC는 감사자가 이 트리에서 재실행 GREEN (원 보고서 `sync-audit-verdict.md`)
- 리드 판정: **PASS** — 4차원 조화평균 95.3 (Security 100 / Craft 96 / Consistency 98 / Functionality 88→F1 해소로 must-pass 조건 충족), blocking 결함 0

## Minor (미처리 — 기록만)

- F3: content/.moai/state/*.json 런타임 오염 (gitignored, cosmetic) — 불필요
- F4: 사이트맵 섹션 URL 정책 (사전 존재 hugo.toml 정책) — 별도 카드감
