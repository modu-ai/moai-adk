---
id: SPEC-BINARY-LAG-VISIBILITY-001
title: 배포 지연 가시성 — 설치된 바이너리가 저장소 HEAD보다 뒤처졌음을 요청 없이 알린다
version: 0.2.0
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: HIGH
phase: "v3.1.4 target"
module: internal/cli, internal/hook, build
lifecycle: spec-anchored
tags: "deployment-lag, doctor, session-start, version-stamp, observability, fail-open"
tier: M
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-08-27 | manager-spec | 최초 작성. 카드 t326 운영자 재범위(통합 락 결함 → 배포 지연 가시성) 반영. 사전 조사에서 **카드 지시문의 전제 1건이 반증**되어 범위를 재구성했다(§1.4). 요구 7 / 수락 7. Tier M |
| 0.2.0 | 2026-08-27 | manager-spec | clarification 2건 종결(리드·운영자 결정). (a) 발화 표면 = `additionalContext` 확정 — 아울러 리드가 **v0.1.0의 선례 인용을 정정**했다: `computeDeferredAdvisory`의 권고는 `HookOutput.Data`(`internal/hook/types.go:394`, `json:"-"`)로 들어가 **도달하지 않으므로**, 그 형태를 복사하면 결손이 재생산된다. 선례를 「일정」과 「발화」로 분리하고 REQ-BLV-008 + AC-BLV-008을 신설(§1.7.1). (b) `VERSION` 무변경 + 별도 `BUILD_ID` 확정 — REQ-BLV-004 재작성, 감수 비용을 §7.5 잔여 위험으로 명시, 제목 줄 축은 별도 카드로 이관. 인용 행번호 정정 3건(`doctor.go` 495 / 199 / 140). 요구 7→**8** / 수락 7→**8** |

---

## §1 배경

### 1.1 이 카드는 통합 락 결함이 아니다 — 운영자 재범위가 정본이다

큐에 남아 있는 t326 카드 본문은 여전히 통합 락 결함으로 읽힌다. **그 전제는 이 세션에서 반증됐고, 운영자가 카드를 「배포 지연 가시성」으로 재범위했다.** 나중에 이 문서를 읽는 사람이 카드 본문을 근거로 범위를 다시 유도하지 않도록 여기에 명시한다. 락 자체는 이 SPEC의 범위 밖이다(§7).

### 1.2 반증 — 관측된 것은 락 결함이 아니라 구버전 바이너리였다

모든 레인이 호출한 `moai` 바이너리는 커밋 `22df80e90`에서 빌드된 것이었다(빌드 스탬프 `2026-08-27T05:33:03Z`). 카드 t298의 수리는 `52c693327`에 착지했고, 이는 그보다 여섯 시간 뒤다. `git merge-base --is-ancestor 52c693327 22df80e90`는 거짓을 반환한다 — 설치된 바이너리는 수리 이전 코드였다. 「상호 축출」 사슬의 모든 관측은 t298 이전 코드 위에서 이뤄졌다.

### 1.3 A/B 재현 (격리 실행)

`integrationLockRoot()`(`internal/cli/integration.go:46`)가 `CLAUDE_PROJECT_DIR`를 먼저 참조하므로, 두 바이너리 모두 일회용 프로젝트 뿌리에 대고 실행했다 — 살아 있는 락은 건드리지 않았다. 동일 셸(pid 50875, 부모 35417):

| 사슬 단계 | 구버전 `22df80e90` | HEAD 빌드 (t298 이후) |
|---|---|---|
| 기록된 pid | 50882 = `acquire` CLI 프로세스 자신 | 35417 = 셸의 부모 = 세션 소유자 |
| 직후 `status` | `held by a session that is gone (reclaimable)` | `held` |
| 2차 `acquire`, `--force` 없음 | 조용히 축출됨 | 거부, exit 1 |
| `--force` | 해당 없음 | 축출하고 `displaced:` 기록 |

락 코드는 HEAD에서 올바르며, `--force`는 별도 변경 없이 의미를 되찾았다.

