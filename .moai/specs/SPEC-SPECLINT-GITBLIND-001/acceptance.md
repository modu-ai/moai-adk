# Acceptance: SPEC-SPECLINT-GITBLIND-001

Tier M. 11개 AC 전부 명령과 관측 대상을 명시한다. 주관적 판단어("적절히", "잘")를 쓰지 않는다.

**공통 픽스처 규약** — AC-SLGB-001..007 은 `t.TempDir()` 안에 `git init` 으로 만든 저장소를 쓴다.
`internal/spec/` 에 선례가 있다(`drift_chore_skip_test.go`, `archive_git_test.go`, `closer_test.go`).
두 가지 구속이 모든 픽스처 테스트에 걸린다:

- **비병렬 [HARD]** — per-run git-query 캐시는 패키지 전역(`gitQueryCacheMu` / `gitQueryCacheV`,
  `internal/spec/gitquery_cache.go:23-25`)이다. 캐시를 건드리는 테스트는 `t.Parallel()` 을 쓰지 않는다.
  이 구속은 이 SPEC 이 새로 만드는 규칙이 아니다 — 기존 헬퍼의 주석
  (`internal/spec/drift_characterization_test.go:53`)이 이미 같은 문장을 담고 있다.
- **작업 디렉터리 — 기존 선례를 그대로 쓴다** — `cachedMainBranch` 는 `exec.Command` 에 `.Dir` 을
  설정하지 않으므로 프로세스 작업 디렉터리를 따른다(`:121`, `:132`). 이것은 **시그니처 변경을 요구하지 않는다**:
  같은 패키지의 `chdirForTest`(`drift_characterization_test.go:55` — `os.Chdir` + `t.Cleanup` 복원)가
  이미 이 성질을 다루고, `setupDriftCorpusFixture`(`:98`)가 마지막에 그것을 호출한다(`:103`).
  비-git 디렉터리로 진입하는 선례도 있다(`:280`). 새 픽스처는 이 헬퍼를 재사용한다.
- **관측 표면은 기본 표 출력 하나 [HARD]** — 판정은 `printTable`(`internal/cli/spec_lint.go:113`)이
  내보내는 기본 출력에서 읽는다. `--json` 과 `--sarif` 경로의 동작은 이 조사에서 **검증된 바 없으므로**
  (`Finding` 의 JSON 태그 `internal/spec/lint.go:33-45` 는 의도만 보인다),
  어떤 AC 도 그 위에 서지 않는다.

---

## M1 — 조용한 skip 의 관측 가능화

### AC-SLGB-001 — 기준 ref 가 없으면 Info 가 뜬다 (M1 의 핵심 판정)

- **Given** 로컬 `main` / `origin/main` / `master` / `origin/master` 중 **어느 것도 없는** git 픽스처와,
  frontmatter `status` 가 `terminalStatusEnum`(`superseded` / `archived` / `rejected` / `completed`)에
  속하지 **않는** SPEC 문서 1개
- **When** 그 저장소에서 `moai spec lint` 를 실행하고 **기본 표 출력**(`printTable`)을 읽으면
- **Then** 다음 둘이 **모두** 참이다:
  1. 출력에 `StatusGitUnreachable` code 를 가진 finding 이 **1건 이상** 나타난다.
  2. 출력에 `✓ No findings — all SPEC documents are valid` 줄이 **없다**.
- **[HARD] 픽스처 전제 (RED 기준선)**: 픽스처의 SPEC 문서는 다른 lint finding 을 내지 않는
  schema-valid 문서여야 한다. 그래야 **M1 이전**에는 finding 이 0건이라
  `printTable` 이 short-circuit(`internal/cli/spec_lint.go:115-118`)하여 위 2번의 줄을 실제로 찍는다.
  구현 전 그 줄이 찍히는 것을 관측한 뒤 구현한다 — 이것이 이 AC 의 RED 다.
  다른 finding 이 섞여 있으면 short-circuit 이 애초에 발화하지 않아 2번 단언이 공허해진다.
