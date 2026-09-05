# SPEC-PLAN-AUDITOR-RESIDUE-001 — run-phase raw command outputs (card t450)

측정 트리: 워크트리 t450, 브랜치 WT-plan-auditor-residue.
시작 HEAD `194fc9439` (develop `6765a75c0` 흡수 머지 커밋 — `git merge-base --is-ancestor 6765a75c0 HEAD` → DEVELOP-ABSORB-OK).
전제 재검증: `git merge-base --is-ancestor 18fc2c9ef develop` → exit 0 (T367-ANCESTOR-OK).
기준 트리(acceptance.md RED-now): 7835148d3 — RED는 그 트리에서 이미 측정돼 있고, 아래는 GREEN 측(커밋 전 작업 트리) 재측정값이다.

## AC-001 — 반출 조항 존재

```
$ grep -c "Export mandate" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
internal/template/templates/.claude/agents/moai/plan-auditor.md:1
.claude/agents/moai/plan-auditor.md:1
종료코드: 0
```

조항의 반출 위치 문면(양쪽 동일): `.moai/reports/<card-id>/plan-audit.md` (또는 `plan-audit-iter<N>.md`; `.moai/reports/<SPEC-ID>/`).

## AC-002 — 금지 경로 제거

```
$ grep -n "reports/plan-audit/" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
(출력 없음)
종료코드: 1
```

RED-now(:395 매치, exit 0)에서 뒤집힘. 금지 디렉터리 언급은 경로 리터럴 없이 규약 참조로 대체됐다 — "the report directory the convention declares FORBIDDEN (`audit-artifact-convention.md` § Where)".

## AC-003 — 곁말 규약 반영

```
$ grep -c "Side-talk" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
internal/template/templates/.claude/agents/moai/plan-auditor.md:1
.claude/agents/moai/plan-auditor.md:1
종료코드: 0

$ grep -o "measured\|inferred\|assumption" .claude/agents/moai/plan-auditor.md | sort | uniq -c
   3 assumption
   1 inferred
   1 measured

$ grep -o "measured\|inferred\|assumption" internal/template/templates/.claude/agents/moai/plan-auditor.md | sort | uniq -c
   3 assumption
   1 inferred
   1 measured
```

(assumption 3 = 조항 1회 + 파일 기존 문면 2회. 3라벨 모두 존재.)

## AC-004 — 조항 쌍둥이 일치

```
$ diff .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
338c338
< - D7-1: ... (supports multi-segment IDs like SPEC-DOMAIN-WO-001)
---
> - D7-1: ... (supports multi-segment IDs like SPEC-EXAMPLE-DOMAIN-001)
443c443
< This agent is invoked by the orchestrator up to a Tier-resolved number of times per SPEC. ...
---
> This agent is invoked by the orchestrator up to 3 times per SPEC (max_iterations: 3 per harness.yaml).
445,446d444
< **Ceiling bounds ITERATION COUNT, NOT verdict.** ...
종료코드: 1
```

전체 diff는 기존 드리프트 2 hunk(D7-1 예시 식별자, Tier-resolved ceiling 문단 — 판정 대상에서 제외됨)만 보인다. 조항 범위(clause-scoped) diff:

```
$ sed -n '/Export mandate/,/record surface only/p' .claude/agents/moai/plan-auditor.md > /tmp/t450-clause-local.txt
$ sed -n '/Export mandate/,/record surface only/p' internal/template/templates/.claude/agents/moai/plan-auditor.md > /tmp/t450-clause-tmpl.txt
$ diff /tmp/t450-clause-local.txt /tmp/t450-clause-tmpl.txt
AC-004-CLAUSE-SCOPED-DIFF-0
```

## AC-005 — 교차참조 일치

(a) plan-auditor 사본의 판정문 기록처 참조 소멸:

```
$ grep -c "reports/plan-audit/" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
.claude/agents/moai/plan-auditor.md:0
internal/template/templates/.claude/agents/moai/plan-auditor.md:0
종료코드: 1
```

(b) spec-workflow.md § Report Persistence(양쪽 미러 바이트 동일 — `diff` 빈 출력 확인)는 review stream을 반출 패밀리(`plan-audit.md` / `plan-audit-iter<N>.md` → `.moai/reports/<card-id>/` 또는 `/<SPEC-ID>/`)로 서술하고 run-gate stream만 gitignored 런타임 기록 디렉터리로 서술한다. audit-artifact-convention.md § What makes the convention stick의 감사자 측 서술("The plan-auditor and sync-auditor agent definitions carry the export as a HARD completion condition")은 AC-001로 참이 됐다. § Cross-references는 plan-auditor 반출 패밀리를 § Where로 명시 조인했다.

## AC-006 — 방출물 + 카탈로그 해시 재생성

재생성 전 기선 (기준 트리와 동일한 RED, 이번 실행 재측정):

```
$ go test ./internal/template/agentemit/...
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.01s)
    golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.745s
```

실제 메커니즘 전체 재생성(REQ-008, t367 선례 2549f775f):

```
$ AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.515s
```

sync-auditor.toml 복원 후 최종 상태:

```
$ git restore --source=HEAD -- internal/template/templates/.codex/agents/moai/sync-auditor.toml
$ shasum -a 256 internal/template/templates/.codex/agents/moai/sync-auditor.toml internal/template/templates/.codex/agents/moai/plan-auditor.toml
5306b92eaccbf9dff2c6cb4149f0182760071289ceedfc1e2be57d8e0707f905  internal/template/templates/.codex/agents/moai/sync-auditor.toml   # == develop 값(HEAD) — byte-identical
e9408a40d7019fc39501ac51cb8bbb5e304ff23d9f8bc8993500371236cbc6d0  internal/template/templates/.codex/agents/moai/plan-auditor.toml    # 재생성됨

$ git diff develop --stat -- internal/template/templates/.codex/agents/moai/sync-auditor.toml
(빈 출력 — develop과 차이 0)
```

적색 한정 확인 (AC-006 합격 판정: 적색이 sync-auditor.toml 1건으로 한정):

```
$ go test ./internal/template/agentemit/...
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.01s)
    golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	2.562s
```

카탈로그 갱신(정확 기수는 `go run ./internal/template/scripts/gen-catalog-hashes.go` — Makefile:35의 실제 경로, `--entry` 플래그는 `--help`로 확인):

```
$ go run ./internal/template/scripts/gen-catalog-hashes.go --entry plan-auditor
Computing hash for entry "plan-auditor"...
  plan-auditor: 2403bfb3e0747467ba328cdbfe3767c0801db95702568d5cb606596655d771fb
catalog.yaml updated successfully (12899 bytes)

$ git diff internal/template/catalog.yaml   # 요지
-              hash: efb7167ded27512474d993d76099162253c4bef550c557dd6b0849644e9bbfb8
+              hash: 2403bfb3e0747467ba328cdbfe3767c0801db95702568d5cb606596655d771fb
```

카탈로그 해시 유효성(internal/template — TestManifestHashFormat):

```
$ go test ./internal/template/ -run 'TestManifestHashFormat' -v
    catalog_tier_audit_test.go:451: CATALOG_HASH_UNSTABLE: sync-auditor stored hash=f1b4487f...,
      computed hash=545d03d9... (source=.claude/agents/moai/sync-auditor.md)
    catalog_tier_audit_test.go:456: audited 45 catalog entries for hash validity
--- FAIL
```

이 FAIL은 **선존재 상속 적색**이다 — 비교 양변(catalog 저장 해시 f1b4487f…, 임베디드 소스 해시) 모두 HEAD 커밋 바이트다(`git diff --stat HEAD -- .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` → 빈 출력, 본 실행 미터치). 4244c4a06 이래의 sync-auditor 카탈로그 스테일 — t443/t444 소관, record-and-not-repair. plan-auditor 항목은 안정(불안정 발견 1건뿐) — AC-006의 "plan-auditor는 녹색" 충족.

## AC-007 — t367 :72 문면 보존

```
$ grep -c "the fifth GEARS pattern" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
.claude/agents/moai/plan-auditor.md:1
internal/template/templates/.claude/agents/moai/plan-auditor.md:1
$ grep -c "NOT a GEARS pattern" .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
.claude/agents/moai/plan-auditor.md:1
internal/template/templates/.claude/agents/moai/plan-auditor.md:1
```

편집 전후 동일(양쪽 각 1) — 보존 판정 통과.

## AC-008 — 템플릿 중립성 0매치

신규 템플릿 문면(템플릿 plan-auditor.md :395-397) + 템플릿 편집 문서 2종 대상 금지클래스 grep(SPEC-ID 일반형 / 카드 id / 40hex SHA / 내부 날짜):

```
$ grep -cE "SPEC-[A-Z][A-Z0-9]+-[0-9]{3}" /tmp/t450-new-templ-lines.txt   → 0
$ grep -cE "\bt450\b"                      /tmp/t450-new-templ-lines.txt   → 0
$ grep -cE "[0-9a-f]{40}"                  /tmp/t450-new-templ-lines.txt   → 0
$ grep -cE "20[0-9]{2}-[0-9]{2}-[0-9]{2}"  /tmp/t450-new-templ-lines.txt   → 0
$ grep -E "SPEC-[A-Z][A-Z0-9]+-[0-9]{3}|\bt450\b|[0-9a-f]{40}|20[0-9]{2}-[0-9]{2}-[0-9]{2}" \
    internal/template/templates/.moai/docs/audit-artifact-convention.md \
    internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md | wc -l   → 0
```

전트리 가드:

```
$ go test ./internal/template/ -run 'TestTemplateNeutralityAudit$'    → ok
$ go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak$' → ok
$ go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift'     → ok   # spec-workflow.md 미러 포함
$ go build ./...                                                      → BUILD-OK (docs-only 레인 검증 범위)
```

## 이번 실행 미수리 기록 (record-and-not-repair)

1. agentemit 골든 적색 1건 — sync-auditor.toml sha256 mismatch: t443 소관(REQ-008, t367 선례 2549f775f).
2. TestManifestHashFormat CATALOG_HASH_UNSTABLE 1건 — sync-auditor 저장 해시 스테일(4244c4a06 이래): t443/t444 소관. 본 실행이 만들거나 고친 것이 아니다(양변 HEAD 바이트).
3. 기존 쌍둥이 드리프트 2 hunk(D7-1 예시 식별자, Tier-resolved ceiling 문단): SPEC Out of Scope — 기록만.
4. sync-auditor.md 곁말 반영: 별도 카드 소관(Out of Scope).
