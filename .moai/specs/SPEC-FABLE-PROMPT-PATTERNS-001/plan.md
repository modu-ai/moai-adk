# SPEC-FABLE-PROMPT-PATTERNS-001 — plan.md

## §A Context

doc-only Tier M. 8개 Fable 패턴을 9개 기존 canonical 파일(+각 템플릿 미러)에 통합한다.
Go 코드 무변경 — 검증은 grep 기반 AC + 템플릿 CI 가드 + `moai spec lint`로 수행한다.
소스 증거·앵커·기준선은 `research.md`에 자체 보존 (스크래치패드 의존 없음).

접근 원칙:
1. **SSOT 존중**: 패턴은 기존 canonical 파일 **안으로** 통합. 신규 파일 금지 (REQ-FPP-016).
2. **Template-First**: 편집 파일 9개 전부 미러 보유 — 각 마일스톤 내에서 live+미러 동시 갱신.
   미러 편집 전 parity class 재실측 필수 (research.md §A: BYTE-IDENTICAL 2건 실측,
   sanitized-pair 2건 실측, 5건 미측정 — parity는 시변이므로 run-phase에서 9건 전부 재실측).
3. **Baseline-delta AC**: 모든 AC는 기준선(전부 0, 보존 불변량 2건) 대비 delta grep — 공허한
   GREEN 차단.
4. **다이어트 준수**: 상시 로드 파일(CLAUDE.md, session-handoff.md, constitution 등)에는 최소
   문장만 추가; 부피가 있는 본문(대조 쌍 예시)은 examples 동반 파일에 배치.

## §B Known Issues / Risks / [NEEDS CLARIFICATION]

1. **[NEEDS CLARIFICATION: no-narration 규칙과 moai.md §8 Insight 배너의 관계]** — §8 Insight
   배너(What/Why/Alternatives)는 결정 서사를 의도적으로 수행하는 sanctioned surface다.
   권고안: §8 구조화 배너 전체를 no-narration의 명시적 예외로 선언하고, 금지 범위를 사전-결정
   기계 산문(선택 숙고, preload 언급, 로딩 서사)에 한정. 대안: Insight 배너 축소/폐지.
   run-phase 진입 전 orchestrator가 AskUserQuestion으로 해소할 것.
2. **[NEEDS CLARIFICATION: §16 never-deny 가드와 confirm-first 검색 절차의 조화]** — 현행
   CLAUDE.md §16 Process (2)는 과거 세션 검색 전 AskUserQuestion 확인을 요구한다. Fable의
   가드는 "검색 없이 부재 단언 금지"다. 권고안 (a): "이전 논의가 없다고 단언하려면 검색을
   수행했거나 검색을 제안한 후여야 한다" (confirm-first 보존, 부정-단언만 게이트).
   대안 (b): 부재 확인용 경량 인덱스 grep은 무확인 허용 예외 신설. run-phase 진입 전 해소.
3. **드리프트 sentinel**: session-handoff.md ↔ moai.md §8은 SSOT↔render 파리티 계약이 있다
   (session-handoff.md § Cross-references). 두 파일을 같은 SPEC에서 편집하므로 각 편집 후
   파리티 체크 필수 — Good/Bad 쌍이 6-block 스켈레톤·로케일 테이블을 변경하지 않음을 확인.
4. **CLAUDE.md 40K 한도**: 현재 25,357 chars. M1+M3 추가분(추정 +1.5K 미만)은 여유 충분하나
   편집 후 `wc -c` < 40,000 재확인.
5. **sanitized-pair 미러**: verification-claim-integrity.md·askuser-protocol.md 미러는 이미
   DIVERGED — 파일 복사가 아닌 등가-내용 패치로 반영해야 하며, 미러 쪽에는 SPEC ID·REQ 토큰·
   내부 날짜가 없어야 한다 (§25.1). forbidden-phrase 카탈로그의 ko 문자열("검증 완료" 등)은
   금지-문자열 데이터이므로 템플릿 중립성 위반이 아님 (언어 중립성 §15는 프로그래밍 언어 축).
