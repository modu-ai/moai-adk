# SPEC Review Report: SPEC-SEC-SCAN-SURFACE-001

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS**
Overall Score: **0.925** (Tier M PASS threshold = 0.80)
Score trend: iteration 1 **0.65** → iteration 2 **0.925**. No regression ⇒ no STOP signal.

측정 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t217`, 브랜치 `security-scan-surface`, HEAD `2f8f6a6c1dbc16399cd5a5620463923b82135e22` (`git rev-parse HEAD` 실측, working tree clean).
Reasoning context ignored per M1 Context Isolation. 입력은 `spec.md` / `plan.md` / `acceptance.md`(Tier M 집합) + 근거 `investigation.md` + 커밋된 측정 스크립트 2종. 룰셋은 워크트리에 없어 primary 체크아웃과 템플릿 사본에서 읽었다.
Iteration 1 기록은 `plan-audit.md` 에 그대로 보존돼 있다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o "REQ-SSS-[0-9]\{3\}" spec.md | sort -u` → 16건, `REQ-SSS-001`…`REQ-SSS-016` 연속, 결번·중복 없음.
- **[PASS] MP-2 GEARS format compliance** — **요구 레이어(`REQ-XXX` in `spec.md`) 기준**. 16개 전부 다섯 패턴 안에 든다: Ubiquitous(001·003·005·008·010·011·013·014·015·016), Where/When(002·004·007), When(006·012), Where(009). `acceptance.md` 의 Given-When-Then 은 검증 레이어의 정상 형식이므로 여기서 감점하지 않았다(Group 4 소관). 통합으로 생긴 복합 문장 문제는 E5 로 별도 기록 — MP-2 실패가 아니다(이유는 E5 본문).
- **[PASS] MP-3 YAML frontmatter validity** — iteration 1 의 D1 이 해소됐다. 도메인 전용 도구로 재확인:
  ```
  $ ~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md
  ✓ No findings — all SPEC documents are valid
  ```
  `spec.md:13` 이 `tags: "security, hook, pretooluse, posttooluse, ast-grep, performance"` — 캐논 스키마의 콤마 구분 문자열.
- **[PASS] MP-4 Section 22 language neutrality** — REQ-SSS-005(spec.md:121)가 커버 언어 집합을 **설정에서 파생**하도록 강제하므로 5번째 언어 룰이 추가되면 스캔이 자동 재개된다. 언어 목록의 코드 하드코딩을 금지하는 조항이 같은 REQ 에 있다. 특정 언어를 primary 로 승격시키는 서술 없음.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE "SPEC-([A-Z][A-Z0-9]+-)+[0-9]+" spec.md | sort -u` → 자기 자신 1건. retired/superseded 참조 없음.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rc "NEEDS CLARIFICATION" plan.md` → `0`. `research.md` 는 Tier M 이라 부재가 정상.

일곱 개 전부 통과 — M5 Must-Pass Firewall 미발동.

---

## 렌즈 1 — D2 신규 추출 규칙 2행 정면 공격 (최우선)

결론부터: **반례를 찾지 못했다.** 아래가 찾으려 한 것과, 각각 무엇으로 배제했는지다.

### 대상 모집단을 먼저 고정

`regex:` 를 쓰는 룰은 배포 룰셋 전체에 5개뿐이고, 그중 **error 심각도는 4개**(전부 동일한 `sec-hardcoded-credential` 변종)다:

```
$ grep -rn "regex:" internal/template/templates/.moai/config/astgrep-rules/
go/hardcoding.yml:15        regex: '^"https://api\.'                    # severity: warning → 제외
security/credentials.yml:23 regex: "^\"(sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})"   # go
security/credentials.yml:35 regex: "^[\"'](sk-|…)"                      # python
security/credentials.yml:47 regex: "^[\"'](sk-|…)"                      # javascript
security/credentials.yml:59 regex: "^[\"'](sk-|…)"                      # typescript
go/error-handling.yml:28    regex: '^err(e?|s)$'                        # severity: warning → 제외
```

### 교대(alternation) 행 — 건전함, 배포 룰셋 전수 확인

다섯 분기 `sk-` / `AKIA[0-9A-Z]{16}` / `ghp_[0-9A-Za-z]{36}` / `xox[baprs]-` / `AIza[0-9A-Za-z_-]{35}` 는 **각각 필수 리터럴 접두**를 갖는다(`sk-`, `AKIA`, `ghp_`, `xox`, `AIza`). 매치는 어느 한 분기를 요구하므로 다섯 중 하나도 없으면 매치가 불가능하다 — `any:` 와 동일한 논리이고 §C.2 가 그렇게 적었다. 더불어 정규식은 **노드 텍스트**에, 프리필터는 **페이로드 전체**에 적용되므로 후자가 전자의 상위집합이다. 거짓 skip 경로 없음.

