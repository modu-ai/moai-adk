# t572 run-preflight — SPEC-OWNERSHIP-SILENCE-001 (Section C 재측정)

- run-phase 진입 시점: 2026-09-08, lane-11 manager-develop
- 측정 트리: `.claude/worktrees/t572` @ `b642479ec` (plan-phase 커밋, branch `WT-ownership-lint-silent`)
- 진입 검증: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t572` ✓ /
  `git branch --show-current` → `WT-ownership-lint-silent` ✓ /
  `git log --oneline -1` → `b642479ec feat(SPEC-OWNERSHIP-SILENCE-001): plan-phase artifacts (M, 3 artifacts) (t572)` ✓

## C.1 좌표 유효성

```
$ /usr/bin/grep -n "OwnershipTransitionRule" internal/spec/lint.go
147:		&OwnershipTransitionRule{},
149:		// (card t376). Sits beside OwnershipTransitionRule and answers a

$ /usr/bin/grep -n "AuthoredByAgent ==" internal/spec/lint_ownership.go
414:	if rec.AuthoredByAgent == "" {
```

→ 등록 `:147`, 피의자 분기 `:414` — baseline(b642479ec 상위 3ac58b5a1)과 **이동 없음**. 재귀속 불필요.

strict 승급 조건(참고용 판독, lint.go:57-64):

```
$ /usr/bin/sed -n '55,70p' internal/spec/lint.go
		if f.Severity == SeverityError {
			return true
		}
		if r.Strict && f.Severity == SeverityWarning && !f.Advisory {
			return true
		}
```

## C.2 형제 지점 무변경 판독 (baseline 형태와 동일)

- `:400-402` `rec == nil → return nil` ✓ / `:405-407` `expected == ownerNone → return nil` ✓ /
  `:419-422` 미인식 행위자 → `return nil` ✓ (판독: lint_ownership.go 전문, 이번 실행)
- 형제 발견: `:375-385` Skipped(Info), `:388-399` Unreachable(Info), `:424-443` Invalid(Warning) ✓

## C.3 쌍둥이 일치 (편집 전)

```
$ cmp -s internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md .claude/rules/moai/development/spec-frontmatter-schema.md; echo "cmp_rc=$?"
cmp_rc=0
$ wc -l (양 사본)
     256 internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
     256 .claude/rules/moai/development/spec-frontmatter-schema.md
```

## C.4 상속 적색 — baseline 패키지 상태 (변경 전)

```
$ go test ./internal/spec/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	78.980s
```

→ **GREEN baseline** — 기존 적색 0건. 본 카드 GREEN 판정은 이 baseline 대비로 스코프한다.

## 무음-단언 형제 인벤토리 (AC-OWN-003 사전 확정)

```
$ /usr/bin/grep -n 'AuthoredByAgent' internal/spec/lint_ownership_test.go
556:			AuthoredByAgent: "", // legacy / non-MoAI commit — no trailer   ← 유일한 빈 리터럴
518/585/623: 비어있지 않은 값 (fixture)
86/188: 테이블 필드 치환 — 테이블 행 12건 전부 비어있지 않은 트레일러(manager-spec/develop/docs/orchestrator-direct)
```

→ REQ-OWN-006 스코프 = `trailer_absent_silent_skip` 1개 테스트로 확정 (plan-audit 재확인 일치).

## Residual-risk

- develop의 `lint.go`는 t518 축에서 변동 중 — 병합 시점 재측정은 리드 통합 창의 소관(§H.1).
- CI spec-lint 잡 상속 적색 + t577 errcheck 축 — 본 카드 판정 축 밖(dispatch B5 기록 귀속).
