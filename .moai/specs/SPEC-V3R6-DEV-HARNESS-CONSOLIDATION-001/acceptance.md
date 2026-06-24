# SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 — Acceptance Criteria

> Given-When-Then acceptance criteria (each with a binary verification command). 각 AC는 독립적으로 검증 가능(grep/find/file-existence/test 명령). 7개 scope 항목 + Runner/human-gate 정합을 커버한다.

## §D. AC Matrix

| AC ID | REQ | Capability/Topic | 검증 방식 |
|-------|-----|------------------|-----------|
| AC-DHC-001 | REQ-DHC-001 | 단일 devkit 하네스 생성 | file-existence + manifest |
| AC-DHC-002 | REQ-DHC-002 | specialist 콘텐츠 포팅 | file-existence + 구조 grep |
| AC-DHC-003 | REQ-DHC-003 | 번호 커맨드 삭제 + 구 소스 처분 | file-absence (find) |
| AC-DHC-004 | REQ-DHC-004 | dev-only 격리 보존 | find (템플릿 누출 0) |
| AC-DHC-005 | REQ-DHC-005 | CI guard 확장 (embedded-tree-absence) | go test + grep |
| AC-DHC-006 | REQ-DHC-006 | doctrine 업데이트 (3 surface, §2 포함) | grep |
| AC-DHC-007a | REQ-DHC-007 | Runner 비-상호작용 only (negative 증명) | grep (AskUserQuestion + gh/approval 부재) |
| AC-DHC-007b | REQ-DHC-007 | 사람-게이트 specialist 위임 (primitive 검증) | manifest primitive 검증 |

## §D.1 Acceptance Criteria (Given-When-Then)

### AC-DHC-001 — 단일 devkit 하네스 생성

**Given** v4 Builder가 run-phase에서 실행되었을 때,
**When** GENERATE phase가 완료되면,
**Then** 다음이 모두 존재한다:
- `.claude/commands/harness/devkit.md` (thin-wrapper 진입 커맨드, /harness:devkit 해석)
- `.claude/commands/harness/manifest.json` 또는 `.claude/harness/devkit/manifest.json` (SSOT)
- `.claude/workflows/harness-devkit-run.js` (Runner)

manifest는 정확히 3 capability에 대응하는 3개 specialist를 `specialists` 배열에 가진다 (release-update / github / release). **[D4] 그리고 thin-wrapper 커맨드 `devkit.md`는 선택된 manifest 위치(고정 경로)를 본문에 기록하여 Runner가 manifest를 찾을 수 있게 한다** (위치 모호성 폐쇄 — "Runner can locate manifest" 루프).

검증:
```bash
test -f .claude/commands/harness/devkit.md && echo "entry OK"
MANIFEST=$(ls .claude/commands/harness/manifest.json .claude/harness/devkit/manifest.json 2>/dev/null | head -1)
echo "manifest at: $MANIFEST"
test -f .claude/workflows/harness-devkit-run.js && echo "runner OK"
# [D7] manifest specialists 3개 확인 (JSON-aware jq)
jq '.specialists | length' "$MANIFEST"
# 기대: 3
# [D4] thin-wrapper가 선택된 manifest 경로를 기록하는지 확인
grep -qF "$(basename "$(dirname "$MANIFEST")")/manifest.json" .claude/commands/harness/devkit.md \
  || grep -q "manifest.json" .claude/commands/harness/devkit.md && echo "manifest path recorded in entry cmd"
```

### AC-DHC-002 — specialist 콘텐츠 소스 재사용 (구조적 충실도)

**Given** 마이그레이션된 `agents/local/*-specialist.md` body + `release.md` body가 존재할 때,
**When** GENERATE가 devkit specialist를 emit하면,
**Then** 3개 specialist 파일이 `.claude/agents/harness/harness-devkit-{release-update,github,release}-specialist.md`에 존재하고, 각 파일은 원본의 핵심 phase 구조(release-update의 multi-phase tracker, github의 issues/pr sub-command, release의 Enhanced GitHub Flow)를 구조적 충실도로 보존한다.

