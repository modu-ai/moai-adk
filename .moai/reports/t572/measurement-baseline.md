# t572 실측 baseline — OwnershipTransitionRule 무음 결함

- 카드: t572 (배차: lane-11, 2026-09-08)
- 트리: `.claude/worktrees/t572` @ `3ac58b5a1` (= origin/develop tip, fetch 후)
- 배차 인용 SHA: `91d25bc61` — `internal/spec/lint_ownership.go` 는 두 SHA 사이 **무변경**(아래 drift 실측), 좌표 유효. `lint.go` 는 +198/-19 변경(등록 좌표 :139 → **:147** 이동).

## Claim

1. 룰은 등록·실행되지만, "발견된 전환 커밋에 트레일러가 없으면" 조용한 nil 로 통과한다 — 배차문 좌표 `91d25bc61:internal/spec/lint_ownership.go:414` 정확.
2. 배차문 전제 "이 저장소 커밋 전부가 트레일러를 달지 않는다"는 **그대로는 거짓** — 실제 트레일러 보유 커밋 104건(전체 11,917건 중), SPEC 경로 83건. 단, **가장 최근 보유자가 2026-09-03** 이므로 "09-03 이후 전환에는 트레일러가 없다"가 참인 형태이고, 그 기간 착지 카드들의 전환은 전부 무음 통과 = "공허한 초록" 지적은 이 스코프로 유효.
3. (c)안(무음 시 Info `unmeasured` 명시 보고)은 CI `spec lint --strict` 를 깨지 않는다 — strict 승격은 Warning 한정.

## Evidence — 코드 지도 (worktree @ 3ac58b5a1 기준)

- 등록: `internal/spec/lint.go:147` `&OwnershipTransitionRule{}`
- 진입: `internal/spec/lint_ownership.go:368` `Check()`
- **무음 nil 4곳**:
  - `:400-402` `rec == nil` — 50커밋 창(`gitLogWindowSize = 50`, `internal/spec/drift.go:302`) 안에 전환 없음 → 조용한 nil
  - `:405-407` `expected == ownerNone` — 매트릭스 미정의 전환 → 조용한 nil
  - **`:409-416` `rec.AuthoredByAgent == ""` — 전환은 발견됐는데 트레일러 없음 → 조용한 nil (배차 지점)**. 주석: M4 AC-LSG-004 의 false-positive 가드(legacy/non-MoAI 커밋은 검사 대상 아님).
  - `:419-422` 트레일러 있으나 행위자 미인식(예: manager-git) → 조용한 nil
- 캐시: `internal/spec/gitquery_cache.go:258` `cachedOwnershipTransition`, `:99` `cachedGitDirAvailable`
- subject-prefix fallback 은 M4(AC-LSG-004)에서 **퇴역** — `commitOwnerKind`(:112)는 회귀 테스트 전용으로 유지.

## Evidence — 트레일러 실측 (명령 + 출력 전문)

```
$ git log --all --format='%(trailers:key=Authored-By-Agent,valueonly)' | /usr/bin/grep -c '[^[:space:]]'
104
$ git rev-list --all --count
11917
$ git log --all --format='%h%x09%(trailers:key=Authored-By-Agent,valueonly)' -- .moai/specs/ | awk -F'\t' '$2 != "" {n++} END{print n+0}'
83
$ git log --all --format='%h%x09%ad%x09%s%x09%(trailers:key=Authored-By-Agent,valueonly)' --date=short | awk -F'\t' '$4 != "" {print; n++; if(n>=15) exit}'
a20fba05f	2026-09-03	docs(SPEC-LANE-PUSH-DOC-001): M2 run-phase evidence export (t463)	manager-develop
a30edfe98	2026-09-03	fix(SPEC-LANE-PUSH-DOC-001): M1 lane push actor sentence repair (t463)	manager-develop
b1bcce4f4	2026-09-01	fix(SPEC-STATUS-TRANSITION-VALIDITY-001): memoize the status-transition history walk (t376)	manager-develop
f32e9a346	2026-08-27	chore(SPEC-V3R6-GRAPH-FRESHNESS-001): Mx-phase audit-ready signal + 3-phase close	orchestrator-direct
... (이하 생략 — 값 enum: manager-develop / orchestrator-direct 등 §D.1.6 열거와 일치)
```

