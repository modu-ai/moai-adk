# SPEC-FABLE-PROMPT-PATTERNS-001 — acceptance.md

모든 AC는 baseline-delta grep 기반 기계 검증이다. 기준선(2026-07-11 실측)은 `research.md` §C —
전 신규 패턴 grep 기준선 0, 보존 불변량 2건(`score ≥ 7`=1, moai.md `narrat`=1). run-phase
pre-flight에서 기준선 재실측 후 드리프트 시 기대값을 보정한다. 한글 리터럴은 문자열 매칭만
사용한다 (BSD grep 문자-클래스 범위 금지).

경로 축약: `OS` = `.claude/output-styles/moai/moai.md`, `RULES` = `.claude/rules/moai`,
`TPL` = `internal/template/templates`.

## §D AC Matrix

### M1 — 라우팅 (REQ-FPP-001..004)

**AC-FPP-001** (REQ-001): CLAUDE.md §4 stop-at-first-match preamble 존재.
```bash
grep -c "first match" CLAUDE.md        # expected: >= 1 (baseline 0)
awk '/### Selection Decision Tree/,/### Retained Agents/' CLAUDE.md | grep -ci "first match"  # expected: >= 1 (윈도 내 존재)
```

**AC-FPP-002** (REQ-002): moai.md §4 first-match 규율 존재.
```bash
grep -c "first match" "$OS"            # expected: >= 1 (baseline 0)
```

**AC-FPP-003** (REQ-003): CLAUDE.md anti-rationalization 키 문구 존재.
```bash
grep -c "style opinion, not a category mismatch" CLAUDE.md   # expected: 1 (baseline 0)
grep -c "escape hatch" CLAUDE.md                              # expected: >= 1 (baseline 0)
```

**AC-FPP-004** (REQ-004): moai.md §4 Forced Delegation Table anti-rationalization 절 존재.
```bash
awk '/## 4. Delegation Decision/,/## 5\./' "$OS" | grep -ci "subdivid\|escape"  # expected: >= 1
```

**AC-FPP-001b** (REQ-001, spec-workflow 표면): Mode Dispatch first-match 명시화.
```bash
grep -c "first match" "$RULES/workflow/spec-workflow.md"      # expected: >= 1 (baseline 0)
```

### M2 — 대조 쌍 + forbidden phrases (REQ-FPP-005..010)

**AC-FPP-005** (REQ-005): 마커 규약이 askuser-protocol.md에 1회 정의되고 `Why bad:` 요건 포함.
```bash
grep -c 'Why bad:' "$RULES/core/askuser-protocol.md"          # expected: >= 1 (규약 정의 + 쌍 사용)
```

**AC-FPP-006** (REQ-006): askuser-protocol.md 질문 구성 대조 쌍 ≥1, 쌍 균형.
```bash
G=$(grep -c '\*\*Good:\*\*' "$RULES/core/askuser-protocol.md"); B=$(grep -c '\*\*Bad:\*\*' "$RULES/core/askuser-protocol.md")
echo "G=$G B=$B"                                              # expected: G >= 1 && G == B (baseline 0/0)
```

**AC-FPP-007** (REQ-007): 핸드오프 emission 쌍(examples 본문) + 포인터(session-handoff.md).
```bash
G=$(grep -c '\*\*Good:\*\*' "$RULES/workflow/session-handoff-examples.md"); B=$(grep -c '\*\*Bad:\*\*' "$RULES/workflow/session-handoff-examples.md")
echo "G=$G B=$B"                                              # expected: G >= 1 && G == B (baseline 0/0)
grep -c "Good/Bad\|contrastive" "$RULES/workflow/session-handoff.md"  # expected: >= 1 (포인터)
```

**AC-FPP-008** (REQ-008): moai.md Completion Report Good/Bad 쌍 ≥1, 쌍 균형 (파일 전체 카운트 —
windowed-grep undercount 회피; §8 배치는 리뷰로 확인).
```bash
G=$(grep -c '\*\*Good:\*\*' "$OS"); B=$(grep -c '\*\*Bad:\*\*' "$OS")
echo "G=$G B=$B"                                              # expected: G >= 1 && G == B (baseline 0/0)
```

