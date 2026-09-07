# SPEC-CODEX-GHOST-SKILLS-PRUNE-001 — 진행 기록

카드: t506 · 트리: `.claude/worktrees/t506` · 브랜치: `WT-codex-ghost-skills` · base `ace1c5440`

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` / `plan.md` / `acceptance.md` / `progress.md` (Tier M)
- SPEC ID 정규식 자가검사: `PASS` (Bash 실행, 이 세션)
- ID 중복: 없음 — `.moai/specs/SPEC-CODEX-*` 16개 중 이 ID 없음
- 측정 정본: `.moai/reports/t506/baseline-measurement.md` (재측정하지 않고 인용)
- 표면 결정: `moai clean --codex-skills` (plan §A). 기각 2건 근거 기록됨
- 미결로 남기지 않은 설계 결정 3건: 동사 위치(plan §A), 줄 범위 출처(plan §B), 범위 안 내용물 처분(plan §B.2)

### plan-audit iter-1 — FAIL 0.76, 수리 완료

판정서: `.moai/reports/t506/plan-audit-verdict.md`. blocking 5건 + MP-2, 전부 소스에서 재확인했고 **하나도 틀리지 않았다.** 수리 내역:

| 지적 | 수리 |
|---|---|
| MP-2 | REQ-CGP-004/005 를 모달 형태로 재작성(`spec.md §C.2`) |
| D1 | AC-CGP-013 을 재조립 함수에 대한 **단위** 기준으로 재작성(개행 유/무/CRLF 3픽스처) + 실제 제거를 강제하는 AC-CGP-014 신설. `plan.md §C.1` 에 `splitLines` 의 줄 끝 손실을 확정 사실로 기록하고 REQ-CGP-019 신설 |
| D2 | `plan.md §B.1` 에 파서 extent 주장 정정(원문 유지, 정정 병기). 범위 안 내용물은 정교화가 아니라 **탈락**으로 처분 — REQ-CGP-018 + AC-CGP-003 7-a/b/c |
| D3 | REQ-CGP-020 신설, `§D.2` 에 매핑(고아 해소) |
| D4 | 뮤턴트 적용 지점을 프루너 적격 분기로, RED 판별식을 **테스트 이름**으로 못박음 |
| D5 | `plan.md §F` 에서 skip 선택지 제거, `§F.1` 에 stat 씨앗이 **신규 작업**임을 명시(`doctor_codex.go:426` 은 `os.Stat` 직접 호출) |
| D6 (선택) | AC-CGP-003 에 `path = ""` 행(5b) 추가 |
| D7 (선택) | REQ-CGP-021 + AC-CGP-015 — `clean.go:34` 도움말 수정을 범위 안으로 |
| D8 (선택) | sha256 경계를 "run 단계 첫 명령 직전 / 마지막 명령 직후"로 고정 |

**틀렸다고 판단한 지적은 없다.** D1 의 `splitLines`(`internal/codexwiring/configtoml.go:204-209`), D2 의 `anyTableRe`(`configtoml.go:76`)와 `skills.go:113-141`, D5 의 `doctor_codex.go:426` 을 각각 직접 읽어 확인했다.

REQ 수 17 → 21, AC 수 13 → 15.

### plan-audit iter-2 — FAIL 0.83, 범위 한정 수리 완료

판정서: `.moai/reports/t506/plan-audit-verdict-iter2.md`. 점수는 Tier M 문턱(0.80) 위였으나 blocking 2건(E1 모순, E2 삼킴 반례)이 FAIL 을 만들었다. iter-1 지적 6건은 전부 해소 확인.

| 지적 | 수리 |
|---|---|
| E1 (critical) | REQ-CGP-005 의 둘째 문장을 **근거 문단으로 강등**(원문 보존 + 강등 사유 병기). 적격 의무는 REQ-CGP-004 에 이미 있으므로 잃는 것 없음. 추가로 **REQ-CGP-022(우선순위 조항)** 신설 — 실격이 적격을 이긴다 |
| E2 (critical) | REQ-CGP-018 의 판정식을 **겉모양 → 파서 상태**로 이동. `openDelim` 분기가 소비한 줄은 모양과 무관하게 미인식. `plan.md §B.3` 신설, `plan.md:72` 의 거짓 문장은 취소선으로 보존 + 정정. AC-CGP-003 에 행 7-d(삼킴 반례)와 7-d'(헤더 없는 좁은 변형) 추가 |
| E3 (major) | AC-CGP-003 의 Given/Then 을 "여섯" → **"여덟 행(부류 일곱)"** 으로 정정하고, 세는 단위가 **행**임을 명시 |
| E4 (major) | **AC-CGP-016 신설** — EOF 로 끝나는 적격 항목, 마지막 개행 유/무 두 변형 |
| E5 (선택) | §F 위험표의 스테일한 "M2 의 픽스처" 인용을 AC 참조로 교체하고, 마일스톤 번호를 표에만 두도록 함 |
| E6 (선택) | AC-CGP-014 에 CRLF 변형 추가(제거 × CRLF 합성 축) |

**E2 는 손으로 걷지 않고 실행해서 확인했다.** 판정서 스스로 "did not execute" 라고 Gaps 에 적었고, critical 판정의 근거이므로 재현이 필요했다. `internal/codexwiring` 에 임시 테스트를 두고 `ParseSkillEntries` 를 판정서의 7줄 픽스처에 직접 돌린 결과(2026-09-07, 트리 `ace1c5440`):

```
ENTRIES=1
  [0] path="/gone" enabled=0
