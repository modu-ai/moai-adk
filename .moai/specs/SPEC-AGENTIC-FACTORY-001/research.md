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

## R10 `cc -l` 리더 검증 위치 (0.3.0, 트리 `c14feec16d`)

| 사실 | 근거 |
|---|---|
| `probeIdentity = homestate.ProbeProcessIdentity` | `internal/discovery/factory_discovery.go:82` |
| `verifyCandidate`는 `probeIdentity(pid)` 결과가 `ProcessIdentityLive`이고 시작 시각이 비어 있지 않을 때만 리더로 받아들인다. 기록된 시작 시각과 비교하지 않는다 | `internal/discovery/factory_discovery.go:147-154` |
| 패키지 주석은 이를 "the ClassifyOwnerWith predicate family"라 부르지만 `ClassifyOwnerWith`를 부르지는 않는다 | `internal/discovery/factory_discovery.go:10`, `:149` |
| `ClassifyOwnerWith`의 다른 운영 호출처 | `internal/factorymsg/factory_run_retire.go:99`(iteration-2 감사자 측정을 옮김. 이 세션은 다시 실행하지 않음) |

## R11 역할 부여·Codex 기본값 위치 (0.4.0, 트리 `f0a2f86e51`)

| 사실 | 근거(이번 실행의 grep·sed) |
|---|---|
| 대화형 서명 방식 상수 `MethodInteractiveTTY = "interactive-tty"` | `internal/contract/seal.go:47`. `signer_kind` 필드는 `internal/contract/contract.go:82`. 서명 흐름 본문은 읽지 않음 |
| 로컬 `workflow.yaml`의 Codex 키: `audit.model: codex`(L26), `audit.gates.codex: required`(L38), `codex:`(L349) 아래 `review_gate:`(L350)·`enabled: true`(L351)가 다른 줄 | `.moai/config/sections/workflow.yaml` |
| 템플릿 `workflow.yaml`: `audit:`(L169-178)는 `claude`·`codex`·`glm` 하위 블록의 `model`·`effort`만 있고 `audit.model`·`gates` 없음. `codex:`(L189-192)는 `review_gate.enabled: false`만 있고 `task.allow_write` 없음 | `internal/template/templates/.moai/config/sections/workflow.yaml` |
| `defaults.go`: `ReviewGate: CodexReviewGateConfig{`(L1732) 아래 `Enabled: false,`(L1733), `AllowWrite: false,`(L1744) | `internal/config/defaults.go` |
| `llm.yaml`의 codex 세션 기본값 경로는 두 벌 모두 `llm.defaults.sessions.codex.{leader,lane}.{model,effort}`(값 `gpt-6.1-sol`, `high`/`medium`) | `.moai/config/sections/llm.yaml:27-29`, `internal/template/templates/.moai/config/sections/llm.yaml:71-82` |

## R8 판단에 쓴 해석

- claim을 slot lease와 분리한 결정(spec.md §F.1)은 R4의 REQ-RSL-006 설계 근거("잘못된 비어 있음의 비용은 실행 두 개 겹침뿐")가 writer 중복에는 맞지 않는다는 해석에 기댄다. 이 해석은 코드 주석을 읽은 결과이고, 실제 중복 writer 사고를 재현한 것은 아니다.

## R12 0.5.0 → 0.6.0 ID 대응 (0.6.0, 트리 `3275706cc9`)

0.6.0에서 REQ·AC ID의 뜻이 바뀌었다. 옛 판정 파일(iteration-1~5)의 ID는 아래 왼쪽 열의 뜻이다. 0.5.0 뜻은 `git show 3275706cc9:.moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md`의 REQ 행 첫머리로 확인했다.

| 0.5.0 REQ (뜻) | 0.6.0 위치 |
|---|---|
| REQ-AF-001 (상태 전이표·차단·보류) | REQ-AF-016 (T01) |
| REQ-AF-002 (`done` 세 증거) | REQ-AF-006 (G06) |
| REQ-AF-003 (전달 상태 네 단계) | REQ-AF-005 (G05) |
| REQ-AF-004 (원장 필드·비밀 가림) | REQ-AF-013 (G13), design §4.0 |
| REQ-AF-005 (관찰 모니터) | REQ-AF-020 (T05) |
| REQ-AF-006 (owner·writer 하나) | REQ-AF-001 (G01) |
| REQ-AF-007 (중복 dispatch 거절) | REQ-AF-001 (G01) |
| REQ-AF-008 (소유권 확인 실패 시 쓰기 거절) | REQ-AF-003 (G03) |
| REQ-AF-009 (실제 실행 우선 조회) | REQ-AF-004 (G04) |
| REQ-AF-010 (활성 판정·인계) | REQ-AF-004 (G04) |
| REQ-AF-011 (로컬 저장소 권위·거울) | 폐기 → REQ-AF-002 (G02, control ref 권위) |
| REQ-AF-012 (GLM 제외) | REQ-AF-015 (G15) |
| REQ-AF-013 (Codex 명시 설정) | REQ-AF-015 (G15) |
| REQ-AF-014 (인증·권한·429) | REQ-AF-014 (G14, 403 판별 추가) |
| REQ-AF-015 (환경 행렬) | REQ-AF-024 (T09) |
| REQ-AF-016 (브라우저 3선택) | REQ-AF-021 (T06) |
| REQ-AF-017 (설정 검증) | REQ-AF-021 (T06) |
| REQ-AF-018 (병합 게이트·귀속·잠금) | REQ-AF-007 (G07) + REQ-AF-011 (G11) |
| REQ-AF-019 (보호 비약화) | REQ-AF-008 (G08) |
| REQ-AF-020 (증거 저장 분리) | REQ-AF-019 (T04, 정제본 위치) |
| REQ-AF-021 (병합 미상·done·종료·다음 dispatch) | REQ-AF-005 (미상) + REQ-AF-006 (순서 뒤집힘: 종료가 DONE보다 먼저) |
| REQ-AF-022 (Codex 이관 순서) | REQ-AF-015 (G15) |
| REQ-AF-023 (ControlService 단독 쓰기) | REQ-AF-010 (G10) |
| REQ-AF-024 (재연결·목록·재개 명령) | REQ-AF-022 (T07) |
| REQ-AF-025 (기존 진입 명령 불변) | REQ-AF-022 (T07) |

