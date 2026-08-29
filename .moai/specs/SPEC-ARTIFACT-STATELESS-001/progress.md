# SPEC-ARTIFACT-STATELESS-001 — 진행 기록

카드: t357 · Tier M · 브랜치 `WT-tierl-status-transitions`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 계약. `design.md`/`research.md`는 Tier M 산출물이 아니며, 이 SPEC의 주제상으로도 만들지 않는다)
- 운영자 결정: **안 C(무상태 선언)**. 정의는 **D1(status 라인만 제거)**, 모집단은 **전 코퍼스 696**
- REQ 16 / AC 11 (Tier M 상한 각 16)

### 감사 이력

| 회차 | 판정 | 점수 | 결과 |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.86 (Tier M 임계 0.80) | blocking 6건(D1~D6) + optional D9 → v0.2.0에서 전부 반영 |
| 2 | PASS-WITH-DEBT | 0.92 (네 차원 단조 상승) | 1차 10건 중 9건 해소. 잔여 blocking 2건(N1 빈 슬롯 공허 PASS — 개정이 들여옴, R1 부정문 맹점) + minor N2~N4 → v0.3.0에서 반영, R1은 잔여 부채로 §5에 명시 |

감사 보고: `.moai/reports/t357/plan-audit.md` (1차, 감사 트리 `3b1830b96`) · `.moai/reports/t357/plan-audit-iter2.md` (2차, 감사 트리 `aacad4f99`)

### 측정 baseline — 두 트리가 섞여 있다

| 측정 | 트리 | 값 |
|---|---|---|
| 초기 코퍼스 실측 (696 / 633 / 362 / 417 / 170 / 106 / 12) | `c6aa61346` | `.moai/reports/t357/plan-measurement.md` |
| D1 전 코퍼스 (채택 모집단) | `3b1830b96` | **389** = design 27 / research 34 / plan 164 / acceptance 164 |
| D1 종결-한정 (참고값, 389의 부분집합) | `3b1830b96` | 362 |
| 템플릿 미러 바이트 동일성 | `3b1830b96` | 동일 (양쪽 23,317 bytes) |

재측정 명령: `bash .moai/reports/t357/t357_d1_all.sh .` · `bash .moai/reports/t357/t357_d1_by_artifact.sh .`

### AC 착지 전 FAIL 증거 (v0.2.0 개정 시 실행, 트리 `3b1830b96`)

아래는 **v1 스크립트의 출력**이며, 같은 경로의 스크립트는 v0.3.0에서 v2로 교체됐으므로 `3b1830b96`에서 다시 돌려도 이 인용문은 재현되지 않는다 — 기록이 틀린 것이 아니라 라벨에 스크립트 판번호가 빠져 있었을 뿐이다.

`bash .moai/reports/t357/t357_ac_precheck.sh .` 출력:

```
section bytes = 0
S1 spec.md-only:      FAIL
S2 four-artifacts+status: FAIL
S3 Tier-independent:  FAIL
P  permission stated: FAIL
N  blanket-prohibition matches = 0 (must be 0, AND P must be PASS)
mirror identical: yes
anchor in local: FAIL (section empty)
anchor in mirror: FAIL
```

`moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md` → `✓ No findings`, rc=0, `ArtifactStatusFieldForbidden` 매치 **0** (AC-04 비공허성 가드가 실제로 0을 낸다).

### 기준값 — run-phase에서 채운다

`acceptance.md`의 `AC-AST-001-07` / `-08` / `-10`이 이 표를 `sed`로 읽는다. 형식(`| <이름> | \`<값>\` |`)을 바꾸지 않는다.

| 시점 | 값 |
|---|---|
| SPEC 착수 직전 | `096aaf527` |
| M3 착수 직전 | `` |
| M3 착수 시 D1 baseline N | `` |