내가 뒤진 무효화 요인과 결과:

| 무효화 요인 | 배포 룰셋에서 | 근거 |
|---|---|---|
| 대소문자 무시 플래그 `(?i)` | **없음** | `grep -rn "(?i)" …astgrep-rules/` → rc=1, 0건 |
| 선택 그룹이 앞에 오는 분기(`(sk-)?AKIA`) | 없음 | 위 5개 정규식 전수 판독 |
| 문자클래스로 시작하는 분기 | 없음(그룹 **밖**의 `["']` 는 접두로 채택되지 않음) | 동상 |
| 앵커 `^` 를 리터럴로 오채택 | 해당 없음 | 동상 |

### `kind:` + `regex:` 행 — 건전함, **실측으로** 확인

가장 위험한 가설은 "정규식이 원본 소스 텍스트가 아니라 **디코드된 문자열 값**에 걸린다"였다. 그렇다면 `"\x73k-…"`(디코드하면 `sk-`)가 오늘 deny 되면서 프리필터에는 `sk-` 리터럴이 없어 skip — 정확히 요청받은 형태의 치명 반례가 된다. 실측:

```
$ cat /tmp/t217r2/esc.go
package main

const a = "\x73k-abcdef1234567890"

$ sg scan -c …/astgrep-rules/sgconfig.yml --json /tmp/t217r2/esc.go
findings: []                                   ← 오늘도 deny 되지 않는다

$ sg scan -c …/astgrep-rules/sgconfig.yml --json /tmp/t217r2/raw.go   # const a = "sk-abcdef1234567890"
"ruleId": "sec-hardcoded-credential"           ← 대조군: 원본 형태는 잡힌다
```

ast-grep 의 `regex:` 는 **원본 소스 텍스트**에 걸린다. 룰과 프리필터가 같은 문자열을 본다 ⇒ **오늘 deny 되는 구성이 프리필터를 빠져나가는 경로 없음.**

`kind` 가 정말 좁히기만 하는지도 실측했다 — 주석 안의 `sk-`:

```
$ printf 'package main\n\n// sk-abcdef1234567890 in a comment, not a string literal\nvar x = 1\n' > kindtest.go
$ sg scan -c … --json kindtest.go
findings: []
```

룰은 안 걸리는데 프리필터는 토큰이 있으니 escalate 한다 — **보수적 방향**(불필요한 escalate)이지 누락이 아니다. `rule:` 안의 동급 원자 규칙은 AND 결합이므로 `kind` 는 매치 집합을 좁히기만 하고, 정규식의 필수 리터럴은 그대로 필수로 남는다. §C.2 가 이를 `all:`-형 결합이라 적은 것은 정확하다.

**렌즈 1 판정: 두 행 모두 배포 룰셋에 대해 건전하다. 반례 0건.** 유일한 잔여는 규칙 표가 커버하지 않는 정규식 문법 클래스(E2) — 오늘 룰셋에는 존재하지 않으므로 잠복이다.

---

## 렌즈 2 — 측정 스크립트 감사

### 재현

```
$ python3 .moai/reports/t217/skiprate.py .
go files=2438 wouldSKIP=30 rate=1.2%
js files=81 wouldSKIP=78 rate=96.3%
py files=14 wouldSKIP=13 rate=92.9%
```
`2f8f6a6c1` 에서 §A.1 / §C.3 의 숫자와 정확히 일치한다.

### §C.2 표를 그대로 구현하는가 — **아니다(python 한정)**. → E1

`sec-command-injection-shell`(python, error)은 네 분기 `any:` 다: `subprocess.call` / `subprocess.run` / `subprocess.Popen` / `os.system`. §C.2 의 `any:` 행은 모든 분기가 토큰을 내면 **합집합**을 채택하라고 한다. 스크립트의 python 토큰은 `['subprocess.call','os.system','sk-','AKIA','ghp_','xox','AIza']` — `subprocess.run` 과 `subprocess.Popen` 이 **빠졌다**(skiprate.py:31). 즉 스크립트는 명세보다 **더 관대한** 근사를 구현한다. 실측 영향:

