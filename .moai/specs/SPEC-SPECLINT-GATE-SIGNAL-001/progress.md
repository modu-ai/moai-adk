# progress.md — SPEC-SPECLINT-GATE-SIGNAL-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
baseline_tree: dd1439502 (origin/develop tip at plan time)
baseline_measurement: 0 error(s), 4368 warning(s), rc=0 without --strict (.moai/reports/t525/spec-lint-baseline-dd1439502.txt)
kickoff_gate: 기제 모양 (i)-(iv) 선택은 운영자 결정 — plan.md §F M2 / spec.md §2.2

참고: plan-auditor 판정은 이 신호 이후 오케스트레이터가 수행한다.

## §E.2 Run-phase Evidence

### M1 — 판정 (2026-09-08, card t525)

측정 규율: 모든 수치는 **트리 빌드 `go run ./cmd/moai`**(installed 바이너리 e79c010b8 미사용 —
1069 커밋 지연, 리드 지시 2026-09-08). 카운팅/부재 판정은 `/usr/bin/grep`. attribution triple
(명령 / 관측 출력 / 트리 SHA) — 상세 원문 `.moai/reports/t525/{m1-demographics.md,verdict.md,census-b6efc874f.json}`.

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| M1-1 | 인구통계 재도출 (AC-SLGS-001) | `go run ./cmd/moai spec lint --json > .moai/reports/t525/census-b6efc874f.json` | rc=0, `jq length` → **4378** findings (warning 4378, error 0; advisory=true 4376 / advisory=false **2**) | `b6efc874f` |
| M1-2 | 비-strict 경로 초록 | `go run ./cmd/moai spec lint` | `0 error(s), 4378 warning(s)`, RC_NONSTRICT=0 | `b6efc874f` |
| M1-3 | (b) 구조 입증 (AC-SLGS-002 rc=1 쪽) | `go run ./cmd/moai spec lint --strict` | `0 error(s), 4378 warning(s)`, RC_STRICT=**1** | `b6efc874f` |
| M1-4 | 비-advisory 재고 = M4 쌍 | `jq -r '[.[] \| select(.severity=="warning" and (.advisory // false) == false)] \| .[] \| "\(.code)\t\(.file)"' census-b6efc874f.json` | `SpecsDirMissingSpecFile` × 2 (SPEC-V3R4-CC2X-ADOPT-001/002) | `b6efc874f` |
| M1-5 | 수치 동결 없음 (AC-SLGS-004) | `/usr/bin/grep -rnE '4[,.]?3(44|68)' .moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001 internal/spec internal/cli` | 12 히트 — 전부 SPEC 문서 4건의 출처 명시 인용; internal/spec·internal/cli **0** 히트 | `b6efc874f` |
| M1-6 | t518 미착지 (REQ-SLGS-011 전제) | `git merge-base --is-ancestor 6cfcfef00 origin/develop` | 비-조상 (T518_NOT_LANDED) — 흡수 병합 전후 2회 재측정 | `b6efc874f` / `2cee65571` |

M1 측정 전 흡수: `origin/develop` `a849d99d2` → `19cf21408` 병합 (창 67 커밋, `internal/spec/` 변경 0).

**M1 판정 요약** (원문 verdict.md):

- **(a)** 성립·지연 — advisory 재고 4,376건(12 규칙)은 실재하는 부채; t518 착지까지 상환 금지 (REQ-SLGS-011).
- **(b)** 구조 입증 — error 0 + 비-advisory 경고만으로 rc=1. 단, 전제 교정: **비-advisory 재고는 2건**(M4 쌍)이며 "수천 건" 전제는 역사적 표본이 advisory 비-표시 표였던 데서 온 추론으로 기각. era demotion(`dd644b5e0`, 2026-07-21)이 이미 대량 재고를 흡수한 상태. M4가 쌍을 "왜"와 함께 닫으면 bare `--strict`는 0 비-advisory → 초록(신호 회복). M2 기제(iii)의 잔여 가치: 규칙별 델타 가시성 + t518 인구 이동 흡수(M3 게이트 재기준). — kickoff 승인 범위 변경 아님, 전제 교정 보고.
- **(c)** 소관 외 — advisory 경계 자체는 t518 소관; 측정된 t518 스코프 = advisory 표시 4,376건.
- **AC-SLGS-002 mutation 절반**: M2 관측 창 소관 — M1에서 미관측, pending으로 기록 (통과 아님).

