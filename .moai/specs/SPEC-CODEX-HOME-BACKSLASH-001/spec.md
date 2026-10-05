---
id: SPEC-CODEX-HOME-BACKSLASH-001
title: "codex 삭제 경로의 비대칭 가드 — 백슬래시가 홈-상대 선언에서만 거절을 빠져나간다"
version: "0.1.0"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "security, path-classification, codex, prune, internal/cli, t571"
tier: S
depends_on: [SPEC-CODEX-SKILL-PATH-SLASH-001]
---

# SPEC-CODEX-HOME-BACKSLASH-001 — 홈-상대 선언의 백슬래시 우회

## HISTORY

### 2026-09-08 — 최초 작성 (카드 t571, 실측 재현 기준)

카드 `t571`에서 열렸다. 워크트리 `.claude/worktrees/t571`, 브랜치 `WT-codex-home-backslash`, HEAD `ee194493f`. 재현 산출물은 `.moai/reports/t571/repro-asymmetry.log`(`EXIT=0`)이며, 이 SPEC이 인용하는 수치·판정은 전부 그 로그에서 가져온 것이다 — 재유도하지 않았다.

선행 의존은 `SPEC-CODEX-SKILL-PATH-SLASH-001`(카드 t540)이다. **t540은 로컬 `develop` `ee194493f`에 착지했으나 `origin/develop`(`3ac58b5a1`)에는 아직 없다.** 이 워크트리 브랜치는 `origin/develop`에서 잘렸고 로컬 develop을 이미 흡수했으므로, 여기서 관측되는 분류기는 t540 이후 상태다. 이 사실은 서술이며 이 카드에서 재검증하지 않는다.

#### 2026-09-08 — `depends_on` 처분: 로그 남기는 override

위 문단은 **브랜치 착지 축**을 말한다. 별개의 축이 하나 더 있고, 그것이 run-phase 게이트를 건드린다 — **선행 SPEC의 `status` 필드 값**이다. 실측:

```
$ grep -n '^status:' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/spec.md
5:status: implemented
```

`spec-workflow.md` § Depends_on Pre-flight Check 의 충족 술어는 **엄격히 `completed`** 이며 부분 점수가 없다 — 나머지 7개 status 값은 전부 미충족으로 읽힌다. 따라서 `implemented` 는 **미충족**이고, `/moai run SPEC-CODEX-HOME-BACKSLASH-001` 은 Phase 1 사전 점검에서 3-옵션 게이트를 만난다. 이 카드는 그 게이트를 예고 없이 만나지 않도록 여기에 미리 적어 둔다.

**처분 — 대기하지 않고 override 를 택하되, 기록을 남긴다.** 근거: 이 카드가 의존하는 것은 선행 SPEC의 *행정 상태*가 아니라 *코드*이고, 그 코드는 이 트리에 이미 있다(이 워크트리가 로컬 develop 을 흡수했고, §1.2 재현 로그가 t540 이후 분류기를 관측한다). 미충족인 것은 status 필드 하나뿐이다. run-phase 진입 시 그 게이트에서 override 를 선택하고, 미충족 의존 ID(`SPEC-CODEX-SKILL-PATH-SLASH-001`)와 이 근거를 `.moai/logs/depends-on-override.log` 에 남긴다. **이 로그 기록은 run-phase 진입의 전제이며, 기록 없이 진행하지 않는다.**

---

## 1. 컨텍스트 — 결함의 실체

### 1.1 결함 자리

`internal/cli/doctor_codex.go`의 `classifyCodexSkillPath`(정의 시작 `:667`). 오늘의 순서는 네 단계다.

```
IsAbs → ("~" | "~/" 접두)                → home-relative
      → ("~" 접두 | ContainsRune '\\')   → oddly-formed
      → relative
```

