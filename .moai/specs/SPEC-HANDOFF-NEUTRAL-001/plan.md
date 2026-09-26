# Plan — SPEC-HANDOFF-NEUTRAL-001 (card t1273)

> 마일스톤 구조: **M1이 이 카드의 run 범위** (리드 B안+조정 승인, 2026-09-26). M2·M3는 병합 의존 후속 — 이 plan은 범위·선행 조건만 기록하고 run하지 않는다.

## §A 마일스톤 분해 (M1)

### M1.1 — RED 선관측 (테스트 먼저)
- AC-HN-001..007(+4b) 테스트 작성 → `unset … && go test …` FAIL 출력 관측 → `.moai/reports/t1273/red-*.txt` 저장 (첫 줄 = 관측 시점 HEAD SHA).
- **Fixture 계약**: 저장소는 factory.db이므로 fixture는 `SavePending` 호출/homestate 세팅으로 만든다 — legacy `pending.json` fixture는 호환 분기 검증(AC-HN-004b)에만.
- 판정 패키지: `internal/cli/` (show), 워크트리 생성 패키지 (시딩 — 실제 materializer 소재는 이 시점에 실측 확정해 acceptance 표에 회기).

### M1.2 — `moai handoff show` (P3)
- `internal/homestate/handoff.go`: 소비 이력 최신 조회 함수 1개(status='consumed', consumed_at DESC) — 폴백 소스.
- `internal/cli/handoff.go`: `newHandoffShowCmd` — `ReadPending`(DB pending row) → 소비 이력 최신 폴백, Body verbatim, `--json`, 4-로케일 헤더.
- 재사용: `internal/hook/handoff`의 `ReadPending`·`handoffLocaleStrings` 관례 (신규 복제 금지).
- 무상태 계약: 이 동사는 상태 전이·write를 일으키지 않는다.

### M1.3 — 워크트리 시딩 (P1 전제 1)
- materializer(`moai worktree new` 트리 생성 경로)에서 `.codex/hooks.json` 시딩 — **`internal/codexwiring`(wire.go)의 생성 로직 재사용**. `internal/cli/update_codex_wiring.go`는 "creates nothing" 존재 게이트라 생성 로직 소재가 아니다(1차 감사 정정).
- 런처 진입 보완: `moai cc -w`/`moai codex -w` 해석 경로에서 부재 시 같은 시딩.
- 멱등·fail-open (REQ-HN-005/006/007).

### M1.4 — LIVE 검증 (관문 b·c — 쿼터 회복 9/28 14:37 이후)
- 격리 CODEX_HOME·스크래치 프로젝트 구축은 **회복 전에 미리** 준비 (모델 호출 없는 부분).
- 관문 (a) moai 훅 발화 / (b) additionalContext 도달(마커) / (c) 길이·강등 — 관측을 `.moai/reports/t1273/live-*.txt`.
- 상한: 관문당 3회·벽시계 30분 (선언됨).

### M1.5 — P1 판정·조건부 구현
- 관문 (b) 통과 → `additionalContextEvents` + `EventSessionStart` (+테스트, AC-HN-008). 실패 → P1 기각 기록, P3 단독 종결.

### M1.6 — GREEN 일괄 관측·sync
- AC 표 회기(명령+출력), 형식 게이트(vet/gofmt/lint), sync-audit, 리드 통합 창 요청.

## §B M2·M3 (후속 — run하지 않음)

| 마일스톤 | 선행 조건 | 범위 |
|---|---|---|
| M2 | t1175(룰 다이어트) develop 병합 | ultrathink·하네스 고유 키워드 제거 — SSOT 문서(session-handoff.md·-examples.md·output-styles §8) + 렌더 코드(handoff.go·pending.go·handoff_inject_render.go) **같은 마일스톤 동시 변경**. t1175 압축 baseline에서 150k 재측정 |
| M3 | t1243(지시 파일 통합) develop 병합 | AGENTS.md 소비 계약 문구 배치(design §D5 초안) + 옵션 C 후속 카드 판정 |

## §C PRESERVE (무접촉 목록)

- `internal/hook/handoff/persist.go` (SessionEnd→memory 흐름) — AUTORESUME의 기존 절반, 무접촉.
- `internal/hook/handoff_inject.go`의 claim/소비 로직 — show는 읽기만 (REQ-HN-002).
- 6블록 본문 렌더 형식·Directives 구조 (M2 소관).
- `internal/template/templates/**` 전체 (M1 템플릿 무접촉).
- `AGENTS.md.tmpl` (M3 소관).
- 운영자 `~/.codex/**`, primary `.codex/hooks.json` (LIVE 격리 계약).

## §D 검증 계획 (레인-로컬·대상 한정)

- `go test ./internal/cli/ -count=1 -run 'TestHandoff*'` + `TestRenderHandoffContext` 등 기존 형제.
- `go test ./internal/worktree/ -count=1` (시딩 추가 후 관련 그룹) / `./internal/codexadapter/`.
- `go vet`·`gofmt -l`·`golangci-lint run` 대상 패키지.
- 전체 스위트는 develop push 후 CI (레인에서 `go test ./...` 금지).

## §E 리스크 대응

| 리스크 | 대응 |
|---|---|
| 쿼터 미회복·LIVE 지연 | M1.2·M1.3은 LIVE 무관 완료 — P1 판정만 9/28+ 로 분리 (design §E) |
| materializer 소재 파악 착오 | M1.1에서 함수 소재를 실측 확정 후 acceptance 표 회기 (추측 경로로 테스트 경로 고정 금지) |
| 관문 (b) 기각 | P3 단독 종결 — 설계상 이중화, 재협상 불필요 |
| 기존 워크트리 트리 다수 | 런처 보완(M1.3)이 진입 시 채움 — 일괄 마이그레이션 스크립트 불요 |
