# SPEC-AGENTIC-FACTORY-001 — 조사 기록

측정 트리: `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4` (브랜치 `WT-agentic-factory-contract`, 작업 트리 깨끗함). 측정일 2026-10-29. 아래 각 항목은 이번 세션에서 직접 실행하거나 읽은 것이다. 다른 SPEC·다른 시점의 숫자는 옮겨 오지 않았다.

## R1 규범 원천

- Issue #1948: `gh issue view 1948` 결과 사본(조정자가 저장한 JSON)을 읽었다. `state: OPEN`, `comments: []`, `labels: []`, `updatedAt: 2026-10-29T03:12:53Z`, 본문 3,368자.

## R2 설계 입력 다섯 건 (읽기 전용 `gh issue view <n> --json title,body,comments,state`)

| Issue | 종료 코드 | 상태 | 읽은 범위 |
|---|---|---|---|
| #1938 | 0 | OPEN | 본문 전체, 댓글 0건 |
| #1926 | 0 | OPEN | 본문 전체, 댓글 2건(둘째는 앞 1,200자) |
| #1924 | 0 | OPEN | 본문 전체, 댓글 0건 |
| #1847 | 0 | OPEN | 본문 전체, 댓글 2건(앞 800자씩) |
| #1813 | 0 | OPEN | 본문 전체, 댓글 5건(앞 800자씩) |

## R3 충돌 PR (읽기 전용 `gh issue view <n> --json number,title,state`)

| 번호 | 상태 | 제목 |
|---|---|---|
| #1937 | OPEN | [t1980] fix(config): flip the shipped codex session defaults to terra/xhigh and harden the audit effort override |
| #1944 | OPEN | [t1985] merge: absorb origin/main (...) into the instruction-files card |
| #1943 | OPEN | fix(factory): recover parked leases and isolate hook Git environments (t1995) |
| #1916 | OPEN | fix(launcher): judge the real target of a lone '-' and slash worktree names (t1958, #1847) |
| #1921 | OPEN | [t1791] fix(cli): refuse -w <name> shadowed by an existing MoAI tree (GH #1813) |

변경 파일 목록은 읽지 않았다.

## R4 코드 사실

| 사실 | 근거 |
|---|---|
| 큐는 SQLite(`modernc.org/sqlite`) | `internal/factory/backlog_sqlite.go:35` |
| slot lease: `.moai/state/slot-leases/<resource>.json` + mutation lock + 감사 로그 | `internal/factory/slot_lease.go` 머리 주석 |
| slot lease는 기한이 지나면 소유자가 살아 있어도 `--force` 없이 인계(REQ-RSL-006) | 같은 파일 머리 주석 |
| `Stale()`: `pid <= 0`이면 살아 있다고 봄. `Expired()`: 해석 불가 기한은 만료 아님 | `internal/factory/slot_lease.go:179-198` |
| 자원 이름은 `^[a-z0-9-]{1,64}$` | `internal/factory/slot_lease.go` `slotResourceNameRE` |
| 카드 임대 키 범위 검사, 깎지 않고 거절 | `internal/homestate/factory_lease.go:39-90` |
| `renew` 결과값 `refused-expired` 등 | `internal/cli/factory_card.go:2817`, `internal/cli/mcp_factory_card.go:90` |
| `factory_card.go`는 `internal/factory`에 없고 `internal/cli`에 있음 | `find . -name factory_card.go` → `./internal/cli/factory_card.go` |
| factory 하위 명령: `next`, `stage`, `renew`, `complete`, `assign`, `status`, `decide`, `set-spec`, `reclassify`, `rebind-evidence`, `bundle`, `adopt`, `relaunch` 등 | `internal/cli/factory*.go`의 `Use:` 문자열 |
| `moai cc -f`(leader), `-l`(lane, `lane-<n>` 라벨), `--leader <name>` | `internal/cli/cc.go` 도움말 60-100행 |
| `internal/github`: Issue 파싱·종료·SPEC 연결·PR 병합·리뷰만 있고 원장 코드는 없음 | `ls internal/github`, `grep 'func .*Issue'` |
| `internal/agenticfactory` 없음 | `go test ./internal/agenticfactory/... -count=1` → `lstat ./internal/agenticfactory/: no such file or directory`, `FAIL ./internal/agenticfactory/... [setup failed]`, exit 1 |
| `agentic_factory` 문자열 없음 | `grep -rn 'agentic_factory\|agenticfactory\|AgenticFactory' internal .moai/config` → 출력 없음 |
| Codex 설정 타입·기본값 | `internal/config/types.go:1107-1135`(`CodexConfig`, `AllowWrite`), `internal/config/defaults.go:1730-1746`(`ReviewGate.Enabled: false`, `AllowWrite: false`) |

## R5 설정 사실

