# progress.md — SPEC-SPECLINT-GATE-SIGNAL-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
baseline_tree: dd1439502 (origin/develop tip at plan time)
baseline_measurement: 0 error(s), 4368 warning(s), rc=0 without --strict (.moai/reports/t525/spec-lint-baseline-dd1439502.txt)
kickoff_gate: 기제 모양 (i)-(iv) 선택은 운영자 결정 — plan.md §F M2 / spec.md §2.2

참고: plan-auditor 판정은 이 신호 이후 오케스트레이터가 수행한다.

## §E.2 Run-phase Evidence

### M1 — 판정 (2026-09-08, card t525)

측정 규율: 모든 수치는 **트리 빌드 `go run ./cmd/moai`**(installed 바이너리 e79c010b8 미사용 —
1069 커밋 지연, 리드 지시 2026-09-08). 카운팅/부재 판정은 `/usr/bin/grep`. attribution triple
(명령 / 관측 출력 / 트리 SHA) — 상세 원문 `.moai/reports/t525/{m1-demographics.md,verdict.md,census-b6efc874f.json}`.

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| M1-1 | 인구통계 재도출 (AC-SLGS-001) | `go run ./cmd/moai spec lint --json > .moai/reports/t525/census-b6efc874f.json` | rc=0, `jq length` → **4378** findings (warning 4378, error 0; advisory=true 4376 / advisory=false **2**) | `b6efc874f` |
| M1-2 | 비-strict 경로 초록 | `go run ./cmd/moai spec lint` | `0 error(s), 4378 warning(s)`, RC_NONSTRICT=0 | `b6efc874f` |
| M1-3 | (b) 구조 입증 (AC-SLGS-002 rc=1 쪽) | `go run ./cmd/moai spec lint --strict` | `0 error(s), 4378 warning(s)`, RC_STRICT=**1** | `b6efc874f` |
| M1-4 | 비-advisory 재고 = M4 쌍 | `jq -r '[.[] \| select(.severity=="warning" and (.advisory // false) == false)] \| .[] \| "\(.code)\t\(.file)"' census-b6efc874f.json` | `SpecsDirMissingSpecFile` × 2 (SPEC-V3R4-CC2X-ADOPT-001/002) | `b6efc874f` |
| M1-5 | 수치 동결 없음 (AC-SLGS-004) | `/usr/bin/grep -rnE '4[,.]?3(44|68)' .moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001 internal/spec internal/cli` | 12 히트 — 전부 SPEC 문서 4건의 출처 명시 인용; internal/spec·internal/cli **0** 히트 | `b6efc874f` |
| M1-6 | t518 미착지 (REQ-SLGS-011 전제) | `git merge-base --is-ancestor 6cfcfef00 origin/develop` | 비-조상 (T518_NOT_LANDED) — 흡수 병합 전후 2회 재측정 | `b6efc874f` / `2cee65571` |

M1 측정 전 흡수: `origin/develop` `a849d99d2` → `19cf21408` 병합 (창 67 커밋, `internal/spec/` 변경 0).

**M1 판정 요약** (원문 verdict.md):

- **(a)** 성립·지연 — advisory 재고 4,376건(12 규칙)은 실재하는 부채; t518 착지까지 상환 금지 (REQ-SLGS-011).
- **(b)** 구조 입증 — error 0 + 비-advisory 경고만으로 rc=1. 단, 전제 교정: **비-advisory 재고는 2건**(M4 쌍)이며 "수천 건" 전제는 역사적 표본이 advisory 비-표시 표였던 데서 온 추론으로 기각. era demotion(`dd644b5e0`, 2026-07-21)이 이미 대량 재고를 흡수한 상태. M4가 쌍을 "왜"와 함께 닫으면 bare `--strict`는 0 비-advisory → 초록(신호 회복). M2 기제(iii)의 잔여 가치: 규칙별 델타 가시성 + t518 인구 이동 흡수(M3 게이트 재기준). — kickoff 승인 범위 변경 아님, 전제 교정 보고.
- **(c)** 소관 외 — advisory 경계 자체는 t518 소관; 측정된 t518 스코프 = advisory 표시 4,376건.
- **AC-SLGS-002 mutation 절반**: M2 관측 창 소관 — M1에서 미관측, pending으로 기록 (통과 아님).

크로스플랫폼 빌드: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (M1 커밋 전, tree `b6efc874f`).

### M2 — 기제: 규칙별 기준선 래칫 (2026-09-08, card t525)

