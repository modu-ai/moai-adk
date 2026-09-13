# sync-audit — SPEC-SPECLINT-GATE-SIGNAL-001 (card t525)

- 감사자: sync-auditor (독립 판정, 읽기 전용)
- 감사 트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`, HEAD `ed4242845` (tree `6a375f6c6f701fe418d05ba32ab79240d02f1c4e`), 착수 시 `git status --porcelain` 0줄
- 판정 바이너리: 이 트리에서 `go build -o <세션 스크래치>/audit/moai-t525-audit ./cmd/moai` (exit 0), 경로로 호출. 설치본 미사용
- 평가 프로필: `default` (SPEC 에 `evaluator_profile` 없음 → `harness.yaml` `default_profile: "default"`)
- 적용 규칙 인용: `verification-claim-integrity.md` §1.1·§2(귀속), `verification-completeness.md` §1.1(관측된 실패)·§2(뮤턴트/잘못된 이유의 초록)

## 종합 판정: **FAIL** (차단 결함 1건 — F1, 수정 범위 작음)

차단 사유는 F1 하나다. 12개 AC 가운데 11개는 증거로 통과했고, AC-SLGS-011 의 로그 절반은 푸시 뒤에만 관측할 수 있어 리드 소관으로 넘어간다(결함 아님). 판정 항목 1(AC-SLGS-010 전반부)은 **PASS** 로 판정했다.

F1 을 고친 뒤의 재감사는 F1 델타(아래 Required fix 의 두 파일 + 교차 프로세스 확인 1회)로 한정한다.

---

## 1. Claim

1. 기준선 래칫(`--baseline`/`--update-baseline`/`--reason`)은 이 HEAD 에서 SPEC 이 요구한 판정 순서(error → 규칙별 증가 → 통과)대로 동작하고, 감소 시 파일을 다시 쓰지 않으며, 사유 없는 재기준을 거절한다 — 실제 CLI 교차 프로세스로 관측했다.
2. 워크플로 배선은 올바르다. strict 분기가 `--baseline` 을 쓰고, 기준선만 바꾼 변경이 기준선 게이트를 건너뛰는 경로는 없다.
3. M3.4 의 간선과 게이트된 재기준 1회는 기록과 일치한다.
4. sync 산출물(CHANGELOG, docs-site 4개 로케일, `status: completed`, §E.4, 트레일러)은 코드와 기록에 부합한다. 다만 CHANGELOG 용어 1건이 부정확하다(F5).
5. **결함(F1)**: 기준선 플래그 거절 경로 전부가 **출력 0바이트로 exit 3** 한다. 경로·원인을 이름 붙인다는 계약은 사용자 표면에서 성립하지 않고, 그 계약을 지킨다는 테스트는 사용자가 볼 수 없는 문자열로 초록이 된다.

## 2. Evidence (명령 + 원문 출력, HEAD `ed4242845`)

### 2.1 게이트·판정 (실 코퍼스)

```
$ <bin> spec lint --baseline .moai/spec-lint-baseline.json > gate.txt 2>&1; echo "gate exit=$?"
gate exit=0
$ tail -5 gate.txt
0 error(s), 3133 warning(s)

baseline: OK — .moai/spec-lint-baseline.json
  inventory: 3133 warning(s) total (advisory included), 0 non-advisory tracked across 0 recorded rule(s)
  recorded at: 4ac93f755 (2026-09-11)

$ <bin> spec lint --strict > strict.txt 2>&1; echo "strict exit=$?"
strict exit=0            (tail: 0 error(s), 3133 warning(s))

$ <bin> spec lint --json > lint.json 2>lint.stderr; echo "json exit=$?"   → json exit=0, stderr 0바이트
$ jq '[length, error, warning, non-advisory-warning, SpecsDirMissingSpecFile, MissingDependency, MovingRefUnpinned(양성 대조), NoSuchRuleCodeXYZ(음성 대조), DuplicateAcceptanceID]'
[3335, 0, 3133, 0, 0, 0, 116, 0, 0]
```

### 2.2 기제 — 스크래치 코퍼스 교차 프로세스 (저장소 밖 사본, SPEC 원본 무변경)

코퍼스: 이 SPEC 과 `SPEC-SPEC-LINT-BLIND-AXES-001` 디렉터 사본 2개. 기준선은 기제 경유로 산출.

```
캡처   --update-baseline --reason "audit scratch capture"      → exit 0, recorded: 0 non-advisory warning(s) across 0 rule(s)
                                                               sha256 3f5c53e6…a1cae8daa52