6. **BSD grep 한글**: AC의 한글 리터럴 grep은 문자-클래스 범위(`[가-힣]`)가 아닌 리터럴 문자열
   매칭만 사용 (BSD grep false-negative 회피). 필요 시 `rg` 사용.
7. **lint**: 신규 문장에 legacy `IF/THEN` 사용 금지 (LegacyEARSKeyword). GEARS 형식 유지.

## §C Pre-flight (run-phase 진입 시 최우선 실행)

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — 병렬 세션
   race 확인 (다중 세션 활성 이력 있는 리포).
2. 9개 대상 파일의 live↔mirror parity 재실측: `for f in <9 files>; do diff -q "$f" "internal/template/templates/$f"; done`
   → 결과를 §E.2 증거로 기록.
3. research.md §C 기준선 grep 11종 재실행 — 기준선 드리프트 시 AC 기대값 보정 후 진행.
4. `wc -c CLAUDE.md` 기준선 기록.
5. [NEEDS CLARIFICATION] 2건이 해소되었는지 확인 — 미해소 시 blocker 보고 후 중단.

## §D Constraints

- doc-only: `internal/`, `pkg/`, `cmd/` 무접촉 (미러 `internal/template/templates/**` 마크다운 제외).
- 커밋 규율: 마일스톤당 1커밋, `feat(SPEC-FABLE-PROMPT-PATTERNS-001): M<n> …` (plan-phase 산출물
  커밋은 본 계획 밖 — orchestrator 소관).
- 미러 커밋에 SPEC ID 포함 금지 대상은 **파일 내용**이며 커밋 메시지는 무관.
- 시간 예측 금지 — priority 순서만.
- 기존 로케일 테이블·6-block 스켈레톤·배너 스키마 byte-보존 (메타 지침 추가만).

## §E Self-Verification (run-phase 완료 게이트)

1. acceptance.md AC-FPP-001..016 전 grep 실행 + 기대값 대조 (단일 턴 병렬 배치).
2. 보존 불변량: `grep -c "score ≥ 7" orchestration-mode-selection.md` = 1 유지;
   session-handoff 6-block 스켈레톤 무변경 diff 확인.
3. `go test ./internal/template/...` exit 0 (neutrality/leak 가드 포함).
4. `grep -rn "SPEC-FABLE-PROMPT-PATTERNS\|REQ-FPP-" internal/template/templates/` = 0 매치.
5. `moai spec lint .moai/specs/SPEC-FABLE-PROMPT-PATTERNS-001 --strict` exit 0.
6. `wc -c CLAUDE.md` < 40,000.

## §F Milestones (owner: manager-develop, cycle_type=ddd — 기존 문서 행위 보존 리팩터)

### M1 — 라우팅 순회 규율 + anti-rationalization (P1, P2) — Priority High

| Step | 작업 | 파일 | REQ |
|------|------|------|-----|
| M1-1 | §4 Selection Decision Tree에 stop-at-first-match preamble + anti-rationalization 절 추가 | `CLAUDE.md` | 001, 003 |
| M1-2 | §4 Delegation Decision/Forced Delegation Table에 first-match 규율 + anti-rationalization 절 추가 | `.claude/output-styles/moai/moai.md` | 002, 004 |
| M1-3 | Mode Dispatch 우선순위 목록에 first-match 순회 한 문장 명시화 (D5 — 최소 편집) | `.claude/rules/moai/workflow/spec-workflow.md` | 001 |
| M1-4 | 미러 동기 3건 (parity class별 절차: BYTE-IDENTICAL → 동일 패치; sanitized → 등가 패치) + 중립성 grep | `internal/template/templates/{CLAUDE.md, .claude/output-styles/moai/moai.md, .claude/rules/moai/workflow/spec-workflow.md}` | 015 |
| M1-5 | AC-FPP-001..004 grep 검증 + `wc -c CLAUDE.md` | — | — |

