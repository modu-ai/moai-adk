---
id: SPEC-SPECLINT-GITBLIND-001
title: "SPEC Lint 의 git 눈멂 — 조용한 skip 을 관측 가능하게 만들고 기준 ref 를 해소한다"
version: "0.4.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/spec/lint.go, internal/spec/drift.go, internal/spec/gitquery_cache.go, .github/workflows/spec-lint.yml"
lifecycle: spec-anchored
tags: "spec-lint, git-blindness, shallow-clone, silent-skip, ci-checkout, main-ref"
tier: M
era: V3R6
---

# SPEC: SPEC Lint 의 git 눈멂 — 조용한 skip 을 관측 가능하게 만들고 기준 ref 를 해소한다

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-31 | manager-spec | 최초 작성 — 카드 t371 의 사전 측정(A/B 3회)을 근거로 M1(관측 가능화) → M2(ref 해소 체인) → M3(CI 체크아웃) 순서 고정 |
| 0.2.0 | 2026-08-31 | manager-spec | plan-audit iter-1(PASS-WITH-DEBT 0.75) 차단 결함 5건 반영 — error 모양을 **3종**으로 정정하고 모양별 발화 여부 명시(D1), Unreachable 을 **lint 실행당 1건**으로 결정·상한 AC 신설(D2), ref 이름 추적 AC 신설(D4), AC-003 에 non-terminal status 제약 명시(D7), AC-003/006 에 mutation 절차 추가(D8). 범위 추가: 워크플로 trigger paths 확대(REQ-SLGB-009). §1.2 인용 `:1310-1313` → `:1305-1308` 정정(D5a), `DetectDrift` 결과를 §4 잔여 위험으로 승격(D6). AC 11건·대상 파일 6개로 **Tier S → M** 상향 |
| 0.3.0 | 2026-08-31 | manager-spec | iter-1 정정 2건 반영 — D12 를 필수로 승격: `printTable` 의 zero-finding short-circuit(`✓ No findings`, `internal/cli/spec_lint.go:115-118`)이 곧 눈감긴 상태의 출력이라는 사실을 §1.2 서사에 편입하고, 관측 표면을 기본 표 출력으로 못박음(AC-SLGB-001 / 004 에 해당 줄 부재 단언 추가). `cachedMainBranch` 의 cwd 의존은 시그니처 변경이 아니라 기존 `chdirForTest` 선례로 해소됨을 확인 — plan.md M2 단계 0(디렉터리 파라미터 분기) 철회. `--json` / `--sarif` 동작에 기대는 AC 를 금지하는 조항을 §4 에 명시 |
| 0.4.0 | 2026-08-31 | manager-spec | 인용 갱신 + 기준값 재측정 — `origin/develop`(`9328a5242`) 흡수 후 HEAD `35bc0715f` 에서 인용 좌표를 전량 재측정해 `internal/spec/lint.go` 5개 인용을 갱신(`:1287`→`:1306`, `:1305-1308`→`:1324-1327`, `:1316`→`:1335`, `:284-300`→`:296-312`, `plan.md` `:1287-1323`→`:1306-1342`, `acceptance.md` `:1299-1301`→`:1318-1320`). 그 밖의 파일 인용은 불변으로 확인. 린트 기준값을 `0 error / 1096 warning` 으로 재귀속하고 종전 "1098"이 warning 수가 아니라 `wc -l` 이었다는 단위 오류를 정정, `SyncSHASlotFormat` +5 델타와 t382 무영향 반증을 `progress.md` §E.1 iter-3 에 기록. **요구사항 · AC 의미 변경 없음** |

## 1. 문제 — 측정된 형태

CI 잡 `SPEC Lint`(`.github/workflows/spec-lint.yml`)은 `actions/checkout@v7` 를 `with:` 블록 없이 쓴다.
`fetch-depth` 기본값이 1 이므로 체크아웃은 얕고, 로컬 `main` 브랜치도 없다. 그 상태에서
`go run ./cmd/moai spec lint --strict` 가 돌면 lint 규칙 두 개가 **아무것도 보지 못한 채 초록**을 낸다.
초록이 "일치한다"인지 "안 봤다"인지 출력만으로는 구분되지 않는다.

### 1.1 A/B 측정 — 변수 하나씩

측정 트리: develop `b9149857c`(작업 트리 동일, `main` = `48239c7dc`).
측정 명령: `go run ./cmd/moai spec lint --strict`. 원문 출력은 `.moai/reports/t371/` 및 스크래치 클론에 있다.