S1 무변경                                                       → exit 0, baseline: OK, 8 warning(s) total, 0 non-advisory
S2 spec.md 없는 SPEC 디렉터 주입                                → exit 1
     baseline: EXCEEDED — ../scratch-baseline.json
       SpecsDirMissingSpecFile: recorded 0 -> current 1 (+1)
S3 주입 제거                                                    → exit 0, baseline: OK
     기준선 sha256 3f5c53e6…a1cae8daa52 (캡처와 동일 — 적색 실행 포함 불변)
S4 error(가짜 의존 id) + 증가(디렉터 주입) 동시                  → exit 1
     ERROR     MissingDependency …
     baseline: ERROR-GATED — 1 error-severity finding(s); the baseline never absorbs an error
     EXCEEDED 출현 횟수: 0
     (복원 뒤 cmp 로 원본 spec.md 와 바이트 동일 확인, cmp exit 0)
S5 감소 (기록 SpecsDirMissingSpecFile:1, 현재 0)                → exit 0
     decreased: SpecsDirMissingSpecFile recorded 1 -> current 0 (-1)
     sha256 before 9881bc35…2b37c3b / after 9881bc35…2b37c3b (불변)
```

S4 는 M3.4 간선의 양성 대조도 겸한다 — 린터가 `dependencies:` 줄을 읽고 없는 id 에 `MissingDependency` error 를 낸다.

### 2.3 F1 — 거절 경로의 출력 (실제 바이너리)

```
$ <bin> spec lint --baseline .moai/spec-lint-baseline.json --strict > combo.txt 2>&1; echo $?   → 3, 출력 0바이트
$ <bin> spec lint --baseline /nonexistent/nope.json > missing.txt 2>&1; echo $?               → 3, 출력 0바이트
$ <bin> spec lint --baseline .moai/spec-lint-baseline.json --update-baseline --reason "   " > noreason.txt 2>&1; echo $?
                                                                                             → 3, 출력 0바이트
  이후 sha256 .moai/spec-lint-baseline.json = e9124a70…15455f6 (불변), git status 0줄
```

거절 자체(exit 3, 파일 불변)는 옳다. 문제는 **무엇이 거절됐는지 아무것도 출력하지 않는다**는 점이다. 원인은 코드에서 확인된다:

- `internal/cli/spec_lint.go:162-194` `validateBaselineFlags` — 모든 거절이 `&exitCodeError{...}` 만 반환하고 stderr 에 쓰지 않는다. `:233-236`(LoadBaseline 실패, exit 3)과 `:211-213`(WriteBaseline 실패, exit 2)도 같다.
- 같은 파일 `:338-344` 가 이 침묵을 이미 알고 있다("an ExitCoder returned from RunE … is rendered as nothing") — 그래서 `resolveLintTargets` 는 `argumentError(stderr, …)`(`:405-409`)로 stderr 에 먼저 쓴다. 새 코드는 그 파일 안의 관례를 따르지 않았다.
- 테스트가 결함을 가린다: `internal/cli/spec_lint_test.go:600` `slRunLint` 는 `combined + "\n" + err.Error()` 를 반환한다. 같은 파일 `:102-107`(`runSpecLint`)이 금지한 바로 그 방식이다 — "Appending it would let a test assert on text the user never sees". 그래서 `TestSpecLintBaseline_FlagContractRejections`(`:883-917`)의 "the message must name %q" 단언은 사용자가 보지 못하는 문자열로 초록이 된다(verification-completeness §2 의 잘못된 이유의 초록).

### 2.4 워크플로 (`.github/workflows/spec-lint.yml`)

```
$ python3 -c "yaml.safe_load(...)"
pr   ['.moai/specs/**', 'internal/spec/**', '.github/workflows/spec-lint.yml', '.moai/spec-lint-baseline.json']
push ['main', 'develop'] ['.moai/specs/**', 'internal/spec/**', '.github/workflows/spec-lint.yml', '.moai/spec-lint-baseline.json']
[('Run baseline-gated SPEC lint', "steps.policy.outputs.strict == 'true'", 'go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json'),
 ('Run error-only SPEC lint',     "steps.policy.outputs.strict == 'false'", 'go run ./cmd/moai spec lint')]

