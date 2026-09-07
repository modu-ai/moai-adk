# SPEC-CODEX-GHOST-SKILLS-MEASURE-001 — 수락 기준

> **Tier M** (발행 시 S → plan-audit 후 격상). Tier M 은 `acceptance.md` 를 정식 산출물로 갖는다 — 발행 등급이 S 이던 동안 이 파일은 「예외적 분리」였으나, 격상으로 예외가 아니라 규정 산출물이 됐다.
>
> 좌표: tree `4b07b9eb68e98f6ef8a2a861e854431a330cc27d`, HEAD `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`, 워크트리 `.claude/worktrees/t533`.

## §D. 층 1 — 지금 닫히는 AC

### AC-CGM-001 — `~/.codex` **하위 전체** 쓰기 0건 (REQ-CGM-001)

REQ 의 범위는 `~/.codex` 하위 전부이므로 파일 하나의 mtime 으로는 부족하다. **특히 `--force` 가 가장 먼저 하는 일이 새 백업 파일 생성이므로**(`codex_skills_prune.go:204`, 주석 "The backup lands BEFORE the write"), 백업만 남기고 중단된 쓰기는 `config.toml` mtime 검사에 걸리지 않는다.

**Given** run 시작 시점
**When** 아래로 디렉터리 스냅숏을 뜨고 **증거 파일로 산출물에 남기면**

```
ls -la ~/.codex | awk 'NR>1{print $NF}' | sort > .moai/reports/t533/codex-entries.before
stat -f '%N %m' ~/.codex/* ~/.codex/.[!.]* 2>/dev/null | sort > .moai/reports/t533/codex-mtimes.before
```

**Then** 두 파일이 `.moai/reports/t533/` 에 존재하고 비어 있지 않다 — 기준값이 산출물에 없으면 이 AC 는 사후 재검증이 불가능하다.

**And** run 종료 시점에 같은 명령으로 `.after` 두 개를 떠서 `diff` 하면 **양쪽 모두 무출력**이다. 어느 한쪽이라도 출력이 있으면 FAIL 이며 즉시 중단한다.

**뮤턴트 (판별력 증명, 필수) — `~/.codex` 밖에서 돌린다**: 셀렉터가 새 엔트리를 잡는다는 것은 `mktemp -d` 한 격리 디렉터리에서 증명한다. **가드의 물성을 증명하는 데 실제 사용자 트리를 쓰지 않는다** — 이 카드는 읽기 전용이고, t502 가 같은 트리를 쓰고 있다. 절차와 실측 결과는 plan.md §C 5번(`2a3 > .mutant`, rc=1). `mktemp -d` 산출물이라 중단돼도 사용자 트리에 잔재가 남지 않는다.

> **[HARD] `~/.codex` 안에서의 뮤턴트는 금지다.** 종전 수리안은 `touch ~/.codex/.t533-mutant-probe` → 후행 `rm` 형태였는데, ① 위치가 plan 사전점검이라 읽기 전용 [HARD] 를 스스로 어겼고 ② 잔재 보장이 **후행 `rm` 뿐**이라 중단되면 그 줄에 도달하지 못한다 — 「후행 kill 은 cleanup 이 아니다」와 같은 축이다. 어떤 사정으로 `~/.codex` 안에서 돌려야 한다면 그것은 **층 2**이며, 선행 조건은 t502 착지 + 리드 지시이고, **중단 시 잔재가 남을 수 있으므로 `.after` 재측정 전에 수동 제거가 선행 조건**이다 — 보장할 수 없으므로 보장한다고 쓰지 않는다.

보조 확인: `.moai/reports/t533/measure.sh` 에 `--force` 문자열이 없다 — `grep -c -- '--force' …/measure.sh` → `0`, 대조 `grep -c 'codex' …/measure.sh` → 1 이상(대조가 비면 파일을 잘못 짚은 것이다).

### AC-CGM-002 — 유령 개수, 두 계수기 일치 (REQ-CGM-002)

**Given** 이 머신의 `~/.codex/config.toml`
**When** `grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml` 과 `awk '/^\[\[skills\.config\]\]/{n++} END{print n}' ~/.codex/config.toml` 을 각각 실행하면
**Then** 두 출력이 모두 `49` 다.

### AC-CGM-003 — 계수 함정이 기록됐다 (REQ-CGM-003)

