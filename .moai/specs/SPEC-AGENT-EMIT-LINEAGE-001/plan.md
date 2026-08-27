# PLAN: SPEC-AGENT-EMIT-LINEAGE-001

## A. 맥락

- card: t317 · worktree `.claude/worktrees/t317` (branch `WT-agent-emit-lineage`, base `48eb945df`)
- 근거 문서: `.moai/reports/t317/measurement.md` (이 트리 이 실행에서 잰 증거 8건)
- 이 SPEC 은 **존재하는 드리프트를 고치는 카드가 아니다.** 현재 트리는 드리프트 0 이다(실측 4: `go test ./internal/template/agentemit/...` → `ok`). 다루는 것은 드리프트가 **어느 지점에서 늦게 잡히는지**다.

## B. Tier 판정 — M (v0.4.0 재판정; v0.3.0 까지 S)

v0.3.0 까지 이 절은 영향 파일을 **"3-5"** 로 추정했다. 그 추정은 doctor 항목 등록(v0.3.0 결정)이 들어오기 **전** 값이었고, 그 개정에서 다시 계산하지 않았다 — 이 SPEC 스스로 "범위가 늘면 Tier 를 다시 판정한다"고 적어 놓고 하지 않은 것이 감사 iter-2 D8 이다. v0.4.0 이 추정을 **전수 열거**로 바꾼다.

### B.1 영향 파일 전수 열거 (추정 아님)

| # | 파일 | 신규/편집 | 무엇을 하는가 | 어느 마일스톤 |
|---|---|---|---|---|
| 1 | `Makefile` | 편집 | `.PHONY` 행에 `agents-emit` + 새 검사 타깃 추가(`:16`), `build` 선행에 드리프트 검사 추가(`:23`), `embed-check` 타깃 신설 | M1·M2·M3 |
| 2 | `internal/cli/doctor_<name>.go` | **신규** | 임베드 축 판정 본체 + doctor 항목. 이 리포의 doctor 항목은 파일 1개에 사는 것이 규약이다(`doctor_mcp_version.go`, `doctor_disk.go`, `doctor_hook.go` …) | M1 |
| 3 | `internal/cli/doctor_<name>_test.go` | **신규** | 위 본체의 테스트. 같은 규약(`doctor_mcp_version_test.go` 등 모든 doctor 항목이 짝 테스트를 갖는다) | M1 |
| 4 | `internal/cli/doctor.go` | 편집 | 항목 등록 1행. 등록 지점은 `moaiChecks` 슬라이스이며, 선례가 `:201` 에 있다(`{mcpServerVersionCheckName, func(v bool) DiagnosticCheck { return checkMCPServerVersion(cwd, v) }}`) | M1 |
| 5 | `CLAUDE.local.md` | 편집 | 편집 절차 명문화(REQ-AEL-006). 배포 템플릿 중립성(§25) 때문에 템플릿 트리에는 넣지 않는다 | M3 |

**5건.** 이 세션에서 확인한 사실 두 가지가 2·3·4 를 뒷받침한다: doctor 항목은 파일 1개 + 짝 테스트 1개로 살고(`ls internal/cli/doctor*.go`), 등록은 `doctor.go` 의 슬라이스 1행이다(`grep -n 'checkMCPServerVersion' internal/cli/doctor.go` → `:201`). `internal/cli/doctor_golden_test.go` 는 항목 이름 목록을 고정하지 않으므로(같은 세션 grep, 0 히트) 여섯 번째 편집 파일이 되지 않는다.

**하한이다, 상한이 아니다.** `make embed-check` 는 `BIN=<path>` 로 임의 바이너리를 겨눠야 하는데 `moai doctor` 에는 그 파라미터가 없다. run-phase 가 이를 별도 진입점(예: `internal/template/scripts/` 의 `go run` 스크립트, `gen-catalog-hashes.go` 선례)으로 푸는 형태를 고르면 파일이 6-7 건으로 는다. 그 경우 `module:` 에 해당 패키지를 추가해야 한다 — 형태가 바뀌면 `module:` 도 함께 바뀐다.