크로스플랫폼 빌드: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (M1 커밋 전, tree `b6efc874f`).

### M2 — 기제: 규칙별 기준선 래칫 (2026-09-08, card t525)

측정 규율: 모든 수치는 **트리 빌드**(`go build -o /tmp/t525-moai2 ./cmd/moai`, tree `b09b012ef`)
로 잰다. 부재·카운팅 판정은 `/usr/bin/grep`. attribution triple(명령 / 관측 출력 / 트리 SHA).

산출물: `internal/spec/lint_baseline.go` + `_test.go`(신규), `internal/cli/spec_lint.go`(플래그
3종 + 게이트), `internal/cli/spec_lint_test.go`(신규). 기준선 **파일 자체는 만들지 않았다** —
최초 산출은 M3.2 소관(`ls .moai/spec-lint-baseline.json` → `No such file or directory`).

**확정한 설계 결정** (plan.md §F M2 / 경계 사례 절이 run 단계에 위임한 항목):

| 항목 | 결정 | 근거 |
|---|---|---|
| 게이트 차원 | 규칙별 **비-advisory 경고 수**. advisory 는 기록도 게이트도 하지 않는다 | t518 이 advisory 인구를 통째로 옮기므로, 기록하면 착지 즉시 재기준이 강제돼 감사 절차가 일상이 된다 |
| 파일 위치 | `--baseline <path>` 로 호출자가 지정. 기본 경로 없음 | 기본 경로 폴백은 §D.3 의 "침묵 폴백 금지"와 충돌 |
| 직렬화 | `json.MarshalIndent(indent=2)` + 후행 개행. 규칙 키 정렬(encoding/json 이 map 키를 정렬), OS 경로 미포함 | §D.6 결정론. 플랫폼 무관 바이트 동일성 |
| 판정 순서 | ① error → ② 규칙별 증가 → ③ 통과. 둘 다 exit 1 이므로 **출력이 어느 원인인지 이름 붙인다**(`ERROR-GATED` vs `EXCEEDED`) | exit code 만으로는 순서가 관측 불가 — M-b 뮤턴트가 이를 실증 |
| `--baseline` 없는 파일 | exit 3, 경로를 이름 붙임. 오늘의 동작으로 폴백하지 않음 | 경계 사례 절, 침묵 폴백 금지 |
| `--baseline` + `--json`/`--sarif` | exit 3 거절 | 두 페이로드는 원시 finding 목록이라 기준선 차원이 없다 — 받아들이면 플래그를 조용히 무시하는 것 |
| `--baseline` + `--strict` | exit 3 거절 | spec.md §2.2 가 하이브리드 (iv) 를 기각한 사유(escalation 의미 중복)와 동일 |
| `--update-baseline` 중 error 존재 | 파일은 쓰되 exit 1 | REQ-SLGS-009 — 기준선은 error 를 담지 않으므로 재기준이 깨끗한 run 으로 읽히면 안 된다 |

