---
id: SPEC-V3-PLG-001
title: Plugin Manifest & Marketplace Parity v1 (agents + skills + commands)
version: 0.1.0
status: draft
created: 2026-04-22
updated: 2026-04-22
author: manager-spec
priority: High
phase: "Phase 6a — Tier 2 Strategic Differentiators"
module: "internal/plugin/"
dependencies:
  - SPEC-V3-SCH-001
  - SPEC-V3-SCH-002
  - SPEC-V3-MIG-001
related_gap: [140, 141, 143, 144, 145]
related_theme: "Theme 6 — Plugin Ecosystem Parity"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3, plugin, marketplace, manifest, extension, P2"
---

# SPEC-V3-PLG-001: Plugin Manifest & Marketplace Parity v1

## HISTORY

- 2026-04-22 v0.1.0: 최초 작성. master-v3 §3.6 Theme 6, gap-matrix #140-#148 (scope-reduced per §9 open question #2), W1.5 §4.1-§4.7 근거. Claude Code의 플러그인 시스템 reduced-scope 파리티.

---

## 1. Goal (목적)

moai-adk-go는 template 기반 배포 모델만 제공해 (W1.6 §4), 사용자가 외부 skill/agent/command를 재사용하거나 공유할 경로가 없다. Claude Code는 `.claude-plugin/plugin.json` manifest + 3 plugin kinds(built-in/marketplace/session inline) + 5 marketplace source types + 6 capabilities로 완성된 ecosystem을 보유한다 (W1.5 §4.1).

본 SPEC은 master-v3 §9 open question #2의 권장 default에 따라 **reduced-scope v1**을 출시한다:
- Capabilities: **agents + skills + commands** 3종만 (hooks / mcpServers / outputStyles는 v3.2 이후)
- Marketplace source: **github + directory** 2종만 (git / url / file은 v3.2 이후)
- Install scopes: **user / project / local** 3종 (SPEC-V3-SCH-002와 일관)

CC와의 manifest 포맷 파리티를 확보해 (향후 full parity 마이그레이션 경로 유지), 플러그인 authoring 관점의 생태계 성장을 시작한다.

### 1.1 배경

W1.5 §4.3 manifest format:
```json
{
  "name": "example-plugin",
  "description": "...",
  "version": "1.0.0"
}
```

W1.5 §4.1 plugin kinds:
- Built-in: `registerBuiltinPlugin()` at startup (v3.0은 없음)
- Marketplace: `~/.claude/plugins/installed.json` (moai 로컬 대응: `~/.moai/plugins/`)
- Session inline: `--plugin-dir` flag (v3.0 지원)

gap-matrix:
- #140 (High, XL): plugin system 부재 → scope-reduced v1
- #141 (High, L): 5 marketplace sources → 2개로 축소 (github + directory)
- #142 (High, XL): 6 capabilities → 3개로 축소
- #143 (Medium, M): install scopes user/project/local — 유지
- #144 (Medium, S): `plugin validate` — 유지
- #145 (Medium, S): `--plugin-dir` session-only — 유지

### 1.2 비목표 (Non-Goals)

