---
id: SPEC-AGENT-EMIT-LINEAGE-001
title: "에이전트 정의 방출 계보 — 로컬 즉시 드리프트 검사 + 임베드 축 판정 지점"
version: "0.6.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "Makefile, internal/cli, internal/template/agentemit, internal/template/templates/.codex/agents/moai, CLAUDE.local.md"
lifecycle: spec-anchored
tags: "agentemit, codex-agents, template-first, drift-guard, embed, mutation-testing"
tier: M
era: V3R6
---

# SPEC: 에이전트 정의 방출 계보 가드

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-27 | manager-spec | 최초 작성 — t317 실측(`.moai/reports/t317/measurement.md`) 위에 (a′)+(b′)+(c 잔여) 3갈래로 구성 |
| 0.2.0 | 2026-08-27 | manager-spec | iter-1 감사(`.moai/reports/t317/plan-audit-iter1.md`) 수리 — D2(REQ/AC-AEL-008 폐기), D3(REQ-AEL-004 기수·바이너리 부재 조항 + AC-AEL-003 확장), D4(REQ-AEL-002 주어 확대), D5(AC 헤딩에 REQ id 병기), D6(AC-AEL-006 RED 대조군 + 기대 출력 정정). 요구 8→7 / 수락 8→7 |
| 0.3.0 | 2026-08-27 | manager-spec | D1 clarification 종결(운영자 결정) — 임베드 축 검사의 자동 호출 지점을 `moai doctor` 항목 편입으로 확정. REQ-AEL-004 본문 확장(verb + doctor 이중 도달 + CI 빌드 잡 금지). 요구/수락 개수 불변(7 / 7) |
| 0.4.0 | 2026-08-27 | manager-spec | iter-2 감사(`.moai/reports/t317/plan-audit-iter2.md`) 수리 — D7(REQ-AEL-004 에 **적용가능성 술어** + 배포 프로젝트 not-applicable 거동 명문화, AC-AEL-003 에 verb 도달·doctor 도달·CI 미부착·적용 불가 4개 판정 단계 추가), D8(영향 파일 전수 열거 **5건** → Tier S 의 `< 5 files` 위반 → **Tier S → M 승격**, `module:` 에 `internal/cli` 추가, 수락 기준을 `acceptance.md` 로 분리). 요구/수락 개수 불변(7 / 7) |
| 0.5.0 | 2026-08-27 | manager-spec | iter-3 감사(`.moai/reports/t317/plan-audit-iter3.md`, PASS 0.90) D9 단독 수리 — AC-AEL-003 에 「하위 디렉터리 앵커」 게이트 1개 추가(적용 가능한 트리가 하위 디렉터리 실행에서 적용 불가로 뒤집히지 않음을 판정), REQ-AEL-004 적용가능성 술어에 **기준점 해석 방법** 절 1개 추가(`.moai/` marker 상향 탐색 — doctor 배선이 `os.Getwd()` 원값을 넘기므로). 요구/수락 개수 불변(7 / 7). D10·D11·D12 는 부채로 지고 run-phase 진입 |
| 0.6.0 | 2026-08-27 | manager-spec | run-phase 실측에 맞춘 요구 문면 정정 — REQ-AEL-004 적용가능성 술어의 **기준점 해석 방법**을 `.moai/` marker 상향 탐색에서 **커밋 방출물 자체를 앵커로 삼는 상향 탐색**으로 교체. 계기: v0.5.0 문면대로 구현한 뒤 D9 하위 디렉터리 앵커 게이트가 실패했다 — `internal/cli/.moai/state/`(untracked·gitignored, `internal/cli` 테스트 부산물)가 marker 탐색을 가로채 `internal/cli` 실행이 `ok — 적용 불가`, 저장소 루트가 `fail` 로 갈렸다. 재앵커 후 재측정: `root_exit == sub_exit == 1`(둘 다 `fail`), 원복 후 `sub_clean_exit=0`(`ok`). 구현은 `f3e5006ce` 로 이미 착지했고 이 판은 요구 문면을 구현에 맞춘 것이다. `findProjectRoot()` 선례 인용은 **상향 탐색의 선례**로 유지(앵커 대상의 선례가 아님을 명시). AC-AEL-003 은 문면 변경 없음(게이트가 거동으로만 서술돼 marker 를 언급하지 않는다). 요구/수락 개수 불변(7 / 7) |