### 1.4 [반증] 지시문 전제 정정 — 「Link 1은 미커버」는 거짓이다

카드 지시문은 세 고리 사슬(**저장소 HEAD → 설치된 바이너리 → 실행 중 프로세스**)에서 Link 1이 미커버이며 그것이 이 SPEC의 전부라고 규정했다. **사전 조사에서 이 전제가 반증됐다.**

`checkBinaryFreshness`(`internal/cli/doctor.go:495`)는 **이미 존재하고, 이미 doctor 검사 목록에 등록돼 있으며**(`internal/cli/doctor.go:199`의 `{"Binary Freshness", checkBinaryFreshness}`), 정확히 Link 1을 계산한다 — 바이너리 커밋을 `git rev-parse HEAD`와 대조하고, 조상이면 WARN한다. 이 워크트리에서 실측:

```
$ moai doctor --check "Binary Freshness"
  ok  Binary Freshness  binary matches source HEAD (343399d2f)
rc=0
```

오늘의 실패 시점에 이 명령을 실행했다면 `binary is behind source tree (binary: 22df80e90, HEAD: …)`를 출력했을 것이다 — 비교 로직이 정확히 그 조건에서 WARN하도록 쓰여 있다.

**따라서 결손은 「비교의 부재」가 아니라 「판정의 도달 실패」다.** 판정은 존재했지만, 그것을 발화하는 유일한 경로가 사람이 손으로 `moai doctor`를 치는 것이었고, 5단계 실패 사슬 어디에도 그 단계가 없었다. 자동 호출 지점을 전수 조사한 결과 0건이다(`.claude/hooks/`, `.claude/settings.json`, `internal/hook/` 대상 grep — `moai doctor` / `runDoctor` 호출 0건).

이 SPEC은 그러므로 **비교를 새로 만들지 않는다**. 이미 있는 판정을 요청 없이 관측자에게 도달시키는 일과, 사람이 실제로 읽는 신원 문자열이 순서를 뒤집지 않게 하는 일, 둘뿐이다.

### 1.5 같은 병의 두 번째 사례 — 신원 문자열이 순서를 뒤집는다

교체 후 설치된 바이너리는 `v3.1.2`를, 교체 전 바이너리는 `v3.1.3-rc.5`를 보고한다. 그런데 `git merge-base --is-ancestor 22df80e90 343399d2f`는 참이다 — 새 바이너리가 옛 것을 포함하는데도 **사람이 보는 버전 문자열은 더 낮게 읽힌다.** 버전 문자열만 비교한 독자는 사실과 정반대의 결론에 도달한다.

기계적 원인을 실측으로 규명했다. `Makefile:6`:

```make
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || …)
```

`--abbrev=0`이 커밋 거리 접미사를 제거한다. 이 워크트리 실측:

```
$ git describe --tags --abbrev=0   → v3.1.2
$ git describe --tags              → v3.1.2-494-g343399d2f
```

**494개 커밋이 하나의 문자열로 붕괴한다.** v3.1.2 이후의 모든 기본 빌드가 동일하게 `v3.1.2`로 찍히므로, 버전 문자열은 빌드 신원이 아니라 **태그 하한**이다. 앞선 rc 바이너리가 더 높게 읽힌 것은 git-flow rc 절차가 `VERSION=`을 명시로 넘겼기 때문이며, 명시 rc가 이후 기본 빌드를 앞지른다 — 이것이 관측된 역전의 정확한 기전이다.

`moai version`이 커밋 pill을 함께 출력한다는 점(`internal/cli/version.go:34`)은 사실이지만 결손을 메우지 못한다: 독자는 `343399d2f`를 보고도 그것이 자기 트리보다 앞인지 뒤인지 알 방법이 없다. 부족한 것은 커밋의 표시가 아니라 **대조**다.

### 1.6 기각된 탐지 방법 — run-phase에서 재발명하지 말 것

