# ACCEPTANCE: SPEC-AGENT-EMIT-LINEAGE-001

> Tier M 산출물. v0.3.0 까지 이 내용은 `spec.md §3` 에 인라인이었다(Tier S). v0.4.0 이 영향 파일 전수 열거로 Tier 를 **M** 으로 재판정하면서 여기로 분리했다 — 근거와 산술은 `plan.md §B`. **수락 기준의 개수·문면은 분리로 바뀌지 않았고**, v0.4.0 이 더한 것은 AC-AEL-003 안의 판정 단계 4개뿐이다(감사 iter-2 D7 의 판정 공백). 총 **7** 건, 001-007.

각 항목은 판정 명령을 함께 적는다. 측정 트리는 `.claude/worktrees/t317` @ `48eb945df` 기준이며, run-phase 는 자신이 잰 트리 SHA 를 함께 인용한다.

---

## AC-AEL-001 *(REQ-AEL-001)* — 소스 층 드리프트가 로컬 build 에서 즉시 빨간불이 된다 **[뮤테이션 확립]**

- **Given** 깨끗한 트리(`go test ./internal/template/agentemit/... -count=1` → `ok`)
- **When** 커밋된 방출물 한 개에 뮤턴트를 심고 `make build` 를 실행한다
  ```bash
  printf '\n# mutant-t317-a\n' >> internal/template/templates/.codex/agents/moai/manager-git.toml
  make build; echo exit=$?
  ```
- **Then** exit ≠ 0 이고, 출력이 `.codex/agents/moai/manager-git.toml` 을 이름으로 지목하며, 바이너리 컴파일 단계에 도달하지 않는다
- **Restore** `git checkout -- internal/template/templates/.codex/agents/moai/manager-git.toml` → `make build` exit=0

> 이 AC 가 뮤테이션 확립이어야 하는 이유: 실측 8 이 **바로 이 패키지에서** 그럴듯한 가드가 공허했음을 보였다. 실패시킬 수 있음을 보이지 못한 가드는 가드가 아니다.

## AC-AEL-002 *(REQ-AEL-002, REQ-AEL-003)* — build 의 검사는 쓰지 않는다 (재생성 금지) **[뮤테이션 확립]**

- **Given** AC-AEL-001 의 뮤턴트가 심긴 상태에서 `make build` 가 실패한 직후
- **When** 뮤턴트 마커의 생존과 트리 변경 범위를 잰다
  ```bash
  grep -c 'mutant-t317-a' internal/template/templates/.codex/agents/moai/manager-git.toml   # 1 이어야 한다
  git status --short
  ```
- **Then** 마커 카운트가 `1` — 즉 검사가 방출물을 재생성해 손편집을 덮어쓰지 **않았다**. `git status --short` 에 이 뮤턴트 파일 외의 신규·수정 경로가 검사 때문에 추가되지 않는다
- **Restore** 위와 동일

> 재생성은 증거를 지우고, 검증은 증거를 세운다. 손편집을 build 가 조용히 덮으면 그 손편집이 있었다는 사실 자체가 사라진다.

## AC-AEL-003 *(REQ-AEL-004, REQ-AEL-005)* — 임베드 축 판정 지점이 스테일 바이너리를 잡는다 **[뮤테이션 확립 — 실측 8 이 죽이지 못한 바로 그 뮤턴트]**

- **Given** 깨끗한 트리에서 `make build` 가 성공해 `bin/moai` 가 존재한다. 이하 모든 단계에서 `REPO=$(pwd)` 를 잡아 두고, 판정 대상 바이너리는 **이 트리에서 방금 빌드한 `$REPO/bin/moai`** 로 고정한다(설치본이 아니다)
- **When** 커밋된 방출물 한 개를 바꾸고 **재빌드 없이** 임베드 축 검사를 돌린다
  ```bash
  printf '\n# mutant-t317-b\n' >> internal/template/templates/.codex/agents/moai/manager-git.toml
  make embed-check; echo exit=$?          # 판정 명령 (타깃명은 run-phase 확정)
  ```