```
py files=14 script_skip=13 correct_skip=12
divergent: ['./.moai/reports/t82/measure-surface.py']
```
올바른 토큰 집합에서 python skip rate 는 **85.7%** 이고, 보고된 **92.9%** 는 7.1%p 과대다. 오차 방향이 절감을 부풀리는 쪽이다.

go 는 반대로 **보수적**이다: 8개 error 룰 중 1개의 토큰(`,` `=`)만 쓰므로 토큰을 더하면 skip 은 내려갈 뿐 — 1.2% 는 상한이다. 스크립트 docstring 이 그렇게 적었고 맞다. js/ts 는 error 룰이 2개뿐이고 토큰이 전부 들어 있어 정확하다(ts 를 'js' 버킷에 합친 것은 두 언어의 토큰 집합이 동일해 결과에 영향 없음).

### 파일 수인가 Write 수인가 — 정직하게 적혀 있다

§C.3 "Gaps in this measurement"(spec.md:279-283)가 (a) Writes 가 아니라 저장소에 존재하는 **파일**을 센다는 것, (b) js 81 / py 14 모집단이 작고 도구 스크립트 쪽으로 치우쳤다는 것, (c) `child_process.exec` / `os.system` 을 실제로 쓰는 프로젝트는 훨씬 덜 skip 한다는 것을 먼저 밝힌다. 방향성 측정이지 보증이 아니라고 명시했다.

모집단 편향은 한 가지 더 확인했다 — `skipdirs` 가 걷어낸 js/ts 파일 수:
```
$ find … \( -path '*/public/*' -o -path '*/resources/*' -o -path '*/dist/*' -o -path '*/node_modules/*' -o -path '*/vendor/*' \) -a \( -name '*.js' -o … \) | wc -l
0
```
빌드 산출물 제외가 이 저장소에서는 아무 표본도 걷어내지 않았다. 81 은 필터링으로 줄어든 수가 아니다.

### 16-언어 주장의 근거로 충분한가 — 부분적으로만. → E6

§C.3 의 유지 논거 중 "would deny a **93-96% saving to every** JavaScript, TypeScript, and Python project"(spec.md:274-275)는 이 저장소의 81+14 파일에서 모든 배포 프로젝트로 일반화한다. 같은 절의 Gaps 문단이 그 전이 가능성을 스스로 부정하므로 두 문단이 긴장한다. 다만 유지 결정 자체는 그 전이를 **필요로 하지 않는다** — "기전이 데이터 구동이라 언어별 룰셋이 자라면 비율이 자동 조정된다"는 논거가 독립적으로 성립한다. 수사적 과잉이지 논거 붕괴는 아니다.

---

## 렌즈 3 — 16개 AC 전부의 미구현 트리 측정값 재실측

| AC | PASS 값 | SPEC 기재 "measured today" | 내 실측 (`2f8f6a6c1`) | 반전? |
|---|---|---|---|---|
| 001 | 동일 쌍 재현 | (i) M0 에서 생성, (ii) M2 전 미컴파일 | 코퍼스·테스트 부재 확인 | 구조적 RED |
| 002 | ScanFile 0 | 1 | **1** — `pre_tool.go:638` 이 지원 확장자에서 `ScanFile` 1회 호출 | ✅ |
| 003 | temp 파일 0 | 1 | **1** — `pre_tool.go:623` `os.CreateTemp` 가 스캔 **전**에 실행 | ✅ |
| 004 | scanner측 0 / caller측 1 | scanner측 1 / caller측 0 | **scanner 1 / caller 0** — `grep -rn "FindRulesConfig" internal --include="*.go" \| grep -v _test` 가 pre-write 경로 유일 호출로 `security/scanner.go:84`(ScanFile 내부)만 보고 | ✅ 두 카운터가 반전 |
| 005 | 0 (11케이스) + .go 대조 1 | 11 | **11** — `supportedLanguages` 15개(ast_grep.go:358-374) − 커버 4개 = 11 | ✅ |
| 006 | 0 / 1 분기 | 두 팔 모두 1 | **1 / 1** — 언어 필터가 없어 두 팔이 같음 | ✅ 0/1 분기는 하드코딩 목록으로 불가 |
| 007 | 3팔 모두 1 | 3팔 모두 1 (**행위보존형**으로 라벨됨) | **1 / 1 / 1** | ⚠️ 비반전 — 라벨 있음 |
| 008 | — | 미컴파일 | 함수 부재 확인 | 구조적 RED |
| 009 | 토큰 정확히 5개 | 미컴파일 | 함수 부재 확인 | 구조적 RED |
| 010 | underivable + escalate 1 | 미컴파일 | 함수 부재 확인 | 구조적 RED |
| 011 | ScanFile 0 (.js) | 1 | **1** | ✅ |
| 012 | 두 필드 생존 + 도달 | 미컴파일(가디언 핸들러 부재) | `deps.go:220-264` 에 가디언 PostToolUse 핸들러 등록 없음 확인 | 구조적 RED |
| 013 | 바이트 동일 | 미컴파일 | 동상 | 구조적 RED |
| 014 | block 없음 | 미컴파일 | 동상 | 구조적 RED |
| 015 | jq `0` | jq `1` | **`1`** (재실측) | ✅ |
| 015 (3행) | `exit=0` | `exit=0` (**행위보존형** 라벨) | `printf '{}' \| ~/go/bin/moai hook security-scan` → **exit=0** | ⚠️ 비반전 — 라벨 있음. 단 인용 경로는 E4 |
| 016 (drift) | 0줄 | 가드 있으면 0 / 없으면 31 | **guarded=0, unguarded=31** (`driftcheck.sh` 로 양쪽 재현) | ✅ |
| 016 (leak) | pass | pass | **미관측(Gap)** — 아래 Gaps 참조 | — |
| 016 (PR) | 선언 확인 | M4 전 PR 없음 | PR 부재 확인 | 구조적 |

