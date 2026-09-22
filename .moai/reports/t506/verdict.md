# t506 판정서 — SPEC-CODEX-GHOST-SKILLS-PRUNE-001 sync-phase close

카드: t506 · 트리 `.claude/worktrees/t506` · 브랜치 `WT-codex-ghost-skills`
sync 착수 시점 HEAD: `e27f00b5a` · 측정 일자: 2026-09-07
정책 인용: `.claude/rules/moai/core/verification-claim-integrity.md` §1/§2/§3 (5절 형식·baseline 귀속·도구 출처 §2.2), `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix / § Artifact Statelessness

---

## 1. Claim — 이 sync 단계가 주장하는 것

| # | 주장 |
|---|---|
| C1 | `CHANGELOG.md` `[Unreleased] > Added` 첫 항목으로 이 SPEC 의 항목이 하나 들어갔고, 중복은 없다 |
| C2 | 새 사용자 표면 `moai clean --codex-skills` 가 docs-site 4개 로케일과 README 4개 로케일에 문서화됐다 |
| C3 | docs-site 가 경고 없이 빌드된다 |
| C4 | SPEC 은 `in-progress → implemented → completed` 를 단일 sync 커밋에서 종결했고, 형제 산출물 3개는 status 축에서 stateless 라 전이 대상이 아니다 |
| C5 | sync 단계에서 `.go` 파일은 한 줄도 바뀌지 않았고, close 이후에도 착지하지 않는다 |
| C6 | 이 기계의 라이브 `~/.codex/config.toml` 은 sync 단계를 거치고도 바이트 그대로다 |
| C7 | MX Tag 검증을 sync 하위 단계로 수행했고, 새로 요구되는 태그가 없음을 측정으로 확인했다 |

---

## 2. Evidence — 명령과 그 출력 원문

### C1 — CHANGELOG (B12 자가검사 3건)

```
$ grep -c 'SPEC-CODEX-GHOST-SKILLS-PRUNE-001' CHANGELOG.md      # 방출 전
0

$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/acceptance.md | sort -u | wc -l
      17

$ ls internal/codexwiring/skills.go internal/codexwiring/skills_extent_test.go \
     internal/cli/clean.go internal/cli/codex_skills_prune.go \
     internal/cli/codex_skills_prune_test.go internal/cli/update_preserve_inventory.go
-rw-r--r--@ 1 goos  staff   6051 Sep  7 14:16 internal/cli/clean.go
-rw-r--r--@ 1 goos  staff   8384 Sep  7 14:37 internal/cli/codex_skills_prune.go
-rw-r--r--@ 1 goos  staff  21224 Sep  7 14:57 internal/cli/codex_skills_prune_test.go
-rw-r--r--@ 1 goos  staff  19734 Sep  7 14:17 internal/cli/update_preserve_inventory.go
-rw-r--r--@ 1 goos  staff  11103 Sep  7 14:12 internal/codexwiring/skills.go
-rw-r--r--@ 1 goos  staff   6972 Sep  7 14:11 internal/codexwiring/skills_extent_test.go
```

자가검사 (b) 의 **0 은 적색 신호**라는 규정에 따라 값을 확인했다 — 17 이고 `acceptance.md §D.3` 이 산문으로 적은 17, `progress.md §E.3` 의 `ac_pass_count: 17` 과 세 갈래가 일치한다. CHANGELOG 항목도 17 을 주장한다.

방출 후 위치:

```
$ sed -n '8,12p' CHANGELOG.md | cut -c1-80
## [Unreleased]

### Added

