# v3.1.1 배치 커밋 분석 + 문서 커버리지 매트릭스

- 작성: 2026-08-18 (oss-docs content-author, phase 1)
- 대상: `release/v3.1.1` @ `1bd9140eb`, 범위 `v3.1.0..HEAD`
- 커밋 수: 총 233 (비머지 151)
- 산출 방법: `git log --oneline --no-merges v3.1.0..HEAD` 전량 분류 + 각 항목의
  문서 표면(README×4, docs-site/content/ko) grep 대조 + 소스 코드 검증
  (`internal/cli`, `internal/statusline`, `internal/web`, `internal/kanban`).

## 분류 요약

| 분류 | 개수 |
|---|---|
| A. 사용자 체감 신기능 (user-facing features) | 22 |
| B. 사용자 체감 동작 변경 (behavior changes) | 8 |
| C. 사용자 체감 수정 (noticeable fixes) | 12 |
| D. 문서 카드로 이미 착지 (t47/t59/t32/t98/t115/cc234) | 14 |
| E. 내부 전용 (internal-only: SPEC 증거·테스트·CI·보고서) | 95 |

CHANGELOG `[Unreleased]`에는 현재 SPEC 3-phase close 3건(HOME-PATHS,
CLEAN-HOME, ALWAYS-LOADED-DIET)만 기록돼 있고, 아래 A/B/C 대부분은
릴리즈 노트에 아직 없다 — 릴리즈 노트 생성은 release 프로세스 소관(본
감사 범위 밖, 참고로만 기록).

## A. 사용자 체감 신기능 — 커버리지 매트릭스

문서화 열: ✓ = 반영됨 / △ = 부분 반영(낡은 서술 공존) / ✗ = 미반영.

