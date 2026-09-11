# SPEC-HARNESS-EVIDENCE-WRITE-001 — sync-audit verdict (card t569)

- 감사자: sync-auditor (독립 감사 — 레인 산출물을 신뢰하지 않고 핵심 검증을 전부 재실행함)
- 날짜: 2026-09-08
- 측정 트리: worktree `.claude/worktrees/t569` · branch `WT-harness-evidence-write` · HEAD `3db94543e` (sync commit)
- diff 베이스: `fccc31d7e` (plan-phase commit) — 모든 diff 귀속은 `git diff fccc31d7e..HEAD`
- 평가 프로필: 기본(default) — 가중치 Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security, 하트닉 평균

---

## Overall Verdict: **PASS** — 96.9/100 (harmonic mean)

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence (이번 감사 실행의 원본 출력) |
|-----------|-------|---------|----------|
| Functionality (40%) | 100/100 | PASS (must-pass) | 10개 AC 식별자 전부 감사자가 직접 재실행해 관측 — 아래 AC 표. 대표: `--- PASS: TestNoTestWritesRepoTree (0.01s)` / `measurement written to /var/folders/.../TestCorpusREQWideningMeasurement4241157098/001/m1-corpus-measurement.txt` / no-clobber 재실행에서 `--- FAIL` + sha256 `234dc16585962e12...f9a93` BEFORE=AFTER |
| Security (25%) | 100/100 | PASS (must-pass) | test-only 변경 (`git diff fccc31d7e..HEAD --stat` — 비-테스트 소스 0건). 비밀값·외부입력·네트워크 노출 없음. 증거 무결성 통제가 오히려 강화됨: no-clobber(`os.Stat` → `t.Fatalf`)가 연산자 실수에 의한 핀 고정 증거 파괴를 차단. 파일 모드 `0o644`/`0o755` 최소 권한 |
| Craft (20%) | 92/100 | PASS | `go test -cover ./internal/spec/ -count=1` → `coverage: 90.5% of statements` (임계 85% 충족, 재측정값 = 레인 주장과 일치). `golangci-lint run ./internal/spec/...` → 기존 baseline 1건(`zz_t528_overacceptance_test.go:100` errcheck, PRESERVE 파일)뿐, 신규 0건. `go vet` 네이티브 OK, `GOOS=windows GOARCH=amd64 go vet ./internal/spec/` exit 0 (테스트 파일 크로스컴파일 포함), `gofmt -l internal/spec/` 빈 목록. 오류 처리 충실 — guard의 post-loop `sc.Err()` 조기-중단 검출(guard 146-148행), `errors.Is(err, fs.ErrNotExist)` 판별. 감점: F1-F3의 guard 완결성 공백 |
| Consistency (15%) | 96/100 | PASS | 네이밍(`t362ReportPath`/`t362EvidenceOutEnv`가 기존 `corpusScanEnv` 게이트 네이밍과 일관), 주석 영어, Conventional Commits + SPEC-ID 스코프 전 커밋 확인, 카드 id `t569`가 sync 커밋 제목에 존재, CHANGELOG B12 셀프테스트(`grep -c` = 1). 소유권 체인 정합: M1 커밋 `db16e5be4`가 `draft → in-progress`, sync 커밋이 `in-progress → completed` (status 1행만). 감점: F5 표기 경계 |

---

## AC별 판정 (10 식별자 — 전부 감사자 재실행)