- **[HARD] `fixtureSpecMD` 선례는 전면만 따른다 (iter-3 D1)**: `fixtureSpecMD`
  (`drift_characterization_test.go:109`)의 본문은 절이 아예 없어(`# <ID>` 한 줄) 그대로 쓰면
  `MissingExclusions` 를 내어 위 전제가 깨진다. 픽스처 본문은 `## 4. Scope` 절 안에
  `### 4.1 Out of Scope — <suffix>` 소제목과 항목 1개를 **반드시** 포함한다. 검증된 형태:
  `.moai/reports/t371/repro/withscope/spec.md` (이 전제로 `✓ No findings` rc=0 실측 — 아래 2-cell 대장 AC-001 행).

이 AC 가 실패하면 M1 은 미달이다 — 기준 ref 가 없는 체크아웃에서 Info 가 나오지 않는 것이 곧 결함이다.
2번 단언이 없으면 Info 를 `--json` 경로에만 내고 표의 거짓 전면 무결 선언은 그대로 두는 구현이 통과한다
(`spec.md` §1.2).

### AC-SLGB-002 — 메시지가 시도한 ref 이름을 담는다 (REQ-SLGB-002 추적)

- **Given** AC-SLGB-001 과 동일한 픽스처
- **When** lint 를 실행하고 `StatusGitUnreachable` finding 의 `Message` 를 읽으면
- **Then** 메시지는 해소를 시도한 후보 ref 이름을 담는다 — 최소한 문자열 `main` 과 `master` 가
  모두 나타나며, 조건이 저장소 전체에 걸린다는 사실이 메시지에 명시된다.

이 AC 가 없으면 구현이 ref 이름을 생략해도 나머지 AC 가 전부 통과한다.

### AC-SLGB-003 — 발화는 lint 실행당 정확히 1건 (REQ-SLGB-001 상한)

- **Given** AC-SLGB-001 의 픽스처에 비-terminal status 를 가진 SPEC 문서를 **10개** 둔 상태
- **When** `moai spec lint` 를 1회 실행하면
- **Then** `StatusGitUnreachable` finding 의 개수는 **정확히 1** 이다(10 이 아니다).

### AC-SLGB-004 — 얕은 저장소에서 shape ② 와 ③ 이 발화한다

`spec.md` §2.1 의 shape 표를 두 하위 픽스처로 나눠 검증한다. 두 픽스처 모두
`git rev-parse --is-shallow-repository` 가 `true` 이고 기준 브랜치는 해소된다.

- **AC-SLGB-004a (shape ②)** — Given 얕은 창 안에 해당 SPEC-ID 를 담은 커밋이 **하나도 없는** 픽스처,
  When lint 를 실행하고 기본 표 출력을 읽으면, Then `StatusGitUnreachable` 이 나타나고
  `✓ No findings — all SPEC documents are valid` 줄이 **없다**.
- **AC-SLGB-004b (shape ③)** — Given 얕은 창 안에 해당 SPEC-ID 를 담은 커밋은 있으나
  전부 분류 불가(`chore(spec): sweep` 형태)인 픽스처, When lint 를 실행하고 기본 표 출력을 읽으면,
  Then `StatusGitUnreachable` 이 나타나고 `✓ No findings` 줄이 **없다**.
- **[HARD] 픽스처 전제 (RED 기준선)**: AC-SLGB-001 과 동일 — 두 픽스처의 SPEC 문서는 다른 finding 을
  내지 않는 schema-valid 문서여야 하며, 구현 전 `✓ No findings` 줄이 실제로 찍히는 것을 관측한다.
  `fixtureSpecMD` 본문 무단 복사 금지와 Out of Scope 절 의무를 포함해 AC-SLGB-001 의 [HARD] 전제 전체가
  이 픽스처들에도 그대로 적용된다(iter-3 D1).

004b 가 iter-1 D1 이 지적한 사각이다 — `drift.go:366` 의 창 소진 경로.
`✓ No findings` 부재 단언을 두 하위 픽스처에 모두 거는 이유는 §1.2 와 같다: 이 줄이 곧 눈감긴 상태의
출력이므로, 그것이 사라지는 것이 M1 의 실제 수리다.

### AC-SLGB-005 — 완전한 저장소에서는 침묵한다 (REQ-SLGB-004, 소음 방지 반대 방향 가드)