측정 규율: 모든 수치는 **트리 빌드**(`go build -o /tmp/t525-moai2 ./cmd/moai`, tree `b09b012ef`)
로 잰다. 부재·카운팅 판정은 `/usr/bin/grep`. attribution triple(명령 / 관측 출력 / 트리 SHA).

산출물: `internal/spec/lint_baseline.go` + `_test.go`(신규), `internal/cli/spec_lint.go`(플래그
3종 + 게이트), `internal/cli/spec_lint_test.go`(신규). 기준선 **파일 자체는 만들지 않았다** —
최초 산출은 M3.2 소관(`ls .moai/spec-lint-baseline.json` → `No such file or directory`).

**확정한 설계 결정** (plan.md §F M2 / 경계 사례 절이 run 단계에 위임한 항목):

| 항목 | 결정 | 근거 |
|---|---|---|
| 게이트 차원 | 규칙별 **비-advisory 경고 수**. advisory 는 기록도 게이트도 하지 않는다 | t518 이 advisory 인구를 통째로 옮기므로, 기록하면 착지 즉시 재기준이 강제돼 감사 절차가 일상이 된다 |
| 파일 위치 | `--baseline <path>` 로 호출자가 지정. 기본 경로 없음 | 기본 경로 폴백은 §D.3 의 "침묵 폴백 금지"와 충돌 |
| 직렬화 | `json.MarshalIndent(indent=2)` + 후행 개행. 규칙 키 정렬(encoding/json 이 map 키를 정렬), OS 경로 미포함 | §D.6 결정론. 플랫폼 무관 바이트 동일성 |
| 판정 순서 | ① error → ② 규칙별 증가 → ③ 통과. 둘 다 exit 1 이므로 **출력이 어느 원인인지 이름 붙인다**(`ERROR-GATED` vs `EXCEEDED`) | exit code 만으로는 순서가 관측 불가 — M-b 뮤턴트가 이를 실증 |
| `--baseline` 없는 파일 | exit 3, 경로를 이름 붙임. 오늘의 동작으로 폴백하지 않음 | 경계 사례 절, 침묵 폴백 금지 |
| `--baseline` + `--json`/`--sarif` | exit 3 거절 | 두 페이로드는 원시 finding 목록이라 기준선 차원이 없다 — 받아들이면 플래그를 조용히 무시하는 것 |
| `--baseline` + `--strict` | exit 3 거절 | spec.md §2.2 가 하이브리드 (iv) 를 기각한 사유(escalation 의미 중복)와 동일 |
| `--update-baseline` 중 error 존재 | 파일은 쓰되 exit 1 | REQ-SLGS-009 — 기준선은 error 를 담지 않으므로 재기준이 깨끗한 run 으로 읽히면 안 된다 |

