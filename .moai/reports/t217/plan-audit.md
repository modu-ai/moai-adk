# SPEC Review Report: SPEC-SEC-SCAN-SURFACE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.65** (Tier M PASS threshold = 0.80)

측정 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t217`, 브랜치 `security-scan-surface`, HEAD `c5d08ce0ed916d992ab4e79423e8a3b9bac3aa2d` (`git rev-parse HEAD` 실측).
Reasoning context ignored per M1 Context Isolation — 감사는 `spec.md` / `plan.md` / `acceptance.md` (Tier M 입력 집합) + 근거 파일 `investigation.md` 만 읽고 수행했다. 룰셋 파일은 워크트리에 없어 primary 체크아웃(`/Users/goos/MoAI/moai-adk-go/.moai/config/astgrep-rules/`)과 템플릿 사본에서 읽었다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o "REQ-SSS-[0-9]\{3\}" spec.md | sort -u` → 20건, `REQ-SSS-001`…`REQ-SSS-020` 연속, 결번·중복 없음. 굵은 선언(`^- \*\*REQ-SSS-`) 수도 20으로 일치.
- **[PASS] MP-2 GEARS format compliance** — **요구 레이어(`REQ-XXX` in `spec.md`)에 대해** 판정. 20개 전부 다섯 패턴 중 하나에 맞는다: Ubiquitous(001·002·004·006·009·011·012·013·015·016·017·018·019·020), Where/When 복합(003·005·008·010, spec.md:94·104·115·121), When(007, L109), Unwanted(011 "shall not be derived from…", L124). `acceptance.md` 의 Given-When-Then 은 검증 레이어이므로 여기서 감점하지 않았다(Group 4 소관).
- **[FAIL] MP-3 YAML frontmatter validity** — 도메인 전용 도구로 확인. **D1 참조.**
  ```
  $ ~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md
  ERROR  ParseFailure  …spec.md  1  SPEC parsing failed: frontmatter parsing error:
    YAML parsing error: yaml: unmarshal errors:
      line 13: cannot unmarshal !!seq into string
  1 error(s), 0 warning(s)
  ```
- **[PASS] MP-4 Section 22 language neutrality** — 이 SPEC 은 `internal/hook` Go 서브시스템 범위이고, 언어별 도구명을 "primary"로 승격시키지 않는다. 오히려 REQ-SSS-006(spec.md:106)이 커버 언어 집합을 **설정에서 파생**하도록 강제하므로, 5번째 언어 룰이 추가되면 스캔이 자동 재개된다 — 중립성을 코드에 고정하지 않고 데이터로 옮기는 방향이다. 하드코딩 언어 목록 금지 조항도 같은 REQ 에 있다.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE "SPEC-([A-Z][A-Z0-9]+-)+[0-9]+" spec.md | sort -u` → `SPEC-SEC-SCAN-SURFACE-001` 자기 자신 1건뿐. retired/superseded 참조 없음 ⇒ BLOCKING 없음.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -n "syscall" spec.md` → 0건(rc=1). D8-4에 따라 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-SEC-SCAN-SURFACE-001/` → 0건(rc=1). `research.md` 는 Tier M 이라 부재가 정상.

MP-3 단독으로 M5 Must-Pass Firewall 이 발동해 총점과 무관하게 `Verdict: FAIL` 이다.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.50 | 0.50 | §C.2(spec.md:207-210)의 추출 규칙이 **실제 룰셋의 지배적 형태**에 대해 두 갈래로 읽히고, 두 해석의 결과가 정반대다 — D2 |
| Completeness | 0.75 | 0.75 | 필수 섹션 전부 존재, Out of Scope H3 6개(L226·234·241·247·254·261) 각각 구체 bullet 보유. 감점: frontmatter 타입 오류(D1) + Tier M REQ 예산 초과(D6) |
| Testability | 0.50 | 0.50 | 16개 AC 중 3개가 관측을 못 한다: 공허 통과(D3), 미구현 트리에서 이미 실패(D4), 존재하지 않는 seam 의존(D5) |
| Traceability | 0.85 | 0.75~1.0 경계 | §E(spec.md:270-278)가 REQ 20개를 AC/표면에 전부 사상. REQ-SSS-012(순수함수 요구)만 AC-SSS-008/009 를 통해 **간접** 커버 — 직접 검증 AC 없음 |

산술 평균 0.65. Tier M 임계 0.80 미달.

---

## Defects Found (structured defect-list)

**D1** — `spec.md`:L13 — `tags: [security, hook, pretooluse, posttooluse, ast-grep, performance]` 가 YAML **시퀀스**인데, 캐논 스키마(`.claude/rules/moai/development/spec-frontmatter-schema.md` § Field Reference)와 `internal/spec/lint.go:403` `Tags string \`yaml:"tags"\`` 는 **콤마 구분 문자열**을 요구한다. 결과는 필드 하나가 비는 정도가 아니라 **frontmatter 전체 파싱 실패**(`ParseFailure`, ERROR): 12필드 검증·era 분류(`internal/spec/era.go`)·드리프트 탐지가 이 SPEC 에 대해 통째로 죽는다. — Severity: **critical** — Class: **blocking** — Required fix: `tags: "security, hook, pretooluse, posttooluse, ast-grep, performance"` 로 교체하고 `~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md` 가 0 error 로 떨어지는 것을 확인할 것.

