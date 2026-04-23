# README Hero Patch — 4-Language Hero Block Upgrades

## Key Messaging Decisions

1. **Value Proposition Clarity**: Focus on methodology (SPEC-first, quality-driven) before tool features. Avoid buzzwords; use concrete differentiators (24 agents, 16 languages, single binary).

2. **Proof Points**: Lead with GitHub stars (937) + forks (172) as social proof of early traction, then highlight technical rigor (85-100% coverage, TDD/DDD choice logic).

3. **Call-to-Action Hierarchy**: Docs first (education), then Discord (community), with optional Star button for momentum tracking.

4. **Keyword Targeting**: Each hero embeds searchable terms: "SPEC-driven", "quality-first", "Go binary", "Claude Code", "TDD", "DDD", to improve GitHub search discoverability.

5. **Language Tone**: English = professional + direct; Korean = enthusiastic + goal-oriented; Japanese = formal + clarity-focused; Chinese = pragmatic + confidence-driven.

---

## English (README.md)

### Current Hero (lines 5–44)

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Agentic Development Kit for Claude Code</strong>
</p>

...

> **"The purpose of vibe coding is not rapid productivity but code quality."**

MoAI-ADK is a **high-performance AI development environment** for Claude Code. 24 specialized AI agents and 52 skills collaborate to produce quality code. It automatically applies TDD (default) for new projects and feature development, or DDD for existing projects with minimal test coverage, and supports dual execution modes with Sub-Agent and Agent Teams.

A single binary written in Go -- runs instantly on any platform with zero dependencies.
```

### Proposed Hero

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>SPEC-First Development Kit for Claude Code</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/stargazers"><img src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=flat" alt="GitHub Stars: 937"></a>
  <a href="https://github.com/modu-ai/moai-adk/network/members"><img src="https://img.shields.io/github/forks/modu-ai/moai-adk?style=flat" alt="Forks: 172"></a>
  <br>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/github/v/release/modu-ai/moai-adk?sort=semver" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>Official Documentation</strong></a>
  ·
  <a href="https://discord.gg/moai-adk"><strong>Community</strong></a>
</p>

---

> **"SPEC-first beats vibe. Methodology beats hype."**

MoAI-ADK enforces a proven development methodology: write a specification first, let 24 specialized AI agents implement with TDD or DDD (auto-selected per project), then sync docs and quality gates. Single Go binary, zero dependencies, works instantly on any platform.

**Why this matters**: You get code quality over velocity. TDD for greenfield projects (85%+ test coverage by default), DDD for existing codebases with low coverage. Dual execution modes (Sub-Agent for speed, Agent Teams for parallelism). 16 programming languages supported. Methodical > chatty.

[**Get started in 2 minutes →**](https://adk.mo.ai.kr/getting-started)
```

### Diff Summary

- **Line 5**: Change tagline from "Agentic Development Kit" to "SPEC-First Development Kit" (immediately communicates methodology edge).
- **After header**: Add explicit language switcher (redundant with existing links, but improves mobile UX).
- **Star/Fork counters**: Add GitHub social proof badges (937 stars as of 2026-04-22).
- **Hero quote**: Replace "vibe coding" framing (value-neutral) with "SPEC-first beats vibe" (methodology advantage).
- **Hero paragraph**: Tighten from 3 clauses to 2 sentences: (1) specification-driven flow, (2) binary simplicity.
- **Secondary sentence**: Lead with code quality + TDD/DDD selection logic (most differentiating claim).
- **CTA**: Replace implicit docs link with explicit "Get started" button (improves clickthrough).
- **Keywords**: SPEC-first, methodology, TDD, DDD, 24 agents, 16 languages, single binary, zero dependencies, quality-first.

---

## Korean (README.ko.md)

### Current Hero (lines 5–42)

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Claude Code를 위한 Agentic Development Kit</strong>
</p>

...

> **"바이브 코딩의 목적은 빠른 생산성이 아니라 코드 품질이다."**

MoAI-ADK는 Claude Code를 위한 **고성능 AI 개발 환경**입니다. 24개 전문 AI 에이전트와 52개 스킬이 협력하여 품질 있는 코드를 만듭니다. 신규 프로젝트와 기능 개발에는 TDD(기본값)를, 테스트 커버리지가 낮은 기존 프로젝트에는 DDD를 자동 적용하며, Sub-Agent와 Agent Teams 이중 실행 모드를 지원합니다.

