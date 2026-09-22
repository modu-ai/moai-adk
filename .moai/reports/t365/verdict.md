# t365 판정서 — discoverSPECs 발견 경로 blind spot

- 카드: t365
- 워크트리: `.claude/worktrees/t365` · 브랜치 `WT-spec-discovery-blind`
- 베이스: 로컬 develop `e45054c56` (워크트리 최초 베이스는 `f7cabfc29`였고 진입 직후 fast-forward 흡수)
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t365`

---

## 1. 착수 전 판정 [HARD]

### 판정 ① — 모집단 계수

명령 (측정 트리, HEAD `e45054c56`):

```
for d in .moai/specs/*/; do [ -f "$d/spec.md" ] || echo "NO_SPEC_MD: $d"; done
```

출력:

```
NO_SPEC_MD: .moai/specs/_archive/
NO_SPEC_MD: .moai/specs/SPEC-V3R4-CC2X-ADOPT-001/
NO_SPEC_MD: .moai/specs/SPEC-V3R4-CC2X-ADOPT-002/
```

`.moai/specs/*/` 총 747개 중 3개. `_archive/`는 SPEC 디렉터리가 아니라 아카이브 루트이며
(`SPEC-` 접두사가 없고, 하위 6개는 전부 spec.md 보유 — `for d in .moai/specs/_archive/*/` 무출력),
`lintSpecsDirRootIntegrity`의 밑줄 화이트리스트가 이미 면제한다.

**실질 모집단 = 2건.** 리드가 develop 트리에서 잰 값과 일치한다 — 다만 이 수치는 이 트리에서
다시 잰 것이고, 리드 값을 옮겨 적은 것이 아니다.

심각도: **Tier S**. 모집단이 2건이고 둘 다 같은 출처(#861, 2026-05-12)의 umbrella research
문서이며, 새로 생산되는 경로가 없다.

### 판정 ② — 허용되는 상태인가 (이 카드의 본질)

**허용되지 않는다.** 근거는 스키마의 티어 정의다 —
`.claude/rules/moai/development/spec-frontmatter-schema.md:167`:

> `tier` | enum (S\|M\|L) | Tier S = 2 files (spec.md + plan.md); Tier M = 3 files
> (spec.md + plan.md + acceptance.md); Tier L = 5 files (…)

**모든 티어가 spec.md를 포함한다.** spec.md 없는 SPEC 디렉터리를 정당화하는 티어가 하나도 없다.
반증을 찾기 위해 `research-only` / `precursor` 개념이 워크플로 규약에 있는지 훑었으나
(`spec-workflow.md`, `spec-frontmatter-schema.md`) 무출력이었다.

두 디렉터리의 실체:

```
.moai/specs/SPEC-V3R4-CC2X-ADOPT-001/  → research.md 1개 (19343 B)
.moai/specs/SPEC-V3R4-CC2X-ADOPT-002/  → research.md 1개 (17588 B)
```

`research.md` 서두는 `spec_id:` / `phase: research` 를 쓴다 — 현행 스키마의 `id:` 가 아닌
구형 키다. 즉 현행 스키마 이전(#861, `8851c2b3c`)의 umbrella research 산출물이 spec.md를
끝내 받지 못한 채 남은 것이고, 정당한 형태가 아니라 **결함 상태**다.

**따라서 수리 방향은 "발견 경로를 넓힌다"가 아니라 "그 디렉터리 자체를 막는다".**

---

## 2. 결함의 실제 모양 — 두 필터를 모두 빠져나간다

`internal/spec/lint.go`의 자동 발견 경로는 두 겹이다.

1. `discoverSPECs(baseDir)` — `SPEC-*/spec.md` 를 stat 해서 **있을 때만** 경로에 담는다.
   spec.md가 없으면 그 디렉터리는 목록에 아예 오르지 않는다.
2. `lintSpecsDirRootIntegrity(baseDir)` — 루트에 있으면 안 되는 항목을
   `SpecsDirForeignEntry` 경고로 표면화한다. 그런데 이 함수는
   `entry.IsDir() && strings.HasPrefix(name, "SPEC-")` 인 항목을 **이름만 보고 통째로 면제**했다.

`SPEC-*` 이면서 spec.md가 없는 디렉터리는 ①에서 이름이 없고 ②에서 이름 때문에 면제된다.
어느 규칙도 그것을 **통과시킨 것이 아니라 방문한 적이 없다.** t357 M2가 관측한
"정리 대상 393 대 lint finding 392"의 차분 1건이 정확히 이 경로다.

---

## 3. 수리

`internal/spec/lint.go` — `lintSpecsDirRootIntegrity` 에 `SpecsDirMissingSpecFile` 을 추가.
`SPEC-*` 디렉터리 면제를 유지하되, 면제하기 전에 spec.md 존재를 stat 하고 없으면 경고를 낸다.

**discoverSPECs는 건드리지 않았다.** 발견 경로를 넓히면 존재하지 않는 파일에서 파싱한
SPECDoc이 모든 per-doc 규칙에 넘어가 `ParseFailure` 에러 + 빈 문서 기반의 허위 finding이
쏟아진다 — 판정 ②가 "허용되지 않는다"로 났으므로 넓히는 쪽은 애초에 틀린 수리다.

심각도는 **warning**. 형제 규칙 `SpecsDirForeignEntry` 의 선례를 그대로 따랐고, 근거는
`lint_artifact_status.go:50-60` 이 적어둔 원칙이다 — 정리(cleanup)가 같은 SPEC에서 함께
착지할 때만 `error` 가 안전하다. 이 카드는 두 디렉터리를 처분하지 않으므로(범위 밖, §5 참조)
`error` 로 두면 착지 즉시 코퍼스가 붉어진다.

---

## 4. 회귀 [HARD] — 양방향, 한 테스트

`TestLinter_SpecsDirMissingSpecFile_BothDirections` (`internal/spec/lint_test.go`).
한쪽만 보면 "발견 경로를 넓힌 것"과 "규칙을 끈 것"이 구별되지 않으므로 둘을 한 테스트에 넣었다.

심은 것:

| 디렉터리 | 내용 | 기대 |
|---|---|---|
| `SPEC-BLIND-001/` | `research.md` 만 (spec.md 없음) | `SpecsDirMissingSpecFile` 1건, warning |
| `SPEC-VALID-001/` | `valid` 픽스처 spec.md | 무고발 |
| `SPEC-LIVE-001/` | `missing-coverage` 픽스처 spec.md | `CoverageIncomplete` 가 **여전히** 발화 |

- **A 방향(구멍이 닫힘)**: `SPEC-BLIND-001` 이 정확히 1건 표면화되고 warning 이다.
- **B 방향(끈 것이 아님)**: 정상 두 디렉터리는 고발되지 않고, `ListDocs(root)` 가
  **정확히 2건**만 돌려주며 그 안에 `SPEC-BLIND-001` 이 없다(발견 경로가 안 넓어짐).
  그리고 `SPEC-LIVE-001` 의 per-doc 규칙이 그대로 발화한다(규칙이 안 꺼짐).
  덧붙여 `SPEC-BLIND-001` 이 `SpecsDirForeignEntry` 로 잡히지 않는 것도 못박았다 —
  두 코드는 다른 결함을 가리키며 겹치면 안 된다.

### 뮤테이션으로 RED 확보

새 테스트가 공허하지 않음을 보이기 위해 수리를 되돌린 상태에서 먼저 관측했다.

```
git show HEAD:internal/spec/lint.go > internal/spec/lint.go
go test ./internal/spec/ -run TestLinter_SpecsDirMissingSpecFile_BothDirections -count=1
```

```
--- FAIL: TestLinter_SpecsDirMissingSpecFile_BothDirections (0.35s)
    lint_test.go:1042: expected exactly 1 SpecsDirMissingSpecFile (SPEC-BLIND-001), got 0: []
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.878s
```

수리본 복원 후:

```
--- PASS: TestLinter_SpecsDirMissingSpecFile_BothDirections (0.32s)
ok  	github.com/modu-ai/moai-adk/internal/spec	0.692s
```

---

## 5. 미발행 카드 후보의 처분 (리드 요청)

리드가 별도로 들고 있는 후보 — "AuditError 2건: SPEC-V3R4-CC2X-ADOPT-001/-002 spec.md 부재" —
는 **판정 ②가 그 처분도 정한다**: 두 디렉터리는 정당한 형태가 아니므로 그 후보는 실재하는
작업이다. 처분(spec.md를 채우느냐 / `_archive` 로 옮기느냐 / 지우느냐)은 이 카드 범위 밖이며,
카드 발행은 운영자 소관이므로 여기서 만들지 않았다.

이 수리가 착지하면 두 디렉터리는 lint 표면에서도 보이게 되므로, 그 후보를 처리할 때
`spec lint` 가 `SpecsDirMissingSpecFile` 0건이 되는 것이 완료 판정이 된다.

`audit` 표면은 이미 보고 있었다 — `Audit()` (`internal/spec/audit.go:181`) 은
`SPEC-*` 전부를 spec.md 유무와 무관하게 방문하므로 blind spot이 없다. 실측:

```
go run ./cmd/moai spec audit --json  →  finding_type "AuditError", severity INFO, 두 SPEC 모두
```

즉 이 수리 이후 두 표면(audit INFO / lint WARNING)이 같은 2건을 가리키며 서로 정합한다.

---

## 6. 검증

측정 트리 `.claude/worktrees/t365`, 수리 적용 상태.

| 명령 | 결과 |
|---|---|
| `gofmt -l internal/spec/lint.go internal/spec/lint_test.go` | 무출력 |
| `go vet ./internal/spec/` | 무출력 |
| `go test ./internal/spec/ -count=1 -skip TestCatalogHashParity` | `ok … 82.366s` |
| `go test ./internal/epic/ ./internal/web/ -count=1` | `ok` / `ok` |
| `go test ./internal/cli/ -count=1 -timeout 900s` | `ok … 510.377s` |
| `go run ./cmd/moai spec lint` (실 코퍼스) | `SpecsDirMissingSpecFile` **정확히 2건**, exit 0 |

소비 표면은 `spec.NewLinter` / `ListDocs` 를 부르는 곳을 grep 해서 정했다 —
`internal/web/viewmodel_ops.go`, `internal/cli/spec_lint.go`, `internal/cli/mcp_server.go`,
`internal/epic/discover.go` → `internal/web`, `internal/cli`, `internal/epic` 전부 초록.

### 상속된 적색 1건 (이 카드 소관 아님)

`TestCatalogHashParity` 가 `CATALOG_HASH_DRIFT` 34건으로 실패한다.
**베이스에서도 동일하게 34건**임을 확인했다 — 내 두 파일을 `git show HEAD:` 로 되돌린 뒤
같은 테스트를 돌려 34를 셌고, 수리본 복원 후에도 34다. 내 diff는
`internal/spec/lint.go` + `internal/spec/lint_test.go` 2파일뿐이고
템플릿 `SKILL.md` 를 건드리지 않는다. develop 팁의 선재 적색이다.

**흡수로는 풀리지 않는다.** 처음에는 해시 재생성 커밋 `86456663e` 가 내 base 에 없는 것이
원인이라는 가설이 있었고, 조상 관계는 내 트리에서 맞았다
(`git merge-base --is-ancestor 86456663e HEAD` → rc=1,
 `… refs/heads/develop` → rc=0). 그러나 그 가설은 **틀렸다** — 조상 관계는
"재생성 커밋이 내 base 에 없다"만 보일 뿐 34가 0이 된다를 보이지 않는다.

실제 원인은 **세지 않은 세 번째 소비자**이고, 내 트리에서 직접 읽어 확인했다:

- `internal/spec/catalog_hash_test.go` 의 `resolveHashSourcePath` 는 skill 디렉터리에 대해
  `SKILL.md` / `skill.md` **단일 파일**만 돌려준다 — v1 해시 방식이다.
- 같은 파일의 주석이 이미 자기 중복을 선언해 두고 있다: *"a deliberate duplicate kept here
  because the script lives in `package main` and cannot be imported"*.
- 그런데 `internal/template/catalog_tree_hash.go` 의 `ComputeDirTreeHash` 문서는
  소비자를 **둘로** 센다 — *"Both consumers … the generator … and the audit test
  (TestManifestHashFormat …)"*. `TestCatalogHashParity` 가 그 목록에 없다.

즉 t323 이 없애려던 쌍 드리프트가, 세지 않은 세 번째 소비자에서 다시 났다.
`catalog.yaml` 은 생성기 기준으로 옳고 붉은 것은 테스트 쪽이라는 것은 lane-14 의 실측이며
(병합 트리에서 `gen-catalog-hashes --all` 이 diff 0), 나는 그 명령을 돌리지 않았다 —
`catalog.yaml` 을 쓰는 명령이라 이 카드 범위 밖이다.

**수리하지 않는다.** 카드 범위 밖이고, 수리 방향 자체가 결정을 요구한다
(테스트를 `ComputeDirTreeHash` 로 옮길 것인지, 해시 개념 둘을 유지하기로 하고 그 사실을
문서에 적을 것인지). 따라서 이 카드의 재측정은 **`TestCatalogHashParity` 를 이름을 밝혀
제외**하고 나머지 `internal/spec` + 건드린 소비 패키지로 판정했다.

---

## 7. 범위 밖으로 남긴 것 (보고만)

1. **처분 결정**: 두 ADOPT 디렉터리 자체의 처분 — §5, 운영자 소관.
2. **스테일 인용**: `lint.go` 의 `SpecsDirForeignEntry` 메시지와 함수 주석이
   `spec-frontmatter-schema.md § Root Integrity` 를 가리키는데, 그 절이 실재하지 않는다
   (`grep -rn "Root Integrity" .claude/rules/` 무출력). 기존 메시지 문자열이라 이 카드에서
   건드리지 않았다.
3. **`TestCatalogHashParity` 상속 적색** — §6.

---

## 8. Gaps / 잔여 위험

- **Gaps**: CI 판정을 읽지 않았다(레인은 CI를 직접 요청하지 않는다 — 판독은 리드 몫).
  darwin/windows 매트릭스와 `-race` 는 관측하지 않았다. 전체 스위트(`go test ./...`)는
  로컬에서 돌리지 않았다(§4.1 규율).
- **잔여 위험**: 새 코드가 `warning` 이라 `--strict` 모드에서는 게이트를 붉힌다.
  이 리포의 코퍼스에 2건이 남아 있으므로, `--strict` 로 도는 경로가 있다면 그쪽은
  §5의 처분이 끝나기 전까지 붉어진다. `spec lint` 기본 경로는 exit 0 으로 실측했다.