이 세션은 처음에 `strings <binary> | grep -c 'The pid recorded is the OWNING SESSION'`가 0을 반환한다는 것을 t298 부재의 근거로 제시했다. **리드가 이를 정당하게 기각했다**: 같은 grep은 **새 바이너리에서도 0을 반환한다**(해당 문구는 Go 주석이라 컴파일 산출물에 없다). 결론을 지탱한 것은 오직 빌드 커밋 스탬프와 조상 판정이었다. 바이너리 문자열 스캔은 이 도메인에서 판정 근거가 아니다 — REQ-BLV-007이 이를 요구 수준으로 고정한다.

### 1.7 운영자가 제기한 설계 질문과 그 답

> 오늘의 실패 경로: 세 레인이 각각 증상을 독립 관측 → 리드가 락 결함으로 카드 발행 → 배차 → 그제서야 원인 발견. **어느 단계에서 신호가 이것을 멈췄겠는가?**

| # | 단계 | 기계적 개입 가능성 | 판정 |
|---|---|---|---|
| 1 | 레인이 `moai integration acquire`를 호출, 이상 pid 관측 | 매 호출 조상 판정은 비용·소음 과다 | 부적합 |
| 2 | 레인이 증상을 리드에 보고 | 사람/에이전트 판단 시점, 기계 훅 없음 | 부적합 |
| 3 | 리드가 3건을 상관해 카드 발행 | 절차 층, 코드 아님 | 범위 밖 |
| 4 | 카드 배차 | 동상 | 범위 밖 |
| 5 | 원인 발견 | 이미 늦음 | — |
| **0** | **레인이 무엇이든 관측하기 이전 — 세션 시작** | **세션당 1회, 저비용, 모든 레인에 선행 도달** | **적합** |

답은 **단계 0**이다. 다섯 단계 중 어느 것도 기계적 개입점으로 적합하지 않고, 유일하게 값싸고 신뢰할 수 있는 지점은 관측이 시작되기 **전**이다.

두 번째 축은 단계와 무관하게 상시 필요하다: 지연을 의심한 사람이 가장 먼저 손을 뻗는 표면(`moai version`)이 오늘은 **적극적으로 오도한다**(§1.5).

#### 1.7.1 [정정] 선례는 「일정」에만 적용된다 — 「발화」에 적용하면 결손을 재생산한다

이 SPEC 초안(v0.1.0)은 `computeDeferredAdvisory`의 `status_drift_warning`을 발화 선례로 인용했다. **리드가 이를 정정했고, 실측으로 확인했다.**

그 권고 키들은 핸들러의 `Data` 맵으로 들어가는데, `HookOutput.Data`는 `internal/hook/types.go:394`에서 **`json:"-"`** 다 — 바로 위 주석이 「Internal data (not serialized to JSON)」라고 적고 있다. **그 권고는 관측자에게 도달하지 않는다.** 그 형태를 복사하는 것은 이 SPEC이 닫으려는 도달 불가를 **정확히 그대로 다시 만드는 일**이다 — 올바른 판정을 계산해서 아무도 렌더하지 않는 필드에 넣는 검사.

같은 파일이 이 결함 계열의 선례를 이미 기록하고 있다. `internal/hook/session_start.go:296` 인접 주석:

> `the existing `Data` field carries `json:"-"` and is internal-only (research.md §D.0 — structural root cause of the attribution dead feature)`

**도달이 실측된 표면은 `AdditionalContext`다.** `session_start.go:305`가 세션 귀속 문자열을 `hookSpecificOutput.AdditionalContext`에 쓰고, 그 문자열은 **이 세션의 시작 컨텍스트에 그대로 도착했다**(`moai session attribution: source_session_id=ae79f362-…`). `:343-346`과 `:369`가 비어 있으면 대입하고 아니면 `\n\n`로 덧붙이는 확립된 패턴을 보여준다.

따라서 `computeDeferredAdvisory`는 **일정(scheduling)의 선례**로만 채택한다 — 지연 실행, 최선 노력, 경계된 조인, 다음 세션에서의 멱등 재유도. **발화(emission)의 선례로는 채택하지 않는다.** REQ-BLV-008이 이 구분을 요구 수준으로 고정한다.