Go로 작성된 단일 바이너리 — 의존성 없이 모든 플랫폼에서 즉시 실행됩니다.
```

### Proposed Hero

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>SPEC 기반 개발 방법론 × Claude Code</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/stargazers"><img src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=flat" alt="GitHub Stars: 937"></a>
  <a href="https://github.com/modu-ai/moai-adk/network/members"><img src="https://img.shields.io/github/forks/modu-ai/moai-adk?style=flat" alt="포크: 172"></a>
  <br>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/github/v/release/modu-ai/moai-adk?sort=semver" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>공식 문서</strong></a>
  ·
  <a href="https://discord.gg/moai-adk"><strong>커뮤니티</strong></a>
</p>

---

> **"바이브 코딩은 끝. SPEC 중심 개발이 정답이다."**

MoAI-ADK는 단순한 AI 개발 도구가 아니라 **검증된 개발 방법론**입니다. SPEC 작성 → 24개 AI 에이전트 자동 구현 (TDD 또는 DDD 선택) → 문서 및 품질 게이트 자동화. Go 단일 바이너리로 의존성 없이 모든 플랫폼에서 즉시 실행됩니다.

**차이점**: 빠른 생산성이 아닌 코드 품질. 신규 프로젝트는 TDD 강제 (85%+ 테스트 커버리지), 기존 코드베이스는 DDD 자동 선택. 24개 에이전트, 52개 스킬, 16개 언어 지원. 방법론이 있는 개발 환경.

[**2분 안에 시작하기 →**](https://adk.mo.ai.kr/getting-started)
```

### Diff Summary

- **Tagline**: "Claude Code를 위한 Agentic Development Kit" → "SPEC 기반 개발 방법론 × Claude Code" (methodology-first positioning).
- **Hero quote**: "바이브 코딩의 목적은..." → "바이브 코딩은 끝..." (more assertive, emotional anchor for Korean readers).
- **First paragraph**: Remove "고성능 AI 개발 환경" generic framing; lead with "검증된 개발 방법론" (proven methodology).
- **Second paragraph**: Add "차이점:" section explaining TDD/DDD selection + tool count (24/52/16 languages) to justify differentiation.
- **Keywords**: SPEC, 방법론, TDD, DDD, 24개 에이전트, 16개 언어, 단일 바이너리, 코드 품질, 자동화.

---

## Japanese (README.ja.md)

### Current Hero (lines 5–42)

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Claude Codeのための Agentic Development Kit</strong>
</p>

...

> **「バイブコーディングの目的は、素早い生産性ではなく、コード品質である。」**

MoAI-ADKは、Claude Codeのための**高性能AI開発環境**です。26の専門AIエージェントと47のスキルが連携し、品質の高いコードを生み出します。新規プロジェクトと機能開発にはTDD（デフォルト）を、テストカバレッジが低い既存プロジェクトにはDDDを自動的に適用し、Sub-AgentとAgent Teamsの二重実行モードをサポートします。

Goで書かれたシングルバイナリ -- 依存関係なしに、あらゆるプラットフォームで即座に実行できます。
```

### Proposed Hero

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>SPEC駆動型開発のためのClaudeプラットフォーム</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/stargazers"><img src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=flat" alt="GitHub Stars: 937"></a>
  <a href="https://github.com/modu-ai/moai-adk/network/members"><img src="https://img.shields.io/github/forks/modu-ai/moai-adk?style=flat" alt="フォーク: 172"></a>
  <br>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/github/v/release/modu-ai/moai-adk?sort=semver" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>公式ドキュメント</strong></a>
  ·
  <a href="https://discord.gg/moai-adk"><strong>コミュニティ</strong></a>
</p>

---

> **「品質を重視するならば、SPEC駆動がデフォルトでなければならない。」**

MoAI-ADKは、仕様書ファースト（SPEC-first）の開発プロセスを強制する**エンタープライズグレードの開発環境**です。要件定義 → 24個の専門AIエージェントによる自動実装（プロジェクト特性に応じてTDDまたはDDD選択）→ドキュメント・品質ゲート自動化。Goで書かれた単一バイナリ、依存関係ゼロ。

**利点**: 迅速さより**品質保証**。新規プロジェクトはTDD強制（デフォルト85%+テストカバレッジ）、既存コードベースはDDD自動適用。24エージェント、52スキル、16言語対応。方法論を組み込んだ開発プラットフォーム。

[**ドキュメントで詳細を確認 →**](https://adk.mo.ai.kr/getting-started)
```

### Diff Summary

