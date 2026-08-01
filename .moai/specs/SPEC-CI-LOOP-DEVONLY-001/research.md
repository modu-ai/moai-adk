# research — SPEC-CI-LOOP-DEVONLY-001

모든 명령은 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/ci-loop-devonly`
(브랜치 `plan/SPEC-CI-LOOP-DEVONLY-001`, HEAD `a64548a2a`) 안에서 실행했다.
아래 각 항목은 **실행한 명령**과 **관측한 출력**을 함께 기록한다.

---

## §A 영향 표면 재측정

### A.1 스크립트 경로 참조

```bash
grep -rln 'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/ | sort
grep -rn  'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/ | wc -l
```

관측: **8개 파일, 27개 참조.** 파일 목록은 미션 브리프의 8개와 정확히 일치한다.

```
internal/template/templates/.claude/agents/moai/manager-develop.md
internal/template/templates/.claude/rules/moai/core/zone-registry.md
internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md
internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md
internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md
internal/template/templates/.claude/skills/moai-workflow-ci-loop/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/run.md
```

### A.2 스킬 식별자 참조 (브리프에 없던 추가 표면)

```bash
grep -rln 'moai-workflow-ci-loop' internal/template/templates/ internal/template/catalog.yaml | sort
```

관측: **9개 항목.** A.1과 3개(`ci-autofix-protocol.md`, `ci-watch-protocol.md`,
`moai-workflow-ci-loop/SKILL.md`)가 겹치고, 다음 6개는 A.1에 없다:

```
internal/template/catalog.yaml
internal/template/templates/.claude/skills/moai/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/fix.md
internal/template/templates/.claude/skills/moai/workflows/loop.md
internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
internal/template/templates/.moai/config/sections/delegation.yaml
```

`delegation.yaml`은 3곳에서 스킬을 기본 위임 대상으로 지정한다
(`grep -c 'moai-workflow-ci-loop' …/delegation.yaml` → `3`).
`sync/delivery.md:239`는 `Skill("moai-workflow-ci-loop")`를 직접 호출한다.

### A.3 합집합

```bash
{ grep -rl 'scripts/ci-watch\|scripts/ci-autofix' internal/template/templates/;
  grep -rl 'moai-workflow-ci-loop' internal/template/templates/ internal/template/catalog.yaml; } | sort -u | wc -l
{ grep -rl 'scripts/ci-watch\|scripts/ci-autofix' .claude/;
  grep -rl 'moai-workflow-ci-loop' .claude/ .moai/config/sections/delegation.yaml; } | sort -u | wc -l