- **[SPEC-CODEX-GHOST-SKILLS-PRUNE-001](.moai/specs/SPEC-CODEX-GHOST-
```

### C2 — 4로케일 문서 동기화

```
$ wc -l docs-site/content/*/utility-commands/moai-clean.md
     289 docs-site/content/en/utility-commands/moai-clean.md
     289 docs-site/content/ja/utility-commands/moai-clean.md
     289 docs-site/content/ko/utility-commands/moai-clean.md
     289 docs-site/content/zh/utility-commands/moai-clean.md

$ for f in docs-site/content/*/utility-commands/moai-clean.md; do echo -n "$f "; grep -c '^##' "$f"; done
docs-site/content/en/utility-commands/moai-clean.md 23
docs-site/content/ja/utility-commands/moai-clean.md 23
docs-site/content/ko/utility-commands/moai-clean.md 23
docs-site/content/zh/utility-commands/moai-clean.md 23

$ for f in docs-site/content/*/utility-commands/moai-clean.md; do echo -n "$f "; grep -c 'codex-skills' "$f"; done
docs-site/content/en/utility-commands/moai-clean.md 4
docs-site/content/ja/utility-commands/moai-clean.md 4
docs-site/content/ko/utility-commands/moai-clean.md 4
docs-site/content/zh/utility-commands/moai-clean.md 4

$ grep -n 'moai clean \[--home\] \[--codex-skills\]' README.md README.ko.md README.ja.md README.zh.md | cut -c1-60
README.md:749:| `moai clean [--home] [--codex-skills]` | Clear left
README.ko.md:749:| `moai clean [--home] [--codex-skills]` | 오래된 실행 산출
README.zh.md:749:| `moai clean [--home] [--codex-skills]` | 清理旧的运行产物
README.ja.md:749:| `moai clean [--home] [--codex-skills]` | 古い実行成果物の
```

줄 수·`##` 헤딩 수·언급 수가 4로케일에서 동일하다. 4개 로케일이 **같은 줄 수**라는 것은 문서가 이미 그렇게 관리돼 온 성질이고, 이 카드가 추가한 블록도 그 성질을 깨지 않았다.

표 셀 안에 `|` 를 쓰면 GFM 렌더에서 셀이 갈라지므로 README 첫 열은 `[--home|--codex-skills]` 가 아니라 `[--home] [--codex-skills]` 로 적고, 스코프가 배타적이라는 사실은 설명 문장이 진다.

### C3 — hugo 빌드

```
$ cd docs-site && hugo --quiet
(출력 없음)
hugo exit: 0
```

### C4 — 상태 전이

```
$ grep -n 'status:\|updated:' .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/spec.md | head -3
5:status: completed
7:updated: 2026-09-07
```

전이 전 값은 `status: in-progress` (스크립트가 정확히 1회 매치를 단정한 뒤 치환). `updated:` 는 이미 sync 커밋 날짜인 2026-09-07 이므로 값 변경이 없다 — 갱신을 건너뛴 것이 아니라 이미 맞는 값이었다.

`progress.md §E.4` 를 sync-phase Audit-Ready Signal 로 채웠다(`sync_commit_sha: pending-backfill-sync` — 커밋이 자기 해시를 인용할 수 없으므로 스키마 독트린 D3 이 허용한 자리표시자이며, 후속 커밋에서 실제 SHA 로 백필한다. 빈 값으로 두는 것은 금지 — 빈 슬롯은 갚아야 할 빚을 아무에게도 알리지 않는다).

`plan.md` / `acceptance.md` / `progress.md` 는 frontmatter 의 status 축에서 **stateless** 다(`spec-frontmatter-schema.md` § Artifact Statelessness: "A SPEC's lifecycle state lives in exactly one place, `spec.md`"). 실제로 세 파일 모두 frontmatter 블록 자체가 없다:

```
$ head -3 .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/plan.md
# SPEC-CODEX-GHOST-SKILLS-PRUNE-001 — 구현 계획
```

따라서 "네 산출물 전부를 completed 로 전이"라는 지시는 이 SPEC 에서 **spec.md 1건의 전이 + progress.md §E.4 기록**으로 이행된다. 나머지 둘에 `status:` 를 새로 심는 것은 스키마 위반이므로 하지 않았다 — 지시의 문자 대신 실질을 택했고, 그 사실을 여기에 적어 되돌릴 수 있게 남긴다.

### C5 — sync 단계에 `.go` 변경 0

```
$ git status --porcelain
 M .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/progress.md
 M .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/spec.md
 M CHANGELOG.md
 M README.ja.md
 M README.ko.md
 M README.md
 M README.zh.md
 M docs-site/content/en/utility-commands/moai-clean.md
 M docs-site/content/ja/utility-commands/moai-clean.md
 M docs-site/content/ko/utility-commands/moai-clean.md
 M docs-site/content/zh/utility-commands/moai-clean.md
?? .moai/reports/t506/lane-live-dryrun.md
```

`.go` 는 한 건도 없다. close 이후 판정은 §5 에 적는다.

### C6 — 라이브 설정 불변 (세 번째 측정)

```
$ shasum -a 256 ~/.codex/config.toml
9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a  /Users/goos/.codex/config.toml

$ grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml
49
```

run 단계의 두 측정(첫 명령 직전 / 마지막 명령 직후)과 **같은 값**이다. 49개 항목이 그대로 있고, `spec.md §D` 의 첫 번째 Out of Scope("이 기계의 49건을 제거하지 않는다")가 sync 단계를 통과한 뒤에도 지켜졌다.

### C7 — MX Tag 검증 (sync 하위 단계)

```
$ grep -n '@MX:' internal/cli/codex_skills_prune.go internal/codexwiring/skills.go internal/cli/clean.go
internal/cli/codex_skills_prune.go:57:// @MX:WARN: [AUTO] deletion guard — every removal site MUST take its decision from this predicate, in the same pass that produced the candidate
internal/cli/codex_skills_prune.go:58:// @MX:REASON: [AUTO] a wrong "eligible" here deletes a registration the user wrote by hand, ...

$ grep -rn 'ParseSkillEntries' --include='*.go' internal/ | grep -v '_test.go' | grep -v 'skills.go'
internal/cli/codex_skills_prune.go:120:	entries := codexwiring.ParseSkillEntries(content)
internal/cli/doctor_codex.go:386:	entries := codexwiring.ParseSkillEntries(raw)
```

`@MX:ANCHOR` 의무 문턱은 fan_in ≥ 3 인데 `ParseSkillEntries` 의 실제 호출자는 **2개**(doctor 와 프루너)로 측정됐다 — 의무가 발생하지 않는다. 위험 지점(삭제 판정)에는 run 단계가 이미 `@MX:WARN` + `@MX:REASON` 을 달았다. 따라서 이 sync 단계에서 **추가한 태그는 0건**이며, 그것이 "검사를 건너뛰었다"가 아니라 "검사했고 요구가 없었다"임을 위 두 명령이 보인다.

### 부수 — spec lint (도구 출처 명시)

```
$ /tmp/t506-moai-run spec lint .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/spec.md
✓ No findings — all SPEC documents are valid
rc=0
```

**이 초록이 무엇을 덮지 않는지 명시한다(리드 지시 항목 5).** `moai spec lint` 의 `No findings` 는 이 SPEC 에 대해 **frontmatter 축과 REQ id 축만** 덮는다. `internal/spec/lint.go:790` 의 `isModalityMalformed` 는 `WHEN `/`WHILE `/`WHERE `/`IF `/`THE ` 라는 영문 접두사에만 반응하므로 한국어 REQ 본문에 대해 무조건 false 를 돌려주고, 그 결과 **GEARS 모달리티 축에서 이 초록은 공허하다.** 그 공허함이 실제로 무엇을 놓쳤는지도 기록돼 있다 — 린트가 `No findings` 를 내는 동안 plan-audit iter-1 은 MP-2 로 모달리티 위반 2건을 잡았다. 이 판정서는 린트 초록을 모달리티 근거로 **인용하지 않는다.**

**도구 출처(VCI §2.2).** 위 lint 는 이 트리에서 빌드된 바이너리(`/tmp/t506-moai-run`, `go build ./cmd/moai` — 레인이 `e27f00b5a` 에서 컴파일)로 돌렸다. 설치본 `~/go/bin/moai` 는 커밋 `e79c010b8`(2026-09-03) 로 이 트리보다 뒤처져 있어 판정 근거로 쓰지 않았다 — 실제로 그 설치본으로 시도한 무인자 `moai spec lint` 는 4분 넘게 반환하지 않아 배경으로 밀렸고, 그 실행의 출력은 이 판정서의 어느 주장도 뒷받침하지 않는다.

---

## 3. Baseline-attribution — 무엇에 대고 쟀는가

- 트리 좌표: `.claude/worktrees/t506`, 브랜치 `WT-codex-ghost-skills`, sync 착수 HEAD `e27f00b5a`. 위 명령은 전부 **이 실행에서 이 트리에 대고** 돌렸다.
- CHANGELOG 중복 검사(0)는 방출 **전에** 잰 값이다. 방출 후 grep 은 당연히 1 이 되므로, 0 이라는 값은 방출 전 시점에만 의미가 있다.
- AC 개수 17 은 `acceptance.md` 를 이 트리에서 직접 세어 얻었고, `progress.md §E.3` 의 `ac_pass_count: 17` 과 대조했다 — 어느 한쪽을 인용해 다른 쪽을 주장하지 않았다.
- 라이브 설정 sha256 은 run 단계가 기록한 값과 **같은 명령으로 이 sync 단계에서 다시 잰** 값이다. 건네받은 값을 재사용하지 않았다.
- lint 판정의 **두 좌표**를 모두 밝힌다: 읽힌 트리 `e27f00b5a`, 판정한 빌드 `e27f00b5a`(트리에서 컴파일).
- 실행 증거 정본: `.moai/reports/t506/run-evidence.md`(run 단계), `.moai/reports/t506/lane-live-dryrun.md`(레인 측 라이브 dry-run). 둘 다 이 커밋에 포함되므로 감사 시점에 경로가 해석된다.

---

## 4. Gaps — 관측하지 않은 것

- **`internal/cli` 패키지 커버리지 80.7% 는 85% 목표 미달이다.** 이 카드가 낮춘 것이 아니라 패키지의 기존 baseline 이며(run-evidence §7), 이 카드의 신규 판정 함수 2개는 100%다(`internal/codexwiring` 는 89.5%). 패키지 전체를 끌어올리는 것은 이 카드의 범위 밖이고, **이 sync 단계는 그것을 고치지 않았다.**
- **sync 단계에서 Go 테스트를 다시 돌리지 않았다.** `.go` 파일이 한 건도 바뀌지 않았으므로 재측정할 대상이 없다 — 다만 그 사실이 "테스트가 지금도 통과한다"를 이 단계에서 관측했다는 뜻은 아니다. 마지막 실측은 run 단계의 `go test -count=1 ./internal/cli/... ./internal/codexwiring/...` exit 0 이며, 전 패키지 판정은 CI 몫이다.
- **`--force` 쓰기 경로는 라이브 설정에 대해 여전히 미실행이다**(의도적 — t504/t502 의 관측 대상 보존). 쓰기 경로의 근거는 픽스처 테스트와 그 백업+sha256 단언뿐이다.
- **docs-site 렌더 결과를 눈으로 확인하지 않았다.** `hugo` 가 경고 없이 exit 0 을 냈다는 것은 빌드가 성공했다는 뜻이지, 새 절이 의도대로 보인다는 뜻이 아니다. 특히 4로케일 표의 셀 분할은 렌더를 열어 보지 않았다.
- **README 4로케일 이외의 표면은 훑지 않았다.** `moai clean` 을 언급하는 다른 문서(예: `docs-site/content/*/advanced/home-hygiene.md`, `cli-reference/doctor.md`)는 `--home` 스코프에 대한 서술이라 이 카드가 만지지 않았고, 그 판단이 옳은지 전수로 확인하지도 않았다.
- **번역 품질을 원어민이 검수하지 않았다.** ja/zh/ko 문장은 calque 를 피해 작성했으나, 그 판정은 이 세션의 것이다.
- **`moai spec lint` 전수 실행 결과가 없다.** 이 SPEC 파일 1건에 대해서만 돌렸다.

---

## 5. Residual-risk — 관측했음에도 여전히 틀릴 수 있는 것

- **`sync_commit_sha` 가 자리표시자로 남아 있다.** 백필 커밋이 실패하거나 잊히면 그 슬롯은 자리표시자인 채 커밋된다. 자리표시자는 최소한 갚아야 할 빚의 이름을 남기지만, 빚이 갚혔음을 보장하지는 않는다.
- **close 이후 `.go` 착지 금지는 이 커밋 시점의 성질이다.** 판정 명령 `git diff --name-only <close>..HEAD | grep '\.go$'` 이 무출력임을 §6 에 기록하지만, 이후 이 브랜치에 코드를 얹으면 그 불변은 깨진다. 다음 작성자에게 남기는 경계다.
- **문서가 구현을 앞지르거나 뒤처질 수 있다.** 이 판정서가 문서화한 일곱 부류·백업 형식·배타 스코프는 오늘의 `codex_skills_prune.go` 를 읽고 쓴 것이다. 파서의 다섯 갈래 분류가 넓어지면(run-evidence §8) 지금 문서가 "절대 지우지 않는다"고 적은 항목이 적격이 되고, 그 변화를 문서가 자동으로 따라가지는 않는다.
- **AC-CGP-010 가드는 모양 검사다.** `filepath.IsAbs` + `HasPrefix(p, "~` 조합을 찾으므로 파라미터 이름이 다른 두 번째 분류기는 빠져나간다(run-evidence §8). 이 sync 단계는 그 한계를 좁히지 않았다.
- **백업 파일명 형식이 기존 `~/.codex/` 의 3건과 다르다**(이 카드는 `config.toml.bak-<UTC>T<...>Z`, 기존은 로컬시 `-<YYYYMMDD>-<HHMMSS>`). 사용자는 두 형식을 나란히 보게 되고, 문서는 그 사실을 설명하지 않는다.

---

## 6. Close 이후 `.go` 미착지 검증

sync 커밋이 착지한 뒤 아래를 돌려 무출력을 확인한다(이 판정서를 담은 커밋이 close 커밋이다):

```
git diff --name-only <close-commit>..HEAD | grep '\.go$'
```

실행 결과는 최종 보고에 원문으로 옮긴다.

---

## 7. 판정

**PASS.** AC 17/17, must-pass 2건(AC-CGP-004 부재 판정 경계, AC-CGP-017 범위 내용물 경계) 모두 뮤턴트로 GREEN→RED→GREEN 관측. 문서 4로케일 동기화 + hugo 경고 없는 빌드, CHANGELOG 항목 1건(중복 0), SPEC 종결 완료. 미해결로 남는 것은 §4 의 커버리지 미달(카드 범위 밖의 기존 baseline)과 §5 의 잔여 위험이며, 어느 것도 이 카드의 인도물을 무효화하지 않는다.
