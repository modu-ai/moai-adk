# PROGRESS: SPEC-AGENT-EMIT-LINEAGE-001

## §E.1 Plan-phase Audit-Ready Signal

- card: t317 · worktree `.claude/worktrees/t317` (branch `WT-agent-emit-lineage`, base `48eb945df`)
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M 산출물 3종 + 워크플로 공통 progress.md)
- Tier: **M** — v0.4.0 재판정(v0.3.0 까지 S). 영향 파일 **전수 열거 5건** ≥ Tier S 의 `< 5 files` 경계 → 실격, Tier M 밴드 `5 - 15` 에 해당. 열거·산술은 plan.md §B.1/§B.2. plan-auditor PASS 임계 0.75 → **0.80**(iter-2 종합 0.82 는 이 임계도 넘는다)
- REQ **7** / AC **7** (Tier M 상한 각 16, 두 축 독립). **개수는 v0.2.0 이후 불변** — v0.3.0·v0.4.0 수리는 전부 기존 본문 확장이며, Tier 를 올린 축은 항목 수가 아니라 파일 수다. AC 중 4건(AC-AEL-001/002/003/006)이 뮤테이션 확립
- 근거 문서: `.moai/reports/t317/measurement.md` (실측 1-10), `.moai/reports/t317/plan-audit-iter1.md`, `.moai/reports/t317/plan-audit-iter2.md`
- 미해결 결정: **없음**. v0.3.0 에서 자동 호출 지점이 운영자 결정으로 종결됐고(`moai doctor` 항목 편입), v0.4.0 이 그 결정에서 **파생되는** 적용가능성 거동(배포 프로젝트 = 적용 불가 → `ok`, 종료 코드 불변)을 요구 층에 못박았다. 반영 지점은 REQ-AEL-004 「Applicability」 + AC-AEL-003 의 v0.4.0 4게이트
- 감사 이력: iter-1 D1-D6 전건 RESOLVED → iter-2 PASS-WITH-DEBT 0.82, 신규 D7·D8 → v0.4.0 에서 수리 → **iter-3 PASS 0.90 (0.74 → 0.82 → 0.90 단조 상승), 감사 iteration 상한 도달 — plan-phase 감사 종료.** 신규 D9-D12 는 전건 optional, blocking 0
- v0.5.0 델타(감사 후 편집, **재감사 대상 아님**): D9 단독 수리. AC-AEL-003 에 「하위 디렉터리 앵커」 게이트 1개 + REQ-AEL-004 적용가능성 술어에 기준점 해석 절 1개. 요구/수락 개수 불변(7 / 7), Tier M 불변
- **부채로 지고 가는 iter-3 결함 3건 (OPEN — 운영자 승인 하에 run-phase 진입)** — run-phase 가 이것들을 새 발견으로 놀라지 않도록 여기 이름으로 남긴다:
  - **D10 (minor, OPEN)** — `plan.md §B.1` 의 "전수 열거 5건" 이 `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` 3본을 빠뜨린다. doctor 항목을 하나 더하면 이 3본은 **반드시** 재생성된다(항목 행 + 그룹 카운터 `8 ok, 3 warn, 0 fail` 을 담고 있음). 실제 편집 파일은 **8건**이며 Tier 는 M 그대로. → M1 종료 조건에 골든 3본 재생성을 포함하고, 갱신 diff 가 새 항목 행 + 카운터 증가에 한정되는지 확인할 것
  - **D11 (minor, OPEN)** — `plan.md §B.1` 의 "이 리포의 doctor 항목은 파일 1개에 사는 것이 규약" 은 과잉 주장이다. MoAI-ADK 그룹 13개 중 자기 파일 보유 2건(`doctor_mcp_version.go:39`, `doctor_disk.go:66`), `doctor.go` 인라인 6건(`:417,:478,:495,:641,:941,:987`). 파일 #2·#3 은 규약이 아니라 **선택 가능한 형태**이며, 어느 쪽을 골라도 Tier 는 M
  - **D12 (minor, OPEN)** — `spec.md` 의 `golden_test.go:285` 인용은 `:284-285` 가 정확하다(`if` 는 `:284`, `want 11` 메시지가 `:285`). 표기 정정만 남았다