**§E.2 Run-phase Evidence — M2**

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| M2-1 | RED(컴파일) — 구현 전 | `go test ./internal/spec/ -run 'TestComputeBaselineRules\|TestMarshalBaseline\|TestCompareBaseline\|TestLoadBaseline\|TestWriteAndLoadBaseline'` | `undefined: ComputeBaselineRules` 외 10건, `FAIL ... [build failed]`, rc=1 | `b09b012ef` |
| M2-2 | RED(단언) — 스텁 대상 | 위와 동일 `-v` | `rule count: got 0 (map[]), want 2` · `ErrorCount: got 0, want 1` · `expected an error for a missing baseline file` 등 8건 FAIL, rc=1 | `b09b012ef` |
| M2-3 | 공허한 초록 적발·제거 | 위 `-v` 관측 | `TestMarshalBaseline_Deterministic…` 가 스텁의 `(nil, nil)` 에 **PASS**. 비어있지 않음 가드 추가 후 재측정 → `MarshalBaseline produced no bytes — equality below would assert nothing`, rc=1 | `b09b012ef` |
| M2-4 | RED(CLI) | `go test ./internal/cli/ -run 'TestSpecLint' -short` | `unknown flag: --baseline` × 8 FAIL, rc=1 | `b09b012ef` |
| M2-5 | 틀린 이유 초록 적발·제거 | 위 관측 | `TestSpecLintUpdateBaseline_ErrorStillExitsOne` 이 `unknown flag` 의 rc=1 로 PASS. `unknown flag` 배제 가드 추가 후 → `rc=1 came from an argument error, not from the standing error`, rc=1 | `b09b012ef` |
| M2-6 | GREEN(spec) | `go test ./internal/spec/ -run 'TestComputeBaselineRules\|…' -v` | 9/9 PASS, `ok github.com/modu-ai/moai-adk/internal/spec 0.380s`, rc=0 | `b09b012ef` |
| M2-7 | GREEN(cli) | `go test ./internal/cli/ -run 'TestSpecLint' -short` | `ok github.com/modu-ai/moai-adk/internal/cli 3.908s`, rc=0 | `b09b012ef` |
| M2-8 | **AC-SLGS-005** 양방향 + 크로스프로세스 | `go test ./internal/cli/ -run 'TestSpecLintBaseline_CrossProcessBothDirections' -v` | `--- PASS: TestSpecLintBaseline_CrossProcessBothDirections (5.43s)`, rc=0 | `b09b012ef` |
| M2-9 | **AC-SLGS-002 mutation 절반** — `--strict` 쪽 | `/tmp/t525-moai2 spec lint --strict` (실제 저장소 코퍼스) | `0 error(s), 4378 warning(s)`, **RC_STRICT=1** | `b09b012ef` |
| M2-10 | **AC-SLGS-002 mutation 절반** — 기준선 흡수 쪽 | `/tmp/t525-moai2 spec lint --baseline /tmp/t525-real-baseline.json` (동일 코퍼스·동일 바이너리) | `baseline: OK` / `inventory: 4378 warning(s) total (advisory included), 2 non-advisory tracked across 1 recorded rule(s)`, **RC_BASELINE=0** | `b09b012ef` |
| M2-11 | **AC-SLGS-006** 실측 (재고 초록 + 총수 가시) | M2-10 과 동일 실행 | 위 `inventory:` 줄 — 재고가 사라진 게 아니라 흡수됐음이 관측됨 | `b09b012ef` |
| M2-12 | 재기준 기록(**AC-SLGS-008** 정상 경로) | `/tmp/t525-moai2 spec lint --baseline /tmp/t525-real-baseline.json --update-baseline --reason "…"` | `tree_sha: b09b012ef` / `date: 2026-09-08` / `reason: …` / `recorded: 2 non-advisory warning(s) across 1 rule(s)`, rc=0 | `b09b012ef` |
| M2-13 | **AC-SLGS-004** 수치 동결 없음 | `/usr/bin/grep -rnE '4[,.]?3(44\|68\|78)' internal/spec internal/cli` | 무출력, rc=1. 대조군 `/usr/bin/grep -rc 'baseline' internal/cli/spec_lint.go` → `41`, rc=0 (grep 유효 — 위 부재는 실측) | `b09b012ef` |
| M2-14 | 크로스 플랫폼 | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | 각각 rc=0 | `b09b012ef` |
| M2-15 | 커버리지 | `go test -cover ./internal/spec/... ./internal/cli/...` | `internal/spec` **90.3%**, `internal/cli` **81.4%** (전 18 패키지 ok, rc=0) | `b09b012ef` |
| M2-16 | 린트 | `golangci-lint run --timeout=5m internal/spec/... internal/cli/...` | `0 issues.`, rc=0 (착수 전 baseline 도 `0 issues.`) | `b09b012ef` |
| M2-17 | 서브에이전트 경계 | `/usr/bin/grep -rn 'AskUserQuestion' internal/spec/lint_baseline.go internal/spec/lint_baseline_test.go internal/cli/spec_lint.go internal/cli/spec_lint_test.go` | 무출력, rc=1 | `b09b012ef` |
| M2-18 | 전체 대상 패키지 | `go test ./internal/spec/ ./internal/cli/` (크로스프로세스 포함) | `ok internal/spec 70.114s` · `ok internal/cli 443.866s`, rc=0 | `b09b012ef` |

**AC-SLGS-002 판정 갱신**: M1 이 pending 으로 남긴 mutation 절반이 **관측됐다** — 동일 코퍼스·
동일 트리·동일 바이너리에서 `--strict` rc=1 ↔ `--baseline` rc=0. (b) 축 구조 입증 완결.

**뮤턴트 프로브** (verification-completeness §2 — 잡은 것과 **못 잡은 것** 양쪽 기록):

