---
id: SPEC-CI-DOCTOR-BIN-001
title: "doctor 임베드 축 검사 — 바이너리 부재 시 정보성 skip (develop CI 구조적 적색 수리)"
version: "0.1.0"
status: completed
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, embed-check, fail-open, severity, ci-red"
tier: S
era: V3R6
related_specs: [SPEC-AGENT-EMIT-LINEAGE-001]
---

# SPEC: doctor Agent Emit Embed 검사의 바이너리 부재 verdict 수리

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-28 | manager-spec | 최초 작성 — 카드 t346 실측( `gh run list --branch develop --workflow ci.yml` 전부 실패, 812ee01fc 이후) + 본 워크트리 1차 재현 위에 구성 |

---

## 1. 배경 — 왜 develop CI가 구조적으로 빨간가

t317(SPEC-AGENT-EMIT-LINEAGE-001)이 도입한 `moai doctor`의 "Agent Emit Embed" 검사
(`internal/cli/doctor_agentemit_embed.go`)는, 검사 대상 트리가 커밋 방출물 세트
(`internal/template/templates/.codex/agents/moai/*.toml`)를 갖고 있을 때 **적용 가능** 트리로 판정되고,
그 트리에 판정 대상 바이너리 `bin/moai`가 없으면 `CheckFail`을 낸다
(`doctor_agentemit_embed.go:119-124` — "no readable binary to judge at …").

이 fail-closed 판정은 종전 SPEC 의 요구였다 — REQ-AEL-004 의 공허성 봉투 조항:
*"While the judgment point is applicable, absence of a judgment target is failure, never success."*
그러나 그 조항이 간과한 것은 **CI 의 go test 잡들(`test` — `ci.yml:104`, `test-race` — `ci.yml:209`)은 바이너리를 빌드하지 않는다**는 사실이다 — 그 잡들의 체크아웃에는 `bin/moai` 가 없다. (대조: `lint` 잡은 `ci.yml:367` 에서, `constitution-check` 잡은 `ci.yml:464` 에서 각자의 체크아웃에 `bin/moai` 를 빌드한다 — 다만 그 두 잡에서 doctor 테스트는 돌지 않는다.)

실패 전파 사슬 (전부 본 트리에서 확인):

1. CI 잡은 전체 스위트를 돌린다 (`.github/workflows/ci.yml:183` `go test -coverprofile… ./...`, `:238` `go test -race -count=1 ./...`).
2. `internal/cli`의 doctor 종단 테스트들(`TestRunDoctor_*`)이 `runDoctor(cmd, nil)` 로 **전체 검사 스위트**를 실행한다 (`coverage_improvement_test.go:699-721`).
3. 테스트 바이너리의 cwd는 `internal/cli/` — 검사의 상향 탐색(`findEmbedCheckRoot`)이 저장소 루트까지 올라가 커밋 방출물 세트를 찾고, 검사는 **적용 가능**으로 판정된다.
4. `bin/moai` 부재 → `CheckFail` 1건 → `doctorExitStatus(failCount)`(`doctor.go:140-146`)가 `doctor: 1 check(s) failed` 오류를 반환 → 테스트 `t.Fatalf`.

즉 **바이너리를 빌드하지 않는 모든 환경(go test 를 돌리는 CI 잡 전체)에서 이 검사는 정의상 항상 실패**한다.
판정하려는 대상(빌드 산출물의 임베드 바이트)이 그 트리에 존재하지 않는데 실패로 보고하는 것은,
바이너리 부재를 "판정 불가"가 아니라 "판정 결과 실패"로 분류하는 범주 오류다.

본 실측 (이 세션, 브랜치 `WT-ci-doctor-bin` @ `4fdbd55c1`, `bin/moai` 부재 상태):

```console
$ go test ./internal/cli/ -run 'TestRunDoctor_WithExport$' -count=1
--- FAIL: TestRunDoctor_WithExport (6.11s)
    coverage_improvement_test.go:715: runDoctor error: doctor: 1 check(s) failed
FAIL	github.com/modu-ai/moai-adk/internal/cli	6.816s
```

카드 t346 본문의 대조 실측: `make build` 로 `bin/moai`를 만들면 같은 테스트들이 통과한다
(t234 워크트리에서 `ok 328.863s`). 원인 귀속이 확정돼 있다.