---

## 1. 배경 — 이 카드가 왜 절차가 아니라 기계적 가드를 요구하는가

에이전트 정의는 사본이 셋이다.

| 사본 | 경로 | 성격 |
|---|---|---|
| C1 | `.claude/agents/moai/*.md` | 로컬 도그푸드, 손편집 |
| C2 | `internal/template/templates/.claude/agents/moai/*.md` | 배포 미러, 손편집 — 중립 원본 층 |
| C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **C2 로부터 기계 방출** (`internal/template/agentemit`) |

C2 → C3 만이 생성 관계다. C1 ↔ C2 는 바이트 동일 관계가 **아니며**(의도된 분기) 이 카드 범위 밖이다(실측 7).

### 1.1 발견 경위와 lane-16 의 짝 — 둘을 나란히 놓아야 크기가 보인다

이 결함은 **t301 감사의 부산물**이다. 감사관이 방출 배선을 겨눠서 찾은 것이 아니라, 어휘 토큰 전역 검색 중에 `.codex` 사본이 세 번째 사본으로 존재한다는 사실이 눈에 걸렸다. 우연이 잡아낸 것이므로 같은 경로로 다시 잡힐 것을 기대할 수 없다 — **그래서 이 카드의 처방은 절차가 아니라 기계적 가드여야 한다.**

lane-16(t316)이 같은 시기에 짝이 되는 발견을 냈다: 검증 기준이 **소스 파일 두 개를 비교**하는데, 사용자가 실제로 받는 것은 컴파일 시점에 박히는 임베드 자산이다(`internal/template/embed.go:28`, `//go:embed all:templates`). 한쪽은 "재생성이 호출되지 않는다"고 말하고, 다른 쪽은 "재생성이 일어났는지를 소스로는 판정할 수 없다"고 말한다. **둘을 겹쳐 놓으면, 되돌려진 수리도, 되돌려졌다는 사실도 아무도 관측하지 못한다.**

### 1.2 실측이 정리한 것 (`.moai/reports/t317/measurement.md`)

- **실측 1**: `make build` 의 선행은 `templ-generate` 뿐 — `agents-emit` 은 어떤 타깃의 선행으로도 등장하지 않는다.
- **실측 2**: 소스 층 드리프트는 **CI 에서 이미 잡힌다.** `AGENTEMIT_UPDATE` 없는 기본 실행이 sha256 불일치를 세우고, CI 는 `go test ./...` 를 돌린다. 따라서 스테일한 `.toml` 이 머지되는 경로는 막혀 있다.
- **실측 5**: `build` 는 이미 트리를 변형한다(`*_templ.go` 쓰기, `catalog.yaml` in-place 쓰기). "빌드는 읽기전용이어야 한다"는 반론은 이 리포에서 약하다.
- **실측 6**: `catalog.yaml` 에 `.codex` 항목이 0건 — 방출물은 catalog 무결성 층 밖에 있다.
- **실측 8 (결정적)**: `TestEmbedFSPresenceAndByteEquality` 는 실패 메시지로 "embedded bytes differ from committed (run make build)" 를 선언하지만 **공허하다.** 커밋된 `.toml` 에 두 줄을 주입한 뮤턴트가 **생존**했고(대조군인 골든 본체는 죽었다), 원인은 `go test` 가 테스트 바이너리를 매번 새로 컴파일해 `//go:embed` 가 같은 커밋본을 읽어 들이기 때문이다 — 양변이 함께 움직이는 동어반복이다. **이 패키지에서 임베드 축을 판정하는 테스트는 하나도 없다.**

### 1.3 남은 진짜 창

실측 2 가 머지 경로를 막고 있으므로, 남는 창은 두 가지다.