- 상태: `draft` — run-phase 미착수. Implementation Kickoff Approval 은 감사 PASS 로 대체되지 않는다

## §E.2 Run-phase Evidence

측정 트리: `.claude/worktrees/t317` @ `f3e5006ce` (branch `WT-agent-emit-lineage`, base `48eb945df`). 아래 모든 출력은 **이 트리 이 실행**에서 관측했다.

결정한 항목명: **`Agent Emit Embed`** (상수 `agentEmitEmbedCheckName`, `mcpServerVersionCheckName` 패턴). `acceptance.md` 의 `<항목명>` 5곳을 이 값으로 치환해 모든 게이트가 문면 그대로 실행 가능하다.

### AC PASS/FAIL 매트릭스

| AC | 판정 | 검증 명령 | 관측 출력 |
|---|---|---|---|
| AC-AEL-001 | **PASS** | 뮤턴트 심고 `make build; echo exit=$?` | `exit=2`. `golden_test.go:109: .codex/agents/moai/manager-git.toml: committed artifact differs from emission (sha256 mismatch)` + `agent-emit drift: … run \`make agents-emit\`` + `make: *** [agents-emit-check] Error 1`. **`go build` 줄 미출력 — 바이너리 컴파일 단계 미도달.** 원복 후 `make build` → `exit=0` |
| AC-AEL-002 | **PASS** | `grep -c 'mutant-t317-a' …/manager-git.toml` / `git status --short` | `1` — 검사가 방출물을 재생성해 손편집을 덮지 않았다. `git status --short` 는 뮤턴트 파일 1건뿐(` M internal/template/templates/.codex/agents/moai/manager-git.toml`), 검사 기인 신규 경로 0 |
| AC-AEL-003 | **PASS** | 아래 7개 게이트 전건 | 아래 표 |
| AC-AEL-004 | **PASS** | `make build; echo exit=$?` (드리프트 0 트리) | `exit=0`, 직후 `git status --short` 출력 없음 |
| AC-AEL-005 | **PASS** | C2 에 뮤턴트 → `make agents-emit` → `go test ./internal/template/agentemit/... -count=1` | 재생성 전 `FAIL … sha256 mismatch` (exit=1) → `make agents-emit` exit=0 → 재실행 `ok github.com/modu-ai/moai-adk/internal/template/agentemit 0.660s` (exit=0). 재생성 경로가 (a′) 도입으로 막히지 않았다 |
| AC-AEL-006 | **PASS** | `touch <target> && make <target>` × 3 | `agents-emit` / `agents-emit-check` / `embed-check` 셋 다 **레시피를 실제로 실행**(각각 `AGENTEMIT_UPDATE=1 go test …` 출력 / `ok … agentemit 0.241s` / `✓ Agent Emit Embed`). `grep -n '^\.PHONY' Makefile` → 16행에 세 이름 모두 포함 |
| AC-AEL-007 | **PASS** | `grep -rn "agents-emit\|agentemit" --include="*.md" .claude/ .moai/docs/ CLAUDE.md CLAUDE.local.md` | `grep_rc=0`, 5행 히트(baseline 은 0 히트). `CLAUDE.local.md` §2.0 이 소스 층 경로·방출 층 경로·`make agents-emit` 동사 셋을 각각 2회씩 담는다 |

### AC-AEL-003 게이트 세부 (7건)

바이너리는 **뮤턴트 이전에** 빌드했고, 어느 게이트도 재빌드하지 않았다.

