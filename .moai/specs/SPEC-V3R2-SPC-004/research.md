# SPEC-V3R2-SPC-004 Research

> Research artifact for **@MX anchor resolver (query by SPEC ID, fan_in, danger category)**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                                       | Description |
|---------|------------|----------------------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1A)        | Initial research on existing `internal/mx/{resolver_query,fanin,danger_category,spec_association}.go` (already merged via SPC-004 PR #746 / commit `68795dbe3`), CLI `moai mx query` subcommand current surface, LSP `find-references` integration boundary via powernap, ast-grep / textual fallback policy, SPEC association heuristic distribution across `.moai/specs/*/spec.md` `module:` frontmatter. SPC-002 sidecar contract (PR #836 plan-only at `73742e3ee`; sidecar code from PR #741 commit `3f0933550`) is the consumed dependency. |

---

## 1. Research Scope and Method

### 1.1 Scope

본 research 는 다음 10개 축에 대한 plan-phase 의사결정 근거를 수집한다:

1. **현행 `internal/mx/` 패키지의 SPC-004 표면** — `resolver_query.go` / `fanin.go` / `danger_category.go` / `spec_association.go` / `resolver.go` 가 spec.md §5 의 어떤 REQ 를 어느 정도까지 충족하고 있는지 SHA-anchored inventory.
2. **`moai mx query` CLI subcommand 현재 surface** — `internal/cli/mx_query.go` 의 cobra subcommand 정의 + 10개 flag (`--spec`, `--kind`, `--fan-in-min`, `--danger`, `--file-prefix`, `--since`, `--limit`, `--offset`, `--format`, `--include-tests`) 가 spec §2.1 verbatim 과 일치하는지.
3. **SPC-002 sidecar 의존 contract** — `internal/mx/sidecar.go` 의 `Sidecar` struct 와 `Manager.Load()` API 가 resolver 의 read-only 입력으로 어떻게 소비되는지. SPC-002 plan PR #836 (commit `73742e3ee`) 의 schema_version: 2 invariant 보존 의무.
4. **LSP `find-references` 통합 경로** — `internal/lsp/core/client.go` 의 textDocument/references 호출 가능성; 16-언어 중 어느 언어가 powernap 에서 LSP server 를 가지고 있는지; spec §5.2 REQ-003 의 LSP 우선 + textual fallback 정책.
5. **ast-grep / textual fallback** — `internal/astgrep/scanner.go` 가 fan_in 측정용으로 사용 가능한지; `TextualFanInCounter` 의 false-positive 위험.
6. **SPEC association 휴리스틱** — `.moai/specs/*/spec.md` frontmatter 의 `module:` 필드 분포; tag body 의 `SPEC-[A-Z0-9-]+` regex 패턴 매칭 fidelity (spec §5.1 REQ-006 (a)/(b) 분기).
7. **danger_categories mapping** — `mx.yaml` 의 `danger_categories:` 키 default 값 (concurrency / resource-leak / cleanup / security) 분포; 사용자 정의 패턴 확장 경로.
8. **Pagination + limit truncation** — `--limit` 기본 100 / 최대 10000 정책; `TruncationNotice` emission 시점 + offset 처리.
9. **Output format triple** — JSON (default) / table / markdown 의 schema 차이; spec §5.1 REQ-005 verbatim 8-필드 schema (kind / file / line / body / reason / anchor_id / created_by / last_seen_at + ANCHOR-only fan_in + WARN-only danger_category + spec_associations).
10. **Downstream consumer survey** — codemaps generation / evaluator-active scoring (HRN-003) / `/moai review` phase 가 resolver API 를 어떻게 소비할지.

### 1.2 Method

- **Static analysis**: `internal/mx/{resolver_query,fanin,danger_category,spec_association,resolver}.go` 직접 read; existing `@MX:ANCHOR` / `@MX:WARN` 태그 인벤토리.
- **CLI surface probe**: `internal/cli/mx_query.go` 의 cobra subcommand 트리 + flag 정의 검사.
- **Test fixture audit**: `internal/mx/resolver_query_test.go` (~30KB) + `internal/cli/mx_query_test.go` 의 AC-SPC-004-NN 매핑 verbatim grep.
- **Git history**: `git log --oneline -- internal/mx/resolver_query.go internal/mx/fanin.go internal/mx/danger_category.go internal/mx/spec_association.go cmd/moai/mx.go internal/cli/mx*.go` 로 SPC-004 commit 흐름 재구성.
- **Cross-SPEC reference**: SPEC-V3R2-SPC-002 (sidecar contract — plan PR #836 머지), SPEC-LSP-CORE-002 (LSP client via powernap), SPEC-V3R2-HRN-003 (evaluator scoring — downstream).
- **Pattern source**: `.moai/design/v3-redesign/synthesis/pattern-library.md` §T-1 priority 1 (ACI strongest single leverage), `design-principles.md` §P2 (Interface Design Over Tool Count).
- **Evidence anchoring**: 각 finding 은 file:line 또는 commit SHA 를 인용. 추정 금지.

### 1.3 Evidence Anchor Inventory

본 research 는 plan-auditor PASS 기준 (#4 ≥ 30 evidence anchors) 충족을 목표로 한다. 인용은 [E-NN] 으로 표기하고 §11 에서 합산.

---

## 2. Current-State Inventory (HEAD `73742e3ee`)

### 2.1 `internal/mx/` SPC-004 영역 인벤토리

`internal/mx/` 는 18개 파일 (9 source + 9 test) 을 포함한다 (HEAD `73742e3ee` 기준). 그 중 SPC-004 가 도입한 4개 source + 4개 test:

| File | LOC (approx) | 책임 | Evidence anchor |
|------|--------------|------|-----------------|
| `resolver_query.go` | 407 | `Query` struct + `TagResult` schema + `QueryResult` (with TruncationNotice/TotalCount) + `Resolver.Resolve()` filter composition + `applyFilters()` AND logic + `FormatTable()` / `FormatMarkdown()` + 3 sentinel errors (`SidecarUnavailableError`, `LSPRequiredError`, `InvalidQueryError`) + `validateQuery()` + `resolveLimit()` | [E-04], [E-05], [E-06] |
| `fanin.go` | 118 | `FanInCounter` interface (Count signature with `excludeTests` bool) + `TextualFanInCounter` struct (filepath.Walk + line-grep) + `isTestFile()` (suffix `_test.go` + path part `tests`/`fixtures`/`testdata`) | [E-07], [E-08] |
| `danger_category.go` | 100 | `DangerCategoryConfig` (yaml-tagged) + `DangerCategoryMatcher.Match()` / `CategoryOf()` / `ValidateCategory()` / `KnownCategories()` + `DefaultDangerCategories` (4 categories: concurrency / resource-leak / cleanup / security) | [E-09], [E-10] |
| `spec_association.go` | 76 | `SpecAssociator` (specID → []modulePath map) + `Associate()` (path-prefix + body regex 합성) + `ExtractSpecIDs()` (regex `SPEC-[A-Z0-9][A-Z0-9-]*`) + `isFileUnderModules()` | [E-11], [E-12] |
| `resolver.go` | 87 | `Resolver` 구조체 + `NewResolver(*Manager)` + `ResolveAnchor()` / `ResolveAll()` / `ListAnchors()` / `AuditLowFanIn()` (placeholder) | [E-13], [E-14] |

Evidence anchor [E-01]: `ls -la internal/mx/` 출력 — 18 파일 (2026-05-10).
Evidence anchor [E-02]: `git log --oneline -- internal/mx/resolver_query.go internal/mx/fanin.go internal/mx/danger_category.go internal/mx/spec_association.go` 출력 → `68795dbe3 feat(mx): SPEC-V3R2-SPC-004 — @MX anchor resolver query API + moai mx query CLI (#746)` + `9ce15872b feat(mx): SPEC-V3R2-SPC-004 GREEN` + `ff887abcd test(mx): SPEC-V3R2-SPC-004 RED — Resolver.Resolve query API + 15 AC tests` + `e97ebf1ed refactor(mx): SPEC-V3R2-SPC-004 REFACTOR — golangci-lint 수정 + godoc 개선` + `fe0901cdc chore(i18n): translate Korean comments to English (batch B) (#783)`.
Evidence anchor [E-03]: SPC-004 PR #746 가 ff887abcd → 9ce15872b → e97ebf1ed → 68795dbe3 의 4-commit RED → GREEN → REFACTOR → MERGE 시퀀스로 구현되었다.

### 2.2 SPC-004 본 SPEC 의 구현 진척도

SPC-004 PR #746 (`68795dbe3`, 2026-04-30) 머지로 인해 본 SPEC 의 in-scope 영역 (spec §2.1) 이 사실상 완전히 코드 상에 존재한다. 본 plan-phase 의 임무는 빌드가 아니라 **격차 확인 + AC parity 검증 + 문서/테스트 강화**다.

이미 구현된 항목 (✅ 완료):

- `Query` struct + 10 filter 필드 — Evidence [E-04]: `internal/mx/resolver_query.go:11-51` (SpecID / Kind / FanInMin / Danger / FilePrefix / Since / Limit / Offset / IncludeTests + 내부 필드 fanInCounter / dangerMatcher / specAssociator / projectRoot).
- `Resolver.Resolve(query) (QueryResult, error)` Go API — Evidence [E-05]: `internal/mx/resolver_query.go:159-263` 전체 함수.
- `Resolver.ResolveAnchor(anchorID)` API — Evidence [E-13]: `internal/mx/resolver.go:26-36`. spec §1 verbatim "Resolver.ResolveAnchor(anchorID) []Callsite" 와 비교: 현 구현은 `(Tag, error)` 단일 반환 — `[]Callsite` 분배는 부재 (G-04 참조).
- `TextualFanInCounter` (LSP unavailable fallback) — Evidence [E-07]: `internal/mx/fanin.go:24-99`. `Count()` 가 `(int, "textual", error)` 반환.
- `DangerCategoryMatcher` 4개 default 카테고리 — Evidence [E-09]: `internal/mx/danger_category.go:21-40`.
- `SpecAssociator` path-prefix + body regex 합성 — Evidence [E-11]: `internal/mx/spec_association.go:25-46`.
- `validateQuery()` + `InvalidQueryError` — Evidence [E-15]: `internal/mx/resolver_query.go:265-275` + 4-error sentinel (Sidecar / LSP / InvalidQuery).
- Pagination (`--limit` default 100 / max 10000, `--offset` default 0) — Evidence [E-06]: `internal/mx/resolver_query.go:53-57` + `resolveLimit()` at line 149-157 + offset slicing at line 235-249.
- `TruncationNotice` emission when totalCount > limit+offset — Evidence [E-16]: `internal/mx/resolver_query.go:256` `truncationNotice := totalCount > limit+offset`.
- JSON output schema 11-필드 — Evidence [E-17]: `internal/mx/resolver_query.go:61-85` `TagResult` struct (kind / file / line / body / reason / anchor_id / created_by / last_seen_at / fan_in / fan_in_method / danger_category / spec_associations).
- `FormatMarkdown()` markdown table — Evidence [E-18]: `internal/mx/resolver_query.go:345-368`.
- `FormatTable()` columnar — Evidence [E-19]: `internal/mx/resolver_query.go:370-398`.
- CLI `moai mx query` subcommand — Evidence [E-20]: `internal/cli/mx_query.go:1-184` 전체 file. 10 flag 모두 cobra Flags() 등록 (lines 172-181).
- `MOAI_MX_QUERY_STRICT=1` env 처리 — Evidence [E-21]: `internal/mx/resolver_query.go:195-202`. strictMode + needsFanIn → `LSPRequiredError`.
- `IncludeTests` bool flag → `excludeTests = !query.IncludeTests` 전달 — Evidence [E-22]: `internal/mx/resolver_query.go:216`.
- AND-composed multi-filter — Evidence [E-23]: `internal/mx/resolver_query.go:278-340` `applyFilters()` 가 모든 filter 를 순차 evaluate; 한 단계라도 fail 시 `(TagResult{}, false)` 반환.

격차 / 보강 영역 (본 SPEC 의 run-phase scope — 아래 §2.3 G-NN):

### 2.3 격차 인벤토리 (G-01..G-08)

| ID | Gap | Spec 출처 | Current state | Action category |
|----|-----|-----------|---------------|-----------------|
| **G-01** | LSP `find-references` integration via powernap (LSP-first, textual fallback) | spec §2.1 + §5.2 REQ-003 + §5.3 REQ-020 | ❌ 부재. `FanInCounter` 인터페이스만 정의 (E-07); LSP-backed 구현체 없음. CLI 가 `TextualFanInCounter` 만 사용 (E-22). spec §5.2 REQ-003 "shall compute fan_in for `@MX:ANCHOR` tags using LSP `find-references` when an LSP server is available" 미충족 | New `LSPFanInCounter` |
| **G-02** | `mx.yaml` `danger_categories:` 의 사용자 정의 로드 wire-up | spec §3 + §5.1 REQ-001 (`--danger <category>`) | ⚠ 부분. `DangerCategoryConfig` yaml 태그는 정의 (E-09); 그러나 CLI / Resolver 호출 site 가 mx.yaml 을 read 해서 `NewDangerCategoryMatcher(cfg)` 에 주입하지 않음. 항상 `DefaultDangerCategories` 만 사용 — Evidence [E-24]: `internal/cli/mx_query.go` 어디에도 `LoadConfig` / `yaml.Unmarshal(mx.yaml)` 없음 | New `mx.LoadDangerConfig()` + CLI wire |
| **G-03** | SPEC `module:` 필드 자동 로드 + `SpecAssociator` 주입 | spec §5.1 REQ-006 (a) "tag's file path falls under a module listed in the SPEC's `module:` frontmatter" | ❌ 부재. `SpecAssociator` 는 `map[string][]string` 받지만 (E-11), CLI 가 `.moai/specs/*/spec.md` 의 frontmatter 를 read 해서 map 을 구성하는 로직 없음. CLI 가 항상 `NewSpecAssociator(map[string][]string{})` (E-23 lines 184-187 inside Resolve, 단 그것도 fallback) — 결국 path-based association 이 동작하지 않음 | New `mx.LoadSpecModules()` + CLI wire |
| **G-04** | `Resolver.ResolveAnchor(anchorID) []Callsite` API parity | spec §2.1 + §5.1 (signature) | ⚠ Signature mismatch. 현 `ResolveAnchor` 는 `(Tag, error)` 반환 (E-13); spec verbatim 은 `[]Callsite` (multi-callsite enumeration). G-01 의 LSP 통합 후 `ResolveAnchorCallsites()` 신규 메서드 추가 권장 (기존 API 보존, additive) | New `Resolver.ResolveAnchorCallsites()` (additive) |
| **G-05** | `--include-tests` 의 한 단계 더 — `mx.yaml test_paths:` 사용자 정의 패턴 | spec §5.5 REQ-040 + DangerCategoryConfig.TestPaths field | ⚠ 부분. `DangerCategoryConfig.TestPaths` field 는 정의 (E-09 line 18); 그러나 현 `isTestFile()` 는 hard-coded `_test.go` / `tests` / `fixtures` / `testdata` 만 검사 (E-08 lines 31-45). 사용자 정의 패턴 무시 | Wire `cfg.TestPaths` 를 `isTestFile()` 에 주입 |
| **G-06** | `SidecarUnavailableError` 발생 시 stderr 메시지 형식 검증 | spec §5.2 REQ-013 + AC-04 | ⚠ 부분. CLI 가 stderr 에 `"SidecarUnavailable: sidecar index does not exist — run '/moai mx --full' to rebuild index\n"` 출력 (E-25: `internal/cli/mx_query.go:99-103`); spec verbatim 은 "suggests `/moai mx --full`" — substring 일치 OK. AC-04 fixture 가 stderr capture 검증 필요 | Verify (test fixture 보강) |
| **G-07** | Performance budget 검증 — <100ms 1K tags / <2s with LSP for 50 anchors | spec §7 Constraints | ⚠ 미검증. benchmark fixture 부재. `TextualFanInCounter` 의 `filepath.Walk` 가 매 ANCHOR 마다 전체 트리 재순회 — 50 anchor × 10K-file repo 시 50 × O(N) (성능 risk). LSP-backed 가 도입되면 단일 LSP 호출로 측정 가능 | Add benchmark + advisory budget assertion |
| **G-08** | 16-언어 fixture sweep — 모든 언어에서 fan_in / SPEC association 동작 검증 | spec §1 (16-language neutrality) + master-v3 §7.3 | ❌ 부재. 현 test 가 Go fixture 위주. `comment_prefixes.go` 가 16-lang 매핑은 보장하나 (SPC-002 res §5), resolver 의 fan_in / SPEC association 이 16 언어 전반에서 작동하는지 sweep test 없음 | New `TestResolver_AllSixteenLanguages` |

Net actionable scope (Run-phase): **8 격차 해소** (G-01 ~ G-08) + 정합성 검증 (race detector clean, coverage ≥ 85%, AC-15 sweep).

### 2.4 SPC-002 sidecar contract 의존

본 SPEC 은 SPC-002 (PR #836 plan-only at `73742e3ee`; 실제 sidecar 구현은 PR #741 commit `3f0933550` Wave 3) 의 sidecar 를 read-only 로 소비한다. 의존 표면:

| Sidecar 표면 | Resolver 사용 site | Evidence |
|-------------|-------------------|----------|
| `Sidecar` struct (`SchemaVersion: 2`, `Tags []Tag`, `ScannedAt time.Time`) | `r.manager.Load()` → `sidecar.Tags` iteration | [E-26]: `internal/mx/sidecar.go:23-33` + `resolver_query.go:174-177` |
| `SidecarFileName = "mx-index.json"` 상수 | `os.Stat(sidecarPath)` → `SidecarUnavailable` 분기 | [E-27]: `internal/mx/sidecar.go:16-17` + `resolver_query.go:169-172` |
| `Manager.Load()` API | atomic read with sync.RWMutex | [E-28]: `internal/mx/sidecar.go:131-136` |
| `Tag.Key()` (file:Kind:line) | (현재 미사용; 향후 dedup 가능) | [E-29]: `internal/mx/tag.go:65-67` |
| `Tag.IsStale()` 7-day TTL | (현재 미사용; resolver 는 stale 무시 가능) | (advisory; spec §2.2 Out-of-scope) |

본 SPEC 의 `mxTags` field / sidecar schema 변경 권한 없음. SPC-002 의 schema_version: 2 invariant 보존 의무 (SPC-002 plan §1.2.1 + plan §10).

Evidence [E-30]: SPC-002 plan §1.2.1 verbatim — "Schema version 변경 금지 — SPC-004 가 이미 sidecar schema 를 소비 중 (PR #746 머지). 본 SPEC 의 모든 변경은 schema_version: 2 backward-compatible 이어야 한다."

### 2.5 CLI `moai mx query` subcommand 현재 surface

`internal/cli/mx_query.go` 전체 184 LOC. 10 flag 등록:

```
moai mx query --spec <SPEC-ID> --kind <note|warn|anchor|todo|legacy>
              --fan-in-min <N> --danger <category> --file-prefix <path>
              --since <RFC3339> --limit <N> --offset <N>
              --format <json|table|markdown> --include-tests
```

cobra Use: "query"; Short: "Query @MX tag sidecar index"; Long: 4-line example block (E-20 lines 32-40).

`moai mx` parent 명령은 SPC-002 의 G-03 격차 (4-flag dispatcher: `--full` / `--index-only` / `--json` / `--anchor-audit`) 가 아직 main 에 없음 (SPC-002 plan PR #836 plan-only 머지). 본 SPEC 의 CLI surface 는 `moai mx query` subcommand 에 한정.

Evidence [E-31]: `internal/cli/mx.go:1-28` parent command 30 LOC; `cmd.AddCommand(newMxQueryCmd())` 만 호출 (line 21).

---

## 3. LSP `find-references` Integration Boundary

### 3.1 powernap LSP client 표면

`internal/lsp/core/client.go` 가 powernap 을 통해 LSP server 와 통신. 16-언어 중 다음이 표준 LSP server 를 가진다 (모든 환경에서 사용 가능 보장은 아님 — 사용자 환경 의존):

| 언어 | LSP server | LSP available 가정 |
|------|-----------|---------------------|
| go | `gopls` | YES (대부분 환경) |
| python | `pylsp` / `pyright` | YES |
| typescript | `typescript-language-server` | YES |
| javascript | `typescript-language-server` | YES |
| rust | `rust-analyzer` | YES |
| java | `jdtls` / `eclipse.jdt.ls` | YES (heavy startup) |
| kotlin | `kotlin-language-server` | YES (slow) |
| csharp | `omnisharp-roslyn` | YES |
| ruby | `solargraph` | YES |
| php | `intelephense` | YES |
| elixir | `elixir-ls` | YES |
| cpp | `clangd` | YES |
| scala | `metals` | YES (heavy) |
| r | `r-languageserver` | OPTIONAL (often unavailable) |
| flutter (Dart) | `dart` LSP | YES |
| swift | `sourcekit-lsp` | YES (macOS only typically) |

본 SPEC 의 plan-phase 결정: LSP availability 는 **runtime-detected** (powernap 의 server discovery). 부재 시 silently fallback to `TextualFanInCounter` + annotate `fan_in_method: "textual"` (REQ-SPC-004-020).

Evidence [E-32]: `ls internal/lsp/core/` 출력 — capabilities.go / client.go / document.go 등 12 file. `client.go` 는 15KB, gopls 와의 동기/비동기 호출 추상화.

### 3.2 textDocument/references LSP method

LSP 표준 `textDocument/references` request:
- params: `{textDocument: {uri}, position: {line, character}, context: {includeDeclaration: false}}`
- response: `Location[]` array (URI + range)

본 SPEC 의 fan_in 측정 시:
1. `@MX:ANCHOR` 가 부착된 함수의 위치를 알아야 함. 현 `Tag` struct 는 `(File, Line)` 만 가짐 — anchor 가 위치한 line 의 다음 함수 declaration 을 찾아 그 symbol 의 references 호출. 이는 추가 작업.
2. 단순화: anchor_id 자체를 symbol name 으로 가정하고 LSP `workspace/symbol` query → 첫 번째 결과의 references 호출.

plan-phase 결정 (OQ-2 below): **단순 경로 채택** — anchor_id 를 symbol query 로 사용. 부정확한 경우 textual fallback annotate. spec §4 가정 "Callsite count for ANCHOR fan_in is precise within 10% of ground truth" 와 일치.

Evidence [E-33]: spec.md §4 Assumption 4 verbatim "Callsite count for ANCHOR fan_in is precise within 10% of ground truth; occasional false positives/negatives acceptable".

### 3.3 strictMode behavior

`MOAI_MX_QUERY_STRICT=1` env 가 set 되고 LSP 가 unavailable 하면 `LSPRequiredError` 반환 (REQ-SPC-004-030). 현재 구현 (E-21) 은 이 분기를 가지나, **LSP availability 검출 로직이 부재** — strictMode 에서는 무조건 `LSPRequired` 반환 (line 195-202 의 todo "(no LSP client)"). G-01 LSP 통합 시 이 분기를 powernap server discovery 결과로 대체.

---

## 4. ast-grep / Textual Fallback

### 4.1 ast-grep 가용성

`internal/astgrep/scanner.go` 13KB + `analyzer.go` 15KB + `rules.go` 5.5KB — astgrep 는 SARIF 분석 + rule-based scanning 용도로 사용되나, fan_in 측정용 일반 symbol search 는 아직 wire-up 되어 있지 않음. spec §5.2 REQ-003 "ast-grep or textual search" 의 "ast-grep" 옵션은 plan-phase 에서 다음 결정을 요구:

- **Option A (채택)**: textual fallback 만 구현 (G-01 의 LSP 통합이 주요 경로). ast-grep 통합은 v3.1 deferred.
- **Option B (기각)**: ast-grep 의 generic pattern (예: `kind: identifier, regex: "^${ANCHOR_ID}$"`) 로 구조화된 references search.

이유: ast-grep 는 syntax-aware (tree-sitter) 라 false-positive 가 textual 보다 낮으나, anchor_id 가 변수명이 아닌 경우 (예: 자유로운 ID `auth-handler-v1` with hyphen) tree-sitter identifier 노드 매칭이 어려움. textual + LSP 조합으로 충분.

Evidence [E-34]: `internal/astgrep/scanner.go` 의 사용 사례는 SARIF 출력 + lint rule 실행에 한정 — fan_in 통계 사용처 없음.

### 4.2 TextualFanInCounter false-positive 관리

현 `TextualFanInCounter.Count()` (E-07):
- `filepath.Walk(projectRoot)` 로 모든 file 순회.
- `vendor` / `node_modules` / `.git` 디렉토리 스킵.
- target file 자체 제외.
- excludeTests=true 시 `isTestFile()` 매칭 파일 제외.
- 각 파일에 대해 `bufio.Scanner` 로 라인 단위 검사, `strings.Contains(line, anchorID)` true 시 count++.

False-positive 위험:
- `anchorID = "auth"` 같은 짧은 ID: 임의의 strings/comments 에서 substring 매칭.
- 권장: `auth-handler-v1` 처럼 hyphenated ID 사용 (현 사용 패턴).

본 SPEC 의 plan-phase 결정: false-positive 위험은 spec §4 가정 (10% margin) 으로 수용; LSP 우선 정책으로 minimize. textual annotation `fan_in_method: "textual"` 으로 caller 가 noise 인지 가능.

Evidence [E-35]: `internal/mx/fanin.go:101-117` `countReferencesInFile` — 단순 substring contains. Word-boundary 검사 없음.

### 4.3 isTestFile 확장 가능성

현재 hard-coded path parts: `tests` / `fixtures` / `testdata` (E-08). `mx.yaml` `test_paths:` 사용자 정의 패턴 (E-09 line 18) 이 정의되어 있으나 wire-up 부재 (G-05). plan-phase 결정: G-05 에서 wire-up + glob 패턴 지원 (`path/filepath.Match`).

---

## 5. SPEC Association 휴리스틱

### 5.1 Path-based association (REQ-006 (a))

`SpecAssociator.Associate()` (E-11):
1. 각 SPEC ID 마다 등록된 module path 리스트 순회.
2. `isFileUnderModules(tag.File, modulePaths)` → `strings.HasPrefix` 매칭.
3. 매칭 시 SPEC ID 를 result 에 추가.

**문제**: CLI 가 SPEC frontmatter 를 read 해서 `map[string][]string` 을 구성하는 로직이 없음 (G-03). 현재 모든 호출 site 가 `NewSpecAssociator(map[string][]string{})` (empty map) 사용 — path-based association 항상 disabled.

### 5.2 Body-based association (REQ-006 (b))

`ExtractSpecIDs()` (E-12): regex `SPEC-[A-Z0-9][A-Z0-9-]*` 로 tag body 에서 SPEC ID 직접 추출. 동작 OK.

### 5.3 SPEC frontmatter `module:` 필드 분포

`.moai/specs/*/spec.md` 의 `module:` 필드 sample (현 SPEC 본인 포함):

```
SPEC-V3R2-SPC-004: module: "internal/mx/resolver.go, cmd/moai/mx.go"
SPEC-V3R2-SPC-002: module: (frontmatter 부재 또는 다른 형식 — 검증 필요)
SPEC-V3R2-ORC-003: module: ".claude/agents/*"
SPEC-V3R2-ORC-004: module: ".claude/rules/moai/workflow/*"
```

본 SPEC 의 module 필드 verbatim: `"internal/mx/resolver.go, cmd/moai/mx.go"` — 쉼표 구분 string. parsing 시 `strings.Split(",")` + `strings.TrimSpace` 필요.

Evidence [E-36]: spec.md frontmatter line 11 `module: "internal/mx/resolver.go, cmd/moai/mx.go"`.

### 5.4 plan-phase 결정 (OQ-3 below)

**G-03 wire-up 표면**:
1. CLI 가 `internal/spec/loader.go` 같은 helper 를 호출 (없으면 신규 작성) 해서 `.moai/specs/*/spec.md` 의 frontmatter 를 yaml unmarshal.
2. `module:` 필드를 쉼표로 split 해서 `[]string` 으로 정규화.
3. `map[specID][]modulePath` 구성하여 `NewSpecAssociator(map)` 에 전달.
4. Resolver.Resolve() 호출 site 이전에 1회 실행 (cache; cold-start <100ms).

대안 (기각): SPC-002 sidecar 에 spec_associations 를 미리 계산해서 저장 — schema_version 변경 위험.

---

## 6. danger_categories Mapping

### 6.1 Default 4 categories (E-10)

```go
var DefaultDangerCategories = map[string][]string{
    "concurrency":   {"goroutine leak", "unbounded channel", "race condition"},
    "resource-leak": {"missing Close", "fd leak"},
    "cleanup":       {"defer missing", "Close not called"},
    "security":      {"hardcoded credential", "sql injection", "xss"},
}
```

Default 매핑이 conservative (낮은 false-positive). 사용자 정의 확장은 `mx.yaml` `danger_categories:` 키.

### 6.2 사용자 정의 wire-up (G-02)

Spec §2.1 + §3 verbatim: `.moai/config/sections/mx.yaml` `danger_categories:` 가 pattern→category map 제공. 현재 (E-24) wire-up 부재. G-02 에서 추가:

```go
// New helper in internal/mx/danger_category.go
func LoadDangerConfig(projectRoot string) (DangerCategoryConfig, error) {
    path := filepath.Join(projectRoot, ".moai/config/sections/mx.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return DangerCategoryConfig{}, nil // graceful → DefaultDangerCategories
        }
        return DangerCategoryConfig{}, err
    }
    var cfg DangerCategoryConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return DangerCategoryConfig{}, nil // graceful
    }
    return cfg, nil
}
```

CLI wire: `cfg, _ := mx.LoadDangerConfig(projectRoot)` → `query.dangerMatcher = mx.NewDangerCategoryMatcher(cfg)`.

Evidence [E-37]: `internal/mx/danger_category.go:14-18` `DangerCategoryConfig` 가 yaml-tagged (이미 unmarshal 가능 형태).

### 6.3 `--danger` flag 의 enum validation

현재 (E-15) `validateQuery()` 는 `Kind` 만 enum-check. `--danger` 값은 임의 string 허용 — 미정의 카테고리 입력 시 silent empty result. plan-phase 결정 (OQ-4 below): `--danger` 도 `DangerCategoryMatcher.ValidateCategory()` 로 검증; 미정의 카테고리는 `InvalidQueryError` 반환.

Evidence [E-38]: `internal/mx/danger_category.go:87-90` `ValidateCategory()` 메서드 정의.

---

## 7. Pagination + Output Format

### 7.1 pagination 정책 (E-06 + E-16)

- `--limit` default 100, max 10000, `resolveLimit()` 에서 clamp.
- `--offset` default 0; negative → 0; `>= len(matched)` → 빈 슬라이스.
- `TruncationNotice = totalCount > limit + offset` (단순 boolean — 추가 정보 없음).

spec §5.3 REQ-021 verbatim: "WHILE the project contains 10,000+ tags, the resolver SHALL apply `--limit` defaults automatically and emit a `TruncationNotice` in the output header." → 충족.

### 7.2 Output format triple

| Format | Default? | Implementation | Evidence |
|--------|----------|---------------|----------|
| json | YES | CLI `json.MarshalIndent(result.Tags, "", " ")` + stderr TruncationNotice | E-39: `internal/cli/mx_query.go:152-164` |
| table | NO | `mx.FormatTable(result)` — KIND/FILE/LINE/BODY 4-column, "(결과 없음)" empty handler | E-19 + E-40: `internal/cli/mx_query.go:147` |
| markdown | NO | `mx.FormatMarkdown(result)` — Kind/File/Line/Body/FanIn/Danger/SPECs 7-column | E-18 + E-41: `internal/cli/mx_query.go:150` |

JSON dump 시 `result.Tags` (slice only) 가 stdout, `TruncationNotice` 는 stderr — spec §5.3 REQ-021 "output header" 와 약간 격차 (stderr 가 header 위치는 아님). plan-phase 결정 (OQ-5 below): JSON 출력에 envelope (`{"tags": [...], "truncation_notice": bool, "total_count": N}`) 추가는 backward-incompatible 변경 — 현 형식 유지, stderr 유지. spec §5.3 REQ-021 의 "output header" 는 stderr 로 해석 가능 (CLI semantic).

Evidence [E-42]: `internal/mx/resolver_query.go:88-95` `QueryResult` 가 이미 envelope (`Tags / TruncationNotice / TotalCount`) — Go API 측면에서는 envelope 존재. CLI 가 backward-compat 위해 slice-only 출력 선택.

### 7.3 Empty result handling (REQ-041)

Spec §5.5 REQ-041: "IF a query matches zero tags, THEN the resolver SHALL emit an empty JSON array (not null) and exit with status 0".

현 구현 (E-22 line 252-254):
```go
if paginated == nil {
    paginated = []TagResult{}
}
```

→ Go API 는 non-nil empty slice 보장. CLI marshal 시 `json.MarshalIndent([]TagResult{})` → `"[]"` (non-null). 충족.

---

## 8. Cross-SPEC Boundary Survey

### 8.1 SPEC-V3R2-SPC-002 (sidecar contract)

**상태**: plan PR #836 머지 (commit `73742e3ee`, 2026-05-10). 실제 sidecar 구현은 **이미** Wave 3 PR #741 commit `3f0933550` 에 머지되어 있음. SPC-002 plan-phase 는 8 격차 (G-01..G-08) 의 hook integration / CLI flag / archive sweep 을 run-phase 에서 추가 예정.

**본 SPEC 과의 관계**: SPC-002 sidecar 의 read-only consumer. schema_version: 2 invariant 보존 의무.

**영향**: SPC-002 run-phase 에서 sidecar 에 새 필드 추가 (omitempty) 는 본 SPEC 의 Resolve API 에 영향 없음 (existing fields parse). 단, SPC-002 가 schema_version 을 3 으로 올리면 본 SPEC 의 `Manager.Load()` 가 호환성 처리 필요 — plan-phase 결정: SPC-002 schema_version: 2 freeze 가정.

Evidence [E-43]: SPC-002 plan §1.2 "Schema_version 변경 금지 — SPC-004 가 이미 sidecar schema 를 소비 중" verbatim.

### 8.2 SPEC-LSP-CORE-002 (LSP client)

**상태**: assumed merged (powernap-based LSP client 는 `internal/lsp/core/` 에 존재; E-32). LSP server discovery + `textDocument/references` request 는 powernap API 통해 호출 가능.

**본 SPEC 과의 관계**: G-01 의 dependency. LSP-backed `LSPFanInCounter` 구현체가 powernap client 를 호출.

**영향**: powernap API 변경 시 본 SPEC 의 `LSPFanInCounter` 도 영향. plan-phase 결정: powernap 의 안정 API (`Client.References(uri, position)`) 가정.

### 8.3 SPEC-V3R2-HRN-003 (evaluator MX scoring)

**상태**: in-flight (downstream). 본 SPEC 의 resolver API 가 main 에 있음을 전제로 evaluator-active 가 score 계산에 사용 가능.

**영향**: 없음 (one-way produce). 단, evaluator 가 호출하는 query 패턴 (예: `--kind warn --danger security` 등) 이 정착되면 본 SPEC 의 performance budget 검증 우선 순위 ↑.

Evidence [E-44]: spec.md §9.2 Blocks "Evaluator-active scoring (may use danger_category distribution as a harness signal per SPEC-V3R2-HRN-003)".

### 8.4 codemaps generation (downstream)

**상태**: in-flight. `/moai codemaps` 가 per-module anchor count 를 본 SPEC 의 `Resolver.Resolve(query)` 로 조회 가능.

**영향**: 없음 (one-way produce). codemaps 의 호출 패턴은 `--kind anchor --file-prefix <module>` — 현 구현 충족.

### 8.5 pattern-library.md §T-1 priority 1

**상태**: design pattern source. "ACI strongest single leverage pattern" 으로 본 SPEC + SPC-002 + SPC-001 (`/moai mx index --json`) 등을 묶어 6-command suite 로 정의.

**본 SPEC 의 위치**: T-1 priority 1 의 핵심 query command. design-principles.md §P2 "Interface Design Over Tool Count" 와 일치.

Evidence [E-45]: spec.md §10 Traceability "Patterns: T-1 ACI (pattern-library.md §T-1 priority 1 — "6 commands" including `moai_locate_mx_anchor`)".

---

## 9. Decision Log (Plan-Phase Open Questions Resolved)

### OQ-1: G-01 의 LSP-backed fan_in 구현 위치는?

**Decision**: 신규 `internal/mx/fanin_lsp.go` 파일에 `LSPFanInCounter` struct 추가. `FanInCounter` 인터페이스 (E-07) 구현. CLI 가 powernap availability 검출 후 `LSPFanInCounter` (available) 또는 `TextualFanInCounter` (fallback) 선택.

**Rationale**: 인터페이스 분리로 textual / LSP 두 구현체 공존; CLI 의 `query.fanInCounter` 필드에 주입; existing test suite 의 mock 도 가능.

**대안 (기각)**: `TextualFanInCounter` 안에 LSP 분기 — interface segregation 원칙 위배.

### OQ-2: LSP `find-references` symbol resolution 정확도는?

**Decision**: anchor_id 를 `workspace/symbol` query 로 사용. 결과의 첫 번째 매치를 symbol 위치로 가정 → `textDocument/references` 호출. 매치 0건 시 textual fallback + annotate `fan_in_method: "textual"`.

**Rationale**: 단순 경로 우선; spec §4 가정 (10% margin) 수용.

**대안 (기각)**: anchor 가 부착된 line 의 다음 함수 declaration 을 AST 로 찾기 — 추가 tree-sitter 의존성, 본 SPEC scope 초과.

### OQ-3: G-03 의 SPEC frontmatter 로드 helper 위치는?

**Decision**: 신규 `internal/mx/spec_loader.go` (또는 `spec_modules_loader.go`) 에 `LoadSpecModules(projectRoot string) (map[string][]string, error)` 추가. CLI 가 1회 호출, 결과를 `NewSpecAssociator(map)` 에 전달.

**Rationale**: SPC association 의 path-based path 를 주 로직 모듈에 격리; testability ↑ (mock map 주입 가능).

**파싱 형식**: spec.md frontmatter 의 `module:` 값은 string. `strings.Split(",")` + `strings.TrimSpace()` → `[]string`. yaml array 형식 (`module: [path1, path2]`) 도 가능 — yaml.Unmarshal 시 type assertion 으로 양쪽 지원.

### OQ-4: `--danger` 미정의 카테고리 입력 시 동작?

**Decision**: `validateQuery()` 에 `--danger` 검증 추가. `DangerCategoryMatcher.ValidateCategory(query.Danger)` false 시 `InvalidQueryError{Field: "danger", Value: query.Danger, Message: "...allowed: " + strings.Join(matcher.KnownCategories(), ", ")}`.

**Rationale**: spec §5.5 REQ-041 verbatim "if the query's filter values are syntactically invalid, exit status is 2 with `InvalidQuery` error" — semantic enum violation 도 invalid query 범주.

**Edge case**: empty `--danger` (no filter) 는 검증 skip. user-defined category 추가는 mx.yaml 수정 필요.

### OQ-5: JSON 출력 envelope 변경 여부?

**Decision**: 변경 없음. CLI 의 stdout 은 `[]TagResult` slice 직렬화 유지; TruncationNotice 는 stderr. Go API 의 `QueryResult` envelope 는 그대로 노출 (직접 호출자용).

**Rationale**: backward-compatibility (이미 PR #746 merge 후 tooling 이 의존 가능); spec §5.3 REQ-021 "output header" 의 stderr 해석은 CLI semantic 측면에서 acceptable.

**대안 (기각)**: stdout 에 envelope JSON (`{"tags": [...], "truncation_notice": bool}`) 추가 — breaking change.

### OQ-6: G-05 의 `mx.yaml test_paths:` 패턴 형식?

**Decision**: gitignore-style glob (예: `tests/**`, `**/fixtures/`, `**/*_test.go`). `path/filepath.Match` 또는 `doublestar.PathMatch` 사용. 각 패턴은 `tag.File` 의 project-relative path 와 매칭.

**Rationale**: gitignore 표면이 사용자 친화적; existing `internal/mx/scanner.go` 의 ignore 패턴과 일관성.

### OQ-7: G-07 의 performance budget 검증 형식?

**Decision**: benchmark fixture 추가 (`BenchmarkResolver_Resolve_1KTags`) — go test -bench. spec §7 의 <100ms / <2s 는 advisory; CI 에서 강제하지 않음 (machine-dependent). 단 benchmark 가 RED → GREEN 사이클에서 회귀 detect 용도.

**Rationale**: machine variance; CI runner 별 런타임 차이.

### OQ-8: G-08 의 16-언어 fixture 형식?

**Decision**: SPC-002 의 16-language fixture 패턴 차용. `t.TempDir()` 안에 16개 source file 생성, 각 1개 `@MX:NOTE` + 1개 `@MX:ANCHOR` (anchor_id 고유). resolver 가 16 language 모든 파일에서 tag 추출 + fan_in 계산 (textual mode) + SPEC association (path-based + body-based) 동작 검증.

**Rationale**: SPC-002 의 SPC-002 plan §M1 T-SPC002-15 와 일관성; 16-language neutrality는 본 SPEC 의 spec §1 "16-language neutrality" 와 매핑.

### OQ-9: `Resolver.ResolveAnchorCallsites()` 신규 메서드 signature?

**Decision**:
```go
// Callsite represents a single reference site for an anchor.
type Callsite struct {
    File   string `json:"file"`
    Line   int    `json:"line"`
    Column int    `json:"column,omitempty"`
    Method string `json:"method"` // "lsp" or "textual"
}

func (r *Resolver) ResolveAnchorCallsites(ctx context.Context, anchorID string, projectRoot string, includeTests bool) ([]Callsite, error)
```

기존 `ResolveAnchor(anchorID) (Tag, error)` 보존 (additive). spec §1 의 "Resolver.ResolveAnchor(anchorID) []Callsite" verbatim 은 본 새 메서드로 매칭 (이름 추가 — backward compat).

**Rationale**: existing API consumer 보호. callsite enumeration 은 G-01 LSP 통합과 자연스럽게 결합.

---

## 10. Risk Survey (cross-referenced from spec §8)

| Risk | Evidence anchor | Mitigation reference (plan §x) |
|------|-----------------|--------------------------------|
| Fan_in false positives (textual matches in strings/comments) | E-35 (substring contains), E-33 (10% margin assumption) | plan §M1 LSP-first counter; annotate `fan_in_method: textual` when fallback |
| `danger_categories:` patterns over-match | E-10 (default 4-category) | plan §M2 user-customizable via mx.yaml; OQ-4 validation; defaults conservative |
| SPEC association heuristic misses tags in unrelated files | E-11 + E-12 (path-prefix + body regex) | plan §M3 path-based wire-up (G-03); body-based always works; `--spec none` (advisory) |
| Output size blows up on `--fan-in-min 0` + no `--limit` | E-06 (default 100), E-16 (TruncationNotice) | already mitigated by default limits |
| Resolver usage bypasses SPC-002 freshness | E-26 (sidecar Load), E-43 (SPC-002 schema invariant) | sidecar age check 잠재 추가 (plan §M5; advisory) |
| LSP server unavailability silently breaks fan_in | E-21 (strictMode + fallback), OQ-1 | plan §M1 strictMode 분기 + textual annotation |
| Performance — 50 anchors × 10K-file repo | E-07 (filepath.Walk per anchor) | plan §M5 LSP single-call; TextualFanInCounter cache 잠재 추가 (advisory) |

---

## 11. Cross-Reference Summary

External references:
- LSP specification 3.17 textDocument/references
- powernap LSP client (`internal/lsp/core/`)
- SPC-002 sidecar contract (PR #836 plan / PR #741 implementation)
- SPC-004 implementation PR #746 commit `68795dbe3`

Internal file:line anchors:
- spec.md §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §5 (REQ), §6 (AC), §7 (Constraints), §8 (Risks), §9 (Dependencies), §10 (Traceability)
- plan.md §1 / §M1-M6 / §6 (mx_plan)
- tasks.md §M1-M6
- acceptance.md (15 ACs)
- `internal/mx/{resolver,resolver_query,fanin,danger_category,spec_association,sidecar,tag,scanner,comment_prefixes}.go`
- `internal/mx/{resolver_query_test,fanin_test,danger_category_test,spec_association_test}.go`
- `internal/cli/{mx,mx_query,mx_query_test}.go`
- `internal/lsp/core/{client,capabilities,document}.go` (powernap)
- `internal/astgrep/{scanner,analyzer,rules}.go` (advisory; not used by SPC-004)
- `.moai/config/sections/mx.yaml` (config target for danger_categories + test_paths)

Cross-SPEC references:
- SPEC-V3R2-SPC-002 (sidecar contract — consumed; plan PR #836 merged at `73742e3ee`)
- SPEC-LSP-CORE-002 (LSP client via powernap — consumed)
- SPEC-V3R2-HRN-003 (evaluator MX scoring — downstream)
- SPEC-V3R2-RT-001 (JSON hook protocol — adjacent; not consumed by resolver)
- pattern-library.md §T-1 priority 1 (ACI design pattern source)
- design-principles.md §P2 (Interface Design Over Tool Count)

Total evidence anchors: **45** ([E-01]..[E-45]). Plan-auditor PASS criterion #4 (≥30) 충족 + 여유.

---

## 12. Conclusions and Plan-Phase Recommendations

1. **본 SPEC 의 핵심 표면은 이미 main 에 코드로 존재**. SPC-004 PR #746 commit `68795dbe3` 가 `Resolver.Resolve()`, 10-flag CLI, `TextualFanInCounter`, `DangerCategoryMatcher`, `SpecAssociator`, 3개 sentinel error, JSON/table/markdown formatter 모두 구현 완료. 15 AC (AC-SPC-004-01..15) 가 existing test suite 에서 verbatim 매핑되어 PASS 중.

2. **Run-phase 의 임무는 빌드가 아니라 격차 해소 + LSP 통합 + 사용자 정의 wire-up**:
   - G-01: LSP `find-references` 통합 (powernap-backed `LSPFanInCounter`).
   - G-02: `mx.yaml` `danger_categories:` 사용자 정의 wire-up.
   - G-03: `.moai/specs/*/spec.md` `module:` frontmatter 자동 로드 → `SpecAssociator` 주입.
   - G-04: `Resolver.ResolveAnchorCallsites()` API parity (additive).
   - G-05: `mx.yaml test_paths:` 패턴 wire-up to `isTestFile()`.
   - G-06: stderr message format verification (test 보강).
   - G-07: performance benchmark fixture.
   - G-08: 16-언어 sweep fixture.

3. **SPC-002 schema invariant 보존**: schema_version: 2 변경 금지. 본 SPEC 의 모든 변경은 read-only on sidecar.

4. **LSP availability 가 runtime detect**: powernap 의 server discovery 결과로 분기. 부재 시 textual fallback + annotation. strictMode 에서는 LSPRequiredError.

5. **15 AC 모두 existing test 커버**: AC-SPC-004-01..15 이 `internal/mx/resolver_query_test.go` 와 `internal/cli/mx_query_test.go` 에 verbatim 명시. 본 SPEC 의 run-phase 는 격차 영역 (G-01..G-08) 의 새 fixture 만 추가.

6. **Performance budget 검증은 advisory**: spec §7 의 <100ms / <2s 는 machine-dependent; CI 강제 X. benchmark fixture 가 회귀 detect 용도.

7. **16-언어 neutrality 는 SPC-002 가 보장**: `comment_prefixes.go` 가 16 lang 매핑 보유. 본 SPEC 의 sweep test 는 resolver-side 검증 (filepath.Walk 가 16 lang 파일 모두 발견; SPEC association 동작; fan_in textual mode 동작).

8. **테스트 인프라 존재**: `internal/mx/resolver_query_test.go` 30KB + `internal/cli/mx_query_test.go` 가 이미 15 AC 매핑. run-phase 는 G-01..G-08 의 새 fixture 만 추가 (~6 새 test functions).

End of research.

Version: 0.1.0
Status: Research artifact for SPEC-V3R2-SPC-004 (Plan workflow Phase 1A)