### M2 — 대조 예시 쌍 + forbidden phrases (P3, P4) — Priority High (M1 완료 후)

| Step | 작업 | 파일 | REQ |
|------|------|------|-----|
| M2-1 | 마커 규약(`**Good:**`/`**Bad:**`+`Why bad:`) 정의 + 질문 구성 대조 쌍 ≥1 | `askuser-protocol.md` | 005, 006 |
| M2-2 | 핸드오프 emission 대조 쌍 본문(examples) + 포인터(session-handoff.md) | `session-handoff-examples.md`, `session-handoff.md` | 007 |
| M2-3 | Completion Report Good/Bad 쌍 (bad = 증거 경로 없는 미관측-주장 배너) | `.claude/output-styles/moai/moai.md` §8 | 008 |
| M2-4 | vci §6 forbidden-phrase 카탈로그 신설 (≥8 en+ko 항목 + 조건부 허용 대안 2단 구조) | `verification-claim-integrity.md` | 009 |
| M2-5 | moai.md §8 Verification Matrix/Completion Report 규칙 목록에 §6 포인터 추가 (중복 금지) | `.claude/output-styles/moai/moai.md` | 010 |
| M2-6 | 미러 동기 5건 + 중립성 grep + session-handoff↔moai.md §8 파리티 sentinel 체크 | 미러 5건 | 015 |
| M2-7 | AC-FPP-005..010 grep 검증 (쌍 균형: Good 수 == Bad 수) | — | — |

### M3 — 스케일링 + 언어 신호 + no-narration + 메모리 방어 (P5-P8) — Priority Medium (M2 완료 후)

| Step | 작업 | 파일 | REQ |
|------|------|------|-----|
| M3-0 | [NEEDS CLARIFICATION] 2건 해소 결과를 편집안에 반영 (미해소 시 blocker) | — | — |
| M3-1 | §B.1c tool-call volume heuristic 신설 (§B.1b와 §B.2 사이; SSOT 임계 재기술 금지) | `orchestration-mode-selection.md` | 011 |
| M3-2 | §16 Search-when에 언어 신호 3클래스 + never-deny-without-search 가드 (해소안 반영) | `CLAUDE.md` | 012 |
| M3-3 | §10 Output Rules에 [HARD] no-narration 불릿 (해소된 §8 배너 예외 스코프 반영) | `.claude/output-styles/moai/moai.md` | 013 |
| M3-4 | §Lessons Protocol에 Memory-as-Data Boundary 규칙 4항 | `moai-constitution.md` | 014 |
| M3-5 | 미러 동기 4건 + 중립성 grep | 미러 4건 | 015 |
| M3-6 | §E Self-Verification 전체 게이트 실행 (AC 전건 + 전역 게이트) | — | 016 |

## §G Anti-Patterns (run-phase 금지)

- 미러에 live 파일 blind copy (sanitized-pair 파괴 + SPEC ID 누출).
- §B.1 임계값(≥3/≥10/≥7)을 §B.1c에 재기술 (SSOT 이중화).
- vci §6 카탈로그 내용을 moai.md에 복제 (포인터만 허용, REQ-FPP-010).
- 대조 쌍 마커 변형 (`**GOOD:**`, `✅ Good` 등) — grep 파리티 파괴.
- 신규 rules 파일 생성, agents/skills 본문 편집 (REQ-FPP-016).
- 기존 배너 스키마·로케일 테이블·6-block 스켈레톤 변경.
- Fable consumer 정책(안전/저작권/wellbeing) 문장 이식.

## §H Cross-References

- `research.md` — 소스 발췌·앵커·기준선·결정 D1-D5
- `acceptance.md` — AC 매트릭스 + Given-When-Then + DoD
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1/§25.3 — 미러 중립성 체크리스트
- `.claude/rules/moai/workflow/spec-workflow.md` § Plan Audit Gate — Phase 0.5 진입
- CLAUDE.local.md §2 Template-First Rule — 미러 동기 의무의 상위 규정