**§E.2 Run-phase Evidence — M2**

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| M2-1 | RED(컴파일) — 구현 전 | `go test ./internal/spec/ -run 'TestComputeBaselineRules\|TestMarshalBaseline\|TestCompareBaseline\|TestLoadBaseline\|TestWriteAndLoadBaseline'` | `undefined: ComputeBaselineRules` 외 10건, `FAIL ... [build failed]`, rc=1 | `b09b012ef` |
| M2-2 | RED(단언) — 스텁 대상 | 위와 동일 `-v` | `rule count: got 0 (map[]), want 2` · `ErrorCount: got 0, want 1` · `expected an error for a missing baseline file` 등 8건 FAIL, rc=1 | `b09b012ef` |
| M2-3 | 공허한 초록 적발·제거 | 위 `-v` 관측 | `TestMarshalBaseline_Deterministic…` 가 스텁의 `(nil, nil)` 에 **PASS**. 비어있지 않음 가드 추가 후 재측정 → `MarshalBaseline produced no bytes — equality below would assert nothing`, rc=1 | `b09b012ef` |
| M2-4 | RED(CLI) | `go test ./internal/cli/ -run 'TestSpecLint' -short` | `unknown flag: --baseline` × 8 FAIL, rc=1 | `b09b012ef` |
| M2-5 | 틀린 이유 초록 적발·제거 | 위 관측 | `TestSpecLintUpdateBaseline_ErrorStillExitsOne` 이 `unknown flag` 의 rc=1 로 PASS. `unknown flag` 배제 가드 추가 후 → `rc=1 came from an argument error, not from the standing error`, rc=1 | `b09b012ef` |
| M2-6 | GREEN(spec) | `go test ./internal/spec/ -run 'TestComputeBaselineRules\|…' -v` | 9/9 PASS, `ok github.com/modu-ai/moai-adk/internal/spec 0.380s`, rc=0 | `b09b012ef` |
| M2-7 | GREEN(cli) | `go test ./internal/cli/ -run 'TestSpecLint' -short` | `ok github.com/modu-ai/moai-adk/internal/cli 3.908s`, rc=0 | `b09b012ef` |
| M2-8 | **AC-SLGS-005** 양방향 + 크로스프로세스 | `go test ./internal/cli/ -run 'TestSpecLintBaseline_CrossProcessBothDirections' -v` | `--- PASS: TestSpecLintBaseline_CrossProcessBothDirections (5.43s)`, rc=0 | `b09b012ef` |
| M2-9 | **AC-SLGS-002 mutation 절반** — `--strict` 쪽 | `/tmp/t525-moai2 spec lint --strict` (실제 저장소 코퍼스) | `0 error(s), 4378 warning(s)`, **RC_STRICT=1** | `b09b012ef` |
| M2-10 | **AC-SLGS-002 mutation 절반** — 기준선 흡수 쪽 | `/tmp/t525-moai2 spec lint --baseline /tmp/t525-real-baseline.json` (동일 코퍼스·동일 바이너리) | `baseline: OK` / `inventory: 4378 warning(s) total (advisory included), 2 non-advisory tracked across 1 recorded rule(s)`, **RC_BASELINE=0** | `b09b012ef` |
| M2-11 | **AC-SLGS-006** 실측 (재고 초록 + 총수 가시) | M2-10 과 동일 실행 | 위 `inventory:` 줄 — 재고가 사라진 게 아니라 흡수됐음이 관측됨 | `b09b012ef` |
| M2-12 | 재기준 기록(**AC-SLGS-008** 정상 경로) | `/tmp/t525-moai2 spec lint --baseline /tmp/t525-real-baseline.json --update-baseline --reason "…"` | `tree_sha: b09b012ef` / `date: 2026-09-08` / `reason: …` / `recorded: 2 non-advisory warning(s) across 1 rule(s)`, rc=0 | `b09b012ef` |
| M2-13 | **AC-SLGS-004** 수치 동결 없음 | `/usr/bin/grep -rnE '4[,.]?3(44\|68\|78)' internal/spec internal/cli` | 무출력, rc=1. 대조군 `/usr/bin/grep -rc 'baseline' internal/cli/spec_lint.go` → `41`, rc=0 (grep 유효 — 위 부재는 실측) | `b09b012ef` |
| M2-14 | 크로스 플랫폼 | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | 각각 rc=0 | `b09b012ef` |
| M2-15 | 커버리지 | `go test -cover ./internal/spec/... ./internal/cli/...` | `internal/spec` **90.3%**, `internal/cli` **81.4%** (전 18 패키지 ok, rc=0) | `b09b012ef` |
| M2-16 | 린트 | `golangci-lint run --timeout=5m internal/spec/... internal/cli/...` | `0 issues.`, rc=0 (착수 전 baseline 도 `0 issues.`) | `b09b012ef` |
| M2-17 | 서브에이전트 경계 | `/usr/bin/grep -rn 'AskUserQuestion' internal/spec/lint_baseline.go internal/spec/lint_baseline_test.go internal/cli/spec_lint.go internal/cli/spec_lint_test.go` | 무출력, rc=1 | `b09b012ef` |
| M2-18 | 전체 대상 패키지 | `go test ./internal/spec/ ./internal/cli/` (크로스프로세스 포함) | `ok internal/spec 70.114s` · `ok internal/cli 443.866s`, rc=0 | `b09b012ef` |