1. **로컬 지연**: C2 를 고치고 재생성을 빠뜨린 상태는 **CI 가 빨간불을 세울 때까지** 아무 신호가 없다. 그 사이 `make build` 로 만든 `bin/moai` 로 작업이 계속된다.
2. **바이너리 노후**: `bin/moai`(또는 설치본)가 **커밋된 방출물보다 오래된** 상태는 소스↔소스 비교로는 원리상 보이지 않는다(실측 8). 판정 지점이 소스가 아니라 **빌드 산출물**이어야 하는 이유다.

---

## 2. 요구사항 (GEARS)

**REQ-AEL-001** — When `make build` is invoked, the build shall run a read-only agent-emit drift check **before** compiling the binary, and shall abort with a non-zero exit status naming every drifted path when the committed `.codex/agents/moai/*.toml` artifacts differ from the emission of the current `.md` source layer.

**REQ-AEL-002** — Every check this SPEC introduces — the source-layer drift check of M2 and the embed-axis check of M1 alike — shall not write to, create, or delete any file inside the repository tree; scratch output belongs outside the tree.

> 주어를 넓힌 이유(감사 D4): M1 의 실증된 추출 경로는 스크래치 디렉터리에 **프로젝트 전체를 배포**하는 무거운 쓰기 동작이다(감사관이 git init 과 훅 설치 동반을 관측했다). 대상 디렉터리를 잘못 잡았을 때 저장소 트리를 오염시키지 않을 보장이 요구 층에 있어야 한다. 종전 문면("The drift check")은 문맥상 M2 만 가리켰고, M1 의 무쓰기 의무는 `plan.md` 에 "같은 규율"이라는 표현으로만 존재했다 — 그 표현 자체가 구속하지 않음을 인정하는 것이었다.

**REQ-AEL-003** — Regeneration of the emitted artifacts shall occur only through the explicit `make agents-emit` verb. The build path shall never regenerate.

**REQ-AEL-004** — The project shall provide an embed-axis judgment point that compares the `.codex/agents/moai/*.toml` bytes **carried by a built binary** against the committed artifacts, and that runs without rebuilding the binary under test. The judgment point shall be reachable **both** as an explicit maintainer verb **and** as a `moai doctor` check item, and shall not be attached to a CI build job as its automatic trigger.

**Applicability.** The judgment point shall take the presence of the committed emission set in the tree under check as its applicability predicate: it is **applicable** when the glob `internal/template/templates/.codex/agents/moai/*.toml`, resolved relative to the project root under check — which the judgment point shall itself locate by walking up from the invoking working directory to the nearest ancestor **at which that same glob matches**, because the `moai doctor` wiring hands every check the raw `os.Getwd()` value and resolves no root of its own (`internal/cli/doctor.go:180`). The walk shall anchor on the committed emission set itself and **shall not** anchor on a marker directory such as `.moai/`, **because a marker directory can appear anywhere in the tree as build or test residue, so anchoring on a marker that is not the thing being judged makes the predicate answer a different question than the one asked** — "where is the nearest thing that looks like a project?" instead of "where is the committed emission set this check compares against?". A `.moai/` upward walk MAY still be used to phrase the not-applicable reason, which is a message concern and not the predicate. The in-package precedent `findProjectRoot()` at `internal/cli/glm.go:1058` remains cited as the precedent for **walking upward**, not for **what to anchor on**. The glob shall match one or more paths, and **not applicable** when it matches zero. The predicate shall be keyed on that committed path alone — a project-root `.codex/agents/moai/*.toml` set is a deployment output of this repository's templates, not a committed artifact of the tree under check, and shall not be substituted for the committed emission set. **Where the judgment point is not applicable**, it shall report `ok` naming the absent committed emission set as its reason, shall not report failure or warning, and shall leave `moai doctor`'s exit status unchanged — so a distributed user project, which carries no committed emission set, sees exactly one added `ok` row and the same exit status it had before.