[HARD] **빈 슬롯은 AC를 통과시키지 않는다.** 추출은 `\{7,\}`(SHA) / `\{1,\}`(숫자) 패턴 + `-n` 검사 + `git rev-parse --verify` 3중 가드를 거치며, 빈 슬롯은 stderr에 `FAIL — 「…」 슬롯이 비어 있거나 …`를 내고 exit 1 한다. 가드 없이 `[0-9a-f]*`로 읽으면 빈 슬롯이 매치되어 빈 문자열을 캡처하고, `""..HEAD`가 `HEAD..HEAD`로 조용히 해석돼 AC가 공허하게 PASS한다 — iter-2 감사 N1이 지적한 결함이다.

- `SPEC 착수 직전` / `M3 착수 직전`: 7자리 이상 hex SHA
- `M3 착수 시 D1 baseline N`: `bash .moai/reports/t357/t357_d1_all.sh .` 의 「D1 전체 696 모집단」 값 (착수 시점 참고: 389 @ `3b1830b96`)

### 미해결 Gap

1. 카드가 인용한 `SPEC-AC-COUNT-DISCRIMINATOR-001`이 develop 코퍼스에 없어 원 사례 미재현
2. `origin/develop`이 `48d8ef4be`로 26 커밋 전진 — M3 착수 전 재측정 필수(REQ-AST-001-009)
3. 배포 사용자 코퍼스의 기존 위반 여부는 이 리포에서 측정 불가 — lint 심각도 결정의 잔여 위험(`plan.md` §B3/§D)
4. 362를 도출한 두 경로가 `fm_of` 추출기를 공유하므로 추출 단계 실패 모드는 교차검증되지 않음(spec.md §1.6)

다음: M2(재발 방지 lint) + M3(D1 코퍼스 정리) — 운영자 결정으로 함께 보류 중이다. M2만 먼저 착지시키면 코퍼스 잔여 위반이 남은 상태에서 `spec-lint.yml`이 develop push에 `--strict`로 돌고 `--strict`가 warning을 error로 승격시켜 develop SPEC Lint가 적색이 된다. 병합 큐가 빠지면 둘을 함께 재개한다.

## §E.2 Run-phase Evidence

### 기준 SHA

| 시점 | 값 |
|---|---|
| M1 착수 직전 (§E.1 「SPEC 착수 직전」 슬롯과 동일) | `096aaf527` |
| M3 착수 직전 | (미착수 — M3 미승인) |

> §E.1 「기준값」 표가 SHA의 SSOT다. 이 표는 같은 이름의 행을 만들지 않는다 — `AC-AST-001-10`의 `sed` 추출이 같은 이름을 두 번 매치하면 값이 두 줄이 되어 `git rev-parse --verify`가 실패한다.

### M1 — 규약 명문화 (착지)

브랜치 `WT-tierl-status-transitions`, 기준 HEAD `096aaf527`. **M1만** 수행했다. M2(lint 규칙)·M3(코퍼스 정리)는 이 실행에서 승인되지 않았으므로 착수하지 않았고, `.moai/specs/` 아래에서 이 SPEC의 `progress.md` 외에는 어떤 파일도 건드리지 않았다.

변경 파일 2개 (두 벌이 바이트 동일):

| 파일 | 변경 |
|---|---|
| `.claude/rules/moai/development/spec-frontmatter-schema.md` | `## Canonical 12 Required Fields` 바로 아래, `## Field Reference` 직전에 `### Artifact Statelessness` 소절 신설 |
| `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` | 위와 동일 내용을 같은 위치에 미러 (Template-First) |

바이트: 양쪽 23,317 → **24,677** (+1,360). 소절 본문 1,327 bytes.

템플릿 중립성: 신설 소절에 SPEC ID·REQ 토큰·내부 날짜·커밋 SHA·로컬 문서 참조가 없다. 따라서 로컬본과 미러본을 갈라 쓸 필요가 없었고 두 벌은 바이트 동일하다.