VARIANT-ENTRIES=1
  [0] path="/gone" enabled=0
```

`/exists` 는 파서에 보이지 않는다 — 반례는 참이며, 헤더 없는 좁은 변형도 동일하다. 임시 테스트 파일은 확인 후 삭제했다(트리에 남기지 않음).

**틀렸다고 판단한 지적은 없다.** E1 은 내가 MP-2 수리로 심은 결함이고, E3·E5 도 내가 행을 추가하면서 세는 문장을 갱신하지 않아 생긴 것이다.

REQ 21 → 22, AC 15 → 16.

### plan-audit iter-3 — FAIL 0.81, 범위 한정 수리 완료 (최종 편집 회차)

판정서: `.moai/reports/t506/plan-audit-verdict-iter3.md`. E1 해소 확인, E2 의 본질적 손실 차단 확인(7-d 가 두 구현을 실제로 가른다). 감사 루프는 여기서 중단하고 diff 판독으로 대체(리드 지시).

| 지적 | 수리 |
|---|---|
| F2 (major) | 운반체를 boolean → **`FirstUnrecognizedLine` 인덱스**로 교체. `plan.md §B.3` 의 옛 결정은 취소선 보존 + 두 오류 명시. 정당화 전제("파서만 안다")가 (b)·(c)에 대해 거짓임을 `skills.go:133-138` 인용으로 기록. `spec.md` REQ-CGP-018 에 계산·보고 의무 [HARD] 추가 |
| F2 (가족 인식) | `plan.md §B.4` 신설 — D2/E2/F2 를 한 실수의 세 판본으로 명명하고, 네 번째를 위한 판별 질문 하나를 [HARD] 로 남김 |
| F3 (major) | **AC-CGP-017 신설(must-pass)** — 텍스트 재훑기 뮤턴트가 7-d 담당 테스트를 RED 로 만들어야 함. 더해 7-d/7-d' 에 **선행 단언**(`ENTRIES=1`) 추가 — 픽스처 미발화 시 조용한 초록 대신 시끄러운 실패 |
| F4 (minor) | `§D.2` 산문의 "REQ 21 / AC 15" → **22 / 17** |
| F1 (선택, F2 와 함께 무료) | REQ-CGP-018 이 지고 있던 보고 의무를 명문화 — 판정 계산과 보고가 파서의 몫임을 [HARD] 로 |

**틀렸다고 판단한 지적은 없다.** F2 의 핵심 반박(`case inEntry:` 무동작 경로가 빈 줄·주석·미지 키를 구분하지 못한다)은 `internal/codexwiring/skills.go:133-138` 을 직접 읽어 확인했다 — 두 정규식 분기뿐이고 else 가 없다. 내 §B.3 전제가 (b)·(c)에 대해 거짓이었다는 판정이 맞다.

must-pass AC 2건 체제가 됐다: AC-CGP-004(부재 판정 경계) + AC-CGP-017(범위 내용물 경계). REQ 22(불변), AC 16 → 17.

### [HARD] 이 SPEC 에서 `moai spec lint` 초록은 모달리티 근거가 아니다

`internal/spec/lint.go:790` 의 `isModalityMalformed` 는 `WHEN `/`WHILE `/`WHERE `/`IF `/`THE ` 라는 **영문 접두사에만** 반응한다. 한국어 REQ 본문은 어느 접두사도 만족하지 않아 무조건 false 를 돌려주므로, 이 SPEC 에 대한 린트 초록은 **모달리티 축에서 공허하다.**

그 공허함이 실제로 무엇을 놓쳤는지가 기록되어 있다: 린트가 `No findings` 를 내는 동안 plan-audit 은 MP-2 로 GEARS 모달리티 위반 2건을 잡았다. 두 결과는 모순이 아니라 **서로 다른 축**이다.

따라서 이후 어떤 단계에서도 "`moai spec lint` 통과"를 GEARS 준수의 증거로 인용하지 않는다. 그 초록이 실제로 덮는 것은 frontmatter 와 REQ id 축이다.

## §E.2 Run-phase Evidence

증거 정본: `.moai/reports/t506/run-evidence.md` (이 트리, 2026-09-07 실행). 아래 표는 그 파일의 §4 를 옮긴 것이며, 실패 출력 원문과 뮤턴트 적용 지점은 정본에 있다.

커밋: `8c418440d` (M1 — 파서 범위/인식 보고 + 무손실 split/join), `a18ce4e55` (M2~M6 — 판정기·쓰기 경로·명령 배선).

| AC | 판정 | 검증 수단 | 관측 |
|---|---|---|---|
| AC-CGP-001 | PASS | `TestPruneCodexSkillEntriesRemovesMissingAbsolute` | ok |
| AC-CGP-002 | PASS | `TestPruneCodexSkillEntriesRemovesMissingHomeRelative` | ok |
| AC-CGP-003 | PASS | `TestPruneCodexSkillEntriesNeverPruneClasses`(12항목 선행 단언 + 행 1~7c) + `...SwallowedRegistrationSurvives`(7-d) + `...NarrowSurvives`(7-d') + `TestRunCleanCodexSkillsReportsSkippedEntries` | ok |
| AC-CGP-004 | **PASS (must-pass)** | 뮤턴트 1 — indeterminate→missing, 적용 지점 `judgeCodexSkillEntry` stat switch `default:` | GREEN(exit 0) → RED(`TestPruneCodexSkillEntriesNeverPruneClasses`) → GREEN(exit 0) |
| AC-CGP-005 | PASS | `TestPruneCodexSkillEntriesEnabledIsNotAGate` | ok |
| AC-CGP-006 | PASS | `TestRunCleanCodexSkillsDryRunWritesNothing` | ok (`bytes.Equal` — sha256 동일성보다 강함) |
| AC-CGP-007 | PASS | `TestPruneCodexSkillEntriesIgnoresHeaderInsideDocString` | ok |
| AC-CGP-008 | PASS | `TestRunCleanCodexSkillsBacksUpBeforeWriting` | ok — 백업 내용 == 실행 전 설정, 보고에 경로 + sha256 |
| AC-CGP-009 | PASS | `TestRunCleanCodexSkillsFailsOpen` (4 서브테스트) | ok |
| AC-CGP-010 | PASS | `TestSinglePathShapeClassifier` | ok — 두 번째 분류기를 실물로 심어 발화 확인(정본 §5) |
| AC-CGP-011 | PASS | `TestSkillsParserStaysReadOnly` | ok |
| AC-CGP-012 | PASS | `TestCleanCmdRejectsBothScopeFlags` | ok |
| AC-CGP-013 | PASS | `TestConfigLinesRoundTrip` (재조립 함수 직접 호출, 변형 a/b/c + 4) | ok |
| AC-CGP-014 | PASS | `TestPruneCodexSkillEntriesPreservesUntouchedBytes` (3 변형) | ok |
| AC-CGP-015 | PASS | `TestCleanCmdHelpNamesCodexScope` | ok |
| AC-CGP-016 | PASS | `TestPruneCodexSkillEntriesEntryAtEOF` (2 변형) | ok |
| AC-CGP-017 | **PASS (must-pass)** | 뮤턴트 2 — 파서 판정 무시 후 텍스트 재훑기, 적용 지점 `pruneCodexSkillEntries` 실격 판정 자리 | GREEN(exit 0) → RED(`...SwallowedRegistrationSurvives`, `...NarrowSurvives`) → GREEN(exit 0) |

불변 축:

| 불변 | 명령 | 관측 |
|---|---|---|
| 라이브 `~/.codex/config.toml` 불변 (spec §D) | `shasum -a 256 ~/.codex/config.toml`, run 첫 명령 직전 / 마지막 명령 직후 | 양쪽 `9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a` — 동일 |
| 파서 read-only | `TestSkillsParserStaysReadOnly` | 쓰기 호출 0 |
| 분류기 단일 | `TestSinglePathShapeClassifier` | 두 번째 구현 0 |
| 서브에이전트 경계 | `grep -rn 'AskUserQuestion\|mcp__askuser' <touched files>` | 매치 0 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: a18ce4e55
run_status: complete
ac_pass_count: 17
ac_fail_count: 0
must_pass_ac: [AC-CGP-004, AC-CGP-017]   # 둘 다 PASS, 미검증 처분 없음
mutant_gates: 2                           # 각각 GREEN → RED(지목 테스트) → GREEN
preserve_list_post_run_count: 0           # PRESERVE 목록 밖 변경 0
new_warnings_or_lints_introduced: 0       # golangci-lint: 0 issues
cross_platform_build:
  darwin_native: pass                     # go build ./... exit 0
  windows_amd64: pass                     # GOOS=windows GOARCH=amd64 go build ./... exit 0
coverage:
  internal_codexwiring: 89.5%
  internal_cli: 80.7%                     # 패키지 기존 baseline — 미달은 이 카드 소관 밖(§7 Gaps)
verification_scope: ["./internal/cli/...", "./internal/codexwiring/..."]
full_local_suite_run: false               # 로컬 go test ./... 금지 — 전 패키지 판정은 CI
total_run_phase_files: 5
m1_to_mN_commit_strategy: "M1 단독 커밋 + M2~M6 통합 커밋 (2건)"
evidence_path: .moai/reports/t506/run-evidence.md
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: pending-backfill-sync   # 커밋은 자기 해시를 인용할 수 없다 — 후속 커밋에서 백필
sync_status: complete
b12_self_test_a: pass                    # grep -c 'SPEC-CODEX-GHOST-SKILLS-PRUNE-001' CHANGELOG.md → 0 (중복 없음)
b12_self_test_b: pass                    # acceptance.md 고유 AC 17개 == CHANGELOG 가 주장하는 17개
b12_self_test_c: pass                    # CHANGELOG 가 인용한 파일 경로 6개 전부 ls 로 확인
changelog_entry_position: "[Unreleased] > Added, 첫 항목"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (단일 sync 커밋에 병합)"
  plan_md: "해당 없음 — status 축에서 stateless (spec-frontmatter-schema.md § Artifact Statelessness)"
  acceptance_md: "해당 없음 — status 축에서 stateless"
  progress_md: "해당 없음 — frontmatter 블록 없음, 본문 절로 진행을 기록"
mx_tag_validation:
  ran_as: "sync 하위 단계"
  deletion_guard_annotated: true         # judgeCodexSkillEntry 에 @MX:WARN + @MX:REASON (codex_skills_prune.go:57-58)
  anchor_required: false                 # ParseSkillEntries fan-in = 2 (doctor_codex.go, codex_skills_prune.go) — @MX:ANCHOR 의무 문턱 3 미만
  tags_added_this_phase: 0
docs_sync:
  docs_site_4locale: true                # docs-site/content/{en,ko,ja,zh}/utility-commands/moai-clean.md — 4개 파일 전부 289줄, `##` 헤딩 23개로 동일
  readme_4locale: true                   # README{,.ko,.ja,.zh}.md CLI 표의 `moai clean` 행
  hugo_build: "exit 0, 경고 없음"
canary_compliance_check: "해당 없음 — 이 SPEC 은 자기 sync 가 시험할 전향적 정책을 정의하지 않는다"
verdict_path: .moai/reports/t506/verdict.md
```

### 이 SPEC 에서 `moai spec lint` 초록의 범위 (재확인)

`moai spec lint` 가 `No findings` 를 내더라도 그 초록이 덮는 것은 **frontmatter 축과 REQ id 축뿐이다.** `internal/spec/lint.go:790` 의 `isModalityMalformed` 는 `WHEN `/`WHILE `/`WHERE `/`IF `/`THE ` 라는 영문 접두사에만 반응하므로 한국어 REQ 본문에 대해 무조건 false 를 돌려주고, 따라서 **GEARS 모달리티 축에서 이 SPEC 의 린트 초록은 공허하다.** sync 단계에서도 그 초록을 모달리티 준수의 근거로 인용하지 않았다 — 그 축은 plan-audit 3회차가 사람이 읽어 판정한 것이다(§E.1 의 [HARD] 절).
