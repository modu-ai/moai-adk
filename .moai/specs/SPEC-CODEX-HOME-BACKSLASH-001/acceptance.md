# SPEC-CODEX-HOME-BACKSLASH-001 — 인수 기준

모든 AC 는 실행 가능한 명령과 엄격한 통과 기준을 갖는다. 시간 추정 없음.

테스트 파일: `internal/cli/codex_skills_path_shape_test.go`(신규). 시접을 덮는 하위 테스트는 `t.Parallel()` 없이 `t.Cleanup` 복원(REQ-CHB-007).

---

## AC-CHB-001 — 두 팔 대칭 (구속력 있는 판별식)

_maps REQ-CHB-001, REQ-CHB-002_

**Given** 같은 백슬래시를 품은 두 선언 `~/x\SKILL.md`(홈-상대 형태)와 `x\SKILL.md`(상대 형태)가 주어지고,
**When** `classifyCodexSkillPath`를 두 선언에 각각 적용하면,
**Then** 두 반환값은 **서로 같아야** 하고, 그 공통 값이 `codexPathOddlyFormed` 여야 한다.

- 테스트는 두 단언을 **모두** 수행한다: (a) `got(homeArm) == got(relArm)`, (b) `got(homeArm) == codexPathOddlyFormed`. (a) 가 이 AC의 본체다 — 결함이 비대칭이므로 대칭이 검사 대상이다.
- 명령: `go test ./internal/cli/ -run 'TestCodexSkillPathBackslashSymmetry' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT=0` 이고 출력에 `--- PASS: TestCodexSkillPathBackslashSymmetry` 가 있으며, 하위 테스트 이름에 두 팔이 모두 등장한다.

---

## AC-CHB-002 — 뮤턴트 (판별식이 실제로 무언가를 잡는다)

_maps REQ-CHB-001, REQ-CHB-002_

**Given** `classifyCodexSkillPath`의 분기 순서를 옛 순서(`IsAbs → "~"/"~/" 껍질 → 백슬래시 → relative`)로 일시 복원하고,
**When** AC-CHB-001 의 명령을 다시 실행하면,
**Then** 그 테스트는 **실패**해야 하며, 실패 메시지가 홈-상대 팔의 관측값을 상수 이름 `codexPathHomeRelative` 로 지목해야 한다.

- **[HARD] Then 과 통과 기준이 같은 것을 요구한다.** 종전에는 Then 만 상수 이름을 요구하고 통과 기준은 `--- FAIL:` 줄만 요구해, 기준의 문자만 지키면 Then 이 미충족인 채로 통과하는 테스트를 쓸 수 있었다. `codexSkillPathShape` 는 `String()` 메서드가 없는 맨 `int` 이므로(`grep -n "codexSkillPathShape)" internal/cli/*.go` → 매치 0건) `%v` 는 상수 이름이 아니라 `1` 을 찍는다. 따라서 **테스트가 shape → 이름 매핑을 직접 들고 실패 메시지에 이름을 적어야 하며**, 그 요구를 아래 통과 기준에 넣는다. 이 매핑은 테스트 파일 안의 표이며 새 production 시접도 새 shape 상수도 아니다(§plan.md D 제약 위반 아님).
- 명령: (뮤턴트 주입 후) `go test ./internal/cli/ -run 'TestCodexSkillPathBackslashSymmetry' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT != 0` 이고, 출력에 `--- FAIL: TestCodexSkillPathBackslashSymmetry` 가 있으며, **출력에 문자열 `codexPathHomeRelative` 가 1회 이상 등장한다**(홈-상대 팔의 관측값이 상수 이름으로 지목됐다는 증거). 출력을 `.moai/reports/t571/mutant-old-ordering.log` 로 남기고 소스를 원상 복구한 뒤, AC-CHB-001 을 다시 실행해 `EXIT=0` 임을 보인다.
- **[HARD]** 뮤턴트가 RED 를 세우지 못하면 AC-CHB-001 은 공허한 초록으로 판정한다.

---

## AC-CHB-003 — 보존 칸 (바뀌는 칸은 하나뿐이다)

_maps REQ-CHB-003, REQ-CHB-005_

