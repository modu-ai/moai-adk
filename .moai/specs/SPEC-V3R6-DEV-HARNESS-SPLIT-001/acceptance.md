# Acceptance — SPEC-V3R6-DEV-HARNESS-SPLIT-001

> 각 AC는 Given-When-Then + binary 검증 명령(grep/find/test/파일 존재)을 갖는다.
> 7개 scope 항목 전부 커버.

## §D. AC Matrix

| AC | REQ | 검증 방식 |
|----|-----|-----------|
| AC-DHS-001a | REQ-DHS-001 | 3개 thin command 파일 존재 |
| AC-DHS-001b | REQ-DHS-001 | per-command surface가 argument-hint에 선언됨 |
| AC-DHS-002a | REQ-DHS-002 | 3개 specialist가 새 이름으로 존재 + 옛 이름 부재 |
| AC-DHS-002b | REQ-DHS-002 | specialist frontmatter `name` 갱신 + self-reference 갱신 |
| AC-DHS-003a | REQ-DHS-003 | 통합 devkit.md + manifest.json 부재 |
| AC-DHS-003b | REQ-DHS-003 | Runner가 release-update 전용으로 rename + 재범위화 |
| AC-DHS-004 | REQ-DHS-004 | CI guard 재지향 + stale `devkit` 토큰 0 (파일명/함수/sentinel rename, RED→GREEN) |
| AC-DHS-005 | REQ-DHS-005 | dev-only 누출 invariant 보존 |
| AC-DHS-006 | REQ-DHS-006 | 5개 doctrine surface에서 devkit 참조 제거 (standalone title/header 포함) |
| AC-DHS-007 | REQ-DHS-007 | (sync-phase) 메모리 갱신 — plan/run-phase에서는 미작성 확인 |

---

## §D.1 — AC 상세 (Given-When-Then)

### AC-DHS-001a — 3개 독립 thin command 파일 존재

- **Given** devkit 단일 진입이 3개로 분리됨
- **When** run-phase 완료 후
- **Then** 3개 thin command 파일이 존재한다

```bash
test -f .claude/commands/harness/release-update.md \
  && test -f .claude/commands/harness/github.md \
  && test -f .claude/commands/harness/release.md \
  && echo PASS || echo FAIL
# 기대: PASS
```

### AC-DHS-001b — per-command surface가 argument-hint에 선언됨

- **Given** 사용자 지정 per-command surface
- **When** 각 thin command frontmatter를 검사
- **Then** §B.1 surface가 argument-hint에 선언되어 있다

```bash
# argument-hint frontmatter 라인에 anchor (file-wide prose match 방지 — D4)
grep -E '^argument-hint:.*since' .claude/commands/harness/release-update.md \
  && grep -E '^argument-hint:.*(issues|pr)' .claude/commands/harness/github.md \
  && grep -E '^argument-hint:.*(hotfix|VERSION)' .claude/commands/harness/release.md \
  && echo PASS || echo FAIL
# 기대: PASS — surface가 argument-hint: 필드 자체에 선언되어야 함 (prose 어딘가가 아님)
```

### AC-DHS-002a — specialist 새 이름 존재 + 옛 이름 부재

- **Given** specialist 이름 변경 (`harness-devkit-X` → `harness-X`)
- **When** `.claude/agents/harness/` 검사
- **Then** 새 이름 3개 존재, 옛 `harness-devkit-*` 부재

```bash
test -f .claude/agents/harness/harness-release-update-specialist.md \
  && test -f .claude/agents/harness/harness-github-specialist.md \
  && test -f .claude/agents/harness/harness-release-specialist.md \
  && echo NEW_OK || echo NEW_FAIL
find .claude/agents/harness -name 'harness-devkit-*' | head -1
# 기대: NEW_OK + harness-devkit-* find 결과 비어있음
```

### AC-DHS-002b — frontmatter name 갱신 + self-reference 갱신

- **Given** specialist body는 verbatim 보존하되 name/self-reference만 변경
- **When** 새 specialist 파일 검사
- **Then** frontmatter `name`이 `harness-X-specialist`이고 `harness-devkit` 자기-참조가 남아있지 않다