**D2** — `spec.md`:L207-210 (§C.2 두 번째 bullet) — **pre-filter 추출 규칙이 실제 룰셋의 지배적 형태를 다루지 않는다.** §C.2 는 `pattern:` 의 리터럴 런과 `regex:` 의 "필수 리터럴 접두"만 정의하고, "`kind:`-only rules"를 underivable 로 분류한다. 그런데 배포 룰셋의 `sec-hardcoded-credential` 은 **네 커버 언어 전부**에서 `pattern:` 없이 `kind:` + `regex:` 형태다:
```
# internal/template/templates/.moai/config/astgrep-rules/security/credentials.yml
rule:
  kind: interpreted_string_literal          # go  (js/ts/python 은 kind: string)
  regex: "^\"(sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})"
```
이 형태를 "`kind:`-only" 로 읽으면 **go/js/ts/python 네 언어 전부 underivable** ⇒ REQ-SSS-010 에 따라 무조건 escalate ⇒ **A1 은 어떤 언어에서도 단 한 번도 skip 하지 않는다**(순수 비용). 반대로 "리터럴 앵커가 있으니 derivable" 로 읽으면 정규식 **교대(alternation)** 처리 규칙이 필요한데 §C.2 에 그 규칙이 없다 — `all:`/`any:` 만 정의돼 있다. 두 해석은 AC-SSS-011(§B.4, `.js` payload 가 skip 되어야 함)의 달성 가능 여부를 정반대로 만든다. 덧붙여 §C.2 의 정직성 문단(spec.md:214-218)은 "Go 만 잘 안 걸리고 나머지 세 언어는 절감이 크다"고 단언하는데, 이 룰이 네 언어에 동일하게 걸려 있으므로 그 단언은 D2 가 해소되기 전까지 **미검증 전제**다(VCI §1.1 surface 4). — Severity: **critical** — Class: **blocking** — Required fix: §C.2 에 (a) `pattern:` 없이 `kind:`+`regex:` 인 룰의 처리(derivable/underivable)를 명시하고, (b) derivable 로 갈 경우 정규식 최상위 교대의 각 분기에서 필수 리터럴 접두를 뽑아 **합집합**을 토큰으로 삼는다는 규칙(= `any:` 와 동일 논리, 건전함)을 추가할 것. 그 위에서 §C.2 의 언어별 절감 서술을 실제 룰셋 기준으로 다시 쓸 것.

