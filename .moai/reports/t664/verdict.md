# t664 — GH #1680 파생: sync-phase 품질 게이트의 언어 감지 깊이

워크트리: `.claude/worktrees/t664` · 브랜치 `WT-gate-scan-depth` · base 로컬 develop `fb8bfff95`

표면 구분 [HARD]: 이 카드는 **셸 표면**(`.claude/hooks/moai/sync-phase-quality-gate.sh`)만 다룬다.
t559 가 고친 Go 표면(`internal/hook/quality/gate.go` 의 `detectToolchain`)은 건드리지 않았다.

## Claim

1. 공유 프로브 `has_suffix` 의 `-maxdepth 3` 때문에, 소스 접미사가 유일한 근거인 언어는
   자기 관용 배치에서 **원리상 감지되지 않았다**. 감지가 비면 게이트는 `GATE_LANG` 이 빈
   문자열이라 조용히 `exit 0` 한다.
2. 깊이 상한을 제거하고 **의존·빌드·캐시 디렉터리 prune** 으로 대체해, 15개 언어가 모두
   관용 배치에서 감지된다. Kotlin 전용 우회로였던 `has_kotlin_source` 는 공유 프로브
   위임으로 축소됐다.
3. 같은 함수 안에서 `-maxdepth 3` 사본을 따로 들고 있던 `.csproj` 프로브도 같은 결함이었고,
   공유 프로브로 합류시켰다.

## Evidence

이 워크트리, 이 트리에서 실행한 명령과 그 출력.

### 재현 — BEFORE (수리 전)

픽스처: 언어별 관용 배치 1건씩, **루트 매니페스트 없음**(접미사 프로브가 유일한 판단 근거).
하네스는 `.moai/reports/t664/measure-depth.sh`.

```
$ bash .moai/reports/t664/measure-depth.sh .claude/hooks/moai/sync-phase-quality-gate.sh
lang	want	got	depth
go	go	NONE	5
python	python	NONE	5
node	node	NONE	5
rust	rust	NONE	5
java	java	NONE	6
kotlin	kotlin	kotlin	6
csharp	csharp	NONE	5
ruby	ruby	NONE	5
php	php	NONE	5
elixir	elixir	NONE	5
cpp	cpp	NONE	5
scala	scala	NONE	6
r	r	NONE	4
flutter	flutter	NONE	5
swift	swift	NONE	5
```

**15개 중 14개 감지 실패.** Kotlin 만 잡히는 것은 t604 가 prune 기반 전용 프로브를
따로 넣어 두었기 때문이며, 카드 본문의 서술과 일치한다. 전문은 `before.tsv`.

### AFTER (수리 후)

```
$ bash .moai/reports/t664/measure-depth.sh .claude/hooks/moai/sync-phase-quality-gate.sh
(15행 전부 got == want)
```

전문은 `after.tsv`. 감지 실패 **14 → 0**.

| 지표 | BEFORE | AFTER |
|---|---|---|
| 관용 배치 픽스처 수 | 15 | 15 |
| 감지 성공 | 1 (kotlin) | 15 |
| 감지 실패(NONE) | 14 | 0 |

### 계약 시험

```
$ bash .claude/hooks/tests/test-language-routing-contract.sh
PASS: language routing distinguishes Kotlin, preserves monorepo candidates, and reaches conventional deep layouts
```

기존 4개 단정(Kotlin version-catalog / Kotlin alias / Java Gradle 대조군 / 코드-델타 패턴)은
그대로 통과한다. 추가된 것은 3군: 15개 언어 관용 배치 루프, prune 8종 각각에 대한
**미검출 단정**, 중첩 `.csproj`.

### 뮤턴트

`has_suffix` 에 `-maxdepth 3` 을 되돌린 뒤:

```
$ bash .claude/hooks/tests/test-language-routing-contract.sh
FAIL: version-catalog Kotlin project detected as java, want kotlin
EXIT=1
```

[주의] 뮤턴트는 **기존 t604 단정에서 먼저 죽었다** — `has_kotlin_source` 가 이제 공유
프로브에 위임하므로, 상한이 Kotlin 까지 함께 무너뜨린다. 그래서 새 단정이 독자적으로
힘을 갖는지 따로 쟀다:

```
$ bash .moai/reports/t664/measure-depth.sh <뮤턴트 훅>
NONE count under mutant: 15 of 15
```

수리본에서 15/15 감지, 뮤턴트에서 15/15 미감지 — 새 단정이 겨누는 행동은 뮤턴트에서
확실히 깨진다. 원복 후 해시 일치:

```
$ shasum -a 256 -c .moai/reports/t664/pre-mutant.sha256
.claude/hooks/moai/sync-phase-quality-gate.sh: OK
```

### 템플릿 미러 + 빌드

동일 편집을 `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` 에
적용했다(cp 아님 — 두 사본은 의도적으로 분기할 수 있다). 결과적으로 두 파일은 현재 동일하다.

```
$ bash -n internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh
TEMPLATE-SYNTAX-OK
$ diff -q .claude/... internal/template/templates/.claude/...
MIRROR-IDENTICAL
$ bash .moai/reports/t664/measure-depth.sh <템플릿 사본>
template mismatches: 0
$ make build
EXIT=0        # 전문 make-build.log
```

## Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t664`, base 로컬 develop `fb8bfff95`
  (`git rev-parse --short HEAD` → `fb8bfff95`, 측정 시점 미커밋).
- BEFORE 수치는 수리 **전** 이 트리에서 잰 값이고, AFTER 는 아래 변경 파일 목록을 적용한
  뒤 같은 하네스·같은 픽스처 형태로 다시 잰 값이다. 판별식(관용 배치 + 루트 매니페스트 없음)은
  양쪽 동일하다.
- 비용 측정의 BEFORE 훅은 `git show fb8bfff95:.claude/hooks/moai/sync-phase-quality-gate.sh`
  로 꺼낸 사본이며, 측정 후 삭제했다(같은 명령으로 재현 가능). `cost-before.txt` 의
  `hook` 줄이 가리키는 경로는 그래서 현재 해소되지 않는다 — 재현 명령이 귀속이다.
- 부하 게이트: 착수 `load 11.91`, 측정 직전 `load 10.29`/`16.00` — 모두 임계 30 이하.

## 변경 파일

- `.claude/hooks/moai/sync-phase-quality-gate.sh` — `has_suffix` 의 `-maxdepth 3` 을
  prune 집합으로 교체(`internal/hook/quality/gate.go` 의 `sourceScanSkipDirs` 를 미러:
  `.git .hg .svn node_modules vendor .venv venv site-packages __pycache__ .tox .nox
  .mypy_cache .ruff_cache .pytest_cache dist build target .next .output`).
  `has_kotlin_source` → `has_suffix '*.kt'` 위임. `.csproj` 프로브 → 공유 프로브 합류.
- `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` — 동일 편집.
- `.claude/hooks/tests/test-language-routing-contract.sh` — 15언어 관용 배치 루프,
  prune 8종 미검출 단정, 중첩 `.csproj` 단정 추가.

## Gaps

- **`.kts` / 매니페스트 계열 프로브는 재측정하지 않았다.** 루트 `-f` 검사(`go.mod`,
  `package.json` 등)는 손대지 않았으므로 동작이 바뀔 수 없다고 판단했으나, 그 판단을
  명령으로 확인하지는 않았다.
- **설치된 바이너리 대상 E2E 미실행.** 실제 sync-phase 커밋에서 게이트를 끝까지 돌린
  실행 증거는 없다. 위 증거는 `detect_languages` / `detect_language` 를 직접 부른
  단위 수준이다.
- **Windows·Linux 미검증.** `find` 의 prune·`-quit` 동작은 macOS(BSD find)에서만 쟀다.
  GNU find 에서도 같은 문법이 유효하다고 알려져 있으나 이 트리에서 확인하지 않았다.
- **prune 집합의 완전성 미평가.** Go 쪽과 맞춘 것이지, 16개 언어 생태계의 의존 디렉터리를
  전수 조사해 도출한 목록이 아니다(예: `Pods`, `DerivedData`, `_build`, `deps` 는 없다).

## Residual-risk

- **미적중 언어가 전체 순회를 한다.** `-quit` 은 적중 시에만 조기 종료하므로, 트리에 없는
  언어는 매번 prune 된 전체 순회를 지불한다. 이 저장소 실측:

  ```
  sec_per_detect_languages   0.319 (BEFORE, 상한 있음)  →  1.598 (AFTER)
  ```

  약 5배. 다만 이 비용은 **sync-phase 커밋에서만** 발생한다 — 게이트는 커밋 제목이
  sync-phase 가 아니면 감지 이전에 `exit 0` 한다. 훅 타임아웃은 60s 이므로 여유가 크지만,
  훨씬 큰 저장소에서는 다시 재야 한다.

- **테스트 픽스처가 언어 후보로 올라온다 (실측).** 이 저장소의 후보가
  `go,node,kotlin,cpp` (4) 에서 `go,python,node,rust,kotlin,ruby,php,elixir,cpp,scala,swift,csharp` (12)
  로 늘었다. 출처를 추적한 결과 전부 테스트 픽스처였다 —
  `internal/navigator/astx/testdata/polyglot/sample.swift`,
  `internal/astgrep/testdata/fixtures/php/*.php` 등.

  영향 범위는 좁다: 툴체인은 `GATE_LANG`(= `head -1` = `go`, 불변) 하나만 돌고,
  늘어난 후보는 `code_delta_pattern` 합산(243-249행)에만 쓰인다. 즉 "문서만 바뀐 커밋"
  판정이 조금 덜 자주 성립할 뿐, 새 툴체인이 실행되지는 않는다.

  `testdata` / `fixtures` 를 prune 에 넣지 **않은** 이유: (1) 다른 프로젝트에서는 그 경로가
  실제 소스일 수 있고, (2) Go 쪽 `sourceScanSkipDirs` 에도 없어서 넣으면 두 표면이 갈라지며,
  (3) 카드 범위 밖이다. 이것이 문제로 판정되면 별도 카드가 맞다.

- **prune 이 실제 소스를 가릴 수 있다.** `build/` `dist/` `target/` 아래에 손으로 쓴 소스를
  두는 프로젝트에서는 그 언어가 감지되지 않는다. 상한을 prune 으로 바꾼 대가이며,
  t604 가 Kotlin 에 대해 이미 내린 판단을 그대로 확장한 것이다.