**While the judgment point is applicable**, absence of a judgment target is failure, never success. When no readable binary exists at the judgment target path, the check shall exit **failure** — "0 comparisons → pass" shall not be reachable in a tree that carries a committed emission set. The check shall report how many paths it compared, and shall exit non-zero when that count is lower than the committed artifact count — a partially-successful extraction shall not pass by comparing a subset.

> 호출 지점(v0.3.0 결정): 종전 문면은 판정 지점의 **존재**만 요구하고 **어디서 자동으로 불리는지**에 침묵했으며, 그 공백이 `plan.md` M1 의 미해결 결정 마커였다. 운영자 결정으로 `moai doctor` 항목 편입이 확정돼 요구 층에 못박는다. CI 빌드 잡 금지 조항이 함께 들어가는 이유는 취향이 아니다 — CI 는 자신이 검사하는 커밋에서 빌드하므로 임베드 바이트와 커밋 산출물이 정의상 일치하고, 거기 건 검사는 **결코 실패할 수 없다**(실측 8 의 동어반복과 동형). 근거 전문은 `plan.md` M1 「자동 호출 지점」.

> 적용가능성(v0.4.0 결정, 감사 iter-2 D7): 종전 문면은 「doctor 항목으로 도달 가능」과 「대상 부재 시 실패」를 한 문장 안에 겹쳐 놓아, 문면대로 구현하면 **모든 배포 사용자 프로젝트의 `moai doctor` 가 exit 1 로 뒤집힌다.** 배포 프로젝트에는 `bin/moai` 도 커밋 산출물 경로도 없고(이 세션 실측: `moai init` 한 스크래치 프로젝트에 `internal/template/templates/.codex/agents/moai/` 부재, 루트 `.codex/agents/moai/*.toml` 은 11건 존재, `moai doctor` → `doctor_exit=0`), Fail 하나가 exit 1 로 승격되기 때문이다(`internal/cli/doctor.go:121` 의 `doctorExitStatus(failCount)`, 정의는 `:140-146`; `:47-48` 의 `Long` 도움말이 같은 계약을 문서로 적는다). 위 조항은 **새 결정이 아니라 운영자가 이미 준 결정에서 파생된 것**이다: 운영자가 승인한 비용은 `plan.md` 「받아들인 비용」의 **행 1개 추가**이고, 종료 코드 변경은 그 비용에 포함되지 않는다. 그래서 서로 다른 두 부재를 갈랐다 — **적용 불가**(대조할 커밋 산출물이 애초에 없다 → `ok`, 종료 코드 불변)와 **적용 가능하나 판정 불가**(커밋 산출물은 있는데 바이너리를 읽을 수 없다 → 실패). 후자가 D3 이 요구한 공허성 봉투이며 그대로 남는다. `ok` 로 흘리는 형태는 이 코드베이스의 선례와 같다 — `internal/cli/doctor_mcp_version.go` 의 `checkMCPServerVersionAgainst` 는 `len(live) == 0` 일 때 `uikit.CheckOK` + `"no running moai MCP server recorded"` 를 낸다(파일을 읽어 확인). `uikit` 의 상태 열거는 `ok`/`warn`/`fail` 셋뿐이라(`internal/cli/uikit/types.go:12-17`) "건너뜀" 상태는 존재하지 않으므로, 적용 불가는 `ok` 로 표현하는 것이 유일한 선택지다.

> 공허성 봉투(감사 D3): 이 검사가 보완하는 기존 테스트는 기수를 단언한다 — `golden_test.go:285` 의 `if count != 11 { … want 11 }`(이 트리 이 실행에서 `ls internal/template/templates/.codex/agents/moai/*.toml | wc -l` → `11` 로 재확인). 새 검사에 대응 조항이 없으면, 추출이 부분 성공해 일부 경로만 비교해도 통과하고, 바이너리 부재 시 "비교 0건 → exit 0" 이 규정상 허용된다. 둘 다 이 SPEC 이 근절하겠다고 선언한 실패 유형 그 자체다.

**REQ-AEL-005** — When the bytes carried by the binary under test differ from the committed artifacts, the embed-axis check shall exit non-zero and name every differing path.