**Given** 아래 네 선언이 주어지고,
**When** `classifyCodexSkillPath`를 각각 적용하면,
**Then** 각 반환값이 표의 기대값과 일치해야 한다.

| 선언 | 기대 shape | 무엇을 지키는가 |
|---|---|---|
| `/tmp/a\b` (POSIX 호스트의 절대 경로 + 백슬래시) | `codexPathAbsolute` | `IsAbs` 가 백슬래시 검사보다 **먼저** 남는다 |
| `~someone/SKILL.md` | `codexPathOddlyFormed` | `~user` 형태 불변 |
| `rel/SKILL.md` | `codexPathRelative` | 백슬래시 없는 상대 불변 |
| `~/ok/SKILL.md` | `codexPathHomeRelative` | 깨끗한 홈-상대 불변 (§spec.md 1.2 양성 대조군) |

- 첫 행이 이 AC의 핵심이다. `IsAbs` 위로 백슬래시 검사가 잘못 올라가면 이 칸만 붉어진다 — POSIX 호스트에서 「`IsAbs` 우선」을 기계적으로 확인할 수 있는 유일한 판별식이다.
- 명령: `go test ./internal/cli/ -run 'TestCodexSkillPathPreservedShapes' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT=0`, 네 하위 테스트 전부 `PASS`, 건너뛴(`SKIP`) 하위 테스트 0건.

---

## AC-CHB-004 — 소비자 판정과 stat 부재

_maps REQ-CHB-004_

**Given** 선언 `~/x\SKILL.md` 를 담은 `codexwiring.SkillEntry`(`FirstUnrecognizedLine = -1`)가 주어지고, `osStatFn` 이 호출 횟수를 세는 함수로 덮여 있고,
**When** `judgeCodexSkillEntry` 를 호출하면,
**Then** `Eligible == false` 이고, `SkipReason` 이 `oddly-formed` 를 포함하며, `osStatFn` 호출 횟수가 **정확히 0** 이어야 한다.

