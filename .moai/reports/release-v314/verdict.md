# release/v3.1.4 브랜치 준비 — 검증 보고

작성: lane-4 (session `1a37e6e9-f4fd-4411-abd4-c3c4eeb338be`), 2026-08-31
워크트리: `.claude/worktrees/release-v314`
브랜치: `release/v3.1.4`
base: `origin/develop = 9328a52422baa13fdb7d7fd0c8409151da3ba3c1` (푸시 직전 재fetch 후 동일 확인)

> **배포 아님.** t204 게이트에 따라 `scripts/release.sh` · `git tag` · GoReleaser 모두 미실행.
> 이 브랜치는 PR 대기 상태이며 PR 생성은 리드 소유.

---

## 왜 v3.1.3 이 아니라 v3.1.4 인가

`release/v3.1.3` 은 이미 존재하며 PR #1602 로 **2026-08-23 main 에 병합**됐다.
그러나 태그는 없다(`git tag` 최신 = `v3.1.2`) — 병합됐고 배포만 보류된 상태다.

```
origin/release/v3.1.3 = b37e86b64
merge-base --is-ancestor release/v3.1.3 origin/main    → YES
merge-base --is-ancestor release/v3.1.3 origin/develop → YES
rev-list --count release/v3.1.3..origin/develop        → 719
```

버전 범프(`eba919e44`)와 CHANGELOG `[3.1.3]` 섹션이 이미 착지해 있어,
같은 번호를 재사용하면 하나의 버전이 서로 다른 두 내용을 가리키게 된다.
**운영자 판정으로 이번 배치를 v3.1.4 로 끊고, 태그 없는 3.1.3 은 결번으로 남긴다.**

---

## Claim

`release/v3.1.4` 브랜치가 `origin/develop@9328a5242` 위에 준비됐다.
버전 스탬프 7파일 9줄 범프, CHANGELOG 98항목 승격(순수 삽입), 재측정 4종 전부 통과.

## Evidence

### 커밋 2건

```
c11efe37b docs: update CHANGELOG for v3.1.4
61921f1ba chore: bump version to v3.1.4
9328a5242 Merge WT-state-evidence-canon: SPEC-EVIDENCE-CITATION-CANON-001 (card t375)   ← base
```

### 작업 1 — 버전 스탬프 전수 (v3.1.3 → v3.1.4)

`git diff --stat` 기준 **7파일 9줄**:

| 파일 | 줄 | 근거 |
|---|---|---|
| `pkg/version/version.go` | 8 | 범프 커밋 `eba919e44` 선례 |
| `.moai/config/sections/system.yaml` | 45 (`template_version`), 47 (`version`) | 동일 |
| `README.md` | 24 (Release 배지) | 동일 |
| `README.ko.md` | 24 | 동일 |
| `README.ja.md` | 24 | 동일 |
| `README.zh.md` | 24 | 동일 |
| `docs-site/hugo.toml` | 55 (`version`), 56 (`releaseDate` → 2026-08-31) | **선례 밖 — 아래 참조** |

**`docs-site/hugo.toml` 은 범프 커밋 선례에 없다.** 포함시킨 근거는 그 파일 자신의 주석이다:

```toml
  # SSOT — Version 갱신은 이 두 줄만 동시에 수정
  version = "v3.1.3"
  releaseDate = "2026-08-24"
```

스스로를 버전 SSOT 로 선언하고 있고, v3.1.3 때는 범프 커밋이 놓쳐 별도 카드
(`SPEC-DOCS-V313-CATCHUP-001`, 카드 t274)가 뒤늦게 고쳤다(CHANGELOG:100 에 기록).
같은 누락을 반복하지 않으려 이번엔 범프에 포함했다. **선례를 한 파일 넘어선 확장이므로 명시한다.**

#### 바꾸지 않은 `v3.1.3` 출현 — 전수 분류

버전 스탬프가 **아닌** 것들이며 의도적으로 보존했다:

| 위치 | 성격 |
|---|---|
| `README.{md,ko,ja,zh}.md:503,505,781` | statusline 예시 출력 (`🗿 v3.1.2 -> 🗿 v3.1.3` 은 업데이트 지시자 예시) — 범프 커밋도 24행만 건드렸다 |
| `internal/cli/gate_summary_cli_test.go:86` | 테스트 뮤턴트 리터럴 |
| `internal/cli/mcp_project_root_test.go:38,259` | 테스트 픽스처 frontmatter |
| `internal/binlag/binlag_test.go:154` | 테스트 픽스처 (`v3.1.3-rc.5`) |
| `docs-site/content/{ja,zh,...}/advanced/codex-dual-harness.md:5` | `added_in: "v3.1.3"` — 정확한 이력 |
| `docs-site/content/*/getting-started/faq.md` | 업데이트 예시 |
| `docs-site/content/*/guides/claude-cloud.md:67` | `go install ...@v3.1.3` 설치 예시 |
| `CHANGELOG.md` `[3.1.3]` 섹션 및 본문 언급 | 이력 |