두 번째 축은 단계와 무관하게 상시 필요하다: 지연을 의심한 사람이 가장 먼저 손을 뻗는 표면(`moai version`)이 오늘은 **적극적으로 오도한다**(§1.5). 그 표면은 고쳐져야 한다.

---

## §2 목표

1. 이미 존재하는 지연 판정이 **요청 없이** 관측자에게 도달한다.
2. 빌드 신원 문자열이 **커밋 순서를 뒤집지 않는다**.
3. 배포판 사용자에게 아무 해도 끼치지 않는다 — 비교할 저장소 HEAD가 없는 프로젝트에서는 조용히 적용 불가로 판정한다.

## §3 요구사항 (GEARS)

**REQ-BLV-001** — The project shall emit a binary-lag verdict without an operator explicitly invoking a diagnostic command.

**REQ-BLV-002** — **When** a session starts in a tree where the installed binary's build commit is a strict ancestor of the tree HEAD, the session-start handler shall emit an advisory that names both the binary commit and the tree HEAD.

**REQ-BLV-003** — **Where** the executing tree provides no repository HEAD to compare against (a distributed user project, or any non-git directory), the lag verdict shall report not-applicable, shall emit no advisory, and shall leave the `moai doctor` exit status unchanged.

**REQ-BLV-004** — The build shall carry a monotone build identity, distinct from the release version string. **When** two commits stand in an ancestor relation, the build identity derived for the descendant shall differ from the one derived for the ancestor and shall carry a component identifying the descendant commit. The release version string (`VERSION`) shall not be changed by this SPEC.

**REQ-BLV-005** — The lag comparison shall have exactly one implementation, shared by the `moai doctor` check item and the unprompted emission point through a single substitutable seam.

**REQ-BLV-006** — The unprompted emission shall not block, delay beyond a bounded time box, or fail a session start. **When** the underlying comparison exceeds its time box or errors, the handler shall proceed and emit nothing.

**REQ-BLV-007** — The lag verdict shall not rest on the semantic-version string, nor on scanning the binary's embedded strings. The verdict shall rest on the build commit stamp and its ancestry relation to the tree HEAD.

**REQ-BLV-008** — The advisory shall be emitted on a surface that reaches the observer. The advisory shall be written into `hookSpecificOutput.additionalContext`, and shall not be written **only** into a field the hook contract excludes from serialization.

### 3.1 요구 ↔ 수락 추적

| 요구 | 수락 |
|---|---|
| REQ-BLV-001 | AC-BLV-001 |
| REQ-BLV-002 | AC-BLV-001, AC-BLV-002 |
| REQ-BLV-003 | AC-BLV-003 |
| REQ-BLV-004 | AC-BLV-004 |
| REQ-BLV-005 | AC-BLV-005 |
| REQ-BLV-006 | AC-BLV-006 |
| REQ-BLV-007 | AC-BLV-007 |
| REQ-BLV-008 | AC-BLV-008 |

수락 기준 본문은 `acceptance.md`(Tier M 분리)에 있다.

---

## §4 영향 파일 (전수 열거 — Tier 판정 근거)

| # | 경로 | 성격 |
|---|---|---|
| 1 | `internal/cli/doctor.go` | 기존 `checkBinaryFreshness`를 공유 seam 호출로 환원 |
| 2 | `internal/cli/`의 신규 파일 1건 (예: `binary_lag.go`) | 단일 비교 구현 + 적용가능성 술어 |
| 3 | `internal/hook/session_start.go` | `computeDeferredAdvisory`에 권고 1건 추가 |
| 4 | `Makefile` | `BUILD_ID` 도입 + ldflags 주입 (`VERSION ?=` 행은 불변) |
| 5 | `internal/cli/binary_lag_test.go` (신규) | AC-BLV-002·003·004·007 |
| 6 | `internal/hook/` 테스트 1건 | AC-BLV-001·005·006 |

6건 ≥ 5 → **Tier S의 `< 5 files` 기준 위반 → Tier M**. (t317이 동일 사유로 S→M 승격된 선례를 따른다.)