- **Then** exit ≠ 0 이고 출력이 `manager-git.toml` 을 이름으로 지목한다
- **And (기수 — 게이트)** 같은 출력이 **비교한 경로 수**를 보고하고, 그 수가 커밋된 방출물 개수와 같다
  ```bash
  ls internal/template/templates/.codex/agents/moai/*.toml | wc -l   # 이 트리 이 실행: 11
  ```
  일부 경로만 비교하고도 통과하는 상태는 실패로 판정한다 — 뮤턴트가 우연히 비교 대상에 들어간 것만으로는 이 AC 를 만족시키지 못한다
- **And (바이너리 부재 — 게이트)** 판정 대상이 없을 때 **성공이 아니라 실패**로 끝난다
  ```bash
  BIN=/nonexistent/moai make embed-check; echo exit=$?   # exit ≠ 0 이어야 한다. "비교 0건 → exit 0" 은 실패다
  ```
  이 게이트는 **적용 가능한** 트리(= 커밋 산출물을 이고 있는 이 저장소)에서만 의미를 갖는다. 적용 불가 트리의 거동은 아래 별도 게이트가 판정한다
- **And (verb 도달 — 게이트, v0.4.0)** 판정 지점이 **명시적 유지자 동사**로 닿는다. 뮤턴트를 원복한 뒤:
  ```bash
  grep -n '^embed-check:' Makefile     # 타깃 정의 1행 (타깃명은 run-phase 확정)
  make embed-check; echo exit=$?       # exit=0
  ```
  → 결정하는 조항: REQ-AEL-004 의 "reachable … as an explicit maintainer verb"
- **And (doctor 도달 — 게이트, v0.4.0) [뮤테이션 확립]** 같은 판정이 **`moai doctor` 항목**으로도 닿고, 항목명이 `--check` 로 개별 호출된다. 위 뮤턴트가 심긴 상태(바이너리는 뮤턴트 이전에 빌드됨)에서:
  ```bash
  "$REPO/bin/moai" doctor --check "Agent Emit Embed"; echo exit=$?   # exit ≠ 0, 해당 항목 상태 fail
  ```
  원복 후 같은 명령이 exit=0 + 항목 상태 ok 를 낸다. 필터가 실제로 **한 항목만** 남겼음은 출력 말미 카운터의 합이 **1** 인 것으로 판정한다. 이 트리 이 실행에서 선례 항목으로 그 형태를 확인했다:
  ```console
  $ moai doctor --check "MCP Server Version" 2>&1 | tail -3
  │    1 ok, 0 warn, 0 fail                                             │
  │   Pass 1    Warn 0    Fail 0                                        │
  ```
  → 결정하는 조항: REQ-AEL-004 의 "reachable … as a `moai doctor` check item"
- **And (CI 빌드 잡 미부착 — 금지 게이트, v0.4.0)** 판정 지점이 CI 빌드 잡의 자동 트리거로 붙지 않았다
  ```bash
  grep -rn 'embed-check' .github/workflows/; echo grep_rc=$?   # rc=1 (0 히트)
  ```
  Baseline (이 트리 이 실행, 구현 전): `grep -rn 'embed-check\|embed_check' .github/workflows/` → `grep_rc=1`.
  이 조항은 **금지 게이트라서 도착 시점부터 초록이다** — 뮤테이션 확립 대상이 아니다. 실패시킬 수 있음은 뮤턴트로 보인다: 워크플로 파일 아무 곳에나 타깃명을 한 줄 넣으면 `grep_rc=0` 이 되어 게이트가 붉어진다. 확인 후 즉시 원복한다
  → 결정하는 조항: REQ-AEL-004 의 "shall not be attached to a CI build job as its automatic trigger"
