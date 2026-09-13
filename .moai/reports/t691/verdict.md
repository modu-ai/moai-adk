# Card t691 — Verdict

- card: t691 (GH #1682 — moai glm의 CLAUDE_CODE_HARBOR_KITE 주입 판정)
- branch: WT-harbor-kite-judgment (base: develop e7b93c120)
- evidence: .moai/specs/SPEC-HARBOR-KITE-001/ + .moai/reports/t691/{verdict,plan-audit,issue-reply-draft}.md
- date: 2026-09-13, lane-6 · Class C, 리드 지시로 plan(판정)까지만 — run 미승인

## Claim (주장)

1. **판정: (b) 런처 비주입 유지.** 근거 G1 상류 비공개 플래그(이름·의미 변동 위험, 묶인 기본값의 침묵 변질) · G2 사용자 기기 전역 상태 덮어쓰기 · G3 공급자 간 일관성(gateway GPT 세션도 슬롯 비기록자 — glm만 주입은 비일관, 전체 주입은 G2 증폭) · G4 유실이 제한적(nudge 층)이며 큐-온-디스크 위임 채널은 무관하고 자가 치유됨.
2. **카드의 [HARD] 재현을 행동 수준에서 완료했다**: 슬롯 하나만 다른 두 격리 config에서 같은 바이너리·같은 z.ai 경로의 헤드리스 세션이 — true: ListAgents 존재("peer messaging itself is available") / false: ListAgents가 툴셋에서 제거("There is no tool named ListAgents in my available toolset"). 실가동 ~/.claude.json(현재 true)은 건드리지 않았다.
3. plan-audit CONCERNS(0.81 — Tier M 임계 0.80 충족, must-pass 7/7)의 차단 결함 3건(D1 AC-004 doc-mirror 오탐·D2 REQ-004 무AC·D3 초안의 미측정 주장)을 수정하고 오케스트레이터가 재검증 통과했다.
4. 이슈 #1682 한국어 답변 초안이 완비돼 있다 — 게시는 리드 몫.

## Evidence (증거 — 이번 run에서 직접 관측)

- **재현 셀**: `CLAUDE_CONFIG_DIR=$HOME/.moai/claude-profiles/t691-true claude -p "Call the ListAgents tool…"` → 도구 실행, "peer messaging itself is available". 동일 명령의 t691-false → "The ListAgents tool is not available. There is no tool named `ListAgents` in my available toolset (…)". 단일 변수(CLAUDE_CONFIG_DIR의 슬롯값) 통제, 2026-09-13.
- **t400 문서 측정 인용**: 슬롯 플립 4회 기동 — true→소켓 2회, false→없음 2회 (cross-session-messaging-detail.md § The shared flag slot).
- **감사 재검증(오케스트레이터 손)**: `git grep -n "CLAUDE_CODE_HARBOR_KITE" -- '*.go' ':!internal/template'` → exit 1 (0히트, doc-mirror 제외 확인) · AC-005 존재 · 초안 29행에 측정범위 문구("…검증하지 않았습니다") 확인 · `moai spec lint SPEC-HARBOR-KITE-001` → No findings.

## Baseline-attribution (baseline 귀속)

- 브랜치 WT-harbor-kite-judgment, base develop e7b93c120 — 이번 run, 이 워크트리.
- 재현 셀은 claude 2.1.270 + z.ai(glm-flash) 경로, /tmp/t691-harbor 스크래치 cwd에서 실행.

## Gaps (미검증 — 명시적)

1. 플래그 평가 자체가 꺼진 환경(DISABLE_TELEMETRY 계열)에서 이 env의 동작 — 미측정. 초안은 측정범위 문구로 한정됨.
2. 재현은 claude 직접 실행(moai 래퍼 생략) — 래퍼의 기여는 CLAUDE_CONFIG_DIR 재지향뿐임을 코드로 확인하고 복제했으나, 래퍼 경유 end-to-end는 아님.
3. 카드 원문의 "두 세션 ListAgents 가시성" 대신 단일 세션 도구-가용성 프로브로 대체 — 격리는 강해지지만 가시성 형상 자체는 미관측.
4. 이슈 게시·마감, 프로필 폴더 정리(~/.moai/claude-profiles/t691-true·false)는 리드/후속 몫.

## Residual-risk (잔여 위험)

- 상류가 플래그 이름·의미를 바꾸면 본 판정의 전제가 흔들린다 — 그때 재판정.
- 슬롯은 자가 치유되므로 유실이 재발-소멸을 반복한다 — 사용자 혼란의 여지, doctor 감지 카드로 완화 가능.
- SPEC은 status: draft로 남는다(구현 없는 판정 카드) — 종결 경로는 리드 결정.

## 후보 카드 (리드 판정)

1. `moai doctor` 감지: 서드파티 세션 + 슬롯 false → 문서화된 escape hatch(CLAUDE_CODE_HARBOR_KITE=1) 조언 — 관측성 제공, 강제 없음.
2. (t672 판정서에서 이월) 웨지 안전 재뿌리 정책.