---

## §5 적용가능성 술어 — 배포판 사용자를 깨뜨리지 않는다

`doctorExitStatus`(`internal/cli/doctor.go:140`)는 **Fail이 하나라도 있으면 exit 1로 승격한다.** 배포판 사용자의 프로젝트에는 대조할 저장소 HEAD가 없으므로, 적용 불가 상황을 Fail로 처리하면 **모든 다운스트림 사용자의 `moai doctor`가 exit 1이 된다.**

t317 SPEC(`SPEC-AGENT-EMIT-LINEAGE-001` REQ-AEL-004, v0.5.0, plan-audit iter-3 PASS 0.90)이 임베드 축에 대해 이 문제를 이미 상세히 측정하고 결론지었다 — 「적용 불가 → ok 보고, exit 상태 불변」 경로가 필수라는 것. **그 추론을 재유도하지 않고 재사용한다.** 아울러 t317은 기준점 해석을 `.moai/` marker 상향 탐색으로 확정했는데(doctor 배선이 `os.Getwd()` 원값을 넘기므로), 하위 디렉터리 실행에서 적용 가능한 트리가 적용 불가로 뒤집히는 문제를 막기 위함이다 — 동일 위험이 여기에도 있으므로 같은 해석을 채택한다.

기존 `checkBinaryFreshness`가 이미 다섯 갈래를 ok로 처리한다(커밋 메타데이터 없음 / cwd 확인 불가 / git 트리 아님 / HEAD 일치 / 조상 아님=다른 브랜치). 이 SPEC은 그 관용을 **보존**하며, 축소하지 않는다.