검증:
```bash
for cap in release-update github release; do
  test -f ".claude/agents/harness/harness-devkit-${cap}-specialist.md" && echo "${cap} specialist OK"
done
# release-update: multi-phase tracker 구조 보존 grep
grep -q "since\|CC\|upstream\|release notes" .claude/agents/harness/harness-devkit-release-update-specialist.md && echo "ru fidelity OK"
# github: issues/pr sub-command 구조 보존
grep -qi "issue\|pr\|gh " .claude/agents/harness/harness-devkit-github-specialist.md && echo "gh fidelity OK"
```

### AC-DHC-003 — 번호 커맨드 삭제 + 구 소스 처분

**Given** devkit 하네스가 live 상태일 때,
**When** run-phase가 완료되면,
**Then** 세 번호 커맨드 파일이 삭제되고, 구 specialist 소스(`agents/local/*-specialist.md` + `release.md`)도 삭제된다(M5 권장: dual-source drift 방지).

검증:
```bash
# 번호 커맨드 삭제 확인 (find 결과 비어있음)
find .claude/commands -name "97-release-update.md" -o -name "98-github.md" -o -name "99-release.md" | head -1
# 기대: (출력 없음)
# 구 소스 삭제 확인
test ! -f .claude/agents/local/release-update-specialist.md && echo "local ru removed"
test ! -f .claude/agents/local/github-specialist.md && echo "local gh removed"
test ! -f .claude/skills/moai/workflows/release.md && echo "release wf removed"
```

> 참고: 만약 sync-phase에서 구 소스 보존이 정당화되면(예: 외부 참조 잔존), 그 결정은 progress.md에 명시 기록한다. 기본은 삭제.

### AC-DHC-004 — dev-only 격리 보존 (템플릿 누출 0)

**Given** devkit 하네스 artifacts가 user-owned namespace에 생성되었을 때,
**When** 템플릿 트리를 검사하면,
**Then** 어떤 devkit artifact도 `internal/template/templates/`에 등장하지 않는다.

검증:
```bash
find internal/template/templates -path "*harness-devkit*" | head -1
# 기대: (출력 없음)
find internal/template/templates -path "*commands/harness*" | head -1
# 기대: (출력 없음)
find internal/template/templates -path "*agents/harness*" | head -1
# 기대: (출력 없음)
```

### AC-DHC-005 — CI guard 확장 (devkit namespace embedded-tree 부재 단언)

**Given** [D1 re-anchored] devkit artifact의 실제 CI 보호는 embedded-tree 부재이며 `dev_only_skill_test.go`는 `.claude/skills/`만 walk하는 skills-only walker라 commands/agents/workflows artifact를 보호하지 못할 때,
**When** 신규 embedded-tree-absence 테스트(`internal/template/embedded_namespace_test.go` 모델)가 실행되면,
**Then** 그 테스트는 `EmbeddedTemplates()`(또는 `internal/template/templates/`)에 (a) `.claude/commands/harness/` 경로 부재, (b) `harness-devkit*` agent 파일 부재, (c) `.claude/workflows/harness-devkit-*` 파일 부재를 단언하며 PASS한다. RED: walker가 검출 가능한 누출 형태(embedded 트리 하위에 심은 `harness-devkit` 경로)를 주입 → FAIL; 제거 → PASS(GREEN).

검증:
```bash
# 신규 embedded-tree-absence 테스트 실행 (embedded_namespace_test.go 패턴; dev_only_skill_test.go 아님)
go test ./internal/template/... -run 'TestTemplateAgentsStructure|HarnessDevkit|EmbeddedTree' -v
# 기대: PASS
# 신규 단언이 commands/harness + harness-devkit + workflows 부재를 커버하는지 확인
grep -Eq "commands/harness|harness-devkit|workflows/harness-devkit" internal/template/*_test.go && echo "embedded-tree-absence guard present"
# 기존 {moai}-only allowlist가 agents/harness/ 부재를 이미 보호함을 재확인
grep -q "HARNESS_NAMESPACE_LEAK" internal/template/embedded_namespace_test.go && echo "agents/harness allowlist intact"
```