| run | 조건 | StatusGitConsistency | OwnershipTransitionInvalid | rc |
|---|---|---|---|---|
| (1) | `git clone --depth 1 --branch develop` | 0 | 0 | 0 |
| (2) | (1) + `git fetch --unshallow` | 0 | 1 | 0 |
| (3) | (2) + `git fetch origin main:main` | 18 | 1 | 0 |

두 축은 **서로 독립**이며, 각각이 설명하는 건수가 정확히 갈린다.

- **축 A — 이력 깊이**: 1건(`OwnershipTransitionInvalid`). `OwnershipTransitionRule`
  (`internal/spec/lint_ownership.go`)이 SPEC 파일에 `git log --follow` 를 걸기 때문이다.
- **축 B — `main` ref**: 18건(`StatusGitConsistency`). `internal/spec/lint.go:1306` 의
  `StatusGitConsistencyRule.Check` 가 `getGitImpliedStatus`(`internal/spec/drift.go:300`)를 부르고,
  그 안에서 `git log <branch> --oneline --no-merges --grep=<specID> -50` 를 돌린다. 여기서
  `branch = cachedMainBranch()`(`internal/spec/gitquery_cache.go:88-118`)인데, 이 헬퍼는
  **로컬 브랜치만** `git rev-parse --verify main` 으로 확인하고 실패하면 리터럴 `"master"` 로 떨어진다.

`fetch-depth: 0` 만으로는 18건 중 **0건**이 회복된다. 체크아웃된 ref 하나만 로컬 브랜치가 되고
`main` 은 여전히 없기 때문이다 — run (2)가 그 측정이다.

### 1.2 더 깊은 결함 — 조용한 skip

`cachedMainBranch()` 가 존재하지 않는 `"master"` 를 돌려주면 `git log` 가 실패하고,
`getGitImpliedStatus` 는 error 를 반환하며, `StatusGitConsistencyRule.Check` 는
`internal/spec/lint.go:1324-1327` 에서 `return nil` 한다. **"일치함"과 "관측 못 함"이 같은 출력**이다.
게다가 이 실패는 SPEC 하나가 아니라 전 코퍼스에 동시에 걸린다 — 기준 ref 하나가 없으면 규칙 전체가 눈을 감는다.

**그리고 그 침묵은 침묵으로 끝나지 않는다.** `printTable` 은 finding 이 0건이면
`internal/cli/spec_lint.go:115-118` 에서 짧게 끊고 다음 한 줄을 찍는다(메시지는 `:116`).

```
✓ No findings — all SPEC documents are valid
```

눈이 감긴 CI 체크아웃이 실제로 내보내는 출력이 이것이다. 규칙이 전 코퍼스에 대해 건너뛰어졌는데
출력은 **모든 SPEC 문서가 유효하다고 단언한다**. 이 거짓 전면 무결 선언이 이 SPEC 이 지우려는 대상이며,
동시에 수리의 기제이기도 하다: M1 이 눈감긴 상태에서 Info 를 내면 finding 이 0건이 아니게 되고,
short-circuit 이 발화하지 않으며, 그 한 줄이 사라진다.

따라서 이 SPEC 의 성공 판정은 "Info 가 어딘가에 생겼다"가 아니라
**"사람이 실제로 보는 표 출력에서 `✓ No findings` 줄이 사라졌다"** 이다.
관측 표면을 이렇게 못박지 않으면, Info 를 `--json` 경로에만 내고 표의 거짓 선언은 그대로 두는 구현이
모든 AC 를 통과한다.

형제 규칙은 바로 이 상황을 위해 `OwnershipTransitionUnreachable`(Info, `lint_ownership.go:380-400`)을
정의해 두었다. 그러나 (a) `StatusGitConsistencyRule` 에는 대응물이 없고,
(b) 그 Info 조차 얕은 클론에서는 발화하지 않는다 — 잘린 이력에서 `git log --follow` 는 error 가 아니라
**성공적으로 짧은 결과**를 내기 때문이다. run (1)에서 `OwnershipTransitionUnreachable` 은 0건으로 측정됐다.

### 1.3 지금은 붉어지지 않는다 (그리고 그것이 위험을 없애지는 않는다)

두 규칙 모두 `--strict` 에서 error 로 승격되지 않는다. `StatusGitConsistency` 는 emission 시점에
`Advisory: true`(`internal/spec/lint.go:1335`), 유일한 `OwnershipTransitionInvalid` 대상
SPEC-LSPMCP-001 은 terminal status 라 `applyEraDemotion`(`internal/spec/lint.go:296-312`)이 advisory 로 낮춘다.
run (3)은 19건이 모두 보이는 상태에서 `rc=0` 으로 측정됐다.