**AC-SLGS-002 판정 갱신**: M1 이 pending 으로 남긴 mutation 절반이 **관측됐다** — 동일 코퍼스·
동일 트리·동일 바이너리에서 `--strict` rc=1 ↔ `--baseline` rc=0. (b) 축 구조 입증 완결.

**뮤턴트 프로브** (verification-completeness §2 — 잡은 것과 **못 잡은 것** 양쪽 기록):

| 뮤턴트 | 내용 | 결과 | 판별 증거 |
|---|---|---|---|
| M-a | 비교가 항상 "증가 없음" 반환 (`case false:`) | **잡힘** | `TestCompareBaseline_IncreaseNamesTheRule` + CLI 2건 FAIL |
| M-b | error 검사를 기준선 검사 **뒤로** 이동 | **처음엔 못 잡음** → 테스트 추가 후 잡힘 | error 단독으로는 `Increases` 가 비어 fall-through 되어 같은 출력. **error 와 증가가 동시에 있을 때만** 순서가 드러난다 → `TestSpecLintBaseline_ErrorOutranksASimultaneousIncrease` 추가, 뮤턴트가 `EXCEEDED` 를 내며 FAIL |
| M-c | `--reason` 을 "존재함"으로만 검사(trim 안 함) | **잡힘** | `RejectsMissingOrEmptyReason/reason_whitespace_only` 만 FAIL — 이 서브케이스가 유일한 판별식 |
| M-d | 감소 경로에서 기준선을 조용히 재기록 | **잡힘** | `DecreasePassesAndLeavesFileByteIdentical` FAIL (sha256 비교) |
| M-e | advisory 를 기준선에 포함 | **잡힘 — 단, 단위 테스트에서만** | `TestComputeBaselineRules_CountsNonAdvisoryWarningsOnly` FAIL. **CLI 층 테스트는 전부 통과** — 가드의 경계: advisory 배제 결정은 CLI 층에서 커버되지 않는다 |
| M-g | `--baseline` 존재 사전검사 제거 | **못 잡음 — 그러나 결함이 아님** | `LoadBaseline` 이 같은 exit 3 + 경로 이름을 낸다. 사전검사는 **중복 가드**이고, 테스트는 구현이 아니라 관측 가능한 계약(exit 3 + 경로)을 단언한다 |
| M-g2 | 진짜 조용한 폴백(로드 실패 시 오늘의 동작으로) | **잡힘** | `FlagContractRejections/missing_baseline_file_…` FAIL |

**Gaps (관측하지 않은 것)**:

- M3 범위 전체 — CI 배선(`.github/workflows/spec-lint.yml` 는 여전히 벌거벗은 `--strict`),
  기준선 파일 최초 산출·체크인, t518 잠금 점검 기록. **이 마일스톤에서 건드리지 않았다.**
- M4(CC2X ADOPT-001/002) 미착수. `SpecsDirMissingSpecFile` 2건은 그대로 서 있다.
- 기준선 직렬화의 **실제 windows 실행** 미관측 — `GOOS=windows` 는 컴파일만 증명한다.
  결정론 근거는 (i) `encoding/json` 의 map 키 정렬, (ii) 레코드에 파일 경로가 없어 구분자
  의존이 없음, (iii) 삽입 순서 독립성 단위 테스트 — 실행 관측은 CI 의 windows 매트릭스 몫.
- 실 저장소 코퍼스에서의 **주입/제거 양방향**은 미관측(합성 코퍼스에서만 관측). 실 코퍼스
  주입은 `.moai/specs/` 를 오염시키므로 하지 않았다.

**Residual risk**:

- `--update-baseline` 은 error 가 있어도 파일을 **쓴 뒤** exit 1 한다. 의도된 설계(재측정
  결과는 남기되 판정은 적색)지만, error 상태에서 재기준한 파일이 커밋되면 그 사실이
  기준선 파일 자체에는 남지 않는다 — 사유 문구와 커밋 메시지가 유일한 기록이다.
- `resolveTreeSHA` 는 git 이 답하지 못하면 `"unknown"` 을 기록한다. 귀속 없는 수를 만들지
  않으려는 선택이지만, `unknown` 이 박힌 기준선이 커밋되면 그 자체가 감사 불가다 — M3 의
  최초 산출은 반드시 git 트리 안에서 실행해 실제 SHA 가 박히는지 확인해야 한다.