| 게이트 | 명령 | 관측 |
|---|---|---|
| 뮤턴트 사망 | `make embed-check; echo exit=$?` | `exit=2` · `fail Agent Emit Embed  moai embeds stale agent-emit artifacts (11/11 compared): manager-git.toml` |
| 기수 보고 | 위 출력의 `11/11` vs `ls …/*.toml \| wc -l` | 커밋본 **11**, 보고 **11/11** — 일치 |
| 바이너리 부재 | `BIN=/nonexistent/moai make embed-check` | `exit=2` · `fail … no readable binary to judge at /nonexistent/moai (11 committed artifacts to compare)`. "비교 0건 → exit 0" 도달 불가 |
| verb 도달 | `grep -n '^embed-check:' Makefile` / 원복 후 `make embed-check` | `47:embed-check: ## …` (rc=0) · `exit=0` · `ok … 11/11 embedded agent-emit artifacts match the committed set (moai)` |
| doctor 도달 | 뮤턴트 상태 `./bin/moai doctor --check "Agent Emit Embed"` | `root_exit=1` · 항목 상태 `fail` · 카운터 `0 ok, 0 warn, 1 fail` / `Pass 0 Warn 0 Fail 1` — **합 1**, 필터가 한 항목만 남겼다. 원복 후 `exit=0` + `ok` |
| CI 미부착 | `grep -rn 'embed-check\|embed_check' .github/workflows/` | `grep_rc=1` (0 히트). 실패 가능성 확인: `ci.yml` 에 타깃명 1줄 주입 → `mutant_grep_rc=0` (`ci.yml:486`), 즉시 `git checkout --` 원복 → `restored_grep_rc=1` |
| 적용 불가 | 스크래치 배포 프로젝트에서 `doctor --check` + 전체 `doctor` | `init=0`; `…/internal/template/templates/.codex/agents/moai/: No such file or directory` (술어 거짓); 루트 `.codex/agents/moai/*.toml` 은 **11건 존재하나 대조 대상으로 삼지 않았다**; `check_exit=0` + `ok  not applicable: no committed emission set at internal/template/templates/.codex/agents/moai/` + `Pass 1 Warn 0 Fail 0`; `doctor_exit=0` — **종료 코드 불변** |
| 하위 디렉터리 앵커 | `internal/cli` 에서 같은 명령 | 초회 **FAIL 관측 → 수리 → 재측정 PASS**. 아래 「발견한 결함 2」 참조. 최종: `root_exit=1` / `sub_exit=1`, 두 실행 모두 항목 `fail` + 카운터 합 1; 원복 후 `sub_clean_exit=0` + `ok` |
| 공허성 대조 (기록용) | 같은 뮤턴트에 `go test … -run TestEmbedFSPresenceAndByteEquality -count=1` | `exit=0` · `ok … agentemit 0.394s` — **기존 테스트는 여전히 통과한다.** 실측 8 이 죽이지 못한 뮤턴트를 새 검사가 죽인다 |

### RED 증거 (test-first)

| # | RED | 관측 출력 | 처분 |
|---|---|---|---|
| R1 | 신규 테스트 12건 (구현 전) | `undefined: committedEmissionRelDir` / `undefined: findEmbedCheckRoot` / `undefined: checkAgentEmitEmbedAgainst` … `FAIL … [build failed]` | 구현 후 전건 PASS |
| R2 | 상대 BIN 경로 (`TestExtractEmissionViaInit_ResolvesRelativeBin`) | `extractEmissionViaInit with a relative bin path: moai init: fork/exec bin/moai: no such file or directory ()` | 수리 후 PASS |
| R3 | 스트레이 마커 2건 | `findEmbedCheckRoot(".../pkgdir") = ".../pkgdir", want ".../001" — a stray marker must not anchor the walk` / `status = "ok", want fail — an applicable tree must not flip to not-applicable…` | 수리 후 PASS |

### 심은 뮤턴트 (전건 RED 관측 + 원복)