```bash
grep -q '^name: harness-release-update-specialist' .claude/agents/harness/harness-release-update-specialist.md \
  && grep -q '^name: harness-github-specialist' .claude/agents/harness/harness-github-specialist.md \
  && grep -q '^name: harness-release-specialist' .claude/agents/harness/harness-release-specialist.md \
  && echo NAME_OK || echo NAME_FAIL
# 옛 진입 참조 잔존 검사 (Migration Provenance의 역사적 인용은 허용 — '/harness:devkit' 활성 라우팅만 금지)
grep -rn '/harness:devkit' .claude/agents/harness/harness-release-update-specialist.md \
  .claude/agents/harness/harness-github-specialist.md \
  .claude/agents/harness/harness-release-specialist.md \
  | grep -v 'Migration Provenance\|Ported from\|구\|formerly\|이전' | head -5
# 기대: NAME_OK + 활성 '/harness:devkit' 라우팅 참조 비어있음
```

### AC-DHS-003a — 통합 devkit.md + manifest.json 부재

- **Given** 통합-진입 결정 번복
- **When** `.claude/commands/harness/` 검사
- **Then** 통합 `devkit.md` + 통합 `manifest.json` 부재

```bash
test ! -f .claude/commands/harness/devkit.md \
  && test ! -f .claude/commands/harness/manifest.json \
  && echo PASS || echo FAIL
# 기대: PASS
```

### AC-DHS-003b — Runner release-update 전용 rename + 재범위화

- **Given** release-update만 fan-out을 가짐 (Runner asymmetry)
- **When** workflows 디렉터리 검사
- **Then** `harness-release-update-run.js` 존재 + 옛 `harness-devkit-run.js` 부재 + release-update manifest 존재 + github/release Runner 부재

```bash
test -f .claude/workflows/harness-release-update-run.js \
  && test ! -f .claude/workflows/harness-devkit-run.js \
  && test -f .claude/commands/harness/release-update/manifest.json \
  && test ! -f .claude/workflows/harness-github-run.js \
  && test ! -f .claude/workflows/harness-release-run.js \
  && echo PASS || echo FAIL
# 기대: PASS — github/release는 Runner/manifest 없음 (Runner asymmetry, §B.2)
```

### AC-DHS-004 — CI guard 재지향 + stale `devkit` 토큰 0 (RED→GREEN)

- **Given** CI guard가 단일 `harness-devkit` namespace를 단언 + 파일명/함수/sentinel에 stale `devkit` 토큰 보유
- **When** 3개 분리 namespace로 재지향 + 파일/함수/sentinel rename 후 `go test`
- **Then** 갱신된 guard가 PASS하고 `internal/template/*namespace_test.go`에 `devkit` 토큰이 0이다

```bash
# 갱신된 guard PASS (파일명 split_namespace_test.go + 함수 TestSplitHarnessNamespaceNoLeak)
go test ./internal/template/... -run 'Namespace' 2>&1 | tail -5
# [MANDATORY] stale devkit 토큰 0 검사 — 파일명/함수/sentinel(DEVKIT_NAMESPACE_LEAK)까지 잡음
#   (harness-devkit-only grep이 놓치는 단독 'devkit' 파일명/sentinel 토큰을 잡는다)
grep -rn 'devkit' internal/template/*namespace_test.go
# 옛 파일명이 git mv로 사라졌는지 (split_namespace_test.go 존재 + devkit_namespace_test.go 부재)
test -f internal/template/split_namespace_test.go && test ! -f internal/template/devkit_namespace_test.go && echo RENAME_OK || echo RENAME_FAIL
# 기대: go test PASS + grep 결과 EMPTY(stale devkit 토큰 0) + RENAME_OK + 3개 분리 namespace 보호
```

### AC-DHS-005 — dev-only 누출 invariant 보존

- **Given** harness 자산은 사용자 프로젝트에 배포 금지
- **When** embedded template tree 검사
- **Then** `commands/harness/`, `harness-{release-update,github,release}*`, `workflows/harness-*` 모두 부재