파급 범위 — 이 결함은 한 테스트의 실패로 그치지 않는다. doctor 의 종료 상태는
`doctorExitStatus`(`doctor.go:140-146`)를 거쳐 CI 전체 스위트 판정으로 흘러가므로(위 전파 사슬),
`812ee01fc` 이후 develop 에 착지한 모든 카드 — t303, t298, t317, t313, t335, t241, t340, t322,
t326, t234, 열 장 — 은 사용할 수 있는 녹색 CI 판정 없이 착지했다. 실측(본 세션 재관측, 카드 t346
본문과 일치): `gh run list --branch develop --workflow ci.yml` — 취소(concurrency) 런을 제외한
창 내 완료 런은 전부 실패며, 이 검사가 착지한 `da03d91`(t317 병합)부터는 `TestRunDoctor_*`가
`doctor: 1 check(s) failed`로 전량 적색인 본 결함의 구조적 적색이고(CI 로그 직접 확인), 그 이전
두 런(`812ee01`·`d34a789`)의 적색은 kanban Race 플레이크·Lint 였다. 후속 head 의 CI 결과를
선행 커밋의 판정으로 갈아 대는 조상 기반 판정(정당한 절차다)도 같은 적색을 그대로 물려받는다 —
신호 자체가 죽어 있어 어떤 후속 head 를 대입해도 녹색이 나오지 않는다. 그래서 이 수리의 실질
이익은 검사 한 개가 아니라, 일괄 착지분이 CI 판정을 다시 받을 수 있게 되는 것이다.

범위 귀속을 둘로 갈라 적는다 — 배경 사실과 본 카드의 소관은 다른 명제다. 배경: 이 구간
전체에서 CI 초록을 얻은 착지는 0 장이다(판정 난 것은 전부 failure, 나머지는 동시 push 의
concurrency 취소). 소관: t346 이 책임지는 구간은 `da03d9188`(t317) 이후뿐이고, 그 앞 두 런의
적색은 원인이 다르므로 이 카드가 고치지 않는다. 이 구분이 없으면 연속된 붉음을 하나의 원인으로
읽는 시간 축 압축이 생긴다 — 배차자가 처음 "열 장"으로 집계한 것이 바로 그 압축이었고
(붉음의 연속을 원인의 단일성으로 읽은 것), head 별 원인을 갈라 재산 두 축이 함께 참이 된다.
이는 앞서 `Graph Freshness` 상속 레드를 전부로 본 오독과 같은 형태다.

## 2. 요구사항 (GEARS)

**REQ-CDB-001** — While the embed-axis doctor check is applicable (the tree under check carries the committed emission set) and no readable binary exists at the resolved judgment-target path, the check shall report an informational skip — status `ok` with a message stating that no binary was judged — and shall contribute no fail to `moai doctor`'s exit status.

> 이 조항은 SPEC-AGENT-EMIT-LINEAGE-001 REQ-AEL-004 의 bin-absent-failure 절("While the judgment point is applicable, absence of a judgment target is failure, never success")을 **바이너리 부재 케이스에 한해 대체**한다. 선례는 `checkBinaryFreshness`(카드 t184, `internal/cli/doctor.go:502`)다 — 판정 불가 상황(개발 빌드, git 트리 밖, 비조상 커밋)에서 `ok` + 정보성 메시지를 내고 doctor를 gate 하지 않는다. 바이너리가 **있는데** 판정에 실패하는 경로는 REQ-CDB-003 이 그대로 지킨다 — 두 요구가 합쳐져야 t317 의 원래 목적(재생성 누락이 임베드 증거를 지우는 것을 막음)이 살아 있다. 이 대체는 검증 계층 파생물까지 포함한다 — 구 SPEC `acceptance.md:49-53` 의 AC-AEL-003 바이너리 부재 게이트(`BIN=/nonexistent/moai make embed-check` → exit ≠ 0)와 라이브 테스트 `TestAgentEmitEmbed_MissingBinaryFails` 의 기대 역전이며, 후자의 갱신은 plan.md B-기대역전·M1 이 소관이다.

**REQ-CDB-002** — The informational skip shall be distinguishable from a disabled or vacuously-passing check: the skip message shall name the judgment-target path that was absent and name the remedy (`make build`, or the `MOAI_EMBED_CHECK_BIN` override).

> 스킵이 `ok`로 표현되는 이상, "건너뛴 ok"와 "비활성화된 ok"를 메시지로만 가를 수 있다(`uikit` 상태 열거는 `ok`/`warn`/`fail` 셋뿐 — `internal/cli/uikit/types.go:12-17`, skip 상태는 존재하지 않는다). REQ-CDB-001 의 스킵이 실제로 발화했는지를 판정 가능하게 만드는 조항이다.

**REQ-CDB-003** — When a readable binary exists at the judgment-target path, every existing fail verdict shall be preserved unchanged: extraction failure, comparison failure, cardinality shortfall, and stale-bytes drift shall each still report fail. The check shall not be disabled, downgraded, or made unreachable on any input where the binary is present and judgeable.

> t317 이 이 검사를 만든 이유(SPEC-AGENT-EMIT-LINEAGE-001)는 수리로 무력화돼서는 안 된다. 바이너리가 있는 트리에서 낡은 임베드는 여전히 fail 이어야 한다.

**REQ-CDB-004** — The doctor test suite shall pass on a checkout with no `bin/moai`: the doctor-judgment tests that currently fail on a bin-absent tree (the `TestRunDoctor_*` family, e.g. `TestRunDoctor_WithExport`) shall pass without building the binary, and shall continue to pass when the binary is present.