| # | 뮤턴트 | 죽은 테스트 / 관측 | 원복 |
|---|---|---|---|
| M-A | 기수 게이트 무력화 (`if compared < len(committed)` → `if false`) | `TestAgentEmitEmbed_PartialExtractionFails`: `status = "ok", want fail (compared 1 of 2)` | `grep -c mutant` → `0`, 재실행 GREEN |
| M-B | 바이너리 부재를 `CheckOK` 로 | `TestAgentEmitEmbed_MissingBinaryFails`: `status = "ok", want fail (applicable tree, no judgment target)` | 동일 |
| M-C | 적용가능성 술어를 배포 산출물 경로로 교체 | `TestAgentEmitEmbed_NotApplicable_NoCommittedSet`: `status = "fail", want ok (not applicable)` — **SPEC 이 금지한 치환이 실제로 배포 프로젝트를 exit 1 로 뒤집는다는 실측** | 동일 |
| M-D | 루트 상향 탐색 제거 | `TestFindEmbedCheckRoot_WalksUpToMoaiMarker`: `= not found, want the .moai-bearing ancestor` | 동일 |
| M-a | 커밋 방출물에 2줄 주입 (AC-AEL-001) | `make build` exit=2, 컴파일 미도달 | `git checkout --` |
| M-b | 동일 (AC-AEL-003, 재빌드 없음) | `make embed-check` exit=2, `manager-git.toml` 지목 | `git checkout --` |
| M-c | C2 소스에 2줄 주입 (AC-AEL-005) | 골든 FAIL → `make agents-emit` → GREEN | `git checkout --` |
| M-ci | `ci.yml` 에 타깃명 1줄 | 금지 게이트 `grep_rc` 0↔1 반전 | `git checkout --` |
| M-phony | `.PHONY` 에서 세 이름 제거 | ``make: `agents-emit' is up to date.`` exit=0, **레시피 미실행** — AC-AEL-006 의 RED 대조군을 이 실행에서 재관측 | `git checkout --` |

### 발견한 결함 2건 (둘 다 실제 실행이 잡았고, 유닛 테스트만으로는 보이지 않았다)

**결함 1 — 상대 BIN 경로.** 추출은 대상 바이너리를 스크래치를 작업 디렉터리로 삼아 실행하므로, `make embed-check` 의 기본값인 상대 경로 `bin/moai` 가 스크래치 안에서 조회돼 `no such file or directory` 로 실패했다. 스텁 extractor 를 쓰는 유닛 테스트로는 보이지 않는 축이다. exec 전에 호출자 작업 디렉터리 기준으로 절대화하고, 셸 스탠드인 바이너리를 쓰는 회귀 테스트를 붙였다.

**결함 2 — 스트레이 `.moai/` 가 루트 탐색을 가로챈다 (D9 게이트가 실제로 붉었다).** `internal/cli` 에서 돌린 첫 실행이 `ok — not applicable: no committed emission set` / `sub_exit=0` 을 냈다. 저장소 루트는 같은 시점에 `fail` 이었다 — **적용 가능한 트리를 적용 불가로 오판**, 즉 AC-AEL-003 하위 디렉터리 앵커 게이트가 봉쇄하겠다고 선언한 바로 그 형태다.

원인: `internal/cli/.moai/state/`(`config-cache.json`, `kanban/`, `factory/`)가 존재한다. 추적되지 않고 gitignore 돼 `git status` 에도 안 나오는 **테스트 부작용 잔재**이며, `doctor` 골든 테스트가 격리로 방어하는 바로 그 쓰기다. REQ-AEL-004 가 적은 「가장 가까운 `.moai/` 보유 조상」 walk 가 이 잔재에 앵커했다.

수리: walk 의 판정을 **커밋 산출물 보유 여부**로 바꿨다 — 적용가능성 술어가 이름으로 지목하는 것이 바로 그것이고, `.moai/` 는 어느 디렉터리나 획득할 수 있는 마커이기 때문이다. `.moai/` 는 「적용 불가」 사유 두 갈래를 문면상 가르는 데만 쓰이고 아무것도 결정하지 않는다. 두 면(`findEmbedCheckRoot` / 검사 자체)에 재현 테스트를 붙여 RED 를 관측한 뒤 수리했다.

**이것은 요구 문면과 실제 트리의 충돌이며, 위 해석이 REQ-AEL-004 의 술어("the glob … matches one or more paths")와 AC-AEL-003 의 앵커 게이트를 동시에 만족시키는 유일한 해석이다.** 요구 본문을 고치지 않았다 — 문면의 술어는 그대로 두고 그것을 실제로 만족시키는 탐색으로 구현했다. 잔재 자체(`internal/cli/.moai/`)는 이 카드 범위 밖이며 별도 카드 후보로 남긴다.

### 영향 파일 (실제 8건 — D10 이 예고한 대로)

| # | 파일 | 신규/편집 | 마일스톤 |
|---|---|---|---|
| 1 | `internal/cli/doctor_agentemit_embed.go` | 신규 | M1 |
| 2 | `internal/cli/doctor_agentemit_embed_test.go` | 신규 | M1 |
| 3 | `internal/cli/doctor.go` | 편집 (등록 4행) | M1 |
| 4 | `internal/cli/testdata/doctor-light.golden` | 편집 (재생성) | M1 |
| 5 | `internal/cli/testdata/doctor-dark.golden` | 편집 (재생성) | M1 |
| 6 | `internal/cli/testdata/doctor-nocolor.golden` | 편집 (재생성) | M1 |
| 7 | `Makefile` | 편집 (`.PHONY` + `agents-emit-check` + `embed-check` + `build` 선행) | M1·M2·M3 |
| 8 | `CLAUDE.local.md` | 편집 (§2.0 신설) | M3 |

D11 이 옳았다 — 「파일 1개」는 규약이 아니라 선택 가능한 형태였고, 자기 파일 형태를 골랐다(`doctor_mcp_version.go` 선례). D10 이 예고한 골든 3본 재생성도 그대로 필요했고, 갱신 diff 는 **항목 행 1개 + 그룹 카운터**에 한정된다(`8 ok, 3 warn, 0 fail` → `9 ok, 3 warn, 0 fail`, `[Pass 15]` → `[Pass 16]`, **Fail 은 0 불변**). D12(`golden_test.go:285` → `:284-285`)는 표기 정정만 남은 항목이라 손대지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: pending-backfill-final   # M1 742a9485d · M2 6335b731b · M3 9b5d5b10a · fix f3e5006ce
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 8          # plan.md §B.1 예고 5 + D10 골든 3본
l44_pre_commit_fetch: not-run            # 리드가 통합 창을 쥐고 있어 push/rebase 미수행 (지시)
l44_post_push_fetch: not-run             # push 하지 않았다
new_warnings_or_lints_introduced: 0      # golangci-lint run ./internal/cli/... ./internal/template/... → "0 issues."
cross_platform_build:
  darwin_arm64: pass                     # go test ./internal/cli/... ./internal/template/... → exit 0, ok 19/19
  windows_amd64_build: pass              # GOOS=windows GOARCH=amd64 go build ./... → exit 0
  windows_amd64_vet: pass                # GOOS=windows GOARCH=amd64 go vet ./internal/cli/... → exit 0
  linux_amd64_vet: pass                  # GOOS=linux GOARCH=amd64 go vet ./internal/cli/... → exit 0
total_run_phase_files: 8
m1_to_mN_commit_strategy: "4 commits on WT-agent-emit-lineage, no push (lead holds the integration window)"
coverage_new_file:                       # go tool cover -func, internal/cli 패키지 전체 79.6%
  checkAgentEmitEmbed: 100.0
  checkAgentEmitEmbedAgainst: 93.2
  findEmbedCheckRoot: 100.0
  nearestProjectRoot: 100.0
  compareEmission: 93.8
  extractEmissionViaInit: 82.4
  walkUp: 73.3
  committedEmissionSet: 80.0
  resolveEmbedCheckBin: 100.0
subagent_boundary_grep: 0                # grep -rn 'AskUserQuestion|mcp__askuser' <new files> → rc=1
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: pending-backfill-sync   # backfilled in the immediately following commit
sync_status: complete
b12_self_test_a: pass                    # grep -c 'SPEC-AGENT-EMIT-LINEAGE-001' CHANGELOG.md → 0 before emission (no duplicate)
b12_self_test_b: pass-with-note          # AC-ID token sweep of acceptance.md → 8 unique tokens, 7 live
b12_self_test_c: pass                    # every path claimed in the CHANGELOG entry verified with `ls` (rc=0, 9/9)
changelog_entry_position: "[Unreleased] → ### Added, first bullet"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (merged sync close)"
  plan.md: "n/a — no frontmatter block"
  acceptance.md: "n/a — no frontmatter block"
  progress.md: "n/a — no frontmatter block"
canary_compliance_check: n/a             # this SPEC defines no forward-looking policy its own sync must test
```