- 호출 0회가 이 AC의 두 번째 축이다. `Eligible=false` 만 보면 「stat 은 했지만 결과가 달랐다」와 구별되지 않는다.
- **[HARD] 계수기의 양성 대조군을 같은 테스트 안에 둔다.** 아무 데도 연결되지 않은 계수기와 진짜 0 은 똑같이 `0` 을 찍는다. 그러므로 이 테스트는 두 번째 하위 케이스를 갖는다: 같은 방식으로 덮은 `osStatFn` 아래에서 **깨끗한 선언 `~/ok/SKILL.md`** 를 `judgeCodexSkillEntry` 에 통과시키고, 그때 계수가 **정확히 1** 임을 단언한다. 대조군이 1 을 세우지 못하면 주 케이스의 0 은 측정이 아니라 부재이며, 이 AC 는 미충족으로 판정한다.
- 두 하위 케이스는 계수기를 공유하지 않는다(각 하위 케이스가 자기 계수기를 새로 설치하고 `t.Cleanup` 으로 복원한다). 공유하면 순서에 따라 0 이 1 로 오염된다.
- **이 AC 는 RED-now 칸을 이미 갖고 있다 — 물려받은 것이라 적어 둔다.** AC-CHB-002 의 옛-순서 뮤턴트가 주입된 창 안에서는 `~/x\SKILL.md` 가 home-relative 로 분류돼 확장되고 stat 이 걸리므로, 이 AC 의 「계수 0」 단언도 **같은 창에서 함께 붉어진다.** 별도 뮤턴트가 필요 없다는 뜻이고, 명시하지 않으면 이 0 이 근거 없는 단언처럼 읽힌다. AC-CHB-002 를 실행할 때 이 테스트도 함께 돌려 그 적색을 같은 로그에 남긴다.
- `osStatFn` 오버라이드이므로 이 테스트는 `t.Parallel()` 없이 `t.Cleanup` 으로 복원한다.
- 명령: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT=0`, 하위 테스트 **2건**이 각각 `--- PASS` 로 출현(거절 케이스 + 양성 대조군), `SKIP` 0건, 네 단언 모두 수행(거절 케이스의 세 단언 + 대조군의 계수 `== 1`).

---

## AC-CHB-005 — 읽기 쪽과 쓰기 쪽의 대칭

_maps REQ-CHB-002_

**Given** 선언 `~/x\SKILL.md` 가 주어지고, 호스트 구분자가 `'/'` 이고(`configPathSeparator = '/'`), 최소한의 `~/.codex/config.toml` 내용을 담은 바이트 fixture 가 주어지고,
**When** 쓰기 쪽 **production 진입점** `upsertCodexSkillDisable(content []byte, skillPath string)`(`internal/cli/codex_skills_disable.go:231`)을 그 fixture 와 그 선언으로 호출하고, 같은 선언에 읽기 쪽 `classifyCodexSkillPath` 를 적용하면,
**Then** 양쪽 모두 그 선언을 거절해야 한다 — 쓰기 쪽은 반환된 `codexSkillDisableVerdict` 의 `Action` 이 `codexSkillDisableSkipped` 이고 `Reason` 이 비어 있지 않으며 반환 content 가 입력과 동일(무변경), 읽기 쪽은 `codexPathOddlyFormed`.

- **[HARD] 판별식을 테스트 안에 다시 쓰지 않는다.** 종전 문안은 거절을 *표현식* `strings.ContainsAny(skillPath, "\"\\\n\r")` 으로 지목했다. 그 문안을 만족시키는 가장 쉬운 테스트는 같은 표현식을 테스트 파일에 인라인으로 복사해 사본을 사본과 대조하는 것이고, 그런 테스트는 **production 판별식을 통째로 지워도 계속 초록이다.** 따라서 이 AC 는 표현식이 아니라 **그것을 평가하는 함수**를 부른다. 테스트 안에서 `strings.ContainsAny`(혹은 그 판별식의 어떤 재구현)를 호출하는 것은 **금지**이며, 그런 호출이 있으면 이 AC 는 미충족이다.
- 실측으로 확인된 진입점: `func upsertCodexSkillDisable(content []byte, skillPath string) ([]byte, codexSkillDisableVerdict)` — 같은 패키지(`internal/cli`)라 테스트에서 직접 호출 가능하다. 거절 시 `skip(...)` 헬퍼가 `Action: codexSkillDisableSkipped` 와 사유 문자열을 담은 verdict 를 돌려주며 content 를 그대로 반환한다.
- 이 AC 가 잡는 뮤턴트: `codex_skills_disable.go` 의 `if strings.ContainsAny(skillPath, "\"\\\n\r") { return skip(...) }` 블록을 삭제하면 이 테스트가 붉어져야 한다. (선택 검증 — 실행 시 로그를 `.moai/reports/t571/mutant-write-guard-removed.log` 로 남기고 원상 복구.)
- `configPathSeparator` 를 덮으므로 `t.Parallel()` 금지 + `t.Cleanup` 복원.
- 명령: `go test ./internal/cli/ -run 'TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT=0`, `--- PASS: TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite` 출현.
  - **종전 세 번째 조항(「테스트 함수 본문에 `ContainsAny` 문자열 0회」)은 삭제했다.** 명령이 붙어 있지 않아 판정 불가였고, 함수 단위 판독을 요구해 사람이 읽어야 했으며, 무엇보다 **철자만 바꾸면 빠져나간다**(`strings.IndexAny`, 룬 루프 등). 완전성 가드가 아니라 장식이었다. 재구현 금지는 위 [HARD] 조항이 규범으로 세우고, 그것을 실제로 지키는 것은 **`upsertCodexSkillDisable` 호출**과 바로 위의 쓰기-가드 제거 뮤턴트다 — 술어를 지웠을 때 붉어지는지가 유일하게 기계적인 증거다.

---

## AC-CHB-006 — 회귀 범위 (고정된 기준선과의 정확한 일치)

_maps REQ-CHB-005_