```bash
find internal/template/templates -path '*commands/harness*' | head -1
find internal/template/templates \( -name 'harness-release-update*' -o -name 'harness-github*' -o -name 'harness-release*' \) | head -1
find internal/template/templates -path '*workflows/harness-*' | head -1
# 기대: 세 find 모두 비어있음
```

### AC-DHS-006 — 5개 doctrine surface 갱신 (standalone `devkit` 포함)

- **Given** doctrine가 `harness:devkit` / `harness-devkit` 참조 + title/header에 standalone `devkit`
- **When** 5개 surface 검사
- **Then** 활성 통합-진입 참조 + standalone `devkit`(title/header)가 제거되고 3개 분리 이름으로 갱신됨 (migration note의 역사적 인용은 허용)

```bash
# 5개 surface에서 활성 devkit 참조 검사 — broadened to standalone 'devkit' (title/header까지 잡음, D2)
# (migration/historical note는 grep -v로 제외)
for f in .moai/docs/dev-only-commands-isolation.md \
         CLAUDE.local.md \
         .claude/rules/moai/development/skill-authoring.md \
         .claude/skills/moai-foundation-core/modules/INDEX.md \
         .claude/skills/moai/references/reference.md; do
  echo "=== $f ==="
  grep -niE 'devkit' "$f" \
    | grep -v 'migration\|Migration\|구 \|formerly\|이전\|historical\|SPEC-V3R6-DEV-HARNESS-CONSOLIDATION' | head -3
done
# 기대: 각 surface에서 활성(비-역사적) devkit 참조 비어있음 (harness:devkit/harness-devkit + 단독 devkit title/header 포함);
#       3개 분리 이름(release-update/github/release) 언급 존재
```

### AC-DHS-007 — 메모리 갱신은 sync-phase (plan/run에서 미작성)

- **Given** 메모리 갱신은 sync-phase 책임 (REQ-DHS-007)
- **When** plan-phase / run-phase 완료 시점
- **Then** 본 SPEC 관련 메모리 entry가 아직 작성되지 않음 (sync-phase에서만 작성)

```bash
# plan/run-phase 시점: SPLIT-001 메모리 entry 부재가 정상
grep -rl 'DEV-HARNESS-SPLIT-001' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/ 2>/dev/null | head -1
# 기대: plan/run-phase에서는 비어있음 (sync-phase에서 manager-docs가 작성)
```

---

## §D.2 — Quality Gate / Definition of Done

- [ ] 3개 thin command 생성 (AC-DHS-001a/b)
- [ ] 3개 specialist 이름 변경 + body verbatim (AC-DHS-002a/b)
- [ ] 통합 devkit.md + manifest.json 제거 (AC-DHS-003a)
- [ ] Runner release-update 전용 rename + 재범위화, github/release Runner 없음 (AC-DHS-003b)
- [ ] CI guard 재지향 + `go test ./internal/template/...` PASS (AC-DHS-004)
- [ ] dev-only 누출 invariant 보존 (AC-DHS-005)
- [ ] 5개 doctrine surface 갱신 (AC-DHS-006)
- [ ] Layer B harness specialist 무변경 (touch 금지 준수)
- [ ] CONSOLIDATION-001 closed artifacts 무변경 (completed 유지)
- [ ] 메모리 갱신은 sync-phase로 보류 (AC-DHS-007)

## §D.3 — Edge Cases

- **Edge: Migration Provenance의 역사적 `harness-devkit` 인용**: specialist body의 Migration Provenance + doctrine migration note는 역사적 인용으로 `harness-devkit` / `harness:devkit`를 언급할 수 있다 — 이는 anti-pattern이 아니라 이력 보존. AC 검증 명령은 `grep -v`로 역사적 인용을 제외하고 *활성* 참조만 검사한다.
- **Edge: github/release manifest 일관성 판단**: 기본은 no-manifest. run-phase에서 일관성 가치가 정당화되면(plan §B.2) 최소 per-harness manifest를 추가할 수 있으나, 그 경우 AC-DHS-003b의 `github/release Runner 부재` 단언은 유지된다 (manifest ≠ Runner).
- **Edge: `git mv` 이력 보존**: specialist/Runner 이름 변경은 `git mv` 후 Edit로 수행 — 신규 Write+삭제가 아니라 이력 보존 경로.