| # | 기능 (커밋) | README×4 | docs-site ko |
|---|---|---|---|
| F1 | **팩토리 모드 `moai cc -k <N>`** — 리드 루프(큐 폴링·빈 슬롯 픽·스태거 팬아웃, 워커 기본 8), `moai cc -k <N> --name worker-<i>` 진입, A/B클래스 통째·C클래스 직렬 3단계 라우팅, `-f`/`--factory` 명시적 에러 (`b56f5dbc2`, `59dd62d02`, `a34494665`, `c5816c927`, `4419a3744`) | ✗ | ✗ `advanced/kanban-mode.md`·`cli-reference/launchers.md` 전무 |
| F2 | 칸반 동반 세션 이름 = bare role + 생존 충돌 시 숫자 bump (plan-1, plan-2…) (`4419a3744`) | ✓ (README.ko L63) | △ `kanban-mode.md` L128 "합류한 역할은 안내에 없고" 낡은 서술 |
| F3 | **`moai graph build` / `moai graph query`** — edges.jsonl writer+reader, `--callers` `--blast` `--fanin` `--specs-no-code` `--milestones-no-card` (`557877c49`, `02d9cac05`, `c5eed2456`) | ✗ | ✗ cli-reference 페이지 없음 |
| F4 | **`moai tokens record`** — 풀별(glm/claude/other)·origin별(메인/사이드체인) 토큰 회계 → `.moai/state/token-accounting.jsonl` (`dd060a191`) | ✗ | ✗ |
| F5 | **`moai memory doctor` / `moai memory archive`** — MEMORY.md↔토픽 파일 불일치·개수 상한 진단, 아카이브 (`504797021`) | ✗ | ✗ |
| F6 | **glm_task MCP 도구군 4개** (`glm_task`, `glm_job_status/result/cancel`) + MCP 카탈로그 17→**21개·여섯 그룹** (`9865e87ed`, `d7f9f3b3a`) | ✗ "17개·다섯 그룹" (README.ko L277-285) | ✗ `guides/mcp-server.md` "17-도구 카탈로그" 전면 |
| F7 | **`/moai review` 판정 단계에 audit_multi 수렴 연결** (`0ba4fa466`) + codex 리뷰 실패 턴 노출(`f389295ba`) | ✗ | ✗ `advanced/multi-model-audit.md`·`utility-commands/moai-review.md` 무언급 |
| F8 | **`moai clean --home`** (~/.moai 정리, allowlist-only, dry-run 기본, `state.home_retention_days`) + doctor **Home Disk Usage** 진단 + `MOAI_HOME` 환경변수 (SPEC-CLEAN-HOME + SPEC-HOME-PATHS M1-M6: `6ec0ad212`, `051a2fa94`, `6600bfd8e` 등) | ✗ | ✗ `utility-commands/moai-clean.md`·`cli-reference/doctor.md` 무언급 |
| F9 | **todo CLI 확장** — bare `moai todo` = 목록, 2단어 이상 자연어 fallthrough→add, `add --pick` / `next <n>` 경쟁 안전 잠금쓰기 / **`unpick`** 동사, 워크트리에서 실행해도 **primary 체크아웃 큐로 귀속** (`504797021`, `3f5705d8d`, `b018bf1e9`, `6ba8ef90e`, `f5297037f`) | △ README.ko L310 `<add\|list\|next\|done>`에 unpick/bare 없음 | △ `moai-todo.md` CLI 표에 unpick·--pick·bare·primary 귀속 없음 |
| F10 | **statusline 개편** — 세션 라인(🏷️ 세션명 · 👤 에이전트 · 🔄 `TODO: picked/queued` · 🔀 issues/PRs) 조건부 4번째 줄로 추가, 리포 표시 📡, git 상태 📫 사서함 계열, repo|branch 조인 ASCII 파이프(`1492e6a37`, `062e05017`, `572c3b732`, `caf435ec4`) | ✗ README.ko L443-460 예시·표 낡음 | ✗ `advanced/statusline.md` "3-line" 서술 + 예시 전부 낡음 |
| F11 | **웹콘솔 Cross-Session(포스처) 설정 탭** — inbound·isolate_machines·dialog_expiry 편집, 설정 탭 10→**11개** (`aa4387d83`) + 콘솔 크롬 i18n(`a2adaaba9`, `470c6af40`) | ✗ README.ko L670 "10개 탭" | ✗ `moai-web-console.md` L25 "9개 탭" (셋 다 불일치) |
| F12 | **Agent Teams 실험적 재허용 + 기본 탑재** — RETIRED→experimental 전환, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`이 배포 템플릿 settings에 기본 포함(`94bc55370`) | ✗ (README 언급 없음 — 구조상 불필요) | ✗ `claude-code/agentic/agent-teams.md` L62 "기본적으로 꺼져 있습니다…직접 설정" STALE |
| F13 | **`/loop` 포어맨 드라이버** — bare `/loop`가 칸반 백로그를 스스로 굴리는 드라이버(loop.md + moai-kanban-foreman 스킬 배포) (`5178e0b3c`) | ✗ | ✗ |
| F14 | **GLM 1M 컨텍스트 선언** — `CLAUDE_CODE_MAX_CONTEXT_TOKENS`로 Claude Code에 1M 창 선언, 4표면 전부 1,000,000 보고(`225a51e24`) | △ 매핑 표 1M는 맞음, env 무언급 | △ `statusline.md` GLM 보정 절은 있으나 env 무언급, `multi-llm/` 전무 |
| F15 | **`moai update` 삭제 전 백업** — 관리 뿌리 정리 전 미관리 파일을 `.moai-backups/<ts>/pre-clean/`에 백업, 백업 실패 시 중단(`aefaddb71`) + 조용한 실패 노출(`44a85a364`) + retained-key TUI 안내(`14a3a4c0a`) | ✗ | ✗ `cli-reference/update.md` config 백업만 언급 |
| F16 | ast-grep 배포 기본 룰 6종 승격 + `gate.yaml` 규칙 디렉터리 SSOT 단일화(`77922a36b`, `af5c7160d`, `525247ecc`) | ✗ | ✗ `cli-reference/ast-grep.md` 미확인(아래 Gaps G3) |
| F17 | **Windows 설치 스크립트** — `install.bat` 출시(cmd.exe용 CRLF + `/releases/latest` 리다이렉트로 최신 버전 확인), install.ps1/install.sh가 rate-limit 걸린 REST API 없이 최신 버전 확인(`2febbbac6`, `4d97f622c`, `11df56ab1`, `2ce06b439`) | ✗ | ✗ `getting-started/installation.md` 무언급 |
| F18 | **WorktreeCreate active-creator 계약** — `isolation: "worktree"` 스폰이 hook 거부로 실패하던 것 수정, 세션 상태 파일에 activeCreator 기록(`fa667b067`) | ✗ | 미확인(Gaps G4) |
| F19 | 칸반 5-열 보드(review 칸 은퇴 → sync 게이트가 판정 흡수) + 동반 3역할 체계(`327c6453d`, `c5816c927`) | ✓ t47+t115 착지 | ✓ |
| F20 | 교차세션 메시징 가용성 제약(OS·프로바이더·버전·플래그) 문서화(`ce7fb16cb`) | △ | △ `kanban-mode.md` L190 "자동으로 허용돼 있습니다" — 제약과 병존 서술 필요 |
| F21 | 오케스트레이션 모드 카탈로그 6→4 (spawn-count 축 재명명) (`ebf95526b`, `4c5994237`) | ✓ t47 반영 | ✓ t98 착지 |
| F22 | GLM per-agent 모델 → 세션 inherit 접힘 + effort-reasoning 맵 노출(`62c2bf939`) | △ | 미확인(Gaps G5 — `multi-llm/model-policy.md`) |

## B. 사용자 체감 동작 변경

| # | 변경 (커밋) | README | docs-site ko |
|---|---|---|---|
| B1 | `-f`/`--factory` 토큰 = 명시적 에러(은퇴) — F1에 포함 | ✗ | ✗ |
| B2 | `moai mcp` 루트 등록 + 등록 가드(`f7de9eb85`) | — | 미확인(G3) |
| B3 | GLM per-agent fold (F22와 동일) | △ | 미확인 |
| B4 | kanban 알림: epic 포인터 → todo 백로그 요약(`194eed3a0`) | — | △ hooks/kanban 문서 |
| B5 | 세션 레지스트리 CwdChanged 재배치 + `cc -w` 레인 앵커 감지(`bf2201e5f`) | — | 내부 동작, 문서 불필요 판정 |
| B6 | branch-guard가 읽기 전용 `git branch`(조회) 거부 안 함 + 거부 사유에서 도달 불가능한 경로 제거(`2df70172f`, `ee2a3dfb2`) | — | △ `advanced/hooks-reference.md` 미확인(G3) |
| B7 | hook 래퍼 `.tmpl`↔`.sh` 쌍 동기화 + 드리프트 가드(`b06cec203`) | — | 내부 |
| B8 | worktree `done` 잠금 안내·stderr 노출, 앵커 가드 확장(remove/clean/PR-merge) (`901e5244f`, `af749fafe`, `08dc10f20`) | — | △ `worktree/faq.md` 미확인(G3) |

## C. 사용자 체감 수정 (대표)

| # | 수정 (커밋) | 문서 영향 |
|---|---|---|
| C1 | `moai update` rate-limit 헤더 노출 + 릴리스 확인에 인증 토큰(`d6b80a01c`) | update.md 간단 언급 가능(선택) |
| C2 | 프로필 런치 레저를 경로 표기가 아닌 디렉터리로 매칭(`e63551b7d`) | 문서 불필요 |
| C3 | 설정 캐시 미초기 디렉터리 스킵(`3f0077d60`) | 문서 불필요 |
| C4 | codex 리뷰 실패 턴 노출(거짓 PASS 방지, `f389295ba`) | F7과 함께 multi-model-audit에 반영 가치 |
| C5 | Windows 경로 휴대성 테스트 픽스(`44ea01575` 등) | 내부 |
| C6 | GLM 429/과금 경로 등 기타 — 이슈 #1571/#1572/#1574/#1575/#1576/#1562/#1561/#1560/#1559/#1557/#1563/#1564/#1565 (main 병합분) | README #1561/#1562 착지 완료 |

## D. 문서 카드로 이미 착지 — 재작업 금지

t47(README ko 정본 승격·4-로케일 재유도: `3fa31ef4d`, `3c887d508`, `c07804841`,
`26935922e`), t59(용어 통일: `bd06ea127`, `548ca9504`), t32(docs-site 8페이지
KO-SSOT v3.1 + 4-로케일 유도: `cdc1e4c6f`), t98(오케스트레이션 모드·툴 필터·카드
클래스: `d3ddd3406`), t115(5-열 보드: `99b1e5638`, `a916c8137`, `72cc1b220`),
cc234-align(CC 2.1.234 teammate-model 문서: `1c341dd80`).

## 정확성 검증 근거 (소스 대조)

- MCP 도구 21개: `internal/cli/mcp_server.go` — `add("...")` 리터럴 20건 +
  `add(auditMultiToolName, ...)`(mcp_audit_multi.go 상수) 1건 = 21. 그룹은 6개
  (SPEC 3 + 검증 2 + 골/세션 3 + 교차모델감사 4 + codex 위임 5 + **GLM 위임 4**).
- 팩토리 진입 계약: `internal/cli/cc.go` L20/L44-L96 도움말 + `factory.go`
  (v1.2.0 genealogy, `FACTORY_MODE_UNSUPPORTED_BACKEND`, `rejectRetiredFactoryFlag`).
- todo 동사: `internal/cli/todo.go` L162-L218 (bare=list, 2단어 fallthrough,
  `unpick`, `--pick`, primary-checkout 귀속).
- statusline: `internal/statusline/renderer.go` L122-L146(4줄 조립, 세션 라인
  마지막 조건부), L168-L199(세션 라인 세그먼트), L375/L514(`📡` repo, ASCII `|`).
- 웹콘솔 설정 탭 11개: `internal/web/schemaform.go` `consoleTabs()` — MCP와
  crosssession 탭 포함.
- Agent Teams 기본 탑재: `internal/template/templates/.claude/settings.json.tmpl` L386.
- 보드 저장소 "호출자 없음" 서술 여전히 사실: `LoadBoard(` 프로덕션 호출자
  0건 (internal/kanban 외부, 테스트 제외 grep).

## 커버리지 집계

- A그룹 22항목 중 완전 문서화(✓): **3** (F2, F19, F21) + 부분(△) 5 →
  미반영(✗) **14**.
- 최우선 갭(F1 팩토리, F3 graph, F6 MCP 21도구, F10 statusline, F8 clean --home,
  F15 update 백업)은 모두 "사용자가 직접 치는 명령/보는 화면"이라 노출도가
  가장 크다.