| 뮤턴트 | 내용 | 결과 | 판별 증거 |
|---|---|---|---|
| M-a | 비교가 항상 "증가 없음" 반환 (`case false:`) | **잡힘** | `TestCompareBaseline_IncreaseNamesTheRule` + CLI 2건 FAIL |
| M-b | error 검사를 기준선 검사 **뒤로** 이동 | **처음엔 못 잡음** → 테스트 추가 후 잡힘 | error 단독으로는 `Increases` 가 비어 fall-through 되어 같은 출력. **error 와 증가가 동시에 있을 때만** 순서가 드러난다 → `TestSpecLintBaseline_ErrorOutranksASimultaneousIncrease` 추가, 뮤턴트가 `EXCEEDED` 를 내며 FAIL |
| M-c | `--reason` 을 "존재함"으로만 검사(trim 안 함) | **잡힘** | `RejectsMissingOrEmptyReason/reason_whitespace_only` 만 FAIL — 이 서브케이스가 유일한 판별식 |
| M-d | 감소 경로에서 기준선을 조용히 재기록 | **잡힘** | `DecreasePassesAndLeavesFileByteIdentical` FAIL (sha256 비교) |
| M-e | advisory 를 기준선에 포함 | **잡힘 — 단, 단위 테스트에서만** | `TestComputeBaselineRules_CountsNonAdvisoryWarningsOnly` FAIL. **CLI 층 테스트는 전부 통과** — 가드의 경계: advisory 배제 결정은 CLI 층에서 커버되지 않는다 |
| M-g | `--baseline` 존재 사전검사 제거 | **못 잡음 — 그러나 결함이 아님** | `LoadBaseline` 이 같은 exit 3 + 경로 이름을 낸다. 사전검사는 **중복 가드**이고, 테스트는 구현이 아니라 관측 가능한 계약(exit 3 + 경로)을 단언한다 |
| M-g2 | 진짜 조용한 폴백(로드 실패 시 오늘의 동작으로) | **잡힘** | `FlagContractRejections/missing_baseline_file_…` FAIL |

**Gaps (관측하지 않은 것)**:

- M3 범위 전체 — CI 배선(`.github/workflows/spec-lint.yml` 는 여전히 벌거벗은 `--strict`),
  기준선 파일 최초 산출·체크인, t518 잠금 점검 기록. **이 마일스톤에서 건드리지 않았다.**
- M4(CC2X ADOPT-001/002) 미착수. `SpecsDirMissingSpecFile` 2건은 그대로 서 있다.
- 기준선 직렬화의 **실제 windows 실행** 미관측 — `GOOS=windows` 는 컴파일만 증명한다.
  결정론 근거는 (i) `encoding/json` 의 map 키 정렬, (ii) 레코드에 파일 경로가 없어 구분자
  의존이 없음, (iii) 삽입 순서 독립성 단위 테스트 — 실행 관측은 CI 의 windows 매트릭스 몫.
- 실 저장소 코퍼스에서의 **주입/제거 양방향**은 미관측(합성 코퍼스에서만 관측). 실 코퍼스
  주입은 `.moai/specs/` 를 오염시키므로 하지 않았다.

**Residual risk**:

- `--update-baseline` 은 error 가 있어도 파일을 **쓴 뒤** exit 1 한다. 의도된 설계(재측정
  결과는 남기되 판정은 적색)지만, error 상태에서 재기준한 파일이 커밋되면 그 사실이
  기준선 파일 자체에는 남지 않는다 — 사유 문구와 커밋 메시지가 유일한 기록이다.
- `resolveTreeSHA` 는 git 이 답하지 못하면 `"unknown"` 을 기록한다. 귀속 없는 수를 만들지
  않으려는 선택이지만, `unknown` 이 박힌 기준선이 커밋되면 그 자체가 감사 불가다 — M3 의
  최초 산출은 반드시 git 트리 안에서 실행해 실제 SHA 가 박히는지 확인해야 한다.
- 비-advisory 재고가 현재 2건뿐이므로(M1 판정), 기준선의 실효 가치는 t518 착지 후
  인구 이동을 흡수하는 데 있다. 지금 시점의 래칫은 회귀 방지용으로만 작동한다.

### M4 — 부수: CC2X ADOPT-001/002 (2026-09-08, card t525)

측정 규율: 모든 lint 수치는 **그때-current 트리에서 그때 빌드한 바이너리**로 잰다
(before: `/tmp/t525-m4-moai` @ `185c21cff` 변경 전, after: `/tmp/t525-m4-moai2` @ 같은 HEAD
변경 후 워킹트리). 부재·카운팅 판정은 `/usr/bin/grep`(셸의 `grep` 은 조용히 건너뛰는
ugrep 래퍼다). attribution triple(명령 / 관측 출력 / 트리 SHA).

