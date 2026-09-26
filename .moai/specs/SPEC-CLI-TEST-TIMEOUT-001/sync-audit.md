# sync-audit 보고서 — SPEC-CLI-TEST-TIMEOUT-001 (card t1253)

- **Auditor**: sync-auditor (독립 평가 — executor 의 AC matrix 는 주장이며, 본 감사가 재검증)
- **일시**: 2026-09-26 · **트리**: `.claude/worktrees/t1253` · branch `WT-cli-test-duration` @ `7bbd233a9` · base develop `b4f798dcc`
- **Overall Verdict: PASS** — harmonic mean **1.00** (Tier M threshold 0.80)

## Dimension Scores (flat weighted)

| Dimension | Weight | Score | Verdict | 핵심 근거 |
|---|---|---|---|---|
| Functionality | 40% | 1.00 | PASS | AC 9건 전건 재검증(6건은 감사자 직접 명령 재실행, AC-004 는 원시 스트림 독립 재도출) — 아래 Evidence |
| Security | 25% | 1.00 | PASS | config-only 변경, 신규 입력·비밀·네트워크 면 0 — 전체 diff 직독 + OWASP 체크리스트 적용 (moai-ref-owasp-checklist) |
| Craft | 20% | 1.00 | PASS | 최소 diff(레시피 3파일 + SPEC dir), 전 행 도출 주석, `make help` 추출·`bash -n` 유지, drive-by 0 |
| Consistency | 15% | 1.00 | PASS | Conventional Commits + card id 7/7 + 🗿 MoAI trailer 7/7, 소유 매트릭스 준수(전이 2건 모두 정규 주체), 범위 가드 L396/397/532 무결 |

**Harmonic mean = 4 / (1/1.00 + 1/1.00 + 1/1.00 + 1/1.00) = 1.00 ≥ 0.80 → PASS.** must-pass(Functionality·Security) 독립 충족.

## Claim

1. 5개 COVERED 표면(Makefile `test`/`test-verbose`/`test-race-short` 60m, `test-codex-live` 10m, `scripts/ci-mirror/lib/go.sh:25` 60m)과 CLAUDE.local.md §4/§6 레시피(30m)가 명시적 `-timeout` + 도출 주석(D1/D2/D3 + baseline 포인터)을 갖는다 — REQ-TIMEOUT-001~005·010, REQ-DOC-007.
2. CI workflow 는 byte 불변이며 이미 명시적(20m/25m)이다 — REQ-SCOPE-008.
3. 변경 후 `./internal/cli/` 재검증이 exit 0·panic 0·리프 카운트 불변(7594/50)으로 완주한다 — AC-004.
4. §B.3 18행 재고가 현재 트리의 go-test 발견 집합과 1:1 대응한다 — AC-009.
5. 커밋 체인 7건이 카드 id·MoAI trailer 를 모두 갖고, 상태 전이(draft→in-progress @M1, in-progress→completed @sync `3ce9eaa80`)와 D3 backfill(`bfdcbc9fd`·`7bbd233a9`)이 규범대로다.

## Evidence (감사자 직접 실행, 전부 이번 런)

**AC-001 재도출** — `make -n test / test-verbose / test-race-short / test-codex-live` (발췌):
```
go test -race -coverprofile=coverage.out -covermode=atomic -timeout 60m ./... # -timeout 60m = D1 derivation (...)
go test -race -v -coverprofile=coverage.out -covermode=atomic -timeout 60m ./... # ... D1 ...
@go test -race -short -timeout 60m ./... || (echo "preflight: test-race-short FAIL"; exit 1) # ... D1 ... -short subset unmeasured — spec.md §C.3
MOAI_CODEX_LIVE_PROBE=1 MOAI_AUDIT_PIN_LIVE=1 go test -timeout 10m ./internal/cli/ -run 'Live' -v -count=1 # ... D3 explicit pin ...
```
`grep -c '\-timeout' Makefile scripts/ci-mirror/lib/go.sh` → `Makefile:4`, `go.sh:1` → 5/5.

**AC-002 재도출** — `grep -n 'go test' CLAUDE.local.md` → L265·L394 류 각각 `` `go test -timeout 30m ./internal/<pkg>/...` `` + D2 포인터; L396/397(`-count=1 ./...`/`-race ./...`)·L532 는 diff 미등장 = byte 무결; L161/164(row 18 서술 prose) 무손대.