**iteration 1 의 D3/D4 형태(= PASS 값이 오늘 값과 같아 공허하게 통과)가 재도입된 곳은 없다.** 비반전인 두 항목(AC-SSS-007, AC-SSS-015 3행)은 SPEC 이 스스로 "behaviour-preservation criteria … labelled as such"(acceptance.md §D 마지막 항)로 분류해 두었고, 실제로 그 성격이 맞다 — M1 의 skip 로직이 깨뜨릴 수 있는 것을 지키는 기준이지 변화를 관측하는 기준이 아니다. 이 구분은 iteration 1 에서 요구한 정직성의 정확한 형태다.

---

## 렌즈 4 — D6 통합(20→16)에서 내용이 사라졌는가

`git show c5d08ce0e:…/spec.md` 의 20개와 현재 16개를 1:1 대조했다.

| iter1 | → iter2 | 처리 |
|---|---|---|
| 001 deny 보존 + 002 신규 deny 금지 | **001** | 양방향을 한 문장에 병합 |
| 003 | 002 | 그대로 |
| 004 "exactly once" | **003** | **강화** — "in the caller … scanner shall perform no second resolution" 추가 |
| 005 | 004 | 그대로 |
| 006 | 005 | 그대로 |
| 007 | **006** | **확장** — "or yields an empty covered-language set"(D8 채택) |
| 008 | 007 | 그대로 |
| 009 severity 한정 + 011 GuardianPatterns 금지 + 012 순수함수 | **008** | 3→1 병합. 세 의무 모두 문면에 존재 확인 |
| 010 | 009 | 그대로 |
| 013 | 010 | 그대로 |
| 015 advisory 유지 | 011 | 그대로 |
| 014 두 캐리어 필드 | **012** | **확장** — "No preceding handler's decision shall prevent the guardian scan from being reached"(D7 채택) |
| 016 | 013 | 그대로 |
| 017 | 014 | 그대로 |
| 018 미러/짝 + 019 중립성 | **015** | 2→1 병합. 세 의무(미러·짝·금지 콘텐츠 6종) 모두 존재 확인 |
| 020 | 016 | 그대로 |

산술: 20 − 1(001+002) − 2(009+011+012) − 1(018+019) = **16**. ✅
**사라진 내용 없음.** 유일한 미세 손실은 구 011 의 부수절 "so that it is testable without a filesystem and without `sg`" 가 신 008 에서 빠진 것인데, 이는 의무가 아니라 근거절이고 "pure function of the resolved rule set" 이 같은 것을 함의한다 — 내용 삭제로 보지 않는다. 통합의 부작용은 E5(복합 요구)로 별도 기록.

---

## 렌즈 5 — D5 의 M0 seam 이 실제로 명세되었는가

**되었다. 이름만 붙인 게 아니다.** `plan.md` M0 step 1:

> Introduce a narrow interface in `internal/hook` describing only what `scanWriteContent` uses of the scanner (`IsAvailable`, `ScanFile`, `ShouldAlert`, `GetReport`) and change `preToolHandler.scanner` from `*security.SecurityScanner` to that interface. `NewPreToolHandlerWithScanner` keeps its concrete parameter type so no caller changes.

