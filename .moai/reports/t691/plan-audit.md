# Plan-Phase Audit Report — SPEC-HARBOR-KITE-001 (card t691)

- Auditor: plan-auditor (독립 감사, iteration 1/3 — Tier M ceiling 2 적용 대상이나 본 카드는 run 단계 없는 판정 카드로 1회 전수 감사)
- Date: 2026-09-13
- Tree: worktree `.claude/worktrees/t691`, branch `WT-harbor-kite-judgment`, base `e7b93c120`, 미커밋 산출물 대상
- Audit scope: PLAN PHASE ONLY (문서 감사 — 코드 감사 대상 없음, 구현 없음)

## Verdict: CONCERNS (pass-with-required-fixes)

Aggregate 0.81 (Tier M PASS threshold 0.80 이상). Must-pass 7항목 전부 PASS/N-A. 판정 자체(처분 (b) 비주입)와 재현 기록은 인용된 증거에 충실하고 반박에 성공했다. 다만 verification 계측의 정밀도 결함 3건(blocking class)이 있어, 무결 PASS로 치지 않고 열거된 수정 후 확정을 권한다. 3건 모두 문서 한 줄 수정이다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: REQ-001~REQ-005 연속, 공백·중복 없음 (spec.md:32,36,40,44,48)
- [PASS] MP-2 GEARS 형식 (판정 대상 계층: spec.md §1의 REQ-XXX 요구 계층): REQ-001 Unwanted("shall not set", :34), REQ-002/003/005 Event-driven("When … shall", :38,42,50), REQ-004 Ubiquitous("shall be treated", :46) — 5/5 패턴 적합. AC 계층(acceptance.md의 Given-When-Then)은 검증 계층의 정규 형식으로 MP-2에서 벌점화하지 않음
- [PASS] MP-3 YAML frontmatter: spec.md:2-15에 12개 canonical 필드 전부 존재(id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags), snake_case alias 없음, `phase: "v3.2.0 target"` 릴리스 대상 라벨 적법, `status: draft` enum 적합. plan.md/acceptance.md는 status 축 무상태 규정 준수(status 필드 없음)
- [N/A] MP-4 언어 중립: 단일 언어(Go) 프로젝트 스코프, 다국어 도구 내용 없음 — 자동 통과
- [PASS] MP-5 D7 교차-SPEC 조정: `SPEC-([A-Z][A-Z0-9]+-)+[0-9]+` 추출 결과 3개 아티팩트 모두 자기 ID(SPEC-HARBOR-KITE-001)뿐 — 참조된 타 SPEC 없음, BLOCKING 없음. 카드 t400·커밋 25b341dcb 참조는 SPEC-ID가 아님
- [PASS] MP-6 D8 크로스플랫폼: `syscall` 출현 0회 — 자동 PASS
- [PASS] MP-7 clarification gate: plan.md에 `[NEEDS CLARIFICATION` 0건, research.md 부재는 Tier M 정상 아티팩트 셋(4종: spec/plan/acceptance/progress) — N/A 아닌 PASS

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 1.00 | 1.0 | REQ 각각 단일 해석. 인용 필요 시 spec.md:34-50 전 요구가 측정 가능 서술 |
| Completeness | 0.75 | 0.75 | WHY/WHAT H2 표제 부재 — 내용은 HISTORY(spec.md:20-28)에 접혀 있으나 표제 계층 결여(D4). 그 외 전 섹션 존재: Out of Scope 3개 H3 + `-` 항목(spec.md:78-91), 교차 참조(spec.md:93-98), Tier M 아티팩트 셋 완비 |
| Testability | 0.75 | 0.75 | AC-001/002 이진 판정 가능(acceptance.md:28-36). AC-003 "존중 어조" 비이진 요소(D5), AC-004 명령-결과 불일치로 재량 판정 1회 요구(D1) |
| Traceability | 0.75 | 0.75 | REQ-001→AC-004, REQ-002→AC-001, REQ-003→AC-002, REQ-005→AC-003 매핑 존재하나 REQ-004 무AC(D2) |

## 감사 질문별 답변

### Q1 — 결정 기록 완전성과 근거의 정직성

G1-G4는 인용 증거로부터 실제로 따라 나온다. 대조 근거:

- **G1** ≅ 상세 문서 :77 "upstream internal flag rather than a documented interface … name can change without notice" — 정합. 런처 기본값 결합 위험 서술도 문서의 "not something to wire in by default"과 일치.
- **G2 핵심** ≅ 상세 문서 :65-67,:73 — 슬롯이 기기 전역 last-writer-wins 상태이고 서드파티 세션은 기록에 참여하지 않는다는 것은 문서대로. 그러나 **괄호 안의 DISABLE_TELEMETRY 계열 사례는 과장 소지**(D3): 계열 플래그가 "flag evaluation 자체를 끈다"는 것은 문서화돼 있으나(cross-session-messaging.md § Availability constraints), 그 상태에서 env 게이트가 승리하는지(=1이 실제로 그 사용자의 선택을 덮어쓰는지)는 두 문서 어디에도 확립돼 있지 않다. 상세 문서 :77의 "gate checks … before it reads the slot"은 슬롯값 축의 서술이지 evaluation-off 축의 서술이 아니다. 처분 자체는 양쪽 읽기 모두에서 불변(=1이 안 통하면 주입이 무효일 뿐) — 판정은 흔들리지 않으나, 공개 회신 초안에도 같은 문장이 실려 있어 한 절의 완충 수정이 필요하다.
- **G3** ≅ 상세 문서 :67 — "A session pointed at a third-party endpoint … receives no payload of its own to write". 게이트웨이(127.0.0.1) 세션도 같은 비기록자 클래스라는 분해는 정확하며, t695 verdict의 게이트웨이 서사와 모순 없음.
- **G4** ≅ 상세 문서 :69 (게이트 매 호출 재판독 → 자가 치유) + cross-session-messaging.md § dispatch-not-depend-on-reply (nudge/delegation 위계) — 정합.

### Q2 — 재현 기록의 건전성

- **단일 변수 통제**: 성립. 방법론이 상세 문서 :75의 t400 측정 선례("Holding model and environment fixed and flipping only the slot")와 동일 축이며, 슬롯값 외 변수 고정 선언(spec.md:65)이 있다.
- **격리 주장**: 구조적으로 신뢰 가능(CLAUDE_CONFIG_DIR가 config를 프로파일로 재지향 — 슬롯이 config dir의 .claude.json에 살므로 실파일 우회가 구조적으로 차단). 그리고 검증 가능한 절반이 참으로 확인됐다: `~/.claude.json` 현재 `"tengu_harbor_kite": true` (실측) — spec.md:65의 "현재 true로 읽힌다"와 일치. 프로파일 디렉터 2개가 실존하며 각각 true/false로 조립돼 있음도 실측 확인(t691-true/.claude.json=true, t691-false/.claude.json=false, 생성 2026-09-13 20:02/20:06).
- **허위-미수행 반입**: 없음. 재현 기록은 자기 측정 범위를 정직하게 표기(행동 수준 확장임을 명시, t400은 소켓 수준으로 구분). 다만 셀당 시작 횟수(각 1회)와 바이너리 버전이 미기재라는 증거 희소성은 남는다(D7).

### Q3 — AC 검증 가능성과 무-run 마감 부호화

- AC-001/002: Read 기반 이진 판정 가능, 대상 내용 실제 존재 확인(§2 처분+G1-G4+doctor 제안 spec.md:54-61, §3 표+그대로 인용+격리 문구 spec.md:65-72).
- AC-003: 5개 필수 요소는 초안에 전부 존재(초안 §1-§5)하나 "존중 어조" 절이 재량 판정 요소(D5).
- AC-004: **의도는 참이고 본 감사가 직접 재실행해 확인**(`git grep -n "CLAUDE_CODE_HARBOR_KITE" -- internal/ cmd/ pkg/` → 환경 설정 코드 0히트). 그러나 명령이 문서 미러(internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md:77)를 훑어 1히트를 반환 — Then 절의 "0히트"와 명령 출력이 서면 제외 조항을 검사자가 손으로 적용해야만 일치(D1). pathspec 수정 필요.
- **무-run 마감 부호화**: 정확. acceptance.md 서두("구현이 수반되지 않으므로…"), §D.2 게이트, §D.3 DoD, plan.md M3 N/A 행, progress.md §E.2/§E.3 "해당 없음" — 5개 표면이 모두 일관되게 run 부재를 선언.

### Q4 — t400 정합성과 회신 초안 정확성