**측정기 결함 기록 (재측정 이력)**: 최초 시도 2건(`--format` 에 `%h` 를 트레일러 값 뒤에 붙인 형태)은 트레일러 부재 커밋도 줄이 비지 않아 전체 커밋수를 셌다(4481 = SPEC 경로 전체 커밋수). 값-먼저+awk 빈칸 필터 형태로 재채워 위 수치 확정. `--grep` 782건 히트는 본문 "언급"이지 트레일러가 아님.

## Evidence — strict/소비자 (Info 폭발력)

```
$ git grep -n "Strict" origin/develop -- internal/spec internal/cli | grep -v _test | grep lint.go
internal/spec/lint.go:62:		if r.Strict && f.Severity == SeverityWarning && !f.Advisory {
```
→ strict 승격은 **Warning 한정**. Info 는 게이트 안 건드림. CI: `.github/workflows/spec-lint.yml:58` `go run ./cmd/moai spec lint --strict` — push(main/develop) + PR(paths: `.moai/specs/**`, `internal/spec/**`), `fetch-depth: 0`(t371 REQ-SLGB-008 — 전체 이력 보장).

불게이트 관례: `warning + Advisory: true` 에미션 사이트 명시(`lint.go:1056`, MovingRefUnpinned 계열). 본 파일의 기존 Info 형제(`OwnershipTransitionSkipped:375-385`, `OwnershipTransitionUnreachable:388-399`)는 무-Advisory plain Info — **파일 내 일관성 기준 plain Info 권장**.

## Evidence — 초록을 인용하는 자리 (인용 지도)

- `.claude/rules/moai/development/spec-frontmatter-schema.md:198`(로컬) + `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md:198`(미러) — 발동 조건을 "commit **subject prefix** 불일치"로 기술 = **M4 이전 구현을 서술하는 문서 드리프트**(현 구현은 트레일러 SSOT + 트레일러 게이트).
- `.claude/agents/moai/manager-develop.md:185` + 템플릿 미러 + `.codex/agents/moai/manager-develop.toml:173` — "진행 넘기면 `OwnershipTransitionInvalid` 에 걸린다" 억지 인용. 트레일러 없는 커밋에서는 실제로 불이 안 붙음 = 배차문 "공백 위에 선 인용"의 실례. (단, (c)안 채택 후에도 이 문장 자체는 거짓이 되지 않음 — 걸릴 조건이 트레일러 존재를 전제할 뿐.)
- `.claude/rules/local/lifecycle-sync-gate.md:208,212` — `lint.skip: [OwnershipTransitionInvalid]` 옵트아웃 안내(로컬 전용, AC-LSG-012).

## 세 갈래 판정 재료

- **(a) 트레일러 부착 부활**: 관례는 존재(workflow-specialist.md:83, SPEC §D.1.6 HARD)하나 09-03 부터 실천이 끊김. 유일하게 "새 전환을 실제로 측정 가능"하게 하는 축. 그러나 범위가 전 행위자(에이전트 정의·커밋 게이트·템플릿 미러 전파)라 본 카드 스코프 초과 — **후속 카드 권고**로 기록.
- **(b) 트레일러 없이 판정(fallback 부활)**: M4 AC-LSG-004 기록 결정(subject prefix 는 WHO 로 불신뢰 — 퇴역 사유 명시)과 정면 충돌. t260 감사가 당시 fallback 이 유일 살아있는 경로였음을 실측한 이력 있음. **기각** — 기각 사유를 SPEC 에 기록.
- **(c) 무음 시 `unmeasured` 명시 보고(Info)**: 판정 기준「공허한 초록보다 시끄러운 미측정」에 직결. M4 false-positive 가드 훼손 아님(Info 는 Invalid 위장이 아니라 측정 상태 진술). CI strict 안전(위 실측). 예상 소음: 09-03 이후 전환을 가진 SPEC 수준(유한, advisory). **1차 후보** — manager-spec 이 판정문으로 성문화.

## Evidence — 변경 전 baseline lint 실측 (worktree @ 3ac58b5a1, 2026-09-08)