세 가지가 갖춰져 있다: (a) **순서** — M0 이 M1/M2/M3 앞에 오고 의존 AC 전부가 M1 이후에 닫히므로 seam 이 먼저 선다; (b) **범위** — 4개 메서드로 좁힌 인터페이스, 동작 변화 없는 타입 축소; (c) **호환** — 생성자 시그니처를 유지해 호출부 변경 0. 내가 iteration 1 에서 제안했던 "핸들러 레벨 스텁" 이 왜 그대로는 불가능했는지도 §A.1(spec.md:63)에 실측과 함께 기록됐다(`pre_tool.go:325` 가 구조체 포인터).

`acceptance.md` 의 "The instrument" 문단은 한발 더 나가 **잘못된 계측기를 명시적으로 폐기**한다: "`astGrepScanner.scanFunc` is **not** the instrument: it is consulted only inside `ScanMultiple` (`ast_grep.go:199-200`) and the single-file `Scan` execs `sg` directly at `:137`." 내가 실측한 것과 일치한다(ast_grep.go:31 주석 "for ScanMultiple", :195-200 선택 로직, :136 직접 exec). 그리고 "`ScanFile` is the only route from this path to an `sg` spawn, so a `ScanFile` count of 0 proves a spawn count of 0" — 이 함의도 코드상 참이다(`ScanFile` → `astGrep.Scan` 이 유일 경로).

---

## 렌즈 6 — handle-pre-tool.sh 드리프트: 확인 + 처리 타당성

**확인된다.** `diff .claude/hooks/moai/handle-pre-tool.sh internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl` → 세 덩어리, **전부 주석**:

- 배포본에만: `# --- Bash Risk-Amplifier warn signal (SPEC-V3R6-BASH-RISK-GOVERNANCE-001) ---` — **SPEC ID**. 템플릿 중립성 CI 가드가 금지하는 C-class 콘텐츠다.
- 템플릿에만: stdin 취급에 관한 7줄 문서 블록 + `§Bash Risk-Amplifier Doctrine (4)` 의 항목번호 제거.

즉 이 축의 바이트 동일성은 **미수복 상태가 아니라 구조적으로 달성 불가**다 — 같게 만들려면 템플릿에 SPEC ID 를 넣어 중립성 가드를 위반하거나, 배포본에서 문서 블록을 지워야 한다. 따라서 §D 의 "byte-equality across the `.sh` / `.sh.tmpl` axis is therefore **not** a valid repository-wide invariant" 는 **편의가 아니라 사실 진술**이고, 범위를 자기 diff 로 좁힌 것(REQ-SSS-015)은 옳은 판단이다. 이 SPEC 은 훅 래퍼를 하나도 건드리지 않는다(`git diff --stat origin/main...HEAD` → 변경 8파일 전부 `.moai/` 산출물).

다만 축이 어긋난 지점이 하나 있다 → E3.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 1.0 근접 | §C.2 가 7행 추출표 + 지배적 형태 워크드 예제(spec.md:227-241)로 대체돼 두 갈래 독법이 제거됨. 잔여는 표가 이름 붙이지 않은 정규식 문법 클래스(E2) |
| Completeness | 0.95 | 1.0 | 필수 섹션 전부, Out of Scope H3 6개 각각 구체 bullet, REQ 16/AC 16 = Tier M 상한 정확히 충족, 측정 스크립트 2종 커밋 |
| Testability | 0.85 | 0.75~1.0 | 계측기 정정 + 오계측기 명시 폐기, 16개 AC 전부 미구현 측정값 기재(비반전 2건은 라벨). 감점: E3(기계화된 축 불일치), E4(인용 명령 미실행 가능) |
| Traceability | 1.00 | 1.0 | §E 표가 16개 REQ 전부를 닫는 AC 에 사상; 인용된 AC-SSS-001..016 이 `acceptance.md` 에 전부 실재(16 unique 실측) |

산술 평균 **0.925**. Tier M 임계 0.80 초과.

---

## Defects Found (structured defect-list)

