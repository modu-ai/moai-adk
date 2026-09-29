---
id: SPEC-JEV-SKILL-SUGGESTION-001
title: "acceptance — Jev skill-suggestion guidance skill"
version: "0.1.0"
created: 2026-09-30
author: manager-spec
---

# acceptance.md — SPEC-JEV-SKILL-SUGGESTION-001

## D. AC Matrix

| AC | Scenario | Severity | REQ |
|----|----------|----------|-----|
| AC-JSK-001 | 두 사본 존재 + 바이트 동일 | Blocker | REQ-JSK-001 |
| AC-JSK-002 | 신설 가드 테스트 green (양성 대조 포함) | Blocker | REQ-JSK-002 |
| AC-JSK-003 | 중립성 + 철수 명명 부재 grep 0 | Blocker | REQ-JSK-003 |
| AC-JSK-004 | display-only 계약 산문 존재 | Blocker | REQ-JSK-004 |
| AC-JSK-005 | 기각 기제 기록 존재 | Blocker | REQ-JSK-005 |
| AC-JSK-006 | 운영 체크리스트 6항목 존재 | Blocker | REQ-JSK-006 |
| AC-JSK-007 | 게이트-오프 fail-open 산문 존재 | Blocker | REQ-JSK-007 |
| AC-JSK-008 | catalog 엔트리 + 해시 재생성 | Blocker | REQ-JSK-008 |
| AC-JSK-009 | 빌드 매트릭스 green | Blocker | REQ-JSK-008 |
| AC-JSK-010 | 비테스트 Go diff 0 + 기존 가드 green | Blocker | REQ-JSK-009 |
| AC-JSK-011 | lint 신규 이슈 0 | High | REQ-JSK-009 |

## D.0 RED-now 기준선 (plan 단계 관측, 이 트리)

관측 시점: 2026-09-30, worktree `.moai/worktrees/t1340` (develop @ `7a713a9a8` 기반).
4개 산출물 부재 — 등록·존재 AC 의 RED-now 셀:

- `ls .claude/skills/ | grep -c moai-jev-skill-suggestion` → `0`
- `ls internal/template/templates/.claude/skills/ | grep -c moai-jev-skill-suggestion` → `0`
- `grep -c moai-jev-skill-suggestion internal/template/catalog.yaml` → `0`
- `ls internal/cli/ | grep -c jev_skill_suggestion` → `0`
- `grep -rln "REQ-JSK" .moai/specs/ internal/` → 공백 (REQ-JSK 접두사 유일성)

## D.1 AC-JSK-001 — 두 사본 존재 + 바이트 동일

**Given** M2 가 적용된 트리, **When** `cmp .claude/skills/moai-jev-skill-suggestion/SKILL.md internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md` 가 실행되면, **Then** exit 0 (차등 0)이고 두 파일 모두 비어 있지 않다(`wc -c` 각각 > 0). flip 은 옳은 이유로: RED-now 셀은 "파일 없음(카운트 0)"이고, 추가된 것이 곧 파일 2개다.

## D.2 AC-JSK-002 — 신설 가드 테스트 green

**Given** M2 트리, **When** `go test ./internal/cli/ -run '^TestJevSkillSuggestionSkill(CarriesNoCallPath|CopiesStayIdentical)$'` 가 실행되면, **Then** 정확히 2 테스트가 sweep 되어 green 이고, 양성 대조가 검사 발화 능력을 증명한다(대조 파일에서 토큰 발견 — 테스트 내부 단언). RED→GREEN 쌍의 RED 증거는 plan §E8 (M1 시점 read-fail verbatim).

## D.3 AC-JSK-003 — 중립성 + 철수 명명 부재

**Given** M2 트리, **When** 다음을 양 사본에 대해 실행하면, **Then** 각각 `0`:

```bash
grep -Ec 'SPEC-[A-Z]{2,}-[0-9]{3}|REQ-[A-Z]{2,}-[0-9]{3}|t1340|[0-9a-f]{40}' <copy>
grep -c 'jev-suggest' <copy>
```

`<copy>` 는 로컬·템플릿 양쪽. 내부 흔적 클래스와 철수된 명령명이 어느 사본에도 없다.

## D.4 AC-JSK-004 — display-only 계약 산문