**D3** — `acceptance.md`:L58-65 (AC-SSS-004) — **미구현 트리에서 이미 통과하는 공허한 기준.** AC 는 "`FindRulesConfig` 호출 카운터가 정확히 **1**"을 요구하고 baseline 을 "**2** on the pre-implementation tree" 라고 적었다. 실측은 1이다:
```
$ grep -rn "FindRulesConfig" internal --include="*.go" | grep -v _test
internal/hook/security/scanner.go:84    # ScanFile — pre-write 경로가 쓰는 유일한 호출
internal/hook/security/scanner.go:109   # ScanFiles — pre-write 경로 아님
internal/hook/security/rules.go:91      # GetEffectiveRules — pre-write 경로 아님
```
`pre_tool.go:638` 은 `h.scanner.ScanFile(ctx, tmpFile.Name(), h.projectRoot())` 만 부르고 자기 쪽에서는 해석하지 않는다 ⇒ 현 트리 카운트 = **1** = AC 의 PASS 값. AC 자신도 괄호에서 "measured after M1's reorder" 라고 인정하고 있어, 적힌 baseline 은 존재한 적 없는 중간 상태의 수치다. — Severity: **major** — Class: **blocking** — Required fix: baseline 을 1로 정정하고, 관측 대상을 "caller 가 해석하고 scanner 는 재해석하지 않는다"로 바꿀 것 — 예: **scanner 쪽 `FindRulesConfig` 호출 0회** + caller 쪽 1회를 각각 세는 두 카운터. 그래야 미구현 트리에서 (0,1) 이 아니라 (1,0) 이 나와 RED 가 성립한다.

**D4** — `acceptance.md`:L191 (AC-SSS-016 두 번째 명령) — **미구현 트리에서 이미 실패하는 기준.** 명령은 "must print nothing" 인데 실측 31줄이 찍힌다:
```
$ python3 …  # for f in templates/.claude/hooks/moai/*.tmpl; b=${f%.tmpl}; diff -q "$b" "$f"
tmpl files: 35
MISSING base (.sh absent): 31
content DRIFT among existing pairs: 0
AC-SSS-016 loop would print DRIFT lines: 31
security-scan base exists: False
```
템플릿 디렉터리에는 대부분 `.tmpl` 만 있고 `.sh` 짝이 없다(`handle-security-scan.sh` 도 없다). 원본(`CLAUDE.local.md` §2.3)에는 `[ -f "$b" ] &&` 가드가 있는데 AC 로 옮기면서 그 가드가 빠졌다. 현 상태로는 이 AC 가 영원히 닫히지 않아 §A Definition of Done 전체를 막는다. — Severity: **major** — Class: **blocking** — Required fix: `[ -f "$b" ] && { diff -q "$b" "$f" >/dev/null || echo "DRIFT $b"; }` 로 가드를 복원하고, baseline 을 "현 트리에서 0줄"로 실측해 명시할 것. 아울러 이 SPEC 이 실제로 건드리는 짝은 `.claude/hooks/moai/*.sh` ↔ `templates/.claude/hooks/moai/*.sh.tmpl` 이므로, 검사 범위를 그 축으로 다시 잡을 것.

**D5** — `plan.md`:L54 (§D 두 번째 제약) + `acceptance.md` AC-SSS-002/005/006/007/010/011 — **비용 계측 seam 이 이 경로에 존재하지 않는다.** plan 은 "The scanner's `sg` call is injectable (`astGrepScanner.scanFunc`), so the count is observable" 라고 전제하지만, `scanFunc` 는 **`ScanMultiple` 전용**이다:
```
internal/hook/security/ast_grep.go:31   scanFunc func(...)   // 주석: "for ScanMultiple"
internal/hook/security/ast_grep.go:195-200   // scanFn 선택 — ScanMultiple 안에서만
internal/hook/security/ast_grep.go:100-137   // Scan(단일 파일): exec.CommandContext(ctx,"sg",…) 직접 호출, scanFunc 미참조
```
`scanWriteContent` → `ScanFile` → `astGrep.Scan` 이 이 SPEC 이 바꾸는 경로인데 여기엔 주입 지점이 없다. "counting stub" 을 쓰는 6개 AC 는 현재 실행 불가능하고, seam 을 만드는 작업을 소유한 마일스톤도 없다. — Severity: **major** — Class: **blocking** — Required fix: M1 에 "단일 파일 `Scan` 경로에 주입 seam 추가" 를 명시적 단계로 넣거나(권장 — `ASTGrepScanner` 인터페이스를 통한 스텁 주입은 이미 `preToolHandler` 가 `NewPreToolHandlerWithScanner` 로 받으므로 상위 레벨 스텁이 더 쉽다), AC 들의 계측 지점을 `scanner.ScanFile` 호출 횟수 + temp 파일 생성 유무로 재정의할 것.