**잔여 위험**: 앞으로 grandfather 가 아닌 SPEC 이 `OwnershipTransitionInvalid` 를 맞으면 승격된다 —
그 규칙은 emission 시점(`internal/spec/lint_ownership.go:429`)에 `Advisory` 를 달지 않는다.
즉 이 SPEC 은 "지금 붉게 만들기"가 아니라 "눈을 뜨게 하기"이고, 붉어질 날은 그 다음에 온다.

## 2. 요구사항 (GEARS)

### M1 — 조용한 skip 을 관측 가능하게

- **REQ-SLGB-001** (Ubiquitous): `StatusGitConsistencyRule` 은 SPEC 의 git 함의 상태를 **관측하지 못한 경우**
  Info severity 의 `StatusGitUnreachable` finding 을 내며, 그 finding 은 **lint 실행 1회당 최대 1건**이다.
  관측 못 함과 일치함은 서로 다른 출력이어야 한다.
- **REQ-SLGB-002** (event-driven): When `getGitImpliedStatus` 가 기준 브랜치 ref 를 해소하지 못해 실패하면,
  lint 는 **시도한 후보 ref 이름을 메시지에 담은** `StatusGitUnreachable` 을 낸다.
- **REQ-SLGB-003** (state-driven): While 저장소가 얕은 상태(shallow)이면, `StatusGitConsistencyRule` 은
  §2.1 의 shape ② `no git history found` 와 shape ③ `no classifiable commit within window` 를
  **관측 실패**로 취급하고 `StatusGitUnreachable` 을 낸다.
- **REQ-SLGB-004** (unwanted): 기준 브랜치가 해소되고 저장소가 완전한(non-shallow) 경우,
  lint 는 shape ② 또는 ③ 에 해당할 뿐인 SPEC 에 대해 `StatusGitUnreachable` 을 내지 **않는다**.
- **REQ-SLGB-005** (Ubiquitous): `StatusGitUnreachable` 은 Info severity 이며 `--strict` 의 종료 코드를 바꾸지 않는다.

### M2 — 기준 ref 해소 체인

- **REQ-SLGB-006** (Ubiquitous): `cachedMainBranch()` 는 기준 브랜치를 순서 있는 체인
  로컬 `main` → `origin/main` → 로컬 `master` → `origin/master` 로 해소하며,
  넷 다 없으면 **해소 불가**를 보고한다. 존재하지 않는 리터럴을 반환하지 않는다.
- **REQ-SLGB-007** (state-driven): While `Lint()` 의 per-run git-query 캐시가 활성이면,
  기준 브랜치 해소는 한 번의 `Lint()` 실행당 최대 1회만 계산되며, **해소 불가 역시 캐시된다**.

### M3 — CI 워크플로

- **REQ-SLGB-008** (Ubiquitous): `SPEC Lint` 워크플로는 lint 실행 전에 완전한 이력을 체크아웃하고,
  `main` ref 를 **명시적으로** fetch 한다.
- **REQ-SLGB-009** (Ubiquitous): `SPEC Lint` 워크플로의 trigger `paths` 는 lint 규칙이 사는
  Go 소스와 워크플로 파일 자체의 변경에도 발화한다. SPEC 문서만 감시하는 현재 필터로는
  이 SPEC 의 M1/M2 변경이 잡을 재실행시키지 못하고, AC-SLGB-008 이 영구히 닫히지 않는다.

### 2.1 설계 결정 ① — error 모양은 셋이다 (둘이 아니다)

`getGitImpliedStatus` 는 error 를 **세 모양**으로 낸다. 모양별 발화 여부는 다음과 같다.

| # | 반환 지점 | 문자열 | 의미 | 완전한 저장소 | 얕은 저장소 |
|---|---|---|---|---|---|
| ① | `drift.go:312` | `git log failed: …` | 기준 ref 해소 실패 / git 사용 불가 | **발화** | **발화** |
| ② | `drift.go:316` | `no git history found for <ID>` | git 은 돌았고 `--grep` 매치 0 | 침묵 | **발화** |
| ③ | `drift.go:366` | `no classifiable commit within window of 50 for <ID>` | 매치는 있으나 창 안에 분류 가능 커밋 없음 | 침묵 | **발화** |