**순서 기록 — M4 를 M3 보다 먼저 돌렸다.** plan.md §F 는 M3 → M4 순이나 리드 판정으로
뒤집었다. 근거: M4 가 유일한 비-advisory 발견 2건을 닫으므로, M3 의 기준선을 M4 뒤에
산출하면 규칙 집합이 비고 이후의 **어떤** 비-advisory 경고도 즉시 적색이 된다. 먼저
산출했으면 `SpecsDirMissingSpecFile: 2` 가 동결돼 영구 구멍으로 남는다. 범위는 불변,
순서만 이동.

#### "왜 spec.md 가 없는가" — 디렉터마다 따로 (REQ-SLGS-012)

두 디렉터는 **둘 다 의도적 안치(deliberate placement)이지 유실이 아니다.** 근거는 각각
다르다 — plan.md §B.5 관측을 run 단계에서 문서를 열어 확인했다.

| | SPEC-V3R4-CC2X-ADOPT-001 | SPEC-V3R4-CC2X-ADOPT-002 |
|---|---|---|
| 판정 | 의도적 안치 | 의도적 안치 |
| frontmatter | 있음 — `spec_id`, `phase: research`, `created: 2026-05-12`, `child_specs:` 17개 열거 | **없음** — `#` 제목으로 시작 |
| 자기 서술 | §0 Purpose 축자: "본 문서 자체는 plan/spec/acceptance를 포함하지 않으며, 17개 child SPEC의 공통 참조 자료로 사용됩니다" | 제목 축자: "CC Upstream Change Analysis — 2.1.237 → 2.1.239 (Umbrella **Research**)" |
| 출처 | 2026-05-12 마스터 리서치(v2.1.0~v2.1.139, 1,500+ 변경 전수 분석) | 2026-08-23 `/harness:release-update` Phase 5 Option C 스윕 |
| 결정적 증거 | `phase: research` + "plan/spec/acceptance를 포함하지 않으며" — spec.md 를 담을 의도가 애초에 없었다 | 같은 스윕이 만든 child stub 둘(`MSGR-001`, `MCP-001`)은 **12필드 frontmatter 를 갖춘 진짜 spec.md 를 받았다.** 스윕은 SPEC 을 만들 줄 알았고, umbrella 는 SPEC 으로 만들지 않았다 |

즉 어느 쪽도 "spec.md 가 있어야 하는데 사라진" 상태가 아니다. `spec.md` 를 보태는 처분은
**AC-SLGS-012 가 금지한 억압의 다른 얼굴**이다 — 작업을 서술하지 않는 spec.md 는 발견만
지우고 답을 만들지 않는다(실제 작업은 17개/2개 child SPEC 이 담고 있다). 따라서 두 디렉터
모두 **`.moai/specs/` 밖으로 이전**이 유일한 정합 처분이다.

#### 처분 — `.moai/research/` 로 이전(삭제 아님)

```
git mv .moai/specs/SPEC-V3R4-CC2X-ADOPT-001/research.md .moai/research/SPEC-V3R4-CC2X-ADOPT-001-research.md
git mv .moai/specs/SPEC-V3R4-CC2X-ADOPT-002/research.md .moai/research/SPEC-V3R4-CC2X-ADOPT-002-research.md
rmdir .moai/specs/SPEC-V3R4-CC2X-ADOPT-001   # rc=0
rmdir .moai/specs/SPEC-V3R4-CC2X-ADOPT-002   # rc=0
```

- **삭제는 금지였고 하지 않았다.** 002 가 자기 정본이라 지목한 `.moai/research/cc-update-2.1.237-to-2.1.239.md`
  는 이 트리에 **없다**(`ls` → `No such file or directory`) — 즉 SPEC 디렉터 사본이 유일본이었다.
  `git mv` 로 이력을 따라가게 했다.
- **`rmdir` 이 빠지면 처분이 성립하지 않는다.** 규칙은 `os.ReadDir(.moai/specs)` 로 디렉터
  엔트리를 읽고 `SPEC-*` 이면서 `spec.md` 가 없으면 발화한다(`internal/spec/lint.go`
  `lintSpecsDirRootIntegrity`). 파일만 옮기고 빈 디렉터를 남기면 git 은 조용하지만 발견은
  그대로다.