**Given** 같은 파일
**When** `grep -c 'enabled = false' ~/.codex/config.toml` (파일 전체)와 블록 내 awk 스캔을 각각 실행하면
**Then** 전자는 `53`, 후자는 `49` 이고, spec.md §B.2 가 이 불일치를 계수 함정으로 명시한다.
**And** §B.2 **안에서** `53` 이 블록 개수로 인용되지 않는다 — 이 절의 판정 범위는 §B.2 본문으로 한정한다. (문서 전체에 `grep -n '53'` 을 거는 형태로는 닫을 수 없다: 카드 id `t533` 과 요구 id `REQ-CGM-003` 이 같은 문자열을 포함하므로 §B.2 밖 적중이 구조적으로 남는다. 닫히지 않는 절을 AC 에 남기지 않기 위해 범위를 좁힌다.)

### AC-CGM-004 — 경로 부재, 루트는 생존 (REQ-CGM-002 보조)

**Given** 유령 경로 표본
**When** `grep -o '/Users/goos/.codex/skills/[^"]*' ~/.codex/config.toml | head -5` 의 각 경로를 `test -e` 로 검사하면
**Then** 5/5 가 MISSING 이다.
**And** `ls ~/.codex/skills/` 는 `.system` 과 `hatch-pet` 만 나열한다 — 루트는 존재하며, 사라진 것은 `moai-*` 디렉터리뿐이다.

### AC-CGM-005 — prune 미실행, 판별식은 파일명 형식 (REQ-CGM-004)

**Given** `~/.codex/`
**When** 엄격 선택자 `ls ~/.codex/ | grep -E 'config\.toml\.bak-[0-9]{8}T[0-9]{6}Z$'` 를 실행하면
**Then** 출력이 없고 rc=1 이다.
**And** 같은 코퍼스의 느슨한 대조 `ls ~/.codex/ | grep 'config\.toml\.bak-'` 는 **2건**(`config.toml.bak-20260822-022202`, `config.toml.bak-20260901-133347`)을 내어 빈 결과가 공허하지 않음을 보인다.
**And** 그 두 파일이 `YYYYMMDD-HHMMSS` 형식임이 기록되고, 종전의 "백업 파일 부재" 서술이 정정으로 남는다.

### AC-CGM-006 — 타임라인 미괄호 (REQ-CGM-005)

**Given** 스냅숏 4종
**When** 각각에 `grep -c '^\[\[skills\.config\]\]'` 를 실행하면
**Then** 네 값이 모두 `49` 이고, 발생 시점을 백업으로 좁힐 수 없음이 기록된다.

### AC-CGM-007 — 재직렬화 주체가 moai 가 아니다 (REQ-CGM-006)

**Given** 이 워크트리의 Go 소스
**When** `grep -rln 'skills\.config' internal/ --include='*.go' | grep -v _test.go` 를 실행하면
**Then** 정확히 `clean.go`, `codex_skills_prune.go`, `doctor_codex.go`, `codexwiring/skills.go` 네 개가 나온다.
**And** `grep -n 'os.WriteFile\|os.Create\|atomicWrite\|\.Encode(\|WriteString' internal/codexwiring/skills.go` 가 rc=1 (무출력) 이며, 같은 파일에 대한 느슨한 대조 `grep -c 'func ' internal/codexwiring/skills.go` 가 1 이상이다.
**And** 2026-08-22 스냅숏과 현재 파일의 `path =` 목록 diff 가 **순서 변경 1건**만 낸다.

### AC-CGM-008 — 판정 함수가 `enabled` 을 보지 않는다 (REQ-CGM-007)

**Given** `internal/cli/codex_skills_prune.go` 의 판정 함수 본문(59-108행)
**When** `sed -n '59,108p' internal/cli/codex_skills_prune.go | grep -c 'Enabled\|enabled'` 를 실행하면
**Then** 출력이 `0` 이다.
**And** 같은 범위의 대조 `grep -c 'Path'` 가 `10` 이어서 무출력이 공허하지 않음을 보인다.
**And** spec.md §B.10 이 적격 판정을 「`fs.ErrNotExist` 단일 양성 검사 + 그 앞 관문들」로 기술하며, 어느 관문도 `enabled` 을 보지 않음을 명시한다.

### AC-CGM-009 — 호출 여부를 "없음"으로 보고하지 않는다 (REQ-CGM-008)

**Given** 호출 이력을 물을 수 있는 이 머신의 코퍼스
**When** 아래를 실행하면

