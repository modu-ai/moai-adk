# CC Upstream Change Analysis — 2.1.247 → 2.1.263

- **수행 일자**: 2026-09-06 (release-update research sweep; 승인 범위 = Phase 0-5, 연구 + 분류 + 업데이트 플랜 권고까지. 문서/템플릿/코드 편집 없음, 상태 파일 갱신 없음, 커밋 없음, PR 없음)
- **작업 트리**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop` (branch `develop`, HEAD `615d18c1f`) — primary 체크아웃(`main`)에는 쓰지 않았다.
- **기준선**: 상태 파일 `.moai/state/last-cc-version.json`의 `last_analyzed_version` = **2.1.246**, `last_analyzed_date` = 2026-08-27. 지시된 기준선과 일치 — 조정 불필요. 분석 창 = **2.1.246 초과 .. 2.1.263 이하**.
  - 주의: 상태 파일과 직전 리포트들은 **primary 체크아웃에만 존재**한다(`.moai/state/`는 gitignore 대상, `.moai/research/cc-update-2.1.*.md`는 미추적). 이 워크트리에서는 읽기 전용으로 참조했다.
- **1차 소스**: `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` (curl, HTTP 200, 630,840 bytes, 2026-09-06 fetch; 스냅샷 `.moai/reports/release-update-20260906/CHANGELOG-upstream.md`, 창 추출본 `window.md`)
- **npm 관측**: `npm view @anthropic-ai/claude-code version` = `2.1.263`; versions 목록 스냅샷 `.moai/reports/release-update-20260906/npm-versions.json`
- **로컬 바이너리**: `claude --version` = `2.1.263 (Claude Code)` — CHANGELOG head 및 npm latest와 3자 일치 (**doc-lag 0**)
- **개수**: total_items = **435** (11개 버전) · tier1 = **9** · tier2 = **24** · tier3 = **402**
  - 분류 방법: 435개 불릿을 전수 통독하며 Tier 1/2 해당 항목만 **개별 식별·열거**했다. Tier 3(402)은 **차감으로 산출한 잔여값**이며 개별 주석이 달려 있지 않다. 402을 "개별 분류된 402건"으로 읽으면 안 된다.
- **플랜 분류**: 원시 카운트 tier1+tier2 = 33 ≥ 10이지만 **실행 가능(actionable) 정밀 조사 = 7건, 전부 문서/설정 수준** → **umbrella SPEC 불필요, 소규모 직접 편집 권고** (선례: 2026-08-19 런 tier1+tier2=20 중 actionable 3, 2026-08-27 런 35 중 actionable 3 — 둘 다 umbrella 미트립)

## Executive Summary

이번 창은 **시리즈 최대 규모**다. 435개 불릿은 직전 최대치(2026-08-27 런 122개)의 3.5배이며, 2.1.257 한 버전만으로 104개 항목이 나왔다. 17개 버전 번호 중 **11개가 게시, 6개가 미출시**이고, npm에 있으나 CHANGELOG에 없는 doc-lag는 **이번 창에 0건**이다 — 상태 파일 이력상 5회 연속 재발하던 doc-lag가 처음으로 끊겼다.

정밀 조사(genuine drift) **7건**이 확인되었고 전부 문서·설정 수준이다. Go 코드 변경은 불필요하다. 그중 **2건은 이 저장소의 파일에서 직접 실측**했다(설정 파일 1줄, 룰 표 2행). 나머지 5건은 공식 문서와 CHANGELOG 원문으로 교차 확인했다.

가장 무거운 것은 **컨텍스트 윈도 표(GD-1)** 와 **크로스세션 메시징 가용성(GD-2)** 두 건이다. 둘 다 항상 로드되는 룰이라 모든 세션이 매 턴 틀린 사실을 읽는다.

**GD-1 (High)** — `context-window-management.md`의 6행 임계 표에서 **2행이 틀렸다**. 공식 문서 실측(2026-09-06 fetch)상 Sonnet 5 = **1M**, Fable 5.1 = **1M**인데 우리 표는 각각 200K / 256K로 적어 두었다. 결과는 안전 방향의 오류지만 비용을 태운다: Sonnet 5 세션은 실제 500K까지 갈 수 있는데 180K에서 핸드오프를 걸어, 불필요한 `/clear` + 프리픽스 재기입(1.25×)을 반복한다.

**GD-2 (High)** — `cross-session-messaging.md` § Availability constraints의 **3개 불릿 중 3개가 모두 낡았다**. 2.1.248이 Bedrock/Vertex/Foundry에 same-machine 메시징을 추가했고, 공식 문서는 **"모든 프로바이더"**(Claude Platform on AWS, Google Cloud Agent Platform 포함) 지원을 명시한다. 네이티브 Windows도 2.1.234부터 지원된다 — 이건 이번 창이 아니라 **이전 스윕이 놓친 드리프트**다. telemetry 플래그 게이팅도 2.1.248이 풀었다. 이 룰은 `kanban-dispatch.md`가 레인 nudge 도달 범위를 판정할 때 인용하므로, 틀린 제약이 실제로는 쓸 수 있는 채널을 배제하게 만든다.

**GD-6 (Medium, 실측)** — 이 저장소의 `.claude/settings.json:411`이 `permissions.defaultMode: "bypassPermissions"`를 설정하는데, 2.1.257부터 프로젝트 레벨 `.claude/settings.json`/`settings.local.json`의 이 키는 **무시된다**. 템플릿에는 없으므로 배포 사용자 영향은 0이고, dev 저장소 한정 사문(死文) 설정이다.

**주목할 반증 — 워크트리 가드는 아직 고쳐지지 않았다.** 2.1.257과 2.1.259가 각각 "worktree-isolated 세션이 git을 건드리지 않는 Bash 루프/`$VAR`/`"$(…)"`/heredoc/xargs 파이프라인을 거부하던 문제"를 고쳤다고 적었다. 그러나 **본 스윕 수행 중 2.1.263에서 git을 전혀 포함하지 않는 평범한 `for` 루프가 그대로 거부되었다**(§ Evidence E5). 즉 `kanban-dispatch.md:211`의 "복합 셸 구조는 거부된다, 단일 plain 명령으로 분해하라" 독트린은 **여전히 유효하며 완화해서는 안 된다**. 릴리스 노트만 읽고 이 독트린을 되돌렸다면 레인들이 반복해서 막혔을 것이다.

## genuine_delta

```
genuine_delta: "2.1.247(33) + 2.1.248(49) + 2.1.250(1, 내용 없음) + 2.1.251(71) +
                2.1.252(4) + 2.1.257(104) + 2.1.258(2) + 2.1.259(37) +
                2.1.260(66) + 2.1.261(67) + 2.1.263(1, 내용 없음) = 435 bullets"
