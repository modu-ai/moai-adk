# SPEC-PLAN-AUDITOR-RESIDUE-001 — 수용 기준

기준 트리: 7835148d3 (origin/develop과 동일 — 이번 실행 `git rev-parse --short HEAD` 관측). 모든 RED-now 셀은 이 트리에서 이번 실행 직접 측정한 값이다. 커밋 시점에 재측정해 갱신한다.

## §D AC Matrix

| AC | 내용 | REQ | blocking 여부 | 검증 |
|----|------|-----|---------------|------|
| AC-001 | 반출 조항 존재(양쪽 사본) | REQ-001 | release-blocking | grep |
| AC-002 | 금지 경로 제거(양쪽 사본) | REQ-002 | release-blocking | grep |
| AC-003 | 곁말 규약 반영(양쪽 사본) | REQ-003 | release-blocking | grep |
| AC-004 | 조항 쌍둥이 일치 | REQ-004 | release-blocking | diff |
| AC-005 | 교차참조 일치 | REQ-005 | release-blocking | grep |
| AC-006 | 방출물 + 카탈로그 해시 재생성 | REQ-006, REQ-008 | release-blocking | go test |
| AC-007 | t367 :72 문면 보존 | REQ-007 | 회귀 가드 | grep |
| AC-008 | 템플릿 중립성 0매치 | REQ-006 | 회귀 가드 | grep |

## AC 상세

### AC-001 — 반출 조항 존재

**Given** plan-auditor.md 양쪽 사본이 존재하고 **When** 반출 조항 마커 문구("Export mandate")로 양쪽 사본을 검색하면 **Then** 양쪽 사본 각각에서 1회 이상 매치되고, 조항은 `.moai/reports/<card-id>/plan-audit.md` (또는 `plan-audit-iter<N>.md`)를 반출 위치로 명시한다.

RED-now (2026-09-03, 트리 7835148d3):

```
명령: grep -c "Export mandate" .claude/agents/moai/plan-auditor.md
출력: (빈 출력 — 매치 0)
종료코드: 1
```

### AC-002 — 금지 경로 제거

**Given** `.moai/reports/plan-audit/`는 gitignored 위치이고(`.gitignore:230`, `git check-ignore` exit=0 관측) **When** 양쪽 plan-auditor.md 사본에서 쓰기 대상 경로로서의 `.moai/reports/plan-audit/`을 검색하면 **Then** 매치가 0이다 (§ Output Format의 경로 지시가 REQ-001의 반출 위치로 교체됐다).

RED-now (2026-09-03, 트리 7835148d3):

```
명령: grep -n "reports/plan-audit/" .claude/agents/moai/plan-auditor.md
출력: 395:Write the audit report to `.moai/reports/plan-audit/{SPEC-ID}-review-{iteration}.md`.
종료코드: 0
```

### AC-003 — 곁말 규약 반영

**Given** `audit-artifact-convention.md` § Side-talk이 곁말 3규칙(별도 미검증 절 / 측정 지시 형태 / 상태 라벨 measured·inferred·assumption)을 정의하고 **When** 양쪽 plan-auditor.md 사본에서 "Side-talk" 또는 상태 라벨 3종(`measured`, `inferred`, `assumption`)을 검색하면 **Then** 양쪽 사본 모두에서 곁말 조항이 발견되고 3개 라벨이 모두 등장한다.

RED-now (2026-09-03, 트리 7835148d3):

```
명령: grep -n "Side-talk" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
출력: (빈 출력 — 매치 0)
종료코드: 1
```

### AC-004 — 조항 쌍둥이 일치

**판정 대상(도구 모호성 제거)**: 이 AC가 판정하는 것은 **조항 범위(clause-scoped) diff**다 — AC-001~003의 조항 구간만 추출해 비교하며 차이가 0이어야 한다. 전체 파일 diff는 기존 드리프트 2 hunk(D7-1 예시 식별자, Tier-resolved ceiling 문단)를 적법하게 유지할 수 있고, 그 2 hunk의 존재는 이 AC의 실패가 아니다 — 전체 파일 diff는 RED 관측 수단으로만 쓴다.

RED-now (2026-09-03, 트리 7835148d3):

```
명령: diff .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
출력: 338c338 (D7-1 예시 식별자 hunk), 441c441 (Tier-resolved ceiling 문단 hunk), 443,444d442 — 계 2 hunk
종료코드: 1
```

RED-now 보조 기준선 — 조항 구간 자체가 양쪽 모두 부재 (2026-09-03, 트리 7835148d3):

```
명령: grep -c "Side-talk" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
출력: .claude/agents/moai/plan-auditor.md:0
      internal/template/templates/.claude/agents/moai/plan-auditor.md:0
종료코드: 1
```

**Green 경로**: M2가 조항을 양쪽 사본에 동일 문면으로 적용하면 이 AC가 뒤집힌다. 합격 출력 = 조항 범위 diff 0 (조항 구간 추출 비교에서 차이 없음). 이때 전체 파일 diff는 위 2 hunk를 여전히 보일 수 있다 — 적법하며 판정에서 제외된다(§ 판정 대상 참조).

### AC-005 — 교차참조 일치