```
$ go run ./cmd/moai spec lint --strict   # rc=1
0 error(s), 4698 warning(s)
exit status 1
$ # 전체 4,703줄 중 OwnershipTransition 히트는 단 1건:
WARNING OwnershipTransitionInvalid  .moai/specs/SPEC-LSPMCP-001/spec.md  1
  SPEC SPEC-LSPMCP-001 transition "archived" → "superseded" expected owner "manager-spec"
  but commit 94d9fccfc77939af86c7d4ee3222d92dada20afa (Authored-By-Agent: manager-develop)
  maps to "manager-develop"
$ # 심각도×코드 상위 구성:
3599 CoverageIncomplete / 492 ModalityUnjudged / 180 ModalityMalformed /
115 MovingRefUnpinned / 102 StatusTransitionInvalid / 81 REQTableRowsRejected /
 48 LegacyEARSKeyword / 27 MissingExclusions / 18 StatusGitConsistency /
 14 FrontmatterInvalid /  7 StatusTokenUnrecognized /  6 SyncSHASlotFormat
```

**해석**:
1. **"트레일러 104건 → 코퍼스 전체 판정 1건"** — 유일하게 창 안 전환 커밋에 트레일러가 살아있는 SPEC(SPEC-LSPMCP-001)만 실제 판정을 받았고, 나머지 전부 무음 통과. "공허한 초록"의 정량 증거이자, 트레일러만 있으면 룰이 실제로 작동한다는 반증(메커니즘 생존 확인). 본 카드는 이 기존 발견 1건의 동작을 절대 깨지 않는다(AC-2).
2. **develop tip 기준 `spec lint --strict` 는 이미 rc=1** — strict 가 비-advisory Warning 을 승격하는 구조(lint.go:62)상 CoverageIncomplete 3,599건 등이 원인. 이것은 본 카드 이전의 상속 상태이며, 본 카드의 Info finding 은 이 exit 에 어떤 영향도 주지 않는다(Info 는 승격 대상 아님 — AC-4 는 "exit 를 뒤집지 않음"으로 검증). CI 쪽 spec-lint 잡의 실제 판정은 별도 관측 필요(상속 적색 귀속).

## Gaps

- CI spec-lint 잡의 develop 최근 판정은 lane 에서 `gh run list` 로 관측 시도; 결과는 완료 보고에 귀속.
- unmeasured 발화 볼륨의 per-SPEC 정확한 수는 구현 후 lint 실측에서 확정 예정(815 디렉터 중 창 안 전환 보유분만).

## Evidence — 커밋 트레일러 파싱 위험 (REQ-OWN-010 이행 시 필수 지식, 2026-09-08 실측)

plan 커밋에서 실측: git 트레일러 파서는 마지막 문단에 트레일러가 아닌 줄이 섞이면 **블록 전체를
파기**한다(일부 보존이 아니라 전부). 고립 실험(`git interpret-trailers --parse`):

```
$ printf 'subject\n\nbody\n\nAuthored-By-Agent: manager-spec\n' | git interpret-trailers --parse
Authored-By-Agent: manager-spec        # ✓ 트레일러만 있는 마지막 문단
$ printf 'subject\n\nbody\n\n🗿 MoAI\nAuthored-By-Agent: manager-spec\n' | git interpret-trailers --parse
(빈 출력)                               # ✗ 마커 줄이 같은 문단에 있으면 블록 통째로 죽는다
```

**규칙**: `Authored-By-Agent:` 트레일러는 **마지막 문단의 유일한 줄**이어야 하고, `🗿 MoAI`
마커는 그 앞 문단에 둔다(빈 줄 분리). 커밋마다
`git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'` 가 행위자를 출력하는지
확인한다 — 빈 출력은 "관례 부재"와 구분 불가능하게 읽힌다(본 카드가 수리하는 무음과 같은 형태).

## Residual-risk

- develop 의 `lint.go` 는 t518(axes) 축에서 활발히 변동 중 — 등록 줄번호·인접 룰은 병합 시점 재확인 필요.
- CI 에 상속 적색 축 존재(t577/lane-9 errcheck 수리 대기, t518/t536 귀속 기록) — 본 카드 판정은 이 축과 무관하게 스코프해야 함.