`spec-frontmatter-schema.md`는 `paths: "**/.moai/specs/**,internal/spec/**"` 로 범위가 잡힌 파일이라 always-loaded 표면이 아니다 — `rule-authoring.md`의 1,000-byte 성장 진술 의무는 발동하지 않는다.

### 검증 — 명령과 관측 출력

모든 측정은 트리 `096aaf527`(작업 트리, 위 2파일 수정 상태)에서 이 실행 중에 수행했다.

**(1) AC 판정 — `bash .moai/reports/t357/t357_ac_precheck.sh .`**

착지 **전** (동일 명령, 미수정 트리):

```
HEAD=096aaf527
section bytes = 0
--- AC-AST-001-01 (three separate statements + negation proximity) ---
S1 FAIL
S2 FAIL
S3 FAIL
--- AC-AST-001-02 (permission present + not negated, no blanket prohibition) ---
permission=0 blanket_prohibition=0
AC-02 FAIL
--- AC-AST-001-11 (template mirror) ---
mirror identical: PASS
anchor local:  FAIL
anchor mirror: FAIL
```

착지 **후**:

```
HEAD=096aaf527
section bytes = 1327
--- AC-AST-001-01 (three separate statements + negation proximity) ---
S1 PASS
S2 PASS
S3 PASS
--- AC-AST-001-02 (permission present + not negated, no blanket prohibition) ---
permission=1 blanket_prohibition=0
AC-02 PASS
--- AC-AST-001-11 (template mirror) ---
mirror identical: PASS
anchor local:  PASS
anchor mirror: PASS
```

`AC-AST-001-07` / `-08` / `-10`의 기준값 3중 가드는 이 실행에서도 계속 **요란하게 FAIL**한다 — 의도된 상태다. M3가 승인되지 않았으므로 `M3 착수 직전` / `M3 착수 시 D1 baseline N` 슬롯은 비워 둔다:

```
--- 3-guard base-value extraction (AC-07 / -08 / -10) ---
BASE_SPEC=096aaf527 (rc=0)
FAIL — 「M3 착수 직전」 slot empty or shorter than 7
AC-07/-08 FAIL — no base SHA (rc=1)
FAIL — 「M3 착수 시 D1 baseline N」 slot empty or not a number
AC-07 FAIL — no baseline N (rc=1)
```

**(2) 미러 정합 — `diff -q`**

```
$ diff -q .claude/rules/moai/development/spec-frontmatter-schema.md \
          internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
(무출력) rc=0
$ wc -c (양쪽)
   24677 .claude/rules/moai/development/spec-frontmatter-schema.md
   24677 internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
```

**(3) SPEC lint — `moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md`**

```
✓ No findings — all SPEC documents are valid
rc=0
```

**(4) 템플릿 임베드 — `make build` → `make embed-check`**

`make build`: `agents-emit-check` 통과(`ok github.com/modu-ai/moai-adk/internal/template/agentemit 0.455s`) → `templ generate` → catalog 해시 재계산(45 entries, `catalog.yaml updated successfully (12899 bytes)`) → `go build -ldflags … -o bin/moai ./cmd/moai`. 오류 없음.

`catalog.yaml`은 재계산 후에도 **내용이 바뀌지 않았다** — 규칙 파일은 catalog 엔트리가 아니다. `git status --short`가 이 실행 내내 수정 파일 2개(+ 이 `progress.md`)만 보고한다.

```
$ make embed-check
  ✓ Agent Emit Embed
  ok  Agent Emit Embed  11/11 embedded agent-emit artifacts match the committed set (moai)
  Pass 1    Warn 0    Fail 0
```

**(5) Go 트리 — `go build ./... && go vet ./...`**

```
BUILD_VET_OK
```

Go 소스 변경은 없다. 임베드 FS 내용만 바뀌므로 바이너리 재컴파일 여부만 확인한 것이다.

### 소절 원문 (인간 판독용 — DoD 항목)