**판정 대상(녹색 형태 명시)**: 이 AC가 측정하는 것은 **조항**이다 — plan-auditor 쌍둥이 사본에 `.moai/reports/plan-audit/`을 판정문 기록처(verdict destination)로 참조하는 문면이 남아 있지 않은가. 디렉터리 자체의 비움 여부는 판정 대상이 아니다: 런게이트 스트림이 그 디렉터리에 적법하게 남으므로, 저장소 전체 grep에서 `.moai/reports/plan-audit/`이 여전히 매치되는 것은 이 AC의 실패가 아니다.

**Given** 규약 문서와 룰 문서가 plan-auditor의 반출 경로를 서술하고 **When** (a) 양쪽 plan-auditor.md 사본에서 `reports/plan-audit/`을 검색하면 **Then** 양쪽 모두 매치 0이다 (판정문 기록처 참조 소멸). **When** (b) `spec-workflow.md` § Report Persistence(양쪽 미러)와 `audit-artifact-convention.md` § What makes the convention stick(양쪽 미러)의 plan-auditor 관련 서술을 판독하면 **Then** 그 문서들이 plan-auditor의 쓰기 대상을 `.moai/reports/plan-audit/`으로 지정하지 않고, `audit-artifact-convention.md`의 감사자 측 서술("The plan-auditor and sync-auditor agent definitions carry the export...")은 참이 된다.

RED-now (2026-09-03, 트리 7835148d3):

```
명령: grep -n "reports/plan-audit/" .claude/agents/moai/plan-auditor.md
출력: 395:Write the audit report to `.moai/reports/plan-audit/{SPEC-ID}-review-{iteration}.md`.
종료코드: 0
```

### AC-006 — 방출물 + 카탈로그 해시 재생성

**Given** 템플릿 쪽 plan-auditor.md가 바뀌었고 **When** plan-auditor.toml을 범위 재생성하고 카탈로그 해시를 갱신한 뒤 `go test ./internal/template/agentemit/...`을 실행하면 **Then** sync-auditor.toml 1건의 sha256 불일치만 남는다(t443 소관 기록) — plan-auditor.toml은 골든 패리티를 통과하고, TestCatalogHashParity·TestManifestHashFormat에서 plan-auditor는 녹색이다.

RED-now (2026-09-03, 트리 7835148d3):

```
명령: go test ./internal/template/agentemit/...
출력: --- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
      golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
종료코드: 1
```

주의: 이 RED의 원인 자체는 t443 소관이라 이 SPEC이 고치지 않는다. AC-006의 합격 판정은 "plan-auditor 항목 녹색 + 적색이 sync-auditor.toml 1건으로 한정"이며, 전체 녹색은 아니다 — REQ-008의 record-and-not-repair다.

### AC-007 — t367 :72 문면 보존 (회귀 가드)

**Given** t367(18fc2c9ef)이 착지시킨 루브릭 문면이 양쪽 사본에 있고 **When** 양쪽 사본에서 Event-detected 불릿과 Unwanted legacy-only 불릿을 검색하면 **Then** 편집 전후로 양쪽 사본에서 각각 1회씩 매치가 유지된다.

GREEN-now (2026-09-03, 트리 7835148d3): `grep -c "the fifth GEARS pattern"` 양쪽 사본 각 1 (`18fc2c9ef` diff로 문면 확인). 이 AC는 RED-now가 아니라 보존 판정이다 — run-phase에서 편집 전후 동일 값을 기록.

### AC-008 — 템플릿 중립성 0매치 (회귀 가드)

**Given** 템플릿 쪽 편집이 완료됐고 **When** 편집된 템플릿 파일들에 중립성 금지 패턴(SPEC-ID 일반형, 카드 id, 커밋 SHA, 내부 날짜 — §25.1 C1-C8 카탈로그)을 적용하면 **Then** 신규 편집분에서 매치가 0이다. 형제 가드 `internal_content_leak_test.go`(C3 날짜 + C7 해시)도 통과한다.

## §D.1 심각도 정의

- critical: release-blocking AC 실패 — 반출 조항 부재, 금지 경로 잔존, 방출물 불일치
- major: 조항 쌍둥이 불일치 — 즉시 수리
- minor: 문안 표현 차이(의미 동일)

## §D.2 간접 검증

- AC-004는 직접 diff로, AC-006은 골든 테스트 출력으로 간접 확인 — 수동 판독으로 대체하지 않는다.

## §D.3 종결 게이트

- 8개 AC 전부 통과 + M1~M4 완료 + progress.md §E.1 기록 + 증거 브랜치 커밋.

## §D.4 품질 게이트 (TRUST 5)

- Tested: agentemit 골든 + 카탈로그 패리티 테스트 출력 인용 (AC-006)
- Readable: 조항 문안은 sync-auditor 착지 문면의 문체 선례와 일치
- Unified: 쌍둥이 diff 0 (AC-004)
- Secured: 해당 없음 (문서 변경)
- Trackable: 커밋마다 카드 id(t450) 명시, Conventional Commits

## §D.5 Definition of Done

- REQ-001~008 전부 충족, AC-001~008 전부 통과(AC-006은 REQ-008의 한정 판정 포함), 양쪽 미러 커밋, 증거 기록, plan-auditor 실사용 경로(다음 plan 감사)에서 반출 파일이 카드 디렉터리에 생성되는 것으로 행위 확인.
