# t562 sync-audit — SPEC-CODEX-SKILL-PATH-READBACK-001 (lens: --security)

감사 주체: `sync-auditor` (독립 감사, 레인 자기보고 미채택)
감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562` · 브랜치 `WT-codex-read-inverse` · HEAD `bbba4f672`
평가 프로필: `.moai/config/evaluator-profiles/default.md` (`harness.yaml default_profile: "default"`, SPEC 프론트매터에 `evaluator_profile` 없음)

**판정: PASS** — must-pass 두 축(Functionality / Security) 모두 통과. 차단 소견 0건, 비차단 소견 7건.

---

## 1. Claim (주장)

1. 생산 코드 변경은 `codexPathAbsolute` 분기의 한 줄 변환 두 개가 전부이며(`c007e5409`), 홈-상대 분기는 손대지 않았다.
2. 분류(`classifyCodexSkillPath`)는 선언 형태 위에서, 변환보다 **먼저** 돈다.
3. 이 변환은 어떤 입력 계열에서도 해석 가능한 경로를 해석 불가로 뒤집지 않는다 — 즉 허위 `fs.ErrNotExist` → `Eligible: true` → 삭제 경로를 새로 열지 않는다.
4. darwin 에서 변환은 항등이며, Windows 쪽 효과는 구조적 추론뿐이다.
5. run→sync 경계에서 쓰기 가능 에이전트 둘이 겹쳤으나 유실은 0이다.
6. AC 10건 전부 PASS.

---

## 2. Evidence (증거 — 실행한 명령과 그 출력)

### 2.1 트리 접지

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562
$ git rev-parse HEAD
bbba4f672e4a4a5d2a6e90fbecbb3fea991df2a4
$ git branch --show-current
WT-codex-read-inverse
```

### 2.2 생산 diff 가 주장대로인가 (주장 1)

`git show c007e5409 -- internal/cli/codex_skills_prune.go internal/cli/doctor_codex.go` 는 두 파일 각각에서 `-\t\tstatPath = e.Path` → `+\t\tstatPath = fromConfigPath(e.Path, configPathSeparator)` 한 줄과 주석만 낸다. 다른 실행 문장 변경 없음.

카드 커밋 전수 스윕 — `internal/cli/*.go` 를 만진 커밋:

```
bbba4f672 internal-go-files=0    18bf8cc06 internal-go-files=0
37a631a70 internal-go-files=0    757ef601f internal-go-files=0
4aa8915ee internal-go-files=0    835215bab internal-go-files=0
c007e5409 internal-go-files=2  ← 생산 수정 (2파일)
b78d2e425 internal-go-files=1  ← 테스트 개명만
f1654c924 internal-go-files=2  ← M1 테스트 2본
```

REQ-CSRB-006 범위 핀 — 카드 자기 커밋 9개 중 `codex_config_path.go` / `codex_skills_disable.go` / `codexwiring/skills.go` 를 만진 커밋: **0건** (빈 출력).

### 2.3 [보안 핵심] 적대적 입력 행렬 (주장 2·3)