`~/` 껍질이 **백슬래시 검사보다 먼저** 벗겨진다. 그래서 백슬래시를 품은 홈-상대 선언은 oddly-formed 거절에 도달하지 못하고 home-relative로 분류되어 확장·stat 을 거쳐 삭제 판정까지 간다. 같은 백슬래시가 **상대 선언**에 있으면 정확히 거절된다. **이 비대칭이 결함이다** — 백슬래시 자체가 아니라, 같은 문자가 선언 형태에 따라 다른 판정을 받는다는 사실이.

### 1.2 재현 (인용, 재유도 아님)

`.moai/reports/t571/repro-asymmetry.log`, `EXIT=0`:

```
path="~/x\\SKILL.md"    shape=1 eligible=true  skip=""                                   stated=[/Users/goos/x\SKILL.md]
path="x\\SKILL.md"      shape=3 eligible=false skip="oddly-formed path — not resolvable here" stated=[]
path="~/ok/SKILL.md"    shape=1 eligible=true  skip=""                                   stated=[/Users/goos/ok/SKILL.md]
```

`shape=1`은 `codexPathHomeRelative`, `shape=3`은 `codexPathOddlyFormed`다. 세 번째 줄은 **양성 대조군**이다 — 깨끗한 홈-상대 선언이 `eligible=true`를 받는 것은 정상이므로, 첫 줄의 `true`가 「분류기가 안 돌았다」가 아니라 「돌아서 통과시켰다」임을 이 줄이 세운다.

### 1.3 소비자 두 곳

| 소비자 | 자리 | 판정이 하는 일 |
|---|---|---|
| `judgeCodexSkillEntry` | `internal/cli/codex_skills_prune.go:76` | `Eligible=true` → `pruneCodexSkillEntries`가 그 엔트리의 **줄 범위를 `~/.codex/config.toml`에서 제거** |
| `codexStaleSkillFinding` | `internal/cli/doctor_codex.go:831` | 읽기 전용 stale 계수 |

### 1.4 피해의 정확한 크기 — 과장하지 않는다

`Eligible`이 지우는 것은 **설정 파일의 줄**이다. 손으로 쓴 `[[skills.config]]` 등록이 `~/.codex/config.toml`에서 사라진다. `statPath`가 가리키는 **파일은 건드리지 않는다.** 사용자 데이터 파일 삭제가 아니라 등록 소실이며, 그것만으로 충분히 결함이다 — `judgeCodexSkillEntry` 위의 `@MX:WARN`이 이미 「여기서의 잘못된 eligible 은 사용자가 손으로 쓴 등록을 지운다」고 적고 있다.

부수적으로 이 경로는 존재하지 않는 이름(`/Users/goos/x\SKILL.md`)에 `os.Stat`을 한 번 건다. 정보 유출은 관측하지 않았고 주장하지 않는다.

### 1.5 쓰기 쪽 형제 가드 — 대칭의 기준선

`internal/cli/codex_skills_disable.go:261`는 `'/'`-구분자 호스트에서 **백슬래시를 품은 경로를 무조건 거절**한다(`strings.ContainsAny(skillPath, "\"\\\n\r")`). 그 자리의 주석은 거절 이유를 명시한다 — escape 된 형태를 쓰면 codex 는 맞게 읽고 moai 독자는 다른 경로로 읽으며, **「prune 이 그것을 absent 로 분류해 지운다」**. 즉 쓰기 쪽은 이미 이 결함을 알고 방어하고 있고, 읽기/prune 쪽만 비어 있다. 두 쪽을 대칭으로 만드는 것이 이 SPEC의 목표다.

---

## 2. 요구 (GEARS)