**AC-004 원시 스트림 독립 재도출** (8.15MB JSONL, jq/grep):
```
jq -r 'select(.Action=="pass" and .Test==null) | ...Elapsed' → PACKAGE_PASS .../internal/cli Elapsed=1435.829
jq -r 'select(.Action=="pass" and .Test!=null) | .Test' | wc -l  → 7594
jq -r 'select(.Action=="skip" and .Test!=null) | .Test' | wc -l  → 50
grep -c '"Action":"fail"'            → 0   (exit 1 = 부재)
grep -c 'panic: test timed out'      → 0   (exit 1 = 부재)
wc -c go-test-cli-post.stderr.txt    → 0
```
baseline 7594/50 과 리프 레벨 완전 일치 — progress.md §E.2 기재 수치와 1:1.

**AC-005** — `git diff --name-only b4f798dcc..HEAD` → 9파일(SPEC dir 6 + `Makefile` + `CLAUDE.local.md` + `scripts/ci-mirror/lib/go.sh`); `git diff --stat -- .github/` → 빈 출력, `.github/` 경로 0건.

**AC-003 산술 재계산** — 885.287×3.62=3204.7≈3205s ✓ · 1118.093×2.868=3206.7≈3207s(수렴) ✓ · 1800/1118.093=1.61 ✓ · 3600/3205=1.12 ✓ · top25 합계 재가산=278.58s≈278.6s, 278.58/1118.093=24.9%≈25% ✓. 전 피처 attribution 명기(measure-meta.txt / run 36228023389).

**AC-009 재고 대응** — 현 트리 재발견: Makefile `39 48 52 59 67 104 107 110 180 184 193 194`, `go.sh:24/25`, `check-staged.sh:26`, CLAUDE.local.md `161/164/265/394/396/397/532`, CI workflow → §B.3 18행과 정확히 1:1(무멸 row·무미달 row). `go.sh:24` 는 log 문자열이지 발견 항목 아님.

**도구 게이트** — `moai spec lint SPEC-CLI-TEST-TIMEOUT-001` → `✓ No findings — all SPEC documents are valid`, EXIT=0. `moai spec audit` → 본 SPEC 은 `[INFO] (V3R6) — EraAutoDetected` 1건뿐, drift/MUST-FIX 0. (MUST-FIX 2건은 타 SPEC 소유 — SPEC-CI-FLAKE-SERIES-001·SPEC-FULL-SUITE-DOCTRINE-001, 본 카드 표면 밖의 기존 리포 상태.)

**internal/spec AC 스냅숏 가드 — 감사자 독립 재실행으로 귀속** (선택지 중 재실행을 택함):
```
unset MOAI_FACTORY_WORKER ... GIT_PREFIX && go test -count=1 -timeout 30m ./internal/spec/
→ ok  github.com/modu-ai/moai-adk/internal/spec  139.687s   SPEC-TEST-EXIT=0
```

**커밋 체인** — 7건 전부: 제목 Conventional Commits + SPEC scope, 본문 `Card t1253.`·`Authored-By-Agent`·`🗿 MoAI` (trailer 7/7 집계 확인). `585648d08` 이 spec.md frontmatter `draft→in-progress` 실제 수행, `3ce9eaa80` 이 `in-progress→completed`(해 커밋의 spec.md diff 는 frontmatter status 1행만 — 소유 매트리스의 manager-docs 허용 범위 준수), backfill 2건은 D3 면제 규범대로.

**기타 무결성** — `bash -n scripts/ci-mirror/lib/go.sh` OK(주석 부착이 셸 구문 불파); `make help` 에 `test`·`test-verbose`·`test-race-short`·`test-codex-live`·`ci-local` 전부 존재(도출 주석이 `##` 추출 불파); `grep -c 'SPEC-CLI-TEST-TIMEOUT-001' CHANGELOG.md` → 0(NO-ENTRY 결정 사실과 일치); t1252(`19b5321c1`)는 base·HEAD 어느 쪽 조상도 아님(`git merge-base --is-ancestor` 양쪽 부정) + `internal/cli/main_test.go` 범위 diff 빈 출력 → 코드 축 매칭·파일 분리 전제 모두 검증.

## L1 렌즈 — 처방은 로컬 cli 스위트를 timeout 없이 완주시키는가 (판정: **PASS**)