> 참고: `dev_only_skill_test.go`(skills-only walker)는 본 AC의 대상이 아니다 — devkit artifact는 commands/agents/workflows이므로 그 테스트로는 RED를 만들 수 없다(skills dir에 `harness-devkit` skill을 심어도 commands/workflows 누출은 미검출).

### AC-DHC-006 — doctrine 업데이트

**Given** sync-phase가 실행될 때,
**When** doctrine 업데이트가 완료되면,
**Then** (a) `.moai/docs/dev-only-commands-isolation.md`는 harness-namespaced dev-only 도구를 반영한 배포 금지 표/체크리스트 + 이 SPEC credit 마이그레이션 노트 + 번호-커맨드 제거를 명명하고, **[D6] (b) CLAUDE.local.md §2 "Local-Only Files"는 삭제된 97/98/99 + `release-update.md` 항목이 제거되고 신규 `commands/harness/devkit*` + `workflows/harness-devkit-run.js`가 추가된다**.

검증:
```bash
grep -q "harness-devkit\|devkit" .moai/docs/dev-only-commands-isolation.md && echo "doctrine names harness"
grep -q "SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001" .moai/docs/dev-only-commands-isolation.md && echo "migration note OK"
grep -qi "97-release-update\|98-github\|99-release.*removed\|consolidat" .moai/docs/dev-only-commands-isolation.md && echo "removal named"
# [D6] CLAUDE.local.md §2 stale 항목 제거 + 신규 항목 추가
grep -q "harness-devkit-run.js\|commands/harness/devkit" CLAUDE.local.md && echo "section2 updated"
# §2의 구 stale 항목이 제거되었는지 (97-release-update.md가 §2 Local-Only 목록에 더 이상 dev-only 진입점으로 남지 않음)
grep -c "97-release-update.md" CLAUDE.local.md  # 0 또는 historical 노트 맥락만
```

> 참고: `git-workflow-doctrine.md`의 `release.md` 언급은 GitHub-Actions release.yml/Release Drafter를 가리키는 false positive — 건드리지 않는다(D6).

### AC-DHC-007a — Runner는 비-상호작용 fan-out만 (negative 증명: AskUserQuestion + 상호작용 surface 부재)

**Given** GENERATE가 Runner를 emit했을 때,
**When** Runner 스크립트를 검사하면,
**Then** `harness-devkit-run.js`는 (i) AskUserQuestion을 호출하지 않고(dynamic-workflow는 mid-run 사용자 프롬프트 불가), **[D8] (ii) 어떤 상호작용 surface도 inline하지 않는다** — `gh pr`/`gh issue` 생성, 승인(approval) 요청 등을 직접 포함하지 않는다. Runner는 release-update 리서치 sweep 같은 비-상호작용 fan-out 부분만 모델링한다.

검증:
```bash
# (i) AskUserQuestion 부재
grep -n 'AskUserQuestion\|mcp__askuser' .claude/workflows/harness-devkit-run.js | grep -v '^\s*//' | head -1
# 기대: (출력 없음 — 주석 제외 매치 0)
# [D8] (ii) Runner가 상호작용 surface(gh pr/issue, 승인)를 inline하지 않음 (necessary AND sufficient의 sufficient 보강)
grep -nE 'gh (pr|issue)|승인|approval' .claude/workflows/harness-devkit-run.js | grep -v '^\s*//' | head -1
# 기대: (출력 없음 — 상호작용 surface는 specialist에 위임, Runner에 inline 금지)
```

### AC-DHC-007b — 사람-게이트 작업은 specialist 위임 (manifest primitive 검증)