| AC | 판정 | 감사자가 실행한 명령 (요지) | 관측 원본 출력 (요지) |
|----|------|---------------------------|----------------------|
| AC-001 | PASS | `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/ -run TestCorpusREQWideningMeasurement -count=1 -v` | `measurement written to /var/folders/.../TestCorpusREQWideningMeasurement4241157098/001/m1-corpus-measurement.txt` · `--- PASS`; 실행 후 `git status --porcelain .moai/reports/t362/ .moai/reports/t528/` 빈 출력 (exit 0) |
| AC-002 | PASS | 동일 형태, `-run TestCorpusRejectedREQIDDecomposition` | `decomposition written to /var/folders/.../TestCorpusRejectedREQIDDecomposition2006101515/001/m2-gate0-decomposition.txt` · `--- PASS` |
| AC-003 | PASS | `unset ... && go test ./internal/spec/ -run TestT528Anchor -count=1 -v` | `OUTDIR = /var/folders/.../TestT528Anchor3936807198/001` · `--- PASS` |
| AC-004 | PASS | `T528_PROBE_OUT=/tmp/t569-audit-probe-out go test ... -run TestT528Anchor -count=1 -v` (신선한 리터럴 디렉터리) | `OUTDIR = /tmp/t569-audit-probe-out` · `--- PASS` · `ls -la` 로 8개 파일 확인 (baseline-needle, blind-files, filelist, newly-accepted, positive-needle, reqmap, rootac, still-rejected) |
| AC-005 | PASS | `go test ./internal/spec/ -run TestNoTestWritesRepoTree -count=1 -v` | `--- PASS: TestNoTestWritesRepoTree (0.01s)` / `ok ... 0.377s` |
| AC-006 | PASS | 감사자 독자 변이 2종 주입 → guard 실행 → 제거 → 재실행 | 변이 #1(직접 리터럴): `zz_t569_audit_mutant_scratch_test.go:14: os.WriteFile writes a repo-anchored path (direct .moai/ literal)` → `--- FAIL`. 변이 #2(**원 결함 형상 130846ab2 — const → Join 별칭 → 쓰기**: 레인이 시연하지 않은 더 강한 형상): `:19: os.MkdirAll writes via "out" (repo-anchored)` + `:22: os.WriteFile writes via "out" (repo-anchored)` → `--- FAIL`. 제거 후 `--- PASS` 복귀, `git status --porcelain internal/spec/` 빈 출력으로 트리 복원 확인 |
| AC-007 | PASS | `git diff fccc31d7e..HEAD --stat -- .moai/reports/` + `git ls-files .moai/reports/t362/` + diff stat 전수 | `.moai/reports/` diff **빈 출력** · t362 트래킹 12파일 · `ac_count_clause_test.go`/`zz_t528_overacceptance_test.go` diff 부재(미수정) |
| AC-008a | PASS | `MOAI_T362_CORPUS_SCAN=1 MOAI_T362_EVIDENCE_OUT=/tmp/t569-audit-evidence-out go test ... -run TestCorpusREQWideningMeasurement -count=1 -v` | `measurement written to /tmp/t569-audit-evidence-out/m1-corpus-measurement.txt` · `--- PASS` · 해당 경로에 파일 착지 확인 |
| AC-008b | PASS | 동일 호출 재실행(기존 파일 존재) — shasum 전후 대조 | `lint_req_widen_corpus_test.go:246: MOAI_T362_EVIDENCE_OUT no-clobber: target ... already exists; refusing to overwrite — durable capture must write a NEW file` · `--- FAIL` · sha256 `234dc16585962e12dcd92560dbb3d459efab3eed42d1c84192a64c964c3f9a93` BEFORE=AFTER (진행 기록이 인용한 다이제스트와도 일치 — 리포트 내용이 결정론적임의 부수 확인) |
| AC-008c | PASS | `head -8` 리포트 + `/usr/bin/grep -c '\.moai/reports'` | 헤더에 `# output location: a per-run t.TempDir() directory` + `# durable capture: set MOAI_T362_EVIDENCE_OUT=<dir> BEFORE the run` 명시 · `.moai/reports` 언급 0건 (grep count 0). (`scan_glob=.moai/specs/SPEC-*/spec.md` 는 스캔 **입력** glob이지 출력 기본경로가 아니다 — AC 문언과 저촉 없음) |

**게이트 보존 (비-AC 확인)**: 게이트 없이 페어 실행 → `--- SKIP` 2건 + `PASS` / `ok ... 0.209s` — `MOAI_T362_CORPUS_SCAN` 옵트인 의미론 보존 확인.

---

## Guard 적정성 (과업 지시 3번 항목)