**E1** — `.moai/reports/t217/skiprate.py`:31 (+ `spec.md`:64, `spec.md`:263) — **측정 스크립트가 자기가 구현한다고 밝힌 §C.2 규칙과 어긋난다.** python 의 error 룰 `sec-command-injection-shell` 은 4분기 `any:` 인데 토큰 집합에서 `subprocess.run` 과 `subprocess.Popen` 이 누락됐다. §C.2 의 `any:` 행은 전 분기 합집합을 요구한다. 실측 영향: 올바른 토큰 집합에서 `py files=14 script_skip=13 correct_skip=12` — 실제 skip rate 는 **85.7%** 이며 보고된 **92.9%** 는 7.1%p 과대, 오차 방향이 절감을 부풀린다. 발산 파일 1건 특정: `./.moai/reports/t82/measure-surface.py`. 이 수치는 §A.1 표와 §C.3 두 곳에 인용돼 A1 유지 결정을 떠받치므로, "measured, not asserted" 라는 §C.3 의 자기 규정과 어긋난 상태다(VCI §2 귀속 위반). **단, 결론은 뒤집히지 않는다**(go 1.2% vs py 85.7% 의 비대칭은 그대로). — Severity: **major** — Class: **blocking** — Required fix: `skiprate.py` 의 `'py'` 토큰 목록에 `'subprocess.run'`, `'subprocess.Popen'` 두 개를 추가하고 재실행한 뒤, `spec.md:64` 와 `spec.md:263` 의 python 수치를 실측값으로 교체할 것. go/js 수치는 영향 없음.

**E2** — `spec.md`:214-222 (§C.2 추출표) — **표가 정규식 문법의 두 클래스를 이름 붙이지 않는다**: (a) 대소문자 무시 플래그 `(?i)` — 이 경우 대소문자 구분 `strings.Contains` 토큰은 대소문자 무시 룰에 대해 **불건전**(오늘 deny 되는 `SK-…` 를 프리필터가 skip)하다; (b) 분기 앞의 선택/수량 그룹(`(sk-)?AKIA` 형) — 첫 리터럴이 필수가 아닌데 접두로 채택될 수 있다. 배포 룰셋에는 **둘 다 없음**을 실측 확인했다(`grep -rn "(?i)" …astgrep-rules/` rc=1; error 심각도 정규식 4개 전수 판독). 따라서 오늘의 불건전성은 없고 순수 잠복이다 — 그러나 §C.2 는 건전성을 "by construction" 으로 주장하고 REQ-SSS-009 의 underivable 목록도 이 클래스를 담지 않으므로, 향후 룰 추가가 신호 없이 deny 를 좁힐 수 있다. — Severity: **minor** — Class: **blocking**(건전성 주장의 범위 문제이므로) — Required fix: §C.2 표의 underivable 행에 "a `regex:` carrying an inline flag that changes literal matching (e.g. `(?i)`), or a branch whose leading literal is optional/quantified" 를 추가할 것. 한 줄이면 끝난다.

**E3** — `acceptance.md` AC-SSS-016 2행 — **기계화된 짝 검사가 REQ-SSS-015 가 구속하는 축과 다른 축을 돈다.** AC 가 인용한 가드형 루프는 `templates/` **내부** 축(`templates/x.sh` ↔ `templates/x.sh.tmpl`)을 도는데, 실측상 그 축에서 실제로 비교되는 짝은 35개 중 **4개**뿐이다(`pairs actually compared: 4`, DRIFT 0). 반면 REQ-SSS-015 의 의무("any hook wrapper this SPEC changes shall move as a `.sh` / `.sh.tmpl` pair")와 §D 의 out-of-scope 노트가 말하는 축은 **배포본 ↔ 템플릿**(`.claude/hooks/moai/x.sh` ↔ `templates/…/x.sh.tmpl`)이다. `handle-pre-tool.sh` 는 `templates/` 안에 `.sh` 사본이 없으므로 AC 의 루프가 애초에 건드릴 수 없다 — 즉 §D 노트가 그 루프의 범위 축소를 정당화하는 것처럼 읽히지만 두 문장은 서로 다른 축을 말한다. 정작 올바른 축은 SPEC 이 커밋한 `driftcheck.sh` 의 `pairaxis` 모드에 **이미 구현돼 있고**(실행 시 `DRIFT .claude/hooks/moai/handle-pre-tool.sh` 정확히 1건 보고), 어떤 AC 도 그 모드를 인용하지 않는다. — Severity: **minor** — Class: **optional** — Required fix(선택): AC-SSS-016 에 `bash .moai/reports/t217/driftcheck.sh pairaxis` 행을 추가하고 pre-implementation 측정값을 `handle-pre-tool.sh 1건(§D 기록된 기존 드리프트)` 으로 명시하거나, §D 노트에 "두 축은 다르다" 를 한 줄 덧붙일 것.