- `~/.moai/logs/`: 대상 `grep -rc 'codex-skills' ~/.moai/logs/ | grep -v ':0$'` → 무출력, 대조 `grep -rl 'moai' ~/.moai/logs/` → 2건 이상
- `~/.zsh_history`: **먼저 계측기를 검증한다.** 반환을 `rc` 만이 아니라 **출력 길이까지** 재야 두 경우가 갈린다 —
  - 깨진 계기: `out=$(grep -c 'moai' ~/.zsh_history 2>&1)` → **`rc=1 len=0 out=[]`** — `-c` 를 줬는데 **아무것도 찍지 않는다**. 이것은 `0` 이 **아니다**. 반면 `head -1` 은 `moai glm` 을 보인다 ⟹ 계측기 고장.
  - 정상 계기: `out=$(LC_ALL=C grep -ac 'codex-skills' ~/.zsh_history)` → **`rc=1 len=1 out=[0]`** — 실제로 `0` 을 찍는다.
  - rc 는 양쪽 다 1 이므로 **rc 만으로는 구별되지 않는다. 길이가 판별식이다.**
  - 계기 교체 후: `LC_ALL=C grep -ac 'moai'` → `1071`(대조 성립), `codex-skills` → `0`, `moai clean` → `0` — 이 셋은 모두 실제로 수를 찍는 경우다.

**Then** spec.md §F.1 이 이 결과를 **"이 질문을 덮는 범위의 계기가 이 머신에 없다"**(계기 범위 부족)로 서술하고, "호출된 적 없다"로도 **"계기가 없다"로도** 서술하지 않는다.
**And** `~/.zsh_history` 부재 서술이 정정 2, 1차 초안의 "계측기가 없다" 문장이 정정 3 으로 각각 기록된다.
**And** 덮지 못하는 지점 5종이 열거되며, 그중 ③ 은 「미상」이 아니라 **측정된 상한**으로 적힌다 — `grep -n 'SAVEHIST' /etc/zshrc` → `18:SAVEHIST=1000`, `LC_ALL=C grep -ac '^: [0-9]' ~/.zsh_history` → `1649`. 항목 수가 상한을 넘으므로 창 시작은 절단 경계일 수 있다.
**And** ④ 의 「다른 터미널 불가시」가 **거짓임이 기록된다** — `LC_ALL=C grep -ao 'lane-[0-9]*' ~/.zsh_history | sort -u | wc -l` → `20`(대조: 전체 적중 104줄). 여러 터미널 세션이 한 파일을 공유한다.
**And** 「결정적」 라벨이 ⑤(미측정 근거)가 아니라 ③(측정된 사실)에 붙는다.
**And** 이 두 수정이 간극을 **넓힌다**는 사실과, 「모른다」가 사실은 「안 재봤다」였던 계열의 3회차(정정 2·3·4)임이 정정 4 로 기록된다.
**And** spec.md §F 가 **깨진 계기의 반환을 `0` 이 아니라 무출력(len=0)으로** 적고, 정상 계기가 실제로 `0` 을 찍는 것과 **구분해서** 적는다(정정 5). 「무출력은 0 이 아니다」를 가르치는 문서가 그 자리에서 스스로를 반증하지 않게 하는 절이다.
**And** 이 문서 전체에서 `0` 이라 적힌 모든 자리가 **「실제로 0 을 찍는가, 아무것도 안 찍는가」로 재확인된 상태다.** 전수 훑기 결과: 실제로 `0` 을 찍는 자리 — `grep -c 'Enabled\|enabled'`(spec.md §B.10) · `grep -c -- '--force'`(AC-CGM-001) · `grep -c '\.go$'`(AC-CGM-010) · `LC_ALL=C grep -ac` 3종(위). 무출력인데 종전에 `0` 으로 적혀 있던 자리 — `grep -c 'moai' ~/.zsh_history`(수정 완료) · `grep -rc … | grep -v ':0$'`(라운드 1 수정 완료). **라운드 1 에서 후자 1곳만 고친 것이 이 결함이 살아남은 원인이다 — 축이 같으면 전수로 훑는다.**
**And** spec.md §F.2 가 깨진 대조군 사건을 기록하고, 그 처방(프로브와 대조군은 도구·코퍼스·needle 중 무엇이든 **실제로 달라야 한다**)을 오늘의 계열 4항 중 마지막 자리에 연결해 적는다.