### B.2 산술과 판정

SSOT(`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier) 인용:

> | S (Simple) | < 300 LOC | **< 5 files** | **2 files**: spec.md + plan.md (AC inline in spec.md §3) | 0.75 |
> | M (Medium) | 300 - 1000 LOC | **5 - 15 files** | **3 files**: spec.md + plan.md + acceptance.md | 0.80 |

- 파일 수 **5**. Tier S 의 조건은 `< 5` 이므로 `5 < 5` 는 **거짓** → Tier S 실격. `5` 는 Tier M 의 밴드 `5 - 15` 에 **정확히 들어간다**.
- LOC 추정은 여전히 `< 300` 으로 Tier M 의 밴드(`300 - 1000`) 아래다. 두 축이 갈리지만 판정은 M 이다: SSOT 가 LOC 열을 "Scope **guidance**" 로, 파일 수를 조건으로 적고 있고, Tier S 자격은 두 축의 **연언**(`< 300 LOC` AND `< 5 files`)이라 한쪽만 깨져도 실격이기 때문이다. 승격 방향으로 트는 것이 이 리포가 명시한 안티패턴("오버헤드를 피하려고 큰 SPEC 을 S 로 분류")의 반대 방향이기도 하다.
- 마일스톤 **3**(M1-M3), 헌장성 **없음**(빌드 배선 + doctor 항목 + 문서. ZONE 규칙·에이전트 계약을 바꾸지 않는다).

**판정: Tier M.**

### B.3 승격의 결과

| 축 | v0.3.0 (S) | v0.4.0 (M) |
|---|---|---|
| 산출물 | spec.md + plan.md (+ progress.md) | spec.md + plan.md + **acceptance.md** (+ progress.md) |
| 수락 기준 위치 | spec.md §3 인라인 | **`acceptance.md`** — v0.4.0 에서 분리 완료. spec.md §3 은 포인터 + REQ↔AC 사상표만 남는다 |
| REQ/AC 상한 | 각 8 | 각 **16** (두 축 독립). 실제 항목 수는 **7 / 7 로 불변** |
| plan-auditor PASS 임계 | 0.75 | **0.80** |
| 위임 템플릿 | 최소형 허용 | Section A-E 5절 템플릿 **필수**(`manager-develop-prompt-template.md` § Applicability) |

임계 상향은 이미 받은 점수에 소급 위험이 없다 — iter-2 종합 **0.82** 는 0.80 도 넘는다. 다만 다음 델타 재감사는 0.80 기준으로 판정된다.

**다시 늘어나면 다시 판정한다.** 파일 수가 15 를 넘으면 Tier L(산출물 5종 + 임계 0.85)로 재판정한다. 이 규칙을 이번에 지키지 않은 것이 D8 이었으므로, 규칙을 반복해 적는 것으로 끝내지 않고 위 B.1 을 **열거**로 유지한다 — 추정치는 갱신되지 않은 채 남는다는 것이 이번의 교훈이다.

## C. 마일스톤 — 되돌리기 어려운 결정부터

순서는 의존성이 아니라 **바뀔 확률**로 잡았다. M1 이 새 표면(사용자가 부르는 동사)을 만들므로 리뷰가 가장 필요하고, M3 는 기계적이라 뒤에 둔다.

### M1 — 임베드 축 판정 지점 (REQ-AEL-004, 005 / AC-AEL-003) · Priority High

**결정해야 할 것: 판정 지점의 위치와 추출 경로.**

실측 8 이 못박은 제약: **`go test` 안에서는 이 판정을 세울 수 없다.** `go test` 는 테스트 바이너리를 매번 새로 컴파일하고 `//go:embed all:templates` 가 그 시점의 커밋본을 읽으므로, 임베드 바이트와 커밋 바이트가 함께 움직인다(동어반복). 따라서 판정은 **이미 존재하는 빌드 산출물**에 대고 이뤄져야 한다.

명세:

- 새 verb `make embed-check` (타깃명은 run-phase 확정 가능). **`build` 를 선행으로 갖지 않는다** — 갓 빌드한 바이너리는 정의상 커밋본과 일치하므로, build 직후에만 돌 수 있는 검사는 M1 이 겨눈 결함을 다시 동어반복으로 만든다.
- 검사 대상 바이너리는 기본 `bin/moai`, `BIN=<path>` 로 대체 가능(설치본 `~/go/bin/moai` 판정용).
- **같은 검사를 `moai doctor` 항목으로 등록한다**(위 「자동 호출 지점」 결정). verb 와 doctor 항목은 **둘 다** 존재하며 같은 판정 로직을 공유한다 — verb 가 doctor 로 대체되는 것이 아니다. 항목명은 `--check` 로 개별 호출 가능해야 한다(`doctor_mcp_version.go` 의 `mcpServerVersionCheckName` 상수 패턴을 따른다). 등록 지점은 `internal/cli/doctor.go` 의 `moaiChecks` 슬라이스(선례 `:201`), 항목명 문자열은 run-phase 확정.
- **적용가능성은 요구 층이 못박았다 — run-phase 가 다시 고르지 않는다(v0.4.0, 감사 D7).** 항목은 **커밋 산출물을 이고 있는 트리에서만 판정한다**: 술어는 `internal/template/templates/.codex/agents/moai/*.toml` 이 프로젝트 루트 기준으로 1건 이상 잡히는가이며, 0건이면 **적용 불가**로 `ok` 를 내고 종료 코드를 바꾸지 않는다. 배포 사용자 프로젝트가 이 경우다(이 세션 실측: 그 경로 부재, `moai doctor` → exit 0). 배포 프로젝트 루트의 `.codex/agents/moai/*.toml` 11건을 대신 대조 대상으로 삼는 것은 **금지**다 — 배포 산출물이지 커밋 산출물이 아니어서 SPEC 이 정의한 판정과 다른 검사가 된다. 반대로 커밋 산출물이 **있는데** 바이너리를 읽을 수 없으면 그것은 실패다(D3 의 공허성 봉투). 전문은 `spec.md` REQ-AEL-004 「Applicability」.
- **추출 경로 (실증됨 — iter-1 감사 §3).** 판정 대상 바이너리 자신의 배포 경로로 스크래치 디렉터리에 템플릿을 풀어 `.codex/agents/moai/*.toml` 을 꺼내고 커밋본과 sha256 대조한다. 동사는 **`moai init <scratch> --non-interactive`** 로 확정한다 — `update` 가 아니다. 감사가 설치본(v3.1.3-rc.5 / `22df80e90`)으로 실행해 다음을 세웠다:
  - `.codex` 는 **기본 배포 대상**이다. `--agent codex` 도 `--all` 도 필요 없었고, `.codex/**/*.toml` 11건이 그대로 나왔다.
  - 배포 경로는 **바이트를 보존한다.** 11건 중 10건이 커밋본과 완전 일치.
  - 유일한 차이(`manager-develop.toml`)는 배포 경로의 변형이 아니라 **설치본 빌드 커밋과 이 트리 사이의 진짜 내용 스큐**다 — 즉 **이 검사가 잡아야 할 상태가 지금 실제로 존재한다**(살아있는 양성 검출). 임베드 축은 가설이 아니다.
- **2순위 폴백(바이너리에 임베드 자산 덤프 경로 추가)은 폐기했다.** 실행 불가능해서가 아니라 **불필요해서**다: 1순위가 성립한 이상, 유지자 전용 검사 하나를 위해 배포되는 바이너리에 CLI 표면을 더하는 것은 순수한 표면 확대다. run-phase 는 이것을 재설계하지 말고 없는 것으로 취급한다.
- 스크래치는 `$TMPDIR` 하위. 저장소 트리에 아무것도 쓰지 않는다(REQ-AEL-002 가 이 검사를 **직접** 구속한다 — v0.2.0 에서 주어를 넓혔다). 이 조항은 형식이 아니다: 위 추출은 스크래치에 프로젝트 전체를 배포하는 무거운 쓰기이며, 감사는 git init 과 훅 설치가 동반되는 것을 관측했다.

종료 조건: AC-AEL-003 의 뮤턴트가 **죽는다**(RED 관측 + 원복 후 GREEN). 같은 뮤턴트에 기존 `TestEmbedFSPresenceAndByteEquality` 를 걸어 **여전히 통과함**을 기록으로 남긴다 — 새 검사가 기존 검사가 못 보던 축을 실제로 본다는 대조군이다. 여기에 v0.2.0 이 더한 두 게이트를 함께 만족해야 한다(감사 D3): **비교 기수 보고**(대체 대상인 `golden_test.go:285` 의 `count != 11` 단언에 대응하는 조항) 와 **바이너리 부재 시 실패 종료**. 둘 다 없으면 부분 성공한 추출이 조용히 통과한다.

**자동 호출 지점 — 결정됨(운영자, v0.3.0): (iii) `moai doctor` 항목으로 편입한다.** `make embed-check` verb 자체는 어느 안에서도 M1 산출물로 남는다. 갈린 것은 자동 트리거뿐이었고, 두 근거가 이를 취향이 아니라 측정으로 못박는다.

- **근거 1 — (ii) CI 빌드 잡은 기계적으로 탈락한다.** CI 는 자신이 검사하는 그 커밋에서 바이너리를 빌드하므로, 임베드 바이트와 커밋 산출물은 **정의상** 일치한다. 즉 CI 빌드 잡에 건 검사는 거기서 **결코 실패할 수 없다** — 실측 8 이 `TestEmbedFSPresenceAndByteEquality` 에서 드러낸 것과 **동일한 동어반복**이다. 공허성 근절을 주제로 삼은 SPEC 에 두 번째 공허한 가드를 다는 셈이다. 진짜 창은 **노후한 로컬/설치 바이너리**이고, CI 에는 그런 바이너리가 없다.
- **근거 2 — (iii) 는 같은 병에 대한 이 코드베이스의 기존 처방이다.** `internal/cli/doctor_mcp_version.go` 가 자기 목적을 이렇게 적는다: 장수하는 산출물이 이전 빌드를 그대로 이고 있는 상태 — "The host spawns the MCP server once per session and never respawns it on reinstall, so `make install` leaves the host talking to the previous build. **Nothing surfaces that**" — 를, 설치 바이너리와 빌드 스탬프를 대조해 진단한다(`checkMCPServerVersionAgainst`, `commitsMatch`). 같은 병, 같은 처방 형태가 이미 `doctor` 안에 산다. 덧붙여 `moai doctor --check <name>` 이 항목을 개별 호출 가능하게 하므로(`internal/cli/doctor.go:58` 의 `check` 플래그, `doctor_mcp_version.go:24-26` 의 "also the value accepted by `moai doctor --check`"), (iii) 는 자동 트리거와 스크립트 가능한 동사를 **동시에** 제공한다 — (i) 을 버리는 것이 아니라 포함한다.

**받아들인 비용(운영자가 승인한 표면 변화)**: `moai doctor` 출력에 **행 1개**가 추가된다. 그 이상은 없다 — 새 최상위 CLI 동사도, 새 플래그도, **종료 코드 변경도** 아니다.

v0.4.0 정정(감사 D7): 이 비용 서술이 v0.3.0 요구 문면과 어긋나 있었다. 「대상 부재 시 실패」를 배포 문맥까지 문면대로 읽으면 배포 프로젝트마다 `moai doctor` 가 exit 1 로 뒤집히고, 실패 행 하나는 "출력 한 줄"이 아니라 종료 코드 변경이다 — 운영자가 승인한 적 없는 형태다. 위 적용가능성 조항이 그 간극을 닫으므로 이 비용 서술은 **이제 문면과 일치한다**: 배포 프로젝트가 보는 것은 `ok` 행 **하나**, 종료 코드는 **불변**. 재승인이 필요한 안(종료 코드가 바뀌는 안)은 고르지 않았다.

### M2 — build 선행 읽기전용 드리프트 검사 (REQ-AEL-001, 002, 003 / AC-AEL-001, 002, 004, 005) · Priority High

**결정해야 할 것: 검사냐 재생성이냐 — 이미 검사로 확정.**

- `build` 의 선행에 소스 층 드리프트 **검사**를 건다. 현재 선행은 `templ-generate` 뿐이다(실측 1).
- 검사는 `AGENTEMIT_UPDATE` 없이 골든을 돌리는 형태가 1순위다(비용 0.419s, 실측 5).
- **재생성은 절대 하지 않는다.** 실측 5 가 "빌드는 결정적이어야 한다"는 반론을 약화시켰지만(build 는 이미 `*_templ.go` 와 `catalog.yaml` 을 쓴다), 그것이 **재생성**을 정당화하지는 않는다: build 가 잘못된 손편집을 조용히 덮으면 CI 가 볼 증거가 사라진다. 재생성은 지금처럼 `make agents-emit` 이라는 명시적 동사로만.
- 실패 메시지는 드리프트 경로를 이름으로 지목하고, 복구 동사(`make agents-emit`)를 함께 안내한다.

종료 조건: AC-AEL-001(RED) + AC-AEL-002(마커 생존 = 재생성 안 함) + AC-AEL-004(깨끗한 트리 GREEN) + AC-AEL-005(재생성 동사 보존).

### M3 — 절차 명문화 + 미미한 기계적 항목 (REQ-AEL-006, 007 / AC-AEL-006, 007) · Priority Medium

- **문서(REQ-AEL-006)**: 현재 `agents-emit`/`agentemit` 은 문서 어디에도 없다(AC-AEL-007 baseline, 0 히트). 편집 절차에 "C2 를 고쳤으면 `make agents-emit`" 을 소스 층·방출 층 경로와 함께 적는다. 게재 후보는 `CLAUDE.local.md` §2 Template-First Rule 인접(유지자가 실제로 읽는 자리)이며, 배포 템플릿 중립성(§25)에 걸리지 않도록 **템플릿 트리에는 넣지 않는다**.
- **`.PHONY`(REQ-AEL-007)**: `Makefile:16` 목록에 `agents-emit` 이 없다. 새 검사 타깃과 함께 추가한다. 현재 동명 경로가 없으므로 잠재 결함이다 — **작은 사안으로 적고 부풀리지 않는다.** AC-AEL-006 의 RED 대조군은 이미 측정돼 있다(수정 전 `make` 가 레시피를 건너뛰고 exit=0).
- ~~update 분기 되읽기(REQ-AEL-008)~~ — **v0.2.0 에서 요구·수락 함께 폐기.** 겨냥한 두 실패 유형(권한 실패·부분 쓰기)이 이미 기존 코드에서 판정되고, 남는 유형은 되읽기로 닿지 않는다. 판정 근거는 `spec.md` §「폐기 판정」. run-phase 는 이 항목을 구현하지 않는다.

**M3 단독으로는 이 카드의 처방이 될 수 없다.** 문서는 기계적 가드가 아니고, 이 결함은 우연이 잡아낸 것이다(§1.1). M1·M2 없이 M3 만 착지하면 카드는 실패다.

## D. 위험

| 위험 | 형태 | 완화 |
|---|---|---|
| 새 검사가 또 공허하다 | 실측 8 이 이 패키지에서 실제로 일어난 일이다 | AC 3건을 뮤테이션 확립으로 못박았다. RED 를 보이지 못한 가드는 착지 불가 |
| build 가 느려진다 | 모든 로컬 빌드에 검사가 붙는다 | 실측 비용 0.419s. 초과하면 설계 재검토(§5 제약) |
| 추출이 부분 성공한다 | 일부 경로만 비교하고 통과 — 이 SPEC 이 겨눈 공허성의 재발 | REQ-AEL-004 의 기수 조항 + AC-AEL-003 의 기수 게이트가 착지 조건이다 |
| `AGENTEMIT_UPDATE` 가 다른 경로로 주입 | 실측의 잔여 위험 — grep 0 회지만 전 환경 스캔 아님 | 검사 구현이 자기 실행 환경에서 해당 변수를 명시적으로 제거하고 돌린다 |

## E. 자가 검증 (run-phase 가 채운다)

- [ ] AC-AEL-001 RED 관측 + 원복 GREEN
- [ ] AC-AEL-002 마커 생존 카운트 = 1
- [ ] AC-AEL-003 뮤턴트 사망 + 기존 테스트 생존 대조 기록 + **기수 보고 = 커밋본 개수** + **바이너리 부재 시 exit ≠ 0**
- [ ] AC-AEL-003 v0.4.0 4게이트: verb 도달(`make embed-check` exit=0) · doctor 도달(뮤턴트 상태에서 `bin/moai doctor --check "<항목명>"` exit ≠ 0, 원복 후 exit=0, 카운터 합 1) · CI 미부착(`grep -rn 'embed-check' .github/workflows/` rc=1 + 뮤턴트로 rc=0 확인 후 원복) · **적용 불가**(스크래치 배포 프로젝트에서 항목 `ok` + `moai doctor` exit=0)
- [ ] AC-AEL-006 RED 대조군(수정 전 레시피 미실행) 재관측 후 GREEN(레시피 실행)
- [ ] AC-AEL-004 / 005 / 007
- [ ] 영향 패키지 테스트: `go test ./internal/template/... -count=1`
- [ ] 트리 정결: `git status --short` 에 의도한 경로만

## F. 안티패턴 — 하지 말 것

- build 가 방출물을 **재생성**하게 만들기 (증거 소각)
- 임베드 축 판정을 `go test` 안에 세우기 (실측 8 — 원리상 동어반복)
- 소스 층 CI 검출을 새로 만들기 (실측 2 — 이미 있다)
- C1↔C2 차이 7건을 "고치러" 들어가기 (범위 밖, 의도된 분기)
- `.PHONY`·되읽기 두 미미 항목을 카드의 주 성과로 보고하기

## G. 교차 참조

- `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/acceptance.md` — 수락 기준 7건(v0.4.0 Tier M 승격으로 `spec.md §3` 에서 분리)
- `.moai/reports/t317/measurement.md` — 실측 1-8
- `internal/cli/doctor.go:201` (항목 등록 선례), `:121` + `:140-146` (`doctorExitStatus` — Fail → exit 1 승격)
- `internal/cli/doctor_mcp_version.go:54-58` (대조 대상 부재를 `CheckOK` 로 내는 선례), `internal/cli/uikit/types.go:12-17` (상태 열거 ok/warn/fail — "건너뜀" 없음)
- `Makefile:16` (`.PHONY`), `Makefile:23-28` (`build` / `agents-emit`)
- `internal/template/agentemit/golden_test.go:80` (골든 본체), `:255` (공허한 임베드 테스트)
- `internal/template/embed.go:28` (`//go:embed all:templates`)
- lane-16 / t316 — 짝이 되는 발견. `.moai/reports/t316/plan-audit-iter1.md` 는 **primary 체크아웃**에 있으며 형태 참조일 뿐, 이 SPEC 은 그것에 의존하지 않는다
