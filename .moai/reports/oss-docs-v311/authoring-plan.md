# v3.1.1 문서 작성 계획 (authoring plan)

- 원본: `release/v3.1.1` @ `1bd9140eb` (읽기 전용) — 드래프트는
  `.moai/reports/oss-docs-v311/drafts/`에 전체 파일 사본.
- 규칙: ko 정본 작성 → en/ja/zh는 locale-translator가 같은 PR에서 파생.
  i18n HARD 규칙(TD-only mermaid, 본문 이모지 금지→icon shortcode,
  adk.mo.ai.kr URL 화이트리스트, 강조 표기 간격) 드래프트에도 적용.

## Stage 1 — ko 드래프트 (이번 phase, 13파일)

### 1-A. 기존 페이지 편집 (10)

| # | 파일 | 변경 요약 |
|---|---|---|
| E1 | `README.ko.md` | MCP 17→21·여섯 그룹(GLM 위임 행 추가) / statusline 예시·표 4줄화+아이콘 갱신 / 칸반 절에 팩토리 모드 H3 추가 / todo 동사 갱신 / CLI 표에 `moai graph` 행 / 유틸리티 나열에 `todo` / 웹콘손 11탭 / update 설명에 "삭제 전 백업" |
| E2 | `docs-site/content/ko/advanced/kanban-mode.md` | "팩토리 모드 — 번호 붙은 워커" 절 신설(진입 계약·리드 루프·카드 라우팅·-f 은퇴) / L128 안내 서술 갱신 / 관련문서에 `/moai loop` 추가 |
| E3 | `docs-site/content/ko/cli-reference/launchers.md` | `moai cc` 플래그 표에 `-k` 4형태 + `--name` 추가, 팩토리 짧은 절 |
| E4 | `docs-site/content/ko/guides/mcp-server.md` | 17→21·여섯 그룹 전면 교체, "GLM 위임" 그룹 절 추가, 백그라운드 추적 절에 GLM 병기, description 수정 |
| E5 | `docs-site/content/ko/advanced/statusline.md` | 3-line→"3줄 + 조건부 세션 라인(4번째)", 예시 전면 갱신(📡·📫·ASCII 파이프·세션 라인), 세그먼트 키 서술 보강(16 정식 키 + session/backlog/github), MAX_CONTEXT_TOKENS 한 문장 |
| E6 | `docs-site/content/ko/utility-commands/moai-todo.md` | CLI 표에 unpick·`add --pick`·bare=list·자연어 fallthrough, primary 체크아웃 귀속 문단 |
| E7 | `docs-site/content/ko/utility-commands/moai-clean.md` | "홈 디렉터리 정리(`moai clean --home`)" 절: allowlist·dry-run 기본·home_retention_days·가드된 force·doctor Home Disk Usage 링크 |
| E8 | `docs-site/content/ko/claude-code/agentic/agent-teams.md` | "기본 꺼짐" → "배포 기본 켜짐(실험적 지위 유지, 끌 수 있음)" 정정 |
| E9 | `docs-site/content/ko/cli-reference/update.md` | 관리 뿌리 정리 전 미관리 파일 사전 백업(`.moai-backups/<ts>/pre-clean/`, 실패 시 중단) 절 |
| E10 | `docs-site/content/ko/advanced/moai-web-console.md` | 설정 탭 9→11(Cross-Session 포함 나열 + 한 문단) |

### 1-B. 신규 페이지 (3) — 콘텐츠 드래프트, 메뉴 등록은 curator

| # | 파일 | 내용 |
|---|---|---|
| N1 | `docs-site/content/ko/cli-reference/graph.md` | `moai graph build`(edges.jsonl 산출) + `moai graph query` 5 셀렉터, "미연결 ≠ 미구현"/"큐 없음 ≠ 미발급" 주의문 |
| N2 | `docs-site/content/ko/cli-reference/tokens.md` | `moai tokens record` — 풀/origin 회계, token-accounting.jsonl, 플래그 |
| N3 | `docs-site/content/ko/cli-reference/memory.md` | `moai memory doctor/archive` — 색인↔파일 불일치, 상한, 아카이브 |

## Stage 2 — 소형 변경 (다음 phase, 문안만 제시 — 드래프트 미생성)