- **공정성**: 코드 축은 매칭(t1252 가 양 트리 모두에 없음 — 직접 검증). 부하 축은 **비통제** — 사전 측정은 load 27→13 기록(`measure-meta.txt:9,12`), 사후 M3 기록(`go-test-cli-post-summary.txt`)은 load 미기록. 1118.093s→1435.829s(+28%) 델타를 부하/코드로 분해 불가.
- **그러나 결론은 속도 주장이 아니라 메커니즘에 있다**: 명시 상한이 관측 최대 소요를 설계상 웃돈다 — 프로브 35m(2100s)·레시피 30m(1800s) > 관측 최대 1435.829s, race 표면 60m(3600s) > 투영 최악 3205s. 결정적으로, 본 M3 원시 스트림의 `ok ... 1435.826s` + exit 0 은 **600s 기본 상한과 역학적으로 양립 불가**한 출력이다 — 기본 상한이었다면 600s 에 panic 했을 것이다. 즉 같은 커맨드 형태가 600s 선을 넘겨 완주한 것이 메커니즘 작동의 직접 시연이며, t1171 601s panic 사고 기록이 실패 모드 쪽 근거다.
- **과잉 주장 스캔**: 3개 SPEC 산출물 대상 `grep -in 'faster|speedup|speed up|performance|improve'` → 0건. progress.md AC-004 는 카운트 정합만 주장, plan.md 는 35m 프로브 근거를 정직히 기술, D1 은 "headroom thin by construction" 명시, REQ-TIMEOUT-006 재도출 트리거가 정직한 경계로 명문화. verdict.md 도 "부하 변동으로 소요는 커졌으나" 라고 속도 주장 없이 기술. **과잉 인과·성능 주장 없음.**
- 유일한 공정성 유보: D2 헤드룸 1.61x 는 사전 기준선 고정치 — 사후 부하 체제(1435.8s)로는 ~1.25x 로 줄어듦(아래 Residual-risk F2).

## L2 렌즈 — 처방 착지 지점과 lane/CI 양면 도달 (판정: **PASS**)

- **lane 측**: 7개 COVERED 발견 항목 전건 직접 grep 재확인 — Makefile L104/107/110/194 + `go.sh:25` (60m×4·10m×1, 전 행 도출 주석) + CLAUDE.local.md L265/394 (30m + D2 포인터). `make ci-local` → `run.sh` 가 `lib/<lang>.sh` 디스팟치(`run.sh:28` "Dispatching: $lang")로 `lib/go.sh` 도달, 플래그 행 L25 도달 체인 확인.
- **CI 측**: `git diff --stat -- .github/` 빈 출력 — byte 불변, 기존 `-timeout 20m`/`25m` 명시 유지(REQ-SCOPE-008). CI 는 애초 명시적이라 변경 불요라는 SPEC 전제가 사실임을 확인.
- **머지창 실측 표면**: lane 이 실제로 도는 CLAUDE.local.md §4/§6 레시피가 이제 `-timeout 30m` 을 처방(본 감사의 internal/spec 재실행도 그 처방 형태로 수행해 소통 확인).
- **§B.3 ↔ 트리**: 18행 1:1 유지(위 AC-009). 5 COVERED / 1 TRANSITIVE(coverage→test 상속, diff 문맥으로 `coverage: test` 확인) / 7 EXCLUDED / 1 EXTERNAL-EXPLICIT.

## Findings (전부 optional — blocking 0)

- **F1** [Info] [optional] `go-test-cli-post-summary.txt` — M3 사후 기록에 machine load 미기록. 사전/사후 소요 비교(1118.093s vs 1435.829s)가 부하 축에서 비공정하나, 그 델타로 결론을 내는 산출물은 없음(메커니즘 기반). Required fix: 없음. 차기 M3류 프로브는 `measure-meta.txt` 처럼 load_before/after 를 쌍으로 기록할 것(권장).
- **F2** [Info] [optional] spec.md §B.1 D2 — 1.61x 헤드룸은 사전 기준선(1118.093s) 고정. 사후 관측 체제에선 ~1.25x. REQ-TIMEOUT-006 이 재도출을 소유하므로 조치 불요. Required fix: 없음(트리거 명문화로 상쇄됨).
- **F3** [Info] [optional] `scripts/ci-mirror/lib/go.sh:24` — log 문자열 `"step 3/5: go test -race -short"` 은 실제 커맨드(이제 `-timeout 60m` 포함)를 축약 표기. 단, `-count=1` 누락은 변경 전부터의 기존 축약 관행의 연속이지 본 카드가 만든 신규 불일치가 아님. Required fix: 선택적 log 문자열 갱신.
- **F4** [Info] [optional] progress.md §E.4 — 선행 인용 "card t1244 NO-ENTRY" 는 본 감사에서 재검증하지 않음(비상재 근거: NO-ENTRY 결정은 독자 근거로 성립 + CHANGELOG grep 0 확인). Required fix: 없음.