- **Given** `git rev-parse --is-shallow-repository` 가 `false` 이고 로컬 `main` 이 해소되는 픽스처와,
  frontmatter `status` 가 **`draft`**(비-terminal 임을 명시)인 SPEC 문서 — lifecycle 커밋은 없다
- **When** lint 를 실행하면
- **Then** `StatusGitUnreachable` finding 은 **0건**이다.
- **[HARD] 제약**: 픽스처의 `status` 는 반드시 비-terminal 이어야 한다.
  `Check` 는 `internal/spec/lint.go:1367-1369` 에서 `terminalStatusEnum` 인 문서에 대해
  git 에 닿기 **전에** `return nil` 하므로, `completed` 픽스처는 규칙이 한 번도 돌지 않은 채
  "0건"을 만족시킨다 — 즉 공허하게 통과한다.
- **[HARD] mutation 확인**: 구현에서 shallow 술어 가드를 제거해
  shape ②/③ 을 무조건 발화시키도록 바꾸면 이 AC 는 **반드시 실패**해야 한다
  (기대: 0건 → 1건). 실패하지 않으면 픽스처가 규칙에 닿지 않은 것이다.

### AC-SLGB-006 — Info 는 rc 를 바꾸지 않는다 (REQ-SLGB-005)

- **Given** AC-SLGB-001 의 픽스처
- **When** `moai spec lint --strict` 를 실행하면
- **Then** 종료 코드는 `0` 이고, `StatusGitUnreachable` finding 의 severity 필드 값은 `info` 다.

---

## M2 — 기준 ref 해소 체인

### AC-SLGB-007 — 체인이 순서대로 해소한다 (REQ-SLGB-006)

- **Given** 네 가지 픽스처 — (a) 로컬 `main` 만 존재, (b) `origin/main` 만 존재,
  (c) 로컬 `master` 만 존재, (d) 넷 다 없음
- **When** 각 픽스처에서 `cachedMainBranch` 를 호출하면
- **Then** 반환값은 차례로 `main` / `origin/main` / `master` / **해소 불가 신호**이며,
  (d) 에서 리터럴 문자열 `"master"` 는 반환되지 **않는다**.
- 픽스처 진입은 공통 규약의 `t.Chdir` + 비병렬 구속을 따른다.

### AC-SLGB-008 — 해소 결과가 실행당 1회만 계산된다 (REQ-SLGB-007)

- **Given** per-run 캐시가 활성인 상태(`startGitQueryCache()` 호출 후)에서
  `cachedMainBranch` 를 1회 호출한 뒤, 해당 브랜치 ref 를 삭제한 상태
- **When** 같은 실행 안에서 다시 호출하면
- **Then** 첫 호출과 **동일한 값**이 반환된다.
- **[HARD] mutation 확인**: `gitEnvCache.mainBranchSet` 조기 반환
  (`internal/spec/gitquery_cache.go:115-119`)을 제거하면 이 AC 는 **반드시 실패**해야 한다
  (두 번째 호출이 삭제된 ref 를 다시 조회해 다른 값을 낸다). 실패하지 않으면
  이 AC 는 오늘 트리에서 이미 통과하는 상속된 초록이며 아무것도 검증하지 않는다.
- **[HARD] 해소 불가 캐싱**: (d) 픽스처에서도 두 번째 호출이 재조회하지 않아야 한다 —
  해소 불가는 SPEC 마다 4회 `git rev-parse` 를 spawn 하지 않는다.

---

## M3 — CI 워크플로

### AC-SLGB-009 — 체크아웃이 완전한 이력과 `main` ref 를 갖춘다 (REQ-SLGB-008)

- **When** `.github/workflows/spec-lint.yml` 을 읽으면
- **Then** 다음 셋이 모두 참이다:
  1. `actions/checkout@v7` 단계가 `fetch-depth: 0` 을 담은 `with:` 블록을 가진다.
  2. `main` ref 를 fetch 하는 단계가 존재한다.
  3. 그 fetch 단계는 `go run ./cmd/moai spec lint` 실행 단계보다 **앞선다**.

### AC-SLGB-010 — trigger paths 가 규칙 소스와 워크플로 자체를 덮는다 (REQ-SLGB-009)

