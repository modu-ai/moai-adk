---
title: 設定セクションリファレンス
weight: 71
draft: false
description: ".moai/config/sections/ の主要設定ファイル(handoff/delegation/llm/statusline/security) キーリファレンス。"
---

MoAI-ADKのプロジェクト設定は `.moai/config/sections/` 配下の複数のYAMLファイルに分割されています。[settings.jsonガイド](/ja/advanced/settings-json)がClaude Codeランタイム設定を扱うのに対し、このページはMoAI-ADK自身の動作を制御する主要セクションファイルのキーを整理します。

{{< callout type="info" >}}
**1行要約**: `settings.json` はClaude Codeに何を許可するかを定義し、`.moai/config/sections/*.yaml` はMoAI-ADKがどのようにオーケストレーションするかを定義します。
{{< /callout >}}

## handoff.yaml — 自動レジュームハンドオフ

セッション境界で保存されたハンドオフをどのように処理するかを制御します。

```yaml
handoff:
    mode: manual   # manual | auto
    guide: false
```

| キー | 値 | 説明 |
|------|-----|------|
| `mode` | `manual` (デフォルト) | 保存されたハンドオフを自動注入しない (opt-inベースラインUX) |
| `mode` | `auto` | `/clear` 時に保存されたハンドオフをセッションコンテキストに注入した後、audit-trailコピーに移動 |
| `guide` | `false` (デフォルト) | `true` 時、non-`/clear` セッション開始(startup/resume/compact)で待機中ハンドオフがあるというbest-effort stderrヒントを出力。情報提供のみでセッションをブロックしない |

関連: [自律継続ループ](/ja/advanced/autonomous-loops), [moai handoff](/ja/cli-reference/handoff).

## delegation.yaml — エージェントルーティングSSOT

`/moai` サブコマンド別のデフォルトスキル/エージェント割り当てマップです。オーケストレータが実行計画を構築する際(Analyze-First)、このマップを読み、どのエージェントをspawnし、どのスキルを注入するかを決定します。

```yaml
delegation:
    version: 1
    learning:
        observe: routing-ledger
        propose_via: harness-tier-ladder
        auto_apply: false          # Tier-4ゲート — ユーザー承認が必要
    subcommands:
        plan:
            agents: [manager-spec, plan-auditor, Explore]
            skills: [moai-workflow-spec, moai-foundation-thinking]
        # run / sync / project / fix / loop / ...
    domain_skills:
        backend:  [moai-ref-api-patterns, moai-domain-backend]
        security: [moai-ref-owasp-checklist, moai-ref-llm-security, ...]
    agents:
        manager-spec: [moai-workflow-spec, moai-foundation-thinking]
```

| ブロック | 説明 |
|----------|------|
| `learning` | ルーティング使用をappend-only元帳 (`.moai/state/routing-ledger.jsonl`, opt-in·fail-open) で管理し、ハーネス学習サブシステムが4-tier提案ラダーで更新提案。`auto_apply: false` — Tier-4変更は `AskUserQuestion` ユーザー承認が必要 |
| `subcommands` | サブコマンド別 `agents` (spawnする11個retainedエージェント) + `skills` (spawn時注入するworkflowスキル)。0個割り当ても有効 (オーケストレータが直接実行) |
| `domain_skills` | ミッションドメイン別注入スキル (spawn当たり0-3個)。ドメイン信号とマッチング |
| `agents` | エージェント別conditionalスキル (トリガー発生時on-demandロード) |

関連: [エージェントガイド](/ja/advanced/agent-guide), [スキルガイド](/ja/advanced/skill-guide).

## llm.yaml — バックエンド・プロファイルマトリクス

プロファイル、プロファイルマトリクス、エージェント別 override、GLM モデルマッピングを定義します。

```yaml
llm:
  profile: "medium"            # max | medium | low (アクティブマトリクス列)
  performance_tier: "medium"   # legacy エイリアス (profile 不在時に読み込み、high→max)
  profiles:                    # プロファイル列 → 6 グループ → {model, effort}
    max: { ... }               # 詳細表: プロファイルマトリクスページ
    medium: { ... }
    low: { ... }
  agent_overrides: {}          # エージェント別 {model, effort} override (任意)
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opusスロット
      medium: "glm-4.7"        # 202K context — Sonnetスロット
      low: "glm-4.5-air"       # 128K context — 軽量スロット
      fable: "glm-5.2"
```