**REQ-AEL-006** — The maintainer edit procedure documentation shall state that any edit to `internal/template/templates/.claude/agents/moai/*.md` obliges `make agents-emit`, and shall name both the source layer and the emitted layer by path.

**REQ-AEL-007** *(minor)* — The `agents-emit` target shall be declared in the Makefile `.PHONY` list, so that a same-named path on disk cannot make it silently skip.

> **폐기 — 종전 REQ-AEL-008 (update 분기 쓰기 후 되읽기)**. iter-1 감사 D2 가 짝인 AC 의 공허함을 실행으로 확증했고, 나는 이 트리에서 그것을 재현한 뒤 **요구 자체를 폐기하는 쪽**을 골랐다. 근거는 아래 「폐기 판정」 절에 적었다. 요구 번호는 재사용하지 않는다 — 007 이 마지막이며 총 7건이다.

---

## 3. 수락 기준 — `acceptance.md` (Tier M)

수락 기준 **7건(AC-AEL-001 ~ 007)** 은 `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/acceptance.md` 에 있다. v0.3.0 까지는 이 절에 인라인이었고(Tier S 산출물 규약), v0.4.0 이 영향 파일 전수 열거로 Tier 를 **M** 으로 재판정하면서 분리했다 — 산술과 근거는 `plan.md §B`.

| REQ | 피복 AC |
|---|---|
| REQ-AEL-001 | AC-AEL-001, AC-AEL-004 |
| REQ-AEL-002 | AC-AEL-002, AC-AEL-004 |
| REQ-AEL-003 | AC-AEL-002, AC-AEL-005 |
| REQ-AEL-004 | AC-AEL-003 |
| REQ-AEL-005 | AC-AEL-003 |
| REQ-AEL-006 | AC-AEL-007 |
| REQ-AEL-007 | AC-AEL-006 |

미피복 REQ 0건, 고아 AC 0건. AC 4건(001/002/003/006)이 뮤테이션 확립이다.

---

## 4. 범위 밖 (exclusions)

아래는 이 SPEC 의 out of scope 이며, run-phase 가 여기로 흘러가서는 안 된다.

### Out of Scope — C1 ↔ C2 패리티
- `.claude/agents/moai/*.md` 와 배포 미러 사이의 관계는 바이트 동일성이 아니다(실측 7 — 7개 파일 differ, 표본 `manager-develop.md` 의 차이는 `isolation: worktree` 한 줄). 패리티 가드 부재를 결함으로 주장하지 않는다.
- 측정된 7건의 차이를 조사·정렬하지 않는다.

### Out of Scope — CI 파이프라인 일반
- 리드가 언급한 `graph-freshness` CI 실패는 이 계보와 무관하다. 이 SPEC 은 그것을 고치지 않는다.
- 소스 층 드리프트의 CI 검출은 **이미 존재한다**(실측 2). 새로 만들지 않는다.

### Out of Scope — catalog 무결성 확장
- `catalog.yaml` 이 `.codex/agents/moai/*.toml` 을 한 항목도 담지 않는다는 사실(실측 6)은 **관측으로 기록**할 뿐 요구사항이 아니다. 방출물을 catalog 해시 층에 편입하는 일은 별도 후속 카드 후보다.

### Out of Scope — 골든 update 분기 전반
- "재생성 이후에도 update 모드가 판정하게 만들기"는 기각한다(실측 3): 재생성 명령이 재생성을 하고 나서 자기가 한 일을 실패로 보고하는 자기모순이다.
- 쓰기 후 되읽기 확인(종전 REQ/AC-AEL-008)도 v0.2.0 에서 **함께 폐기**해 범위 밖으로 옮겼다. 판정 근거는 아래 「폐기 판정」 절.

### Out of Scope — `agentemit` 밖의 임베드 검증 공허성
- 같은 동어반복이 `internal/template` 의 다른 임베드 테스트에도 있는지는 조사하지 않는다(실측 8 Gaps). 후속 카드 후보로만 기록한다.

---

## 5. 제약