| 파일 | 관측 |
|---|---|
| `.moai/config/sections/workflow.yaml` | `audit.model: codex`, `audit.gates: {claude: off, codex: required, glm: off}`, `factory_lease: {duration: 15m, renew_interval: 5m}`, `factory.self_dispatch: false`, `codex.review_gate.enabled: true`, `codex.task.allow_write: true` |
| 템플릿 `workflow.yaml` | `codex.review_gate.enabled: false` |
| `.moai/config/sections/llm.yaml` | `team_mode: glm`, `glm_env_var: GLM_API_KEY`, `glm.base_url: https://api.z.ai/api/anthropic`, `defaults.sessions.codex.{leader,lane}` |
| 템플릿 `llm.yaml` | `codex.leader` 주석 "reserved: future moai codex -f", `codex.lane` 주석 "moai codex -l" |
| `.moai/config/sections/mcp-matrix.yaml` | 프로젝트 유형별 MCP 추천표. 브라우저 관련 항목은 `playwright`, `chrome-devtools`. egolite·aside 없음 |
| `.moai/config/sections/git-strategy.yaml` | `mode: manual`, `manual.workflow: git-flow`, `merge_method: squash`, `branch_protection: false`, `lead_push_threshold: 20` |

## R6 원격 보호 설정 (읽기 전용 `gh api`)

- `repos/{owner}/{repo}/branches/main/protection`: 필수 검사 `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`(`strict: false`), PR 리뷰 0명·`dismiss_stale_reviews: true`, `enforce_admins: true`, `required_linear_history: false`, `allow_force_pushes: false`, `allow_deletions: false`.
- `repos/{owner}/{repo}`: `allow_merge_commit: true`, `allow_squash_merge: true`, `allow_rebase_merge: false`, `default_branch: main`.
- rulesets: `Release tag immutability (v*)`(tag 대상, active) 하나.
- `git log --merges --oneline -3 origin/main`: `fc17f49b10 Merge pull request #1862 ...` 등 merge commit 존재.

## R7 브라우저·QA 관련

- SPEC-ASIDE-BROWSER-001(status `completed`): REQ-ASB-009 "Aside SHALL NOT be auto-detected, SHALL NOT appear among the Phase 0.5 selection options, and SHALL NOT be recommended", REQ-ASB-012 조용한 대체, REQ-ASB-007 subagent 호출 금지.
- egolite: `grep -rn -i 'egolite\|ego-lite\|ego-browser' .claude .agents .moai/config internal` → 출력 없음.
- QA: `.claude/skills`에 `moai-e2e`·`moai-gate` 디렉터리 없음. `.claude/skills/moai/workflows/{gate,e2e,review}.md` 존재. `.claude/skills/moai-workflow-testing`, `.claude/skills/verify` 존재. `.agents/skills/moai-gate`, `.agents/skills/moai-e2e` 존재.
- t1965: `.moai/reports/t1965` 없음. SPEC-FETCH-WEB-001 spec.md 95·104행이 "t1965 linkage"로 언급.
- lane-22 대응표: `.moai/reports/agentic-factory/`가 이 트리와 주 체크아웃 모두에 없음.

## R9 활성 판정 함수 (0.2.0, 트리 `e14b5bbac2`)

| 사실 | 근거 |
|---|---|
| `FactoryProcessAlive(pid)`: `pid <= 0` → `false`; `syscall.Kill(pid, 0)`이 nil 또는 `EPERM`이면 `true`. 시작 시각을 보지 않음 | `internal/factory/factory_alive_unix.go:22-28` |
| Windows 짝: `OpenProcess` + `GetExitCodeProcess` | `internal/factory/factory_alive_windows.go:28` |
| `FactoryProcessAlive` 호출처 | `grep -rn 'FactoryProcessAlive(' internal cmd`(시험·정의 제외) → `internal/cli/factory_card.go:401`, `internal/cli/codex_review_gate.go:363`, `internal/cli/gate_lock.go:242`, `internal/cli/integration_wait.go:286`, `internal/factory/integration_lock.go:232`, `internal/factory/slot_lease.go:183` |
| slot lease `Stale()`은 `pid <= 0`을 먼저 "살아 있음"으로 처리한 뒤 `FactoryProcessAlive`를 부름 | `internal/factory/slot_lease.go:175-183` |
| `ClassifyOwnerWith(probe, pid, processStart)`: `live` / `dead` / `indeterminate`. 시작 시각 불일치 = `dead`(PID 재사용) | `internal/homestate/factory_run_retire.go:152-168`, 상수 `:20-22` |
| `DefaultOwnerClassifier` → `ProbeProcessIdentity` | `internal/homestate/factory_run_retire.go:139-141`, `internal/homestate/profile_lease.go:210-220` |
| 시작 시각 원천: `ps -o lstart=`(`!windows && !darwin`), macOS `sysctl kern.proc.pid` P_starttime, Windows `GetProcessTimes` | `internal/homestate/process_fingerprint_unix.go:12`, `process_fingerprint_darwin.go`, `process_fingerprint_windows.go` |
| `moai cc -l` 도움말: 합류 시 리더를 "pid + process-start"로 검증 | `internal/cli/cc.go:69-70` |
| `--leader`는 `-l` 필요: `if !p.LaneRole { return p, errors.New(leaderNeedsLaneEntry) }` | `internal/cli/factory.go:284-286` |
| `approve` 명령 `Use: "approve <card-id> --run <run-id>"` | `internal/cli/leader_approval_gate.go:286` |

## R8 판단에 쓴 해석

- claim을 slot lease와 분리한 결정(spec.md §F.1)은 R4의 REQ-RSL-006 설계 근거("잘못된 비어 있음의 비용은 실행 두 개 겹침뿐")가 writer 중복에는 맞지 않는다는 해석에 기댄다. 이 해석은 코드 주석을 읽은 결과이고, 실제 중복 writer 사고를 재현한 것은 아니다.