> t317 인용은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317` 트리의 **읽기 전용 스냅샷**이며, 2026-08-27에 채취했다. 그 트리는 다른 레인의 살아 있는 워크트리이므로 이 SPEC의 어떤 단계도 그곳에 쓰지 않는다.

---

## §6 형제 작업과의 경계 (요구 수준 제약)

세 고리 사슬의 소관을 고정한다. 이 절은 해설이 아니라 **제약**이다 — t317과 이 SPEC이 둘 다 `moai doctor` 검사 항목을 추가하므로, 경계가 없으면 두 run-phase가 충돌한다.

| 고리 | 대조 대상 | 소관 | 상태 |
|---|---|---|---|
| Link 3 | 설치된 바이너리 ↔ 실행 중 프로세스 | `internal/cli/doctor_mcp_version.go` | **이미 종결** |
| Link 2 | 커밋된 산출물 ↔ 바이너리에 임베드된 바이트 | 카드 t317 / `SPEC-AGENT-EMIT-LINEAGE-001` REQ-AEL-004 | **진행 중** |
| Link 1 | 저장소 HEAD ↔ 설치된 바이너리 | **이 SPEC** | 판정은 존재(§1.4), 도달이 결손 |

**제약 C-1** — 이 SPEC의 구현은 `agentemit` 임베드 축을 검사하지 않는다. 그 축은 t317 소관이며, t317의 §Out of Scope가 `agentemit` 밖 임베드 검증을 명시적으로 제외해 이 자리를 비워 두었다.

**제약 C-2** — 이 SPEC의 doctor 항목은 **신규 항목이 아니라 기존 `Binary Freshness` 항목의 재배선**이다. 새 검사 이름을 추가하지 않는다. t317은 별도 이름의 항목을 추가하므로 이름 충돌이 발생하지 않는다.

**제약 C-3** — 이 SPEC은 실행 중 프로세스 축을 건드리지 않는다(Link 3 종결).

---

## §7 범위 밖

### Out of Scope — 통합 락
- 통합 락의 거동, liveness 앵커링, `--force` 의미론은 카드 t298이 전부 수리했고 이미 배포됐다. §1.3 A/B가 그 근거다.
- 「상호 축출」 사슬을 결함으로 재조사하지 않는다.

### Out of Scope — 바이너리 재설치 자체
- 재설치는 이 세션에서 리드가 이미 수행했다(`343399d2f` 설치, `moai version` rc=0). 구제 조치를 요구사항으로 만들지 않는다.
- 재설치를 자동으로 수행하는 기능은 이 SPEC이 만들지 않는다. 판정과 발화까지가 범위이며, 수리는 사람의 몫이다.

### Out of Scope — 임베드 축 (t317 Link 2)
- `.codex/agents/moai/*.toml` 바이트 대조는 t317 REQ-AEL-004 소관이다. 여기서 다루지 않는다.

### Out of Scope — 실행 중 프로세스 축 (Link 3)
- `moai mcp-server`의 실행 중 인스턴스 대조는 `checkMCPServerVersion`이 이미 닫았다. 재구현하지 않는다.

### Out of Scope — 비교 로직의 신규 작성
- `checkBinaryFreshness`의 조상 판정 로직은 정확하다(§1.4 실측). 새로 쓰지 않는다. 이 SPEC의 코드 변경은 재배선·단조성 수리·테스트에 한정된다.

### Out of Scope — 릴리스 버전 정책과 `VERSION` 자체
- SemVer 정책, 태그 명명, rc 절차는 건드리지 않는다.
- **`VERSION` 변수 자체를 바꾸지 않는다**(운영자 결정 b). 단조 신원은 별도 `BUILD_ID`로 들어간다. `RELEASE_BINARY`(`Makefile:14`), `version.json`(`Makefile:35`), `internal/update/local.go:65`의 소비 경로는 무변경으로 남는다.

### Out of Scope — `moai version` 제목 줄 축
- `VERSION`이 그대로이므로 제목 줄은 `v3.1.2`로 남고, 제목 줄만 읽는 독자의 오독 가능성도 남는다(§7.5). **이 축을 닫는 일은 별도 카드 소관이다** — 이 SPEC은 그것을 고치려고 범위를 넓히지 않는다.

### Out of Scope — 절차 층
- 리드의 카드 발행 판단, 레인의 보고 규율 같은 절차 개선은 코드 산출물이 아니므로 이 SPEC이 규정하지 않는다(§1.7 단계 3·4).

---

## §7.5 잔여 위험 — 인정하고 감수하는 비용

운영자 결정(b)에 따라 `VERSION`은 건드리지 않고 별도 `BUILD_ID`를 도입한다. 근거: 이 카드의 목적은 신호 도달이지 릴리스 넘버링 변경이 아니며, `VERSION`은 바깥으로 뻗는다 — `Makefile:14`가 `RELEASE_BINARY`에 보간하고, `Makefile:35`가 `version.json`에 쓰며, `internal/update/local.go:65`가 그 파일을 읽는다. `VERSION`을 고정하면 릴리스 산출물 이름·업데이트 경로·GoReleaser가 이 카드의 폭발 반경 밖에 남고, 변경이 되돌리기 쉬운 상태로 유지된다.

**감수하는 비용을 얼버무리지 않고 명시한다**: `VERSION`이 그대로이므로 `moai version`의 **제목 줄은 여전히 `v3.1.2`로 읽힌다.** 제목 줄만 보는 독자는 새 빌드를 옛 빌드로 여전히 오독할 수 있다 — 오늘 실제로 발생한 바로 그 오독이다. 빌드 신원은 **스탬프 줄에서만 판독 가능**하다.

**제목 줄 축을 닫는 일은 별도 카드로 미룬다.** 이 SPEC은 그것을 고치려고 범위를 조용히 넓히지 않는다.

## §8 미해결 관측 (요구사항 아님)

- `pkg/version/version.go:9`의 컴파일 기본값이 `Version = "v3.1.3"`인데 `Makefile`의 태그 파생 기본값은 `v3.1.2`다. 즉 ldflags 없는 맨손 `go build`가 태그 파생 빌드보다 **높은** 문자열을 찍는다. §1.5의 역전과 같은 계열이지만 별개 경로이므로, 관측으로만 기록하고 REQ-BLV-004의 판정 범위에 넣을지는 plan-phase의 판단에 맡긴다.