```markdown
### Artifact Statelessness

The canonical 12-field obligation above binds `spec.md` only.
Every other artifact in a SPEC directory is governed by this sub-section instead.

The four sibling artifacts — `plan.md`, `acceptance.md`, `design.md`, and `research.md` — are **stateless on the status axis**: they MUST NOT carry a `status:` field in their YAML frontmatter. A SPEC's lifecycle state lives in exactly one place, `spec.md`, which is the file every lint, audit, and close path reads it from.

Frontmatter itself is permitted in those four artifacts.
Statelessness is confined to the status axis: fields such as `id`, `title`, `version`, and `created` are outside the scope of this rule, and an artifact may carry a frontmatter block or omit one entirely, as its author prefers. Reading statelessness as a blanket ban on frontmatter is a misreading of the axis this rule governs.

This declaration is Tier-independent.
It holds for every SPEC directory whatever the `tier:` value says, so a SPEC that carries `design.md` or `research.md` while declaring a Tier other than L falls under exactly the same rule as one that declares `tier: L`.

Two files sit outside this sub-section: `spec.md`, which carries the canonical `status:` field per the schema above, and `progress.md`, which records phase progress in body sections rather than in frontmatter.
```

세 고정 리터럴은 각각 **자기 줄에서 평서 단언**으로 쓰였다. 이것은 우연이 아니라 설계다 — 부정어 근접 검사가 줄 단위이므로, 리터럴을 담은 문장을 후속 설명 문장과 같은 줄에 두면 뒤 문장의 어휘(예: `without`, `not`)가 앞 문장의 극성을 오염시킬 수 있다. 그래서 리터럴 문장과 그 부연을 **줄로 분리**했다(마크다운 렌더는 한 문단으로 합쳐진다).

### Gap — 이 실행이 관측하지 않은 것

1. **M2·M3 미착수.** `AC-AST-001-03` ~ `-10`은 이 실행이 판정하지 않았다. `AC-AST-001-09`(lint와 정리의 동시 착지)는 아직 성립하지 않으며, 그것이 `plan.md` §B2가 "M3를 미루면 era 예외가 필수"라고 적은 상태다 — 현재 SPEC은 그 미결 상태에 있다.
2. **임베드 축의 대상 한정.** `make embed-check`가 대조하는 것은 `.codex` 에이전트 TOML 11개다. 이 소절이 실린 규칙 파일이 바이너리에 임베드된 내용과 일치하는지를 직접 대조하는 검사는 이 리포에 없다 — 방금 빌드한 바이너리는 정의상 일치하므로 실질 위험은 낮으나, **관측하지 않았다.**
3. **여러 줄 부정은 여전히 기계로 닫히지 않는다**(spec.md §5 부채). 위 원문 인용이 그 방어의 전부다.
4. **배포 사용자 코퍼스**에 대한 영향은 이 리포에서 잴 수 없다(불변).

### M2 — 재발 방지 lint (착지)

브랜치 `WT-tierl-status-transitions`, 기준 HEAD `22232bff5`(= 이 실행 직전 브랜치 tip, `origin/develop` `a6bbbf82b`를 흡수한 머지 커밋). M1 착지 이후 M2만 수행했다 — M3(코퍼스 정리)는 이 실행에서 착수하지 않았고, `.moai/specs/` 아래에서 이 SPEC의 `progress.md` 외에는 어떤 파일도 편집하지 않았다(AC-04가 심었다 원복한 이 SPEC의 `plan.md`는 아래 4항에서 `git diff --stat` 빈 출력으로 확인).

변경 파일 3개(+ 픽스처 11개):

| 파일 | 변경 |
|---|---|
| `internal/spec/lint_artifact_status.go` | 신규 — `ArtifactStatusFieldForbiddenRule` + `frontmatterStatusLine` |
| `internal/spec/lint.go` | 규칙 배열에 `&ArtifactStatusFieldForbiddenRule{}` 1줄 등록 (+ 근거 주석) |
| `internal/spec/lint_artifact_status_test.go` | 신규 — 5개 판정, 각각 자기를 빨갛게 만드는 뮤테이션을 주석에 명시 |
| `internal/spec/testdata/artifactstatus/**` | 신규 픽스처 5 케이스(11 파일) |