```

관측: 템플릿 측 **14**, 개발 저장소 미러 측 **13** (미러에는 `catalog.yaml`이 없다).

> **브리프 대비 정정**: 브리프의 "8개 파일"은 스크립트 *경로* 참조 기준으로 정확하다.
> 그러나 스킬 제거까지 승인된 방향에 포함되므로 실제 편집 대상은 **양측 합계 27개 파일**이다.

---

## §B watch 능력의 소재 — v0.1.0 판단의 자기 반증

> **v0.1.0 정정 기록.** 이 절의 초판은 `moai pr watch --help`의 산문을 읽고
> "watch 능력은 CLI로 배포되며 스크립트는 얇은 래퍼"라고 결론지었다.
> 그 결론은 **명령을 실행하지 않고 도움말만 읽은 데서 나온 오류**이며,
> 아래 실행 결과로 반증되었다. 초판 결론은 폐기한다.

### B.1 명령을 실행한 결과

```bash
moai pr watch 999 --branch main; echo "EXIT=$?"
```

관측:

```
[ci-watch] Use scripts/ci-watch/run.sh to start the watch loop.
[ci-watch] Example: MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh 999 main
EXIT=0
```

CLI는 watch를 수행하지 않는다. **셸 스크립트를 실행하라는 안내문을 출력하고 종료한다.**

### B.2 소스 확인

```bash
sed -n '35,80p' internal/cli/pr_watch_cmd.go
grep -rn 'os.Exit(2)' internal/cli/pr_watch_cmd.go internal/ciwatch/*.go
```

관측 (발췌):

```go
			if flags.abort  { return runPRWatchAbort(statePath) }
			if flags.report { return runPRWatchReport(args[0], flags.branch) }

			// Default: just print usage info directing user to the shell script.
			fmt.Fprintf(os.Stderr, "[ci-watch] Use scripts/ci-watch/run.sh to start the watch loop.\n")
```

`os.Exit(2)` grep 결과: **매치 0건.**

CLI의 실제 모드는 셋이다 — 기본(안내문 출력, exit 0), `--abort`(상태 파일 플래그),
`--report`(보고서 서식화). 폴링 루프·30초 주기·30분 타임아웃·`exit 2`는 전부
`scripts/ci-watch/run.sh`(219행)에 있다.

### B.3 도움말과 구현의 괴리

`pr_watch_cmd.go:37-39`의 `Long` 텍스트는 다음을 **주장**한다:

```
On required-failure:  emits a JSON handoff to stdout (exit 2).
On 30-min timeout:    exits with code 3.
```

`RunE`는 이 중 무엇도 구현하지 않는다. **도움말 산문은 구현의 증거가 아니다** —
이것이 초판 오류의 직접 원인이며, 이 SPEC 전체에 걸쳐 적용되는 교훈이다.

### B.4 귀결

- 배포된 사용자는 watch 능력을 **가지고 있지 않다**. 미션 브리프의 원 전제가 옳았다.
- `ci-watch-protocol.md`의 8개 Frozen 절은 **수행 주체가 없는 동작**을 규율한다.
- 초판이 §F.2에 적었던 "능력이 배포되므로 규칙을 유지해야 한다"는 전제 1은 거짓이다.
- 초판이 제안했던 치환 규칙
  `moai pr watch reports a required-check failure (exit 2)`는
  **배포 규칙에 거짓 서술을 기입**하고 M2를 통해 Frozen 절 `CONST-V3R5-014`까지
  오염시켰을 것이다. 이 치환 규칙은 폐기한다.

---

## §C Q1 — 18개 Frozen 절을 어떤 절차로 처리하는가

### C.1 검증 엔진의 실제 동작

```bash
grep -n 'SentinelZoneUnregistered\|SOURCE_FILE_MISSING' internal/constitution/validator.go
sed -n '160,175p' internal/constitution/validator.go
```

관측:

- `SOURCE_FILE_MISSING` → `*ValidationError`, **fatal, exit 2**
- `ZONE_UNREGISTERED` → 소스 파일에 `[ZONE:…]` 마커가 있으나 레지스트리에 없을 때 발생
- `DRIFT` → 레지스트리 `clause` 텍스트가 소스 파일에서 발견되지 않을 때 발생 (non-fatal, exit 1)

**한 방향만 막혀 있다** (초판의 "두 방향 모두 막혀 있다"는 서술은 반증되었다):

| 반쪽 상태 | 기대 (초판) | 실측 |
|---|---|---|
| 파일만 삭제, 레지스트리 항목 유지 | `SOURCE_FILE_MISSING` exit 2 | **확인됨** — `validate-exit=2`, ci-count 18→11, total 77→49 |
| 레지스트리 항목만 삭제, 소스 유지 | `ZONE_UNREGISTERED` | **반증됨** — `validate-exit=1`, `ZONE_UNREGISTERED=0` |

두 번째 상태를 재현한 명령과 관측:

```bash
# 8개 항목(014..021)만 제거, ci-watch-protocol.md 소스는 유지
moai constitution validate; echo "validate-exit=$?"
# → validate-exit=1
grep -c 'ZONE_UNREGISTERED' /tmp/cv2.txt   # → 0
grep -c 'canary_gate: true' .claude/rules/moai/core/zone-registry.md   # → 65
```

**귀결 (중요)**: 이 반쪽 상태에서 `canary_gate` 총계는 정확히 **65** — AC-CLD-011의 통과값 —
이며 소스 파일은 고아로 남는다. 즉 **AC-CLD-011 단독으로는 이 상태를 잡지 못한다.**
잡는 것은 AC-CLD-005(파일 부재)와 AC-CLD-007(`moai pr watch` 잔존)이다.
AC-CLD-009/010은 반대 방향(소스만 삭제)을 `validate-exit=2`로 잡는다.

소스와 레지스트리를 같은 변경에서 함께 움직여야 한다는 요구(REQ-CLD-011/012)는 유효하다.
다만 그 근거는 "양방향 검증 차단"이 아니라 **한 방향만 도구가 막고, 나머지 한 방향은
AC 조합으로 막아야 한다**는 것이다.

### C.2 현재 baseline — 18개 절은 이미 전부 드리프트 상태다

```bash
MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md moai constitution validate > /tmp/cv.txt 2>&1
echo "EXIT=$?"                                    # → EXIT=1
grep -c 'DRIFT\]\|SOURCE_FILE_MISSING\]' /tmp/cv.txt
grep -E '^\s+\[(DRIFT|SOURCE_FILE_MISSING)\]' /tmp/cv.txt | grep -c 'ci-autofix-protocol.md\|ci-watch-protocol.md'
grep -c 'ZONE_UNREGISTERED' /tmp/cv.txt
```

관측:

```
EXIT=1
77        # 전체 findings
18        # ci-autofix-protocol.md / ci-watch-protocol.md 귀속 findings
0         # ZONE_UNREGISTERED
```

`CONST-V3R5-004` ~ `CONST-V3R5-021` **18개 전부가 이미 `[DRIFT]`** 이다 —
레지스트리의 `clause` 텍스트가 현재 소스 파일에서 발견되지 않는다.
따라서 "Frozen 절을 건드리면 통과하던 게이트가 깨진다"는 우려는 성립하지 않는다.
게이트는 이 18개에 대해 **이미 실패 중**이며, 이 SPEC은 그것을 0으로 만든다.

`--- 정정 기준선`: "validate가 깨끗하다"는 판정은 불가능하다(무관 드리프트 59건 존재).
판정은 반드시 **ci 귀속 findings = 0** 이라는 델타 기준이어야 한다 (REQ-CLD-013).

### C.3 CI 게이트의 실제 강도

```bash
sed -n '335,375p' .github/workflows/ci.yml
```

관측:

```yaml
  constitution-check:
    name: Constitution Check
    continue-on-error: true
    ...
      - name: Verify zone registry (list)
        run: … ./bin/moai constitution list --format json | …
      - name: Verify zone registry (frozen entries)
        run: … ./bin/moai constitution list --zone frozen | tail -1
```

세 가지 사실:

1. CI는 `constitution validate`가 아니라 **`constitution list`** 만 실행한다 — 소스 대조를 하지 않는다.
2. `continue-on-error: true` — **자문(advisory) 잡이며 머지를 차단하지 않는다.**
3. 대상은 저장소 루트 `.claude/`이며, **템플릿 트리의 `zone-registry.md`는 CI 검증 대상이 아니다.**

> 미션 브리프의 "라이브 CI 게이트를 통과해야 한다"는 전제는 반증된다.
> 그럼에도 이 SPEC은 게이트가 약하다는 이유로 정합성을 낮추지 않으며,
> `moai constitution validate` 기준으로 판정한다 (acceptance.md AC-CLD-009/010).

### C.4 은퇴 마커 메커니즘의 부재

```bash
grep -rn 'retire\|Retire\|deprecated\|Deprecated\|superseded' internal/constitution/*.go | grep -v _test
```

관측: **매치 0건.** 레지스트리 스키마에 은퇴/폐기 상태 필드는 존재하지 않는다.
`internal/constitution/amendment.go`는 5계층 안전 게이트(Frozen 강등 증거,
canary 평가, 모순 탐지, 레이트 리밋)를 제공하지만, 이는 **런타임 개정 API**이고
레지스트리 YAML의 정적 편집 경로가 아니다.

**Q1 결론 (절차)**:
스키마 수준 은퇴 마커는 **존재하지 않으므로 사용할 수 없다.** 유일하게 검증을 통과하는 경로는
소스와 레지스트리를 **같은 커밋에서 동시에 처리**하는 것이다.
§F.2의 결정 A에 따라 18개 절은 두 갈래로 나뉜다:

| 처리 | 절 | 개수 | 근거 |
|---|---|---|---|
| **레지스트리 항목 삭제** | `CONST-V3R5-014..021` | 8 | 소스 파일 `ci-watch-protocol.md`가 배포에서 제거되므로, 항목을 남기면 `SOURCE_FILE_MISSING`(exit 2). 소스와 항목이 함께 사라져야 한다 |
| **재작성 후 텍스트 정합화** | `CONST-V3R5-004`, `013` | 2 | 스크립트 경로를 오케스트레이터 핸드오프 트리거로 치환 (§F.2) |
| **텍스트만 정합화** | `CONST-V3R5-005..012` | 8 | 스크립트 무관. 기존 드리프트만 해소 |

이 처리 후 `canary_gate: true` 총계는 73 → **65**로 감소한다 (8개 항목 삭제).
감소는 "주제가 소멸한 절의 제거"라는 사유에 한정되므로 허용된다.

우회(`MOAI_CONSTITUTION_SKIP_VALIDATE=1`)는 사용하지 않는다.

---

## §D Q2 — catalog 해시 재생성과 make build 필요 여부

```bash
sed -n '51,55p' internal/template/catalog.yaml
cat internal/template/embed.go
grep -n '^build:' -A 3 Makefile
```

관측:

```yaml
            - name: moai-workflow-ci-loop
              tier: core
              path: templates/.claude/skills/moai-workflow-ci-loop/
              hash: 3761e843a3b440e855012799fe04be1c47f130d2697a0bb34e3a1febfc039306
              version: 0.1.0
```

```go
//go:embed all:templates
//go:embed catalog.yaml
var embeddedRaw embed.FS
```

```make
build: templ-generate ## Build the binary
	@go run ./internal/template/scripts/gen-catalog-hashes.go --all
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai
```

**Q2 결론 (예 — 둘 다 필요하다)**:

- `catalog.yaml`은 `templates/`의 형제이며 **별도 `//go:embed` 지시자로 바이너리에 컴파일**된다.
  따라서 항목 삭제는 재컴파일 없이는 반영되지 않는다.
- 해시는 `make build`의 첫 단계 `gen-catalog-hashes.go --all`이 **자동 재생성**한다.
  손으로 해시를 계산할 필요가 없고, 그래서도 안 된다.
- 정확한 명령 순서:

```bash
# 1) 템플릿 소스 편집 (스킬 디렉터리 삭제 + catalog.yaml 항목 삭제)
# 2) 해시 재생성 + 재컴파일 (단일 타깃이 둘 다 수행)
make build
# 3) 임베드 FS 검증
go test ./internal/template/...
```

`make build`는 `templ-generate`를 선행 의존으로 갖는다. 별도의 `moai build` CLI 명령은 없다.

임베드 반영 여부의 판정은 소스 grep이 아니라 **컴파일된 FS**를 읽어야 한다.
저작 시점에 임시 프로그램으로 baseline을 확보했다:

```go
f, _ := template.EmbeddedTemplates()
// .claude/skills/moai-workflow-ci-loop 하위 파일 수를 센다
```

관측: `embedded files under .claude/skills/moai-workflow-ci-loop: 1`, exit 1.
(임시 프로그램은 측정 후 삭제했다. 구현 단계에서 영구 가드 테스트로 승격한다 — plan.md M4.)

### D.1 임베드 프로브가 탐지하지 **못하는** 것 (초판 근거 정정)

초판은 이 프로브가 "`make build` 누락을 탐지한다"고 서술했다. **거짓이다.**
감사가 변이 테스트로 반증했다: 소스 디렉터리를 삭제하고 **재빌드하지 않은 채**
프로브를 돌리면 0을 반환하고 통과한다. 이유는 단순하다 — `go test`가 컴파일을 수행하므로
`//go:embed`는 **항상 현재 소스를 반영한다.** 따라서 이 프로브는 원리적으로
"재빌드를 잊었는가"를 판별할 수 없다.

프로브의 유효한 역할은 남는다: 소스 트리 grep과 달리 **배포 산출물인 임베드 FS 자체**를
판정하므로, 디렉터리가 남아 있으면 실패한다. 그러나 그 근거를 "재빌드 누락 탐지"로
서술하면 안 된다.

### D.2 재빌드 누락을 실제로 탐지하는 게이트 (변이 검증 완료)

재빌드를 잊었을 때 트리에 남는 흔적은 **catalog.yaml의 낡은 해시**다.
생성기를 돌린 뒤 diff가 나면 커밋된 해시가 낡았다는 뜻이다.

```bash
go run ./internal/template/scripts/gen-catalog-hashes.go --all
git diff --exit-code -- internal/template/catalog.yaml
```

변이 테스트 — **세 상태 모두 실행함**:

```
STATE A  clean                        gen-exit=0  diff-exit=0
STATE B  템플릿 파일 1개 내용 변경        gen-exit=0  diff-exit=1   ← 탐지
STATE C  스킬 디렉터리 삭제 (M3의 동작)   gen-exit=1  diff-exit=0   ← diff는 놓치고 gen이 잡는다
STATE D  공백만 추가                    gen-exit=0  diff-exit=0   ← 정규화되어 탐지 대상 아님
(모든 상태 후 git checkout으로 복원 확인)
```

**초판이 이 절에 STATE B만 기록한 것이 N1 결함의 직접 원인이다.** 내용 변경으로만
변이 테스트했기 때문에 `git diff --exit-code` 단독으로 충분하다고 결론지었으나,
이 SPEC의 M3가 실제로 수행하는 것은 **삭제**(STATE C)이며 그 상태에서 생성기는
해시를 계산할 수 없는 항목의 재작성을 거부해 `catalog.yaml`을 손대지 않는다 —
가장 위험한 상태에서 트리가 깨끗해 보인다.

교훈: **변이는 마일스톤이 실제로 수행하는 연산으로 해야 한다.**
acceptance.md AC-CLD-004가 `gen-exit`와 `diff-exit`를 함께 채택한다.

---

## §E Q3 — 이미 배포된 스킬 폴더의 운명

### E.1 정리 경로 두 갈래

```bash
grep -n 'func scanDeprecatedPaths' -A 20 internal/cli/update_cleanup.go
grep -rn 'DeprecatedPaths = ' internal/
```

관측: `scanDeprecatedPaths`는 `defs.DeprecatedPaths`(`internal/defs/dirs.go:68`)라는
**명시적 등록 목록**만 순회한다. 등록되지 않은 경로는 건드리지 않는다.

```bash
sed -n '28,60p' internal/cli/update_archive.go
```

관측: 두 번째 경로는 `legacySkillIDs`이며 13개 ID를 보유한다. 주석 원문:

```go
// legacySkillIDs lists the skill IDs removed in BC-V3R3-007 that are still
// gone from the template tree. When `moai update` runs, these skills are moved
// to .moai/archive/skills/v2.16/.
...
// TestLegacySkillIDsNotEmbedded (update_archive_guard_test.go) asserts this list
// stays disjoint from the embedded template skill set, so a future revival cannot
// silently recreate the defect.
const archiveVersion = "v2.16"
```

`archiveSkill`은 소스 부재 시 `nil`을 반환하는 멱등 동작을 갖는다.

**Q3 결론 (고아가 된다)**:
어느 목록에도 등록하지 않으면, 이미 배포된 `.claude/skills/moai-workflow-ci-loop/`는
**정리되지도 보존 선언되지도 않은 고아 디렉터리로 남는다.** 사용자에게는
"업데이트했는데 죽은 스킬이 계속 리스트에 뜬다"로 관측된다.

**이 SPEC이 처리한다** (REQ-CLD-017/018). 선택 경로는 `legacySkillIDs` 등록이며,
`defs.DeprecatedPaths`가 아니다 — 전자는 아카이브(보존 후 이동)이고 후자는 삭제이기 때문이며,
REQ-CLD-018(사용자 저작 내용 미삭제)을 만족하는 것은 전자뿐이다. 단 REQ-CLD-018은 현재
어느 AC도 아카이브-대-삭제 동작을 관측하지 않으므로 문서화된 부채다 (acceptance.md 커버리지 표).

**단, 두 가지 선결 제약을 기록한다**:

1. `TestLegacySkillIDsNotEmbedded`가 임베드 스킬 집합과의 **분리(disjoint)** 를 강제한다.
   따라서 템플릿에서 스킬을 제거하고 임베드 FS에서 사라진 뒤에만 목록에 추가할 수 있다.
   의존 대상은 **M4**(재빌드 및 임베드 확인)이며 M2가 아니다 — 초판의 "M2 → M5" 표기는 오류다.
   이는 문체상 권고가 아니라 테스트가 강제하는 **실제 제약**이다.
2. `archiveVersion`은 `"v2.16"`으로 하드코딩되어 있다. v3.x 제거분을 v2.16 아카이브에
   넣는 것은 의미상 부정확하다. 이 부정확성을 감수할지, 버전 태그를 일반화할지는
   구현 단계의 결정 사항으로 plan.md M5에 명시한다.

---

## §F Q4 — 두 프로토콜 파일을 삭제할 것인가 재작성할 것인가

### F.1 결합도 실측

```bash
grep -n 'scripts/ci-watch\|scripts/ci-autofix' \
  internal/template/templates/.claude/skills/moai/workflows/run.md \
  internal/template/templates/.claude/agents/moai/manager-develop.md \
  internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
```

관측 (발췌):

```
manager-develop-prompt-template.md:20:| `autofix` | **DIAGNOSE-PATCH-VERIFY** | CI auto-fix loop after `scripts/ci-watch/run.sh` detects a failing required check; …
run.md:187:| PATCH scope | [HARD] SPEC scope ONLY; MUST NOT touch `.env*` / credentials / `scripts/ci-watch/run.sh` / …
```

`cycle_type=autofix`는 `manager-develop`의 **1급 사이클 타입**이다.
`ci-autofix-protocol.md`를 삭제하면 이 사이클의 정식 계약(반복 상한, 에스컬레이션,
의미론적 실패 처리, 보호 파일)이 배포 트리에서 사라진다.

`ci-watch-protocol.md`의 Frozen 절 8개는 `moai pr watch` / `.github/required-checks.yml`을
언급하지만, §B.2가 확정했듯 **그 CLI는 watch 루프를 수행하지 않는다** — 따라서 이 절들은
배포되는 어떤 산출물에도 구속되지 않는다. 이것이 §F.2가 이 파일을 배포에서 제거하는 근거다.

### F.2 결정 (GOOS 결정 A — 두 파일을 분리 처리한다)

§B.4에 따라 두 파일의 처지가 갈린다. 판별 기준은 **배포되는 산출물 중 그 규칙이 규율하는
동작을 실제로 수행하는 것이 있는가**이다.

| 파일 | 수행 주체가 배포되는가 | 처리 |
|---|---|---|
| `ci-watch-protocol.md` | **아니오** — watch 루프는 미배포 스크립트에만 존재 (§B.2) | **배포 중단** + Frozen 절 8개(`014..021`) 제거 |
| `ci-autofix-protocol.md` | **예** — `cycle_type=autofix`는 `manager-develop`의 실재 사이클 (§F.1) | **유지** + 스크립트 의존성만 제거, Frozen 절 10개(`004..013`) 정합화 |

`ci-watch-protocol.md` 배포 중단 근거: 수행 주체가 없는 동작을 규율하는 규칙을 배포하면,
사용자는 존재하지 않는 루프의 폴링 주기와 타임아웃을 지시받는다. 이는 이 SPEC이 제거하려는
결함과 정확히 같은 형태다.

`ci-autofix-protocol.md` 유지 근거: 파일을 지우면 `cycle_type=autofix`가 반복 상한·
에스컬레이션 계약·의미론적 실패 처리·보호 파일 규칙 없는 사이클이 된다.
문서화된 무능력을 문서화되지 않은 능력으로 바꾸는 것은 개선이 아니다.

재작성 규칙 (`ci-autofix-protocol.md`에만 적용):

| 현재 | 치환 후 |
|---|---|
| `after scripts/ci-watch/run.sh detects a failing required check` | `when the orchestrator hands off a failing required check` |
| 보호 파일 목록의 `scripts/ci-watch/run.sh` | 목록에서 제거 — 배포 트리에 없는 경로는 보호 대상이 될 수 없다 |
| `Skill("moai-workflow-ci-loop")` 호출 (`sync/delivery.md`) | 호출 제거 — 대체 호출 대상 없음 |

> **폐기된 치환 규칙.** 초판은 위 첫 행의 치환 후 값으로
> `moai pr watch reports a required-check failure (exit 2)`를 제시했다.
> §B.2가 이를 거짓으로 확정했다 — CLI에 `os.Exit(2)`는 존재하지 않는다.
> 이 규칙을 채택했다면 배포 규칙과 Frozen 절 `CONST-V3R5-014`에 거짓 서술이 기입되었을 것이다.

치환 후 문구는 스크립트도 CLI도 이름하지 않고 **오케스트레이터 핸드오프**를 트리거로 삼는다.
이는 배포 환경에서 실제로 참인 유일한 서술이다 — 오케스트레이터는 배포되며,
실패 정보를 어떤 경로로 얻든 autofix 사이클에 전달할 수 있다.

---

## §G 해소된 항목

### G.1 required-checks.yml — 이연하되 소멸이 예상된다 (GOOS 결정 B)

배포 트리에서 이 파일을 요구하는 참조를 전수 조사했다:

```bash
grep -rln 'required-checks.yml' internal/template/templates/
```

관측 (4건):

```
internal/template/templates/.claude/rules/moai/core/zone-registry.md
internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md
internal/template/templates/.claude/skills/moai-workflow-ci-loop/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
```

네 번째 항목의 실제 위치를 확인했다:

```bash
grep -n 'required-checks.yml' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
# 460: … 이 skill은 `auto` 모드 sync에서 … `.github/required-checks.yml` 존재 확인 …
```

`delivery.md:460`은 **ci-loop 스킬 설명 절 안에 있다** — M3가 제거하는 범위다.

| 참조처 | 제거 주체 |
|---|---|
| `ci-watch-protocol.md` | M3 (파일 자체 배포 중단) |
| ci-loop `SKILL.md` | M3 (디렉터리 삭제) |
| `zone-registry.md` | M2 (`014..021` 항목 삭제) |
| `delivery.md:460` | M3 (스킬 서술 제거) |

**결론**: 이 전제를 요구하는 **다른 배포 파일은 존재하지 않는다.**
따라서 결함은 본 SPEC 완료 시점에 잔여 참조 0으로 소멸한다.
소멸 여부는 AC-CLD-016이 판정하며, 잔여가 남으면 그때 별도 SPEC으로 승격한다.

### G.2 검증 엔진 버전 — 차이 없음 (마커 종결)

감사가 워크트리 소스로 빌드한 바이너리에서 `moai constitution validate`를 재실행한 결과,
설치된 v3.0.1 릴리스 바이너리와 **동일한 출력(전체 77 / ci 귀속 18)** 을 얻었다.
초판이 남긴 "엔진 버전 재확인 필요" 마커는 종결한다. acceptance.md의 baseline은 유효하다.

## §H 잔여 미해소 항목 (Gaps)

- **`moai init`으로 배포 결과를 직접 관측하지 못했다.** 스크래치 디렉터리에서
  `moai init . --yes`를 실행했으나 파일이 생성되지 않았다(비대화형 인자 불충분 추정).
  대신 임베드 FS를 직접 읽어 baseline을 확보했다(§D). 배포 경로 판정은 임베드 FS와
  catalog 해시 기준이며, 실제 `moai init` 배포 산출물을 관측한 판정은 이 SPEC에 없다.
- **`.github/required-checks.yml`의 사용자 측 생성 경로 자체는 여전히 미확인이다.**
  §G.1은 "배포 규칙이 더 이상 이 파일을 요구하지 않는다"를 확정했을 뿐,
  "사용자가 이 파일을 어떻게 얻는가"에는 답하지 않는다. 결정 B에 따라 이연한다.