| 0.5.0 AC (뜻) | 0.6.0 위치 |
|---|---|
| AC-AF-001~007 (시나리오 하위 시험 01~07) | AC-AF-025 하나(하위 시험 01~07) |
| AC-AF-008 (상태 엔진 조회) | AC-AF-016 |
| AC-AF-009·010 (claim·활성) | AC-AF-001·004 |
| AC-AF-011·012 (거울·원장) | AC-AF-002·013 |
| AC-AF-013 (Codex 계수) | AC-AF-015 |
| AC-AF-014·015 (GLM·인증) | AC-AF-015·014 |
| AC-AF-016·017 (브라우저·설정) | AC-AF-020 |
| AC-AF-018 (병합 게이트·귀속) | AC-AF-007 + AC-AF-011 |
| AC-AF-019 (보호 기준값) | AC-AF-008 (기준값은 regression-guard) |
| AC-AF-020 (Codex 이관 기록) | AC-AF-015 |
| AC-AF-021 (#1926·#1924 시험) | AC-AF-016 |
| AC-AF-022 (관찰·종료 코드, (e) 포함) | AC-AF-019 + AC-AF-021 |
| AC-AF-023·024 (명령·호환) | AC-AF-021 |
| AC-AF-025 (#1938·#1847·#1813 시험) | AC-AF-001 |

0.6.0의 새 AC는 위 행이 덮지 않는 것들이다: AC-AF-003(Issue 쓰기 권한), 009(PR 인계·수리), 012(접수·준비), 017(record), 018(맥락 장부), 022(template), 023(환경 행렬 전용), 024(효율).

## R13 정본 사본과 분석 입력 (0.6.0)

- 정본: `gh issue view 1948 --comments --json comments`(읽기 전용) 사본 v6. 댓글 11개, 마지막 댓글이 S53. 이 실행의 live 사본과 같았다.
- 전이표 E01~E27의 문장은 S46 댓글([7])의 개정 표를 따랐다(E03·E06이 [0]과 다름).
- 리더가 준 분석 입력(로컬 scratchpad, 커밋 안 됨): 정본 영향 분석, 접수·준비, 맥락 장부, API 효율, S52 측정, 결함. 이 SPEC에 살아남아야 할 내용은 design.md로 옮겼다. 그 파일들은 이 SPEC의 증거 위치가 아니다.
- 웹 문서: design §4.5 U1~U8(GitHub·GitLab), GitHub 저장소 문서(https://docs.github.com/en/rest/repos/repos). 모두 2026-10-29 WebFetch 요약이며 원문 대조 안 함(uncollated). 이 실행에서 다시 열지 않았다.

## R14 이 실행의 측정 (0.6.0, 트리 `3275706cc9`)

| 명령 | 결과 |
|---|---|
| `git ls-remote origin refs/heads/moai-control` | `ae3a4edcaa529bc927357b345051b96cd04605a1	refs/heads/moai-control`, exit 0 |
| `internal/homestate/process_fingerprint_unix.go` 읽기 | Linux 등은 `ps -o lstart= -p <pid>` 출력을 trim한 문자열, 비어 있지 않으면 성공 |
| `internal/homestate/factory_run_retire.go:152-168` 읽기 | `fingerprint == processStart`이면 live, 아니면 dead. 순서 비교 없음 |
| grep(worktree `af-settings-keys`, `internal/config/agentic_factory.go`) | 키 `browser.selected`, `codex.enabled`, `pr_queue.wip_limit`, `collaboration.*`, 함수 `NormalizeRepositoryPath`, `ResolveAgenticFactoryBrowser`, `CodexOptIn`, `WIPLimit` |
| SPEC ID 정규식 검사(ID별 한 줄) | `SPEC-AGENTIC-FACTORY-001` PASS, `SPEC-AGENTIC-FACTORY-IDENTITY-001` PASS. 두 ID를 한 줄 반복문으로 검사하려던 명령은 worktree 가드가 거절했다 |

이 실행은 lane-17·18·20·22의 코드를 실행하지 않았다. spec.md §B.2의 사실은 모두 보고이며 미검증이다.