**AC-FPP-009** (REQ-009): vci §6 카탈로그 — 헤딩 + ≥8 항목 + en/ko 혼재 + 조건부 허용 구조.
```bash
F="$RULES/core/verification-claim-integrity.md"
grep -c "^## 6\." "$F"                                        # expected: 1 (신규 §6 헤딩)
sed -n '/^## 6\./,/^## /p' "$F" | grep -c '^- '               # expected: >= 8 (항목 불릿)
grep -c "tests should pass" "$F"                              # expected: >= 1 (baseline 0)
grep -c "검증 완료" "$F"                                       # expected: >= 1 (ko 항목)
grep -c "ONLY" "$F"                                           # expected: >= 1 (조건부 허용 2단 구조)
```

**AC-FPP-010** (REQ-010): moai.md §8에 §6 포인터 존재 + 카탈로그 미복제.
```bash
grep -c "forbidden" "$OS"                                     # expected: >= 1 (포인터, baseline 0 — pre-flight 재확인)
grep -c "tests should pass" "$OS"                             # expected: 0 (내용 미복제)
```

### M3 — 스케일링/신호/서사/메모리 (REQ-FPP-011..014)

**AC-FPP-011** (REQ-011): §B.1c 존재 + 수치 대역 + 승격 출구 + SSOT 비재기술.
```bash
F="$RULES/workflow/orchestration-mode-selection.md"
grep -c "B.1c" "$F"                                           # expected: >= 1 (baseline 0)
grep -c "3–5\|3-5" "$F"                                       # expected: >= 1
grep -c "5–10\|5-10" "$F"                                     # expected: >= 1
grep -c "20+" "$F"                                            # expected: >= 1
grep -c "score ≥ 7" "$F"                                      # expected: 1 (보존 불변량 — SSOT 문장 유일성 유지)
```

**AC-FPP-012** (REQ-012): CLAUDE.md §16 언어 신호 3클래스 + never-deny 가드.
```bash
grep -c "possessive" CLAUDE.md                                # expected: >= 1 (baseline 0)
grep -c "definite article" CLAUDE.md                          # expected: >= 1 (baseline 0)
grep -c "past-tense" CLAUDE.md                                # expected: >= 1 (baseline 0)
grep -c "without having searched\|without searching" CLAUDE.md  # expected: >= 1 (baseline 0)
```

**AC-FPP-013** (REQ-013): moai.md §10 [HARD] no-narration 불릿 (§10 윈도 내).
```bash
sed -n '/^## 10\./,/^## 11\./p' "$OS" | grep -c "narrat"      # expected: >= 1 (baseline: §10 윈도 내 0)
sed -n '/^## 10\./,/^## 11\./p' "$OS" | grep -c "per my guidelines"  # expected: >= 1 (금지 예시 명시)
```

**AC-FPP-014** (REQ-014): constitution §Lessons Protocol memory-as-data 규칙.
```bash
F="$RULES/core/moai-constitution.md"
grep -c "background data" "$F"                                # expected: >= 1 (baseline 0)
grep -c "never executable\|not executable" "$F"               # expected: >= 1
grep -c "verbatim" "$F"                                       # expected: >= 1 (Lessons 절 내 — 리뷰로 위치 확인)
grep -c "cannot override" "$F"                                # expected: >= 1
```

### Cross-cutting (REQ-FPP-015/016)

**AC-FPP-015a** (REQ-015): 미러 내용 패리티 — 각 편집 파일 f에 대해 위 해당 AC grep을
`TPL/<path>`에 동일 실행, 기대 카운트 동일. (sanitized-pair 파일은 카운트 동일성만 요구,
byte-parity 불요.)
```bash
# 예시 (전 9쌍 반복):
grep -c "first match" "$TPL/CLAUDE.md"                        # expected: live와 동일
grep -c "^## 6\." "$TPL/.claude/rules/moai/core/verification-claim-integrity.md"  # expected: 1
```

