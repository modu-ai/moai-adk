---
title: 設定セクションリファレンス
weight: 71
draft: false
description: ".moai/config/sections/ の主要設定ファイル(handoff/delegation/llm/statusline/security/ralph/harness/quality/workflow/mx) キーリファレンス。"
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
  profile: "medium"            # high | medium | low (アクティブマトリクス列、max は high として読み込み)
  performance_tier: "medium"   # legacy エイリアス (profile 不在時に読み込み、同じ語彙)
  profiles:                    # プロファイル列 → 11 エージェント → {model, effort}
    high: { ... }              # 詳細表: プロファイルマトリクスページ
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
| `profile` | アクティブなプロファイルマトリクス列 (`high`/`medium`/`low`。旧 `max` は `high` のエイリアスとして読み込まれる)。空なら `medium` として解釈。全サブエージェント spawn の model+effort のソース |
| `performance_tier` | legacy エイリアスフィールド。`profile` がない場合のみ読み込まれ、`high`/`medium`/`low` の同じ語彙を共有するため正規化ステップは不要 |
| `profiles` | プロファイル列別のエージェント単位 → `{model, effort}` マトリクス (11 エージェント × 3 列 = 33 セル)。Go デフォルト値 (`template.DefaultProfileMatrix`) が欠落セルの権威ある fallback |
| `agent_overrides` | 正規エージェント名別 `{model, effort}` override。アクティブプロファイルのエージェントセルより優先 (カタログ+enum 検証) |
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

## ralph.yaml — Ralph Engine

`/moai loop` の診断ベースの反復修正ループ (Ralph Engine) の動作を制御します。

```yaml
ralph:
  max_iterations: 5         # 反復上限 (デフォルト 5; CLI --max が優先)
  auto_converge: true       # 停滞検出時に自動収束
  human_review: true        # レビュー段階で人の介入により中断
  lint_as_instruction: true # LSP diagnostic を次ターンの指示として注入
  warn_as_instruction: false # エラーがないときのみ warning も注入
```

| キー | 説明 |
|----|------|
| `max_iterations` | 反復上限 (デフォルト 5)。優先順位: CLI `--max` フラグ > `ralph.max_iterations` > `workflow.yaml loop_prevention.max_iterations` |
| `auto_converge` | N ターン連続の無進展時に自動的に収束と判定 |
| `human_review` | レビュー段階で人が目を通すよう中断 |
| `lint_as_instruction` | LSP diagnostic を `systemMessage` として注入し、AI が次のプロンプトとして受け取る (デフォルト true) |
| `warn_as_instruction` | エラーがないとき warning も一緒に注入 (デフォルト false) |

関連: [/moai loop](/ja/utility-commands/moai-loop), [自律連続ループ](/ja/advanced/autonomous-loops).

## harness.yaml — ハーネス深度・評価

ハーネス品質パイプラインの深度 (minimal/standard/thorough) と自動検出、評価者の memory scope、エスカレーションを定義します。

```yaml
harness:
  default_profile: "default"  # default | strict | lenient | frontend
  evaluator:
    memory_scope: per_iteration  # FROZEN — 変更不可 (design-constitution §11.4.1)
  mode_defaults:
    solo: auto
    team: auto
    cg: thorough                 # CG モードは常に thorough
  auto_detection:
    enabled: true
    rules:
      minimal:  # file_count <= 3 AND single_domain, ...
      standard: # file_count > 3 OR multi_domain, ...
      thorough: # security/payment キーワード、critical 優先度、...
  escalation:
    enabled: true
```

| ブロック | 説明 |
|------|------|
| `default_profile` | SPEC に `evaluator_profile` がないときに使うデフォルト評価プロファイル |
| `evaluator.memory_scope` | 評価者のメモリ範囲。`per_iteration` に固定 (FROZEN) |
| `mode_defaults` | 実行モード (solo/team/cg) 別のデフォルト深度 |
| `auto_detection.rules` | Complexity Estimator が minimal/standard/thorough に自動分類する条件 |
| `escalation` | 失敗時に上位深度へエスカレーション |

関連: [ハーネスプロファイルと評価](/ja/advanced/harness-profiles), [moai harness](/ja/cli-reference/harness).

## quality.yaml — 品質ゲート・開発方法論

開発モード (DDD/TDD)、カバレッジ目標、LSP 品質ゲート閾値、品質ゲート強制の有無を制御します。

```yaml
constitution:
  development_mode: tdd        # tdd | ddd | off
  enforce_quality: true
  test_coverage_target: 85
  lsp_quality_gates:
    enabled: true
    plan: { require_baseline: true }
    run:  { max_errors: 0, max_type_errors: 0, max_lint_errors: 0, allow_regression: false }
    sync: { max_errors: 0, max_warnings: 10, require_clean_lsp: true }
```

| ブロック | 説明 |
|------|------|
| `development_mode` | `tdd` / `ddd` / `off`。`/moai run` の cycle_type デフォルト値を決定 |
| `enforce_quality` | `true` なら品質ゲート違反時に run-phase が GREEN にならない |
| `test_coverage_target` | パッケージカバレッジの最小目標 (%)。クリティカルパッケージ (cli/template/hook) は 90%+ を推奨 |
| `lsp_quality_gates.plan` | plan-phase: LSP baseline をキャプチャするか |
| `lsp_quality_gates.run` | run-phase: エラー/型エラー/リントエラー 0、回帰禁止 |
| `lsp_quality_gates.sync` | sync-phase: エラー 0、warning ≤ 10、clean LSP を要求 |