**Given** M2 트리, **When** 양 사본에 `grep -c 'keeps its selection authority'` 와 `grep -c 'a signal a person'` 를 실행하면, **Then** 각각 ≥ 1. 본문은 신호/선택 분리를 서술하고 기존 선택자의 권한 유지를 명시한다(spec.md §B REQ-JSK-004 문언과 정합).

## D.5 AC-JSK-005 — 기각 기제 기록

**Given** M2 트리, **When** 양 사본에 `grep -c 'Hiding the listing'`, `grep -ci 'function-hook'`, `grep -c 'constant-answer baseline'` 를 실행하면, **Then** 각각 ≥ 1. 목록 은닉+본문 주입의 기각 사유와 함수 훅 미채택(조기 실험 전제 포함)이 기록돼 있다. 정확한 전제 문자열(CC 버전·환경변수)은 SPEC 쪽에만 있고 본문에는 중립 산문으로 있다 — 이것이 의도된 배치다.

## D.6 AC-JSK-006 — 운영 체크리스트 6항목

**Given** M2 트리, **When** 양 사본에 대해 `grep -c` 를 6 sentinel 로 실행하면, **Then** 각각 ≥ 1: `'announce'`, `'non-task origins'`, `'invocation-disabled'`, `'Cache the roster'`, `'duplicate injection'`, `'Canonicalize names'`.

## D.7 AC-JSK-007 — 게이트-오프 fail-open

**Given** M2 트리, **When** 양 사본에 `grep -c 'workflow.jev.enabled'` 와 `grep -c 'NO SIGNAL'` 를 실행하면, **Then** 각각 ≥ 1. 첫 절이 게이트-오프 동작(요청 미구성·무신호 진행·부재≠부정)을 서술한다.

## D.8 AC-JSK-008 — catalog 엔트리 + 해시 재생성

**Given** M3 트리, **When** (a) `grep -A 4 'moai-jev-skill-suggestion' internal/template/catalog.yaml` 실행, **Then** name/tier(`core`)/path/hash/version 5필드가 보이고 hash 는 64자 hex; (b) `make build` 직후 재실행 시 hash 가 변하지 않는다(결정적 재생성). 손 계산 hash 삽입은 §G 반패턴.

## D.9 AC-JSK-009 — 빌드 매트릭스

**Given** M3 트리, **When** `go build ./...` 와 `GOOS=windows GOARCH=amd64 go build ./...` 를 실행하면, **Then** 양쪽 exit 0. `make build` 가 카탈로그 해시 재생성을 포함해 성공한다.

## D.10 AC-JSK-010 — 비테스트 Go 무접촉 + 기존 가드 green

**Given** 본 SPEC 의 전체 diff, **When** (a) `git diff --name-only <base>..HEAD -- 'internal/**/*.go' | grep -v _test` 실행하면, **Then** 공백(비테스트 Go 변경 0 — 신설은 `_test.go` 하나뿐); (b) `go test ./internal/jevmeasure/ -run '^TestNoConsumerCallPathShips$'` green; (c) `git diff <base>..HEAD -- internal/jevmeasure/` 공백 (기존 가드 zero-edit).

## D.11 AC-JSK-011 — lint

**Given** M3 트리, **When** `golangci-lint run --timeout=2m ./internal/cli/...` 실행하면, **Then** 이 SPEC diff 에 귀속되는 신규 이슈 0 (§C baseline 관측치와 대비).

## D.12 엣지 케이스

- 스킬 파일이 있고 catalog 엔트리가 없으면(또는 그 반대) → AC-JSK-008 FAIL — 한쪽만 착지한 수정은 반대쪽을 확립하지 않는다.
- 두 사본이 한쪽만 편집돼 갈라지면 → AC-JSK-001/002 FAIL — 사용자 프로젝트가 읽는 내용이 변한다.
- 본문에 호출 경로를 설명하는 "예시"가 들어오면 → AC-JSK-002 가 잡는다(금지 토큰), 토큰 우회 표현은 plan-auditor/sync-auditor 산문 검토가 잡는다.

## D.13 품질 게이트

- TRUST 5: Tested(신설 가드 2테스트 + 기존 가드 회귀), Readable(영어 본문, 형제 스킬 문체 정합), Unified(스킬 프런트매터 스키마 준수), Secured(키 재질 무인라인, 호출 경로 없음), Trackable(Conventional Commits + 카드 id).
- Definition of Done: AC-JSK-001..011 전부 PASS + §E 매트릭스에 (a)명령/(b)출력/(c)SHA 귀속 완비 + `git status --short` 청결.