#### 술어 — 세는 명령보다 넓다 (의도된 비대칭, 측정으로 뒷받침)

`acceptance.md` 스니펫 (B)의 계수 술어는 `^status:[[:space:]]` 라 콜론 뒤 공백을 요구한다. lint 규칙은 `status:` 접두만 본다 — 즉 **계수 술어의 진부분집합이 아니라 상위집합**이다. `status:draft`(공백 없음)와 값 없는 `status:`가 그 차분이다.

이 방향의 비대칭만 안전하다. 검사가 정리보다 **좁으면** 재발이 새고, **넓으면** 정리가 치우지 못한 잔여가 남아 코퍼스가 빨개진다 — 그래서 M3 정리도 같은 넓은 술어를 쓴다. 오늘의 차분이 0임을 가정하지 않고 쟀다:

```bash
$ grep -rnE '^status:[^ 	]' --include=plan.md --include=acceptance.md \
    --include=design.md --include=research.md .moai/specs/
(무출력)
```

즉 두 술어는 현재 코퍼스에서 **같은 집합을 고른다**. 갈라지는 것은 앞으로 누군가 쓸 줄에서뿐이고, 재발 방지 규칙이 잡아야 할 것이 바로 그것이다.

#### 뮤테이션 — 심고, 빨간 것을 보고, 되돌렸다

판정 5개 전부 "존재하지만 아무것도 안 보는 검사"가 아님을 뮤테이션으로 보였다. 각 줄은 실제 실행 출력이다.

| # | 뮤테이션 | 관측된 RED |
|---|---|---|
| 1 | 픽스처의 `status:` 키를 `state:`로 개명 | `TestArtifactStatus_FiresOnPlanStatus`: `findings = 0, want 1` |
| 2 | "블록은 1행에서만 열린다" 가드 제거 | `TestArtifactStatus_IgnoresBodyText`: `findings = 1, want 0` — `no-status/acceptance.md` 7행의 본문 `status: draft`를 frontmatter 필드로 오독 |
| 3 | `statelessArtifacts`에서 `research.md` 제거 | `TestArtifactStatus_FiresOnAllFourArtifacts`: `findings = 3, want 4` |
| 3b | 콜론 뒤 공백 요구(계수 술어와 동일하게 좁힘) | 같은 판정: `findings = 3, want 4` — `design.md`의 `status:draft`가 빠짐 |
| 4 | `eraDemotableCodes`에 코드 추가 | `TestArtifactStatus_SurvivesEraDemotion`: `Severity = "warning", want "error"` + `finding is Advisory` |

뮤테이션 4는 특히 **판정의 비공허성 자체**를 증명한다: `grandfathered` 픽스처가 실제로 grandfather-era로 분류되지 않았다면 이 뮤테이션은 아무 효과가 없었을 것이고, 판정은 조용히 통과했을 것이다. warning + Advisory 로 뒤집힌 것이 그 픽스처가 `applyEraDemotion` 경로를 실제로 통과한다는 증거다.

전부 원복 후 재실행:

```
$ go test ./internal/spec/ -run 'TestArtifactStatus' -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	1.528s
$ gofmt -l internal/spec/lint.go internal/spec/lint_artifact_status.go internal/spec/lint_artifact_status_test.go
(무출력)
```

#### AC 판정 — 명령과 관측 출력

**AC-AST-001-03 (규칙 등록 + era 예외 부재) — PASS**

```
$ grep -c 'ArtifactStatusFieldForbidden' internal/spec/lint.go
2
$ sed -n '/^var eraDemotableCodes/,/^}/p' internal/spec/lint.go
var eraDemotableCodes = map[string]bool{
	"MissingExclusions":  true,
	"FrontmatterInvalid": true,
}
```

