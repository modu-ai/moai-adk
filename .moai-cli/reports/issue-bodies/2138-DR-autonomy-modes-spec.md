# [DR][02] 자율 무중단 개발 모드 설계 — `/moai --leader-mode` · `/moai --agent-mode` (SPEC 단계)

## 배경
종합 보고서 §10.4(사용자 지시 2026-11-09): 론처(`moai cc|glm|codex`)는 역할을 모르고, 세션 안에서 `/moai --leader-mode [--autonomy full|semi]` / `/moai --agent-mode` 로 자율 무중단 개발에 들어간다. 에이전트는 **self-claim**(로컬 미배차 목록 갱신 → journal+lock → Issue 댓글 claim → 최소 live 댓글 id 승 → label/상태 projection → EnterWorktree → … → ExitWorktree(keep) → 반복), 리더는 배차가 아닌 감독(5분 PR 점검·타 레인 수리 지시·배정 기록 대조·목록 보충·`/goal` 붙여넣기 지시문 안내). Class C(설계 변경) — 이 카드는 **SPEC 문서만** 만들고 구현은 후속 카드.

## 범위
- `docs/design-review/autonomy-modes-spec.md` 신규: 진입 방식, 모드별 책임, self-claim 7단계(각 단계의 실패 분기 포함), 리더 감독 루프(5분 주기는 사용자 지정값·측정값 아님 명시), 결정 등급 full/semi의 keep-set 경계, `/goal` 안내 절차(Claude Code 내장 `/goal`은 모델이 입력 불가 → 붙여넣기 지시문; MoAI `goal_arm`은 Go 세대·RC1 route 없음; Codex 동등 명령 미확인), 론처 `-l`/`-a` 존치 여부를 결정 항목으로
- RC1 실물 대조 표: `src/claim.rs`(①journal+lock, ⑤댓글 claim), `src/handoff.rs`, `skills/moai/references/routes.json`(route 8개, goal 없음), `src/main.rs` HostLaunch(-l/-a), `AGENTS.local.md` § Lane loop, `.claude/rules/moai/workflow/auto-semantics.md` §6 ladder·§9 gates
- GEARS 형식 REQ + 수용 기준 초안(후속 구현 카드가 그대로 쓸 수 있게)
- **코드·skill 파일 변경 없음**

## 수용 기준
- AC-1: 문서가 self-claim 7단계 각각에 대해 "RC1 구현 있음(경로:행) / 신규"를 표기 — 전부 소스 인용 확인(가설 금지)
- AC-2: 동시성 논증 1절: 로컬 파일·label이 lock이 될 수 없는 이유와 댓글 id 단조성 가정, 가정이 깨질 때의 fallback(uncertain) — `src/claim.rs:15-23` 인용
- AC-3: `/goal` 자동 입력 가능 여부 3항목(Claude 내장·MoAI goal_arm·Codex)을 "확인됨/미확인"으로 구분, Codex 항목은 공식 문서 URL+확인일 또는 "미확인" 유지
- AC-4: 후속 구현 카드 분해(skill 인자·factory-dispatch 문서·assignments 연동·auto-semantics 연결·론처 결정) 각 1문단
- AC-5: 개인 절대 경로·비밀 0건, check set 실측 exit 기록

## 사용자 결정
- 2026-11-09 00:0x~00:3x: 모드 skill 진입, 완전/반자율, self-claim 기본, 리더 5분 PR 점검·타 에이전트 수리 지시, `/goal` 붙여넣기 안내