### AC-CGM-010 — Go 코드 변경 0줄 (판별력 있는 대조군 필수)

**종전 형태는 오늘 실제로 공허하게 통과했다** — 이 카드 자신의 AP-4 위반이다. `git diff --name-only <base>..HEAD -- '*.go'` 가 0 이었는데, **필터를 뗀 `git diff --name-only <base>..HEAD` 도 0** 이었다(HEAD 가 아직 base 이고 SPEC 디렉터리가 미추적이라 커밋 범위가 비어 있음). 대조군이 「`.go` 무변경」과 「빈 범위」를 구별하지 못했다. 아래는 §F.2 처방(프로브와 대조군이 **실제로 달라야** 한다)의 직접 적용이다.

**[HARD] 게이트는 국면에 따라 분기한다.** 작업 트리 셀렉터를 무조건 적용하면 **커밋 이후의 정상 상태(깨끗한 트리)가 FAIL 로 판정된다** — 판별력을 지키되 정상 상태를 실패로 만들지 않는다. 국면은 `git rev-parse HEAD` 가 base(`6a46c0edb`)와 같은지로 가른다.

**국면 A — 미커밋 (`HEAD == 6a46c0edb`)**

**Given** 산출물이 아직 작업 트리에만 있는 상태
**When** 대조군을 **먼저** 세운다:

```
git status --porcelain -- .moai/specs/SPEC-CODEX-GHOST-SKILLS-MEASURE-001 | wc -l   # ≥ 1
git status --porcelain | wc -l                                                       # ≥ 1  (필터 제거 시 출력 존재)
```

**Then** 두 값이 모두 1 이상이다. 하나라도 0 이면 코퍼스가 비어 있다는 뜻이므로 **아래 주장을 세우지 않고 FAIL(측정 불가)로 보고한다.**
**And** 그 위에서만 프로브를 건다 — `git status --porcelain | grep -c '\.go$'` → `0`(이 셀렉터는 실제로 수를 찍는다).

**국면 B — 커밋 이후 (`HEAD != 6a46c0edb`)**

**Given** 산출물이 커밋된 상태 (작업 트리는 깨끗한 것이 정상이다)
**When** 대조군을 **커밋 범위**에서 세운다 — `git diff --name-only 6a46c0edb..HEAD | wc -l` → ≥ 1
**Then** 그 값이 1 이상이다. 0 이면 범위가 비었으므로 FAIL(측정 불가).
**And** 그 위에서만 프로브를 건다 — `git diff --name-only 6a46c0edb..HEAD -- '*.go'` → 무출력.

**공통 판별식**: 어느 국면이든 **「필터를 떼면 출력이 있다」**가 성립해야 이 AC 가 의미를 갖는다. 성립하지 않으면 PASS 가 아니라 측정 불가다. 국면을 잘못 골라 빈 셀렉터를 쓰는 것도 측정 불가이지 FAIL 이 아니다.

## §D.1 층 2 — 조건부 AC (지금 닫히지 않는다)

> **[HARD] 아래 세 AC 의 선행 조건은 카드 t502 의 착지다.** 착지 전에는 수행하지 않으며, 수행하지 않은 상태를 PASS 로도 FAIL 로도 표시하지 않는다 — `BLOCKED (t502 미착지)` 가 유일하게 옳은 표시다. "닫을 수 없어 생략"이 아니라 "이 조건에서 닫힌다"로 읽어야 한다.

### AC-CGM-011 — dry-run 선행 (REQ-CGM-009)

**전제**: `internal/cli/codex_skills_disable.go` 가 로컬 develop 에 존재하고, 리드가 실행을 지시했다.

**Given** 그 전제가 성립한 트리
**When** `moai clean --codex-skills` 를 **플래그 없이**(기본 dry-run) 실행하면
**Then** 생산자의 두 분기 중 정확히 하나가 나온다(`codex_skills_prune.go:190-197`). 기대 출력은 **분기별로** 적는다 — 종전 AC 는 한쪽만 적어 정상 동작이 FAIL 로 판정될 수 있었다. **AC 가 생산자를 잘못 기술한 것이므로 AC 를 고친다.**

| 분기 | 조건 | 마지막 행 (verbatim) |
|---|---|---|
| A | `eligible >= 1` | `%d of %d entries eligible for removal. Run with --force to actually remove.` |
| B | `eligible == 0` | `No removable entries in %s (%d declared)` — **분기 A 의 문장은 아예 찍히지 않는다** |

