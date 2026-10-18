# 상류 변경 스윕 — Claude Code / Codex CLI / 베스트 프랙티스 (2026-10-15)

> 확장 스코프 상시 스윕 3차 (운영자 지시 2026-10-07: CC·Codex 두 체인지로그 + 베스트 프랙티스 조사 + HTML 제안 보고).
> 실행 주체: release-update 하네스 스페셜리스트 (읽기 전용 조사·보고만 수행).
> 상태 파일·템플릿·규칙·코드·docs-site는 건드리지 않았고, 카드 발행·커밋·PR도 하지 않았다. 카드 발행은 운영자 승인 후 리더가 한다.

## 0. 기준선 정리 (배차문과 상태 파일이 다르다)

배차문은 기준을 CC 2.1.292, Codex 0.160.1로 적었지만, 실제 기록은 그보다 앞서 있다.

| 축 | 배차문 기준 | 실제 기록 | 이번 처리 |
|---|---|---|---|
| CC | 2.1.292 | `last-cc-version.json` = **2.1.294** (2026-10-08 스윕, primary checkout 파일) | 요청대로 2.1.292→2.1.300 전체를 다룬다. 단 2.1.293·2.1.294는 10-08 판정을 그대로 쓴다. 해당 구간 CHANGELOG가 10-08 스냅샷과 바이트 단위로 같음을 `cmp`로 확인했다 (exit 0). 새로 분석한 구간은 **2.1.295–2.1.300** |
| Codex | 0.160.1 | `last-codex-version.json` **없음** → 메인 스페셜리스트 규정(REQ-RDX-004)의 기본값 `rust-v0.161.0` | 0.160.1→0.162.1 전체. 0.161.0은 10-08 판정을 재사용하고, 새로 분석한 구간은 **0.162.0·0.162.1** |

**교차 검증 기준 트리**: 이 워크트리의 HEAD는 `origin/main` = `2aab5f797`이다. 하지만 로컬 `main` = `1ced8cab9`가 1055커밋 앞서 있다 (`git rev-list --count --left-right origin/main...main` → `0 1055`). 원격 푸시는 운영자가 일괄 처리하므로, 이미 로컬에 반영된 작업을 다시 제안하지 않도록 **모든 저장소 대조는 로컬 `main`(1ced8cab9)에 `git grep`으로 수행했다**.
예: GD-1(Haiku 5.5, t1605)은 `origin/main`에는 없지만(`model-policy.md` "Haiku 5.5" 0건) 로컬 `main`에는 반영되어 있다 (merge `669962662`가 `main`에 포함되고 `origin/main`에는 없다). 그래서 이번 보고에서는 해결된 항목으로 본다.

## 1. 주장 (Claim)