**D6** — `spec.md` 전체 (frontmatter `tier: M`, L14) — **Tier M REQ 예산 초과.** `spec-workflow.md` § SPEC Complexity Tier 의 요구 상한은 Tier M = 16 인데 이 SPEC 은 REQ 20개다(AC 는 16으로 상한 정확히 충족). 상한은 "티어를 올리거나 SPEC 을 쪼개라"는 신호이지 완화 대상이 아니며, 부담은 전량을 동시에 붙들어야 하는 감사자에게 떨어진다. — Severity: **minor** — Class: **blocking**(명시된 규칙 위반이므로) — Required fix: `tier: L` 로 승급(단 Tier L 은 `design.md` + `research.md` 를 추가로 요구)하거나, 배포·공시 축(REQ-SSS-018·019·020)을 별도 카드로 분리해 16 이하로 줄일 것.

**D7** — `spec.md`:L134-136 (REQ-SSS-014) — **병합 후 등록 순서 의존성이 요구로 진술되지 않았다.** `internal/hook/registry.go:142-149` 의 `Dispatch` 는 앞선 핸들러가 block 결정을 내면 **남은 핸들러를 단락(short-circuit)** 시키고, `:159` 는 `ExitCode == 2` 에서도 같은 동작을 한다. 지금 guardian 은 Claude Code 가 별도 프로세스로 부르므로 이 단락에 면역이지만, 병합 후에는 등록 순서에 종속된다. 현재는 잠복 위험이다 — `post_tool.go` 는 관측 전용이라(`:144` "Always returns Decision \"allow\" (observation only)") 실제로 막지 않는다. 그러나 REQ-SSS-014 는 "둘 다 non-empty 일 때 둘 다 실린다"만 말하고 단락 불변식을 말하지 않으며, AC-SSS-012 도 행복 경로만 본다. — Severity: **minor** — Class: **optional** — Required fix(선택): REQ-SSS-014 에 "선행 핸들러의 결정이 guardian 핸들러의 도달을 막지 않는다"를 추가하거나, 잠복임을 §A 에 근거와 함께 기록할 것.

**D8** — `spec.md`:L128 참조 지점 (REQ-SSS-005/007) — **"커버 언어 0개" 결과가 "설정 고장"과 구분되지 않는다.** ruleDirs 가 가리키는 디렉터리가 없는 경우(이 리포에 전례가 있다) 워크는 언어를 하나도 못 찾는다. REQ-SSS-007 은 "읽기/파싱/워크 실패"만 escalate 로 규정하므로, "워크는 성공했고 결과가 0" 이면 **모든 언어가 skip** 되어 게이트가 조용히 deny 를 멈춘다. 실측상 지금은 결과가 같아서(아래) 즉시 사고는 아니다:
```
$ sg scan -c /tmp/t217-sgtest/sgconfig.yml --json /tmp/t217-audit-probe.go
Error: Cannot read rule directory /tmp/t217-sgtest/go   # findings 0 — 오늘도 deny 없음
```
즉 deny 함수는 불변이나, 안전성 근거가 SPEC 에 적혀 있지 않다. — Severity: **minor** — Class: **optional** — Required fix(선택): "파생된 커버 언어 집합이 공집합이면 unknown 으로 간주하고 escalate" 를 REQ-SSS-007 에 붙이거나, 위 실측을 §A 에 근거로 기록할 것.

---

## 근거가 확인되어 결함이 아닌 것 (요청된 렌즈별)

- **렌즈 3 — A2 안전성 주장은 실측으로 성립한다.** `investigation.md` 는 `cd /tmp` 에서만 확인했으나, 더 강한 조건(cwd = 프로젝트 루트, 그 아래 `.moai/config/astgrep-rules/sgconfig.yml` 존재)에서도 `sg` 는 자체 상향 탐색으로 그 설정을 찾지 못한다:
  ```
  $ cd /Users/goos/MoAI/moai-adk-go && sg scan --json /tmp/t217-audit-probe.go
  Error: No ast-grep project configuration is found.
  ```
  따라서 `configPath == ""` ⇒ findings 0 이고, REQ-SSS-003 의 skip 은 deny 를 하나도 제거하지 않는다.