### B12 self-test b — the count discrepancy, stated rather than smoothed

`grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` returns **8**
(`AC-AEL-001` … `AC-AEL-008`), while the SPEC carries **7** live acceptance criteria.
The eighth token is not an eighth criterion: `AC-AEL-008` survives only inside the
retirement footnote at `acceptance.md:165` and the matching `spec.md` 「폐기 판정」 section,
both of which exist to record that the criterion was withdrawn in v0.2.0. `acceptance.md:3`
and `:165` both state the live total as 7, `spec.md §3`'s REQ→AC coverage table maps 7,
and §E.2's AC matrix judges 7. The CHANGELOG entry therefore states **7**, and this note
is the attribution for why the mechanical count and the stated count differ.

### Sync-phase changes (this commit)

| File | Change |
|---|---|
| `CHANGELOG.md` | `[Unreleased] → Added` entry (first bullet) |
| `internal/cli/doctor_agentemit_embed.go` | `@MX` annotations only — no behavior change (`go build ./internal/cli/` rc=0, `gofmt -l` empty) |
| `internal/template/CLAUDE.md` | build-prerequisite chain corrected + cross-reference to `CLAUDE.local.md` §2.0 |
| `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/spec.md` | frontmatter `status: in-progress → completed` |
| `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/progress.md` | this §E.4 block |