| キー | 説明 |
|------|------|
| `profile` | アクティブなプロファイルマトリクス列 (`max`/`medium`/`low`)。空なら `medium` として解釈。全サブエージェント spawn の model+effort のソース |
| `performance_tier` | legacy エイリアスフィールド。`profile` がない場合のみ読み込まれ `high`→`max` に正規化 |
| `profiles` | プロファイル列別グループ → `{model, effort}` マトリクス。Go デフォルト値 (`template.DefaultProfileMatrix`) が欠落セルの権威ある fallback |
| `agent_overrides` | 正規エージェント名別 `{model, effort}` override。アクティブプロファイルのグループセルより優先 (カタログ+enum 検証) |
| `glm.base_url` | Z.AI Anthropic互換プロキシエンドポイント |
| `glm.models` | スロット別GLMモデルマッピング。GLMはClaudeの5段階effortを3個reasoning状態 (thinking-off / reasoning-high / reasoning-max) にcollapse |

関連: [プロファイルマトリクス](/ja/advanced/profile-matrix), [3-ティアエージェントアーキテクチャ](/ja/advanced/no-haiku-3tier).

## statusline.yaml — ステータスライン

ステータスラインテーマと16セグメントトグルを制御します。

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  segments:
    model: true
    context: true
    # ... 全16セグメント (すべてデフォルトon)
    task: true
    pr: true
```

| キー | 説明 |
|------|------|
| `theme` | 正確に2つのテーマが存在: `catppuccin-mocha` (デフォルト) または `catppuccin-latte` |
| `segments` | 16セグメント個別トグル (唯一のランタイムレバー)。すべてデフォルトonで、非アクティブ状態はgraceful no-outputで処理 |

セグメントは3行に配置されます — 行1(モデル・バージョン・セッションメタ)、行2(コンテキストウィンドウ・API使用量バー)、行3(ディレクトリ・git・ワークフロー・PR)。

関連: [Statuslineシステム & PRセグメント](/ja/advanced/statusline).

## security.yaml — セキュリティ強化

組み込み `DefaultSecurityPolicy` パターンを**拡張**(交換ではなく)する追加セキュリティ設定です。SOLIDの開放-閉鎖原則に従い、core修正なしでconfigで拡張します。

```yaml
security:
  extra_dangerous_bash_patterns:
    - 'curl\s+.*\|\s*(ba)?sh'
    - 'rm\s+-rf\s+/[^.]'
  extra_deny_patterns: []
  extra_ask_patterns: []
  permission:
    strict_mode: true
    session_rules: []
  sandbox:
    required: false
    network_allowlist: []
    env_scrub_extra: []
    docker_image: "alpine:latest"
```

| キー | 説明 |
|------|------|
| `extra_dangerous_bash_patterns` | 組み込みdenyパターンに**追加**される危険Bashコマンド正規表現 (大文字小文字無視) |
| `extra_deny_patterns` / `extra_ask_patterns` | 追加ファイルdeny/askパターン |
| `permission.strict_mode` | `true` 時、bypassPermissionsモードのエージェントspawnを拒否 |
| `sandbox.required` | `true` 時、`sandbox.justification` なしの `sandbox: none` エージェントを拒否 (デフォルトfalse) |
| `sandbox.network_allowlist` | デフォルト8ホストに**追加**される許可ネットワークホスト |
| `sandbox.env_scrub_extra` | デフォルトscrubリストに**追加**されるenv変数名 (AWS_*, GITHUB_TOKENなど) |
| `sandbox.docker_image` | dockerバックエンドデフォルトイメージ |

関連: [セキュリティノート](/ja/advanced/security-notes), [settings.jsonガイド](/ja/advanced/settings-json).

## 関連ドキュメント

- [settings.jsonガイド](/ja/advanced/settings-json) — Claude Codeランタイム設定
- [ハーネスプロファイルと評価](/ja/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ja/cli-reference/doctor) — `moai doctor config` で結合設定検査