- ① 은 저장소 상태와 무관하게 관측 실패다. 무조건 발화한다.
- ② 와 ③ 은 **완전한 저장소에서는 정상적 무해**다. ② 는 그 SPEC 에 lifecycle 커밋이 없을 뿐이고,
  ③ 은 커밋이 전부 cosmetic 일 뿐이다(`.moai/reports/t371/classification-18.md` 의 C-2 형태가 바로 이것이다).
  얕은 저장소에서는 둘 다 **잘린 창이 만든 산물일 수 있어** 관측 실패로 취급한다.
- 두 경우를 가르는 술어는 저장소 수준의 `git rev-parse --is-shallow-repository` 이며,
  `cachedMainBranch` 와 같은 per-run 캐시에 얹어 SPEC 수 만큼의 subprocess 를 만들지 않는다.

이 결정으로 M1 은 축 A(이력 깊이)와 축 B(ref)를 **둘 다** 덮는다. 남는 사각은 §4 에 있다.

### 2.2 설계 결정 ② — 발화는 실행당 1건 (SPEC 당 아님)

세 모양의 원인은 모두 **저장소 수준**이다: ① 은 ref 부재, ② ③ 의 발화 조건은 shallow 상태다.
그런데 `Check` 는 SPEC 마다 불린다. 조건을 그대로 SPEC 마다 발화시키면
이 카드가 고치려는 바로 그 CI 상태에서 동일한 Info 가 수백 줄 찍히고,
진짜 신호가 소음에 묻힌다 — plan.md §G 가 금지하는 형태다.

따라서 `StatusGitUnreachable` 은 **`Lint()` 실행 1회당 최대 1건**이다.
중복 억제는 per-run 캐시(`gitEnvCache`)에 발화 여부 플래그를 두어 수행하며,
메시지는 이 조건이 **저장소 전체에 걸린다**는 사실과 그로 인해 이 규칙이 전 SPEC 에 대해
건너뛰어졌다는 사실을 명시한다. 캐시가 비활성인 경로(`Lint()` 밖의 직접 호출)에서는
억제할 상태가 없으므로 그대로 발화한다.

### 2.3 설계 결정 ③ — M3 의 명시적 fetch 가 중복이 아닌 이유

`actions/checkout` 이 `fetch-depth: 0` 에서 `refs/remotes/origin/*` 를 채우는지는
이번 조사에서 **검증되지 않았다** — GitHub Actions 내부 동작은 로컬에서 재현할 수 없다.
검증되지 않은 전제 위에 수리를 얹지 않기 위해, 명시적 `main` fetch 단계를 별도로 둔다.
설령 중복이더라도 비용은 fetch 한 번이고, 빠졌을 때의 비용은 CI 가 다시 눈을 감는 것이다.

## 3. 인수 기준

Tier M — AC 는 `acceptance.md` 에 있다(AC-SLGB-001 … AC-SLGB-011).
모든 AC 는 명령과 관측 대상을 명시하며 주관적 판단을 담지 않는다.

## 4. 잔여 위험

- **얕은 저장소에서 워커가 잘못된 커밋을 무는 경우**: 잘린 창 안에 분류 가능한 커밋이 **하나라도**
  남아 있으면 `getGitImpliedStatus` 는 error 없이 값을 낸다(세 shape 중 어디에도 걸리지 않는다).
  이 경우는 `StatusGitUnreachable` 로 잡히지 않고 기존과 동일하게 `StatusGitConsistency` 오탐으로 나타난다.
  M1 은 이것을 덮지 않으며, 이를 덮으려면 워커가 창 소진 여부를 값과 함께 보고해야 한다 — 후속 카드 후보.
- **`DetectDrift` 의 동작이 M2 로 바뀐다**: `internal/spec/drift.go:68` 이
  `branch: cachedMainBranch` 로 배선돼 있어, 이 헬퍼는 lint 전용이 아니다.
  `origin/main` 은 있고 로컬 `main` 이 없는 체크아웃에서 `DetectDrift` 는 오늘 존재하지 않는 `"master"` 를
  걸어 **drift 0건을 조용히 기록**한다. M2 이후에는 실제 record 가 나타난다.
  이는 SessionStart 경로의 관측 가능한 동작 변화이며, **현재 어떤 AC 도 이 경로를 덮지 않는다**.
  M2 의 회귀 표면이 lint 하나가 아니라는 사실을 run-phase 가 알고 들어가야 한다.
- **`OwnershipTransitionRule` 자체의 잘린-이력 사각**: `git log --follow` 가 잘린 이력에서 성공하므로
  `OwnershipTransitionUnreachable` 은 여전히 발화하지 않는다. 이 SPEC 은 이 사각을 후속 카드 후보로만 기록한다.