**Given** M2 의 순서 재배치가 적용된 트리에서, 그리고 **흡수 후·M2 적용 전에 측정·고정된** 현재 유효한 SKIP 기준선 `.moai/reports/t571/skip-baseline-post-absorb.log`(`SKIP_COUNT=30`, `MEASURED_ON_HEAD=768306d2779aae7bce52ac6a78ec0839a498a203`, `EXIT=0`)가 주어지고,
**When** 같은 명령 `go test ./internal/cli/... -count=1 -timeout 600s -v` 를 실행하고 `grep -c '^ *--- SKIP: '` 로 SKIP 을 세면,
**Then** 종료 코드가 0 이고 **SKIP 계수가 정확히 30** 이어야 한다.

- **[HARD] 왜 문안을 바꿨는가.** 종전 문안은 「새로 `SKIP` 으로 바뀐 기존 테스트 0건」이었다. 그것은 **델타**인데 비교 대상이 어디에도 없어서, 아무 수나 기록하고 「기록되지 않은 수와 같다」고 단언하면 만족되는 판정 불가 기준이었다. 이제 기준선은 **M2 이전에 실제로 측정되어** 위 파일에 SHA 와 함께 고정돼 있고, 이 AC 는 그 수와의 **정확한 일치**를 요구한다.
- 계수 표현식의 `^ *` 는 필수다 — `-v` 출력에서 `--- SKIP:` 줄은 들여쓰기돼 있어 `^--- SKIP:` 로 앵커하면 항상 0 이 나오고, 그 0 은 측정이 아니라 앵커 실패다. 기준선과 **같은 표현식**으로 세지 않으면 비교 자체가 성립하지 않는다.
- 하위 테스트를 개별로 센다— 기준선 30건 중 **4건**이 하위 테스트다(실측: 기준선 로그의 `^--- SKIP: ` 행 중 `/` 를 포함하는 것 4건). 세는 단위가 다르면 두 수는 비교 가능하지 않다.
- 불일치가 나오면 **차이 나는 이름을 열거해서** 보고한다. 30 개 중 여럿은 환경 술어(live 백엔드·TTY·크로스컴파일)로 건너뛰므로, 불일치가 곧 회귀는 아니다 — **이름이 무엇인지가 판정**이고, 계수만 보고 결론 내리지 않는다.
- **[HARD] 이름을 대조하기 전에 경과 시간 접미사를 벗긴다.** `-v` 출력의 `--- SKIP:` 줄은 자기 실행 시간을 달고 나오고(`(0.01s)`), 그 값은 실행마다 달라진다. 접미사를 남긴 채 `comm -3` 을 돌리면 **매 실행 가짜 델타**가 난다 — 실측 사례: 흡수 전후 대조에서 `TestMCPServer_StdioRoundTripSubprocess` 가 `(0.01s)` 대 `(0.02s)` 로 어긋나 1건 차이로 보고됐고, 접미사를 벗기니 두 30개 집합이 동일했다. 항상 나는 가짜 델타는 읽는 사람에게 무시하는 습관을 가르치고, 진짜 델타는 바로 그 습관을 타고 통과한다. 대조식:
  ```
  grep '^ *--- SKIP: ' <출력> | sed 's/^ *//' | sed -E 's/ \([0-9.]+s\)$//' | sort   # 양쪽 동일하게
  comm -3 <기준선목록> <이번목록>                                                      # 아무것도 안 나와야 정상
  ```