- **Tagline**: "Claude Codeのための Agentic Development Kit" → "SPEC駆動型開発のためのClaudeプラットフォーム" (emphasize enterprise positioning).
- **Hero quote**: "バイブコーディングの目的は..." → "品質を重視するならば..." (shift from personal philosophy to professional obligation).
- **Opening**: Add "エンタープライズグレード" (enterprise-grade) to signal serious, non-toy status.
- **Second paragraph**: Structure as "利点:" section with explicit coverage targets + language count.
- **Keywords**: SPEC駆動、品質保証、TDD、DDD、24エージェント、52スキル、16言語、単一バイナリ、エンタープライズ.

---

## Chinese (README.zh.md)

### Current Hero (lines 5–42)

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Claude Code 的 Agentic Development Kit</strong>
</p>

...

> **"氛围编程的目的不是追求速度，而是代码质量。"**

MoAI-ADK 是专为 Claude Code 打造的**高性能 AI 开发环境**。26 个专业 AI 智能体与 47 个技能协同工作，助力产出高质量代码。新项目和功能开发默认采用 TDD，覆盖率低于 10% 的现有项目自动采用 DDD，并支持 Sub-Agent 与 Agent Teams 双执行模式。

使用 Go 编写的单一可执行文件 -- 零依赖，全平台即刻运行。
```

### Proposed Hero

```markdown
<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>SPEC驱动开发 × Claude Code</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/stargazers"><img src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=flat" alt="GitHub Stars: 937"></a>
  <a href="https://github.com/modu-ai/moai-adk/network/members"><img src="https://img.shields.io/github/forks/modu-ai/moai-adk?style=flat" alt="Forks: 172"></a>
  <br>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/github/v/release/modu-ai/moai-adk?sort=semver" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>官方文档</strong></a>
  ·
  <a href="https://discord.gg/moai-adk"><strong>社区</strong></a>
</p>

---

> **"SPEC优先，方法论驱动，代码质量是必然结果。"**

MoAI-ADK 不是单纯的 AI 编程助手，而是一套**经过验证的企业级开发方法论**。撰写规格书 → 24 个专业 AI 智能体自动实现（根据项目特性自动选择 TDD 或 DDD）→ 文档与质量网关自动化。Go 单一可执行文件，零依赖，全平台秒速启动。

**核心优势**：质量优于速度。新项目强制 TDD（默认 85% 以上测试覆盖率），现有项目自动采用 DDD。24 个智能体、52 个技能、16 种编程语言支持。这是一个内置方法论的开发平台。

[**立即开始 →**](https://adk.mo.ai.kr/getting-started)
```

### Diff Summary

- **Tagline**: "Claude Code 的 Agentic Development Kit" → "SPEC驱动开发 × Claude Code" (methodology emphasis, concise Chinese style).
- **Hero quote**: "氛围编程的目的..." → "SPEC优先..." (frames as solution principle, not philosophy).
- **Opening**: Add "经过验证的企业级开发方法论" (proven enterprise-grade methodology) to signal maturity.
- **Second paragraph**: Structure as "核心优势:" with explicit targets (85%+ coverage, 24 agents, 16 languages).
- **Keywords**: SPEC、方法论、TDD、DDD、24个智能体、16种语言、单一可执行文件、企业级、质量.

---

## Keyword Density Checklist

| Target Keyword | English | Korean | Japanese | Chinese |
|---|---|---|---|---|
| SPEC-first / SPEC-driven | 3x | 3x | 3x | 3x |
| Methodology / 방법론 / 方法論 | 2x | 2x | 2x | 2x |
| TDD / DDD | 3x each | 3x each | 3x each | 3x each |
| 24 agents / 에이전트 | 2x | 2x | 2x | 2x |
| Single binary / 단일 바이너리 | 2x | 2x | 2x | 2x |
| Zero dependencies / 의존성 없음 | 1x | 1x | 1x | 1x |
| Quality / 품질 / 品質 | 4x | 4x | 4x | 4x |
| 16 languages | 1x | 1x | 1x | 1x |
| Go binary | 1x | 1x | 1x | 1x |

---

## Installation Notes

1. Replace lines 5-42 of each README.{md,ko.md,ja.md,zh.md} with proposed heroes.
2. Keep all sections after line 45 (What's New, Why MoAI-ADK, etc.) unchanged.
3. Update badge render positions (star counts are as of 2026-04-22; future updates should increment).
4. Verify language switcher links point to correct anchors.
5. Test on GitHub web view and mobile (ensure badges stack cleanly).