- **When** `.github/workflows/spec-lint.yml` 의 `on.pull_request.paths` 와 `on.push.paths` 를 읽으면
- **Then** 두 목록 모두 (a) `.moai/specs/**`, (b) lint 규칙이 사는 Go 소스
  (최소한 `internal/spec/**`), (c) 이 워크플로 파일 자체를 덮는 패턴을 담는다.
- 근거: 현재 두 트리거 모두 `paths: ['.moai/specs/**']` 단독이라,
  이 SPEC 의 M1/M2(전부 `internal/spec/` Go 변경)가 착지해도 잡이 재실행되지 않는다.
  그러면 AC-SLGB-011 이 영구히 닫히지 않는다.

### AC-SLGB-011 — 규칙이 실제로 돌았음을 CI 로그로 확인한다

- **When** 이 SPEC 착지 후 `SPEC Lint` 잡이 실제로 완주한 run 의 로그를 읽으면
- **Then** `StatusGitConsistency` finding 이 **1건 이상** 나타난다.
- 0건이면 규칙이 여전히 안 돌고 있다는 뜻이다 — "초록"과 "돌았음"을 가르는 유일한 관측이다.
- **[HARD] 판정 장소**: 이 AC 는 **로컬에서 판정 불가**하며 CI 로그로만 닫힌다.
  run-phase 에서 PASS 로 표시하지 않는다. 착지 후 조용한 head 의 잡 로그를 읽어 닫는다.

---

## 커버리지 표 (REQ → AC)

| REQ | 닫는 AC |
|---|---|
| REQ-SLGB-001 (발화 + 실행당 1건) | AC-SLGB-001, AC-SLGB-003 |
| REQ-SLGB-002 (ref 이름) | AC-SLGB-002 |
| REQ-SLGB-003 (shallow → shape ②③) | AC-SLGB-004a, AC-SLGB-004b |
| REQ-SLGB-004 (완전한 저장소 침묵) | AC-SLGB-005 |
| REQ-SLGB-005 (Info, rc 불변) | AC-SLGB-006 |
| REQ-SLGB-006 (해소 체인) | AC-SLGB-007 |
| REQ-SLGB-007 (실행당 1회 계산) | AC-SLGB-008 |
| REQ-SLGB-008 (체크아웃) | AC-SLGB-009 |
| REQ-SLGB-009 (trigger paths) | AC-SLGB-010 |
| — (반-공허 관측) | AC-SLGB-011 |

`spec.md` §4 의 `DetectDrift` 동작 변화는 **어떤 AC 도 덮지 않는다**. 의도된 공백이며
그 사실 자체가 §4 에 기록돼 있다 — run-phase 는 그 경로를 미검증 잔여 위험으로 안고 들어간다.

## 2-Cell 채택 대장 (verification-completeness §2)

RED-now 셀과 green-path 셀을 쌍으로 기록한다. 관측 트리: **`d31bf744d`**(구현 전, 2026-09-02 측정).
분류 3종 — **실측** = 이 트리에서 RED 관측을 실행해 원문 출력·exit code 를 확보 /
**seam** = 시나리오 구성에 M2 테스트 seam 이 필요(판정서 순환 해소 [HARD] — plan 시점으로 끌어올리지
않고 run-phase §E.2 에서 관측) / **CI** = 로컬 판정 불가(§2.1 undecidable → regression-guard).