- 목적지 선정: `.moai/research/` 는 추적되고 ignore 되지 않으며(`git check-ignore` rc=1),
  이미 `cc-update-*.md` 스윕과 `SPEC-V3R4-CI-FASTTRACK-001-*.md` 형태의 SPEC-ID 접두
  리서치를 담고 있다. plan.md §F 가 예시로 든 `.moai/reports/` 계열도 ignore 되지 않으나
  (rc=1), 그쪽은 카드 verdict·audit 의 자리이고 두 문서는 리서치다.
- `lint.skip` 을 쓰지 않았고, 린터에 예외를 넣지 않았으며, 스텁 spec.md 를 쓰지 않았다.

#### 참조 분류 (`/usr/bin/grep -rln "CC2X-ADOPT-00" --include='*.md' --include='*.go' --include='*.yaml' --include='*.yml' .` → 19 파일)

| 분류 | 위치 | 처분 |
|---|---|---|
| (a) 깨지는 경로 참조 | `SPEC-V3R4-CC2X-MSGR-001/spec.md:23`, `SPEC-V3R4-CC2X-MCP-001/spec.md:23` — 마크다운 상대경로 링크 `../SPEC-V3R4-CC2X-ADOPT-002/research.md` | **수리** → `../../research/SPEC-V3R4-CC2X-ADOPT-002-research.md` (같은 커밋) |
| (b) 살아남는 ID 참조 | 같은 두 파일 `:16` `parent: SPEC-V3R4-CC2X-ADOPT-002`; `.moai/research/cc-update-20260515.md:143,151,201` + `cc-update-20260520.md:18,169,208`(후보 SPEC ID 언급); `SPEC-V3R5-STATUSLINE-V2145-001/{plan.md:177,spec.md:193}`(EXCL-3 번들 대상); 001 자신의 `spec_id`/`child_specs` | **미변경.** ID 는 경로가 아니다 — 이전 후에도 그대로 해소된다. `parent:` 는 역사적 포인터로 유효 |
| (c) 역사 기록 | `SPEC-ARTIFACT-STATELESS-001/progress.md:356,357,375,502`; `.moai/reports/t252/discriminant.md:96,180`; `.moai/reports/t365/verdict.md:24,25,53,54,143`; `.moai/reports/t525/{plan-audit,m1-demographics,verdict}.md`; 이 SPEC 의 `{spec,plan,acceptance}.md` | **미변경.** 완료된 SPEC 의 진행 기록과 카드 verdict 은 그 시점의 관측이다. 소급 수정하면 기록이 낙관으로 휜다 |
| (d) 스테일해지는 주석 | `internal/spec/lint_test.go:1019` — 픽스처를 설명하는 **주석 1줄**. 픽스처 자체는 합성(`SPEC-BLIND-001` in `t.TempDir()`)이라 실 디렉터에 의존하지 않는다 | **갱신**(과거형 + 이전 사실 명기). 테스트 동작은 불변 — `go test ./internal/spec/...` `ok` |

문서 안 자기참조 2건도 손봤다: 001 은 §0 앞에 이전 사유 1문단, 002 는 존재하지 않는 경로를
정본이라 주장하던 헤더를 실제 위치로 정정했다(거짓 주장 → 참인 주장). 002 의 `:108`
`Prior sweep: .moai/research/cc-update-2.1.236-to-2.1.237.md` 는 **다른 파일**에 대한 역사
기록이라 건드리지 않았다(그 파일도 이 트리에 없다 — 같은 dev-only untracked 계열이다).

#### §E — 자가 검증

