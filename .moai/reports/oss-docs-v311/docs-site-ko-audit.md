# docs-site ko 감사 — v3.1.1 델타

- 대상: `docs-site/content/ko/**` (로케일당 약 130페이지) — en/ja/zh는 같은
  페이지 집합이므로 동일 갭 적용(4-로케일 전파 목록은 authoring-plan.md).
- 방법: 커버리지 매트릭스(commit-analysis.md) 키워드 grep + 핵심 페이지
  전수 정독 + 소스 코드 대조.

## 1. 페이지별 감사 결과

### 1-1. `advanced/kanban-mode.md` (267줄) — 팩토리 모드 전무 [최우선]

- ✓ 최신: 5-열 보드(L142), sync 게이트 판정 흡수(L20/L187), 카드 클래스
  A/B/C(L225-239), `moai chain` 5동사(L130), `moai cg` 거부(L127/L257),
  "직접 읽은 증거로만"(L187) — t115/t98 착지분 무손상.
- ✗ **팩토리 모드(`-k <N>`)가 한 줄도 없음**: L154-168 진입 예시, L127
  런처 스위치 설명, L181 "네 터미널로 굴리는 한 런" 모두 3역할 체계만.
  추가 필요: 진입 계약(`moai cc -k <N>` 리드 / `-k <N> --name worker-<i>`
  워커 / bare `--name worker-<i>` = 기본 8), 리드 루프(큐 폴링·빈 슬롯
  픽·스태거 팬아웃), A/B 통째·C 직렬 3단계 라우팅, `-f` 은퇴,
  `FACTORY_MODE_UNSUPPORTED_BACKEND`.
- △ L128 "합류한 역할은 안내에 없고 세션 레코드에 따로 기록됩니다" —
  bare-role 이름 + 숫자 bump 체계(4419a3744)와 어긋나는 낡은 서술.
- △ L142 "보드를 조회하거나 카드를 옮기는 CLI 동사는 존재하지 않습니다" —
  **사실 유지 확인**(`LoadBoard` 프로덕션 호출자 0건, grep 2026-08-18).
  단 `moai todo` 동사가 늘어난 것과 별개로 이 문장은 보드(5-열 상태 파일)에
  한정된 서술이므로 유지 가능. "카드를 고르는 동사" 언급은 todo 절과 정합.
- △ L190 "교차 세션 메시징은 주입된 `--settings`를 통해 자동으로 허용돼
  있습니다" — 가용성 제약(OS·프로바이더·플래그, ce7fb16cb)이 배치에
  문서화됐는데 이 문장은 무조건 허용처럼 읽힘. 한 문장 보강 권장.
- ✗ `/loop` 포어맨(5178e0b3c) 무언급 — "관련 문서"에 loop 링크 추가 또는
  별도 문단(드래프트에는 관련문서 링크로 반영).

### 1-2. `cli-reference/launchers.md` (107줄) — `-k` 계열 플래그 누락

- ✗ L25-33 `moai cc` 플래그 표에 `-k/--kanban`, `--name`이 아예 없음
  (칸반조차!). v3.1 대표 진입 표면이 런처 레퍼런스에 빠져 있음.
- 표에 추가: `-k, --kanban [SPEC-ID]`(칸반 리드), `-k --name <role>`(칸반
  동반 — plan/run/sync, 생존 충돌 시 숫자 bump), `-k <N>`(팩토리 리드),
  `-k <N> --name worker-<i>`(팩토리 워커). `-f` 은퇴 주석.
- `-c, --continue` 등 기존 플래그는 소스와 대조 필요(Gaps G7).

### 1-3. `guides/mcp-server.md` (221줄) — 17→21 [최우선]

- ✗ L5 description "17-도구 카탈로그", L10/L32/L39/L96-98 "17개·다섯 그룹" —
  실제 **21개·여섯 그룹**.
- ✗ "GLM 위임 (백그라운드 작업)" 그룹 절 전체 없음: `glm_task`(동기/백그래운드,
  `DefaultGLMTaskTimeout` 상한), `glm_job_status/result/cancel`. codex 위임
  절(L140-150)과 대칭 구조로 추가.
- △ L180-192 "백그라운드 작업 진행 추적" — codex만 다룸 → GLM 병기.
- 표지 description(front-matter)도 21로.

### 1-4. `advanced/statusline.md` (158줄) — 레이아웃 개편 미반영

- ✗ L2 title "3-line 레이아웃" + L17-29 "한눈에 보는 3줄" + L21-24 예시 —
  v3.1.1은 **조건부 4번째 줄(세션 라인)** 추가: `🏷️ 세션명 · 👤 에이전트 ·
  🔄 TODO: picked/queued · 🔀 이슈/PR`. 첫 줄은 정보줄, 세션 라인은 "마지막"
  줄(renderer.go L122-146 조립 순서).
- ✗ 셋째 줄 예시: `📡 owner/name | 🅱️ branch` (repo 아이콘 🔀→📡, 조인 문자
  `│`→ASCII `|`), git 상태 `💾`→**📫 사서함 계열**(📬 staged/📫 modified/
  📪 untracked/📭 clean).
- ✗ L129-138 설정 스키마 — `session`(🏷️👤)·`backlog`(🔄)·`github`(🔀)
  세그먼트 키 추가 여부를 소스에서 확정해 반영해야 함(Gaps G8 — 드래프트는
  확정된 부분만 반영).