- **현재 유효한 기준선은 `.moai/reports/t571/skip-baseline-post-absorb.log`** 다(`SKIP_COUNT=30`, `MEASURED_ON_HEAD=768306d2779aae7bce52ac6a78ec0839a498a203`, `EXIT=0`, 하위 테스트 4건, `FAIL` 0줄). 흡수 전 `ee194493f`/30 고정은 이제 이력이다 — 두 수가 우연히 같지만, **같다는 것은 실측 결과이지 흡수가 안전하다는 보장이 아니다.**
- 통과 기준: `EXIT=0` **그리고** `SKIP` 계수 `== 30` **그리고** 계수가 30 이 아닐 경우 차이 나는 이름이 전부 열거되고 각각 환경 사유로 귀속됨.
- **[HARD] 재고정 규칙 — 기준선은 흡수를 견디지 못한다.** 이 카드는 병합 창에서 `origin/develop` 을 흡수하고 병합 트리에서 재측정할 의무가 있다(plan.md §B). 그런데 `SKIP_COUNT=30` 은 흡수 **이전** `ee194493f` 에서 잰 값이라, 흡수가 이 패키지를 건드리는 순간 무효가 된다 — 가정이 아니라 실제로 그렇다: 흡수해 들어온 t577 이 `internal/cli/todo.go` 를, 즉 이 AC 가 세는 바로 그 패키지를 고쳤다. 기준선 파일 자신의 단서는 **호스트·환경** 드리프트만 다루고 **트리** 드리프트에는 침묵한다.
  그러므로: `origin/develop` 을 흡수할 때마다 **같은 명령과 같은 `^ *` 표현식으로 병합 트리에서 다시 재고**, 새 HEAD 를 명시한 두 번째 고정 기록을 `.moai/reports/t571/skip-baseline-post-absorb.log` 로 쓰고, 위의 정확 일치 비교는 **그 기록**을 상대로 한다. `ee194493f`/30 고정은 **흡수 이전 실행에 한해서만** 유효하다. 아래 「이름이 무엇인지가 판정」 조항은 그대로 유지한다 — 어느 쪽 기록을 쓰든 델타를 진단 가능하게 만드는 것이 그 조항이다.
- 출력은 `.moai/reports/t571/` 아래에 남긴다.
- 전체 스위트(`./...`)는 로컬에서 돌리지 않는다 — CI 판정 몫.

---

## AC-CHB-007 — 확장 결과에는 검사를 걸지 않는다 (Windows 파괴 방지)

_maps REQ-CHB-006_

**Given** `codexUserHomeDir` 가 백슬래시를 품은 홈(`C:\Users\x` — Windows 확장 결과의 형태)을 반환하도록 덮여 있고, 선언이 `~/ok/SKILL.md`(백슬래시 없음)이고,
**When** `judgeCodexSkillEntry` 를 호출하면,
**Then** 그 엔트리는 oddly-formed 로 거절되어서는 **안 되고**, `osStatFn` 이 정확히 1회, 그리고 **관측된 인자가 정확히 `C:\Users\x/ok/SKILL.md`** 여야 한다.

- **[HARD] 계수만으로는 아무것도 갈리지 않는다.** 이 AC의 판별력 전부가 `codexUserHomeDir` → `C:\Users\x` 스텁에 있는데, 「stat 계수 1」은 스텁이 **먹지 않아 진짜 홈이 쓰였을 때도 똑같이** 나온다. 그러므로 판정 대상은 계수가 아니라 **관측된 stat 인자**다. 기대값이 `C:\Users\x/ok/SKILL.md` 인 근거: `'/'`-구분자 호스트에서 `filepath.Join` 은 백슬래시를 평범한 바이트로 두므로 확장 결과가 백슬래시를 품은 채 남는다.
- **아직 관측되지 않은 기대 — 그렇게 표시한다.** 「검사를 확장 뒤로 옮기면 이 칸이 붉어진다」는 **추론이지 관측이 아니다.** 그 이동은 §spec.md 3.2 가 **기각한** 설계이므로 어떤 절차도 그 뮤턴트를 주입하지 않으며, 따라서 이 AC 는 M2 전에도 후에도 AC-CHB-002·009 두 뮤턴트 아래에서도 초록이다. 관측으로 승격하려면 세 번째 뮤턴트(검사를 `expandCodexHomeRelativePath` 아래로 이동 → 이 AC 적색 확인 → 로그 → 원복 → 재실행 초록)가 필요하고, 이 카드는 그것을 **두지 않는다.** 위 인자 단언이 대신 세우는 것은 「스텁이 실제로 먹었다」 하나뿐이며, 그 이상을 주장하지 않는다.
- `codexUserHomeDir` + `osStatFn` 오버라이드이므로 `t.Parallel()` 금지 + `t.Cleanup` 복원.
- 명령: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntryBackslashHomeStillExpands' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT=0`, 출력에 `--- PASS: TestJudgeCodexSkillEntryBackslashHomeStillExpands` 가 있고, stat 호출 계수 `== 1` **그리고** 기록된 인자가 `C:\Users\x/ok/SKILL.md` 와 문자열 동일, `SkipReason` 이 비어 있음.