関連: [SPEC ワークフロー](/ja/advanced/spec-workflow), [LSP ゲート](/ja/advanced/lsp-gates).

## workflow.yaml — ワークフロー状態

ワークフロー閾値、branch-state guard の opt-in、セッション worktree の opt-in、エージェンティックループの上限を制御します。

```yaml
workflow:
  agentic_loop:
    max_iterations: 10        # パイプラインレベルの completion-loop 上限
  branch_guard:
    enabled: false            # distributed default — hook が INERT (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001)
  session_worktree:
    enabled: false            # distributed default — 自動隔離 INERT
  loop_prevention:
    max_iterations: 5         # per-operation 診断 fix-loop 上限 (ralph.max_iterations とは別軸)
```

| ブロック | 説明 |
|------|------|
| `agentic_loop.max_iterations` | パイプラインレベルの completion-loop 上限 (`AgenticLoopConfig`) |
| `branch_guard.enabled` | Main-Checkout Branch-State Guard の有効化。**分散デフォルト `false`** — 共有チェックアウトでない 1 人開発環境では hook は評価しない。共有マルチセッションチェックアウトの運用者が opt-in |
| `session_worktree.enabled` | `moai init` / `moai profile` / `moai web` 時の自動 worktree 隔離の有効化。**分散デフォルト `false`** (SPEC-SESSION-WORKTREE-001)。`MOAI_SESSION_WORKTREE` env で上書き |
| `loop_prevention.max_iterations` | per-operation の診断 fix-loop 上限。`ralph.max_iterations` とは別軸 — ralph.yaml が優先 |

> branch_guard と session_worktree はどちらも分散デフォルト値が `false` です。単一ユーザーのリポジトリでこれらが `false` であることは意図された動作であり、欠陥ではありません。

関連: [Main-Checkout Branch Guard](/ja/advanced/branch-guard), [Worktree 統合](/ja/advanced/worktree).

## mx.yaml — @MX タグシステム

`@MX` コード注釈タグシステムのタグ種別、言語別検出、検証ルールを定義します。

```yaml
mx:
  version: "2.1"
  languages:
    go:      { enabled: auto, patterns: ["*.go"], exclude: ["*_generated.go", "vendor/**"] }
    python:  { enabled: auto, patterns: ["*.py"], exclude: ["**/__pycache__/**"] }
    # ... 16 言語を同列に (go, python, typescript, javascript, rust, java, kotlin,
    #     csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)
  tags:
    # 各タグ別の説明・有効化状態
  reason_required:  # @MX:REASON が必須のタグ一覧
    - WARN
    - DEBT
```

| ブロック | 説明 |
|------|------|
| `version` | MX タグシステムのスキーマバージョン |
| `languages` | 16 言語を同列に並べる。各言語はプロジェクトマーカー (`go.mod`, `pyproject.toml`, `Cargo.toml` 等) で自動検出 (`enabled: auto`) |
| `tags` | `@MX:NOTE` / `@MX:WARN` / `@MX:ANCHOR` / `@MX:TODO` / `@MX:SPEC` / `@MX:DEBT` / `@MX:LEGACY` などのタグ別メタデータ。`@MX:SPEC` は SPEC 関連を記録 (SPEC-MX-ASSOCIATION-001) |
| `reason_required` | `@MX:REASON` が必須のタグ一覧 (デフォルト: WARN, DEBT) |

> 特定の言語を "PRIMARY" に格下げしたり、一部だけ "enabled" にしたりしないでください — 16 言語はすべて同列です。

関連: [@MX タグプロトコル](/ja/advanced/mx-tag-protocol), [moai mx](/ja/cli-reference/mx).

## 環境変数オーバーライド

YAML セクションの一部の値は環境変数で上書きできます。環境変数はファイル値より優先されます (`internal/config/manager.go` の `applyEnvOverrides`)。

| 環境変数 | 対象 | 説明 |
|----------|------|------|
| `MOAI_DEVELOPMENT_MODE` | `constitution.development_mode` | `tdd` / `ddd` / `off` を強制 |
| `MOAI_LOG_LEVEL` | ログレベル | `debug` / `info` / `warn` / `error` |
| `MOAI_LOG_FORMAT` | ログフォーマット | `text` / `json` |
| `MOAI_NO_COLOR` | カラー出力 | `1` / `true` でカラーを強制 off |
| `MOAI_CONFIG_DIR` | config ディレクトリ位置 | `.moai/config/` の代わりに別のパスを使用 |

> 上記 5 個が config manager が実際に読む環境変数オーバーライドのすべてです。定数は `internal/config/envkeys.go` にあります。`MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` は現在実装されていません — 名前と会話言語は `user.yaml` / `language.yaml` のみを読みます。

## 関連ドキュメント

- [settings.jsonガイド](/ja/advanced/settings-json) — Claude Codeランタイム設定
- [ハーネスプロファイルと評価](/ja/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ja/cli-reference/doctor) — `moai doctor config` で結合設定検査