- **`sc.Err()` 사후 검사 존재**: `guard_no_repo_tree_write_test.go` 146-148행 — 스캐너 중단-조기-종료를 finding으로 승격해 guard 자신의 공허 초록을 막음. 확인.
- **문서화된 경계의 정직성**: 문서화 5종(동적 조립 경로·파라미터 운반 앵커·다중-행 문장·자기-스캔 제외·비-테스트 소스)은 실제 한계와 일치. `filepath.Join(tempDir, ".moai", "logs")` 형태의 temp fixture가 오검출되지 않는 근거(bare segment vs `.moai/` 슬래시 리터럴)도 서술돼 있음. 정직하다고 판정.
- **변이 의무와 술어의 정합**: progress.md §E.2.1의 변이(직접 리터럴)는 guard 술어 3번 분기로 잡히며, 감사자가 추가로 **별칭-체인 형상**(원 결함 `130846ab2`의 실제 모양)을 주입해 술어 2번(식별자 체이닝)도 실동작을 관측 — 두 쓰기 지점 모두 적발. 진행 기록의 변이 증거와 guard 술어는 상호 정합.

---

## Findings (전부 optional — blocking 0건)

- **F1** [low] [optional] `internal/spec/guard_no_repo_tree_write_test.go:72-78` — 원시 리스트가 `os.Rename` / `os.Remove` / `os.RemoveAll` / `os.Truncate` / `os.CreateTemp` 을 빠뜨린다. `os.Rename(tmp, repoPath)` 나 `os.RemoveAll(repoPath)` 는 트래킹 증거를 변형·삭제하면서 finding 없이 통과한다. 시연된 결함 형상(상수-앵커 WriteFile/MkdirAll)은 커버되므로 blocking 아님. 수정 제안: 단일-인수로 모호성 없는 `os.RemoveAll(` / `os.Remove(` 를 먼저 원시 리스트에 추가하고, `os.Rename` 은 후속 SPEC에서 대상-인수 해석과 함께.
- **F2** [low] [optional] 같은 파일 — 미문서화 우회 형상 2종: (a) 쓰기 원시 별칭화(`wf := os.WriteFile` 후 `wf(...)`) — 원시 마커가 같은-행 괄호를 요구하므로 미적발, (b) 백틱 raw-string 리터럴의 `.moai/` — 정규식이 겹따옴표만 매칭. 수정 제안: guard 문서 주석의 경계 목록에 두 형상을 명시(정직성 유지)하거나 후속 강화.
- **F3** [low] [optional] 같은 파일 — 잠재 과검출: 슬래시-포함 리터럴을 temp 아래 조인하는 형태(`filepath.Join(t.TempDir(), ".moai/reports")`)는 repo-앵커로 오검출된다. bare-segment 다중-인수 조인만 안전하다는 문서는 있으나 이 잔여 위험은 미기재. 현재 트리에 해당 형태 없음(guard GREEN으로 확인). 수정 제안: 경계 문서에 "temp 하위라도 슬래시-리터럴 조인은 피하고 bare segment를 쓸 것" 추가.
- **F4** [info] [optional] `progress.md` §E.3 이 인용하는 `/tmp/t569/*.txt` 증거 경로는 오늘 기준 존재하지만 `/tmp` 는 OS가 지운다 — 내구 사본은 §E.2.1의 인라인 원본 인용이 담당한다(그 사본으로 충분히 재검증됐음). 수정 제안: 이후 카드는 검증 원본을 `.moai/state/verify/<session>/` 이나 카드 리포트 디렉터리에 보존.
- **F5** [info] [optional] AC 계수 표기 — §E.1 `ac_count: 8`(plan 시점, AC-008 분할 전) vs 실제 식별자 10개(008a/b/c). §E.4 `b12_self_test_b` 가 "progress.md의 count는 쓰지 않는다"고 명시해 오독 위험은 이미 봉쇄됨. Tier S AC 예산(8)도 "AC-008 = 1개 AC, 3개 GWT 시나리오"로 읽으면 8 logical AC로 문안히 성립하며 plan-auditor iter2가 1.0으로 통과한 지점. 위반으로 판정하지 않고 독해 경계만 기록.

---

## 5-섹션 증거 보고 (VCI §3)