---

## AC-CHB-008 — 시접 오버라이드 테스트의 병렬 금지 규율

_maps REQ-CHB-007_

**Given** 이 카드가 추가한 테스트 파일 `internal/cli/codex_skills_path_shape_test.go` 가 주어지고,
**When** 그 파일 전체에서 `t.Parallel()` 과 `t.Cleanup(` 의 출현을 세면,
**Then** `t.Parallel()` 이 **0건**이고, 시접(`osStatFn`, `codexUserHomeDir`, `configPathSeparator`)을 재대입하는 함수마다 `t.Cleanup(` 이 **1건 이상**이어야 한다.

- **[HARD] 왜 파일 전체 금지로 좁혔는가.** 종전 문안은 명령이 파일 전체 `grep -c` 인데 통과 기준은 함수 단위였고, 기준이 자기 명령으로 판정되지 않아 **사람이 파일을 읽어야만 판가름 나는 AC** 였다. 이 파일은 이 카드가 새로 만드는 것이므로 파일 전체에서 병렬화를 포기하는 비용이 0 이다. 그래서 요구를 좁혀 명령이 그대로 판정이 되게 한다.
- 명령 1(금지): `grep -c '[^A-Za-z]t\.Parallel()' internal/cli/codex_skills_path_shape_test.go`
- 명령 2(존재): `grep -c '[^A-Za-z]t\.Cleanup(' internal/cli/codex_skills_path_shape_test.go`
- **양성 대조군(필수).** 명령 1 이 0 을 낸 것이 「금지가 지켜졌다」인지 「표현식이 아무것도 못 잡는다」인지 가른다: 시접 함수 한 곳에 `t.Parallel()` 을 일시 삽입해 명령 1 이 `1` 이상을 내는 것을 보이고, 그 출력을 `.moai/reports/t571/ac008-positive-control.log` 에 남긴 뒤 원복하고 다시 0 을 확인한다. 대조군 없는 0 은 부재 증거가 아니다.
- 통과 기준: 명령 1 의 출력이 정확히 `0` **그리고** 그 0 에 대한 위 양성 대조군 로그가 존재 **그리고** 명령 2 의 출력이 시접 재대입 함수 수 이상 **그리고** 파일이 존재(`test -f` `EXIT=0` — 파일이 없어도 `grep -c` 는 0 을 내므로 존재 확인이 없으면 이 AC 는 파일 부재로 만족된다).
- 이 패키지의 기존 규율과 같다: `codex_skills_prune_test.go:10`, `codex_config_path_test.go:14`.

---

## AC-CHB-009 — 두 번째 뮤턴트: `IsAbs` 우선을 실제로 관측한다

_maps REQ-CHB-003_

**Given** M2 가 적용된 트리에서 `classifyCodexSkillPath` 의 `strings.ContainsRune(p, '\\')` 검사를 `filepath.IsAbs(p)` 검사 **위로** 일시 이동하고(즉 백슬래시 분기가 첫 번째가 되고),
**When** `go test ./internal/cli/ -run 'TestCodexSkillPathPreservedShapes' -count=1 -v -timeout 600s` 를 실행하면,
**Then** 그 테스트는 **실패**해야 하며, 실패가 `/tmp/a\b` 칸에서 나야 한다(기대 `codexPathAbsolute`, 관측 `codexPathOddlyFormed`).