계수 2 ≥ 1이고 era 블록에 코드가 없다. 착지 전 실측은 0이었다(§E.1).

**AC-AST-001-04 (심어서 거부 관측) — PASS, 네 조건 전부**

lint 호출은 이 SPEC 하나로 한정했다(`plan.md` §B2 — 전 코퍼스 호출은 M3 미착지 상태에서 코퍼스 잔여에 걸려 정상 lint에도 FAIL한다). 바이너리는 이 워크트리에서 방금 빌드한 `./bin/moai`다 — PATH의 설치본은 이 규칙을 모른다.

```
$ ./bin/moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md > …/t357_ac04_before.txt 2>&1; echo "before rc=$?"
before rc=0
$ grep -c 'ArtifactStatusFieldForbidden' …/t357_ac04_before.txt
0
$ cat …/t357_ac04_before.txt
✓ No findings — all SPEC documents are valid
```

심은 뒤:

```
$ ./bin/moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md > …/t357_ac04_after.txt 2>&1; echo "after rc=$?"
after rc=1
$ cat …/t357_ac04_after.txt
SEVERITY  CODE                          FILE                                             LINE  MESSAGE
--------  ----                          ----                                             ----  -------
ERROR     ArtifactStatusFieldForbidden  .moai/specs/SPEC-ARTIFACT-STATELESS-001/plan.md  2     `plan.md` carries `status: draft` in its frontmatter. …

1 error(s), 0 warning(s)
```

원복:

```
$ cp .moai/cache/t357_ac04_backup.md .moai/specs/SPEC-ARTIFACT-STATELESS-001/plan.md
$ git diff --stat -- .moai/specs/SPEC-ARTIFACT-STATELESS-001/plan.md
(빈 출력, rc=0)
```

1항 before=0 · 2항 after 1건 · 3항 after rc=1 비-0 · 4항 원복 후 빈 diff — 넷 다 성립.

증거 파일: `.moai/reports/t357/t357_ac04_before.txt` · `t357_ac04_after.txt`.

#### 검증 — 패키지 스위트

```
$ go test ./internal/spec/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	32.867s
```

전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다(`CLAUDE.local.md` §4 — 병렬 레인 부하 사고). 판정은 PR CI 몫이다. `internal/cli`는 `moai spec lint` CLI가 이 패키지를 소비하므로 함께 돌렸다 — 결과는 아래 Gap 항목 참조.

#### 전 코퍼스 실측 — 두 술어가 1건 어긋나고, 그 1건의 원인을 규명했다

M3 착수 전 참고 측정이다(REQ-AST-001-009의 정식 재측정은 develop 흡수 후에 다시 한다). 트리 `22232bff5`:

```bash
$ bash .moai/reports/t357/t357_d1_all.sh .
HEAD=22232bff5
D1 전체 696 모집단 = 393
D1 종결(633) 모집단 = 366

$ ./bin/moai spec lint > .moai/reports/t357/t357_m2_corpus_lint.txt 2>&1; echo "corpus rc=$?"
corpus rc=1
$ grep -c 'ArtifactStatusFieldForbidden' .moai/reports/t357/t357_m2_corpus_lint.txt
392
```

**393 ≠ 392.** lint가 계수보다 **하나 적다** — 술어를 넓혔으므로 기대와 반대 방향이라 그냥 지나칠 수 없었다. 두 목록을 뽑아 맞춰 보니 차분은 정확히 한 파일이다:

```bash
$ comm -3 <(정렬된 count_d1 목록) <(정렬된 lint 목록)
.moai/specs/SPEC-V3R4-CC2X-ADOPT-001/research.md
$ ls .moai/specs/SPEC-V3R4-CC2X-ADOPT-001/
research.md          # spec.md 가 없다
```

