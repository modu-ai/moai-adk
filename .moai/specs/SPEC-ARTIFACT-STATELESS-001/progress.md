# SPEC-ARTIFACT-STATELESS-001 — 진행 기록

카드: t357 · Tier M · 브랜치 `WT-tierl-status-transitions`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 계약. `design.md`/`research.md`는 Tier M 산출물이 아니며, 이 SPEC의 주제상으로도 만들지 않는다)
- 운영자 결정: **안 C(무상태 선언)**. 정의는 **D1(status 라인만 제거)**, 모집단은 **전 코퍼스 696**
- REQ 15 / AC 11 (Tier M 상한 각 16)

### 감사 이력

| 회차 | 판정 | 점수 | 결과 |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.86 (Tier M 임계 0.80) | blocking 6건(D1~D6) + optional D9 → v0.2.0에서 전부 반영 |

감사 보고: `.moai/reports/t357/plan-audit.md` (감사 트리 `3b1830b96`)

### 측정 baseline — 두 트리가 섞여 있다

| 측정 | 트리 | 값 |
|---|---|---|
| 초기 코퍼스 실측 (696 / 633 / 362 / 417 / 170 / 106 / 12) | `c6aa61346` | `.moai/reports/t357/plan-measurement.md` |
| D1 전 코퍼스 (채택 모집단) | `3b1830b96` | **389** = design 27 / research 34 / plan 164 / acceptance 164 |
| D1 종결-한정 (참고값, 389의 부분집합) | `3b1830b96` | 362 |
| 템플릿 미러 바이트 동일성 | `3b1830b96` | 동일 (양쪽 23,317 bytes) |

재측정 명령: `bash .moai/reports/t357/t357_d1_all.sh .` · `bash .moai/reports/t357/t357_d1_by_artifact.sh .`

### AC 착지 전 FAIL 증거 (v0.2.0 개정 시 실행, 트리 `3b1830b96`)

`bash .moai/reports/t357/t357_ac_precheck.sh .` 출력:

```
section bytes = 0
S1 spec.md-only:      FAIL
S2 four-artifacts+status: FAIL
S3 Tier-independent:  FAIL
P  permission stated: FAIL
N  blanket-prohibition matches = 0 (must be 0, AND P must be PASS)
mirror identical: yes
anchor in local: FAIL (section empty)
anchor in mirror: FAIL
```

`moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md` → `✓ No findings`, rc=0, `ArtifactStatusFieldForbidden` 매치 **0** (AC-04 비공허성 가드가 실제로 0을 낸다).

### 기준값 — run-phase에서 채운다

`acceptance.md`의 `AC-AST-001-07` / `-08` / `-10`이 이 표를 `sed`로 읽는다. 형식(`| <이름> | \`<값>\` |`)을 바꾸지 않는다.

| 시점 | 값 |
|---|---|
| SPEC 착수 직전 | `` |
| M3 착수 직전 | `` |
| M3 착수 시 D1 baseline N | `` |

[HARD] **빈 슬롯은 AC를 통과시키지 않는다.** 추출은 `\{7,\}`(SHA) / `\{1,\}`(숫자) 패턴 + `-n` 검사 + `git rev-parse --verify` 3중 가드를 거치며, 빈 슬롯은 stderr에 `FAIL — 「…」 슬롯이 비어 있거나 …`를 내고 exit 1 한다. 가드 없이 `[0-9a-f]*`로 읽으면 빈 슬롯이 매치되어 빈 문자열을 캡처하고, `""..HEAD`가 `HEAD..HEAD`로 조용히 해석돼 AC가 공허하게 PASS한다 — iter-2 감사 N1이 지적한 결함이다.

- `SPEC 착수 직전` / `M3 착수 직전`: 7자리 이상 hex SHA
- `M3 착수 시 D1 baseline N`: `bash .moai/reports/t357/t357_d1_all.sh .` 의 「D1 전체 696 모집단」 값 (착수 시점 참고: 389 @ `3b1830b96`)

### 미해결 Gap

1. 카드가 인용한 `SPEC-AC-COUNT-DISCRIMINATOR-001`이 develop 코퍼스에 없어 원 사례 미재현
2. `origin/develop`이 `48d8ef4be`로 26 커밋 전진 — M3 착수 전 재측정 필수(REQ-AST-001-009)
3. 배포 사용자 코퍼스의 기존 위반 여부는 이 리포에서 측정 불가 — lint 심각도 결정의 잔여 위험(`plan.md` §B3/§D)
4. 362를 도출한 두 경로가 `fm_of` 추출기를 공유하므로 추출 단계 실패 모드는 교차검증되지 않음(spec.md §1.6)

다음: plan-audit 재감사(델타 범위) → Implementation Kickoff Approval → run

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