**Given** 3개 capability 전부가 사람-게이트/상호작용일 때,
**When** 하네스 설계를 검사하면,
**Then** 상호작용/사람-게이트 작업(사용자 승인, PR 생성, gh CLI, production release 게이트)은 specialist 서브에이전트로 위임되고, orchestrator는 workflow launch 전에 AskUserQuestion 게이트를 보유한다. **[D8] 그리고 각 사람-게이트 capability의 manifest `primitive`는 `sub-agent` 또는 `adversarial-fan-out`이며, `dynamic-workflow`가 아니다** (dynamic-workflow primitive는 비-상호작용 fan-out 전용이므로 사람-게이트 capability에 매핑되면 안 됨).

검증:
```bash
MANIFEST=$(ls .claude/commands/harness/manifest.json .claude/harness/devkit/manifest.json 2>/dev/null | head -1)
# specialist body가 사람-게이트 작업(승인/PR/gh)을 담당하는지 확인
grep -qi "승인\|approval\|PR\|gh pr\|gh issue" .claude/agents/harness/harness-devkit-github-specialist.md && echo "gh specialist holds gate"
# [D8] 사람-게이트 capability(release-update/github/release)의 primitive가 dynamic-workflow가 아님을 검증
jq -r '.specialists[] | select(.role | test("release-update|github|release")) | "\(.role): \(.primitive)"' "$MANIFEST"
# 기대: 각 줄의 primitive가 sub-agent 또는 adversarial-fan-out (dynamic-workflow 금지)
jq -e '[.specialists[] | select(.role | test("release-update|github|release")) | .primitive] | all(. == "sub-agent" or . == "adversarial-fan-out")' "$MANIFEST"
# 기대: true (exit 0)
```

## §D.2 Edge Cases

- **이름 충돌**: `harness-devkit-*`가 기존 Layer B specialist(`cli-template`/`hook-ci`/`quality`/`workflow`)와 충돌하지 않음 — prefix `harness-devkit-`로 분리. PLAN phase pre-flight에서 검증.
- **manifest 위치 모호성**: manifest는 `.claude/commands/harness/manifest.json` 또는 `.claude/harness/devkit/manifest.json` 중 하나의 고정 위치 — entry command가 Runner에게 위치를 기록. AC-DHC-001 검증이 둘 다 허용.
- **구 소스 보존 예외**: 외부 참조가 구 `agents/local/*` body를 가리키면 삭제 대신 보존 + progress.md 기록. 기본은 삭제(M5).

## §D.3 Definition of Done

- [ ] AC-DHC-001 ~ AC-DHC-007b 전부 PASS
- [ ] `.claude/commands/harness/devkit.md` + manifest + Runner 존재
- [ ] 3개 specialist가 `harness-devkit-*`로 포팅(구조적 충실도)
- [ ] 97/98/99 + 구 소스 삭제 (또는 보존 결정 기록)
- [ ] 템플릿 누출 0 (find 결과 비어있음)
- [ ] [D1] embedded-tree-absence 단언 추가(`embedded_namespace_test.go` 모델) + GREEN — `commands/harness/` + `harness-devkit*` + `workflows/harness-devkit-*` 부재 단언
- [ ] doctrine 업데이트 3 surface (dev-only-commands-isolation.md + §21 stub + [D6] CLAUDE.local.md §2)
- [ ] Runner에 AskUserQuestion 부재 AND 상호작용 surface(gh pr/issue, 승인) 부재 [D8], 사람-게이트는 specialist 위임 + manifest primitive ∈ {sub-agent, adversarial-fan-out}
- [ ] `go test ./internal/template/...` (embedded-tree-absence 신규 테스트) PASS

## §D.4 Quality Gate

- Enforce Simplicity: 정확히 3 capability, 추가 추상화 없음.
- Scope discipline: Layer B specialist + 사용자 템플릿 불가침 (find로 확인).
- dev-only 격리: REQ-DHC-004 + AC-DHC-004 (템플릿 누출 0).