$ git diff 81c1d58f9 HEAD -- .github/workflows/spec-lint.yml    (merge-base 81c1d58f9 = 이 카드의 흡수 기준)
  + 두 paths 목록에 '.moai/spec-lint-baseline.json'
  + case: .moai/specs/*|internal/spec/*|.moai/spec-lint-baseline.json) spec_changed=true
  - run: go run ./cmd/moai spec lint --strict
  + run: go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json
  error-only 분기의 run 줄은 불변(주석만 갱신)
```

기준선만 바꾼 변경의 경로 전수:

| 이벤트 | 선택 분기 | 기준선 게이트 실행 |
|---|---|---|
| 일반 PR | 기본 `strict=true` (`:72`) | 예 |
| release/* PR | error-only (`:85`) — head 가 origin/develop 과 같아야 통과(`:81-84`), 그 develop 푸시에서 이미 게이트됨 | 상속 |
| main push | 기본 `strict=true` | 예 |
| develop push, 기준선 변경 포함 | case 적중 → `strict=true` (`:95`, `:99`) | 예 |
| develop push, before 없음 | `strict=true` (`:88-89`) | 예 |
| develop push, before 객체 없음 | `set -euo pipefail` 아래 `git diff` 실패 → 스텝 적색 | 실패 폐쇄 |

`--strict` 와 `--baseline` 을 함께 넘기는 경로는 없고, CLI 는 그 조합을 exit 3 으로 거절한다(2.3 첫 줄). `--no-renames` 로 이동도 옛·새 경로가 모두 잡힌다.

### 2.5 M3.4 — 간선과 게이트된 재기준

```
$ git merge-base --is-ancestor a4fbaeb82 HEAD; echo $?      → 0 (t518 조상)
$ git merge-base --is-ancestor 9fb52f746 a4fbaeb82; echo $? → 1 (대조 — 판별식 유효)
$ /usr/bin/grep -n '^status:\|^updated:\|^dependencies:' .moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001/spec.md
5:status: completed
7:updated: 2026-09-11
15:dependencies: [SPEC-SPEC-LINT-BLIND-AXES-001]
$ git log --format='%h %ad %s' --date=iso -- .moai/spec-lint-baseline.json
377d98c5a 2026-09-11 11:46:33 +0900 feat(…): M3.4 gated re-baseline after t518 absorb (card t525)
715199027 2026-09-08 14:27:03 +0900 docs(…): §E.3 run-phase stopping point + pre-t518-absorb baseline attribution (card t525)
9e1744469 2026-09-08 14:03:53 +0900 feat(…): M3 CI baseline wiring + initial baseline (card t525)
$ shasum -a 256 .moai/spec-lint-baseline.json → e9124a70…15455f6  (m3.4/baseline-sha256-after.txt 와 동일)
```

현재 파일: `tree_sha: 4ac93f755`, `updated_at: 2026-09-11`, 사유 문자열 기록, `rules: {}`. 레인 증거(`m3.4/edge-probe.md` 의 가짜 id 양성 대조, `rebaseline-coordinates.txt`, `exit-codes.txt`, `census-before.txt`)는 위 재관측과 모순이 없다.

리드 결정 확인: t518 밖 비-advisory 코드 `DuplicateAcceptanceID`(카드 t564) — HEAD 인구 `0`(2.1 jq 마지막 값). 재기준 이동폭 기여 0 이므로 SPEC 개정 없이 진행한다는 결정과 사실이 일치한다.

### 2.6 sync 산출물

```
$ git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)' ed4242845
manager-docs
$ git show --stat ed4242845   → sync-gate.txt(+3344), progress.md, spec.md(2줄), CHANGELOG.md(+1), docs-site en/ja/ko/zh 각 +3
$ diff .moai/reports/t525/sync/sync-gate.txt <HEAD 재실행 gate.txt>
1974d1973
< INFO      OwnershipTransitionUnmeasured  …/SPEC-SPECLINT-GATE-SIGNAL-001/spec.md … has no Authored-By-Agent trailer …
$ /usr/bin/grep -c 'SPECLINT-GATE-SIGNAL-001' sync-gate.txt → 8 ; '^ERROR' → 0 ; '^WARNING' → 3133
$ /usr/bin/grep -oE 'AC-SLGS-[0-9]{3}' acceptance.md | sort -u | wc -l → 12
```

- §E.4 의 수치(`exit=0`, `3133 warning(s)`, 8 히트, `^ERROR` 0)는 sync-gate.txt 와 일치한다. HEAD 재실행과의 유일한 차이는 INFO 1줄이 사라진 것인데, 트레일러를 단 sync 커밋이 착지해 소유권 전이가 측정 가능해졌기 때문이다 — 오히려 §E.4 설명(8 = 7 + INFO 1)을 뒷받침한다.
- docs-site: 4개 로케일 모두 같은 자리에 `--baseline` / `--update-baseline` / `--reason` 3행이 들어갔고 뜻이 코드와 맞는다.
- CHANGELOG `[Unreleased] → ### Fixed` 1건(`CHANGELOG.md:12`)을 코드와 대조했다. 판정 순서, `ERROR-GATED`/`EXCEEDED`, 감소 시 무변경, 사유 필수·빈 사유 거절·파일 불변, `--strict` 상호 거절, 워크플로 분기와 carve-out, 기준선 트리거 경로, CC2X 이전 경로 — 모두 코드·트리와 일치한다. 예외 1건은 F5.

### 2.7 품질 도구

```
$ gofmt -l internal/cli/spec_lint.go internal/cli/spec_lint_test.go internal/spec/lint_baseline.go internal/spec/lint_baseline_test.go
(무출력, exit 0)
$ golangci-lint run --timeout=5m ./internal/spec/... ./internal/cli/   → exit 0, "0 issues."
$ go test -count=1 -v -coverprofile=… ./internal/spec/ -run 'Baseline'  → exit 0
  --- PASS 11줄 / --- FAIL 0줄 (선택자가 비지 않았음을 확인)
$ go tool cover -func=… | grep lint_baseline.go   (이 선택자만으로 잰 함수별 커버리지)
  Delta 0.0% · Failed 100% · ErrorGated 100% · ComputeBaselineRules 87.5% · NewBaselineFromReport 0.0%
  MarshalBaseline 66.7% · WriteBaseline 66.7% · LoadBaseline 77.8% · CompareBaseline 96.8%
```

`Delta`·`NewBaselineFromReport` 는 `internal/cli` 테스트에서만 지나간다. 이번 감사는 `internal/cli` 테스트를 돌리지 않았다(슬롯 미배정).

### 2.8 보안 탐침 (카드 코드 diff `81c1d58f9..HEAD -- internal cmd .github`)

```
$ /usr/bin/grep -nEi '^\+.*(password|secret|api[_-]?key|token|sh -c|bash -c|\$\{\{ *github\.event\.(head|pull_request\.title|issue))' code.diff
(무출력, exit 1)
추가된 exec.Command: git -C <cwd> rev-parse --short HEAD (고정 argv, 셸 없음) + 테스트 전용 go build / 빌드 바이너리 실행
go.mod / go.sum 변경: 없음
```

사유 문자열은 `encoding/json` 으로 이스케이프돼 저장된다. 워크플로는 이벤트 값을 `env:` 로 넘기고 셸에서 따옴표로 쓰며, `permissions: contents: read` 다.

---

## 3. Baseline-attribution

- 위 모든 명령과 출력은 이 감사 실행에서 HEAD `ed4242845`(tree `6a375f6c…`)를 대상으로 얻었다. 판정 바이너리는 같은 HEAD 에서 빌드했고, 설치된 `moai` 는 쓰지 않았다(§2.2 도구 귀속).
- 스크래치 코퍼스 결과는 이 HEAD 바이너리로 저장소 **밖** 사본을 잰 것이다. git 이 없는 곳이라 `tree_sha` 는 `unknown` 으로 기록된다(설계대로).
- 인용한 레인 기록(M1 census `b6efc874f`, M2 커버리지 `b09b012ef`, M3.4 좌표 `4ac93f755`)은 그 트리의 값이며 이 실행의 측정으로 귀속하지 않는다.

## 4. AC 판정표

| AC | 판정 | 근거 |
|---|---|---|
| AC-SLGS-001 | PASS | `m1-demographics.md:10,24-28` — 명령 전문과 트리 `b6efc874f`(`git cat-file -t` → commit), `census-b6efc874f.json` 커밋됨 |
| AC-SLGS-002 | PASS | rc=1 절반: `verdict.md:32-38` (b6efc874f). 흡수 절반: progress M2-9/M2-10(동일 트리·동일 바이너리 `--strict` rc=1 ↔ `--baseline` rc=0). 이 HEAD 에서는 비-advisory 가 0 이라 재현할 수 없다 — 레인 기록을 인용 |
| AC-SLGS-003 | PASS | `verdict.md` §(a)/(b)/(c) 각각 명령과 원문 출력 |
| AC-SLGS-004 | PASS | 코드 `grep -rnE '4[,.]?3(44\|68\|78)' internal/spec internal/cli` → exit 1(대조 `baseline` 41히트). SPEC 문서 12히트는 임계·상수·기대값 용도가 없음. 참고 F6 |
| AC-SLGS-005 | PASS | 2.2 S2(exit 1, `+1` 규칙 이름) → S3(exit 0). 이 HEAD 에서 교차 프로세스 재관측. 레인 슬롯 기록 `slot/regression-guard.txt`: `--- PASS: TestSpecLintBaseline_InjectedWarningTurnsRedAndNamesTheRule`, exit=0 (레인 생산, 이 감사의 측정 아님) |
| AC-SLGS-006 | PASS | 2.1 — 실 코퍼스 exit 0 과 `inventory: 3133 warning(s) total` |
| AC-SLGS-007 | PASS | 2.2 S5 — 통과, `decreased:` 표시, 해시 불변 |
| AC-SLGS-008 | PASS | α: 2.3 빈 사유 exit 3·파일 불변. β: S5 불변. 정상 경로: M34-C1 과 현재 파일의 tree_sha·date·reason. (거절이 말없이 끝나는 문제는 F1) |
| AC-SLGS-009 | PASS | 2.2 S4 — error 가 동시 증가보다 앞서고 `ERROR-GATED` 로 이름 붙음, EXCEEDED 0 |
| AC-SLGS-010 전반부 | **PASS** | §5 판정 항목 1 |
| AC-SLGS-010 후반부 | PASS | 2.5 — 조상 판별(대조 포함), 간선, 게이트된 재기준 1회(377d98c5a, 사유 기록). F4 참조 |
| AC-SLGS-011 배선 절반 | PASS | 2.4 — kickoff 모양(`--baseline`), 파싱, 경로, 분기 전수 |
| AC-SLGS-011 로그 절반 | UNVERIFIED (리드 소관) | 녹색 CI run 로그는 develop 푸시 뒤에만 존재한다 |
| AC-SLGS-012 | PASS | 2.1 `SpecsDirMissingSpecFile` 0(양성 대조 116, 음성 0). "왜"는 progress M4 표. `lint.skip`·스텁 spec.md 없음 |

## 5. 판정 항목별 결론

### 5.1 AC-SLGS-010 전반부 — **PASS**

적용한 읽기: **(a)축 부채 상환 = "경고 감축 목적의 SPEC 문서 대량 수정"이며, 그 모집단은 advisory 로 표시된 서 있는 재고다.** 근거는 SPEC 본문의 문언과 요구 사이의 정합성이다.

1. 금지된 행위를 문언이 스스로 정의한다. REQ-SLGS-011: "(a)축 경고 부채 상환, 곧 대량 SPEC 문서 수정이다". AC-SLGS-010: "(a)축 부채 상환(경고 감축 목적의 SPEC 문서 대량 수정)". 요소는 셋이다 — SPEC 문서 수정, 대량, 경고 감축 목적.
2. §5.2 Out of Scope 는 (a) 모집단을 "CoverageIncomplete / ModalityMalformed 등 서 있는 경고 재고의 감축"과 "개별 SPEC 문서의 내용 수리"로 적는다. 반면 §5.1 In Scope 는 "SPEC-V3R4-CC2X-ADOPT-001/002 의 사유 규명과 발견 2건 종결(M4)"을 **이 SPEC 의 작업으로** 명시하고, REQ-SLGS-012 에는 t518 잠금(While 절)이 없다.
3. 효과 축 읽기("t518 착지 전 경고가 줄면 무엇이든 (a) 상환")를 택하면, t518 미착지 시점에 쓰인 계획(M3 → M4)이 스스로 금지한 일을 예정한 셈이 되어 SPEC 이 자기모순에 빠진다. 같은 SPEC 의 형제 요구는 함께 충족 가능하도록 읽는 것이 옳으므로 이 읽기는 기각한다.
4. 사실관계. M4(`9fb52f746`, 2026-09-08 09:03)는 t518 착지(`3ac58b5a1`, 12:37) **전**이다 — 쟁점은 살아 있다. 그러나 이 브랜치의 비병합 카드 커밋 가운데 다른 SPEC 문서를 건드린 것은 9fb52f746 하나이고, 그 범위는 디렉터 2개 이전과 링크 1줄 수리 2건이다(`git log --no-merges --name-only 81c1d58f9..HEAD -- .moai/specs`). "대량"이 성립하지 않는다. 목적은 REQ-SLGS-012 가 따로 명령한 "왜"와 함께 닫기이며, 발견을 억누르는 수단(`lint.skip`, 스텁)은 쓰지 않았다.
5. `--strict` rc 1→0 이라는 효과는 숨겨진 목적이 아니었다. M4 이전 트리(`b6efc874f`)의 M1 verdict(`verdict.md:46-57`)가 이 반전을 미리 적어 두었다.

잔여 위험: 이 판정은 문언 해석에 기댄다. 운영자가 효과 축 읽기를 의도했다면 판정이 아니라 SPEC 을 개정해야 한다(현재 문언으로는 성립하지 않는다).

### 5.2 AC-SLGS-011 — 지금 관측 가능한 것

배선 절반은 PASS(2.4). 같은 명령을 로컬에서 돌리면 녹색이고 `inventory:` 줄이 나온다(2.1). 로그 절반은 푸시 뒤 CI 로그에만 있으므로 UNVERIFIED 로 둔다. 리드가 읽을 때의 참고:

- 이 카드를 싣는 develop 푸시는 `.moai/specs/` 를 바꾸므로 기준선 분기가 선택된다 → 로그에서 `SPEC lint policy: strict=true (develop push changes SPEC inputs or the spec linter)` 다음에 `baseline: OK` 와 `inventory:` 줄을 읽는다.
- 이후 SPEC 입력을 건드리지 않는 develop 푸시는 error-only 분기라 `inventory:` 줄이 없고 `N error(s), M warning(s)` 요약 줄만 남는다. AC 의 "경고 총수 줄"로는 충족되지만 기준선 상태 줄은 없다.

### 5.3 워크플로 해법(리드 결정 A) — 올바름

2.4 표대로 기준선만 바꾼 변경이 기준선 게이트를 건너뛰는 경로는 없다. `--strict`/`--baseline` 동시 전달 경로도 없다. 남은 것은 선택적 개선뿐이다(F2, F3).

### 5.4 M3.4 — 확인

간선, 양성 대조, 조상 판별, 1회 재기준, 재기준 전후 게이트 exit 0, 파일 해시 모두 재관측과 일치한다. `DuplicateAcceptanceID` 인구 0 도 재확인했다.

### 5.5 sync 산출물 — 대체로 정확, 1건 부정확(F5)

---

## 6. 차원 점수 (default 프로필, flat)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 75/100 | UNVERIFIED | AC 11개 PASS(§4). AC-SLGS-011 로그 절반은 푸시 뒤에만 관측 가능. 경계 사례 "명확한 인자 오류"가 사용자 표면에서 미충족(F1): `--baseline /nonexistent/nope.json` → exit 3, 출력 `bytes= 0` |
| Security (25%) | 95/100 | PASS | 2.8 탐침 `sec-grep exit=1`(무히트). exec 는 고정 argv, 워크플로는 env 경유·`contents: read`, 의존성 변경 없음 |
| Craft (20%) | 70/100 | UNVERIFIED | `golangci-lint … 0 issues.`, gofmt 무출력. spec 기준선 테스트 `--- PASS 11 / --- FAIL 0`. 그러나 F1 테스트가 사용자 표면을 단언하지 않는다. `internal/cli` 커버리지는 레인 수치 81.4%(`b09b012ef`, 85% 미만) — 이번에 재측정하지 않음 |
| Consistency (15%) | 75/100 | PASS | 같은 파일의 `argumentError` 관례(`spec_lint.go:338-344,405-409`)와 테스트 헬퍼 관례(`spec_lint_test.go:102-107`)에서 벗어남 — 국소적. 형식·린트 청결 |

가중 조화평균 ≈ 78.8. must-pass 방화벽: Security PASS. Functionality 는 AC 로는 통과지만 AC-011 절반이 구조상 UNVERIFIED 다. 종합 FAIL 은 점수가 아니라 차단 결함 F1 때문이다.

## 7. Findings (structured defect-list)

- **F1** [Medium] [**blocking**] `internal/cli/spec_lint.go:162-194` (+ `:233-236`, `:211-213`), `internal/cli/spec_lint_test.go:600` — 기준선 플래그 거절과 기준선 로드·쓰기 실패가 모두 출력 없이 exit 3/2 로 끝난다(실측 0바이트 3종: `--strict` 조합, 없는 파일, 공백 사유). acceptance.md 경계 사례는 "명확한 인자 오류(exit 3 계열)"를 요구했고 M2 는 "exit 3, 경로를 이름 붙임"으로 정해 기록했다. 그런데 exit 3 을 내는 원인이 여섯 가지(+ 기존 `--json --sarif`)라 코드만으로는 구별되지 않고, CI 에서 기준선 파일이 없어지거나 깨지면 이유 없는 적색이 된다. 가드 테스트 `TestSpecLintBaseline_FlagContractRejections` 는 `err.Error()` 를 덧붙이는 헬퍼 덕분에 초록이다. 확신도 high. — **Required fix**: (1) `validateBaselineFlags` 가 `cmd.ErrOrStderr()` 를 받아 각 거절을 기존 `argumentError(stderr, …)` 로 반환하게 하고, `runBaselineGate` 의 LoadBaseline(exit 3)·WriteBaseline(exit 2) 오류도 같은 방식으로 stderr 에 쓴다. (2) `slRunLint` 가 `err.Error()` 를 덧붙이지 않게 해 `FlagContractRejections` 가 실제로 쓰인 출력에서 원인 단어를 단언하게 한다(수정 전 이 테스트가 붉어지는 것을 먼저 관측). (3) 빌드 바이너리로 `--baseline <없는 경로>` 1건을 돌려 stderr 에 경로가 찍히는지 교차 프로세스로 확인한다.
- **F2** [Low] [optional] `.github/workflows/spec-lint.yml:72,85,89,99,102,111,131` — 정책 출력 이름이 `strict` 로 남아 있고, `:89` 사유 문자열 "fail closed with strict lint" 는 이제 기준선 게이트를 가리킨다. 로그 독자가 오해할 수 있다. 확신도 high. — Required fix(선택): 출력명을 `gate=baseline|error-only` 로 바꾸거나 사유 문자열의 "strict lint"를 "baseline-gated lint"로 고친다.
- **F3** [Low] [optional] `.github/workflows/spec-lint.yml:5-24,93-97` — 게이트 결정 코드(`internal/cli/spec_lint.go`)와 `cmd/moai/**` 는 트리거 경로 밖이다. 그 코드만 바꾼 푸시는 SPEC Lint 를 돌리지 않고 단위 테스트(메인 CI)에만 기댄다. 워크플로 파일만 바꾼 develop 푸시가 error-only 분기를 고르는 것은 흡수 전부터 있던 동작이다. 확신도 medium. — Required fix(선택): paths 와 case 목록에 `internal/cli/spec_lint.go` 를 추가할지 리드가 정한다.
- **F4** [Low] [optional] `progress.md:548`, 기준선 이력 — 715199027 이 기준선을 `tree_sha: 9e1744469` 로 다시 산출했는데, 같은 커밋에 들어간 §E.3 문단은 여전히 `tree_sha: 9fb52f746` 라고 적는다(현재는 SUPERSEDED 표지 아래). 또 기준선은 세 번 쓰였고(9e1744469 최초, 715199027 귀속 재산출 — 건수 불변·사유 기록, 377d98c5a 게이트된 재기준), "재기준 1회"는 715199027 을 재기준으로 치지 않는 읽기에서 성립한다. 확신도 high. — Required fix(선택): progress 에 715199027 재산출을 한 줄로 명시하고 548행의 SHA 를 정정한다(manager-docs 소관 절에 한함).
- **F5** [Low] [optional] `CHANGELOG.md:12` — "one of them (the wiring-completeness half of `AC-SLGS-011`, reading the green CI run's inventory line) … is deferred"라고 쓰는데, 미뤄진 것은 배선 절반이 아니라 CI 로그 절반이다. 배선 절반은 이미 관측됐다. 확신도 high. — Required fix: "the CI-log half of `AC-SLGS-011`"로 고친다.
- **F6** [Info] [optional] `spec.md:74`, `plan.md:14,133`, `acceptance.md:45` — AC-SLGS-004 히트 일부는 같은 불릿 안에 출처(SHA/run id)가 없는 서술·메타 언급이다. AC 실패 모양이 겨냥하는 임계·상수·기대값은 아니라서 PASS 이지만, M1-5 의 "전부 출처 명시 인용"은 약간 과장이다. 확신도 medium. — Required fix: 없음(기록 정정 선택).
- **F7** [Low] [optional] `internal/cli/spec_lint.go:209-230` — error 가 서 있을 때 `--update-baseline` 은 파일을 쓴 뒤 exit 1 한다. 그 사실은 파일 안에 남지 않는다(레인이 이미 잔여 위험으로 기록). `resolveTreeSHA` 는 워킹트리가 더러워도 HEAD 만 기록한다. 확신도 high. — Required fix(선택): 그대로 두되 재기준 절차 문서에 "깨끗한 트리에서 실행" 한 줄을 둔다.
- **F8** [Info] [optional] `internal/spec/lint_baseline.go:60,109` — `Delta`·`NewBaselineFromReport` 는 spec 패키지 테스트만으로는 0% 이고 `internal/cli` 테스트로만 지나간다. `internal/cli` 패키지 커버리지 81.4%(레인 수치)는 85% 미만이다. 확신도 medium. — Required fix(선택): `internal/spec` 에 두 함수의 단위 테스트를 둔다.

## 8. Gaps (관측하지 않은 것)

- `internal/cli` 테스트는 한 번도 돌리지 않았다(슬롯 미배정). 회귀 가드의 초록은 레인이 감사 도중 남긴 `slot/regression-guard.txt`(PASS 1줄, exit=0)를 인용할 뿐이며, 이 감사의 측정이 아니다.
- `internal/cli` 와 `internal/spec` 의 패키지 커버리지를 재측정하지 않았다(M2 `b09b012ef` 수치 인용).
- 교차 모델 감사(`audit_multi`/codex/glm)는 부르지 않았다. `uncommittedChanges` 는 비어 있고, `baseBranch` 는 원격 기본 헤드(main)와 비교해 develop 전체를 싣게 되어 카드 diff(`81c1d58f9..HEAD`)를 겨눌 수 없다. 틀린 트리에 대한 판정을 만들지 않으려고 생략했다.
- 녹색 CI run 로그(AC-SLGS-011 절반)와 windows/darwin 실행.
- 병합 트리. 참고 조기 신호만 있다: develop tip `ee99507fb` 를 `git archive` 로 풀어 그 트리 바이너리로 `--json` 을 돌리면 비-advisory `SpecsDirMissingSpecFile 2`(CC2X 두 디렉터 — 이 카드가 없애는 것)뿐이었다. `.git` 이 없어 git 의존 규칙은 눈이 먼 측정이므로 병합 근거로 쓰면 안 된다.
- AC-SLGS-002 의 흡수 절반은 이 HEAD 에서 재현할 수 없다(비-advisory 0). 레인 기록을 인용했다.

## 9. Residual-risk

- F1 이 남아 있는 동안에는 기준선 파일 삭제·오염·잘못된 플래그 조합이 CI 에서 이유 없는 exit 3 로 나타난다. 게이트는 닫힌 쪽으로 실패하므로 거짓 초록은 없지만, 진단 비용이 운영자에게 떠넘겨진다.
- `"rules": {}` 는 가장 엄격한 상태다. 병합 창의 재흡수에서 새 비-advisory 규칙이 하나라도 발화하면 적색이 되고, 그때 습관처럼 재기준하면 기준선은 장식이 된다. 사유 필수가 유일한 제동 장치다.
- 5.1 판정은 문언 해석에 기댄다.
- 감사 도중 다른 행위자(레인)가 이 워크트리에 미추적 파일 `.moai/reports/t525/slot/`(19:24)을 썼다. HEAD 는 `ed4242845` 그대로였고 커밋은 없었다. 감사 창에 쓰기가 하나 더 있었다는 절차 사실로 리드에게 알린다(`agent-common-protocol.md` § Background Agent Execution — 감사 중인 워크트리의 단일 작성자).