**@MX annotations added (2).** Both are mechanically triggered by
`mx-tag-protocol.md` § When to Add Tags, and nothing else in the file triggered a tag:

| Site | Tag | Trigger that called for it |
|---|---|---|
| `checkAgentEmitEmbedAgainst` | `@MX:WARN` + `@MX:REASON` + `@MX:SPEC` | "If-branches >= 8" — measured 10 |
| `extractEmissionViaInit` | `@MX:ANCHOR` + `@MX:REASON` + `@MX:SPEC` | "External system integration point detected" — it spawns the binary under judgment |

Not added, and why: no function in the file is exported (no public-API-boundary
`@MX:ANCHOR`, no untested-public `@MX:TODO`); measured non-test fan_in is at most 2 for
every function, below the fan_in ≥ 3 `@MX:ANCHOR` trigger; there is no goroutine, channel,
or global-state mutation. No `@MX:DEBT` was added — the carried debt below lives in SPEC
artifacts, not as a bounded simplification in this code.

### Carried debt — still OPEN after sync

This sync does **not** close any of the following. Each was re-read in this tree at
`8d1bfd167` and none was resolved by what shipped:

| Item | Status | Evidence |
|---|---|---|
| D10 — `plan.md §B.1` enumerates 5 impacted files, actual 8 | **OPEN** | `plan.md §B.1` still lists 5 rows; the 3 golden files it omits are recorded only in §E.2's "영향 파일 (실제 8건)" table. `plan.md` body is not manager-docs-editable |
| D11 — `plan.md §B.1`'s "one file per doctor check is the convention" over-claim | **OPEN** | the claim stands verbatim in `plan.md §B.1` row 2; §E.2 records the correction, the claim itself was not edited |
| D12 — `spec.md` cites `golden_test.go:285`, accurate range is `:284-285` | **OPEN** | the citation stands in `spec.md` §2 「공허성 봉투」; §E.2 explicitly recorded it as "표기 정정만 남은 항목이라 손대지 않았다" |
| §G-1 — `internal/cli/.moai/` test residue redirects other `.moai/`-marker upward walks | **OPEN** | this SPEC re-anchored its own walk on the committed emission set (`f3e5006ce`); the residue still exists and other consumers (e.g. `findProjectRoot()` at `internal/cli/glm.go`) are untouched |

