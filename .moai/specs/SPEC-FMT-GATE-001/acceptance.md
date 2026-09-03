# SPEC-FMT-GATE-001 — Acceptance Criteria (Tier M)

> 검증 계층. 각 AC는 `Given-When-Then` 이진 판정. 요구 계층(GEARS REQ)은 spec.md §C.
> 모든 측정은 명령 + 실측 출력을 수반한다(verification-claim-integrity §2).

## §D AC Matrix

### AC-FG-001 — 포맷 위반 트리에서 게이트 명령 실패

```text
Given a tree where `gofmt -l .` at repo root outputs ≥1 path
  (repro: scratch branch, one tracked .go file reformatted off-spec)
When `test -z "$(gofmt -l .)"` runs at repo root
Then the exit code is 1
```

### AC-FG-002 — 녹색 트리에서 게이트 명령 통과

```text
Given the activation commit's tree checked out clean
When `test -z "$(gofmt -l .)"` runs at repo root
Then the exit code is 0
```

### AC-FG-003 — 착지 순서: t457 선행 (기계 판정)

```text
Given the activation commit SHA <act>
When `git merge-base --is-ancestor e1fdf00d1 <act>` runs
Then the exit code is 0 (t457 tip e1fdf00d1 is an ancestor of <act>)
```

### AC-FG-004 — 활성 시점 녹색 (기계 판정)

```text
Given the activation commit SHA <act> checked out
When `gofmt -l . | wc -l` runs at repo root
Then the output is exactly `0`
```

### AC-FG-005 — 배포 표면 불변

```text
Given the full commit range delivered by this SPEC (base d592b0551..branch tip)
When `git diff --name-only d592b0551..HEAD -- internal/template/templates/ | wc -l` runs
Then the output is exactly `0`
```

### AC-FG-006 — 로컬 패리티 (`make fmt-check`)

```text
Given a clean-format tree
When `make fmt-check` runs
Then exit code is 0
Given a dirty-format tree (≥1 unformatted tracked .go file; predicate
  `git ls-files -z '*.go' | xargs -0 gofmt -l` outputs ≥1 path)
When `make fmt-check` runs
Then exit code is non-zero AND the unformatted file list is printed
```

## §D.1 Severity

| AC | Severity | 근거 |
|---|---|---|
| AC-FG-001/002 | MUST | 게이트의 존재 의의 — 이진 판정 실패 시 게이트 아님 |
| AC-FG-003/004 | MUST | 착지 순서 위반 = CI 전면 적색 사고 |
| AC-FG-005 | MUST | 배포 표면 오염은 템플릿 중립 CI 가드 위반 |
| AC-FG-006 | MUST | 로컬/CI 판정 불일치는 개발자 경험 결함 |

## §D.2 Traceability

| AC | REQ | Milestone |
|---|---|---|
| AC-FG-001 | REQ-FG-001 | M1 |
| AC-FG-002 | REQ-FG-002 | M1 |
| AC-FG-003 | REQ-FG-003 | M1 |
| AC-FG-004 | REQ-FG-004 | M1 |
| AC-FG-005 | REQ-FG-005 | 본 SPEC 전체 커밋 범위(문서 + activation) |
| AC-FG-006 | REQ-FG-006 | M1 (단일 activation 커밋 — 리드 결정 D1) |

## §D.3 Edge Cases

- **untracked `.go` 노이즈**: `make fmt-check`은 tracked-files 변형으로 정의 — untracked `.go`가
  판정을 바꾸지 않음을 scratch 파일로 확인(추가 후에도 판정 불변).
- **templ 생성물**: `internal/web/*_templ.go`가 gofmt-dirty가 되는 상황은 게이트 적색으로
  표출되며, 수리는 재생성(`templ generate`)이지 게이트 예외 등록이 아님.
- **testdata fixture**: `internal/navigator/astx/testdata/**`의 `.go`도 게이트 범위임 — fixture
  파스 불가(고의 malformed) 파일은 현재 0건; 등장 시 본 SPEC 개정으로 범위를 재정의해야 함.
- **CI Go 버전 드리프트**: gofmt은 `go-version-file: go.mod`의 툴체인에서 온다. 버전 상승이
  포맷 규칙을 바꾸면 게이트가 붉어질 수 있음 — 그때의 판정은 CI 몫이며 로컬 회피 금지.

## §D.4 Indirect Verification

- CI Lint 잡의 format-gate 스텝 존재: `grep -n 'gofmt' .github/workflows/ci.yml` ≥1히트(스텝 추가의
  정적 증거) — 다만 이것은 스텝 존재 증거일 뿐, 활성(착지) 증거는 AC-FG-003/004가 담당.
- develop push 후 CI Lint 잡 green — 최종 통합 판정(리드가 판독).

## §D.5 Closure Gates (Definition of Done)

- [ ] AC-FG-001..006 전 항목 명령+실측 출력과 함께 progress.md §E.2에 기록
- [ ] activation commit SHA가 progress.md에 기록되어 사후 감사 형태(§C of plan.md)로 재판정 가능
- [ ] develop push 후 CI Lint 잡 green 확인(리드 판독 기록)
- [ ] `internal/template/templates/**` 0변경 유지(AC-FG-005)
- [ ] 본 카드가 내는 커밋에 `gofmt -w`에 의한 `.go` 수정 0건

## §D.6 Forward-Looking

- 본 게이트 활성 후 신규 위반 유입 시: 위반 카드 발행 시점에 `make fmt-check`가 레인 자가검증
  경로가 된다(별도 요구 없음 — M1 산출물의 지속 효과).
- 사용자 프로젝트 배포 표면 포맷 게이트(템플릿 CI / `moai gate` format 스텝)는 별도 카드 —
  본 SPEC 종결이 그 논의를 종결하지 않음을 명시한다.
