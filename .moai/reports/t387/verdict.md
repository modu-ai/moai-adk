# t387 — 감사 판정문 곁말(side-talk) 규약 (verdict)

- **브랜치**: WT-audit-advice-integrity (develop ee1e0a1b6 흡수 + t386 브랜치 8925d89c7 흡수 — 의존 규칙에 따라 미병합 규약 문서를 이 트리에 흡수 후 편집)
- **카드**: t387 lane-10 (G2b 순차 2장째), Tier S

## Claim

감사 판정서에 붙는 곁말(운영 조언)의 형태를 규약으로 정한다: (나)안 — 별도 절 분리 + 미검증 명시 — 를 채택하고, 단절 금지 대신 "측정 지시 형태" 요구와 3단 상태 라벨(measured/inferred/assumption)로 보강. (가) 금지와 (다) 본문 동급 증거 요구는 기각.

## Evidence

### 실증 2건 재측정 (이 트리, 이번 실행)

| # | 카드 서사 | 재측정 |
|---|-----------|--------|
| 1 | iter1이 "44/2 발산은 스테일해지니 재인용 말고 재측정하라"고 경고 — lane-2가 재측정하니 불변, 규율은 옳고 결론이 반대 | `grep -n '44 / 2\|stale' .moai/reports/t377/plan-audit-iter1.md` → 425행 "- The 44 / 2 divergence figures go stale the moment any lane touches either" — 경고 존재 확인 |
| 2 | iter2가 ".moai/reports/는 gitignored라 유실되니 copy-out 필요" — 3중 반증 | `git check-ignore -v .moai/reports/t377/plan-audit-iter2.md` → 무출력 exit=1(무시 안 됨). `grep -n 'moai/reports' .gitignore` → 312행 `.moai/reports/*.md` 규칙 실재 — 인용한 규칙은 참, `*`가 `/`를 넘지 않아 t377/ 한 단계 아래는 못 잡음("규칙 참·추론 거짓" 형태 확인) |

### 채택 판정 근거

- (가) 기각: 실증 1이 "값한 경고가 곁말로 먼저 도착한다"는 반례 — 금지하면 다음 독자가 스테일 재인용 사고를 겪는다.
- (다) 기각: 곁말에 5섹션 증거를 요구하면 감사 비용이 본문과 동급이 된다 — 곁말의 역할은 본문 주제 밖 제안.
- (나) 채택 + 보강 2건:
  1. **측정 지시 형태**: "X는 거짓이다"는 검증 후에만; "X를 측정하라(명령 Y)"는 지시 형태는 검증 가능성을 운반 — 실증 1이 정확히 이 형태였고 그래서 살아남았다.
  2. **상태 라벨 3단**: `measured`(명령+출력 기록) / `inferred`(추론 규칙 명명) / `assumption`(나체 주장, 최약). "규칙은 참, 추론이 거짓"(실증 2)은 inferred 라벨이 규칙을 명명하게 해서 독자가 추론 축만 재판정하게 한다.

### 산출물

- `internal/template/templates/.moai/docs/audit-artifact-convention.md` — § Side-talk (advice attached to verdicts) 절 신규(§ What 다음), 119→154행
- `.moai/docs/audit-artifact-convention.md` — 미러, `diff -q` 동일 확인
- 중립성 스캔: 0매치(SPEC-ID/카드 id/SHA/날짜 패턴)
- 자기 적용: 절 마지막 문단이 "이 절 자신도 위 규칙 대상 — 실증은 카드 산출물에 기록, 여기서 주장하지 않음"을 명시(카드 주의 사항 반영)

## Baseline-attribution

- 측정 트리: WT-audit-advice-integrity (ee1e0a1b6 + 8925d89c7 흡수 후)
- iter1:425행 / .gitignore:312 / check-ignore exit=1 — 모두 이번 실행 직접 관측

## Gaps (미검증)

- plan-auditor.md·sync-auditor.md에 곁말 규약 반영은 **보류** — t386 보류 항목(반출 조항)과 동일하게 lane-9 t367·t302 확정 후 리드 재개 신호 때 함께 반영할 것. 규약 문서가 형식 SSOT이므로 조항 반영 전에도 규약은 유효하나, 감사자에게 도달하는 경로는 에이전트 정의뿐이라 실질 집행력은 반영까지 유보.
- t377 이슈에 적혔다는 감사자 자체 철회 기록("자기가 제기한 N1과 같은 모양")은 이 트리에 없어 관측 불가 — 카드 서사 인용 수준.

## Residual-risk

- 라벨의 정직성은 감사자 자기 보고에 의존 — `measured` 라벨 + 명령 기록 부재 조합을 잡는 기계적 가드는 없음(규약이 "mislabel, not evidence"로 정의하긴 함).
- 곁말과 본문의 경계는 주제 기반이라 정량 판정 불가 — 판정 축: 본문=판정·점수·결함·증거, 그 외=곁말.

## 다음

- 리드가 본 판정서를 읽고 카드 종결 판정
- lane-9 재개 신호 → 에이전트 정의 2건 반영(t386 반출 조항 + t387 곁말 규약 함께)