| AC | RED-now (관측 트리 `d31bf744d`) | 왜 red 인가 | green-path | 분류 |
|---|---|---|---|---|
| AC-SLGB-001 | 실행 관측: `go run ./cmd/moai spec lint .moai/reports/t371/repro/withscope/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0. main-부재 시나리오 자체는 미구성 — run-phase M1 에서 `chdirForTest` 시드로 관측해 §E.2 에 기록 | `StatusGitUnreachable` 가 아직 존재하지 않아 Info 없이 눈감긴 전면 무결이 찍힌다 | M1 — 같은 시나리오에서 `StatusGitUnreachable` ≥1건 + `✓ No findings` 줄 부재 | seam |
| AC-SLGB-002 | seam (finding 존재가 전제 — 발화 이전엔 Message 를 읽을 대상이 없다) | 대상 finding 이 없다 | M1 — 메시지가 `main`·`master` 후보명과 저장소 전체 조건을 명시 | seam |
| AC-SLGB-003 | seam (10문서 픽스처 실행은 새 테스트가 필요) | 대상 finding 이 없다 | M1 — 발화 플래그로 실행당 1건 상한 | seam |
| AC-SLGB-004a | seam (shallow 픽스처가 프로세스 cwd 여야 한다 — 워크트리 격리 가드로 plan 시점 구성 불가) | shallow+빈창에서 조용한 skip 이 계속된다 | M1 — shallow 술어가 shape ② 를 발화 | seam |
| AC-SLGB-004b | seam (AC-SLGB-004a 와 동일 구속) | 창 소진(`drift.go:366`)이 침묵으로 흡수된다 | M1 — shape ③ 도 발화 | seam |
| AC-SLGB-005 | 구현 전 트리에서 이 AC 는 **공허 초록**이다(대상 finding 이 없어 0건이 자동 참) — RED-now 가 원리상 불가하므로 §2 뮤탄트 조항으로 대체: shallow 술어 제거 mutant 가 반드시 실패해야 한다(DoD 명시) | — (red 없음 — 음의 제약) | M1 — 0건이 자동 참이 아니라 술어가 보장하는 침묵이 된다 | seam(뮤탄트 대체) |
| AC-SLGB-006 | seam (`--strict` 에서 새 finding 의 severity 필드 관측은 발화 이후에만 가능) | 대상 finding 이 없다 | M1 — Info 는 rc 를 움직이지 않는다(`HasErrors` 불변) | seam |
| AC-SLGB-007 | seam (4픽스처 체인은 M2 의 함수 대상) | 해소 체인이 아직 `"main"`/`"master"` 2단 | M2 — 4단 체인 + 해소 불가 신호 | seam |
| AC-SLGB-008 | seam (캐시 mutation 시나리오는 새 테스트 필요) | 조기 반환 블록이 있어도 그것을 때리는 관측이 없다 | M2 — mutation 확인 실패 출력이 §E.2 에 기록 | seam |
| AC-SLGB-009 | **실측**: `grep -n 'fetch-depth' .github/workflows/spec-lint.yml` → stdout 없음, **exit 1** | checkout 단계(`:31`)에 `with:` 블록 자체가 없고 main fetch 단계도 없다 — Then 3종 전부 불충 | M3 — checkout 에 `with: fetch-depth: 0` + main fetch 단계가 lint 실행보다 선행; 같은 grep 이 rc=0 으로 행을 낸다 | 실측 |
| AC-SLGB-010 | **실측**: `grep -n -A6 'paths:' .github/workflows/spec-lint.yml` → 두 목록이 각각 `.moai/specs/**` 단독, **exit 0** | (b) `internal/spec/**` (c) 워크플로 자체 커버 부재 — M1/M2 착지해도 잡이 재실행되지 않는다 | M3 — paths 에 (a)(b)(c) 3종 확장 | 실측 |
| AC-SLGB-011 | 본문 [HARD] 대로 로컬 판정 불가 — CI 잡 로그만이 관측면 | §2.1 undecidable → **regression-guard**, release-blocking 아님, run-phase 에서 PASS 표시하지 않음 | sync 이후 조용한 head 의 `SPEC Lint` 잡 로그에서 `StatusGitConsistency` ≥1 관측으로 닫는다 | CI |

## Definition of Done

- AC-SLGB-001 … AC-SLGB-010 전부 통과, 각 테스트 함수를 `go test -run <name> -v ./internal/spec/...`
  로 직접 지목해 **실행 사실**을 확인한다(셀렉터 0건 매치 통과 방지).
- AC-SLGB-005 와 AC-SLGB-008 의 mutation 확인을 각각 수행하고, 심은 mutant 와
  그것이 만든 실패 출력을 progress.md §E.2 에 남긴 뒤 복원한다.
- `go test ./internal/spec/...` rc 0. **`go test ./...` 는 실행하지 않는다.**
- AC-SLGB-011 은 미해결로 명시한 채 run-phase 를 닫고, sync 이후 CI 로그로 닫는다.