versions_existing:               ["2.1.247","2.1.248","2.1.250","2.1.251","2.1.252",
                                  "2.1.257","2.1.258","2.1.259","2.1.260","2.1.261","2.1.263"]
versions_shipped_not_documented: []          # doc-lag 0건 — 5회 연속 재발 후 첫 중단
gaps_never_released:             ["2.1.249","2.1.253","2.1.254","2.1.255","2.1.256","2.1.262"]
infra_only:                      ["2.1.250","2.1.263"]   # "Bug fixes and reliability improvements" 단일 불릿
latest_confirmed_published:      "2.1.263"
latest_installed_binary:         "2.1.263"
npm_latest_at_fetch:             "2.1.263"
```

**미출시 갭 이상징후**: 2.1.258이 "macOS 12(Monterey) 실행 실패 — **2.1.255에서 유입된 회귀**"를 고쳤다고 적는다. 그런데 2.1.255는 npm에도 CHANGELOG에도 없다(`grep -c '^## 2\.1\.255$'` = 0). 공개 노트가 비공개 빌드를 원인으로 지목한 것으로, 미출시 번호대(253-256)에도 실제 코드 변경이 있었음을 시사한다. 이 창의 분석은 **공개된 11개 버전에 한정**되며, 미출시 6개에서 넘어온 변경은 관측 범위 밖이다(§ Gaps).

## Version Table

| Version | CHANGELOG | 항목 수 | 성격 | 비고 |
|---|---|---|---|---|
| 2.1.247 | 있음 | 33 | 중형 — `SendFeedback` 툴, Sonnet 5 auto-compact 1M | 피어 메시지 1줄 프리뷰 접힘 |
| 2.1.248 | 있음 | 49 | 대형 — `--restricted`, `experimental.cacheTtl`, **메시징 프로바이더 확대** | GD-2, GD-7 |
| 2.1.250 | 있음 | 1 | infra-only (내용 없음) | — |
| 2.1.251 | 있음 | 71 | 대형 — `PreModelSwitch`/`PostModelSwitch` 훅, 캐시 계측 | T2 다수 |
| 2.1.252 | 있음 | 4 | 소형 버그 수정 | — |
| 2.1.257 | 있음 | **104** | **시리즈 최대** — Fable 5.1, 권한/샌드박스/게이트웨이 대개편 | GD-1, GD-6 |
| 2.1.258 | 있음 | 2 | 핫픽스 (macOS 12 회귀) | 2.1.255 참조 이상징후 |
| 2.1.259 | 있음 | 37 | 중형 — `managedMcpServers`, 워크트리 가드 완화 시도 | E5 반증 |
| 2.1.260 | 있음 | 66 | 대형 — 권한 규칙 파서 수정, 1M auto-compact | — |
| 2.1.261 | 있음 | 67 | 대형 — `bashOutputMaxChars`, `/skill-doctor` | T2 |
| 2.1.263 | 있음 (head) | 1 | infra-only (내용 없음) | 바이너리와 일치 |
| 2.1.249 / .253 / .254 / .255 / .256 / .262 | **없음** | — | 미출시 (npm에도 없음) | 2.1.255는 .258이 회귀 원인으로 지목 |

## Tier 1 — 정밀 조사 (genuine drift, 조치 필요)

| # | 출처 | 심각도 | 우리 저장소의 현재 상태 | 검증 |
|---|---|---|---|---|
| GD-1 | 2.1.257 + 2.1.247 + 공식 문서 | High | `context-window-management.md:24-25` 표의 `Fable (256K)` / `Sonnet/Opus standard (200K)` 2행이 틀림. 실제 Fable 5.1 = 1M, Sonnet 5 = 1M | **검증됨** — 공식 모델 문서 fetch (E4) + 파일 실측 (E3) |
| GD-2 | 2.1.248 + 공식 문서 | High | `cross-session-messaging.md:22` "unavailable on Amazon Bedrock, Claude Platform on AWS, Agent Platform on Google Cloud, and Microsoft Foundry" — 2.1.248부터 same-machine은 **전 프로바이더 지원** | **검증됨** — CHANGELOG 원문 + 공식 문서 (E6) |
| GD-3 | 2.1.234 (창 이전) + 공식 문서 | Medium | `cross-session-messaging.md:21` "does not provide cross-session messaging on native Windows" — 실제로는 **2.1.234부터 지원**(cross-machine만 제외) | **검증됨** — 공식 문서 (E6). 이번 창이 아닌 **이전 스윕 누락분** |
| GD-4 | 2.1.248 | Medium | `cross-session-messaging.md:24` telemetry 4개 플래그가 채널을 끈다는 서술 — 2.1.248이 "when telemetry is disabled"에 메시징을 추가 | **부분 검증** — `DISABLE_TELEMETRY` 축은 CHANGELOG로 확인, 나머지 3개 플래그 개별 거동은 미관측 |
| GD-5 | 2.1.248 | Medium | `settings-management.md:20` + `cache-aware-execution.md:3` (각각 local/template 2본)이 캐시 TTL 표면을 `promptCacheTtl`/`subagentPromptCacheTtl` 2종으로만 열거. **에이전트 frontmatter `experimental.cacheTtl`** 신규 3번째 표면 누락 | **검증됨** — 저장소 전체 grep 0히트 (E7) |
| GD-6 | 2.1.257 | Medium | `.claude/settings.json:411`이 `permissions.defaultMode: "bypassPermissions"` 설정 — 2.1.257부터 프로젝트 레벨에서 **무시됨**. 템플릿에는 없음(배포 영향 0) | **검증됨** — 파일 실측 + JSON 경로 확인 (E8) |
| GD-7 | 2.1.251 | Low | `cross-session-messaging.md:95` + `kanban-dispatch.md:69`의 "idle notice가 작업에 대해 말하는 것은 없다" 서술 — 2.1.251이 팀메이트 최종 답변을 idle notification에 실어 보내도록 변경 | **미검증** — 아래 주의 참조 |
| GD-8 | 2.1.257 / 2.1.259 | — | 워크트리 가드 복합-셸 거부 완화 주장 → **독트린 유지 권고(조치 없음)** | **반증 관측** — 2.1.263에서 재현 (E5) |
| GD-9 | 2.1.261 | Low | `cache-aware-execution.md:21` + reference(각 local/template 2본)가 출력 상한을 `BASH_MAX_OUTPUT_LENGTH`로만 서술. 신규 `bashOutputMaxChars`/`taskOutputMaxChars`(최대 128K) 누락 | **검증됨** — grep 0히트 (E7) |

### GD-7 주의 — 성급히 고치면 안 되는 항목

2.1.251의 문구는 **"agent teams: a teammate's final answer ... now arrives in the idle notification"** 로, **in-process 에이전트 팀 팀메이트** 채널을 가리킨다. 우리 두 룰의 해당 문장은 **크로스세션 `notify_when_idle`**(별개 세션 간)을 서술한다. 두 통지가 같은 메커니즘인지는 릴리스 노트 문면만으로 확정되지 않으며, 본 런에서 실측하지 않았다.

또한 설령 통지가 최종 답변을 싣더라도 **[HARD] 독트린 자체는 무너지지 않는다** — "주장은 증거가 아니다"는 그대로다. 바뀌는 것은 "통지 내용이 비어 있다"는 **사실 서술**뿐이다. 따라서 권고는 "메커니즘 동일성을 먼저 실측하고, 확인되면 사실 서술만 좁힌다"이며, 독트린 문장은 건드리지 않는다.

## Tier 2 — 검토 가치 있음 (계약 변경 없음, 24건)

| 영역 | 항목 | 관련 표면 |
|---|---|---|
| 훅 | 2.1.251 `PreModelSwitch` / `PostModelSwitch` 신규 이벤트 — 현재 미배선(우리 settings.json은 20개 이벤트 배선) | `hook-development.md`, settings.json |
| 훅 | 2.1.251 `SessionStart` resume 훅이 세션 staleness + 재캐시 비용 추정치를 수신 | `internal/hook/session_start.go` |
| 훅 | 2.1.248 훅 stdout의 유효하지 않은 `{…}` 를 이제 **훅 오류로 보고**(과거엔 평문 취급) | `.claude/hooks/moai/*.sh` 전반 |
| 훅 | 2.1.259 blocking Stop 훅이 다음 턴의 reasoning 유실 + 캐시 미스를 유발하던 문제 수정 | `stop-goal` 훅, `goal-directive.md` |
| 훅 | 2.1.248 `PermissionRequest`/`PreToolUse` 훅의 무효 응답을 `claude agents` 행에 스키마 오류로 표기 | 훅 계약 |
| 캐시/비용 | 2.1.251 `/cost` 세션 프롬프트-캐시 라인 + 상태줄 `prompt_cache` 객체 | `internal/statusline` |
| 캐시/비용 | 2.1.260 프롬프트-캐시 미스 **원인**을 `/cost`·상태줄에 표기 | `cache-aware-execution.md` |
| 캐시/비용 | 2.1.261 in-process 팀메이트가 2턴째에 first-turn 선언을 재전송해 캐시 미스 나던 문제 수정 | agent-team 운용 |
| 캐시/비용 | 2.1.248 OAuth 토큰 갱신에 따른 시간당 1회 캐시 미스 수정 | `cache-aware-execution-reference.md` |
| 컨텍스트 | 2.1.260 1M 모델 auto-compact가 1M 한계 직전에 동작하도록 개선 | `context-window-management.md` |
| 컨텍스트 | 2.1.261 `/context` 토큰 집계를 로컬 추정으로 전환 | 동상 |
| 컨텍스트 | 2.1.260 `/rewind`가 체크포인트 백업 누락 시 성공을 오보하던 문제 수정 | Reduction Ladder |
| 모델 정책 | 2.1.257 `CLAUDE_CODE_SUBAGENT_MODEL` 이 **override → default** 로 의미 변경(에이전트 정의 `model:` 과 per-spawn 이 우선) | `agent-common-protocol.md` § Per-Spawn Model Injection |
| 모델 정책 | 2.1.257 `CLAUDE_CODE_SUBAGENT_MODEL_FORCE` 신규 | `model-policy.md` |
| 모델 정책 | 2.1.251 Opus 5 에서 effort xhigh/max + thinking off 조합이 실패하던 문제(이제 high 로 송신) | effort 독트린 |
| 모델 정책 | 2.1.259 커스텀 커맨드/스킬 frontmatter `model:` 이 인터랙티브 세션에서 무시되던 문제 수정 | 스킬/커맨드 frontmatter |
| 권한 | 2.1.257/251/259/260 Bash 권한 자동승인 우회 4종 차단(zsh `REPORTTIME`, `[[ ]]`, 산술 대입, 복합 명령 내 `permissions.ask`) | settings.json 권한 모델 |
| 권한 | 2.1.260 경로에 괄호가 든 `Edit`/`Write`/`Read` 규칙이 무시되던 문제 수정 — **우리 규칙 168건 중 해당 0건** | settings.json |
| 권한 | 2.1.260 2.1.259의 "Read() deny 를 Bash 인자에 적용" 변경 **철회** | 권한 규칙 해석 |
| 권한 | 2.1.257 프로젝트 `.claude/settings.json` `env` 가 `CLAUDE_CONFIG_DIR`/`TMPDIR` 계열을 더는 설정 못 함 — **우리 미사용, 영향 0** | settings.json |
| 워크트리 | 2.1.248 background 세션이 실행 중 워크트리 락을 보유해 `git worktree remove` 로부터 보호 | `worktree-integration.md` |
| 워크트리 | 2.1.260 미푸시 커밋 있는 background 세션 삭제 시 브랜치·커밋 수를 메시지에 명시 | "원격 착지 전 폐기 금지" 규율과 정합 |
| 다중 세션 | 2.1.259 동시 세션이 서로의 `~/.claude.json` 을 되돌리던 경쟁 조건 수정 | 다중 세션 경쟁 독트린 |
| 다중 세션 | 2.1.260 background 전환 세션이 ListAgents 에 유령 쌍둥이로 2회 표시되던 문제 수정 + 다수 세션이 한 프로젝트 디렉터리를 공유할 때의 "task output swap refused" 수정 | `kanban-dispatch.md` 역할-공백 판정, 레인 운용 |
| 스킬 | 2.1.261 `/skill-doctor` — 로드된 스킬 중 미사용분과 컨텍스트 비용 표시 | progressive disclosure 독트린 |

## Tier 3 — 정보성 (402건, 개별 주석 없음)

VS Code 확장 UI 수정(약 60건), Remote Control / 클라우드 세션, Claude apps gateway·Bedrock·Vertex·Foundry 프로바이더 배관, 터미널 렌더링·입력 처리, 플러그인 마켓플레이스, `/schedule`·`/usage`·`/radio` 등 우리가 쓰지 않는 커맨드, MCP 연결 진단 등. moai-adk-go 표면에 닿지 않는다.

## Affected-Surface Mapping (검증된 파일:라인)

| Drift | 파일 | 라인 | 미러 |
|---|---|---|---|
| GD-1 | `.claude/rules/moai/workflow/context-window-management.md` | 24, 25 (표), 64·80 (임계 산문) | 템플릿 동일본 존재 (8,882 bytes, 내용 일치) |
| GD-2 | `.claude/rules/moai/workflow/cross-session-messaging.md` | 22 | 템플릿 동일 문자열 1히트 |
| GD-3 | 동상 | 21 | 동상 |
| GD-4 | 동상 | 24 | 동상 |
| GD-5 | `.claude/rules/moai/core/settings-management.md` | 20 | 템플릿 미러 존재 |
| GD-5 | `.claude/rules/moai/workflow/cache-aware-execution.md` | 3 | 템플릿 미러 존재 |
| GD-6 | `.claude/settings.json` | 411 (`permissions.defaultMode`) | **템플릿 미러 없음** — 배포 영향 0 |
| GD-7 | `.claude/rules/moai/workflow/cross-session-messaging.md` | 95 | 템플릿 동일 문자열 1히트 |
| GD-7 | `.claude/rules/moai/workflow/kanban-dispatch.md` | 69 | 템플릿 동일본 (34,268 bytes, 크기 일치) |
| GD-8 | `.claude/rules/moai/workflow/kanban-dispatch.md` | 211 | **변경 없음 권고** |
| GD-9 | `.claude/rules/moai/workflow/cache-aware-execution.md` | 21 | 템플릿 미러 존재 |
| GD-9 | `.claude/rules/moai/workflow/cache-aware-execution-reference.md` | 34 | 템플릿 미러 존재 |

**Template-First 주의**: GD-1·2·3·4·5·7·9는 전부 `internal/template/templates/` 에 미러가 있다. 수정 시 local + template **2본을 함께** 고쳐야 하며(CLAUDE.local.md §2 Template-First), 이후 `make build` 가 필요하다. GD-6만 local 전용이다.

## Recommendation (업데이트 플랜)

**권고 처분: umbrella SPEC 불필요 — 단일 chore PR 1건.**

근거: actionable 7건이 전부 문서·설정 텍스트 수정이고 Go 코드·테스트 변경이 없다. 상호 의존이 없어 병렬 검토가 불필요하며, 파일 수는 rules 5개 × 2본(local+template) + settings.json 1개 ≈ 11개다. 2026-08-19 런(tier1+2=20, actionable 3)과 2026-08-27 런(35, actionable 3)이 같은 규모에서 umbrella를 트립하지 않은 선례를 따른다. 다만 actionable 7건은 두 선례의 2배 이상이므로, **본문 수정 규모가 rules 파일당 1개 절(section)을 넘어가면 그 시점에 umbrella 재판정**을 권고한다.

권고 순서 (의존 없음, 우선순위 순):

1. **GD-1** — `context-window-management.md` 표 2행 정정. Sonnet 5 / Fable 5.1 을 1M·50% 행으로 이동. 항상 로드 룰이라 매 세션 비용에 직결.
2. **GD-2 + GD-3 + GD-4** — `cross-session-messaging.md` § Availability constraints 3개 불릿 동시 재작성. 세 건이 같은 절이므로 분리 편집은 낭비다. cross-machine 축의 제외(Bedrock/Google Agent Platform/Foundry/native Windows)는 **살아 있으므로 축을 나눠 서술**해야 한다.
3. **GD-6** — `.claude/settings.json:411` 처분 결정. 두 갈래이며 **운영자 판단 사항**이다: (a) 사문 키 제거, (b) 유지하되 무시됨을 주석. 실제 bypass 모드가 필요하다면 user/managed settings 또는 `--permission-mode` 로 옮겨야 하므로, 단순 제거는 현재 세션 거동을 바꿀 수 있다.
4. **GD-5** — 캐시 TTL 표면 열거에 `experimental.cacheTtl` 추가. "MoAI는 아무것도 설정하지 않는다" 입장은 유지되며 열거만 보강한다.
5. **GD-9** — 출력 상한 서술에 `bashOutputMaxChars`/`taskOutputMaxChars` 추가.
6. **GD-7** — **먼저 실측**(두 idle 통지의 메커니즘 동일성) 후 사실 서술만 좁힘. 실측 전 편집 금지.
7. **GD-8** — **조치 없음**. 독트린 유지. 본 리포트의 E5 관측을 근거로 `kanban-dispatch.md:211` 에 "2.1.263 에서 재확인" 각주를 다는 것은 선택 사항.

## Follow-up List (별도 카드 후보, 본 PR 범위 밖)

- `PreModelSwitch`/`PostModelSwitch` 훅 배선 검토 — 우리 model-policy 가드는 현재 PreToolUse 스폰 감시로만 구현돼 있다. 새 이벤트가 더 정확한 표면일 수 있다.
- `/skill-doctor` 를 스킬 다이어트 계측에 편입 — progressive disclosure 예산 주장을 처음으로 **측정**할 수 있게 된다.
- 상태줄 `prompt_cache` 객체 소비 — `internal/statusline` 확장.
- 훅 stdout JSON 유효성 일제 점검 (2.1.248 변경으로 무효 JSON이 오류가 됨).
- `CLAUDE_CODE_SUBAGENT_MODEL` 의미 변경을 `model-policy.md` 에 반영.

## Evidence

| ID | Claim | Command | Observed |
|---|---|---|---|
| E1 | 설치 바이너리 = npm latest = CHANGELOG head = 2.1.263 (doc-lag 0) | `claude --version` / `npm view @anthropic-ai/claude-code version` / `head -3 CHANGELOG-upstream.md` | `2.1.263 (Claude Code)` / `2.1.263` / `## 2.1.263` |
| E2 | 창 내 게시 11 / 미출시 6 | `npm view … versions --json` → python 필터 | published: 247 248 250 251 252 257 258 259 260 261 263 · never_released: 249 253 254 255 256 262 |
| E2b | 253-256 은 CHANGELOG 에도 없음 | `grep -n '^## 2\.1\.25[3-6]$' CHANGELOG-upstream.md` | rc=1 (0히트) |
| E3 | 우리 컨텍스트 표 2행이 200K/256K | `grep -n 'Sonnet\|Fable…' context-window-management.md` | `24:\| Fable (256K) \| 256,000 tokens \| **90%** …` / `25:\| Sonnet/Opus standard (200K) \| 200,000 …` |
| E4 | 공식: Sonnet 5 = 1M, Fable 5.1 = 1M, Opus 5 = 1M, Haiku 4.5 = 200K | WebFetch `https://platform.claude.com/docs/en/about-claude/models/overview` (2026-09-06) | Context window 행: Fable 5.1 `1M tokens`, Opus 5 `1M tokens`, Sonnet 5 `1M tokens`, Haiku 4.5 `200K tokens` |
| E5 | **워크트리 가드는 2.1.263 에서 여전히 평범한 `for` 루프를 거부** | git 미포함 `for f in …; do …; done` (diff/grep 만 사용) 을 이 워크트리에서 실행 | `This session is isolated in the worktree …, but this command is too complex to verify that it stays inside the worktree. Refusing to run it` |
| E6 | 메시징 same-machine 은 전 프로바이더 지원, native Windows 는 2.1.234+ | WebSearch → `https://code.claude.com/docs/en/cross-session-messaging` | "Same-machine messaging works on every provider, including Amazon Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, and Microsoft Foundry, from 2.1.248." / "On native Windows, it requires Claude Code v2.1.234 or later." / cross-machine 은 Bedrock·Google Agent Platform·Foundry·native Windows 제외 |
| E7 | `experimental.cacheTtl`, `bashOutputMaxChars`, `taskOutputMaxChars`, `PreModelSwitch`, `PostModelSwitch`, `CLAUDE_CODE_SUBAGENT_MODEL` 은 저장소에 0히트 | `grep -rn … .claude/ internal/` | 각 0건 (`promptCacheTtl` 은 4히트 = local 2 + template 2) |
| E8 | `.claude/settings.json` 이 `permissions.defaultMode: bypassPermissions` 설정 | `sed -n '405,418p' .claude/settings.json` + python JSON 경로 순회 | `"defaultMode": "bypassPermissions",` (line 411) · `PATH: .permissions.defaultMode = bypassPermissions` |
| E9 | 권한 규칙 168건 중 괄호-경로/닫는괄호-뒤-텍스트 위반 0건 | python: settings.json allow/deny/ask 순회 | `total rules: 168` · `D12 text-after-closing-paren: 0` · `D7 parens-in-path: 0` · `unclosed-bracket: 0` |
| E10 | 불릿 수 435, 버전별 분포 | `awk` 헤딩·불릿 카운터 on `window.md` | 263:1 261:67 260:66 259:37 258:2 257:104 252:4 251:71 250:1 248:49 247:33 |
| E11 | GD 대상 4개 룰이 템플릿에 미러돼 있음 | `grep -c` 대상 문자열 on `internal/template/templates/.claude/rules/moai/workflow/*` | `Fable (256K)`=1 · `unavailable on Amazon Bedrock`=1 · `env -u VAR`=1 · `what it says about the work is nothing`=1 |

원문 스냅샷: `.moai/reports/release-update-20260906/{CHANGELOG-upstream.md, window.md, npm-versions.json, version-accounting.txt}` (본 워크트리 기준 상대 경로)

## Baseline-attribution

| 수치 | 무엇에 대해, 언제, 어느 트리에서 쟀는가 |
|---|---|
| `last_analyzed_version = 2.1.246` | primary 체크아웃 `/Users/goos/MoAI/moai-adk-go/.moai/state/last-cc-version.json` (mtime 2026-08-27), 2026-09-06 읽기. **이 워크트리에는 파일이 없다** — `.moai/state/` 는 gitignore 대상 |
| `2.1.263` (3자 일치) | 2026-09-06 본 런 실행. 바이너리는 `/Users/goos/.local/bin/claude` → `/Users/goos/.local/share/claude/versions/2.1.263` (네이티브 설치, npm 패키지 트리 없음) |
| 435 불릿 / 11 버전 | 2026-09-06 fetch 한 CHANGELOG 스냅샷에서 `awk` 로 계수. 상류가 이후 편집되면 재계수 필요 |
| 파일:라인 좌표 전부 | 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop`, branch `develop`. **런 도중 HEAD 가 `615d18c1f` → `ce4f96869` 로 이동했다**(타 레인이 카드 t472 를 develop 에 병합 — 본 에이전트는 커밋하지 않았다). `git diff --stat 615d18c1f..ce4f96869` 를 본 리포트가 인용한 8개 경로에 대해 실행한 결과 **빈 출력** — 인용 파일 중 병합에 닿은 것은 없으므로 좌표는 두 커밋 모두에서 유효하다. 병합 범위는 26개 파일, 전부 `internal/kanban/` 계열. 다른 트리·다른 커밋에서는 라인 번호가 다를 수 있다 |
| 권한 규칙 168건 | 같은 트리의 `.claude/settings.json` (이 저장소 dev 전용 파일, 템플릿 아님) |
| 모델 컨텍스트 윈도 값 | `platform.claude.com` 공식 문서, 2026-09-06 fetch. 상류 문서가 갱신되면 재확인 필요 |

## Gaps — 관측하지 않은 것

1. **미출시 6개 버전(249, 253-256, 262)의 내용** — npm·CHANGELOG 어디에도 없어 취득 불가. 2.1.258이 2.1.255를 회귀 원인으로 지목하므로 **이 번호대에 실제 코드 변경이 존재**하지만, 그 내용은 본 분석 범위 밖이다.
2. **Tier 3 402건은 개별 분류하지 않았다.** 전수 통독은 했으나 Tier 1/2 해당분만 열거했고 402는 차감 잔여값이다. 통독 과정에서 놓친 Tier 1/2 항목이 있을 수 있다.
3. **GD-7(idle notice) 메커니즘 동일성 미검증** — agent-team 팀메이트 통지와 크로스세션 `notify_when_idle` 이 같은 경로인지 실측하지 않았다. 그래서 조치를 "실측 후"로 보류했다.
4. **GD-4 telemetry 플래그 3종 미관측** — `DISABLE_TELEMETRY` 축만 CHANGELOG로 확인했다. `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `DO_NOT_TRACK`, `DISABLE_GROWTHBOOK` 각각의 현재 거동은 재보지 않았다.
5. **훅 stdout JSON 유효성 전수 점검 미실시** — 2.1.248이 무효 `{…}` 를 오류로 승격했으나, 우리 훅 래퍼들이 모든 경로에서 유효 JSON을 내는지 실행 검증하지 않았다. 후속 항목으로만 남겼다.
6. **`moai` 측 코드 영향 미검증** — `internal/` 에 대한 검색은 GD 후보 키워드 grep(E7)에 한정했다. Go 코드 전반이 새 CC 거동과 어긋나는지는 보지 않았다.
7. **템플릿 미러 대조는 문자열 히트 수준** — E11은 대상 문자열이 템플릿에 존재함을 보였을 뿐, local/template 전체 diff를 뜨지 않았다(`context-window-management.md` 만 바이트 크기 일치를 관측).
8. **Phase 6/7 미수행** — 승인 범위 밖. docs-site 4로케일·README 동기화 없음, 상태 파일 갱신 없음, 커밋·PR 없음.
9. **본 리포트 파일 자체가 gitignored 다.** `git check-ignore -v` 결과 `.gitignore:313:.moai/research/cc-update-*.md` 에 매칭된다 — 즉 이 파일은 커밋될 수 없고 **이 워크트리에만 존재**한다. 워크트리가 폐기되면 유실된다(직전 2026-08-27 런이 정확히 이 사유로 산출물을 잃고 primary 에 재기입한 전례가 있다). 보존이 필요하면 primary 체크아웃으로 복사해야 하며, 그 판단은 리드 소관이다.

## Residual-risk

- **상류 CHANGELOG는 재작성될 수 있다.** 이 시리즈에서 doc-lag 후 노트가 뒤늦게 게시된 전례가 5회 있다(상태 파일 이력). 이번 창의 "doc-lag 0" 판정은 2026-09-06 시점 스냅샷에 대한 것이며, 미출시 6개 중 일부가 나중에 노트와 함께 게시되면 재스윕이 필요하다.
- **GD-1의 방향은 안전하지만 비용이다.** 표가 틀린 방향은 "너무 일찍 핸드오프"이므로 스톨 위험을 키우지는 않는다. 다만 Sonnet 5·Fable 5.1 세션에서 불필요한 `/clear` 와 프리픽스 재기입이 반복된다. 반대로 **표를 고칠 때 Sonnet 4.x 등 실제 200K 모델을 함께 1M으로 옮기면 그때는 스톨 방향의 위험**이 생긴다 — 행 분리가 필요하다.
- **GD-6 처분이 거동을 바꿀 수 있다.** 현재 세션이 실질적으로 bypass 모드로 돌고 있다면, 그것이 이 키 때문인지 다른 경로(`--permission-mode`, user settings) 때문인지 본 런에서 판별하지 않았다. 단순 제거 전에 실제 유효 모드를 확인해야 한다.
- **E5(워크트리 가드) 는 단일 관측이다.** 한 번의 거부를 보았을 뿐, 어떤 셸 구조가 통과하고 어떤 것이 막히는지의 경계는 재지 않았다. "2.1.257/259 수정이 전혀 효과 없다"고까지는 말할 수 없다 — 말할 수 있는 것은 "git 미포함 `for` 루프는 2.1.263에서 여전히 거부된다"뿐이다.
- **공식 문서와 CHANGELOG가 어긋날 수 있다.** GD-2/GD-3은 문서 쪽이 CHANGELOG보다 넓은 지원 범위를 말한다(문서: 전 프로바이더 / 노트: Bedrock·Vertex·Foundry 3종). 본 리포트는 더 최신·더 구체적인 문서를 채택했으나, 실제 Bedrock 세션에서 검증하지는 않았다.

## Open Questions (운영자 판단 필요)

1. **GD-6 처분** — `.claude/settings.json:411` 의 무시되는 `defaultMode` 를 제거할 것인가, 주석만 달 것인가, 아니면 user/managed settings 로 이설할 것인가?
2. **GD-7 실측을 이번 배치에 포함할 것인가**, 아니면 별도 카드로 분리할 것인가?
3. **처분 확인** — 본 리포트의 권고는 단일 chore PR이다. 7건을 한 PR로 묶을지, GD-1/GD-2(항상 로드 룰, High)만 먼저 낼지?