**And** 어느 분기든 항목별 행이 선행한다 — 적격 항목은 `[dry-run] Would remove: <path>`, 보존 항목은 `Kept: <path> (<사유>)`.
**And** 이 머신의 측정(§B.4: 표본 5/5 부재)에 비추어 분기 A 가 예상되나, **분기 B 도 정상 동작이며 FAIL 이 아니다** — 예컨대 t540 이 경로 해석을 바꿔 전부 「해석됨」으로 이동하면 B 가 나온다(§B.10 전제 취약성).
**And** 실행 전후 AC-CGM-001 의 `~/.codex` 하위 스냅숏 `diff` 가 무출력이다(dry-run 은 쓰지 않는다 — 파일 하나의 mtime 이 아니라 디렉터리 전체로 본다).

### AC-CGM-012 — `--force` 쓰기와 백업 (REQ-CGM-010)

**전제**: AC-CGM-011 의 dry-run 결과가 검토됐고, 리드가 **쓰기를 명시적으로 승인**했다. 승인 없이는 수행하지 않는다.

**Given** 승인된 상태
**When** `moai clean --codex-skills --force` 를 실행하면
**Then** 쓰기 **전에** `config.toml.bak-YYYYMMDDTHHMMSSZ` 형식의 백업이 생성되어, `ls ~/.codex/ | grep -E 'config\.toml\.bak-[0-9]{8}T[0-9]{6}Z$'` 가 **1건 이상**을 낸다(AC-CGM-005 의 rc=1 이 rc=0 으로 뒤집히는 것이 이 AC 의 판별식이다).
**And** `grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml` 이 `49 - N` 이다.
**And** 백업 파일의 블록 수는 여전히 `49` 다.

### AC-CGM-013 — 인계 사실의 자기 증거 대체 (REQ-CGM-011)

**전제**: `internal/cli/codex_skills_disable.go` 가 이 트리에 존재한다.

**Given** 그 파일
**When** 그 verb 의 대상 경로 해석·매칭 방식·멱등 출력·미해석 경로 미발행(REQ-CSD-011)을 이 트리에서 직접 읽으면
**Then** spec.md §C.1 의 네 항목이 "인계 사실"에서 "이 카드의 측정"으로 갱신되거나, 어긋난 항목이 정정으로 기록된다.
**And** C-2 반경 문장과 C-3 갈래 1 의 전제가 자기 증거를 갖는다.

## §E. 완료 정의 (Definition of Done)

- 층 1 AC 10건 전부 PASS, 각각 명령과 실제 출력 인용
- 층 2 AC 3건 전부 `BLOCKED (t502 미착지)` 로 표시 (착지 전 기준)
- `~/.codex` **하위 전체** 쓰기 0건이 **엔트리 목록 + mtime 스냅숏의 `before`/`after` diff 양쪽 무출력**으로 증명됨. **「`config.toml` mtime 불변」은 F5 가 폐기한 기준이므로 종료 게이트로 쓰지 않는다** — 백업만 남기고 중단된 쓰기가 그 검사를 빠져나가기 때문이다. run 이 이 DoD 를 종료 게이트로 읽으므로, 낡은 기준을 남겨 두면 F5 수리가 통째로 우회된다.
- 기준 스냅숏 2본(`codex-entries.before` / `codex-mtimes.before`)이 `.moai/reports/t533/` 에 **증거 파일로 존재**하고 비어 있지 않음 — 기준값이 산출물에 없으면 사후 검산 불가
- 셀렉터 판별력이 **격리 디렉터리**(`mktemp -d`)에서 증명됨. `~/.codex` 안에 뮤턴트 잔재 0건 — `ls ~/.codex/ | grep -c 't533'` → `0`
- 증거가 `.moai/reports/t533/` 에 반출되고, 인용된 경로가 audit 시점에 해석됨
- **정정 5건**이 SPEC 본문에 남음 — ① 백업 파일 부재 ② `~/.zsh_history` 부재 ③ 「계측기가 없다」(→ 범위 부족) ④ 「절단 미상」·「다른 터미널 불가시」(→ 재보지 않은 것) ⑤ 깨진 계기의 무출력을 `0` 으로 적은 것
- spec.md §G 의 `### Out of Scope —` 항목 5개가 유지됨