검증: 범프 후 `grep -c 'v3.1.3'` → `pkg/version/version.go:0`, `system.yaml:0`, `hugo.toml:0`.

#### `.moai/docs/version-management.md` 는 근거로 쓰지 않았다

해당 문서의 "Files Requiring Version Sync"(71-78행) 목록이 부정확하다:

- `internal/template/templates/.moai/config/config.yaml` — **origin/develop 에 존재하지 않음**
  (`git show origin/develop:<path>` → `fatal: path ... does not exist`)
- 실제 범프 대상인 `README.ja.md` / `README.zh.md` / `docs-site/hugo.toml` 이 목록에 **없음**

범프의 기준은 범프 커밋 `eba919e44` 의 실제 diff + 위 전수 스윕이다. 문서 수정은 이 카드 범위 밖이며 리드에 보고했다.

### 작업 2 — CHANGELOG 승격

`## [Unreleased]` 아래 98항목을 `## [3.1.4] - 2026-08-31` 로 승격.
`[Unreleased]` 는 빈 섹션으로 상단에 유지(Keep a Changelog 규약, 파일 헤더가 따른다고 명시).

**diff 는 `+2` 삽입, 삭제 0** — 항목을 지우거나 재배열하지 않았음이 diff 로 증명된다:

```
$ git diff --stat CHANGELOG.md
 CHANGELOG.md | 2 ++
 1 file changed, 2 insertions(+)
```

항목 수 전후 (`grep -c '^- '`):

| 측정 | 전 | 후 |
|---|---|---|
| 파일 전체 | 338 | **338** |
| `[Unreleased]` 구간 | 98 | **0** |
| `[3.1.4]` 구간 | — | **98** |
| `[3.1.3]` 구간 | 26 | **26** (불변) |

구간 계수 명령: `awk '/^## \[<헤딩>\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md | grep -c '^- '`

기존 `## [3.1.3] - 2026-08-24` 는 **건드리지 않았다**. 헤딩 위치만 404 → 406 으로
삽입된 2줄만큼 밀렸다. 중복 버전 헤딩 없음:

```
8:## [Unreleased]
10:## [3.1.4] - 2026-08-31
406:## [3.1.3] - 2026-08-24
536:## [3.1.2] - 2026-08-21
```

### 작업 3 — 재측정 4종

| # | 명령 | 결과 |
|---|---|---|
| 1 | `make build` | **rc=0** |
| 2 | `strings bin/moai \| grep -c c11efe37b` | **4** (트리 귀속 양성) |
| 3 | `./bin/moai spec lint` | **rc=0 — `0 error(s), 1096 warning(s)`** (기대값 일치) |
| 4 | `GOOS=windows go build ./...` | **rc=0** |

빌드가 주입한 ldflags (트리 귀속의 출처):
```
-X pkg/version.Commit=c11efe37b -X pkg/version.BuildID=v3.1.2-956-gc11efe37b
```

`make build` 후 `git status --short` **무출력** — `catalog.yaml` 재계산이 45개 항목 전부
기존 해시와 동일해 생성물 드리프트 없음.

#### [주의] 빌드된 바이너리는 `v3.1.2` 를 보고한다 — 결함이 아님

```
$ ./bin/moai version
moai-adk v3.1.2
```

`Makefile` 의 `LDFLAGS` 가 `git describe` 로 버전을 주입하는데, 태그가 `v3.1.2` 가 최신이라
그 값이 박힌다. `pkg/version/version.go` 의 `v3.1.4` 는 ldflags 가 덮는 **RC/테스트용 기본값**이다.
**태그가 붙기 전까지 `moai version` 으로 범프를 검증할 수 없다** — 검증은 소스 스탬프로 해야 한다.
t204 배포 보류 상태에서는 정상이며, 태그 시점에 자동 해소된다.

## Baseline-attribution