- **Full CC parity 금지** (§9 #2): hooks, mcpServers, outputStyles capabilities는 v3.2 이후.
- **Marketplace git/url/file sources 금지**: v3.0은 github + directory만.
- **Plugin hot-reload 금지**: CC의 `/reload-plugins`와 `clearPluginCache` 복잡성 — v3.2 이후 (gm#146).
- **Cowork plugins 금지**: Anthropic-internal 개념 (W1.5 §4.5).
- **Plugin-authored hooks 금지**: SPEC-V3-HOOKS-003 source precedence는 플러그인 tier를 포함하지 않음 (v3.0 scope).
- **Trust/signing 인증 체계 금지**: github-pinned commit hash만 권장; 암호학적 서명은 v3.2 이후.
- **Plugin analytics / telemetry 금지**: moai는 out-of-scope.
- **Managed enterprise plugin catalog 금지**: v4 이후.
- **Plugin upgrade auto-download 금지**: 명시적 `moai plugin update <name>` 필요.

---

## 2. Scope (범위)

### 2.1 In Scope

**Plugin manifest** (`.claude-plugin/plugin.json` in plugin directory):
- Fields: `name` (required, slug), `version` (required, semver), `description`, `author`, `capabilities`, `engines.moai` (semver constraint, e.g., "^3.0.0")
- Capabilities (v1): `agents` (list of `.md` paths), `skills` (list of skill dir paths), `commands` (list of `.md` paths)

**Install scopes** (SPEC-V3-SCH-002와 일관):
- `user`: `~/.moai/plugins/` (개인 global)
- `project`: `.moai/plugins/` (팀 공유, checked-in)
- `local`: `.moai/plugins/local/` (머신별, gitignored)

**CLI surface** (under `moai plugin` subcommand family):
- `moai plugin install <source>` — source = `github:owner/repo@tag`, `github:owner/repo`, 또는 local path
- `moai plugin uninstall <name>`
- `moai plugin enable <name>` / `moai plugin disable <name>`
- `moai plugin update <name>`
- `moai plugin list` (모든 scope 통합 표시)
- `moai plugin validate <path>` (pre-install manifest + content walk)
- `moai plugin marketplace add <source>` — source = `github:owner/repo` 또는 local directory
- `moai plugin marketplace list`
- `moai plugin marketplace remove <name>`
- `moai plugin marketplace update [name]`

**Session-only plugin**:
- `--plugin-dir <path>` CLI flag (repeatable) — session-only inline plugin, 저장 없음

**Marketplace**:
- Registry: `marketplace.json` at marketplace root → lists available plugins with metadata + install URL
- Source types (v1): `github:owner/repo` (GitHub repo), `directory:/local/path` (local)
- Storage: `~/.moai/marketplaces/{name}/` (clone / symlink)

**Conflict resolution**:
- Duplicate agent `name` across plugins → install-time error
- Duplicate skill name across plugins → install-time error
- Duplicate command name across plugins → install-time error (NO silent override)

**Validation** (`moai plugin validate <path>`):
- Manifest schema check (SPEC-V3-SCH-001 validator/v10)
- Content walk: 선언된 `agents/`, `skills/`, `commands/` 파일 실존 확인
- Agent frontmatter 스키마 (SPEC-V3-AGT-001) 통과
- Skill SKILL.md 존재
- Command thin router pattern (coding-standards.md) 준수

**Trust**:
- `github:*` source → first install 시 `--trust` 플래그 필수
- `directory:*` source → commit hash 선택적 pin (`--pin <sha>`)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Capabilities: hooks, mcpServers, outputStyles (v3.2 이후)
- Marketplace source types: git (non-github), url, file (v3.2 이후)
- Hot-reload / `/reload-plugins` (v3.2 이후)
- Cowork plugin (Anthropic-internal)
- Plugin trust cryptographic signing
- Plugin telemetry
- Enterprise managed catalog
- Auto-download on startup
- Plugin-authored migrations (SPEC-V3-MIG-001 프레임워크의 플러그인 확장 — v3.2+)
- Plugin developer SDK / helper libraries
- `moai plugin search` — 본 SPEC v1에서는 `marketplace list`만
- Dependency resolution between plugins (플러그인 A가 B를 requires) — flat model only
- Plugin-defined custom config sections (`.moai/config/sections/*.yaml` 확장) — v3.2 이후

---

## 3. Environment (환경)

- 런타임: Go 1.26, moai-adk-go v3.0.0-beta.2+ (Phase 6a)
- 의존성:
  - SPEC-V3-SCH-001 (manifest 스키마 검증)
  - SPEC-V3-SCH-002 (user/project/local scope 정렬)
  - SPEC-V3-MIG-001 (plugin uninstall/update 시 migration 패턴 재사용)
- 외부 의존성:
  - `net/http` (stdlib, github tarball 다운로드)
  - 기존 git binary 호출 (`exec.Command("git", "clone")`) — github source 시
  - Binary 최소화를 위해 go-git 등 의존성 거부; plain git CLI shell-out
- 영향 디렉터리:
  - `internal/plugin/` (신규 패키지: manifest.go, installer.go, marketplace.go, registry.go, validator.go)
  - `internal/cli/plugin.go` (신규 subcommand tree)
  - `~/.moai/plugins/`, `.moai/plugins/`, `.moai/plugins/local/` (runtime)
  - `~/.moai/marketplaces/` (runtime)
  - `.gitignore` (template): `/.moai/plugins/local/` 추가
- 대상 OS: macOS / Linux / Windows 동등

---

## 4. Assumptions (가정)

- A-001 (High): 사용자 환경에 `git` CLI가 설치되어 있다. github source 시 `git clone --depth 1` 사용. 부재 시 명확한 에러 메시지.
- A-002 (High): `~/.moai/` 디렉터리는 `os.UserHomeDir()`로 해석 가능 — macOS/Linux/Windows 공통.
- A-003 (Medium): github 저장소는 public or 사용자 인증 설정 완료 상태. private repo 인증은 사용자 `git config` 의존.
- A-004 (Medium): 플러그인 manifest는 UTF-8 JSON (BOM 없음). 비 UTF-8은 parse 에러.
- A-005 (High): 플러그인이 선언한 capabilities가 로컬 manifest에 없는 파일을 가리키면 install-time error로 거부.
- A-006 (Medium): 사용자는 `moai plugin install github:owner/repo@v1.2.3` 형태로 tag를 pin 한다. Unpinned main/master는 `--allow-head` 플래그 필수.
- A-007 (High): Plugin list에서 효과적 상태(`enabled` / `disabled`)는 `~/.moai/plugins/installed.json` (user scope) 또는 `.moai/plugins/installed.json` (project scope)에 기록.
- A-008 (Medium): 플러그인 install 시 target scope (user/project/local) 미지정 → default `project` (CC는 `user`와 다름을 명시; moai는 팀 중심 기본).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-PLG-001-001 (Ubiquitous)**
시스템은 **항상** 플러그인 manifest를 `<plugin-root>/.claude-plugin/plugin.json` 경로에서 로드하고, JSON 스키마(SPEC-V3-SCH-001 validator/v10 패턴)로 검증한다.

**REQ-PLG-001-002 (Ubiquitous)**
시스템은 **항상** `engines.moai` 필드의 semver constraint를 현재 moai 버전과 매칭한다. 불일치 시 install/enable 거부.

**REQ-PLG-001-003 (Ubiquitous)**
시스템은 **항상** v1 capabilities를 `agents`, `skills`, `commands` 3종만 지원한다. manifest에 선언된 다른 capabilities (hooks, mcpServers, outputStyles)는 v3.0에서 warning 후 무시.

**REQ-PLG-001-004 (Ubiquitous)**
시스템은 **항상** 3 install scopes (user/project/local)를 SPEC-V3-SCH-002의 tier 구조와 일치하게 제공한다.

**REQ-PLG-001-005 (Ubiquitous)**
시스템은 **항상** 플러그인 간 agent/skill/command name 중복을 install-time에 감지하고 error로 거부한다. silent override 금지.

**REQ-PLG-001-006 (Ubiquitous)**
시스템은 **항상** 설치된 플러그인 목록을 `<scope-root>/installed.json`에 JSON 레지스트리로 유지한다. scope = user/project/local에 따라 위치 분리.

**REQ-PLG-001-007 (Ubiquitous)**
시스템은 **항상** marketplace 등록 정보를 `~/.moai/marketplaces/{name}/marketplace.json`에 저장한다.

### 5.2 Event-Driven Requirements

**REQ-PLG-001-010 (Event-Driven)**
**When** 사용자가 `moai plugin install <source>`를 실행하면, the 시스템 **shall** source 형식(`github:owner/repo@tag`, `directory:/path`)을 파싱해 해당 소스에서 manifest + content를 획득하고, validate 통과 후 scope 디렉터리에 배포한다.

**REQ-PLG-001-011 (Event-Driven)**
**When** 사용자가 `moai plugin install github:owner/repo`를 tag 명시 없이 실행하면, the 시스템 **shall** 경고 "HEAD is unpinned; add --allow-head to proceed" 후 중단한다.

**REQ-PLG-001-012 (Event-Driven)**
**When** 사용자가 `moai plugin uninstall <name>`를 실행하면, the 시스템 **shall** 해당 플러그인의 모든 capability 파일을 scope 디렉터리에서 제거하고, `installed.json`에서 엔트리 삭제한다.

**REQ-PLG-001-013 (Event-Driven)**
**When** 사용자가 `moai plugin validate <path>`를 실행하면, the 시스템 **shall** (a) manifest schema 검증, (b) 선언된 agent/skill/command 파일 실존 확인, (c) agent frontmatter 스키마 검증, (d) skill SKILL.md 존재 확인, (e) command thin router 패턴 준수를 모두 수행하고 결과를 report한다.

**REQ-PLG-001-014 (Event-Driven)**
**When** 사용자가 `moai plugin list`를 실행하면, the 시스템 **shall** 3개 scope의 `installed.json` 모두를 통합 표시하고, 각 플러그인의 `{name, version, scope, enabled, source}`를 테이블로 출력한다.

**REQ-PLG-001-015 (Event-Driven)**
**When** 사용자가 `moai plugin marketplace add <source>`를 실행하면, the 시스템 **shall** marketplace.json을 fetch하고 `~/.moai/marketplaces/{name}/`에 저장한다.

**REQ-PLG-001-016 (Event-Driven)**
**When** 사용자가 `moai plugin update <name>`을 실행하면, the 시스템 **shall** 해당 플러그인의 source 재fetch, manifest 비교, 버전 상승 시에만 capability 파일 교체한다.

**REQ-PLG-001-017 (Event-Driven)**
**When** moai CLI 실행 중 `--plugin-dir <path>` 플래그(repeatable)가 제공되면, the 시스템 **shall** 해당 경로를 session-only plugin으로 로드하고, `installed.json`에는 기록하지 않는다.

### 5.3 State-Driven Requirements

**REQ-PLG-001-020 (State-Driven)**
**While** 플러그인이 `disabled` 상태인 동안, the 시스템 **shall** 해당 플러그인의 agent/skill/command를 활성 레지스트리에서 제외한다.

**REQ-PLG-001-021 (State-Driven)**
**While** 동일 name의 agent가 2개 이상 플러그인에서 활성(enabled)인 동안, the 시스템 **shall** moai 시작 시 error로 실패하고 사용자에게 한쪽 disable 안내한다.

**REQ-PLG-001-022 (State-Driven)**
**While** `github:` source 플러그인 첫 install 수행 중이며 `--trust` 플래그가 없는 상태라면, the 시스템 **shall** install을 거부하고 사용자에게 `--trust` 안내한다.

**REQ-PLG-001-023 (State-Driven)**
**While** `engines.moai` semver constraint가 현재 moai 버전과 충돌하는 동안, the 시스템 **shall** 해당 플러그인을 inactive 상태로 유지하고 `moai plugin list` 출력에서 `incompatible` 플래그 표시한다.

### 5.4 Optional Requirements

**REQ-PLG-001-030 (Optional)**
**Where** 사용자가 `moai plugin install --scope <user|project|local>`를 지정하지 않으면, the 시스템 **shall** default scope `project`를 적용한다.

**REQ-PLG-001-031 (Optional)**
**Where** 사용자가 `moai plugin list --json`을 사용하면, the 시스템 **shall** 기계 판독 가능한 JSON 배열로 출력한다.

**REQ-PLG-001-032 (Optional)**
**Where** 플러그인이 `author` 필드를 제공하면, the 시스템 **shall** `moai plugin list` 출력에 저자 정보를 함께 표시한다.

**REQ-PLG-001-033 (Optional)**
**Where** marketplace가 category/tags 메타데이터를 제공하면, the 시스템 **shall** `moai plugin marketplace list --filter <tag>` 필터링을 지원한다 (v1: 선택적).

### 5.5 Unwanted Behavior (Must Not)

**REQ-PLG-001-040 (Unwanted)**
시스템은 v3.0에서 hooks, mcpServers, outputStyles 플러그인 capability를 **활성화하지 않아야 한다**. manifest에 선언되어도 warning 후 무시.

**REQ-PLG-001-041 (Unwanted)**
시스템은 플러그인 간 name 충돌 시 silent override를 **허용하지 않아야 한다**. 반드시 install-time error.

**REQ-PLG-001-042 (Unwanted)**
시스템은 `github:` 플러그인을 `--trust` 없이 install을 **허용하지 않아야 한다**. 명시적 opt-in.

**REQ-PLG-001-043 (Unwanted)**
시스템은 플러그인을 자동 업데이트하지 **않아야 한다**. `moai plugin update <name>`는 사용자 명시 실행만.

**REQ-PLG-001-044 (Unwanted)**
시스템은 플러그인의 `local` scope 파일을 git commit 대상으로 **포함하지 않아야 한다**. `.gitignore`에 `.moai/plugins/local/` 자동 추가.

**REQ-PLG-001-045 (Unwanted)**
시스템은 플러그인의 `.claude-plugin/plugin.json` 외 manifest 이름을 **허용하지 않아야 한다** (CC parity 강제).

### 5.6 Complex Requirements

**REQ-PLG-001-050 (Complex)**
**While** `moai plugin install github:owner/repo@v1.2.3 --scope project`가 진행 중, **when** 원격 저장소에서 manifest validate가 실패하면, the 시스템 **shall**:
(a) 다운로드 파일을 임시 디렉터리에서 즉시 삭제,
(b) `installed.json` 수정하지 않음,
(c) 사용자에게 validation error detail 표시,
(d) exit code 1로 종료.

**REQ-PLG-001-051 (Complex)**
**While** `moai plugin list` 실행 중, **when** 3 scope에서 동일 name 플러그인이 다른 버전으로 존재하면, the 시스템 **shall** 활성(enabled) 중인 하나만 선택(precedence: local > project > user, SPEC-V3-SCH-002와 동일) 하고, 나머지는 `shadowed` 플래그로 표시한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

**AC-PLG-001-01**: Local directory 기반 샘플 플러그인(`.claude-plugin/plugin.json` + 1 agent + 1 skill + 1 command) 준비 후 `moai plugin install directory:./sample-plugin` 성공. `.moai/plugins/sample-plugin/` 생성, `installed.json` 갱신.

**AC-PLG-001-02**: `moai plugin validate ./sample-plugin` 실행 시 manifest + content walk 통과, exit code 0. 고의 corrupt manifest로 실패 케이스 exit code 1.

**AC-PLG-001-03**: `moai plugin install github:modu-ai/moai-plugin-example@v1.0.0 --trust` 실행 시 `git clone --depth 1 --branch v1.0.0` 후 `.moai/plugins/moai-plugin-example/` 배포. `--trust` 누락 시 거부.

**AC-PLG-001-04**: 두 플러그인이 동일 agent `researcher`를 선언 상태에서 두 번째 install 시 "duplicate agent name: researcher" error, 두 번째 플러그인 설치 차단.

**AC-PLG-001-05**: v1 capabilities (agents/skills/commands)만 manifest에서 선언한 경우 install 정상. v3.0 미지원 capability(hooks) 선언 시 `warning: hooks capability ignored in v3.0 (deferred to v3.2)` 표시.

**AC-PLG-001-06**: `engines.moai: "^4.0.0"` 선언한 플러그인 install 시 "engine mismatch: requires ^4.0.0, have v3.0.0" error.

**AC-PLG-001-07**: `moai plugin marketplace add github:example/moai-marketplace` 실행 후 `~/.moai/marketplaces/moai-marketplace/marketplace.json` 존재, `moai plugin marketplace list`에 표시.

**AC-PLG-001-08**: 3 scope에 동일 플러그인 3 버전 존재 시 `moai plugin list`에서 `local` 버전만 active, 나머지는 `shadowed`. precedence 검증.

**AC-PLG-001-09**: `moai plugin uninstall sample-plugin` 실행 후 `.moai/plugins/sample-plugin/` 디렉터리 부재, `installed.json` 엔트리 삭제. 해당 플러그인의 agent/skill/command 레지스트리에서 제거 확인.

**AC-PLG-001-10**: `moai --plugin-dir /tmp/inline-plugin plan "test"` 실행 시 inline 플러그인 capability 활성, `installed.json` 불변.

**AC-PLG-001-11**: `moai plugin disable sample-plugin` 후 agent/skill/command 레지스트리에서 제외, `enable` 시 재활성화.

**AC-PLG-001-12**: `.gitignore`에 `.moai/plugins/local/` 포함 확인 (`git check-ignore .moai/plugins/local/foo`).

---

## 7. Constraints (제약)

- **[HARD] SPEC-V3-SCH-001, SCH-002 선행**: manifest schema + 3-tier scope 의존.
- **[HARD] v3.0 capability 제한**: agents/skills/commands만. hooks/mcp/outputStyles v3.2 이후.
- **[HARD] Silent override 금지**: duplicate name은 반드시 error.
- **[HARD] github source trust**: 첫 install 시 `--trust` 명시 필수.
- **[HARD] `.claude-plugin/plugin.json` 경로 강제**: CC parity.
- **[HARD] Semver 엄수**: `engines.moai` constraint 충족 없이 enable 금지.
- Binary size 증가 ≤ 500 KB (plugin 패키지 + stdlib net/http).
- CLI plugin subcommand는 `moai plugin`와 `moai plugins` alias 양쪽 지원 (CC parity).
- Plugin 메타 저장 포맷은 JSON (CC와 일관). YAML 혼용 금지.
- Dependency resolution 없음 — flat model. 플러그인 간 depends_on 은 v3.2 이후.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk ID | 설명 | 확률 | 영향 | 완화 |
|---------|------|------|------|------|
| R-PLG-001-01 | github source 다운로드 시 저장소 악성 코드 | Medium | Critical | `--trust` 필수; commit hash pin 권장; plugin validate로 내용 사전 확인 |
| R-PLG-001-02 | `engines.moai` constraint 충족 안 되는 플러그인 대량 배포 후 사용자 혼란 | Medium | Medium | `moai plugin list`에 incompatible 플래그; marketplace에도 지원 moai 버전 표시 |
| R-PLG-001-03 | Duplicate name 충돌로 install 실패 빈번 | Low | Medium | `moai plugin list` 상세 출력; namespace prefix 권장 (`myorg-myplugin`) |
| R-PLG-001-04 | github rate limit으로 install 실패 | Medium | Low | 캐싱 디렉터리 (`~/.moai/cache/plugins/`); 재시도 전 적절한 backoff |
| R-PLG-001-05 | Git CLI 부재 환경에서 `github:` source 설치 불가 | Medium | Medium | `moai doctor` 에서 git 감지; 명확한 에러 + 설치 가이드 |
| R-PLG-001-06 | Marketplace.json 포맷 드리프트 | Low | Medium | 스키마 검증 (`internal/plugin/marketplace_schema.go`) |
| R-PLG-001-07 | v3.2에서 hooks capability 추가 시 기존 플러그인 재검증 필요 | Low | Low | v1 manifest는 불변 부분만 stable; capabilities 확장 시 backward compatible |
| R-PLG-001-08 | Scope precedence 혼동 (user가 override 안 되는 이유) | Medium | Low | `moai plugin list`에 shadowed 표시; docs-site 명시 |
| R-PLG-001-09 | Plugin install이 moai 콜드 스타트 지연 | Low | Low | Plugin 로딩은 lazy (subcommand invocation 시점); startup fast path 유지 |
| R-PLG-001-10 | 사용자가 local scope에 플러그인 install 후 git commit 실수 | Low | Medium | `.gitignore` template-first; `moai doctor` gitignore 검증 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal Config Schema Framework) — manifest schema 검증 의존
- **SPEC-V3-SCH-002** (Settings 3-tier Source Layering) — install scope 구조 의존
- **SPEC-V3-MIG-001** (Versioned Migration Framework) — plugin uninstall/update 시 migration 패턴 재사용 (선택적)
- **SPEC-V3-AGT-001** (Agent frontmatter v2) — plugin-declared agent 검증 의존 (다른 writer)
- **SPEC-V3-SKL-001** (Skill frontmatter v2) — plugin-declared skill 검증 의존 (다른 writer)

### 9.2 Blocks

- **SPEC-V3-CLI-001** (CLI Subcommand Restructure) — `moai plugin` subcommand family 정의 포함
- v3.2+ **SPEC-V3-PLG-002** (Plugin hooks/mcp/outputStyles 확장)
- v3.2+ **SPEC-V3-PLG-003** (Plugin marketplace git/url/file sources)

### 9.3 Related

- gap-matrix #140, #141, #142 (scope-reduced), #143, #144, #145
- W1.5 §4.1-§4.7 (CC plugin system)
- W1.1 §11 (skill context inline/fork — 플러그인 skill 관련)
- master-v3 §3.6 Theme 6, §9 open question #2 (scope 결정)
- CLAUDE.local.md §2 (Protected Directories), §15 (Language neutrality)

---

## 10. Traceability (추적성)

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| REQ-PLG-001-001, 002, 045 | `internal/plugin/manifest.go` + schema | `TestManifestSchemaValidation`, `TestEnginesMoaiMatch` |
| REQ-PLG-001-003, 040 | `internal/plugin/installer.go` capability filter | `TestV1CapabilitiesOnly`, `TestUnsupportedCapabilityWarning` |
| REQ-PLG-001-004, 030 | `internal/plugin/scope.go` | `TestThreeInstallScopes`, `TestDefaultScopeProject` |
| REQ-PLG-001-005, 021, 041 | `internal/plugin/conflict.go` | `TestDuplicateAgentNameRejected` |
| REQ-PLG-001-006 | `internal/plugin/registry.go` installed.json | `TestInstalledJsonRoundtrip` |
| REQ-PLG-001-007 | `internal/plugin/marketplace.go` | `TestMarketplaceStored` |
| REQ-PLG-001-010, 011, 042, 022, 050 | `internal/cli/plugin_install.go` | `TestInstallGithubWithTrust`, `TestInstallHEADRejected`, `TestInstallAtomicRollback` |
| REQ-PLG-001-012, 009 | `internal/cli/plugin_uninstall.go` | `TestUninstallRemovesFiles` |
| REQ-PLG-001-013 | `internal/cli/plugin_validate.go` + `internal/plugin/validator.go` | `TestValidatorContentWalk` |
| REQ-PLG-001-014, 031, 032, 051 | `internal/cli/plugin_list.go` | `TestListAllScopes`, `TestListJSON`, `TestListShadowed` |
| REQ-PLG-001-015 | `internal/cli/plugin_marketplace.go add` | `TestMarketplaceAdd` |
| REQ-PLG-001-016, 043 | `internal/cli/plugin_update.go` | `TestUpdateOnlyOnVersionBump` |
| REQ-PLG-001-017 | `internal/cli/root.go --plugin-dir` | `TestInlinePluginSession` |
| REQ-PLG-001-020, 023 | `internal/plugin/state.go` enabled flag | `TestDisabledPluginExcluded`, `TestIncompatibleInactive` |
| REQ-PLG-001-033 | `internal/cli/plugin_marketplace.go list --filter` | `TestMarketplaceFilter` |
| REQ-PLG-001-044 | `.gitignore` template + `moai doctor` | `TestLocalPluginsGitignored` |
| AC-PLG-001-01 ~ 12 | 해당 REQ 커버리지 전체 | `go test ./internal/plugin/...` + CLI e2e |

---

End of SPEC-V3-PLG-001.