## Baseline-attribution

- 본 감사 전 명령: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1253`, HEAD `7bbd233a9`, branch `WT-cli-test-duration`, 이번 런 실행·관측.
- SPEC 기준선: plan/run 산출물은 `b4f798dcc`(측정)·`b3ade47e3`~`2f458a2c9`(구현) 귀속, 본 감사는 `7bbd233a9` 에서 재검증.
- internal/spec 가드: **감사자 자 재실행 귀속** — 본 런, 본 트리, `ok ... 139.687s` exit 0. (오케스트레이터 사전 실행 113.317s 는 참고치로만 대체됨 — 귀속 갭 해소.)
- 참조만 하고 재획득하지 않은 외부 근거: GitHub Actions run 36228023389 수치(308.601s/885.287s) — measure-meta.txt 경유 인용.

## Gaps

- **AC-004 의 실행 시점 사실 2건은 미재실행**: (a) 실행 시 `-timeout 35m` 플래그 값 자체, (b) 실행 시점 exit code. 다만 원시 스트림이 명시 상한 > 1435.829s 를 역학적으로 강제하므로(600s 기본이었다면 `ok` 출력 불가능) 플래그 '계열'의 작동은 독립 입증. 24분 슬롯 직렬 heavy 재실행은 config-only 변경 대비 비비례라 생략 — 판단 명시.
- 사후 M3 런의 machine load 미기록(F1) — 소요 델타 분해 불가.
- CI comparator 수치와 t1244 선행은 재획득하지 않음(경유 인용).
- `sync_commit_sha` 가 short SHA(`3ce9eaa80`)로 기록됨 — 현행 slot 포맷 lint 통과(lint exit 0)이나, 40자 전체를 요구하는 도구가 미래에 등장하면 정규화 필요(lane verdict.md 잔여위험 인용).

## Residual-risk

- D1 60m 은 직접 로컬 race 측정이 아닌 투영(1.12x 헤드룸, 단일 표본 증폭 3.62x) — 과부하 로컬 `make test` race 가 60m 을 넘을 수 있다. 실패 양상이 종전의 조용한 600s 기본 panic 에서 가시적 panic→REQ-TIMEOUT-006 재도출 경로로 바뀐 것이 이 SPEC 의 개선이지, 무한 헤드룸 보장이 아니다.
- t1252(TestMain 수리, develop `19b5321c1`)가 develop 병합되면 internal/cli 절대 소요·카운트가 합법적으로 이동 — 본 감사의 수치는 pre-t1252 트리 귀속.
- 리포 레벨: `moai spec audit` 의 MUST-FIX 2건(SPEC-CI-FLAKE-SERIES-001·SPEC-FULL-SUITE-DOCTRINE-001)은 본 카드 밖 기존 상태 — 리드가 본 카드 판정과 혼동하지 않도록 명기(후자는 진행 중 t1219 SPEC 으로 보이며 §D 조정 전제와 정합).

## Recommendations

- 승인: develop 병합 창 진행. 병합 트리 재측정은 Go 0·CI 불변 카드이므로 spec lint + 본 감사서 재확인으로 최소화 가능(lane verdict.md 병합 절차와 일치).
- 후속 카드 후보(SPEC §C 기록분, 큐 진입은 운영자 몫): internal/cli 패키지 분할, slow-test 수리.
- 소소한 위생(선택): `go.sh:24` log 문자열에 `-timeout 60m` 반영(F3); 차기 heavy 프로브 템플릿에 load 쌍기록 필드화(F1).

## 판정 요약

| 항목 | 값 |
|---|---|
| Verdict | **PASS** |
| Harmonic mean | **1.00** (Functionality 1.00 · Security 1.00 · Craft 1.00 · Consistency 1.00) |
| Tier M threshold | 0.80 — 충족 |
| Blocking findings | 0 (optional Info 4건: F1~F4) |
| L1 렌즈 | PASS — 메커니즘 기반 결론, 과잉 주장 0, 부하 축 비통제는 Gap 으로 공개 |
| L2 렌즈 | PASS — lane 7표면 + CI 불변 양면 확인, §B.3 1:1 유지 |