D10/D11/D12 are all `plan.md`/`spec.md` **body** content, which manager-docs may not edit
(`spec-frontmatter-schema.md` § Forbidden ownership crossings). Closing them requires
re-delegation to manager-spec; they are recorded here rather than silently carried.

---

## §G Out-of-Scope Observations

### G-1 — `internal/cli/.moai/` 테스트 잔재가 marker 기반 루트 탐색의 의미를 조용히 바꾼다 **[별도 카드 감]**

**관측(이 트리, 이 실행):**

```console
$ ls internal/cli/.moai/state/
config-cache.json  factory/  kanban/
$ find internal/cli/.moai -type f
internal/cli/.moai/state/config-cache.json
internal/cli/.moai/state/kanban/leads.json
internal/cli/.moai/state/factory/workers.json
$ git check-ignore -v internal/cli/.moai/state
.gitignore:280:internal/cli/.moai/	internal/cli/.moai/state
$ git ls-files internal/cli/.moai | wc -l
       0
```

- **untracked 이고 gitignored 다** — `git status` 에 나타나지 않고, 삭제해도 git 안전망이 없다.
- **`internal/cli` 테스트 실행마다 재생성된다.** 생성 주체는 config 캐시의 `MkdirAll` 부수효과(`internal/config/cache.go`, 파일명 상수 `config-cache.json`). 이 사실은 `internal/cli/doctor_golden_test.go:66` 주석이 **이미 명문화**하고 있다 — "a prior test's config load writes `<cwd>/.moai/state/config-cache.json` (the cache's MkdirAll side effect)". 골든 테스트는 `t.Chdir(t.TempDir())` + 캐시 비활성화로 자신만 격리할 뿐, 잔재 자체를 막지 않는다. `kanban/leads.json`·`factory/workers.json` 도 같은 성격의 cwd-상대 상태 쓰기다(mtime 이 서로 다르다 — 서로 다른 실행이 남겼다).

**왜 이 SPEC 범위 밖인데도 적어 두는가.** 이 잔재는 D9 게이트를 실제로 실패시켰다: `.moai/` marker 를 앵커로 한 상향 탐색이 `internal/cli/.moai/` 에서 멈춰, **적용 가능한 트리를 적용 불가로 오판**했다. 이 SPEC 은 앵커를 커밋 방출물로 바꿔 자기 판정만 구했을 뿐, **잔재를 없애지도, 다른 marker 기반 탐색을 보호하지도 않는다.** 이 저장소에는 `.moai/` 를 marker 로 쓰는 상향 탐색이 더 있고(예: `nearestProjectRoot`, `internal/cli/glm.go:1058` 의 `findProjectRoot()`), 그 소비자들은 `internal/cli` 하위에서 실행될 때 같은 방식으로 조용히 잘못된 루트를 잡는다. 침묵이 위험의 본체다 — 오판이 에러가 아니라 **다른 정답**으로 나타난다.

**카드로 뽑을 것(제안, 이 SPEC 에서 수행하지 않음):** (a) `internal/cli` 테스트가 패키지 cwd 에 상태를 쓰지 않도록 격리(golden 테스트가 자기 몫만 한 것을 패키지 전역으로), 또는 (b) `.moai/` marker 상향 탐색 소비자 전수 조사 후 각자의 앵커가 "판정 대상 그 자체"인지 점검. 둘은 배타적이지 않다.
