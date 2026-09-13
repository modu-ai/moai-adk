# t747 plan-phase 측정 — findACSectionStart 앵커의 in-section 계수 크기·방향

card: t747 · branch: `WT-ac-anchor-scope` · base: local develop `4da5d1c4e` · 측정일: 2026-09-14
프로브: `anchor_scope_probe_test.go`(본 디렉터리, t528 프로브 방법론 계승 — discriminator B `declRe` 동일, 분모 filelist.txt 고정)
재유도: `T747_PROBE_OUT=<abs> go test ./internal/spec/ -run TestT747AnchorScope -v -count=1`

## 측정 결과 (라이브 코퍼스, 기본 환경)

```
DENOMINATOR spec.md read = 860 (filelist.txt 고정)
files with >=1 declaration (whole file) = 152
files anchored by findACSectionStart = 462
files with declarations but NO anchor (narrow-miss) = 14   (anchor-missed.txt)
files anchored but 0 in-section while decls exist elsewhere (empty anchor) = 9   (empty-anchor.txt)
  of which the empty-first-section fallback likely took the anchor = 9
declarations WHOLE FILE = 1405
declarations IN-SECTION (current anchor) = 1240
declarations OUT-SECTION = 165   (outside-decls.txt)
```

t528 기준선 대비: in-section 계수는 t528 당시 분모 815 / in-section 1167 / 동결 앵커 수용 216 → 현재 분모 860 / in-section 1240(라인형 widening 이후). 이번 측정의 새 사실은 앵커(절 시작) 차원의 것:

## 방향 1 — 좁음 (앵커 부재): 14파일, 앵커 없음 + decls 존재

`anchor-missed.txt` 전수 — 대표: SPEC-AC-COLLECTOR-ANCHOR-001 자신(4 decls, vocab 어휘에 안 걸리는 제목), SPEC-CC297-001(19), SPEC-STATUS-AUTO-001(25). 어휘 리스트(`acSectionVocabulary`)+수준≥2 헤딩 조건을 통과하는 절이 없어 파서가 AC 절을 통째로 못 본다. guard 근거(lint_coverage_sibling.go:29-32)의 "spec.md는 혼합 문서" 원칙상 이 14파일의 전부가 수리 대상은 아니고, decl 형태(bullet+AC id)를 갖춘 실 AC 절이 vocab 미등재 제목 아래 있는 경우가 수리 후보.

## 방향 2 — 헐거움 (빈 앵커 선점): 9파일, fallback이 첫 vocab 절을 앵커

`empty-anchor.txt` 전수 — 앵커 제목은 오히려 정확해 보이는 "## Acceptance Criteria"이나 in=0이고 decls는 후행 절에 1~28개. 즉 파일 안에 vocab 제목 절이 둘 이상이고 첫 절이 비어 있으며, 비-빈 절 선호 루프가 모두 실패한 뒤 fallback(`first`)이 빈 첫 절을 앵커한다. 실제 AC는 대개 후행의 다른 제목(예: 요구사항 병합 절 내부, 후속 상세 절) 아래 있음. 9파일 모두 fallback 경유로 측정됨.

## 나머지 out-section 165의 성격 (SPEC 설계 입력)

in=10/out=1 꼴(CLAUDEMD-DIET-V2, DB-SYNC-HARDEN 등)은 산문 속 AC id 언급이 declRe에 걸린 것으로 보이는 올바른 배제일 수 있다 — 165 전체가 좁음 피해는 아니며, 혼합 문서 원칙이 지키려던 바가 정확히 이 층. 수리 설계는 (a) 앵커 부재 14파일과 (b) 빈 앵커 9파일을 수리하면서 (c) out-section 산문층을 새로 흡수하지 않는 선을 명시해야 한다. 샘플 판독은 run 단계 검증 설계의 입력으로 남긴다.

## 범위 메모

- in-section 스캔은 t528 프로브와 동일 형태(모든 `##` 접두 행에서 break)라 216/1167과 비교 가능. `extractACLines`의 앵커-수준 break(### 하위 절 포함 독해)와의 2차 차이는 여기서 해소하지 않고 기록만.
- declRe는 discriminator B의 동결 사본 — 편집 금지.
- 프로브 소스: `probe/anchor_scope_probe_test.go` — 본래 미커밋 측정 도구(zz_t747_anchor_probe_test.go, 커밋 전 삭제)였으나 plan-audit iter1 D4에 따라 내구성을 위해 커밋됨(8665f80b3). 트리 상수 프로브(zz_t528 패턴)로의 승격 여부는 run 단계 M1의 판단이며, 이 커밋 자체가 승격을 뜻하지 않는다.