- ✓ GLM CW 보정 절(L99-101)은 유지 — 단 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`
  선언(225a51e24) 한 문장 보강 가치.

### 1-5. `utility-commands/moai-todo.md` (152줄) — CLI 동사 확장 미반영

- △ L112-138 CLI 표: `unpick`, `add --pick`, bare `moai todo`(=list),
  2단어 이상 자연어 fallthrough, **워크트리에서 실행해도 primary 체크아웃
  큐로 귀속**(6ba8ef90e) 미기술. L56 상태 파일 절에도 primary 귀속 무언급.
- ✓ 슬래시 표면(L37-52)과 "고르는 주체는 사람" 원칙은 그대로 유효.

### 1-6. `utility-commands/moai-clean.md` (211줄) — `--home` 전무

- ✗ `/moai clean`(데드 코드)만 다룸. `moai clean --home`(터미널 CLI,
  ~/.moai 정리: allowlist-only·dry-run 기본·`state.home_retention_days`·
  가드된 force)가 SPEC-CLEAN-HOME으로 들어왔는데 무언급.
- 권장: "CLI 표면" 절 신설(또는 하단 별도 H2) — `/moai clean`과
  `moai clean [--home]`은 다른 표면이라는 구분 포함. doctor의
  Home Disk Usage 진단과 링크.

### 1-7. `advanced/moai-web-console.md` (178줄) — 탭 수·포스처

- ✗ L25 "설정(Settings) … (9개 탭)" — 실제 **11개**(MCP, Cross-Session 포함).
  README는 10개라고 하고 있어 셋 다 다름. 11개로 통일 + Cross-Session
  탭 설명(inbound/isolate_machines/dialog_expiry, `crosssession.yaml`,
  다음 런치부터 적용) 추가.

### 1-8. `claude-code/agentic/agent-teams.md` — 기본 탑재 STALE

- ✗ L62 "기본적으로 꺼져 있습니다… 환경 변수 … 1로 설정해야 켤 수 있습니다" —
  v3.1.1 배포 템플릿 settings에 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`이
  **기본 포함**(settings.json.tmpl L386). "배포 기본은 켜짐(실험적 지위 유지,
  끌 수 있음)"으로 정정.

### 1-9. `cli-reference/update.md` — 삭제 전 백업 미반영

- △ L86-98 config 백업만 언급. 관리 뿌리 정리 전 **미관리 파일 사전 백업**
  (`.moai-backups/<timestamp>/pre-clean/`, 백업 실패 시 중단 — aefaddb71)과
  retained-key TUI 안내(14a3a4c0a) 추가. "moai update가 로컬 전용 파일을
  지운다"는 과거 사고 서술이 있다면 백업 안내장치로 대체.

### 1-10. cli-reference 신규 페이지 부재 [구조 — curator 협력]

- ✗ `moai graph`(build/query) — 페이지 없음. 내용: edges.jsonl 산출물 개념,
  5개 셀렉터(`--callers` `--blast` `--fanin` `--specs-no-code`
  `--milestones-no-card`), "미연결 ≠ 미구현"·"큐 없음 ≠ 미발급" 주의문
  (graph.go에 상수화된 caveat 문구 그대로).
- ✗ `moai tokens record` — 페이지 없음. 풀/origin 회계 개념,
  `.moai/state/token-accounting.jsonl`, `--transcript/--session/--card/--role/--json`.
- ✗ `moai memory doctor/archive` — 페이지 없음. MEMORY.md 색인↔파일 불일치
  진단, 50개 상한, 아카이브(삭제 아님).
- 이들 3페이지는 **신규 콘텐츠 파일**이라 드래프트 작성 가능하나,
  `_meta.yaml`/메뉴 등록은 structure-curator 소관 → handoff 목록으로 이관.

### 1-11. 그 외 스팟 확인

- `getting-started/installation.md` L43 "cmd.exe는 지원하지 않습니다"와
  install.bat 출시(2febbbac6)의 병존 설명 — "cmd.exe는 moai 실행 환경으로
  지원하지 않되, 설치 스크립트(install.bat)는 cmd.exe에서도 돈다" 식의
  한 문장 정리 권장(Phase 2 소형 변경).
- `advanced/multi-model-audit.md` — 수렴 메커니즘은 충실하나 `/moai review`
  판정 단계 연결(0ba4fa466) 무언급 + codex 리뷰 실패 턴 노출(f389295ba).
  한 문단 추가(Phase 2 소형).
- `cli-reference/doctor.md` — Home Disk Usage 진단 행 추가(Phase 2 소형).
- `changelog/_index.md` — 릴리즈 시점 스탬프 소관, 감사 대상 아님.

## 2. 구조 파인딩 (structure-curator 영역)

| 항목 | 내용 | 소관 |
|---|---|---|
| S1 | `cli-reference/`에 graph/tokens/memory 3페이지 신규 등록(`_meta.yaml` weight — launchers 근처 tools 그룹) | curator |
| S2 | 신규 3페이지의 en/ja/zh 사본 등록(4-로케일 파일 존재 검사 통과) | curator + locale-translator |
| S3 | kanban-mode.md에 팩토리 절 추가는 기존 페이지라 메뉴 변경 불요 | — (변경 없음) |
| S4 | 리다이렉트 불필요(페이지 이동/개명 없음) | — |

## 3. 감사 제외 (이미 최신 확인)

`_index.md`(5-열 보드 언급 확인), `core-concepts/kanban-board-terms.md`
(t59 용어 카드 착지), `utility-commands/moai-review.md`·`moai-loop.md`의
기존 내용(변경분은 Phase 2 소형으로만), t32의 8개 v3.1 페이지(로케일
패리티 착지 상태 — 내용 갱신은 본 감사의 갭 항목이 있을 때만).