**E4** — `acceptance.md` AC-SSS-015 3행 — 인용 명령이 `printf '{}' | ./bin/moai hook security-scan` 인데 `./bin/moai` 는 이 트리에 없다(`ls bin/moai` → `No such file or directory`; `make build` 산출물이며 gitignore 대상). "Measured today: `exit=0`" 는 그 명령으로는 측정될 수 없었다. 동작 주장 자체는 참임을 설치본으로 재확인했다(`printf '{}' | ~/go/bin/moai hook security-scan; echo "exit=$?"` → `exit=0`). M4 의 `make build` 이후에는 인용 경로가 유효해지므로 실행 시점 문제이지 사실 오류는 아니다. — Severity: **minor** — Class: **optional** — Required fix(선택): 측정 시점을 "after M4's `make build`" 로 명시하거나 인용을 `~/go/bin/moai`(또는 `moai`)로 바꿀 것.

**E5** — `spec.md`:102-105·134-138·149-154·164-168 — **D6 통합의 부작용으로 복합 요구가 생겼다.** REQ-SSS-001(deny 보존 + 신규 deny 금지), REQ-SSS-008(순수함수 + severity 한정 + GuardianPatterns 금지), REQ-SSS-012(두 캐리어 필드 보존 + 도달성), REQ-SSS-015(미러 + 짝 + 금지 콘텐츠 6종)가 각각 2~3개 의무를 한 항목에 담는다. 결과로 REQ 단위 원자 PASS/FAIL 이 불가능하고, REQ-SSS-015 는 서로 다른 4개 검사를 가진 단일 AC(016)에 사상된다. **MP-2 실패는 아니다** — 각 문장이 개별적으로 GEARS 형식이고, 비형식 언어도 Given-When-Then 의 요구 레이어 침입도 없다. 16 상한을 맞추기 위한 통합의 대가이며, 상한 준수가 이 대가보다 우선한다는 판단은 합리적이다. — Severity: **minor** — Class: **optional** — Required fix(선택): 없음. 대안(Tier L 승급)은 `design.md` + `research.md` 를 추가로 요구해 더 비싸다.

**E6** — `spec.md`:272-277 (§C.3 "Why A1 is kept") — "would deny a **93-96% saving to every** JavaScript, TypeScript, and Python project the tool ships to" 가 단일 저장소의 js 81 / py 14 파일 표본을 전 배포 프로젝트로 일반화한다. 같은 절 Gaps 문단(spec.md:279-283)이 그 전이 가능성을 스스로 부정하므로 두 문단이 긴장한다. 유지 논거는 이 전이를 필요로 하지 않는다 — "기전이 데이터 구동이라 룰셋이 자라면 비율이 자동 조정된다" 가 독립적으로 성립한다. — Severity: **minor** — Class: **optional** — Required fix(선택): 해당 문장을 "would deny the measured saving to JavaScript / TypeScript / Python projects" 정도로 완화하거나, 뒤의 자동조정 논거를 앞세울 것.

---

## Regression Check (iteration 1 결함 6건)

| iter1 | 상태 | 근거 |
|---|---|---|
| D1 frontmatter 파싱 실패 | **RESOLVED** | `moai spec lint` → `✓ No findings`; `spec.md:13` 이 인용 콤마 문자열 |
| D2 §C.2 가 지배적 룰 형태 미커버 | **RESOLVED** | 7행 추출표 + `kind:`+`regex:` 행 + 교대 행 + 워크드 예제(spec.md:227-241). 렌즈 1 에서 반례 0건, 이스케이프 가설은 실측 반증 |
| D3 AC-SSS-004 공허 통과 | **RESOLVED** | 카운터 2개로 분리, 오늘 (scanner 1, caller 0) → PASS (scanner 0, caller 1). 반전 확인 |
| D4 drift 명령 미구현 트리 실패 | **RESOLVED** | `[ -f "$b" ]` 가드 복원. 실측 guarded=0 / unguarded=31 양쪽 재현 |
| D5 계측 seam 부재 | **RESOLVED** | M0 step 1 이 인터페이스 축소를 소유하고 의존 AC 앞에 배치. 내가 제안했던 핸들러 레벨 스텁도 불가함을 §A.1 에 실측 기록하고 정정 |
| D6 Tier M REQ 예산 초과(20>16) | **RESOLVED** | 16 unique 실측. 20→16 전량 사상 확인, 내용 손실 없음(렌즈 4) |
| D7 단락 도달성 | **ADOPTED** | REQ-SSS-012 마지막 문장 + §A.3 근거 기록 + AC-SSS-012 두 번째 절 |
| D8 빈 언어집합 | **ADOPTED** | REQ-SSS-006 "or yields an empty covered-language set" + AC-SSS-007 3팔 |

