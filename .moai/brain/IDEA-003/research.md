# IDEA-003 Research — Skill/Agent Catalog 정리 방안

> Generated: 2026-05-12 · Phase 3 of `/moai brain` workflow

## Topic

moai-adk 기본 설치 시 36개 skills + 29개 agents 전체 deploy 대신, 코어만 deploy하고 `moai-meta-harness`가 프로젝트별 필요 자산을 동적 생성하는 구조로 전환. 핵심 제약: **`moai update` 실행 시 기존 프로젝트도 catalog가 안전하게 동기화되어야 함** (사용자 손실 0).

## Key Findings

### Finding 1: Anthropic 공식 Plugin Marketplace는 2026년 표준 인프라

Claude Code는 **plugin** (배포 단위, skills + agents + hooks + MCP 묶음) 과 **skill** (단일 instruction set) 을 명확히 분리했고, 공식 marketplace (`anthropics/claude-plugins-official` repo, 13개 공식 plugin)를 운영합니다. `/plugin marketplace add` 로 GitHub/Git/local/URL 어디서든 marketplace 추가 가능하며 auto-update 옵션 존재.

- Plugin scopes: **marketplace** (built-in) / **user** (`~/.claude/plugins/`) / **project** (`.claude/plugins/`)
- Plugin trust dialog로 보안 보장
- Auto-update 후 `/reload-plugins` 알림
- Cross-tool: SKILL.md는 Cursor, Gemini CLI, Codex CLI, Antigravity IDE에서도 동작 (Agent Skills open standard)

**Implication**: MoAI의 core/optional/harness-generated 3-tier는 이미 Anthropic의 marketplace/user/project scope와 1:1 대응합니다. 별도 메커니즘 발명할 필요 없이 표준 위에 얹습니다.

### Finding 2: Context Budget 1% 룰 = 코어 슬림화의 결정적 정당화

> "Skill descriptions are loaded into context so Claude knows what's available. All skill names are always included, but if you have many skills, descriptions are shortened to fit the character budget. The budget scales at 1% of the model's context window. When it overflows, descriptions for the skills you invoke least are dropped first."

**계산** (Opus 4.7 1M context):
- 1M × 1% = 10,000 tokens가 skill description 예산
- 현재 MoAI 36개 skills × 평균 250자 (≈ 70 tokens) = ~2,500 tokens — 아직은 OK
- 그러나 사용자가 marketplace에서 추가 plugin install 시 즉시 overflow → MoAI 자체 skill의 description이 drop 가능
- 코어 18개로 줄이면 모든 description이 풀 텍스트로 유지됨 → **trigger 정확도 향상**

`/doctor` 명령으로 budget 상태 확인 가능 (Anthropic 공식 진단 도구).

### Finding 3: Drift Detection은 cruft / Hugo 모델이 검증된 표준

`moai update`의 catalog 동기화 안전성 패턴:

1. **Hugo theme 패턴**: override 파일을 임시 제거 → upstream 비교 → 사용자 결정 후 재적용
2. **Cookiecutter `cruft`**: 템플릿 변경 자동 drift 감지, 사용자 override 보존하며 upstream 변경 merge
3. **DevOps School PR-based migration**: 템플릿 변경 → bot이 PR 자동 생성 → 소유자 review → CI 검증 → deploy. 감사 흔적(audit trail) 완전 보존
4. **Backstage Scaffolder**: 권한/검증/idempotency 분리. 동일 scaffold 반복 실행 corruption 0

**Idempotent + Traceable**은 모든 모던 scaffolding 도구의 공통 invariant.

### Finding 4: Curated Bundles 패턴이 1,234-skill 시대의 표준

3rd-party marketplace는 1,234개 skill을 다 install하지 말고 role-based bundle (예: "Web Wizard" = frontend-design + api-design-principles + lint-and-validate + create-pr) 로 분류해 제공. **사용자는 본인 역할/도메인만 install**.

MoAI에 적용: 코어 18개 + "백엔드 팩", "프런트엔드 팩", "디자인 팩" 등 도메인 팩을 marketplace 형태로. harness가 프로젝트 인터뷰 결과로 적절한 팩을 자동 install.

### Finding 5: Built-in Skills는 이미 존재

`/simplify`, `/batch`, `/debug`, `/loop`, `/claude-api` 등이 Claude Code 기본 번들. 즉 Claude Code 자체가 minimal core + plugin install 모델로 동작 중. MoAI도 이 패턴을 따르면 사용자에게 친숙.

## Implications for MoAI

| 발견 | MoAI 의사결정 영향 |
|------|-------------------|
| Anthropic plugin marketplace 존재 | 자체 메커니즘 발명 X. 표준 plugin scopes 활용 |
| Context budget 1% 룰 | 코어 18개 슬림화는 **성능 최적화이자 trigger 정확도 보장** |
| cruft drift detection | `moai update --catalog-sync` 는 cruft 모델로 구현 (자동 감지 + 사용자 확인 + 백업) |
| Curated bundles | "moai pack add backend" 같은 도메인 팩 도입 검토 |
| /doctor budget 진단 | `moai doctor` 또는 `moai catalog status` 명령에 budget 표시 추가 |

## Limitations of Research

- **cargo-generate, cruft, Yeoman 업데이트 메커니즘 구체 동작 미확인** (검색 결과는 패턴 수준). Phase 5 Critical Evaluation에서 추가 확인 또는 implementation 시 정밀 조사 필요.
- **Anthropic plugin marketplace의 dependency resolution 동작** (한 plugin이 다른 plugin 요구 시) 미확인. MoAI가 marketplace 출판 시 영향 가능.
- **Plugin trust dialog의 UX**가 매번 뜨는지, 한 번만인지 확인 필요 (사용자 마찰).

## Sources

- [Anthropic — Discover and install prebuilt plugins through marketplaces (Claude Code Docs)](https://code.claude.com/docs/en/discover-plugins)
- [Anthropic — Extend Claude with skills (Claude Code Docs)](https://code.claude.com/docs/en/skills)
- [Anthropic — Agent Skills API Docs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview)
- [anthropics/claude-plugins-official — Official plugin marketplace repo](https://github.com/anthropics/claude-plugins-official)
- [anthropics/skills — Public agent skills repository](https://github.com/anthropics/skills/)
- [DeepWiki — Claude Code Plugin Marketplace & Discovery](https://deepwiki.com/anthropics/claude-code/4.1-plugin-marketplace-and-discovery)
- [agensi.io — Claude Code Plugins vs Skills (2026)](https://www.agensi.io/learn/claude-code-plugins-vs-skills-explained)
- [agensi.io — Claude Code Plugin Marketplace Complete Guide (2026)](https://www.agensi.io/learn/claude-code-plugin-marketplace-guide)
- [Nimbalyst — Claude Code Plugins: A 2026 Guide](https://nimbalyst.com/blog/claude-code-plugins-guide/)
- [systemprompt.io — Install Anthropic Marketplace Plugins in Claude Code](https://systemprompt.io/guides/getting-started-anthropic-marketplace)
- [Brian Coords — WP/Woo Plugin Scaffolding in 2026](https://www.briancoords.com/my-wp-woo-plugin-scaffolding-in-2026/)
- [DevOps School — Template Scaffolding Meaning, Examples, Use Cases (Feb 2026)](https://www.devopsschool.nl/template-scaffolding/)
- [Hugo Blox — Update guide (drift handling)](https://bootstrap.hugoblox.com/hugo-tutorials/update/)
- [Backstage Software Catalog — Scaffolder Authorization](https://backstage.io/docs/features/software-templates/authorizing-scaffolder-template-details/)