---

## 3. 수락 기준 (Tier S — 인라인)

| REQ | 피복 AC |
|---|---|
| REQ-CDB-001 | AC-CDB-001 |
| REQ-CDB-002 | AC-CDB-002 |
| REQ-CDB-003 | AC-CDB-003 |
| REQ-CDB-004 | AC-CDB-004 |

미피복 REQ 0건, 고아 AC 0건.

**AC-CDB-001** — Given a fixture tree carrying at least one committed emission artifact and no readable binary at the resolved judgment-target path, When the check runs (via `checkAgentEmitEmbedAgainst` with a nonexistent binary path), Then the status is `ok`, the extraction function is never invoked (a skip must not attempt to execute a nonexistent binary), and the check contributes zero fails to doctor's exit status. (종전 `TestAgentEmitEmbed_MissingBinaryFails`의 기대를 뒤집는 테스트로 잠근다 — extractor 호출 플래그까지 단언한다.)

**AC-CDB-002** — Given the same fixture, When the check runs, Then the skip message names the absent judgment-target path and names a remedy (`make build` 또는 `MOAI_EMBED_CHECK_BIN` 키), so the message content distinguishes "skipped: nothing to judge" from "check disabled".

**AC-CDB-003** — Given a readable binary whose extraction errors, and a readable binary whose embedded bytes differ from the committed set, When the check runs on each, Then the status is `fail` in both cases — 기존 fail 경로 테스트들(`TestAgentEmitEmbed_ExtractionErrorFails` / `TestAgentEmitEmbed_DriftFailsAndNamesPath` / `TestAgentEmitEmbed_PartialExtractionFails`)이 수정 없이 계속 통과한다 (비회귀).

**AC-CDB-004** (two-cell) —

- **RED-now (관측됨)**: 브랜치 `WT-ci-doctor-bin` @ `4fdbd55c1`, `bin/moai` 부재 상태에서 `go test ./internal/cli/ -run 'TestRunDoctor_WithExport$' -count=1` → FAIL, `runDoctor error: doctor: 1 check(s) failed` (2026-08-28 본 세션 재현; 카드 t346 본문이 `TestRunDoctor_*` 9종 + 대조 실측을 전문으로 보유). RED 이유: §1의 전파 사슬 — 이 SPEC 이 고치는 바로 그 verdict 때문에 붉다.
- **Green path**: M1 이후 같은 명령이 같은 트리 상태(여전히 `bin/moai` 없음)에서 통과하고, `make build` 후에도 통과한다 (REQ-CDB-004).

---

## 4. 범위 밖 (exclusions)

### Out of Scope — CI 로그의 hook JSON 노이즈
- CI 로그에서 함께 보이는 훅 JSON 노이즈 줄들은 이 SPEC 이 고치지 않는다 (리드 지시).

### Out of Scope — doctor 의 다른 검사와 적용가능성 술어
- 커밋 방출물 세트 기반 상향 탐색(`findEmbedCheckRoot`), env override(`MOAI_EMBED_CHECK_BIN`), 그 외 모든 doctor 검사는 건드리지 않는다.

### Out of Scope — TestConcurrencyStress 동반 실패의 귀속
- 카드 t346이 같은 실행에서 `internal/kanban`의 `TestConcurrencyStress` 실패를 함께 기록했으나, 귀속 기제는 확립돼 있지 않다(본 트리에서 `internal/kanban`의 doctor 참조 0건 — grep 측정). 수리 후 재발할 때 별도 조사한다.

### Out of Scope — SPEC-AGENT-EMIT-LINEAGE-001 의 현장 수정
- 종전 SPEC 은 completed 상태로 두고, 대체 선언은 본 SPEC 본문으로만 한다. 종전 SPEC 의 frontmatter·본문을 고치지 않는다.

### Out of Scope — CI 워크플로 변경
- `ci.yml` 등 워크플로 정의는 수정하지 않는다. 수리 대상은 검사의 verdict 다.

---

## 5. 제약

- **영향 파일 2개**: `internal/cli/doctor_agentemit_embed.go`(1분기 verdict 변경 + 메시지)와 `internal/cli/doctor_agentemit_embed_test.go`(기대 갱신). 이 범위를 넘으면 설계를 다시 본다.
- **t317 목적 보존**: REQ-CDB-003 이 협상 대상이 아니다. 스킵은 "판정 불가"에만 적용되고 "판정 결과"를 대신하지 않는다.
- **기존 함수 구조 보존**: `checkAgentEmitEmbedAgainst`의 분기 순서 계약(`@MX:REASON` — 적용가능성이 판정대상 요구보다 먼저, 카디널리티 게이트가 바이트 비교보다 먼저)은 유지된다. 변경은 bin-absent 분기의 verdict 와 메시지뿐이다.