- **And (적용 불가 — 게이트, v0.4.0)** 커밋 산출물을 이고 있지 **않은** 트리에서 이 항목은 **실패하지 않고**, `moai doctor` 의 종료 코드를 바꾸지 않는다
  ```bash
  REPO=$(pwd); NA="${TMPDIR:-/tmp}/t317-na"; rm -rf "$NA"
  "$REPO/bin/moai" init "$NA" --non-interactive >/dev/null; echo init=$?
  ls "$NA"/internal/template/templates/.codex/agents/moai/*.toml 2>/dev/null | wc -l   # 0 — 적용가능성 술어 거짓
  ( cd "$NA" && "$REPO/bin/moai" doctor --check "Agent Emit Embed"; echo check_exit=$? )        # exit=0, 항목 상태 ok
  ( cd "$NA" && "$REPO/bin/moai" doctor >/dev/null 2>&1; echo doctor_exit=$? )         # exit=0
  rm -rf "$NA"
  ```
  Baseline (이 트리 이 실행, 설치본 `moai` 로 스크래치 배포):
  ```console
  $ moai init <scratch>/proj --non-interactive >/dev/null 2>&1; echo init_exit=$?
  init_exit=0
  $ ls <scratch>/proj/internal/template/templates/.codex/agents/moai/
  ls: .../internal/template/templates/.codex/agents/moai/: No such file or directory
  $ ls <scratch>/proj/.codex/agents/moai/*.toml | wc -l
        11
  $ cd <scratch>/proj && moai doctor >/dev/null 2>&1; echo doctor_exit=$?
  doctor_exit=0
  ```
  **함정 — 배포된 프로젝트 루트의 `.codex/agents/moai/*.toml` 11건을 대조 대상으로 삼지 않는다.** 그것은 배포 산출물이지 커밋 산출물이 아니며, 그것을 대상으로 잡으면 이 SPEC 이 정의한 판정과 **다른 검사**가 된다(감사 iter-2 D7 의 선택지 (다)). 적용가능성 술어는 오직 커밋 경로에만 건다
  → 결정하는 조항: REQ-AEL-004 의 적용가능성 술어 + "not applicable … shall report `ok` … exit status … unchanged"
- **And (하위 디렉터리 앵커 — 게이트, v0.5.0)** 저장소의 **하위 디렉터리에서** 돌려도 같은 판정이 나온다 — 적용 가능한 트리가 "적용 불가"로 뒤집히지 않는다. 위 뮤턴트가 심긴 상태(바이너리는 뮤턴트 이전에 빌드됨)에서, 작업 디렉터리만 바꿔 두 번 돌린다:
  ```bash
  REPO=$(pwd)
  ( cd "$REPO"              && "$REPO/bin/moai" doctor --check "Agent Emit Embed"; echo root_exit=$? )   # exit ≠ 0, 항목 상태 fail
  ( cd "$REPO/internal/cli" && "$REPO/bin/moai" doctor --check "Agent Emit Embed"; echo sub_exit=$?  )   # 같은 값 — exit ≠ 0, 항목 상태 fail
  git checkout -- internal/template/templates/.codex/agents/moai/manager-git.toml
  ( cd "$REPO/internal/cli" && "$REPO/bin/moai" doctor --check "Agent Emit Embed"; echo sub_clean_exit=$? )  # exit=0, 항목 상태 ok
  ```
  판정: `sub_exit` == `root_exit` 이고 둘 다 ≠ 0 이며 두 실행 모두 항목 상태가 `fail`, 원복 후 `sub_clean_exit=0` + 항목 상태 `ok`. **하위 디렉터리 실행이 `ok` + 「커밋 산출물 부재」 사유를 내면 이 게이트는 실패다** — 그것이 적용 가능한 트리를 적용 불가로 오판하는 형태이며, 이 SPEC 이 봉쇄하겠다고 선언한 공허성의 좁은 재발이다. 항목 상태는 위 doctor 도달 게이트와 같은 방식(출력 말미 카운터 합 = 1)으로 읽는다
  이 게이트가 필요한 이유는 실측이다 — doctor 배선은 모든 항목에 `os.Getwd()` **원값**을 넘기고 프로젝트 루트로 거슬러 올라가지 않는다(`internal/cli/doctor.go:180`, 이 트리 이 실행에서 확인). 따라서 루트 해석은 판정 지점 자신의 몫이다
  → 결정하는 조항: REQ-AEL-004 적용가능성 술어의 기준점("resolved relative to the project root under check")
- **And (공허성 대조 — 기록용, 게이트 아님)** 같은 뮤턴트에 기존 테스트를 걸면 여전히 통과한다는 것을 기록한다
  ```bash
  go test ./internal/template/agentemit/... -run TestEmbedFSPresenceAndByteEquality -count=1   # PASS 예상
  ```
- **Restore** `git checkout -- <path>` → `make build` → `make embed-check` exit=0