감사 전용 프로브 테스트를 주입해 선언 17종 × 분리자 2종(`/`, `\`)을 `judgeCodexSkillEntry` 에 통과시키고, `osStatFn` 에 실제로 넘어간 문자열과 판정을 기록했다. 전문: `.moai/reports/t562/audit-probe.log` (rc=0).

| 선언 | sep | CLASS | stat 대상 | Eligible |
|---|---|---|---|---|
| `C:/Users/u/SKILL.md` | `\` | 2 relative | *(stat 없음)* | false |
| `C:\Users\u\SKILL.md` | `\` | 3 oddly | *(stat 없음)* | false |
| `//?/C:/Users/u/SKILL.md` | `\` | 0 abs | `\\?\C:\Users\u\SKILL.md` | true |
| `//server/share/SKILL.md` | `\` | 0 abs | `\\server\share\SKILL.md` | true |
| `/a/../../etc/passwd` | `\` | 0 abs | `\a\..\..\etc\passwd` | true |
| `/` · `//` · `/tmp/` | `\` | 0 abs | `\` · `\\` · `\tmp\` | true |
| `~/x\SKILL.md` | `\` | 1 home | `/Users/goos/x\SKILL.md` *(미변환)* | true |
| `~root/x` | `\` | 3 oddly | *(stat 없음)* | false |
| `relative/SKILL.md` · `./SKILL.md` · ` /tmp/…` | `\` | 2 relative | *(stat 없음)* | false |

판독:

- **주장 2 성립.** `C:/…` 는 sep 을 `\` 로 주입해도 CLASS=2(relative)로 남고 stat 호출이 0이다. 변환이 분류보다 먼저 돌았다면 `\` 를 품은 문자열이 되어 CLASS=3(oddly)이 됐을 것이고, 그 둘은 서로 다른 SkipReason 문자열을 낸다 — 순서는 여기서 **관측 가능**하다.
- **주장 3 성립(단, 구조적).** 변환은 분리자 조건부이고, 산출은 언제나 그 호스트의 native 형태다. 허위 `ErrNotExist` 를 만들려면 native 형태가 slash 형태보다 **덜** 해석돼야 하는데, 그 반대다. 유일한 이론적 반례는 파일명에 `/` 를 담은 Windows 경로인데 Windows 파일명 문자집합이 `/` 를 금지한다 — 이것은 **인용된 전제이지 이 런에서 잰 사실이 아니다**(§4 G-A).
- **트래버설(`..`)은 삭제 원시함수가 아니다.** `Eligible` 이 지우는 것은 config 의 **줄 범위**이지 stat 대상 파일이 아니다(`pruneCodexSkillEntries`). 게다가 `..` 계열은 변환 전에도 동일하게 eligible 이었다 — 이 카드가 만든 표면이 아니다.
- **부수 소득 — 이 수정이 실제로 값을 하는 유일한 계열을 찾았다.** `//?/C:/…` → `\\?\C:\…`. Windows 는 확장길이 경로(`\\?\`)에서 `/` 를 정규화하지 **않으므로**, 그 계열에서는 미변환 형태가 진짜로 해석에 실패한다. SPEC 은 이 계열을 한 번도 이름 붙이지 않았다(§F6).

프로브는 되돌렸고 트리 청결을 증명했다:
```
$ rm -f internal/cli/zz_audit_probe_test.go && git status --porcelain -- internal/
(빈 출력)
```

### 2.4 기계 게이트 (이 런, 이 트리에서 재실행)

```
$ go vet ./internal/cli/                                   → rc=0            (.moai/reports/t562/audit-vet.log)
$ golangci-lint run --timeout=5m ./internal/cli/           → rc=0, "0 issues." (audit-lint.log)
$ go test ./internal/cli/... -count=1 -timeout 1800s       → rc=0, ok 17개 패키지, (cached) 0건 (audit-test.log)
$ go test ./internal/cli/ -count=1 -cover -timeout 1800s   → ok 449.195s coverage: 81.2% of statements (audit-cover.log)
$ gofmt -l <변경 5파일>                                     → 빈 출력
```

카드 자기 테스트 계열 재실행 — `PASS=33 FAIL=0` (audit-readback.log), 그중 이 카드가 만든 것:
```
--- PASS: TestJudgeCodexSkillEntry_SeparatorConversion (0.00s)
--- PASS: TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion (0.00s)
--- PASS: TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative (0.00s)
--- PASS: TestJudgeCodexSkillEntry_EligibilityGatingPins (0.00s)
--- PASS: TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard (0.01s)
```

### 2.5 AC-CSRB-010 수치 재유도 (레인 수치의 독립 검산)

레인이 인용한 `6938 / 0 / 30` 을 커밋된 로그에서 두 가지 grep 형태로 다시 셌다:

```
$ /usr/bin/grep -c '^--- PASS: ' ac-010.log   → 4246   (최상위만)
$ /usr/bin/grep -c -- '--- PASS: ' ac-010.log → 6938   (서브테스트 포함)
$ /usr/bin/grep -c -- '--- SKIP: ' ac-010.log → 30
$ /usr/bin/grep -c -- '--- FAIL: ' ac-010.log → 0
$ /usr/bin/grep -c '(cached)' ac-010.log      → 0
```

**6938 / 30 / 0 은 정확히 재유도된다** — 앵커 붙인 형태로 처음 세었을 때 4246 이 나와 불일치처럼 보였으나, AC 로그 본문이 세는 대상을 `--- PASS: ` 문자열로 명시하고 있어 앵커 없는 형태가 옳다. 캐시 오염도 0이다. 레인의 수치는 성립한다.

### 2.6 작성자 겹침 — 유실 실측 (주장 5, 레인 보고 미채택·독립 재측정)

```
$ git show 37a631a70 --format='' -- <progress.md> | grep -c '^+'        → 18
$ git show 37a631a70 --format='' -- <progress.md> | grep '^-' | grep -v '^---'  → (빈 출력, rc=1)
$ git show 37a631a70 --format='' -- <progress.md> | grep -c 'E\.4'      → 0
$ git show 18bf8cc06 --format='' -- <progress.md> | grep '^-' | grep -v '^---'  → (빈 출력)
$ grep -c '^### M3 —' <progress.md>                                     → 1
$ git show --stat --format='' 37a631a70 | grep -c 'internal/'           → 0
$ git show --stat --format='' 18bf8cc06 | grep 'internal/'              → (빈 출력)
```

**유실 0 은 독립적으로 성립한다.** 레인이 보지 않은 반대 방향까지 쟀다 — sync 커밋 `18bf8cc06` 이 progress.md 에서 **삭제한 줄은 0**이므로, sync 가 m3 의 `§E.2 M3` 절을 덮어쓰고 되쓴 경우가 아니다. m3 도 내용 줄을 하나도 지우지 않았다(앞서 센 `-1` 은 `--- a/…` 헤더였다). 조용한 화해(silent reconciliation) 흔적 없음. 두 커밋 모두 `internal/` 무변경.

### 2.7 뮤턴트·도달성 증거 검토 (재실행 아님, 로그 판독)

세 뮤턴트가 각각 **서로 다른** 이름의 테스트 하나씩에 잡혔고, 세 지점이 각각 **서로 다른** panic 프로브로 도달 증명됐다 — 「도달하지 않은 뮤턴트와 진짜 생존자는 같은 `ok` 를 찍는다」를 실제로 막은 배치다.

| 로그 | 잡은 테스트 / 프로브 |
|---|---|
| `mutant-bypass.log` | FAIL `TestJudgeCodexSkillEntry_SeparatorConversion` |
| `mutant-blanket-wrap.log` | FAIL `TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative` |
| `mutant-reorder.log` | FAIL `TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion` |
| `probe-p1/p2/p3.log` | `panic: MX-PROBE-P1-CONVERSION-SITE` / `-P2-STAT-SITE` / `-P3-CLASSIFY-SITE` |

세 뮤턴트 전부 **prune 쪽**이다. doctor 쪽 뮤턴트는 없다(→ F2).

---

## 3. Baseline-attribution (이 런, 이 트리에 대한 귀속)

이 보고의 모든 수치는 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t562`, SHA `bbba4f672` 에서 **이번 런에 직접 실행**해 얻었다. 예외는 두 부류이며 각각 성격을 밝힌다:

- §2.5 의 `6938/30/0` 과 §2.7 의 뮤턴트 결과는 **커밋된 로그 파일을 판독**한 것이다(재실행 아님). §2.5 는 원본 로그에서 계수를 재유도했으므로 계수 축은 검증됐고, 실행 축은 검증하지 않았다.
- §2.3 의 「Windows 파일명은 `/` 를 금지한다」는 **인용된 전제**이며 이 런에서 측정하지 않았다.

`primary` 체크아웃과의 드리프트 위험(같은 디렉터리 두 철자)에 대한 대응: 모든 측정 배치를 `git rev-parse --show-toplevel` 로 시작해 접지했고, 상대경로만 사용했으며, `grep` 은 전부 `/usr/bin/grep` 으로 명시 호출했다. `doctor_codex.go` 의 `osStatFn` 계수는 이 트리에서 **2**(`:459`, `:861`)로 읽혔다 — primary 의 0 과 다르므로, 드리프트가 있었다면 이 값이 0으로 나왔을 것이다. 즉 이 계수가 트리 귀속의 양성 증인이다.

---

## 4. Gaps (관측하지 않은 것)

레인이 선언한 G1~G8 은 전부 정직하게 경계 지어져 있고, 어느 것도 결함을 감추고 있지 않다 — 다만 **G6 은 불가피한 것으로 서술돼 있으나 닫을 수 있다**(F1·F2). 그 외 이 감사가 남기는 간극:

- **G-A** Windows 런타임 미관측(G2 와 동일). §2.3 의 안전성 논증은 구조적이며, Windows 파일명 문자집합은 인용된 전제다.
- **G-B** 뮤턴트를 재실행하지 않았다. 로그 판독만 했으므로, 이 감사는 「그 로그가 그런 결과를 담고 있다」를 세울 뿐 「지금 다시 돌려도 같다」를 세우지 않는다.
- **G-C** `-race` 미실행. 전체 저장소 스위트 미실행(범위는 `internal/cli`).
- **G-D** SKIP 30건을 세기만 했고 귀속하지 않았다(레인의 G7 을 그대로 승계).
- **G-E** AC-CSRB-010 의 BEFORE 하한 문제(G1)는 **닫히지 않았다**. AFTER 재유도는 했으나 BEFORE 를 재수립할 방법이 이 트리에 없다. 최대 52건의 조용한 유실이 `AFTER >= BEFORE` 를 여전히 만족시킨다는 잔여 위험은 그대로다.
- **G-F** doctor 쪽 변환의 **행위** 검증을 이 감사도 수행하지 않았다(구조 판독 + prune 쪽과의 형태 동일성 확인까지). 판정은 F2 로 남긴다.
- **G-G** `internal/codexwiring` 파서 자체(선언 문자열이 어떻게 뽑히는가)는 이 감사의 범위 밖이다. 신뢰 경계의 **상류**이므로, 파서가 부여하는 제약은 검증되지 않은 채로 남는다.

---

## 5. Residual-risk (관측했음에도 여전히 틀릴 수 있는 것)

1. **doctor 쪽 되돌림은 어떤 테스트도 잡지 못한다.** darwin 에서 변환이 항등이므로 AC-CSRB-006 기준선 가드는 변환이 있든 없든 바이트 동일한 출력을 낸다. 즉 doctor 쪽 한 줄이 미래에 `e.Path` 로 되돌아가도 CI 는 초록이다. 이것이 이 카드가 남기는 가장 오래 가는 위험이다.
2. **Windows 에서의 순효과가 0일 가능성.** 일반 경로에서는 Windows 가 `/` 를 관용하므로, 이 수정이 실제로 무언가를 고치는 계열은 §2.3 이 찾은 확장길이 경로뿐일 수 있다. 수정이 **해롭지 않다**는 것은 섰지만 **필요했다**는 것은 서지 않았다.
3. **상류 파서가 어떤 문자열을 통과시키는지 모른다.** 이 감사는 `SkillEntry.Path` 에 임의 문자열이 들어온다고 가정하고 적대 입력을 넣었다. 파서가 그보다 좁게 거르고 있다면 §2.3 의 일부 행은 도달 불가이고, 넓게 통과시킨다면 §2.3 이 다 덮지 못한 계열이 있을 수 있다.
4. **홈-상대 분기의 백슬래시 구멍(F3)은 이 카드가 만들지도 고치지도 않았고, 지금도 열려 있다.**

---

## 6. Dimension Scores (default 프로필 앵커 0.25 / 0.50 / 0.75 / 1.00)

| Dimension | 앵커 | 인용한 루브릭 문구 | 근거 |
|---|---|---|---|
| **Functionality (40%)** | **0.75** | "All primary acceptance criteria pass; minor edge cases missing" | AC 10건 PASS, 카드 테스트 계열 33 PASS / 0 FAIL 재실행. 감점: doctor 쪽 변환의 엣지(분리자 주입) 미검증, AC-010 BEFORE 하한. |
| **Security (25%)** | **0.75** | "No Critical/High findings; Medium findings documented with mitigations" | 적대 입력 17×2 에서 회귀 0, 비밀·exec·삭제 표면 추가 0, 삭제 게이트 불변이 AC-CSRB-005 로 고정. Medium 3건(F1·F2·F3) 존재하며 본 보고에 기록·완화안 명시. |
| **Craft (20%)** | **0.75** | "Coverage >= 80%, minor style issues, acceptable naming" | `internal/cli` 커버리지 **81.2%** 실측 — 프로젝트 목표 85% 미달(선행 상태, 2줄 변경이 만든 것 아님). 간극 선언은 모범적. 감점: 커밋된 거짓 주석(F1). |
| **Consistency (15%)** | **1.00** | "Fully consistent with project conventions and existing patterns" | gofmt 빈 출력, lint 0 issues, vet rc=0, t540 seam 을 재창조 없이 소비, 주석 문체·SPEC 인용 관행 일치, REQ-CSRB-006 범위 핀이 커밋 9개 전수에서 성립. |

가중합 = 0.75×0.40 + 0.75×0.25 + 0.75×0.20 + 1.00×0.15 = **0.79**

**must-pass 방화벽**: Functionality(통과 기준 "All acceptance criteria PASS" — 10/10 PASS) 와 Security(통과 기준 "No Critical/High findings" — Critical 0, High 0) 두 축 모두 독립적으로 통과. 판정은 이 방화벽이 정하며 가중합이 정하지 않는다.

---

## 7. Findings (구조화 결함 목록)

| id | severity | confidence | blocking | 위치 | 내용 |
|---|---|---|---|---|---|
| **F1** | Medium | high | no | `internal/cli/codex_stale_skill_readback_test.go:6` | 커밋된 주석이 "doctor_codex.go has zero osStatFn seams" 라고 단언하나 HEAD 에서 거짓이다(`:459`, `:861` 두 곳 존재 — t563 흡수분). 이 거짓 전제가 doctor 쪽을 구조 검증으로만 두는 **근거로 쓰이고 있다**. 필요한 수정: 주석을 사실로 정정하고, G6 을 「불가피한 한계」가 아니라 「의도적 유예」로 다시 진술한다. |
| **F2** | Medium | high | no | `internal/cli/doctor_codex.go:837` (가드 부재) | doctor 쪽 변환에 되돌림을 잡는 테스트가 **하나도 없다**. 뮤턴트 3종 전부 prune 쪽이고, AC-CSRB-006 가드는 darwin 항등 때문에 변환 유무에 무감하다. 능력은 HEAD 에 이미 있다 — `doctor_codex_stale_skill_test.go:378` 의 `stubStatRecording` + `TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths` 가 seam 관측을 이미 한다. 필요한 수정: `TestJudgeCodexSkillEntry_SeparatorConversion` 과 대칭인 doctor 쪽 분리자 주입 테스트 1본. |
| **F3** | Medium | high | no | `internal/cli/doctor_codex.go:667-678` (`classifyCodexSkillPath`) — **t562 소관 아님**, prune/classifier 소유 · 선행 결함 | 분류기가 `~/` 를 백슬래시 검사보다 **먼저** 벗겨내므로, 백슬래시를 품은 홈-상대 선언은 oddly-formed 거부를 빠져나가 stat 되고 **삭제 적격이 된다**. 실측(§2.3): `~/x\SKILL.md` → CLASS=1, stat `/Users/goos/x\SKILL.md`, `Eligible=true`. 같은 백슬래시가 상대 선언에 있으면 거부된다 — 삭제 경로의 비대칭 가드. 필요한 수정: 별도 카드로 홈-상대 분기에도 백슬래시 판정을 적용할지 판단(이 카드에서 고치는 것은 REQ-CSRB-002 위반이므로 **금지**). |
| **F4** | Low | high | no | `spec.md` §A | 사용자 스토리가 "the misread that, on Windows, opens a destructive prune branch" 라고 **사실로** 단언하나, 같은 문서 §B.2 는 "On Windows this still often resolves (the OS tolerates `/`)" 라고 적는다. §A 가 자기 증거절보다 강하다. CHANGELOG 는 과장하지 않는다(구조적 추론임을 명시하고 "No Windows runtime was observed" 를 적는다) — 다만 관용 사실을 옮기지 않아 독자가 강한 쪽으로 읽는다. 필요한 수정: §A 를 §B.2 수준으로 낮추거나 관용 단서를 §A 에 함께 싣는다. |
| **F5** | Low | high | no | `internal/cli/codex_config_path.go:58-59` — **t540 소관** | `fromConfigPath` 를 "the inverse of toConfigPath" 라 적었으나 역함수가 아니다. `toConfigPath` 는 단사가 아니고, `\` 호스트에서 `fromConfigPath(toConfigPath("C:/x"))` = `C:\x` ≠ 입력이다. 의미적으로 무해하며 문구 문제. |
| **F6** | Info (긍정) | high | no | `spec.md` §B.2 | 이 감사가 찾은, 수정이 **진짜로** 값을 하는 유일한 계열: 확장길이 경로. `//?/C:/…` → `\\?\C:\…` 이며 Windows 는 `\\?\` 안의 `/` 를 정규화하지 않는다. SPEC 이 이 계열을 이름 붙이면 정당화의 가장 약한 부분이 구체적 근거로 바뀐다. |
| **F7** | Low | high | no | `internal/cli` 패키지 | 커버리지 81.2% 실측, 프로젝트 목표 85% 미달. 선행 상태이며 이 2줄 변경이 만든 것이 아니다. Craft 축에만 반영. |

### 공정 소견

| id | 내용 |
|---|---|
| **P1** | **작성자 겹침 — 유실 0, 독립 확인.** §2.6 참조. 레인이 재지 않은 반대 방향(sync 가 m3 를 덮었는가)까지 쟀고, sync 커밋의 삭제 줄 0으로 성립한다. 조용한 화해 흔적 없음. `37a631a70` 을 되돌리지 않은 판단도 지지한다 — 그 커밋은 내용 줄을 하나도 지우지 않았고 `internal/` 무변경이다. |
| **P2** | **감사 창 안에 외부 쓰기가 들어왔다.** `.moai/reports/t562/writer-collision.md` (mtime `2026-09-08 07:29:26`, 작성 주체 lane-2, untracked) 가 이 감사 시작 이후에 생겼다 — 감사 시작 시 `git status --porcelain` 은 빈 출력이었다. HEAD 는 `bbba4f672` 로 불변이고 `internal/` 은 바이트 청결을 유지했으므로 판정에는 영향이 없다. 커밋이 아니므로 「외부 커밋」 [HARD] 절은 발동하지 않으나, **감사 창의 단일 작성자 기대는 지켜지지 않았다** — 규율대로 보고한다. |

---

## 8. Recommendations

1. **F2 를 닫는다** — doctor 쪽 분리자 주입 테스트 1본. `stubStatRecording` + `configPathSeparator = '\\'` + 변환 형태 단언. 능력은 이미 트리에 있고 비용은 20줄 남짓이며, 이것이 §5-1 의 잔여 위험을 없애는 유일한 조치다.
2. **F1 을 정정한다** — 거짓 주석은 다음 독자에게 「닫을 수 없는 간극」이라고 잘못 가르친다. 정정 없이 F2 만 닫으면 주석이 새 테스트와 모순된다.
3. **F6 을 SPEC 에 싣는다** — 확장길이 경로 계열은 이 수정의 필요성을 처음으로 구체적으로 세워준다.
4. **F3 을 별도 카드로 발행한다** — 삭제 경로의 비대칭 가드이며 이 카드에서 고치는 것은 범위 위반이다.
5. **F4 를 낮춘다** — 같은 문서 안에서 §A 와 §B.2 가 다른 강도를 말하는 상태를 남기지 않는다.

---

감사자: `sync-auditor` · 2026-09-08 · 트리 `bbba4f672`
증거 파일(이 런 생성): `audit-probe.log` · `audit-vet.log` · `audit-lint.log` · `audit-test.log` · `audit-readback.log` · `audit-cover.log` (전부 `.moai/reports/t562/`)