- **왜 이 AC 가 있는가.** AC-CHB-003 의 첫 행(`/tmp/a\b → codexPathAbsolute`)은 §spec.md 3.3 불변식 1 의 유일한 판별식이라고 선언돼 있지만, 이 SPEC이 예정한 어떤 단계에서도 붉어지지 않는다 — 변경 전 초록, M2 이후 초록, AC-CHB-002 의 뮤턴트(옛 순서도 `IsAbs` 를 첫 번째로 둔다) 아래에서도 초록이다. 실패가 한 번도 관측되지 않은 검사는 그 자리에 있다는 것만 알 뿐 무엇을 잡는지는 모른다. 이 AC 가 그 칸을 **한 번 붉게** 만들어 채택 근거를 만든다.
- 실패 메시지는 AC-CHB-002 와 같은 규율을 따른다 — shape 을 상수 이름으로 지목한다(`codexPathAbsolute` / `codexPathOddlyFormed`).
- 공유 트리에서의 고의 훼손이므로 AC-CHB-002 의 뮤턴트와 **같은 배타 창** 안에서 수행하고, 주입 **직전** 트리 상태(`git status --short`, `git rev-parse --short HEAD`)를 재확인한다. 두 뮤턴트를 동시에 주입하지 않는다 — 하나씩 주입·복구한다.
- 명령: (뮤턴트 주입 후) `go test ./internal/cli/ -run 'TestCodexSkillPathPreservedShapes' -count=1 -v -timeout 600s`
- 통과 기준: `EXIT != 0` 이고, 출력에 **그 칸의 하위 테스트를 지목한 실패 줄** `--- FAIL: TestCodexSkillPathPreservedShapes/posix_absolute_with_backslash` 가 등장한다(하위 테스트 이름은 `/tmp/a\b` 칸의 것이며 구현이 이 이름을 쓴다). 그 실패 줄의 메시지에 `codexPathAbsolute`(기대)와 `codexPathOddlyFormed`(관측)가 상수 이름으로 함께 등장한다.
  - **왜 「출력 어딘가에 `/tmp/a\b` 가 있으면 통과」로 두지 않는가.** `-v` 실행에서 하위 테스트 이름은 **결과와 무관하게** 찍힌다. 그러므로 그 형태의 기준은 「그 칸이 붉어졌다」가 아니라 「실행이 있었다」로 만족돼 버린다. 판정은 반드시 `--- FAIL:` 줄이 그 칸을 지목하는지로 읽는다. 출력을 `.moai/reports/t571/mutant-isabs-hoist.log` 로 남기고 **소스를 원상 복구**한 뒤 같은 명령을 다시 실행해 `EXIT=0` 임을 보인다.
- **[HARD]** 이 뮤턴트가 RED 를 세우지 못하면 AC-CHB-003 의 첫 행은 공허한 초록이고, §spec.md 3.3 불변식 1 은 지켜지지 않은 채 초록이었던 것으로 판정한다.

---

## Definition of Done

- [ ] AC-CHB-001 ~ 009 전부 통과, 각 명령의 종료 코드와 출력이 증거로 인용됨
- [ ] `moai spec lint .moai/specs/SPEC-CODEX-HOME-BACKSLASH-001/spec.md` 이 `EXIT=0`. **주의 — 이 초록은 REQ↔AC 커버리지를 말하지 않는다**: 린터는 넘긴 `spec.md` 만 파싱하고, 이 카드의 AC 는 분리된 `acceptance.md` 에 있어 린터의 시야 밖이다. 커버리지는 아래 줄의 손 판정이 세운다.
- [ ] 일곱 REQ 전부 최소 1개의 AC 에 매핑됨. 판정식은 **줄 시작 앵커**를 쓴다 — `grep -c '^_maps ' acceptance.md` 가 `grep -c '^## AC-CHB-' acceptance.md` 와 같고, `grep -o '^_maps .*' acceptance.md` 에서 뽑은 REQ 집합이 `spec.md` §2 의 일곱 개를 전부 덮는다. **앵커 없는 계수식을 쓰지 않는다**: 매핑 토큰을 본문에서 언급하는 줄(이 줄을 포함해)이 계수에 섞여 들어가 AC 수보다 큰 값이 나오고, 그 초과분은 커버리지가 아니라 자기 자신이다 (§progress.md 의 clarification 마커 계수와 같은 함정)
- [ ] AC-CHB-002 와 AC-CHB-009 의 뮤턴트 로그 2개가 트리에 남고 소스는 각각 원상 복구됨
- [ ] `classifyCodexSkillPath` 의 doc 주석이 **새 순서**를 서술함 (옛 순서를 주장하는 문장 0건)
- [ ] 새 shape 상수 0개, 새 시접 0개
- [ ] 시접 오버라이드 테스트 전부 비-parallel + `t.Cleanup` 복원