- **렌즈 5 — Item B 병합 주장은 코드로 확인된다.** `registry.go:180-230` `mergeHandlerOutput` 은 `SystemMessage`(L182-188)와 `hookSpecificOutput.AdditionalContext`(L224-230)를 **각각** `"\n"` 으로 누적한다. 두 필드가 서로를 덮지 않는다. guardian 의 반출 필드도 `additionalContext` 가 맞다(`guardian.go:56-60`). `spec.md` §A.1-1 이 카드의 전제("additionalContext 끼리 충돌")를 정정한 것은 정확한 지적이다.
- **렌즈 5 — 설정 엔트리 제거가 깨는 소비자는 없다.** `security-scan` 서브커맨드 참조는 `internal/cli/hook.go:167·561`(등록·핸들러), 테스트 3건, 래퍼 `handle-security-scan.sh`(+`.tmpl`) 뿐이다. REQ-SSS-016 + plan §G 의 "래퍼는 남긴다" 결정으로 커버된다.
- **렌즈 6 — 미러 집합 열거는 완전하다.** 이 SPEC 이 건드리는 `.claude/` 파일은 `settings.json` 하나이고, REQ-SSS-017 이 `.json.tmpl` 을 함께 명시한다. `handle-security-scan.sh`/`.sh.tmpl` 은 변경 대상이 아니다. (기계화 방식만 D4 로 결함.)
- **렌즈 7 — PR 공시 요구는 존재한다.** REQ-SSS-020(spec.md:154-157)이 제목 + 본문 첫 문장을 구속하고, AC-SSS-016 마지막 절이 `gh pr view --json title,body` 로 **출력을 읽어** 검증한다(소스 grep 아님). 카드의 [HARD] 를 만족한다.
- **렌즈 4 — 지연시간 단언은 어디에도 없다.** `acceptance.md` §D 가 명시적으로 금지하고, 실제로 16개 AC 중 밀리초·퍼센트를 단언하는 것은 0건이다.

---

## Recommendation

FAIL. 아래 순서로 고치고 재감사(iteration 2, Tier M 상한)를 받을 것. 재감사는 이 결함 델타에 한정된다.

1. **D1 (critical, 즉시)** — `spec.md:13` 을 `tags: "security, hook, pretooluse, posttooluse, ast-grep, performance"` 로 바꾸고 `moai spec lint` 0 error 확인. 이것이 안 고쳐지면 나머지 어떤 점수도 무의미하다(MP-3).
2. **D2 (critical)** — §C.2 에 `kind:`+`regex:` 룰의 처리와 정규식 교대 합집합 규칙을 명시. 그 결정 후 §C.2:214-218 의 언어별 절감 서술을 실제 룰셋 기준으로 재작성하고, AC-SSS-011 이 여전히 달성 가능한지 다시 판단할 것(underivable 로 결정한다면 AC-SSS-011 은 삭제되거나 다른 언어로 옮겨져야 한다).
3. **D5 (major)** — M1 또는 M0 에 "단일 파일 스캔 경로의 주입 seam" 단계를 추가하거나, 6개 계측 AC 의 관측 지점을 재정의. 이것 없이는 A2/A3/A1 의 어떤 절감도 증명할 수 없다.
4. **D3 (major)** — AC-SSS-004 의 baseline 을 1로 정정하고 카운터를 두 개로 쪼개 RED 가 성립하게 할 것.
5. **D4 (major)** — AC-SSS-016 drift 루프에 `[ -f "$b" ]` 가드 복원 + baseline 0줄 실측 명시.
6. **D6 (minor, blocking)** — Tier 승급 또는 REQ 20 → 16 이하로 분리.
7. **D7 / D8 (optional)** — 오케스트레이터 재량. 채택하면 §A 근거 추가 또는 REQ 한 줄 보강으로 끝난다. 채택하지 않아도 verdict 에 영향 없다.

특기 — 이 SPEC 의 §A / §A.1 / §D 는 감사 대상으로서 **드물게 정직하다**: 카드의 전제 두 개를 스스로 반박했고, 절대 시간 수치를 근거로 쓰지 않겠다고 선언했으며, 배제 항목 6개가 전부 이유와 함께 적혀 있다. 위 결함들은 그 판단력의 부재가 아니라, 실제 룰셋 파일(D2)과 실제 코드 seam(D5)을 확인하지 않고 쓴 전제에서 나왔다.