- **t400 모순 없음**: 슬롯 명칭/위치, last-writer-wins, first-party-only 기록, 매 호출 재판독, env 우선 탈출구, "MoAI는 주입하지 않는다" 의도성, 사설 플래그 불안정성 — 전부 상세 문서와 정합. SPEC은 t400의 소켓 수준 발견을 행동 수준으로 확장한다고 정직하게 위치 설정.
- **초안-SPEC 불일치**: 없음. 초안의 5 섹션이 REQ-005의 5 요소와 1:1. doctor 후속 표현(초안 "별도 카드로 접수될 예정" vs plan.md "리드가 신규 카드로 접수") 정합. 초안은 구현을 약속하지 않음("저희는 넣지 않기로 판정했습니다") — 정직. 탈출구의 "env가 슬롯보다 먼저 읽힙니다"는 상세 문서 :77 문증.
- **정밀도 결함**: D3(평가-off 사례의 미측정 기제 단정, G2와 동일 문장), D8(BASE_URL 호스트 조건으로의 기제 압축 — 상세 문서는 평가 엔드포인트 하드코어드+타기팅 속성 구조), D9(공개 회신에 내부 카드 id "t400" 노출 — 표현 위생). 어조는 정중하고 사실 범위 내.

### Q5 — 스코프 규율

이 카드 어디에도 구현 승인이 없다. plan.md §D(구현/코드 변경 금지, 커밋 금지), M3 N/A, §G가 "run 단계 무단 진입"을 반패턴으로 명명, spec §4의 Out of Scope 3개 소주제(런처 주입/doctor/게시·프로파일 정리)가 각각 소관 밖임을 명시. doctor 검사는 신규 카드 후보로만 라우팅됨(§2, §4, plan §F 후속 행, 초안 §5) — 본 카드로의 반입 없음. REQ-001은 불변식 기록이며 부재 증명(AC-004)으로 검증될 뿐 작업을 승인하지 않음. 결론: 스코프 결함 없음.

## Defects Found

D1. AC-004-VERIFY — acceptance.md:24,47-48 — 검증 명령이 문서 미러(`internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md:77`)를 훑어 1히트를 반환; Then 절 "0히트"는 서면 제외를 검사자가 수동 적용해야만 성립 — Severity: minor — Class: blocking — Required fix: 명령에 `':!internal/template'` 제외(또는 `-- '*.go'` 한정)를 넣어 PASS가 명령 자체로 자명하게. 수정 후 Go 파일 스윕 0히트가 기대 green임을 Then에 명기.

D2. REQ-004-UNCOVERED — spec.md:44-46, acceptance.md:21-24 — REQ-004(유실의 성질 규정)에 대응 AC 없음; AC 행렬은 REQ-001/002/003/005만 커버 — Severity: minor — Class: blocking — Required fix: AC-005 추가(예: Given §2 G4와 §4를 읽을 때, Then 유계-비후원 입장과 "launcher가 보상하지 않는다" 규정이 기록돼 있다)하거나 REQ-004를 결정 기록 산문으로 강등.

D3. G2-TELEMETRY-PARENTHETICAL — spec.md:57, issue-reply-draft.md:29 — "비필수 트래픽 제한 계열 플래그로 평가를 끈 경우 포함, =1은 그 이유를 덮어쓴다"는 주장: evaluation-off 상태에서 env 게이트가 승리하는지는 인용된 어느 문서도 확립하지 않음(상세 문서 :77은 슬롯값 축만 서술; 규칙 파일은 계열 플래그가 메시징을 "조용히 끈다"고만 기술). 공개 회신에 미측정 기제가 단정형으로 실림 — Severity: minor — Class: blocking — Required fix: 한 절 완충("이 경우의 =1 동작은 미측정 — 슬롯값 경로에 대해서만 관측됨") 또는 괄호 사례 삭제. 처분은 어느 읽기에서도 불변이므로 판정 자체는 영향 없음.

D4. NO-WHY-WHAT-HEADINGS — spec.md:20-28 — WHY/WHAT 표제 없이 HISTORY에 내용 접힘; 내용은 존재하나 구조 편차 — Severity: minor — Class: optional — Required fix: 없어도 무방. 정규화하려면 HISTORY 아래 소제목 또는 §0 개요 한 단락.