| 파일 | 변경 |
|---|---|
| `cli-reference/doctor.md` | 진단 목록에 "Home Disk Usage"(권고, 임계값 기본 컴파일 상수) 1행 |
| `getting-started/installation.md` | Windows 절에 install.bat 언급 + "cmd.exe는 실행 환경으로 미지원, 설치 스크립트는 지원" 구분 1-2문장 |
| `advanced/multi-model-audit.md` | "`/moai review` 판정 단계가 audit_multi 수렴을 참조(0ba4fa466) + codex 리뷰 실패 턴 노출" 문단 |
| `utility-commands/moai-review.md` | verdict 단계 audit_multi 연결 1-2문장 |
| `multi-llm/model-policy.md` | GLM 세션 1M 선언(CLAUDE_CODE_MAX_CONTEXT_TOKENS) — F22 확인 후(Gaps G5) |

## Stage 3 — 4-로케일 전파 (locale-translator 인계 목록)

동일 PR에서 en/ja/zh 파생. README는 이제 **ko가 정본**(t47 승격) —
ko → en/ja/zh 순서로 파생:

1. `README.md` / `README.ja.md` / `README.zh.md` — E1 동일 섹션 번역
2. `docs-site/content/{en,ja,zh}/advanced/kanban-mode.md` — E2
3. `docs-site/content/{en,ja,zh}/cli-reference/launchers.md` — E3
4. `docs-site/content/{en,ja,zh}/guides/mcp-server.md` — E4
5. `docs-site/content/{en,ja,zh}/advanced/statusline.md` — E5
6. `docs-site/content/{en,ja,zh}/utility-commands/moai-todo.md` — E6
7. `docs-site/content/{en,ja,zh}/utility-commands/moai-clean.md` — E7
8. `docs-site/content/{en,ja,zh}/claude-code/agentic/agent-teams.md` — E8
9. `docs-site/content/{en,ja,zh}/cli-reference/update.md` — E9
10. `docs-site/content/{en,ja,zh}/advanced/moai-web-console.md` — E10
11. 신규 3페이지 × 3로케일 (N1-N3) — curator가 파일 자리 만들면 번역

## 구조 변경 (structure-curator 인계 목록)

| 항목 | 내용 |
|---|---|
| C1 | `content/<locale>/_meta.yaml` cli-reference에 `graph`·`tokens`·`memory` 항목 추가(weight: launchers 인근 tools 그룹과 정렬) — 4로케일 전부 |
| C2 | 신규 3페이지의 메뉴 노출 — `data/menu/main.yaml` 수정 필요 여부 확인(geekdoc _meta 기반이면 불필요할 수 있음 — curator 판단) |
| C3 | 리다이렉트: 불필요(페이지 이동/개명 없음) — `vercel.json` 변경 없음 |

## 파일 수 추정

- ko 드래프트: 13 (편집 10 + 신규 3)
- 4-로케일 착지 시: ko 13 + en/ja/zh 각 13 = **52파일** (README 4 포함:
  README 4 + docs-site 48)
- Stage 2 포함 시 +20 (5파일 × 4로케일)

## 결정 지점 (운영자 판단 필요)

| # | 질문 | 권장 |
|---|---|---|
| D1 | README "v3.1 새 기능" 절 제목/서술을 v3.1.1용으로 갱신할까? (현재 "v3.1은 광복절인 8월 15일에 맞춰 내놓는다" = v3.1.0 서술) | 절 제목 유지 + 팩토리 모드 H3만 추가(드래프트 방식). 절 전체를 v3.1.1 기준으로 다시 쓰는 건 릴리즈 노트와 중복 |
| D2 | "전체 36개 커맨드" 숫자 — v3.1.1 정확 커맨드 수 확정 후 갱신? | 릴리즈 시점 `moai --help`로 확정. 드래프트는 graph 행 추가 + 숫자는 유지(확정 전 임의 변경 금지) |
| D3 | 신규 3페이지(cli-reference graph/tokens/memory)를 이번 릴리즈에 넣을까, 별도 카드로? | 이번에 포함(팩토리·그래프는 v3.1.1 대표 기능). 단 curator 일정이 빠듯하면 tokens/memory만 별도 카드로 분리 가능 |
| D4 | `install.bat` 안내 수준 — README까지 갈까? | docs-site installation.md만(Stage 2). README는 PowerShell 안내 유지 |
| D5 | CHANGELOG [Unreleased]에 A그룹 항목 기록 — 릴리즈 노트 프로세스 소관이지만 본 harness가 초안을 낼까? | 별도 카드 권장(릴리즈 노트는 release-update harness 영역) |

## 검증 계획 (착지 시)

hns-oss-docs-verify 인라인 체크: hugo build 무경고, URL 블랙리스트 grep,
mermaid LR/RL 부재, 4-로케일 파일존재+섹션수 패리티, README 4파일 H2
패리티, 본문 이모지 스캔. 드래프트 단계에서는 원칙 사전 준수 + grep
셀프체크(아래 gaps.md 참조).