- **REQ-CHB-001** (Ubiquitous) — The path classifier shall assign the same shape to two declarations that differ only in whether the path is home-relative, when both carry a backslash.
- **REQ-CHB-002** (Event-driven) — **When** a declared skill path that is not absolute carries a `\` rune, the path classifier shall classify it as oddly-formed, regardless of a `~/` prefix.
- **REQ-CHB-003** (Ubiquitous) — The path classifier shall evaluate `filepath.IsAbs` before the backslash check, so a host-native absolute path containing a backslash stays absolute.
- **REQ-CHB-004** (Event-driven) — **When** the prune verb judges an entry the classifier reported as oddly-formed, the prune verb shall return a non-eligible verdict and shall perform no stat call for that entry.
- **REQ-CHB-005** (Unwanted) — The classifier shall not change the shape it assigns to a `~user` declaration, a plain relative declaration, or a backslash-free home-relative declaration.
- **REQ-CHB-006** (Unwanted) — The classifier shall not run the backslash check against an expanded home directory.
- **REQ-CHB-007** (Ubiquitous) — The test file this card adds (`internal/cli/codex_skills_path_shape_test.go`) shall contain no `t.Parallel()` call, and every test in it that overrides a package-level seam (`osStatFn`, `codexUserHomeDir`, `configPathSeparator`) shall restore that seam through `t.Cleanup`.
  - Narrowed from a per-function State-driven form to a whole-file Ubiquitous one, because the per-function form could not be decided by any command the AC prescribed and required a human to adjudicate by reading. The file is new, so forgoing parallelism across all of it costs nothing. AC-CHB-008 now decides this mechanically.

---

## 3. 채택한 접근과 기각한 대안

### 3.1 채택 — 백슬래시 검사를 `~/` 껍질 벗기기 **위**, `IsAbs` **아래**로 옮긴다

```
IsAbs → ContainsRune('\\') → ("~" | "~/") → "~" 접두 → relative
```

바뀌는 칸은 **홈-상대 + 백슬래시** 하나뿐이다. 다른 세 칸은 이미 같은 값에 도달한다(§1.2의 상대-백슬래시, 그리고 `~user`·plain relative·깨끗한 홈-상대).

### 3.2 기각 — 홈 확장 **뒤에** 다시 검사한다

Windows 를 깨뜨린다. Windows 에서 확장된 홈은 정당하게 `C:\Users\x`이므로, 확장 결과에 백슬래시 검사를 걸면 모든 Windows 홈-상대 선언이 oddly-formed 가 된다. 검사는 **선언**에 걸어야 하며 **확장 결과**에 걸어서는 안 된다(REQ-CHB-006).

### 3.3 보존해야 하는 불변식 — 명시

1. **`IsAbs`가 첫 번째로 남는다.** Windows 네이티브 `C:\...` 절대 경로는 계속 absolute 여야 한다. POSIX 호스트에서 이 순서를 기계적으로 확인할 수 있는 판별식은 **백슬래시를 품은 POSIX 절대 경로**(예: `/tmp/a\b`)다 — `IsAbs`가 참이므로 재배치 후에도 absolute 로 남아야 한다.

    이 칸이 「백슬래시 검사가 `IsAbs` 위로 올라가면 붉어진다」는 것은 **주장이 아니라 관측이어야 한다.** 그 칸은 변경 전에도 초록, M2 이후에도 초록, AC-CHB-002 의 뮤턴트(옛 순서 — `IsAbs` 를 그대로 첫 번째로 둔다) 아래에서도 초록이다. 즉 이 SPEC이 예정한 어떤 단계도 그 칸이 붉어지는 것을 보지 않는다. 그래서 이 불변식은 **AC-CHB-009**(백슬래시 검사를 `IsAbs` **위로** 올리는 두 번째 뮤턴트)가 세우며, 이 문장은 그 AC 를 인용할 뿐 결과를 미리 단언하지 않는다. AC-CHB-009 가 RED 를 세우지 못하면 이 불변식은 지켜지지 않은 채 초록이었던 것이고, 그 사실이 판정이다.
2. `~user` 형태는 계속 oddly-formed.
3. 상대 + 백슬래시는 계속 oddly-formed.
4. 백슬래시 없는 홈-상대는 계속 home-relative.

---

## 4. 제외 범위

### Out of Scope — 백슬래시 경로의 구제

- 백슬래시를 품은 선언을 정규화하거나 정정해 해석 가능하게 만드는 일. 이 SPEC은 **거절**만 대칭으로 만든다.
- 이미 설정에 존재하는 백슬래시 등록의 정리·병합. 그것은 `moai clean --codex-skills`(t506) 소관이다.

### Out of Scope — 다른 분류 축

- 상대 경로의 해석 기준(resolution base) 결정. 지금처럼 「관측된 기준이 없다」로 남긴다.
- 디렉터리가 stat 을 통과하는 문제, `enabled` 필드 의미론, 파서의 `FirstUnrecognizedLine` 축.

### Out of Scope — 쓰기 쪽 변경

- `codex_skills_disable.go`의 거절 집합·`toConfigPath` 변환. 그 쪽은 이미 대칭의 기준선이며 이 카드는 읽기/prune 쪽만 움직인다.

### Out of Scope — 홈-상대 선언의 `..` 이탈 (`~/../../etc/x`)

- **인접하지만 이 카드가 닫지 않는 결함 하나를 명시적으로 남긴다.** `~/../../etc/x` 형태의 선언은 백슬래시를 하나도 품지 않는다. 따라서 오늘의 분류기에서 home-relative 로 분류되고, `filepath.Join(home, "../../etc/x")` 이 내부적으로 `Clean` 을 부르므로(`$GOROOT/src/path/filepath/path_unix.go` 의 `join` → `Clean`) 확장 결과가 **홈 밖**(`/etc/x`)으로 떨어진다. 그 경로에 `os.Stat` 이 걸리고, 존재하면 `Eligible=true` 까지 간다 — 이 SPEC이 닫고 있는 것과 **같은 경로**다.
- **그리고 M2 이후에도 그대로 살아남는다.** M2 가 올리는 것은 `ContainsRune('\\')` 검사뿐이고 이 선언에는 백슬래시가 없으므로, 순서 재배치는 이 칸을 건드리지 않는다. 「홈-상대 이탈은 이 카드에서 고려되어 처리됐다」고 읽지 말 것 — 고려됐고, **처리되지 않았다.**
- 이 SPEC은 `..` 이탈을 고치지 않으며, 그것을 검사하는 AC 도 두지 않는다. 여기서 함께 고치면 판별식이 둘이 되어 두 팔 대칭 AC 의 구속력이 흐려진다.
- **소유자 — 아직 발행되지 않은 후속 카드.** 이 트리 시점의 백로그 큐에는 이 형태를 소유하는 카드가 없다. 실측: `moai todo list --limit 0`(`EXIT=0`, 48행 — `moai todo` 단독형은 20행에서 잘리고 `--limit 0` 은 `Unknown flag` 로 거절되므로 전수 조회는 `list --limit 0` 뿐이다) 출력에 `grep -cE 'home\.escape|홈-상대.*이탈|\.\./\.\.|Clean'` → **0**. 그 0 이 필터 고장이 아님을 세우는 양성 대조군: 같은 출력에 `grep -cE 'codex'` → **11**. 대소문자 무시형은 1건 적중하나 무관한 카드의 `moai worktree clean` 오탐이다. 출력 사본은 `.moai/reports/t571/todo-exhaustive-48rows.txt`. 카드 발행은 운영자의 행위이므로 이 SPEC이 카드를 만들지 않고, 대신 발행문을 여기에 고정해 둔다: **「codex prune: 홈-상대 선언의 `..` 이탈 — `~/../../etc/x` 가 `filepath.Join` 의 `Clean` 을 타고 홈 밖으로 확장돼 stat·삭제 판정에 도달한다. SPEC-CODEX-HOME-BACKSLASH-001 §4 가 범위 밖으로 선언한 잔여 축」.** 후속 카드가 발행되면 그 id 를 이 줄에 적어 넣는다.

### Out of Scope — 검증 범위

- `go test ./...` 전체 스위트 로컬 실행. 판정은 CI 몫이며, 이 카드는 `./internal/cli/...`만 돌린다.