- **읽기전용 불변**: REQ-AEL-002 는 협상 대상이 아니다. build 경로가 방출물을 고치는 순간 이 SPEC 의 목적이 뒤집힌다.
- **비용**: 소스 층 검사의 실측 비용은 `go test ./internal/template/agentemit/...` 0.419s(실측 5). 이 규모를 넘어서면 설계를 다시 본다.
- **기존 동사 보존**: `make agents-emit` 의 의미(재생성)는 바뀌지 않는다(AC-AEL-005 가 이를 잠근다).
- **Tier M 예산**: 요구 **7** / 수락 **7**. 상한은 각각 **16** 이며 두 축에 독립 적용된다(`spec-workflow.md` § SPEC Complexity Tier). v0.4.0 의 Tier 승격을 만든 축은 **항목 수가 아니라 파일 수**다 — 항목 수는 v0.2.0 이 008 쌍을 폐기한 이후 7/7 로 변함이 없고, v0.4.0 의 수리도 기존 REQ·AC 본문 확장뿐이다. 여유가 생겼다고 항목을 채우려 들지 않는다.

## 폐기 판정 — 종전 REQ/AC-AEL-008 을 왜 고치지 않고 없앴는가

감사 D2 는 두 갈래를 제시했다. (a) AC 를 **내용 불일치** 면으로 교체하고 대조군을 붙인다, (b) 요구와 수락을 **함께 폐기**한다. (b) 를 골랐고, 사유는 감사가 든 것("권한 면은 이미 닫혀 있다")보다 강하다.

**1. 공허함을 내 손으로 재현했다.** 미수정 코드에서, 이 트리 이 실행:

```console
$ chmod 444 internal/template/templates/.codex/agents/moai/manager-git.toml
$ make agents-emit; echo exit=$?
    golden_test.go:98: update write .codex/agents/moai/manager-git.toml: open ../templates/.codex/agents/moai/manager-git.toml: permission denied
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.419s
make: *** [agents-emit] Error 1
exit=2
$ chmod 644 internal/template/templates/.codex/agents/moai/manager-git.toml
```

종전 AC-AEL-008 은 REQ-AEL-008 을 한 줄도 구현하지 않은 상태에서 이미 GREEN 이다. 구현 전후를 구분하지 못한다.

**2. 요구가 겨눈 두 실패 유형은 이미 둘 다 닫혀 있다.** 실측 3 이 든 근거는 "부분 쓰기·권한 실패"였다. 권한 실패는 위 실행이 보여준다. 부분 쓰기도 마찬가지다 — `os.File.Write` 는 `n != len(b)` 일 때 `io.ErrShortWrite` 를 돌려주고(`$(go env GOROOT)/src/os/file.go:219-221`), `golden_test.go:97-99` 의 `t.Fatalf` 가 그 오류를 그대로 세운다. 즉 요구가 명시적으로 든 두 갈래 모두 **기존 코드가 이미 판정한다.**

**3. 남는 실패 유형은 되읽기로 닿지 않는다.** 위 둘을 빼면 남는 것은 "쓰기가 오류 없이 끝났는데 저장된 바이트가 다르다" 뿐인데, 쓰기 직후 같은 프로세스에서 되읽으면 페이지 캐시가 방금 쓴 바이트를 그대로 돌려준다 — 디스크에 무엇이 있는지 묻지 못한다. 그러므로 REQ-AEL-008 이 지시하는 가드는 **내구성을 검증하는 것처럼 보이지만 검증할 수 없는 가드**다. 그것이 바로 이 SPEC 이 근절하겠다고 선언한 공허성의 정의다. (a) 로 갔다면 공허한 AC 하나를 공허한 요구 하나로 바꾸는 셈이었다.

**4. 비용 대비.** (a) 를 성립시키려면 쓰기와 되읽기 사이에 외부 변조를 끼워 넣을 시험 이음매를 제품 코드에 내야 한다 — 실측 3 이 "작은 사안"으로 규정한 항목을 위해 표면을 넓히는 일이다. M1 2순위 폴백을 지운 것과 같은 판단 기준을 적용했다.