정체(stagnation) 결함 0건. 미해결 이월 0건.

---

## Gaps (관측하지 못한 것)

- `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` (AC-SSS-016 3행)은 **실행했으나 완료를 관측하지 못했다** — 120초 타임아웃 후 백그라운드로 넘어갔고 출력 파일이 종료 시점까지 0바이트였다. SPEC 이 기재한 "measured today: pass" 를 내가 확인하지도 반증하지도 못했다. 부재는 실패의 근거가 아니다.
- 배포 사용자 환경에서의 실제 deny 발화 여부는 관측 범위 밖이다(investigation Claim 1 의 Gaps 와 동일). SPEC 도 §D 에서 양방향 모두 주장하지 않는다.
- 렌즈 1 의 반례 탐색은 **현재 배포 룰셋**에 한정된다. 사용자가 공급한 임의 룰셋에 대한 건전성은 underivable 폴백(REQ-SSS-009)에 의존하며, 그 폴백의 실제 동작은 M2 구현 후 AC-SSS-010 이 관측할 사항이다.

## Residual-risk

추출기는 이 SPEC 에서 가장 결함이 나기 쉬운 단위이면서 deny 를 지킨다. 명세는 이제 건전하지만 구현은 아직 없다. 이를 묶는 것은 세 가지다: underivable⇒escalate 폴백(REQ-SSS-009), M0 차등 코퍼스(AC-SSS-001 (i)), 그리고 AC-SSS-001 (ii) — "deny 하는 모든 fixture 에 대해 프리필터가 skip 하지 않았을 것". 세 번째가 실질적 안전망이며, 코퍼스가 커버 언어당 최소 1건의 deny 를 갖도록 강제하는 게이트가 그것을 비지 않게 한다.

---

## Recommendation

**PASS.** 일곱 개 must-pass 전부 통과, 총점 0.925 로 Tier M 임계 0.80 을 넘고, iteration 1 의 6개 결함이 모두 해소됐으며 채택형 2건(D7/D8)도 요구 조항으로 들어갔다.

판정 근거를 명시해 둔다 — E1 을 blocking 으로 분류하면서 verdict 를 PASS 로 두는 것은 완화가 아니다. M5 Must-Pass Firewall 의 일곱 기준 중 어느 것도 E1 이 아니고, 루브릭 점수는 0.925 이며, E1 의 크기는 n=14 에서 7.1%p 로 그것이 떠받치는 결정(A1 유지)을 뒤집지 않는다. non-must-pass 결함 목록으로 FAIL 을 제조하지 말라는 M6 이 정확히 이 상황을 가리킨다. 다만 §C.3 이 그 수치를 "measured, not asserted" 로 규정했으므로 **틀린 채로 두는 것은 허용되지 않는다**.

run-phase 진입 전에:

1. **E1 (major, blocking)** — `skiprate.py` python 토큰 2개 추가 → 재실행 → `spec.md:64`·`spec.md:263` 수치 교체. 편집 3곳, 재측정 1회.
2. **E2 (minor, blocking)** — §C.2 underivable 행에 정규식 인라인 플래그 + 선택/수량 접두 클래스 한 줄 추가.
3. **E3 / E4 / E5 / E6 (optional)** — 오케스트레이터 재량. E3 은 `driftcheck.sh pairaxis` 가 이미 존재하므로 AC 한 줄 추가로 끝나고, 나머지 셋은 문면 정리다. 채택하지 않아도 verdict 에 영향 없다.

특기할 점 — 이 개정판이 iteration 1 과 다른 지점은 결함을 고쳤다는 것이 아니라, **고칠 때 근거를 만들어 커밋했다**는 것이다. §C.3 은 이전 판의 단언을 재현 가능한 스크립트로 대체했고, `acceptance.md` 의 "The instrument" 문단은 감사자(나)가 제안했던 잘못된 계측기까지 실측으로 폐기했으며, `driftcheck.sh` 는 D4 의 두 변형을 모두 보존해 31 대 0 을 재현 가능하게 만들었다. 남은 여섯 결함 중 다섯이 minor 인 것은 그 방식의 결과다.