전부 이번 실행에서 이 워크트리(`.claude/worktrees/release-v314`)에 대고 측정.
base `9328a5242`, 측정 시점 HEAD `c11efe37b`.
`spec lint` 로그: 이 저장소 워크트리 경로를 스캔한 것이며 primary 가 아니다(로그의 경로 접두사로 확인 가능).

## 배치 규모 실측 (다음 릴리스 분할 판단용)

`origin/main = 48239c7dc` ↔ `origin/develop = 9328a5242` 기준 — **655 커밋 / 1,457 파일**.
(리드의 선행 측정 644/1,431 은 그 사이 develop 이 움직여 낡은 값이다.)

| 경로 | 파일 수 |
|---|---|
| `.moai/` | 1,010 (69.3%) — specs 564 + reports 431 + 그 외 15 |
| `internal/` | 387 (26.6%) |
| `.claude/` | 24 · `docs-site/` 12 · `.github/` 9 · `scripts/` 4 · README 4 |

확장자별 소재:

- **`.txt` 172개 → 171개가 `.moai/reports/`** (99.4% 가 증거 산출물), 1개는 `scripts/ci-census/`
- **`.md` 873개 → 748개(85.7%)가 `.moai/specs`(564) + `.moai/reports`(184)**.
  나머지 125개: `internal/spec` 픽스처 54, `internal/template` 미러 21, `docs-site/content` 12,
  `.claude/rules` 10, `.moai/project` 6, `.claude/skills` 6, `.claude/agents` 5, `.moai/docs` 3, README 4
- **`.go` 241개 → `_test.go` 150개(62.2%) / 프로덕션 91개(37.8%)**, `testdata/` 0.
  패키지: `internal/cli` 74 · `kanban` 39 · `hook` 30 · `spec` 21 · `guardstate` 15 ·
  `guardliveness` 14 · `web` 8 · `statusline` 7 · `mx`/`graph`/`config` 각 5 · `astgrep` 4

### 결론 — 증거 분리만으로는 리뷰 상한 아래로 못 내려간다

전체 1,457파일 중 **실제 프로덕션 로직은 91개(6.2%)** 다.
그러나 증거·SPEC 산출물 995개(`.moai/specs` + `.moai/reports`)를 **전부 제외해도 462파일**로,
CodeRabbit 150파일 상한의 **3배**다. 즉 다음 릴리스에서 분할을 검토한다면
"증거를 별도 PR 로 분리" 는 해법이 되지 않으며, 로직 변경 자체를 나눠야 한다.

(150 상한 수치는 리드가 t247 실측에서 옮긴 값이며 이 카드가 잰 것이 아니다.)

## Gaps — 관측하지 않은 것

- **테스트 스위트 미실행.** `go test ./...` 는 로컬에서 돌리지 않았다(CLAUDE.local.md §4 —
  레인 동시 실행이 머신을 마비시킨 2026-08-15 사고). 전 패키지 판정은 CI 몫이다.
- **`golangci-lint` 미실행.** 배차의 재측정 4종에 없었다.
- **PR 미생성, main 머지 없음, 태그·배포 없음** — 전부 범위 밖(리드/운영자 소유).
- **719 커밋의 개별 내용 미판독.** 파일 단위 분류만 수행했다.
- **`docs-site` hugo 빌드 미실행.** `hugo.toml` 을 건드렸으나 사이트 빌드로 검증하지 않았다 —
  변경이 `params` 두 줄의 문자열이라 구조 변경이 아니지만, **관측하지 않았다는 사실은 남긴다.**

## Residual-risk

- `docs-site/hugo.toml` 범프는 선례 밖 판단이다. 릴리스 프로세스가 이 파일을 다른 시점에
  갱신하도록 설계돼 있다면 중복 갱신이 된다. 그렇게 설계된 증거는 못 찾았다
  (`.github/workflows/`·`scripts/`·`internal/` 에서 `hugo.toml` 참조 0건).
- 파일 분류는 **경로·확장자 기준**이다. `.moai/reports/` 안에 증거가 아닌 파일이 섞여 있을
  가능성은 배제하지 않았다.
- 655커밋/1,457파일은 `9328a5242` 시점 값이다. 다른 레인이 develop 에 착지하면 즉시 낡는다.
  이 브랜치는 그 base 에 고정되므로 이후 착지분은 다음 릴리스 몫이다.
- CHANGELOG `[3.1.4]` 섹션에 Summary 산문이 없다. v3.1.3 때는 별도로 작성됐다
  (`b4b8bdfbe` 가 178줄 추가). 배차에 없었으므로 쓰지 않았다 — 필요하면 별도 작업이다.