- **`OwnershipTransitionInvalid` 의 미래 승격**: §1.3 참조. 현재 코퍼스에서는 rc=0 이지만
  emission 시점에 `Advisory` 가 없으므로 grandfather 가 아닌 대상이 생기면 `--strict` 를 붉게 만든다.
- **요약 줄은 Info 를 세지 않는다**: `printTable`(`internal/cli/spec_lint.go:113-133`)은 severity 필터 없이
  모든 finding 을 표로 찍지만, 그 아래 요약 줄(`:136-142`)은 error 와 warning 만 센다.
  즉 `StatusGitUnreachable` 행은 **표에 보이되 집계 숫자는 그대로**다.
  M1 의 목적은 이 비대칭에도 살아남는다 — 목적이 "숫자를 바꾸는 것"이 아니라
  "출력에 흔적을 남기는 것"이기 때문이다(REQ-SLGB-005 는 rc 불변을 오히려 요구한다).
  어떤 AC 도 요약 줄에 기대지 않는다: AC 는 전부 finding 의 code/severity/개수와
  `✓ No findings` 줄의 유무를 직접 읽는다.
- **`--json` / `--sarif` 출력 경로는 검증된 바 없다**: `Finding` 구조체가 JSON 태그를 달고 있다는 사실
  (`internal/spec/lint.go:33-45`)은 직렬화 **의도**를 보일 뿐, 그 경로의 실제 동작을 보이지 않는다.
  이 SPEC 은 어떤 AC 도 `--json` 이나 `--sarif` 동작 위에 세우지 않는다.
  관측 표면은 기본 표 출력(`printTable`) 하나로 고정한다 — §1.2 참조.
- **`cachedMainBranch` 는 프로세스 작업 디렉터리를 따른다**: `exec.Command` 에 `.Dir` 을 설정하지 않는다
  (`internal/spec/gitquery_cache.go:102`, `:113`). 이것은 시그니처 변경을 요구하지 않는다 —
  같은 패키지의 기존 픽스처 테스트가 이미 `chdirForTest`
  (`internal/spec/drift_characterization_test.go:55`)로 이 성질을 다루고 있고,
  `setupDriftCorpusFixture`(`:98`)가 마지막에 그것을 호출한다(`:103`).
  새 AC 는 그 선례를 그대로 쓴다(`acceptance.md` 공통 픽스처 규약).

## 5. 범위에서 제외 (Exclusions)

### Out of Scope — SPEC 코퍼스 수리

측정된 18건의 `StatusGitConsistency` finding 은 전수 분류되어 있다
(`.moai/reports/t371/classification-18.md`, 근거 `.moai/reports/t371/statusgit-18-walker-input.txt`).

- **A — 진짜 frontmatter 드리프트, 2건**: SPEC-KANBAN-TODO-CLI-001, SPEC-UPDATE-DOC-DRIFT-001.
- **B — close 상태 불일치, 2건**: SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001.
- **C — 워커 휴리스틱 산물, 14건**: C-1 sync 커밋 주제가 `sync-phase` 리터럴을 안 담음 5건,
  C-2 cosmetic docs/chore 커밋이 최신 슬롯을 차지 6건, C-3 SPEC-ID 를 담은 lifecycle 커밋 자체가 없음 3건.

A 와 B 를 고치는 일은 `.moai/specs/` 를 편집하는 일이라 다른 레인의 파일과 충돌한다. C 는 커밋 규약 축이지
이 카드의 축(워커가 눈을 감고 있다)이 아니다. 셋 다 후속 카드 후보로만 남긴다.

### Out of Scope — 다른 워크플로와 설정

- `.github/workflows/ci.yml`: 이 트리에서 이미 6곳(`:129 :264 :382 :431 :486 :543`)에 `fetch-depth: 0` 을 담고 있다.
- `spec-lint.yml` 의 `concurrency` / `cancel-in-progress` 설정: 운영자 판정으로 그대로 둔다.

### Out of Scope — 구현 세부

- lint 출력 포매터의 Info 렌더 방식 변경, 요약 줄의 집계 대상 변경, severity 체계 재설계,
  `Advisory` 플래그 정책 변경.
- `gitLogWindowSize`(현재 50) 조정, 워커 휴리스틱(`ClassifyPRTitle` 접두 표) 확장.
- `DetectDrift` 의 `logAll` 실패 자체를 관측 가능하게 만드는 일(§4 참조) — M2 가 원인 하나를 없앨 뿐,
  그 경로의 침묵 자체는 이 SPEC 이 건드리지 않는다.
