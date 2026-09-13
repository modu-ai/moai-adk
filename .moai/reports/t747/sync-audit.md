# t747 sync-audit — SPEC-AC-ANCHOR-SCOPE-001

측정: worktree `.claude/worktrees/t747` · branch `WT-ac-anchor-scope` · HEAD `d6947cdbc` (absorb base `188ece2f9`) · sync-auditor(opus/high) 수행, 이번 실행

## 판정: **PASS** — 91/100 (harmonic)

차단 항목: F1 1건(문서-only, 통합 전 수리 권고) — 나머지 전원 must-pass 통과.

## 차원 점수

| 차원 | 점수 | 판정 | 핵심 증거(이번 실행) |
|---|---|---|---|
| Functionality 40% | 88 | PASS | 코퍼스 프로브 독립 재실행: 861 분모·NARROW 14→11(3 수리)·EMPTY 9→3(6 수리)·CONTROL 129 양형상 0·NEWLY 48·FROZEN-CROSSCHECK MATCH; `go test ./internal/spec/ -count=1` ok 126.2s; AC 7건 전부 독립 재유도 PASS |
| Security 25% | 100 | PASS | RE2 선형 정규식(ReDoS 없음), 프로브 저장소 트리 비오염, 신규 신뢰경계 없음 |
| Craft 20% | 85 | PASS | 커버리지 90.9%(findACSectionStart 100%), lint 0 issues, vet 0, gofmt, GOOS=windows OK, RED 증거 존재 — 감점 F1/F2/F5 |
| Consistency 15% | 92 | PASS | t528 2-열 패턴 준수, D3 백필 정확, 단일 sync 커밋 3파일, close 주제 전체 ID+infix — 감점 F1/F7 |

## AC 재유도 요지 (7/7 PASS)

- AC-747-001: 14→11 — 3 REPAIRED 파일별 검증(AC-COLLECTOR `## 1. 배경과 문제` 3 decls / CC297 `## Requirements` 19 / STATUS-AUTO `## Requirements (EARS)` 25), 11 처분 기록. 표기 부정확 → F2.
- AC-747-002: 9→3 — 6 REPAIRED + 3 정당. 잔여 3건 전수 표본: GLM-EFFORT-MAX는 측정 인공물(`**AC-GEM-001** —` 볼드/대시형, declRe 불가 — 파서 결함 아님)로 정직하게 확인; OUTOFSCOPE·_archive 2건은 콜론 없는 bullet로 라인 문법 축 소관. 아티팩트 혼동 → F3.
- AC-747-003: 구조적으로 보장(제어 파일은 변경 불도달 불가 경로 — 초기 vocab 절 early-return 동일) + 실측 0. 강제 테스트 미비 → F5.
- AC-747-004: 48 = 47 narrow(3+19+25) + 1 empty(I18N), 제어 0; qualifier 가드 추적 완료(bullet+id+구분자 필수). 의미적 주름 → F4.
- AC-747-005: 단일 실행 2-열; 861 vs 860 델타 = 자기 SPEC spec.md(sorted diff로 검증). 자기-측정 수치 고정 → F6.
- AC-747-006: 미접촉 표면 diff 0, declRe/baselineRe 바이트 무변경, TestT528Anchor 216 직접 실행.
- AC-747-007: sibling 파일 diff-empty, 패키지 green.

## 결함

- **F1 [Medium][차단]** CHANGELOG.md:23 + progress.md §E.4 canary — 메커니즘 기술이 diff와 불일치("fixed scan window" 존재 안 함, 실제 변경은 declaration-bearing-region candidacy parser.go:143-200) + "TestT528Anchor, 14-file + 9-file probes"의 t528 오귀속(t747 결함 목록). 수치는 전부 정확. 수정: 문서-only 커밋으로 메커니즘 문장을 착지한 두 축으로 재기술 + t528 귀속 정정 — develop 통합 전.
- F2 [Medium][선택] §E.2 처분 기록 — 11건 중 4건은 실제 콜론 없는 AC bullet(CODERABBIT-ADOPTION:48-51, AGENT-MODEL-ROUTING:330, SKILL-COMPRESS:168, SKILL-CONSOLIDATE:230)이지 산문이 아님. leave-unrepaired 결정은 옳음(라인 문법 축 소관). 수정: 두 클래스로 재분류.
- F3 [Medium][선택] verdict 헤드라인 — 3 클래스 혼동: (i) end-to-end 기준 회복(+4 criteria/2파일), (ii) anchor-only 선택·parse 0(CC297/STATUS-AUTO — numeric-sub `AC-1.1` 미매치, lint 동일), (iii) strict-shape 측정 인공물(6 중 5 base==live). 수정: 양 클래스 병기.
- F4 [Low] I18N fallback가 Background 절의 인용성 AC-LCL-005를 흡수 — REQ 준수·CoverageRule 방향 안전. 원하면 증거 주석/후속 카드.
- F5 [Low] 프로브가 control-delta를 로그-only — t.Errorf 없음(t528 동일 형상). 원하면 단언화로 AC-747-003/004 내구 가드화.
- F6 [Low] AC-747-005 "860 일치"는 자기-측정 코퍼스에서 불가요구(실측 861=860+자기 spec.md) — verdict가 정직 공개. 향후 문구 권고.
- F7 [Info] TestT565AnchorVocabularyIsNotWide 이름 유지+역전 계약 — 실질 미약화(ColonlessProse 동반 양방향 고정).

## 권고

F1을 리드 batch push 전 doc-only 후속 커밋으로 착지(유일 차단 항목, 한 파일 문장군). F2/F3 재분류 문구를 같은 흐름에 편입(코드 변경 없음). Gaps: 전체 스위트·-race·CI(미push라 원본 판정 없음), `moai spec lint` 실측 diff(정적 분석으로 대체 — CoverageRule findings 감소만 가능). 잔여 위험: 비-vocab 제목 아래 `- AC-x:` bullet을 두는 향후 SPEC 재앵커 가능(오늘 코퍼스 비용 0); numeric-sub 파일의 "앵커 있음=AC 있음" 해석 변화 가능성.

— 본 파일은 sync-auditor가 전달한 보고를 lane이 보존한 것(하네스 규약상 감사자 미작성).
