# init-tui-audit-20260909.md 발췌 — t583 증거용 사본

원본: primary checkout `.moai/reports/init-tui-audit-20260909.md` (미추적 파일, 읽기만 함). 아래는 원문 행을 그대로 옮긴 것이며, 원본의 file:line 인용은 2026-09-09 트리 기준이라 현재 트리와 행번호가 어긋날 수 있다(재측정 결과는 verdict.md).

## L12-31 — init 16질문 행선지

| # | 질문 | 기본값 | 행선지 | 안 물으면 |
|---|---|---|---|---|
| 1 | 대화 언어 | en | language.yaml 렌더 | 영어 |
| 2 | 이름 | 빈 값 | user.yaml user.name | 빈 값 |
| 3 | 프로젝트 이름 | 디렉터명 | project.yaml + 렌더 | 디렉터명(위치 인자 대체) |
| 4 | 모델 정책 | medium | llm.yaml profile+performance_tier | medium |
| 5 | 리포트 형식 | html+md | report.yaml | html+md |
| 6 | 프로젝트 모드 | personal | project.yaml mode (Go 소비자 0 — F8) | personal |
| 7 | 워크트리 자동 생성 | false | workflow.worktree.auto_create — "예"가 안 쓰임(F1) | false |
| 8 | 백로그 큐 | true | workflow.todo.enabled | true |
| 9 | 피드백 자동 제출 | false | feedback.auto_submit | false |
| 10 | 감사 모델 | claude | workflow.audit.model | claude |
| 11-13 | 게이트 3종 | required/required/advisory | workflow.audit.gates.* | 배포 기본값 |
| 14 | codex 검토 게이트 훅 | false | workflow.codex.review_gate.enabled (훅 등록은 settings.json 무조건, 키로 판정 — mcp_codex.go:1595) | false |
| 15 | MCP 프로비저닝 | true | .mcp.json | **생략 경로에선 설치 안 됨(F4)** |
| 16 | 자율성 등급 | semi-auto | ~/.claude/settings.json defaultMode (+MOAI_AUTONOMY_TIER가 런타임 최종) | semi-auto |

우선순위: CLI 플래그 > 위저드 > 기존 파일 > 컴파일 기본값 (Flags().Changed 판정, init.go:232-288). 로더는 부분 재정의 계약(키 없으면 defaults.go 시딩값, loader.go:36).

## L50-53 — 결함 전수 중 이 카드 소관

🔴 F1 워크트리 "예" drop(추적자 플래그 전용 — init_workflow_flags.go:41-44, 테스트가 false만 재어 미탐지)
🟡 F4 --non-interactive MCP 미설치(init.go:542-567 시드+911)

## L56-60 — 개선 제안

- **남길 3개**: 대화 언어(첫 경험 직결) · 이름(기본값 없음) · 자율성 등급(권한 직결)
- **빼는 13개**: 프로젝트 이름(인자 없을 때만 질문) · 모델정책·리포트·todo·feedback·감사모델·게이트3·codex훅(전부 web 기존 표면, 기본값 유지) · 프로젝트 모드(F8과 함께 폐기) · 워크트리(F1 수리 후 web) · MCP(질문 제거, 기본 on + `--no-mcp` 신설, F4 수리 포함)
- **묶음**: ① init 정온화+F1·F4 수리+"미설정=기본값" 실행 검증 테스트(생략 경로를 실제로 돌려야 F1 재발을 막음) ...

## L63-66 — 운영자 결정 반영 (2026-09-09 확정)

1. **자율 모드 재정의**: `bypassPermissions` / `auto mode` / `accept edits on` 3선택, **기본 = accept edits on**. 구현 전제: REQ-007 zero-delta 재정의(새 기본은 기록을 씀), REQ-006+샌드박스 게이트는 bypass 유지, `defaultMode` 토큰 유효값 검증(현재 코드는 "auto" — autonomy_bundle.go:84).
2. **하네스 질문 추가**(유지 4번째): claude 단독=현행 / both=`.claude`+codex 전체+CLAUDE.md=`@AGENTS.md` 래퍼+AGENTS.md 범용 지침 / codex 단독=AGENTS.md+codex만.

## L83-92 — 최종 카드 분할

| # | 카드 | 의존 |
|---|---|---|
| C1 | init 정온화 — 16→4 질문, 13개 제거+기본값 위임, F1·F4 수리, 생략 경로 실행 테스트 | — |
| C2 | 자율 모드 재정의 — 3 실제 모드·기본 acceptEdits·불변식 재검토·토큰 검증 | C1 병행/선행 |
| C3 | 하네스 3-way 배포 — 질문 신설+both(commandemit 착지·AGENTS.md 보강)→codex-only(필터·캐노닉·정합·--agent 판정) | C1 프레임 후 |
| C4 | TUX 렌더+i18n (F10·F11) | 독립 |
| C5 | update 수리 (F2·F3·F5·F6·F7) | 독립 |
| C6 | 정합성·문서 (F8·F9·F12-14·F16·F17; F15·F17 수용 판정 포함) | 독립 |
