# References — External Sources for IDEA-003 Review

## Anthropic Official Documentation (2026 Stable)

### Claude Code Plugins & Marketplace

- [Discover and install prebuilt plugins through marketplaces — Claude Code Docs](https://code.claude.com/docs/en/discover-plugins)
- [Extend Claude with skills — Claude Code Docs](https://code.claude.com/docs/en/skills)
- [Agent Skills — Claude API Docs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview)
- [Plugins for Claude Code and Cowork — Anthropic](https://claude.com/plugins)

### Official Repositories

- [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official) — 13개 공식 plugin 카탈로그
- [anthropics/skills](https://github.com/anthropics/skills/) — Public agent skills repository (Apache 2.0)
- [claude-plugins-official/.claude-plugin/marketplace.json](https://github.com/anthropics/claude-plugins-official/blob/main/.claude-plugin/marketplace.json) — marketplace.json schema 예시

### Community Marketplace Aggregators

- [claudemarketplaces.com](https://claudemarketplaces.com/) — 4,200+ skills, 770+ MCP servers, 2,500+ marketplaces 추적
- [DeepWiki — Claude Code Plugin Marketplace & Discovery](https://deepwiki.com/anthropics/claude-code/4.1-plugin-marketplace-and-discovery)

### Guide Articles (검증된 2026 콘텐츠)

- [Claude Code Plugins: A 2026 Guide — Nimbalyst](https://nimbalyst.com/blog/claude-code-plugins-guide/)
- [Claude Code Plugins vs Skills — What's the Difference? (2026)](https://www.agensi.io/learn/claude-code-plugins-vs-skills-explained)
- [Claude Code Plugin Marketplace: Complete Guide (2026)](https://www.agensi.io/learn/claude-code-plugin-marketplace-guide)
- [Install Anthropic Marketplace Plugins in Claude Code — systemprompt.io](https://systemprompt.io/guides/getting-started-anthropic-marketplace)

## Scaffolding & Template Update Patterns

### 검증된 Drift Detection 모델

- [Hugo Blox — Update guide (drift handling, override preservation)](https://bootstrap.hugoblox.com/hugo-tutorials/update/) — "remove overrides → compare upstream → re-apply" 패턴
- [Brian Coords — WP/Woo Plugin Scaffolding in 2026](https://www.briancoords.com/my-wp-woo-plugin-scaffolding-in-2026/) — AI agent 기반 scaffold 진화
- [DevOps School — Template Scaffolding (Feb 2026)](https://www.devopsschool.nl/template-scaffolding/) — PR-based migration 패턴

### Backstage Software Catalog (Spotify 오픈소스)

- [Authorizing scaffolder tasks, parameters, steps, and actions — Backstage Docs](https://backstage.io/docs/features/software-templates/authorizing-scaffolder-template-details/) — 권한/idempotency 패턴
- [Backstage scaffolder template composability — Issue #15241](https://github.com/backstage/backstage/issues/15241) — 템플릿 합성성 토론
- [Top Backstage Plugins for Infrastructure Automation (2026 Edition)](https://stackgen.com/blog/top-backstage-plugins-for-infrastructure-automation-2026-edition)

### 참고 도구 (search 결과 미직접 포함, 검토자가 추가 참조)

- **Cookiecutter `cruft`** — https://github.com/cruft/cruft — 템플릿 drift 자동 감지 + 사용자 override 보존하며 upstream merge. 이 제안의 영감 모델.
- **Yeoman** — https://yeoman.io/ — generator-based scaffold update flow
- **Copier** — https://copier.readthedocs.io/ — modern template update with versioning

## MoAI-ADK Internal References (검토에 필수)

### Constitution & Rules

- `.claude/rules/moai/core/moai-constitution.md` — TRUST 5 + AskUserQuestion 규칙
- `.claude/rules/moai/development/skill-authoring.md` — Skill YAML frontmatter 스키마
- `.claude/rules/moai/development/coding-standards.md` — Template-First, thin command 패턴
- `.claude/rules/moai/design/constitution.md` — Frozen vs Evolvable zone 정책 (이 제안의 안전성 모델 영감)

### Existing Harness Infrastructure

- `.claude/skills/moai-meta-harness/SKILL.md` — 7-Phase harness workflow (Apache 2.0 from revfactory/harness)
- `.claude/skills/moai-harness-learner/SKILL.md` — Tier 4 진화 메커니즘
- `.claude/agents/moai/builder-harness.md` — Unified artifact factory (agents/skills/plugins/commands/hooks/MCP/LSP)

### CLAUDE.local.md (Local Development Guide)

- §2 Protected Directories — `.claude/`, `.moai/project/`, `.moai/specs/` 보호 규칙
- §15 16-language Neutrality — `internal/template/templates/` 언어 중립성 강제
- §18 Enhanced GitHub Flow — branch/merge/release 표준

## How to Use This References File

검토자는:

1. 먼저 Anthropic 공식 plugin docs (상단 4개 URL) 을 읽어 marketplace/scope 이해
2. Hugo update + DevOps School template scaffolding 패턴으로 drift detection 모범 사례 확인
3. `cruft` 모델을 별도 검색하여 3-way merge + version pinning 메커니즘 학습
4. MoAI internal references는 제안의 제약사항 (constitutional, 16-lang 등) 이해를 위해 참조