**AC-FPP-015b** (REQ-015): 미러 중립성 — SPEC ID/REQ 토큰 무누출 + CI 가드 green.
```bash
grep -rn "SPEC-FABLE-PROMPT-PATTERNS\|REQ-FPP-" internal/template/templates/  # expected: 0 matches
go test ./internal/template/... > /tmp/moai-verify/tpl.log 2>&1; echo "exit=$?"  # expected: exit=0
```

**AC-FPP-016** (REQ-016): 신규 파일 무생성 + agents/skills 무접촉.
```bash
git diff --name-status <run-base>..HEAD -- .claude/rules .claude/output-styles | grep -c '^A'  # expected: 0
git diff --name-only <run-base>..HEAD -- .claude/agents .claude/skills | wc -l                 # expected: 0
```

### 전역 게이트

**AC-FPP-G01**: `moai spec lint .moai/specs/SPEC-FABLE-PROMPT-PATTERNS-001 --strict` → exit 0.

**AC-FPP-G02**: 보존 불변량 — `wc -c CLAUDE.md` < 40000; session-handoff.md 6-block 스켈레톤·
cut-line 마커·로케일 테이블 diff 무변경; moai.md §8 배너 템플릿 블록(코드펜스 내부) diff 무변경.

## Given-When-Then 시나리오

**시나리오 1 — 라우팅 합리화 차단 (P1+P2)**
- Given: 오케스트레이터가 "간단한 Go 헬퍼 함수 하나 추가" 요청을 받음 (§4 tree 4번 항목 매치)
- When: 편집된 CLAUDE.md §4를 순회
- Then: 4번 항목에서 first-match로 중단해 manager-develop 위임이 결정되고, "한 파일이라 직접
  해도 됨" 식의 하위분류 탈출이 anti-rationalization 절에 의해 차단된다 (문서상 규율 —
  AC-FPP-001/003 grep으로 문장 존재를 검증).

**시나리오 2 — 미관측-주장 자기점검 (P4)**
- Given: manager-develop이 §E 자기검증 보고서를 작성 중
- When: vci §6 카탈로그의 금지 문구("tests should pass", "검증 완료" 등)와 대조
- Then: 증거 인용 없는 금지 문구가 보고서에 존재하면 Gap으로 재분류해야 함이 문서에 명시되어
  있고, 각 금지 항목에 조건부 허용 대안이 짝지어져 있다 (AC-FPP-009).

**시나리오 3 — 미러 중립성 (REQ-015)**
- Given: M2 완료 후 vci 미러가 갱신됨
- When: `grep -rn "SPEC-FABLE-PROMPT-PATTERNS\|REQ-FPP-" internal/template/templates/` 실행
- Then: 0 매치 — 미러에는 패턴 내용만 반영되고 SPEC 흔적은 없다 (AC-FPP-015b).

## Edge Cases

1. **기준선 드리프트**: pre-flight 재실측에서 기준선이 0이 아니면 (병렬 SPEC이 동일 문구 추가)
   기대값을 baseline+delta로 보정하고 §E.2 증거에 기록.
2. **미러 parity 재분류**: 미측정 5쌍 중 sanitized-pair가 추가 발견되면 등가 패치 절차로 전환
   (blind copy 금지) — AC-FPP-015a는 카운트 동일성 기준이므로 영향 없음.
3. **`20+` 리터럴 충돌**: orchestration-mode-selection.md에 "20"이 다른 맥락으로 등장할 수 있음 —
   AC-FPP-011은 `20+` 리터럴(플러스 포함)로 한정.
4. **한글 grep**: "검증 완료"는 리터럴 매칭 — 정규화(NFC/NFD) 차이로 실패 시 `rg -F` 폴백.

## Definition of Done

- [ ] AC-FPP-001..016 (015a/015b, 001b 포함) 전건 PASS — 단일 턴 병렬 grep 배치 증거 첨부
- [ ] AC-FPP-G01 (spec lint --strict) / AC-FPP-G02 (보존 불변량) PASS
- [ ] [NEEDS CLARIFICATION] 2건 해소 기록이 plan.md에 반영됨
- [ ] 미러 9쌍 동기 + `go test ./internal/template/...` exit 0
- [ ] progress.md §E.2/§E.3 증거 기록 (manager-develop 소관)
