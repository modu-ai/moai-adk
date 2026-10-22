# SPEC-FACTORY-BOOT-MINIMAL-001 — 구현 계획

> 카드 t1923 · 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` · Tier M · 개발 방식 TDD(`quality.yaml` `constitution.development_mode: tdd`)

## §A 맥락

- 대상 파일(변경): `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/role_injection_budget_test.go`, `internal/hook/role_composite_e2e_test.go`, `internal/hook/session_start_guide_i18n_red_test.go`, `internal/hook/session_start_guide_i18n_render_test.go`, 재배치 목적지 문서 1쌍(템플릿 + 배포), `AGENTS.md`.
- 보존 대상(PRESERVE): `.claude/rules/moai/workflow/factory-dispatch.md`, `.claude/rules/moai/workflow/cross-session-messaging.md`, 그리고 두 파일의 `internal/template/templates/` 미러(REQ-FBM-007). `internal/hook/role_rules.go`의 상수·로직(`roleRulesContextLimit`, `roleRuleSizeGate`, `buildRoleCore`). `internal/config/token_budget_guard.go`의 `CodexContractByteCeiling`. `internal/template/templates/AGENTS.md.tmpl`.
- Template-First: 재배치 목적지가 `.claude/` 아래 문서라면 `internal/template/templates/.claude/...` 미러를 먼저 고치고 `make build`로 배포본을 맞춘다. Go 소스(`internal/hook`)는 템플릿 대상이 아니다. AGENTS.md 두 문단은 루트 전용이다.

## §B 연구 결과 (plan 단계 실측)

모든 측정은 기준 트리에서 이 세션이 직접 실행했다. 원문 출력은 `probe/`·`probe/research/`에 커밋 대상으로 보관한다(scratch 경로는 인용하지 않는다). Go 측정은 `moai slot acquire --resource go-heavy` 임대 아래, load 27~32에서 실행했다.

### §B.1 측정 도구

- `probe/research/measure_t1923_test.go.txt` — `go test -overlay`로 `internal/hook`에 가상 주입해 실제 렌더러(`factoryLeaderNotice`, `factoryBootstrapNoticeForSource`, `factoryLaneRuleForSource`, `roleRuleInjectionFor`, `assembleInjectionComposite`)의 UTF-16 길이를 출력한다. 트리에 파일을 쓰지 않는다.
- `probe/zz_t1923_boot_probe_test.go.txt` + `probe/overlay.json` — 수용 기준 탐침(AC-FBM-001). overlay 경로는 저장소 루트 기준 상대 경로라 어느 클론에서도 재실행된다.

### §B.2 필드별·조합별 크기 (UTF-16, 기준 트리)

명령: `go test -v -overlay <scratch overlay> -run 'TestT1923Measure$' -count=1 ./internal/hook/` → exit 0. 원문: `probe/research/measure-5995962990.txt`. 발췌(원문 그대로):

```text
FIELDS lang=en leaderHeader=37 leaderIdentity=319 leaderManual=202 entryGuide=140 agentFanout=60 leaderSocket=17 leaderClasses=375 leaderStagger=444 gateSummary=226 operationalStatus=99 settingsAuto=69 settingsVerify=113 laneJoin=58 laneJoinNoCount=55 laneNextCardRule=1926 laneManualDispatchRule=317 laneRecheckRule=779 laneSpawnAuthority=467 autoModeGuideLeader=464 autoModeGuideLane=457 docsPointer=51
MEASURE role=leader face=deployed lang=en source=startup bootstrap=2573 laneRule=0 roleBlock=4303 composite_factory=6878
MEASURE role=lane face=deployed lang=en source=startup bootstrap=1036 laneRule=2713 roleBlock=4301 composite_factory=8054
MEASURE role=lane face=deployed lang=en source=clear bootstrap=0 laneRule=2713 roleBlock=4301 composite_factory=7016
MEASURE role=lane face=deployed lang=en source=compact bootstrap=0 laneRule=0 roleBlock=4301 composite_factory=4301
MEASURE role=leader face=deployed lang=ko source=startup bootstrap=1650 laneRule=0 roleBlock=4303 composite_factory=5955
MEASURE role=lane face=deployed lang=ko source=startup bootstrap=759 laneRule=2074 roleBlock=4301 composite_factory=7138
MEASURE role=lane face=deployed lang=ja source=startup bootstrap=766 laneRule=2084 roleBlock=4301 composite_factory=7155
MEASURE role=lane face=deployed lang=zh source=startup bootstrap=584 laneRule=1680 roleBlock=4301 composite_factory=6569
```

정리(배포·템플릿 두 면 동일값):

| 조합 | 부팅 공지 | 레인 규칙 | 팩토리 고정분 | 역할 블록 |
|---|---:|---:|---:|---:|
| leader · startup · en / ko / ja / zh | 2,573 / 1,650 / 1,651 / 1,229 | 0 | 같음 | 4,303 |
| lane(auto) · startup · en / ko / ja / zh | 1,036 / 759 / 766 / 584 | 2,713 / 2,074 / 2,084 / 1,680 | 3,751 / 2,835 / 2,852 / 2,266 | 4,301 |
| lane(auto) · clear · en / ko / ja / zh | 0 | 2,713 / 2,074 / 2,084 / 1,680 | 같음 | 4,301 |
| 모든 역할 · compact | 0 | 0 | 0 | 4,301 / 4,303 |

clear·compact에서 부팅 공지 0, compact에서 레인 규칙 0은 소스 게이트(`session_start_factory.go:93-95`, `:128-130`)와 일치한다.

### §B.3 23,145자 넘침의 분해 (카드 관측의 원인)

명령: `python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py /Users/goos/.moai/claude-profiles/moai-adk/projects 2026-10-21T00:00:00Z` → exit 0. 원문: `probe/research/live-scan-20261021.txt`. 열 = 시각, 역할, 합본, 역할 블록 앞, 역할 블록, 넘침 여부. 발췌:

```text
2026-10-21T05:41:49.333Z lane 23569 5202 18367 True -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1929/38c8307d-6f95-4adf-
2026-10-21T06:14:54.768Z lane 8068 3775 4293 False -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1737/87d69d3d-984a-4963-
2026-10-21T20:54:29.502Z lane 9071 4778 4293 False -Users-goos-MoAI-moai-adk-go--moai-worktrees-t1909/4a6d0ae6-fa22-4f63-
2026-10-22T18:26:15.251Z leader 5762 1467 4295 False -Users-goos-MoAI-moai-adk-go/3936623d-62a7-4ce8-b44e-74ba0ebe94b8.json
count 15
overflow 10
```

- 23,569 = 넘침 지시문 422 + 결합자 2 + **23,145**(카드 수치와 일치). 역할 블록 18,367은 t1617 압축 전 core(budget 테스트 주석의 「this-tree role block 18,369」와 같은 계열)다.
- 해당 세션의 `cwd`는 primary checkout(`/Users/goos/MoAI/moai-adk-go`)이었고, 현재 그 경로의 `factory-dispatch.md`에는 압축 전 문장(「The leader is the queue's sole producer」)이 0건이다(`grep -c` 0). 즉 10-21 당시 primary checkout이 낡은 규칙을 서빙했다(559커밋 뒤 detach 사고와 같은 날).
- 넘침 10건은 모두 10-21T05:42Z 이전. 이후 5건은 넘침 0, 최대 9,071.

라이브 lane·ko·startup 세그먼트 분해(`probe/research/decompose.py`로 t1909 세션 분해, 세그먼트 머리만 인용): attribution 509 · messaging bound 111 · GLM 359 · 합류 줄 38 · 스폰 권한 347 · 자동 모드 370 · 다음 카드 규칙 1,573 · 정지 복구 499 · binary-lag 421 · guard-liveness 118 · 버전 불일치 안내 411 · 역할 블록 4,293. 비-팩토리 합계 = 509 + 111 + 359 + 421 + 118 + 411 + 결합자 6×2 = **1,941** → 허용분 1,950.

### §B.4 현재 테스트가 위험을 못 보는 이유

명령: `go test -count=1 -run 'TestRoleInjectionAssemblyBudget$|TestCompositeBudgetRealAssembly$' -v ./internal/hook/` → exit 0. 원문: `probe/red-now-5995962990-intree.txt`. 발췌:

```text
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_composite_e2e_test.go:90: RESULT lane_ko_glm: composite utf16=9030 (cap 10000, margin=970)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
```

`TestRoleInjectionAssemblyBudget`는 생산자를 `strings.Repeat("a", 4797)`로 대체하므로 공지·규칙 텍스트가 커져도 결과가 변하지 않는다. `TestCompositeBudgetRealAssembly`는 실제 렌더를 쓰지만 en·ko 두 언어, startup 한 소스만 보고, 상한이 10,000(여유 판단 없음)이다. 영어 lane이 여유 54자로 통과 중이다.

### §B.5 AGENTS.md 바이트 상한

| 명령 | exit | 출력(원문) |
|---|---|---|
| `go test -count=1 -run 'TestCodexContractByteCeiling$\|TestCodexNestedTemplateDiscoveryBudget$' -v ./internal/config/` | 0 | `contract document AGENTS.md = 24501 bytes (ceiling 24576, headroom 75)` / `contract document internal/template/templates/AGENTS.md.tmpl = 23784 bytes (ceiling 24576, headroom 792)` / `--- PASS: TestCodexContractByteCeiling` / `--- PASS: TestCodexNestedTemplateDiscoveryBudget` (원문 `probe/red-now-5995962990-ceiling.txt`) |
| `git apply --check /tmp/adk-pr-recovery-20261021-xlr15jib/proposal-1862-byte-ceiling/BYTE-CEILING-PROPOSAL.patch` | 1 | `error: patch failed: AGENTS.md:164` / `error: AGENTS.md: patch does not apply` |
| `git cat-file -s 3f81377dc9:AGENTS.md` | 0 | `24807` |
| `grep -n "No main-variant\|WT worktrees live\|Three traceability" internal/template/templates/AGENTS.md.tmpl` | 0 | `169:**Three traceability carriers are then mandatory**, because the branch name no longer identifies` (교체 대상 두 문단은 템플릿에 없음) |

- 상한 상수: `internal/config/token_budget_guard.go:101` `const CodexContractByteCeiling = 24576`. 두 테스트는 `internal/config/token_budget_guard_test.go:84`, `:245`.
- CI FAIL(24,807바이트, main@3f81377dc9) 이후 `a1762cabe1`(#1869)이 AGENTS.md를 24,501로 줄여 기준 트리에서는 녹색이다. 카드의 RED는 현재 트리에서 재현되지 않으므로 해당 기준은 regression-guard로 분류한다(verification-completeness.md §2.1 undecidable disposition).
- 운영자 패치를 기준 트리의 같은 두 문단에 옮긴 후보 측정(scratch 스크립트, 트리 무변경): `head_bytes 24501` / `old_block_bytes 927 new_block_bytes 644` / `candidate_bytes 24218 ceiling 24576 headroom 358`. `only` 복원(+3바이트)을 더해도 24,221.
- 의미 대조(운영자 패치 vs 현재 문단): `WT-*main*` 금지, `main`은 PR 병합으로만 전진, 착지마다 local `main` 재동기화·앞서 있으면 결함, primary가 점유한 브랜치는 거절하고 HEAD를 detach하지 않음, `.claude/worktrees/`는 런타임 `agent-*` 전용, 카드 작업에 채택된 런타임 트리는 제자리에서 `WT-`로 개명·위치 유지 — 모두 보존. 약화 1건: 「live **only** under」→「belong under」(배타성 상실) → `only` 복원으로 수리. 제거 1건: 2026-10-21 사고 괄호(근거 문장, 의무 아님).

### §B.6 연결 카드

- t1637 (hold): (d) 「role-core 밖 SessionStart 고정분 4,797자 축소 — 줄어든 만큼 t1617 OUT 행을 IN으로 되돌림」. 본 SPEC은 앞 절반(고정분 축소)을 흡수하고 뒤 절반(core 행 복원, core 상한 개정)은 t1637에 남긴다. 4,797 자체는 숫자로 유지하되 「허용분 1,950 + 팩토리 고정분 2,847」로 분해해 실측 가능하게 만든다 — 이후 고정분 실측이 2,847보다 작아진 만큼이 t1637이 청구할 여유다.
- t1780 (picked): 훅 비동기 전환. 출력 텍스트 크기와 무관한 별개 축이라 범위 겹침 없음.
- t1617(SPEC-ROLE-INJECTION-BUDGET-001, completed): 예산 계약 원류. 0.6.0 수치 유지.

## §C 사전 점검 (run 단계 진입 시)

1. `git rev-parse HEAD`가 이 SPEC의 기준 트리 또는 그 후손인지 확인.
2. `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/` 가 RED(exit 1)임을 재관측 — run 단계의 RED 증거(E8).
3. 테스트가 고정한 공지 문구 목록을 먼저 뽑는다: `grep -rn "autoModeGuide\|laneSpawnAuthority\|docsPointer\|leaderClasses\|leaderStagger\|gateSummary\|laneNextCardRule\|laneRecheckRule" internal/hook/*_test.go`.
4. 무거운 `go test` 앞에는 `moai slot acquire --resource go-heavy` 임대, `go test ./...` 금지.

## §D 제약

- role core 무손실: 역할 게이트 규칙 2개와 그 템플릿 미러는 편집 금지(REQ-FBM-007). 재배치 목적지는 상세 동반 문서(예: `factory-dispatch-detail.md` / `factory-dispatch-mechanics.md`) 또는 todo 워크플로 스킬 — run 단계가 고르되 역할 게이트 파일은 제외.
- 검사 삭제·skip·상한 상향 금지(REQ-FBM-009).
- 프로토콜 토큰 번역 금지 규칙(i18n 파일 머리 주석) 유지 — 명령·경로·도구 이름은 4개 언어 모두 원문.
- AGENTS.md 편집은 상시 로드 파일이므로 작업 마지막에 한 번(cache-aware-execution 지시 3).

## §F 마일스톤 (결정 가능성이 큰 것부터)

### M1 — 공지 계약 결정과 테스트 RED (Priority High)

- 최소 원소 목록(REQ-FBM-001/002)을 테스트로 먼저 고정: 기존 guide 테스트 중 자동 모드 설명·리더 블록을 요구하는 것(`TestRedAutoModeGuidancePresent`, `TestFactoryGuideAutoModeMarkerBlocksEveryLocale`, `TestFactoryGuideNewFieldsCompleteInEveryLocale`, `TestFactoryGuideAutoLeaderDescribesInSessionWorker`, `TestFactoryGuideLaneAudienceLeaderModeNoSelfReference` 등)을 새 계약(자동 모드 설명 부재 + 포인터 줄 존재 + docs URL 존재)으로 다시 쓴다. 삭제가 아니라 계약 반전이다(REQ-FBM-009).
- 예산 테스트 재도출(REQ-FBM-008): `TestRoleInjectionAssemblyBudget`의 합성 생산자 4,797을 「허용분 1,950 + 실제 렌더 고정분」으로 바꾸고 고정분 ≤ 2,847 단언 추가, 역할 × 소스 × 언어 × 디스패치 × 두 면 전부 순회, 순회 칸 수 단언(빈 sweep 방지). `TestCompositeBudgetRealAssembly`의 언어(ja·zh)와 소스(clear·compact) 확장, 비-팩토리 패드를 실측 1,950으로 갱신. 헤더 주석에 0.6.0 → 본 SPEC 회계 변경 이력 기록.
- 이 단계 끝에서 새 테스트는 RED여야 한다(E8 증거).

### M2 — 재배치 목적지 작성 (Priority High)

- 제거될 원소(자동 모드 2플래그 설명, 카드 라우팅, stagger, 배치 게이트 요약, 운영 상태 단서, 진입점 안내, 팬아웃 상한, 레인 규칙 상세 문장)를 목적지 문서에 먼저 옮긴다. 이미 같은 의무가 다른 규칙에 있으면 새로 쓰지 않고 그 위치를 포인터로 쓴다(단순성 사다리 2단).
- 템플릿 미러 먼저, `make build` 후 배포본 대조.

### M3 — i18n 메시지 테이블 축소 (Priority High)

- `session_start_factory_i18n.go` 4개 언어에서 제거 필드 삭제, 포인터 줄 필드 추가, 레인 규칙 상세 문장을 포인터로 대체(REQ-FBM-006 토큰 보존).
- `session_start_factory.go`의 `factoryLeaderNotice`/`factoryLaneNotice`를 최소 원소만 조립하도록 수정. 운영자 사본(`factoryBootstrapNotice`)도 같은 렌더러라 자동으로 맞춰진다.
- 목표 크기: 영어 lane startup 고정분 3,751 → ≤ 2,847(공지에서 자동 모드 블록 −458 → 약 578; 레인 규칙 2,713 → ≤ 2,267, 즉 ≥ 446 감축).

### M4 — 검증 (Priority Medium)

- 탐침(AC-FBM-001) 녹색, 예산·guide 테스트 녹색, 스폰 권한 테스트 녹색.
- 역할 게이트 파일 diff 0(AC-FBM-007).

### M5 — AGENTS.md 문단 교체 (Priority Medium, 마지막)

- §B.5 후보 + `only` 복원을 적용 → ≤ 24,276바이트, Codex 상한 테스트 녹색.

### M6 — 라이브 확인 (sync 단계, Priority Medium)

- 착지 바이너리 설치 뒤 6레인 + 리더 기동, `probe/live_scan.py`로 합본·넘침 확인(AC-FBM-008).

## §G 위험

- 레인 규칙 축소 중 REQ-SD-019 등 다른 SPEC이 고정한 토큰을 잘라낼 위험 → REQ-FBM-006과 기존 테스트가 잡는다.
- 공지가 짧아져 레인이 자동 모드나 착지 절차를 모를 위험 → 같은 의무가 role core(「Before any work, read factory-dispatch.md to the end」)와 재배치 목적지에 남는다. 포인터 줄이 목적지를 이름으로 가리킨다.
- 비-팩토리 허용분 1,950은 GLM 세션 실측 기준. 새 생산자가 SessionStart에 추가되면 허용분을 넘을 수 있다 → 실조립 테스트(`TestCompositeBudgetRealAssembly`)가 10,000 상한으로 계속 감시한다.
- AGENTS.md 여유 300바이트 목표는 판단값이다(rule-authoring의 1,000바이트 단일 편집 고지 의무보다 작은 완충). 다음 규칙 추가가 여유를 다시 잠식할 수 있다.

## §H 참조

- SPEC-ROLE-INJECTION-BUDGET-001 (예산 계약 원류, 0.6.0)
- SPEC-SESSION-START-GUIDE-I18N-001 (자동 모드 안내·docs 포인터의 기존 계약 — 본 SPEC이 자동 모드 안내 부분을 대체)
- SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019 (레인 규칙 토큰)
- `internal/hook/role_rules.go:52`, `:638` / `internal/hook/session_start.go:600-680` / `internal/config/token_budget_guard.go:101`