| # | Claim | Command | Observed output | Tree SHA |
|---|---|---|---|---|
| E1 | **AC-SLGS-012**: `SpecsDirMissingSpecFile` before 2 → after **0** | `<bin> spec lint --json`, 그 뒤 `jq -r '[.[] \| select(.code=="SpecsDirMissingSpecFile")] \| length'` | before `2` / after `0` | `185c21cff` (before=변경 전, after=변경 후 워킹트리) |
| E1-ctrl-A | 세는 식이 작동한다(양성 대조) | 같은 식, `.code=="MovingRefUnpinned"` | before `115` / after `115` — 0 이 아니므로 식은 살아 있다 | 〃 |
| E1-ctrl-B | 세는 식이 거짓 양성을 내지 않는다(음성 대조) | 같은 식, `.code=="NoSuchRuleCodeXYZ"` | before `0` / after `0` | 〃 |
| E1-text | AC 축자 명령 형태에서도 0 | `<bin> spec lint` → `/usr/bin/grep -c SpecsDirMissingSpecFile` | before `2` / after `0` | 〃 |
| E1-total | 총 경고 정확히 −2 (부수 피해 없음) | `<bin> spec lint` 마지막 줄 | before `0 error(s), 4378 warning(s)` / after `0 error(s), 4376 warning(s)`, 양쪽 rc=0 | 〃 |
| E1-new | 목적지에서 새 발견이 생기지 않았다 | `jq -r '[.[] \| select(.file \| contains("/.moai/research/"))] \| length'` | `0` | 〃 |
| E1b | **비-advisory 재고 after = 0** (M3 의 기준선 입력) | `jq -r '[.[] \| select(.advisory != true)] \| length'` + 코드별 group_by | before `2` (`SpecsDirMissingSpecFile: 2`) / after `0` (남은 항목 없음) | 〃 |
| E1c | `--strict` 가 지금 rc=0 | `<bin> spec lint --strict` | `0 error(s), 4376 warning(s)`, `rc=0` | 〃 |
| E2 | 크로스 플랫폼 빌드 | `go build ./... > f 2>&1; echo "rc=$?"` / `GOOS=windows GOARCH=amd64 go build ./... > f 2>&1; echo "rc=$?"` | `host-build rc=0` / `windows-build rc=0`, 양쪽 출력 없음 | 〃 |
| E3 | 대상 패키지 테스트 | `go test ./internal/spec/... ./internal/cli/...` | `rc=0`; `ok internal/spec 68.214s`, `ok internal/cli 403.381s`, 하위 16개 패키지 `ok (cached)` — FAIL 0 | 〃 |
| E5 | 린트 청결 | `golangci-lint run --timeout=5m internal/spec/... internal/cli/...` | `0 issues.`, `rc=0` | 〃 |

파이프 없이 종료코드를 읽었다 — 이 셸은 zsh 라 `${PIPESTATUS[0]}` 가 비어 판정이 조용히
사라진다. `cmd > file 2>&1; echo "rc=$?"` 형태만 썼다.

**이 §E 절을 progress.md 에 쓴 뒤 재측정했다** — 이 파일도 `.moai/specs/` 안이라 lint 대상이고,
위 after 수치는 이 절을 쓰기 **전** 측정이었다. 재측정 결과 동일: `SpecsDirMissingSpecFile` `0`,
비-advisory `0`, 대조군 `MovingRefUnpinned` `115`, 총계 `0 error(s), 4376 warning(s)` rc=0.
즉 이 절 자체는 새 발견을 만들지 않았다.

**Gaps (관측하지 않은 것)**:

- **변경 전 `--strict` 의 rc 를 내가 직접 재지 않았다.** 배차문이 전한 M1 실측
  (`0 error(s), 4378 warning(s)`, rc=1)은 인계받은 값이지 이 런의 관측이 아니다. 따라서
  "재고 2 → 0 이 `--strict` rc 를 1 → 0 으로 뒤집었다"는 **인과 주장은 절반만 실측**이다 —
  after rc=0 은 내 관측, before rc=1 은 전달값이다. 인과 자체는 소스에서 기계적으로 자명
  하나(`lintSpecsDirRootIntegrity` 가 그 두 디렉터를 이름으로 지목했고 규칙은
  `.moai/specs` 엔트리만 읽는다), 그것은 코드 읽기이지 실행 관측이 아니다.
- 커버리지(E4/E3-cover)를 재지 않았다 — M4 는 Go 동작을 바꾸지 않았다(주석 1줄).
- windows/darwin 매트릭스에서 **테스트 실행**은 관측하지 않았다. E2 는 컴파일만 증명한다.
- 이전된 두 문서가 이 저장소 **밖**(docs-site, 외부 링크, 다른 체크아웃)에서 옛 경로로
  참조되는지는 조사하지 않았다 — 리포 내부 스윕만 했다.

**Residual risk**:

- 옛 경로(`.moai/specs/SPEC-V3R4-CC2X-ADOPT-00{1,2}/research.md`)를 기억하고 찾아오는
  독자는 빈손이 된다. 완화: 두 문서 모두 헤더에 이전 사유와 옛 경로를 적었고, `git log
  --follow` 가 이력을 따라간다. 그러나 **옛 경로에 남는 표지는 없다** — 디렉터를 지웠기
  때문이다(남기면 발견이 되살아난다).