원인은 술어가 아니라 **발견 경로**다. `discoverSPECs`(`lint.go:307,328`)는 `SPEC-*/spec.md` 패턴으로 SPEC을 찾으므로, `spec.md`가 없는 디렉터리는 lint가 애초에 방문하지 않는다 — 규칙이 그 파일을 통과시킨 것이 아니라 **보지 못한** 것이다. 반면 `count_d1`은 디렉터리를 직접 순회하므로 본다.

이것이 M3나 AC 판정을 깨지는 않는다. 정리(M3)도 디렉터리 순회형이라 이 파일을 포함하고, 정리 후 두 잔여는 모두 0이 된다. 다만 **재발 방지 가드에는 구조적 사각지대가 남는다** — 아래 Gap 5.

원자료: `.moai/reports/t357/t357_d1_files.txt`(393행) · `t357_lint_files.txt`(392행) · `t357_m2_corpus_lint.txt`.

또한 이 측정에서 **카드가 인용한 원 사례 `SPEC-AC-COUNT-DISCRIMINATOR-001`이 이 트리에 존재한다**(4개 산출물 전부가 D1 대상). SPEC §5가 "develop 코퍼스에 없다"고 적은 것은 그 시점 사실이었고, t338 병합으로 들어왔다. 원 사례가 이제 정리 대상에 포함된다 — 재현하지 않겠다는 스코프 제외는 유효하되(그 SPEC을 판정 근거로 삼지 않는다), 사실 관계는 갱신됐다.

### Gap — M2 실행이 관측하지 않은 것

1. **코퍼스는 지금 빨갛다 — 의도된 중간 상태.** `ArtifactStatusFieldForbidden`이 error 심각도로 켜졌고 D1 정리(M3)는 아직 착수하지 않았으므로, 전 코퍼스 lint는 지금 다수의 error를 낸다. `plan.md` §B2가 적은 그대로이며, **M2와 M3가 같은 브랜치에서 함께 착지해야** 해소된다. 이 상태로 develop에 넣으면 안 된다.
2. **AC-05·-06·-09 미판정.** 셋 다 전 코퍼스를 보므로 M3 착지 후에 판정한다(`acceptance.md` AC-05 주석).
3. **배포 사용자 코퍼스 영향 미측정**(불변) — error 심각도의 잔여 위험(`plan.md` §B3/§D).
4. **`lint.skip` 상호작용 미관측.** 이 규칙은 per-SPEC 규칙이라 `lint.skip: [ArtifactStatusFieldForbidden]`으로 개별 SPEC이 빠져나갈 수 있다. 기존 규칙과 같은 성질이고 의도된 것이지만, 이 실행이 그 경로를 실제로 돌려보지는 않았다.
5. **`spec.md` 없는 SPEC 디렉터리는 가드의 사각지대다.** 위에서 규명한 393 vs 392의 원인이다. `discoverSPECs`가 `SPEC-*/spec.md`로 SPEC을 찾으므로, `spec.md` 없이 산출물만 있는 디렉터리(현재 `SPEC-V3R4-CC2X-ADOPT-001` 1건)에는 이 규칙이 도달하지 않는다. M3 정리는 디렉터리 순회형이라 이번엔 치우지만, **그런 디렉터리에 다시 `status:`가 생기면 lint는 잡지 못한다.** 이 SPEC은 그것을 고치지 않는다 — 발견 경로를 바꾸는 것은 `discoverSPECs`의 계약 변경이고 모든 규칙에 영향을 주므로 별도 카드 소관이다. 여기서는 관측 기록으로만 남긴다.
6. **이 워크트리 바이너리로만 관측했다.** AC-04와 코퍼스 측정은 `./bin/moai`(이 트리에서 `go build`로 방금 만든 것)로 돌렸다. PATH의 설치본은 이 규칙을 모르므로 같은 명령이 다른 답을 낸다 — 재현 시 바이너리를 확인할 것.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