| # | 주장 | 판정 |
|---|---|---|
| C1 | CC 신규 배포 버전은 2.1.295 · 2.1.296 · 2.1.299 · 2.1.300 네 개. 2.1.297·2.1.298은 npm에 게시되지 않았다. 10-08에 미배포였던 2.1.294는 이제 npm에 게시되었다 | 확정 (npm 버전 목록 + CHANGELOG 헤딩) |
| C2 | 신규 구간 불릿은 500개 (88 + 190 + 79 + 143). 선별 결과 **Tier 1 = 48 / Tier 2 = 27 / Tier 3 = 425**. 10-08 구간을 합한 2.1.292→2.1.300 전체 559개는 T1 64 / T2 30 / T3 465 | 확정 (버전별 `grep -c '^- '`), 티어 경계는 판단치 |
| C3 | 바로 손대야 할 MoAI 문서 드리프트는 2건: rm 프롬프트 변경 미반영(`coding-standards.md:147`), 서브에이전트 모델 env 잔여 미기재(`model-policy.md:38`). 나머지 Tier 1은 대부분 수혜(positive)이거나 실측으로 NO-OP를 확인했다 | 확정 (`git grep` on main) |
| C4 | Codex 0.162.0이 **관리형 git 워크트리 생성 도구**(#50148)를 들여왔다. AGENTS.md §3(`main:AGENTS.md:131-136`)은 Codex의 워크트리 경로를 `moai worktree new` + `codex -C`로만 정의하고 있어서, 이 새 경로를 다루는 계약 조항이 없다 | 확정 (릴리스 본문 + grep) |
| C5 | 설치 바이너리는 codex-cli 0.162.1인데 저장소 fixture는 `codex-0.161.0`에 고정되어 있다. MoAI 팩토리는 `codex app-server --listen` + `codex resume --remote`를 쓰므로(`managed_codex_factory.go:692`, `managed_codex_tui.go:341`), 0.162.1 #52648과 0.162.0 #50013이 직접 닿는다 → conformance 재측정이 필요하다 | 확정 |
| C6 | 2.1.300부터 로컬 트랜스크립트에 `api-request` / `api-request-blob` / `api-request-shape` 레코드가 기록된다. 이번 세션에서 실측하니 서브에이전트 트랜스크립트 바이트의 26.4%, 부모 세션의 27.3%를 차지했고, 줄 최대 572,475바이트로 1 MiB 미만이었다. served-model 리더 예산(32 MiB/200,000줄, `served_model.go:71`)에는 여유가 있지만 장시간 감사 트랜스크립트에서는 재측정이 필요하다 | 실측 (본 세션 트랜스크립트) |
| C7 | 베스트 프랙티스: 10-07·10-08 이월 과제였던 BP-3(context engineering) 정확 경로를 확정했다. 신규 공식 게시물 "Redesigning source control at Anthropic"(2026-10-13)을 원문으로 확인했고, 10-07 보고 전용 리드였던 "How we contain Claude across products"(2026-05-25)도 원문으로 확인했다 | 확정 (WebFetch) |
| C8 | 10-07의 CARD-2(tool result clearing 래더)와 CARD-3(멀티에이전트 실패 패턴 매핑)는 로컬 main에도 아직 반영되지 않았다 | 확정 (`git grep` 0건) |

## 2. 증거 (Evidence)

### 2.1 Claude Code 2.1.292 → 2.1.300

**수집**
- `curl -fsSL https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` → 9073줄 / 1,028,694바이트, sha256 `8c6fdbbd692bd9b2e108e9946948aa77200b6c8ffec4f807ea9c8cee92545139`. 스냅샷: `.moai/research/cc-changelog-snapshot-2.1.300.md` (원문 그대로).
- 헤딩: `## 2.1.300`(3행) · `## 2.1.299`(94행) · `## 2.1.296`(287행) · `## 2.1.295`(369행) · `## 2.1.294`(515행).
- 515행 이후 꼬리를 10-08 스냅샷 3–8561행과 `cmp` → exit 0. 즉 2.1.294 이하는 다시 쓰이지 않았다.
- `claude --version` → `2.1.300 (Claude Code)`, `npm view @anthropic-ai/claude-code version` → `2.1.300`. npm 버전 목록에 2.1.297·2.1.298은 없다.

**Tier 1 (48) — MoAI 관련 표면별 판정** (`main` = 1ced8cab9 기준 실측)

| # | 버전 | 항목 | 판정 | MoAI 표면 / 근거 |
|---|---|---|---|---|
| 1 | 295 | command·HTTP 훅 `onFailure: "block"` | **채택 검토 (P6)** | MoAI 훅은 모두 `bash -c '[ -f "$0" ] && exec …; exit 0'` 래퍼(`settings.json.tmpl:36-44` 등)라서, 스크립트가 없으면 이미 exit 0이다. `onFailure`가 실제로 작동하는 경우는 타임아웃·크래시뿐이다. 공식 문서: Stop/SubagentStop/TaskCompleted/TeammateIdle에는 효과가 없다 |
| 2 | 295 | 서브에이전트 skills 사전 로드 상한 32개 | NO-OP 실측 | 템플릿 에이전트 최대 2개 (`git grep -c '^  - (moai|anthropic)'`) |
| 3 | 295 | tool search 경유 MCP 설명 16,384자로 상향 | 수혜 | moai mcp-server 도구 설명 |
| 4 | 295 | 여러 줄 async 훅 JSON 무시 문제 수정 | 수혜 | 훅 출력 |
| 5 | 295 | async SessionStart 훅이 재개 때마다 같은 컨텍스트를 다시 주입하던 문제 수정 | 수혜 | 역할 코어 주입(SessionStart) |
| 6 | 295 | `/resume`·`/branch` 뒤 `CLAUDE_ENV_FILE` 반영 | 수혜 | 훅 env |
| 7 | 295 | `-p`에서 스킬 `allowed-tools`·`effort` 누락 수정 | 수혜 | 헤드리스 스킬 실행 |
| 8 | 295 | 링크된 워크트리의 서브에이전트에 부모 브랜치·상태가 보이던 문제 수정 | **직접 수혜** | `isolation: worktree` 에이전트 (이 스윕 자체도 해당) |
| 9 | 295 | in-process 팀원의 첫 SendMessage 실패 수정 | 수혜 | 메모리 "named spawn runs as in-process teammate" |
| 10 | 295 | `--tools`·`--restricted`가 나중에 등록된 내장 도구에도 적용 | 수혜 | 권한 |
| 11 | 295 | Workflow 서브에이전트에서 fork한 스킬 결과가 호출 에이전트로 전달 | 수혜 | workflow |
| 12 | 295·299 | claude.ai 커넥터·클라우드 세션 MCP 프로토콜 2026-07-28 기본 협상 | watch | 메모리 `reference_mcp_2026_07_28_check` |
| 13 | 295 | for-loop glob Bash 권한 판정 정확도 개선 | 수혜 | 권한 |
| 14 | 296 | 서브에이전트 frontmatter `autoCompactWindow` (100000–1000000, 정수로 기재) | **문서·채택 검토 (P7)** | `agent-authoring.md:51-65` 표에 행 없음 |
| 15 | 296 | `CLAUDE_CODE_WORKFLOW_SUBAGENT_MODEL` | **문서 드리프트 (P5)** | `model-policy.md:38` 잔여 목록에 없음 |
| 16 | 296 | Read `allow_large` | 메모 | 도구 표 보강 후보 |
| 17 | 296 | MCP 설명·서버 지침 기본 상한 2,048→4,096 | 수혜 | moai mcp-server instructions |
| 18 | 296 | `--debug`가 에이전트 파일의 미인식 frontmatter 필드를 보고 | 도구 | 에이전트 lint 보조 수단 |
| 19 | 296 | `--debug`에 훅 명령·결과·소요시간 기록 | 도구 | 훅 지연 측정 수단 |
| 20 | 296 | `CLAUDE_CODE_OVERLOADED_RETRY_MAX_DELAY_MS` (529 백오프 상한) | **GLM 검토 (P10)** | `glm.go`·`launcher.go`에 retry env 0건 |
| 21 | 296 | managed PreToolUse `continue:false` 수정 | NO-OP | MoAI는 managed 설정이 아님 |
| 22 | 296 | 플러그인 힌트 태그와 닮은 훅 출력이 변형되던 문제 수정 | 수혜 | 훅 출력 |
| 23 | 296 | `BASH_ARGV0` 우회 권한 수정 | 수혜 (보안) | 권한 |
| 24 | 296 | 재개된 서브에이전트에 "사용자 거절" 오표 수정 | 수혜 | 서브에이전트 |
| 25 | 299 | **유휴 압축(idle compaction)**: 200K+ 대화가 약 54분 유휴 상태면 1시간 캐시 만료 전에 압축. `idleCompaction: false`, `CLAUDE_CODE_IDLE_COMPACT_MIN_TOKENS` | **정책 검토 (P3)** | `cache-aware-execution.md:23` 지시 9 ">5-min idle gap"; 레퍼런스 TTL 표 `:19-20`; 팩토리 리더·레인은 오래 유휴 대기함 |
| 26 | 299 | UserPromptSubmit 훅이 스킬·슬래시 명령 확장보다 먼저 실행 | NO-OP 실측 | `user_prompt_submit.go:589`가 이미 `/` 접두를 처리 |
| 27 | 299 | 스킬 자체 frontmatter UserPromptSubmit 훅 타이밍 | NO-OP | 템플릿 스킬에 해당 훅 0건 |
| 28 | 299 | MCP 허용 규칙이 이름 붙은 서버에만 적용 | NO-OP 실측 | settings 템플릿 MCP 규칙 0건 |
| 29 | 299 | 에이전트 `tools:`가 이름 붙은 MCP 서버 도구만 취함 | NO-OP 실측 | 전부 `mcp__moai__<tool>` 정식명, 서버 단위 항목 0건 |
| 30 | 299 | plan 모드에서 PreToolUse 허용 쓰기가 실행되던 문제 수정 | 수혜 (보안) | 권한 |
| 31 | 299 | Windows: shell 미지정 PreToolUse·PermissionRequest 훅이 Git for Windows 없을 때 fail-open → **이제 차단** | **watch (P8)** | MoAI 훅은 exec 형식 `"command": "bash"`. Windows 동작은 미측정 |
| 32 | 299 | 플러그인 훅 전달 실패 시 차단 | NO-OP | MoAI는 플러그인 훅을 출하하지 않음 |
| 33 | 299 | 숨김·방향 제어 문자·`<>`가 든 스킬·명령 이름 거부 | NO-OP 실측 | 템플릿 스킬 디렉터리 비정상 문자 0건 |
| 34 | 299 | 재시작 후 SendMessage가 이름으로 서브에이전트를 찾음 | 수혜 | 레인 재개 |
| 35 | 299 | auto 모드 백그라운드 서브에이전트 최종 보고 중복 수정 | 수혜 | 감사 보고 |
| 36 | 299 | `__` 포함 MCP 서버 deny/ask 규칙 미적용 수정 | NO-OP | 해당 규칙 없음 |
| 37 | 299 | 샌드박스 symlink·glob 우회 수정 3건 | 수혜 (보안) | 샌드박스 |
| 38 | 299 | 정책 차단 MCP 서버를 접속 전에 차단 (2건) | NO-OP | |
| 39 | 299 | `/reload-skills`가 압축 후 스킬 목록 전체를 다시 보냄 | 메모 | 스킬 라우팅 독트린 각주 후보 |
| 40 | 299 | 런타임 조립 rm은 프롬프트하지 않음. `CLAUDE_CODE_STRICT_INLINE_SHELL_RM_PROMPT=1`이면 다시 프롬프트 | **문서 드리프트 (P4)** | `coding-standards.md:147` (live·템플릿 모두)에 해당 env 0건 |
| 41 | 299 | 스킬 이름에 `anthropic-skills:` 접두를 잘못 붙였을 때의 에러 안내 개선 | 수혜 | |
| 42 | 300 | PreToolUse·PermissionRequest `updatedInput` 스키마 검증 (미지 키 제거, 거부 입력은 deny) | NO-OP 실측 | 운영 코드에서 설정하는 곳 0건 (`registry.go:313` 병합만 존재) |
| 43 | 300 | agent 훅 `model` 별칭 해석·폴백 | NO-OP | agent/prompt 타입 훅 0건 |
| 44 | 300 | 백그라운드 명령 실행 중 종료할 때 SessionEnd 훅이 누락되던 문제 수정 | 수혜 | `settings.json.tmpl:36` SessionEnd |
| 45 | 300 | `--worktree`에서 플러그인 WorktreeCreate/Remove 훅 실행 | NO-OP | 등록 0건 |
| 46 | 300 | MCP 입력 스키마 400이면 그 도구를 파라미터 없이 전송 | 수혜 | MCP |
| 47 | 300 | 압축 후 접속한 MCP 서버의 도구·지침 고지 | 수혜 | moai mcp-server |
| 48 | 300 | 재개 시 이미 로드된 지연 도구를 다시 검색하던 문제 수정 | 수혜 | ToolSearch |

**Tier 2 (27)**: `-p` 대기 사유 stderr 출력·`-p` 응답 누락 수정, `CLAUDE_AUTO_BACKGROUND_TASKS` 수정, OSC 7501 Program Status Protocol, `CLAUDE_CODE_RETRY_WATCHDOG_MAX_WAIT_MS`, 출력 없는 `cat` 뒤 읽음 처리·mtime 불변 Edit 오판 수정, `/loop` 백그라운드 수정, `←` 백그라운드 전환 시 진행 중 턴 정지, Sonnet 5.5 캐시 읽기 단가 $0.10 (MoAI는 자체 단가 계산 없음 → NO-OP), TRANSCRIPT_LOCAL_GC 10% 규칙, Workflow CRLF 스크립트 허용, `/usage` 경고(max effort·workflows·**상시 로드 컨텍스트**·대형 도구 결과), 종료 시 재개 명령의 ID·`-w -r` 표기, **워크트리 격리 가드의 awk 차단 강화**, `--worktree` reftable 잔여 폴더 수정, 에이전트 뷰 `/effort`·`/fast`, 출력 상한에서 응답이 조용히 끝나던 문제 수정, Write의 비 UTF-8 인코딩 거부, `/btw` 헤드리스 11분 타임아웃, **트랜스크립트의 API 요청 기록 (P9, 실측)**, 백그라운드 출력 폴링·sleep 수정, 종료 1.5초 뒤 백그라운드 명령 강제 종료, `--worktree` 시작 커밋·브랜치명 수정, 스크래치패드 전체 경로, 백그라운드 세션이 Workflow 승인으로 재개, 헤드리스·SDK 재개 압축 수정 2건 + OTel `cache_creation_5m/1h` 분리.

**Tier 3 (425)**: mods(`$.ui.*`·`prompt.*`·`agent.spawn` 등), Claude Tag(Slack), Code Review, VSCode, 클라우드 세션, Remote Control, 자체 호스팅 러너, Claude apps gateway, Artifact, Windows·WSL 단말 렌더링, vim·키 입력, 텔레메트리 등. MoAI 표면 없음. 벌크 분류이므로 누락 가능성이 남는다 (Gaps 참조).

### 2.2 Codex CLI 0.160.1 → 0.162.1

**수집**: `gh api 'repos/openai/codex/releases?per_page=40'` 안정판(prerelease=false): `rust-v0.162.1` 2026-10-09T19:44:51Z (본문 453자), `rust-v0.162.0` 2026-10-08T18:55:59Z (19,434자), `rust-v0.161.0` 2026-10-07T15:58:45Z. `npm view @openai/codex version` → `0.162.1`, `codex --version` → `codex-cli 0.162.1`. 공식 changelog(`developers.openai.com/codex/changelog` → 308 → `learn.chatgpt.com/docs/changelog`)에는 CLI 버전 항목이 보이지 않아서 GitHub 릴리스 본문을 1차 자료로 삼았다.

**0.161.0 (이월, 10-08 판정)**: 14불릿 T1 4 / T2 4 / T3 6. GPT-6.1 Sol 기본 카탈로그, keyring 문서화, thread resume·SQLite 백업, filesystem escalation 의미 확장.

**0.162.0 + 0.162.1 큐레이팅 16불릿: T1 4 / T2 5 / T3 7**

| # | 항목 | 티어 | 판정·MoAI 표면 |
|---|---|---|---|
| X-1 | 관리형 git 워크트리 생성·목록 도구 (#50148, worktrees 기능 활성 시) | 1 | **계약 공백 (P1)**: `main:AGENTS.md:131-136`은 Codex CLI를 `moai worktree new` + `codex -C`로만 정의하고, `.moai/worktrees/` L1 재료화기·`WT-` 명명 규약과의 관계는 정의돼 있지 않다 |
| X-2 | 새 TUI 스레드가 서버의 모델·추론 요약 기본값을 따르고, 명시적 실행 옵션은 유지 (#50013/#50811/#50913) | 1 | **conformance (P2)**: `managed_codex_tui.go:341`은 `s.model`이 비어 있지 않을 때만 `-m`을 붙인다. 비어 있으면 서버 기본값이 적용된다 |
| X-3 | Linux 샌드박스 deny 다중 파일·쓰기 가능한 bwrap 거부·ripgrep 설정이 deny glob을 약화하던 문제 수정 (#50059/#51211/#51407/#51527) | 1 | 수혜 (보안) |
| X-4 | (0.162.1) 실행 중인 백그라운드 서버와 CLI 기본 기능 설정 차이로 기동이 실패하던 문제 → 호환성 검사를 명시적 CLI 기능 오버라이드에만 적용 (#52648) | 1 | **conformance (P2)**: 팩토리가 `codex app-server --listen` + `resume --remote` 사용 (`managed_codex_factory.go:692`) |
| X-5 | 사용자 정의 Responses 호환 공급자에 웹·원격 압축 기능 지정 (#50459) | 2 | NO-OP |
| X-6 | Code Mode JS 스트리밍 헬퍼·순위 기반 도구 검색 (#51126/#51209) | 2 | NO-OP |
| X-7 | `apply_patch`가 CRLF를 무조건 보존 (#51203) | 2 | 수혜 |
| X-8 | Responses·WebSocket `Retry-After` 존중 (#50418/#51440) | 2 | 수혜 |
| X-9 | 안정판 포인터가 뒤로 가지 않도록 보장 (#51186/#51425) | 2 | 방법론 정합 (alpha 추적 신뢰성) |
| X-10 | 핀 고정·`/copy`·URL 클릭·Windows 드라이브·PowerShell 체크섬·서명 설치기·0.162.1 TUI 크래시 (#51866) | 3 | MoAI 무관 |

**PR 제목에서 도출한 추가 watch 항목** (릴리스 본문 "Changelog" 목록. PR 본문은 열람하지 않았다):
#51329 부분 이력 서브에이전트 fork 제거 + #50082 V2 서브에이전트 동적 도구 상속 + #51463 서브에이전트 활동에 모델·effort 기록 (codex_role_audit는 `spawn_agent`가 아닌 최상위 `codex exec`이므로 영향은 제한적) ·
#51156/#51188 기본 지침을 Responses 입력 메시지로 전송 (`moai codex`의 `developer_instructions` 주입면 → P2 측정 항목) ·
#51482 스킬 식별에 PathUri 사용 (`.agents/skills` 미러) ·
#50402 명령 출력을 `aggregated_output`으로 통합 (`mcp_codex.go:565`의 주석 1건만 확인, 운영 파서 0건) ·
#50360 세션 설정 이벤트에서 initial messages 제거 (`sessionConfigured`·`initialMessages` 0건 → NO-OP 실측) ·
#50525 미지 TUI 키 strict 거부 (MoAI가 쓰는 `tui` 키 0건 → NO-OP 실측).

**fixture 상태**: 로컬 main 테스트데이터는 `codex-0.161.0`·`codex-0.161.0-auth` (10-08 CARD-4 반영 확인). 설치본은 0.162.1.

### 2.3 베스트 프랙티스

| ID | 원문 | 요지 | MoAI 매핑 |
|---|---|---|---|
| BP-3 (확정) | [Context engineering: memory, compaction, and tool clearing](https://platform.claude.com/cookbook/tool-use-context-engineering-context-engineering-tools) (Claude Cookbook, 2026-03-20) | "This is the cheapest of the three primitives: no inference cost, just a mechanical edit to the message list." / "If your context bloat is mostly re-fetchable tool output, clearing is cheaper and lossless" | 10-07 CARD-2의 정확한 경로. Reduction Ladder에 "tool result clearing" 단계 0건 (main grep) → P11 |
| BP-5 (신규) | [Redesigning source control at Anthropic](https://www.anthropic.com/engineering/redesigning-source-control) (2026-10-13) | 하루 병합이 2,500건에서 30,000건 이상으로 늘면서 소스 관리가 병목이 됨. 서버 중심 VCS(Scion)로 전환. 머지 큐 대신 병합 후 테스트·배포 전 테스트. "concurrent changes can conflict, but Claude can typically fix these conflicts faster" | 팩토리 직렬 통합 창·serial-slot 설계와 비교할 가치가 있음 (P12, 참조 전용. 우리 규모에서 그대로 적용할 근거는 없음) |
| BP-6 (리드 → 원문 확인) | [How we contain Claude across products](https://www.anthropic.com/engineering/how-we-contain-claude) (2026-05-25) | "users approved roughly 93% of permission prompts" → 프롬프트보다 격리(containment)가 신뢰할 수 있음. 경계는 환경 계층에서 먼저 강제 | branch guard·worktree 격리·read-don't-trust가 같은 방향. `onFailure: "block"` 검토(P6)의 근거 |
| BP-1·2 (이월) | managed-agents, Prompting Claude Opus 5.5 | 10-08 보고 참조 | CARD-3 범위에 합류 (P12) |
| BP-4 (유지) | "Claude 5 세대 시스템 프롬프트 80% 축소" | 공식 근거 없음 | 인용 금지 유지 |

공식 문서 확인 (Phase 3, `code.claude.com`):
- hooks: `onFailure` (v2.1.295+, Stop/SubagentStop/TaskCompleted/TeammateIdle에는 효과 없음). `updatedInput` 검증과 UserPromptSubmit 순서는 페이지에 명시되지 않음 (CHANGELOG만 출처).
- sub-agents: `autoCompactWindow` (v2.1.296+, 정수 표기 필수 — "`200k`는 200으로 읽힘"), skills "first 32 distinct names", 모델 해석 순서와 `CLAUDE_CODE_SUBAGENT_MODEL_FORCE` (v2.1.257+).
- costs: 캐시 수명은 구독 1시간이고, 사용 크레딧을 쓰는 중이거나 API 키·클라우드 공급자면 5분. 유휴 압축(idleCompaction)은 이 페이지에 언급이 없다.

## 3. 기준 귀속 (Baseline-attribution)

- CC: curl 원문 sha256(위)과 npm·`claude --version` 3자 대조. 이번 실행, 2026-10-15.
- Codex: `gh api` 릴리스 목록·본문, npm, `codex --version`. 이번 실행.
- 저장소: **로컬 `main` @ `1ced8cab9`** (2026-10-15 06:56:52 +0900, "Merge WT-3-2-0 into main (card t1630)")에 `git grep` / `git show`. 워크트리 HEAD `2aab5f797`(= origin/main)은 대조 기준으로 쓰지 않았다. 단 GD-1 반영 여부 판정에는 두 ref를 모두 사용했다.
- 트랜스크립트 실측: 이 세션의 서브에이전트 트랜스크립트(574줄 / 4,309,758바이트)와 부모 세션 트랜스크립트(4,250,455바이트). Python `-I`로 타입별 바이트를 집계했다.
- 이 보고의 수치는 MoAI 자체 도구(moai CLI)로 낸 것이 없다. 따라서 §2.2 도구 출처 귀속 문제는 해당하지 않는다.

## 4. 미검증 (Gaps)

- **워크트리 가드 거부 3건** (VCI §3.1): (a) `curl … -o`·`wc`·`grep`을 한 줄에 묶은 복합 명령 → URL의 "git" 문자열 때문에 거부되어 단일 `curl`로 재실행. (b) 에이전트별 skills 개수를 세는 `for`+`awk`+`sed` 루프 → 거부되어 `git grep -c`로 대체(본문 목록 줄 오집계 가능성은 남음). (c) `cd … && git merge-base --is-ancestor` → 거부되어 `git -C <절대경로> branch --contains`로 대체. 세 경우 모두 결론을 바꾸지 않았다.
- Runner(`hns-release-update-run.js`) 병렬 스윕은 실행하지 않았다 (오케스트레이터 소관). 스페셜리스트가 직접 분류했다.
- Phase 3 공식 문서 6종 중 hooks·sub-agents는 앞 100K자만 읽었고, settings는 원시 페이지라 파싱하지 못했다. skills·plugins·mcp는 가져오지 않았다. 따라서 `idleCompaction`의 공식 설정 문서는 확인하지 못했다 (CHANGELOG만 출처).
- Codex 공식 changelog 페이지에서 CLI 항목을 찾지 못했다. PR 본문은 열람하지 않았다 (PR 제목 기반 watch 항목은 그렇게 표기).
- Windows 실측 없음 (#31·P8). Codex conformance 실측 없음 (설치·fixture 대조는 카드 소관).
- Tier 3 425개는 벌크 분류다.
- 상태 파일 2종 갱신하지 않음 (Phase 7 인간 게이트).

## 5. 잔여 위험 (Residual-risk)

- 티어 경계는 판단치다. mods 계열은 "플러그인 훅" 신호가 있지만 MoAI 출하가 없어 Tier 3으로 내렸다.
- 트랜스크립트 증가율은 한 세션(요청 47·114회) 표본이다. 수백 회 요청하는 감사 실행에서 32 MiB에 근접하는지는 미지수다.
- 로컬 main이 원격보다 1055커밋 앞서 있어서, 원격 반영 순서에 따라 이 보고의 좌표(줄 번호)가 바뀔 수 있다.
- CHANGELOG는 상류가 다시 쓴 전례가 있다. 다음 스윕에서 sha256으로 확인해야 한다.

## 6. 개선 제안 (카드 후보 — 발행은 리더)

| ID | 제목 | 티어 | 클래스 | 근거 (상류 / 저장소) | 제안 변경 |
|---|---|---|---|---|---|
| **P1** | Codex 관리형 워크트리 도구와 AGENTS.md §3 계약 정합 | T1 | C | [rust-v0.162.0 #50148](https://github.com/openai/codex/releases/tag/rust-v0.162.0) / `main:AGENTS.md:131-136`, `AGENTS.md.tmpl:133-138` | 0.162.1에서 worktrees 기능을 켜고 생성 경로·브랜치명·잠금을 실측한다. 그 결과로 Codex 행에 "네이티브 생성 도구 사용 여부(허용/금지/`moai worktree new` 경유)"를 명시하고 `WT-` 명명·hoist 의무 적용 범위를 정한다 |
| **P2** | Codex 0.162.1 conformance 재측정 | T1 | B | [#52648](https://github.com/openai/codex/releases/tag/rust-v0.162.1), #50013, #51156, #50402 / `managed_codex_factory.go:692`, `managed_codex_tui.go:341`, `testdata/codex-0.161.0` | fixture를 0.162.1로 갱신(`resume --help`), app-server 기동·`resume --remote` 회귀, `-m` 미지정 시 실제 모델, `developer_instructions` 주입 결과, `aggregatedOutput` 형태를 대조한다 |
| **P3** | 유휴 압축(2.1.299)을 캐시 독트린과 팩토리 세션 정책에 반영 | T1 | C | [CHANGELOG 2.1.299](https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md) / `cache-aware-execution.md:23`, `cache-aware-execution-reference.md:19-20` | 지시 9에 "구독 1시간 캐시 + 200K 이상 약 54분 유휴 시 자동 압축"을 반영한다. 리더·레인 세션에 `idleCompaction` 기본값을 유지할지 끌지 결정한다 (압축 뒤 맥락이 바뀌면 read-don't-trust 근거가 흔들릴 수 있음) |
| **P4** | rm 프롬프트 변경 반영 | T1 | A | CHANGELOG 2.1.299 / `coding-standards.md:147` (live·템플릿 동일) | "읽을 수 없는 스크립트에서 런타임에 조립된 rm은 프롬프트하지 않음 — `CLAUDE_CODE_STRICT_INLINE_SHELL_RM_PROMPT=1`로 복원" 한 문장 추가, MoAI 미설정 명시 |
| **P5** | 서브에이전트 모델 env 잔여 갱신 | T1 | A | CHANGELOG 2.1.296, [sub-agents 문서](https://code.claude.com/docs/en/sub-agents) / `model-policy.md:38` | `CLAUDE_CODE_WORKFLOW_SUBAGENT_MODEL`(workflow 에이전트 전체 고정)과 `CLAUDE_CODE_SUBAGENT_MODEL_FORCE`(v2.1.257+)를 사용자 환경 잔여로 추가한다 |
| **P6** | PreToolUse 가드 훅의 `onFailure` 정책 결정 | T1 | C | [hooks 문서](https://code.claude.com/docs/en/hooks) (v2.1.295+), BP-6 / `settings.json.tmpl` 래퍼, `main-checkout-branch-guard.md` fail-open 규범 | 가드별로 타임아웃·크래시를 차단으로 볼지 판단한다. 기본 fail-open 유지가 기본안이며, 바꾸려면 근거와 예외 목록이 필요하다. Stop 계열에는 효과가 없다는 점을 문서에 명기한다 |
| P7 | `autoCompactWindow`·skills 32개 상한 문서화와 무거운 감사 에이전트 적용 검토 | T1 | A(문서) / C(적용) | sub-agents 문서 / `agent-authoring.md:51-65` | 표에 행 추가(정수 표기 주의 포함), skills 행에 "앞 32개 고유 이름" 추가. plan-auditor(본문 70KB) 등에 적용할지는 별도 판단 |
| P8 | Windows: exec 형식 `bash` 훅의 2.1.299 fail-closed 영향 측정 | T1 | B | CHANGELOG 2.1.299 / settings 템플릿 훅 정의 | Git for Windows가 없는 Windows에서 PreToolUse 훅이 모든 도구 호출을 막는지 실측하고, 막는다면 설치 진단(`moai doctor`)에 경고를 추가한다 |
| P9 | 트랜스크립트 API 요청 기록이 served-model 예산에 미치는 영향 | T2 | B | CHANGELOG 2.1.300 / `served_model.go:71`, `served_model_gate.go:138` | 장시간 감사 트랜스크립트로 바이트·줄 수를 재측정하고, 필요하면 `api-request-blob` 줄을 건너뛰게 사전 필터를 둔다 (현재 `"assistant"` 사전 필터만 있음) |
| P10 | GLM 세션 529 백오프 상한 | T2 | C | CHANGELOG 2.1.296 / `glm.go`, `launcher.go` retry env 0건 | z.ai 529 빈도를 근거로 `CLAUDE_CODE_OVERLOADED_RETRY_MAX_DELAY_MS` 주입 여부와 값을 정한다 |
| P11 | (이월 CARD-2) Reduction Ladder에 tool result clearing 단계 추가 | BP | A | BP-3 원문 / `context-window-management-detail.md` Reduction Ladder | 원문 경로 확정. "재조회 가능한 도구 출력이면 clearing, 대화·추론이면 compaction" 결정 규칙을 반영한다 |
| P12 | (이월 CARD-3 확장) 멀티에이전트·격리·소스관리 BP 매핑 참조 문서 | BP | C | BP-1·2·5·6 / 해당 문서 없음 (`git grep` 0건) | 상관 실패·격리 우선·병합 큐 대안을 MoAI 팩토리 방어 장치와 대응표로 정리한다 (참조 전용) |

**우선순위 상위 5**: P2 → P1 → P3 → P6 → P4.
umbrella SPEC은 불필요하다. 즉시 실행할 문서 드리프트는 2건(P4, P5)이고 나머지는 판단·측정 카드다.

## 7. Phase 7.5 — 개선 발견 (REQ-HRR-003 형식, confidence는 보수적 실행 시점 추정, 하한 0.70)

```jsonc
"findings": [
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md (Phase 0 / cross-reference baseline)",
    "kind": "gap",
    "summary": "Isolated sweep worktrees are cut from origin/main, which trailed local main by 1055 commits this run; the body names no rule for which ref the repo cross-reference reads, so a sweep grepping its own tree re-proposes already-landed work (GD-1 Haiku 5.5 read as open on 2aab5f797, landed on main 1ced8cab9).",
    "confidence": 0.8,
    "suggested_tier": "rule"
  },
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md (dispatch vs state)",
    "kind": "drift",
    "summary": "Dispatch named CC 2.1.292 / Codex 0.160.1 as baselines while last-cc-version.json already read 2.1.294 and last-codex-version.json did not exist; the specialist had to reconcile by hand. Phase 0 could state that the state file wins over a dispatch-quoted baseline unless --since is given, and seed the missing codex file.",
    "confidence": 0.75,
    "suggested_tier": "auto_update"
  },
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md (Phase 1 collection commands)",
    "kind": "friction",
    "summary": "Worktree guard refused a compound curl line because the raw.githubusercontent.com URL contains 'git'; the Phase 1 primary fetch should be documented as a single standalone curl -o invocation.",
    "confidence": 0.7,
    "suggested_tier": "auto_update"
  }
]
```

## 8. 다음 스윕 기준

- CC: CHANGELOG sha256이 `8c6fdbbd…`와 다르거나 npm 버전이 2.1.300보다 크면 델타 분석.
- Codex: 안정판이 rust-v0.162.1보다 새로우면 분석. 상태 파일 `last-codex-version.json`을 시드하는 일은 Phase 7a 소관이다 (이번에는 갱신하지 않음).
- BP: Scion 후속 게시물, Claude 5 프롬프팅 가이드 갱신, Cookbook context engineering 후속을 확인한다.

## Sources

- [Claude Code CHANGELOG (raw)](https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md)
- [Claude Code docs — Hooks](https://code.claude.com/docs/en/hooks)
- [Claude Code docs — Subagents](https://code.claude.com/docs/en/sub-agents)
- [Claude Code docs — Settings](https://code.claude.com/docs/en/settings)
- [Claude Code docs — Manage costs](https://code.claude.com/docs/en/costs)
- [Codex rust-v0.162.0 release](https://github.com/openai/codex/releases/tag/rust-v0.162.0)
- [Codex rust-v0.162.1 release](https://github.com/openai/codex/releases/tag/rust-v0.162.1)
- [Codex rust-v0.161.0 release](https://github.com/openai/codex/releases/tag/rust-v0.161.0)
- [Codex changelog (redirect target, CLI 항목 미확인)](https://learn.chatgpt.com/docs/changelog)
- [Claude Cookbook — Context engineering: memory, compaction, and tool clearing](https://platform.claude.com/cookbook/tool-use-context-engineering-context-engineering-tools)
- [Anthropic Engineering — Redesigning source control at Anthropic](https://www.anthropic.com/engineering/redesigning-source-control)
- [Anthropic Engineering — How we contain Claude across products](https://www.anthropic.com/engineering/how-we-contain-claude)
- [Anthropic Engineering index](https://www.anthropic.com/engineering)
- 보조 (검색 결과, 원문 아님): [Developers Digest — Codex CLI 0.162](https://www.developersdigest.tech/blog/codex-cli-0-162-managed-worktrees-pinned-tasks-2026), [Havoptic — rust-v0.162.1](https://www.havoptic.com/r/openai-codex-rust-v0.162.1)