- E1b 가 0 이라는 것은 M3 의 기준선이 **빈 규칙 집합**이 된다는 뜻이다. 그 상태의 래칫은
  회귀 방지로만 작동하고, t518 착지 시 advisory 인구가 움직이면 재기준이 필요할 수 있다.
- `parent: SPEC-V3R4-CC2X-ADOPT-002` 는 이제 `.moai/specs/` 에 실재하지 않는 디렉터를
  가리키는 ID 다. ID 참조로서는 유효하나, 장래에 `parent:` 를 디렉터 존재로 검증하는 규칙이
  생기면 두 stub 이 발화한다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### 잔여 실측 열거 (M1 이후, 2026-09-08, tree `b09b012ef`)

리드 지시(2026-09-08)에 따라 M2 이후 잔여를 문서 추론이 아니라 실측으로 확정한다.
카운팅·부재 판정은 `/usr/bin/grep`(셸 `grep` 은 ugrep 래퍼라 조용히 건너뛴다).

| # | 대상 | Command | Observed output | 판정 |
|---|---|---|---|---|
| R-1 | M2 기제 (CLI 측) | `/usr/bin/grep -c baseline internal/cli/spec_lint.go` | `0`, rc=1 | 미착지 |
| R-2 | M2 기제 (린터 측) | `/usr/bin/grep -c baseline internal/spec/lint.go` | `0`, rc=1 | 미착지 |
| R-2c | R-1/R-2 대조군 | `/usr/bin/grep -c strict internal/cli/spec_lint.go` | `3`, rc=0 | grep 유효 — 위 `0` 은 실측된 부재 |
| R-3 | 기준선 파일 | `ls -la .moai/spec-lint-baseline.json` | `No such file or directory` | 미생성 |
| R-4 | 신규 소스 파일 | `ls internal/spec/lint_baseline*.go` | `no matches found` | 미생성 |
| R-5 | CI 배선 | `/usr/bin/grep -n 'spec lint' .github/workflows/spec-lint.yml` | `58:        run: go run ./cmd/moai spec lint --strict` | 벌거벗은 `--strict` — 미배선 |
| R-6 | M4 대상 디렉터 | `ls .moai/specs/SPEC-V3R4-CC2X-ADOPT-001/ -002/` | 각각 `research.md` 1건뿐 | 미종결 |
| R-7 | t518 착지 (REQ-SLGS-011 게이트) | `git merge-base --is-ancestor a4fbaeb82 origin/develop` | rc=1 | **비-조상 — 미착지** (`origin/develop` = `91d25bc61`) |

**잔여 = M2 전체 + M3.1~M3.3 + M4.** M3.4(DAG 간선 `dependencies:` 추가 + 게이트된 재기준
1회)는 R-7 로 잠금 유지 — 이 카드에서 열지 않고 t518 착지 시 별건으로 연다.
M1 에서 pending 으로 남긴 AC-SLGS-002 mutation 절반은 M2 기준선 경로 착지 후 관측한다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters** (orchestrator, 2026-09-08):

- tier: M
- scope (files): ~8 (internal/spec/lint.go, internal/cli/spec_lint.go, .github/workflows/spec-lint.yml, baseline file, tests, docs)
- domain count: 2 (Go linter/CLI policy surface, CI wiring)
- file language mix: Go + YAML + JSON baseline
- concurrency benefit: LOW — coding-heavy, sequential dependency M1 verdict → M2 mechanism → M3 wiring
- Agent Teams prereqs: not requested

**Mode evaluation table**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | Multi-file Go + CI change, not a typo fix |
| serial | **yes** | Coding-heavy Tier M; M1→M4 ordered dependencies; single writer per milestone |
| fanout | no | Not research-heavy; coding-task parallelism caveat |
| sweep | no | <30 files, semantic (not mechanical-uniform) change |

Decision: serial
Justification: Anthropic coding-task parallelism caveat applies — the mechanism (M2) depends on the M1 demographic verdict, and CI wiring (M3) depends on the mechanism; nothing downstream of M1 is parallelizable. One manager-develop spawn per milestone, sequential.

Kickoff outcomes (operator decisions, 2026-09-08, AskUserQuestion round):

- kickoff: **approved** (plan-audit PASS 0.875 prerequisite met)
- mechanism (spec.md §2.2 reservation): **(iii) per-rule baseline ratchet** — CI gates via `--baseline <checked-in file>`; `--strict` retained for non-CI use
- progression mode: **autonomous** — ac_converge goal armed by the orchestrator after this gate; no per-milestone operator prompts; evidence reported at close