## AC-AEL-004 *(REQ-AEL-001, REQ-AEL-002)* — 깨끗한 트리에서 build 가 초록이고 검사가 비용을 더하지 않는다

- **Given** 드리프트 0 인 트리
- **When** `make build; echo exit=$?`
- **Then** exit=0. 검사에 기인한 신규 트리 변경이 없다(기존 build 쓰기 — `*_templ.go`, `catalog.yaml` — 는 이 AC 의 대상이 아니다)

## AC-AEL-005 *(REQ-AEL-003)* — 재생성 동사는 여전히 재생성한다

- **Given** C2 소스 한 개를 바꿔 방출 결과가 커밋본과 갈린 상태
  ```bash
  printf '\n<!-- mutant-t317-c -->\n' >> internal/template/templates/.claude/agents/moai/manager-git.md
  ```
- **When** `make agents-emit && go test ./internal/template/agentemit/... -count=1`
- **Then** 방출물이 갱신되어 테스트가 `ok` 를 낸다 (재생성 경로가 (a′) 도입으로 막히지 않았음을 보인다)
- **Restore** `git checkout -- internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.codex/agents/moai/manager-git.toml`

## AC-AEL-006 *(REQ-AEL-007)* — `.PHONY` 누락이 닫힌다 **[뮤테이션 확립 — 미미(minor)]**

- **Control (RED — 수정 전에 관측하고 기록한다)** 같은 명령을 `.PHONY` 수정 **이전** 트리에서 돌리면 레시피가 실행되지 않는다. 이 트리 이 실행에서 이미 측정했다:
  ```console
  $ touch agents-emit; make agents-emit; echo exit=$?; rm -f agents-emit
  make: `agents-emit' is up to date.
  exit=0
  ```
  RED 는 이 **레시피 미실행**이지 문구가 아니다 — 아래 판정 주석을 함께 읽는다
- **Given** 저장소 루트에 `agents-emit` 이라는 이름의 파일이 없다
- **When** 동명 파일을 만들고 타깃을 부른다
  ```bash
  touch agents-emit && make agents-emit; echo exit=$?; rm -f agents-emit
  ```
- **Then** make 가 위 Control 처럼 건너뛰지 않고 **레시피를 실제로 실행한다** — 판정은 골든 update 실행 흔적(`AGENTEMIT_UPDATE=1 go test …` 의 출력)이 나타나는지로 한다
- **And** `grep -n '^\.PHONY' Makefile` 출력에 `agents-emit` 과 새 검사 타깃이 포함된다

> 기대 문구를 판정 기준으로 삼지 않는 이유: 감사(D6)는 종전의 "nothing to be done" 표기가 틀렸다고 옳게 지적했으나, 대체안으로 제시한 `make: 'agents-emit' is up to date.` 도 이 머신에서는 맞지 않는다. 실측(위 Control): **GNU Make 3.81** 은 비대칭 인용 ``make: `agents-emit' is up to date.`` 를 낸다(GNU Make 4.x 는 `'agents-emit'`). 문구 대조는 make 버전에 취약하므로 판정은 **레시피 실행 여부**라는 관측 가능한 사실에 건다.

## AC-AEL-007 *(REQ-AEL-006)* — 편집 절차가 문서에 명문화된다

- **Baseline (측정됨, 이 트리 이 실행)**: 현재 문서 어디에도 없다
  ```console
  $ grep -rn "agents-emit\|agentemit" --include="*.md" .claude/ .moai/docs/ CLAUDE.md CLAUDE.local.md
  (출력 없음 — 0 히트)
  ```
- **Given** 위 baseline
- **When** 같은 grep 을 다시 돌린다
- **Then** 히트가 ≥1 이고, 그 문단이 (1) 소스 층 경로 `internal/template/templates/.claude/agents/moai/*.md`, (2) 방출 층 경로 `internal/template/templates/.codex/agents/moai/*.toml`, (3) `make agents-emit` 동사 셋을 모두 이름으로 담는다

---

> **폐기 — 종전 AC-AEL-008 (update 분기 되읽기).** REQ-AEL-008 과 함께 폐기했다. 판정 근거는 `spec.md` 의 「폐기 판정」 절. 수락 기준은 001-007 총 7건이다.