- **Claim**: SPEC-HARNESS-EVIDENCE-WRITE-001의 10개 AC 전부 충족; 4차원 감사 PASS(96.9/100); blocking finding 0건.
- **Evidence**: 위 AC 표의 전부 — 감사자가 이번 실행에서 직접 관측한 원본 출력. 핵심 6건: guard GREEN(`ok ... 0.377s`), AC-001 TempDir 경로(`lint_req_widen_corpus_test.go:251`), AC-004 오버라이드(`OUTDIR = /tmp/t569-audit-probe-out` + 8파일), AC-008b no-clobber(`--- FAIL` + sha 동일), AC-006 독자 변이 2종 RED(직접 리터럴 :14 / 별칭-체인 :19+:22)와 제거 후 GREEN, 커버리지 재측정 `coverage: 90.5% of statements`.
- **Baseline-attribution**: 전부 (이번 실행, 이 트리) — worktree `.claude/worktrees/t569` @ `3db94543e`, diff 베이스 `fccc31d7e`. 레인의 /tmp 산출물은 인용하지 않았고(존재는 확인 — 참고용), PASS 근거는 전부 감사자의 재실행이다.
- **Gaps**: (1) `go test ./...` 전체 스위트는 로컬 금지 규율에 따라 미실행 — 통합 판정은 origin/develop CI의 몫. (2) F1에 열거한 미커버 원시들(`os.Rename` 등)에 대한 guard 우회 여부는 변이를 만들어 관측하지 않았다 — 소스-판독 기반 추론이며, 위 표의 다른 모든 관측과 달리 이 항목만 도구-검증된 결함이 아니다. (3) findRepoRoot가 사용되는 다른 패키지(external caller) 영향은 test-only 범위 밖이라 미측정.
- **Residual-risk**: guard는 텍스트 스캔이라 F1-F3의 우회 형상이 언제든 재도입 가능하다 — 다만 재도입 시점에도 기본 경로는 TempDir이므로 증거 파괴는 no-clobber와 git이 이중 방어한다. `MOAI_T362_EVIDENCE_OUT` 을 repo 내부 신선 디렉터로 찍으면 트래킹되지 않은 새 파일이 작업 트리에 생긴다(증거 파괴는 아니고 `git status` 로 보임; §E가 T528_PROBE_OUT 에 대해 명시한 것과 같은 status-quo).

---

## Recommendations

1. (F1) 후속 소규모 카드로 `os.RemoveAll` / `os.Remove` 를 guard 원시 리스트에 추가 + 그 변이로 RED 관측까지 완료해 채택 — verification-completeness §1.1의 관측-실패 완결 축을 그대로 적용.
2. (F2/F3) guard 문서 주석에 우회 2형상 + 슬래시-리터럴-조인 과검출 1형상을 명시 — 코드 변경 없이 guard의 정직한 경계 선언을 완성.
3. (F4) 이후 레인의 검증 원본은 `/tmp` 대신 `.moai/state/verify/` 또는 카드 리포트 디렉터리에 보존.

---

## Scope discipline 확인 (과업 지시 4번 항목)

`git diff fccc31d7e..HEAD --stat` — 선언된 파일 집합과 정확히 일치:

```
 .moai/specs/.../progress.md    | 120 ++++++++++++++-
 .moai/specs/.../spec.md        |   2 +-          (status: draft→completed, 1행 — frontmatter 전용 확인)
 CHANGELOG.md                   |   1 +
 internal/spec/guard_no_repo_tree_write_test.go | 182 +++ (신규)
 internal/spec/lint_req_widen_corpus_test.go    |  24 +-
 internal/spec/lint_req_widen_decompose_test.go |  17 +-
 internal/spec/t362_evidence_out_test.go        |  43 ++ (신규)
 internal/spec/zz_t528_anchor_probe_test.go     |  13 +-
```

범위 이탈 0건. REQ-005 확인: `corpusMeasurementRelPath` / `decomposeReportRelPath` / `t528/probe/out` 잔존 grep = **0매치** — repo-상대 쓰기 경로는 더 이상 구성 불가.