- 비-advisory 재고가 현재 2건뿐이므로(M1 판정), 기준선의 실효 가치는 t518 착지 후
  인구 이동을 흡수하는 데 있다. 지금 시점의 래칫은 회귀 방지용으로만 작동한다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### 잔여 실측 열거 (M1 이후, 2026-09-08, tree `b09b012ef`)

리드 지시(2026-09-08)에 따라 M2 이후 잔여를 문서 추론이 아니라 실측으로 확정한다.
카운팅·부재 판정은 `/usr/bin/grep`(셸 `grep` 은 ugrep 래퍼라 조용히 건너뛴다).

| # | 대상 | Command | Observed output | 판정 |
|---|---|---|---|---|
| R-1 | M2 기제 (CLI 측) | `/usr/bin/grep -c baseline internal/cli/spec_lint.go` | `0`, rc=1 | 미착지 |
| R-2 | M2 기제 (린터 측) | `/usr/bin/grep -c baseline internal/spec/lint.go` | `0`, rc=1 | 미착지 |
| R-2c | R-1/R-2 대조군 | `/usr/bin/grep -c strict internal/cli/spec_lint.go` | `3`, rc=0 | grep 유효 — 위 `0` 은 실측된 부재 |
| R-3 | 기준선 파일 | `ls -la .moai/spec-lint-baseline.json` | `No such file or directory` | 미생성 |
| R-4 | 신규 소스 파일 | `ls internal/spec/lint_baseline*.go` | `no matches found` | 미생성 |
| R-5 | CI 배선 | `/usr/bin/grep -n 'spec lint' .github/workflows/spec-lint.yml` | `58:        run: go run ./cmd/moai spec lint --strict` | 벌거벗은 `--strict` — 미배선 |
| R-6 | M4 대상 디렉터 | `ls .moai/specs/SPEC-V3R4-CC2X-ADOPT-001/ -002/` | 각각 `research.md` 1건뿐 | 미종결 |
| R-7 | t518 착지 (REQ-SLGS-011 게이트) | `git merge-base --is-ancestor a4fbaeb82 origin/develop` | rc=1 | **비-조상 — 미착지** (`origin/develop` = `91d25bc61`) |

**잔여 = M2 전체 + M3.1~M3.3 + M4.** M3.4(DAG 간선 `dependencies:` 추가 + 게이트된 재기준
1회)는 R-7 로 잠금 유지 — 이 카드에서 열지 않고 t518 착지 시 별건으로 연다.
M1 에서 pending 으로 남긴 AC-SLGS-002 mutation 절반은 M2 기준선 경로 착지 후 관측한다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters** (orchestrator, 2026-09-08):

- tier: M
- scope (files): ~8 (internal/spec/lint.go, internal/cli/spec_lint.go, .github/workflows/spec-lint.yml, baseline file, tests, docs)
- domain count: 2 (Go linter/CLI policy surface, CI wiring)
- file language mix: Go + YAML + JSON baseline
- concurrency benefit: LOW — coding-heavy, sequential dependency M1 verdict → M2 mechanism → M3 wiring
- Agent Teams prereqs: not requested

**Mode evaluation table**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | Multi-file Go + CI change, not a typo fix |
| serial | **yes** | Coding-heavy Tier M; M1→M4 ordered dependencies; single writer per milestone |
| fanout | no | Not research-heavy; coding-task parallelism caveat |
| sweep | no | <30 files, semantic (not mechanical-uniform) change |

Decision: serial
Justification: Anthropic coding-task parallelism caveat applies — the mechanism (M2) depends on the M1 demographic verdict, and CI wiring (M3) depends on the mechanism; nothing downstream of M1 is parallelizable. One manager-develop spawn per milestone, sequential.

Kickoff outcomes (operator decisions, 2026-09-08, AskUserQuestion round):

- kickoff: **approved** (plan-audit PASS 0.875 prerequisite met)
- mechanism (spec.md §2.2 reservation): **(iii) per-rule baseline ratchet** — CI gates via `--baseline <checked-in file>`; `--strict` retained for non-CI use
- progression mode: **autonomous** — ac_converge goal armed by the orchestrator after this gate; no per-milestone operator prompts; evidence reported at close