D5. AC-003-TONE-CLAUSE — acceptance.md:42 — "존중 어조로 쓰여 있다"는 비이진 판정 요소가 Must AC에 포함 — Severity: minor — Class: optional — Required fix: 어조 절을 AC 밖 검토 지침으로 이동하거나 기계적 대리 판정(존댓말 사용, 비난 표현 0건)으로 대체.

D6. REQ-005-EXTERNAL-ACTOR — spec.md:48-50 대비 :90 — "lead shall post"가 외부 행위자의 행위를 규범화하는데 그 행위는 §4가 카드 범위 밖으로 선언; AC-003은 초안 내용만 검증 — Severity: minor — Class: optional — Required fix: REQ-005을 "회신 초안은 …을 담는다"로 초안-내용 규범으로 재서술하거나 현행 유지(모순은 아님, 재량 해소 가능).

D7. REPRO-SPARSENESS — spec.md:63-72 — 셀당 시작 횟수(각 1회)와 바이너리 버전 미기재; "동일 바이너리" 무귀속 — Severity: minor — Class: optional — Required fix: 버전 문자열과 셀당 시행 수 한 줄 추가(프로파일 디렉터 실존과 슬롯값은 본 감사가 실측 보강 완료).

D8. MECHANISM-COMPRESSION — plan.md:17, issue-reply-draft.md:11 — "ANTHROPIC_BASE_URL 호스트가 api.anthropic.com인 세션만 기록"은 상세 문서의 구조(평가 엔드포인트 하드코어드, BASE_URL은 타기팅 속성으로만 기여)의 압축 서술; 운용적 결론은 동일 — Severity: minor — Class: optional — Required fix: 공개 초안 쪽만 "1st-party(기본 엔드포인트) 세션만"으로 완화 권고.

D9. INTERNAL-ID-LEAK — issue-reply-draft.md:22 — 공개 회신에 내부 카드 id "t400" 노출 — Severity: minor — Class: optional — Required fix: "이전 조사"로 일반화.

심사 균형 고지(M6): optional 6건은 verdict를 움직이지 않는다. blocking 3건은 전부 검증 계측/기록 정밀도의 한 줄 수정이며, 처분(b)·재현·스코프 규율이라는 카드의 실질은 이번 감사의 대조·실측으로 전부 생존했다.

## Regression Check

해당 없음 (iteration 1 — 선행 결함 목록 없음).

## Recommendation

1. D1: acceptance.md AC-004 명령에 `':!internal/template'` 제외 추가 — 그대로 재실행 가능한 형태로.
2. D2: REQ-004에 AC-005 추가 또는 산문 강등.
3. D3: spec.md G2와 초안 §3 두 번째 불릿의 평가-off 사례에 미측정 완충 한 절(리드가 게시 전 초안을 보므로 이 수정이 공개 노출을 막는다).
4. D4-D9: 임의. D8+D9는 공개 초안 품질 차원에서 리드 게시 전 적용 권장.

3건의 blocking 수정은 manager-spec의 문서 편집으로 즉시 가능하며, 수정 후 재감사는 위 열거 결함 delta에 한정된 확인 재감사(전수 재감사 불요)다. 수정 없이 진행할 경우 CONCERNS 상태로 적립 기록(accept-with-debt)도 가능하나, D3은 공개 게시 전 반드시 완충할 것.

## 감사자 실측 근거 일람 (baseline attribution)

- `git grep -n "CLAUDE_CODE_HARBOR_KITE" -- internal/ cmd/ pkg/` → 1히트(internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md:77) — 이번 run, 이 워크트리
- `grep -Eo 'SPEC-…'` 3 아티팩트 → 자기 ID만 (D7)
- `grep -c syscall spec.md` → 0 (D8); `grep -rn '\[NEEDS CLARIFICATION' plan.md` → 0건, research.md 부재 (MP-7)
- `grep -o '"tengu_harbor_kite":[^,}]*' ~/.claude.json` → true; `~/.moai/claude-profiles/t691-true/.claude.json` → true; `t691-false/.claude.json` → false (Q2 격리·조립 주장 보강)
- `.moai/reports/t672/` 는 이 워크트리에 부재 — t672는 진행 중 카드로 판단(교차 참조 자재 중 verdict 미존재, t691 판정에는 영향 없음). t695 verdict는 읽었으며 게이트웨이/서드파티 서사와 정합.
