---
description: "Detail companion for the MoAI output-style Localization Contract (§8) — the worked ko-canonical anti-pattern catalogues for label rendering and banner-body prose, relocated off the always-loaded surface"
paths: "**/output-styles/moai/*.md,**/output-style-localization-catalogue.md"
---

# Output-Style Localization — Anti-Pattern Catalogues (Detail Companion)

> Detail companion of `.claude/output-styles/moai/moai.md` §8 Localization Contract (the
> always-loaded stub). The stub owns the [HARD] obligation — read `conversation_language`,
> translate every label and every banner-body sentence into natural native prose, preserve the
> verbatim set. This file owns the exhaustive worked examples behind that obligation.
>
> The examples are ko-canonical because ko is where the violations were observed. The
> naturalization principle is locale-neutral: for `ja` / `zh` / any other ISO-639 code, render
> what a native reader of that locale would expect, never a word-by-word mapping of the English.

## Label rendering — **Anti-pattern catalogue (HARD violations observed in production):**

When `conversation_language: ko`, emitting raw English literals from the §8 templates is a HARD violation. The reader expects equivalent natural Korean phrasing. The catalogue below shows wrong (raw English) and correct (ko canonical) renderings for every surface that has produced violations. The same translation principle applies to `ja` / `zh` / any other ISO-639 code — render in the user's configured language naturally.

| §8 surface | Raw English (wrong) | ko canonical (right) |
|------------|---------------------|----------------------|
| Gate header | `🤖 MoAI ★ Gate [2/4]` | `🤖 MoAI ★ 게이트 [2/4]` |
| Gate criteria | `Functional / Minimal / Verified / Traceable / Safe` | `기능성 / 최소성 / 검증 / 추적성 / 안전성` |
| Preconditions header | `Preconditions:` | `전제 검증:` |
| Complete: Files | `Files: 6` | `파일: 6` |
| Complete: Tests | `Tests: 7 ACs PASS` | `테스트: 7 ACs 통과` |
| Complete: Coverage | `Coverage: 100%` | `커버리지: 100%` |
| Complete: Deliverables | `Deliverables:` | `산출물:` |
| Complete: Specialists used | `Specialists used:` | `위임 specialist:` |
| Complete: Cleanup | `Cleanup: temp files removed` | `정리: 임시 파일 정리됨` |
| Insight banner header | `🤖 MoAI ★ Insight` | `🤖 MoAI ★ 인사이트` |
| Insight: What | `What:` | `결정:` |
| Insight: Why | `Why:` | `이유:` |
| Insight: Alternatives | `Alternatives:` | `대안:` |
| Insight: Implications | `Implications:` | `함의:` |
| Insight: Your call (pull mode) | `Your call:` | `판단은 사용자 몫:` |
| Delegation: Specialist | `Specialist:` | `전문가:` (또는 `Specialist:` 그대로 — technical role identifier) |
| Delegation: Scope | `Scope:` | `범위:` |
| Delegation: Constraints | `Constraints:` | `제약:` |
| Delegation: Return | `Return:` | `반환:` |
| Step labels (Step 1-4) | `Step 1: Clarify` / `Step 2: Delegate` / `Step 3: Execute` / `Step 4: Verify` | `1단계: 명확화` / `2단계: 위임` / `3단계: 실행` / `4단계: 검증` |
| Recovery options | `Pause / Retry as-is / Alt approach / Abort+preserve` | `일시 중지 / 현재대로 재시도 / 대안 접근 / 중단+보존` |

The catalogue above provides the ko canonical mapping for every label observed in production. For locales beyond ko/ja/zh, follow the same naturalization principle — do not transliterate. (The anti-pattern this Contract prevents — anchoring to the literal English example labels — is restated as a binding directive in §9.)

## Banner-body prose — **Banner body prose Anti-pattern catalogue (extended — ko canonical; same naturalization principle applies to ja / zh / other ISO-639 codes):**

| Surface | Raw English prose (wrong) | ko natural language (right) |
|---------|---------------------------|-----------------------------|
| Discovery `Findings:` body | `manager-develop pre-flight discovered scope ground-truth divergence` | `manager-develop이 사전 점검 중 범위 기준이 두 가지로 갈리는 문제를 발견` |
| Discovery `Findings:` body | `bash-grep literal substring narrow (35 files) vs Go regex word-boundary + prefix-allowlist (45 files) — 11 extras` | `spec.md §A.4에서 35개 파일로 측정한 누출 목록이 Go 테스트 regex로는 45개로 잡힘 — 11개가 추가로 식별됨` |
| Discovery `Recommended action:` body | `User 4-option 결정 (A/B/C/D, manager-develop alt recommendation = Option A 44 files comprehensive cleanup)` | `사용자가 A/B/C/D 4개 선택지 중 결정 필요 (manager-develop 대체 권장 = A안, 44개 파일 전체 정리)` |
| AskUserQuestion `description` field | `Clean all 44 files to match Go test scope. AC GREEN proof = clean PASS.` | `Go 테스트가 잡아내는 44개 파일을 모두 정리. 결과: 해당 AC가 명확하게 통과로 마무리됩니다.` |
| AskUserQuestion `preview` field | `actual cleanup: 39 files (45 - .gitignore - allowlist 5)` | `실제 정리 대상: 39개 파일 (45개 중 .gitignore 1개 + 교육 예외 5개 제외)` |
| AskUserQuestion `preview` field | `장점: doctrinally 정확 + 해당 AC 명확 PASS` | `장점: 정책 의도에 정확히 부합 + 해당 AC 명확 통과` |
| AskUserQuestion `preview` field | `단점: scope expansion +11 files, +1-2 commits` | `단점: 정리 범위가 11개 파일 늘어남, 커밋이 1-2개 추가됨` |
| Step/round update prose | `빠른 독립 verify 후 사용자 결정 surface합니다` | `빠르게 독립적으로 확인한 뒤 사용자 결정을 받겠습니다` |
| Gate body prose | `comprehensive cleanup` | `전체 정리` (또는 맥락에 따라 `포괄적 정리`) |
| Gate body prose | `scope discipline` | `범위 절제` (또는 `범위 규율 준수`) |
| Gate body prose | `narrow canonical` | `좁은 기준 채택` (또는 `좁은 정의 우선`) |
| Gate body prose | `silent semantic divergence` | `의미 차이가 조용히 누적된 상태` |
| Insight body prose | `decision required pending blocker` | `차단 사유로 사용자 결정이 필요한 상황` |
| Race Absorbed body prose | `parallel session race-absorbed clean fast-forward` | `병렬 세션 commit이 fast-forward로 흡수됨 (충돌 없음)` |

이 catalogue는 ko canonical. ja / zh / 기타 locale은 동일한 자연화 원칙으로 prose를 풀어쓴다 — 단어 단위 치환이 아닌 native speaker가 자연스럽게 듣는 문장 구조 채택. transliteration (음역) 금지.

---

Classification: Evolvable detail companion — worked examples only; every [HARD] obligation lives in the always-loaded stub.
