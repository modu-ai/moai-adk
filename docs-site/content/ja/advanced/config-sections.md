---
title: 設定セクションリファレンス
weight: 71
draft: false
description: ".moai/config/sections/ の主要設定ファイル(handoff/delegation/llm/statusline/security/workflow/crosssession) キーリファレンス。"
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
      high: "glm-5.3"          # 1M context — Opusスロット
      medium: "glm-5.3"        # 1M context   — Sonnetスロット
      low: "glm-5.3"          # 1M context   — 軽量スロット
      fable: "glm-5.3"
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

ステータスラインテーマ、`github` セグメントが数えるホスティングサービス、そして16セグメントトグルを制御します。

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  # forge: gitlab             # github | gitlab | none (未指定なら origin のホストで判定)
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
| `forge` | `github` · `gitlab` · `none` のいずれか。`github` セグメントが開いている作業をどのホスティングサービスで数えるかを決めます。未指定なら origin リモートのホストで判定します — `github.com` なら `gh`、`gitlab.com` なら `glab`。セルフホストのインスタンスは名前に手がかりが無いため、そこでは明示的に指定します。認識できない値は判定へ戻らず何もレンダリングしないので、打ち間違いは誤った件数ではなくセグメントの不在として現れます |
| `segments` | 16セグメント個別トグル (唯一のランタイムレバー)。すべてデフォルトonで、非アクティブ状態はgraceful no-outputで処理 |

セグメントは3行に配置されます — 行1(モデル・バージョン・セッションメタ)、行2(コンテキストウィンドウ・API使用量バー)、行3(ディレクトリ・git・ワークフロー・PR)。

関連: [Statuslineシステム & PRセグメント](/ja/advanced/statusline).

## security.yaml — セキュリティ強化

組み込み `DefaultSecurityPolicy` パターンを**拡張** (交換ではなく)する追加セキュリティ設定です。SOLIDの開放-閉鎖原則に従い、core修正なしでconfigで拡張します。

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

## workflow.yaml — branch_guard

主チェックアウト（primary checkout）のブランチ状態を守る opt-in ガードです。1 つのチェックアウトを複数のセッションが同時に使うとき、片方が実行した `git switch` · `git checkout` · `git reset --hard` · `git stash` · `git rebase` は、もう一方のセッションの作業ツリーを何の合図もなく変えてしまいます。このガードはそれらのコマンドを主チェックアウトでのみ拒否します。

```yaml
workflow:
    branch_guard:
        enabled: false   # 配布時の既定値
```

| キー | 値 | 説明 |
|------|-----|------|
| `enabled` | `false`（既定） | ガードは完全に不活性です。判定のための `git rev-parse` すら実行しないため、付随コストがありません |
| `enabled` | `true` | 主チェックアウトでブランチ状態を変えるコマンドを拒否します。ワークツリー内ではこれまでどおり許可されます |

**既定が無効な理由。** このガードが防ぐ危険は、1 つのチェックアウトを複数セッションで共有するときにだけ生じます。1 人で使うリポジトリでは起こらない問題なので、配布版はガードを無効にしたまま出荷されます。複数セッションを同時に動かすリポジトリの管理者が、上のキーを自分で書いて有効にします。

**適用範囲。** ガードは主チェックアウトとワークツリーを区別し、ワークツリー内のブランチ操作は妨げません。`git status` · `git log` · `git diff` · `git fetch` のような読み取りコマンドと、`git stash list` · `git merge-base` は有効時でも通過します。

**例外と失敗の向き。** ブランチを作る必要のある git 担当エージェントは識別子で例外扱いされ、`MOAI_BRANCH_GUARD_EXEMPT=1` 環境変数でも回避できます。判定が不確実なとき（リポジトリでない、`git rev-parse` の失敗など）は拒否せず通し、監査ログだけを残します — 確かな根拠があるときにのみ拒否します。

ブランチを変える必要のある作業は、拒否させるのではなくワークツリーに移すのが定石です。手順は [moai worktree](/ja/cli-reference/worktree/) を参照してください。

## workflow.yaml — audit

クロスモデル監査バックエンド（`codex_audit` · `glm_audit` · `audit_multi`）が実際にどのモデルと effort で動くかを指定します。バックエンドごとに `{model, effort}` の 1 組で、配布時の既定値はすべて空です。

```yaml
workflow:
    audit:
        codex:
            model: ""   # 例: gpt-5.6-sol — codex が提供できるモデル id
            effort: ""  # 例: high — low | medium | high | xhigh | max
        glm:
            model: ""   # 例: glm-5.3
            effort: ""  # 例: max — low | high | max（z.ai の reasoning 状態名）
```

| キー | 説明 |
|------|------|
| `audit.codex.{model, effort}` | codex 監査がセッション開始とレビューターンに載せる 1 組。model が空か、codex が提供できない id ならピンを捨てて、既存の SSOT 解決に戻ります |
| `audit.glm.{model, effort}` | GLM 監査が z.ai リクエストに載せる 1 組。effort は z.ai の reasoning 状態名 `low` · `high` · `max` のみを受け付けます。それ以外の値では reasoning 指示を外し、モデルのピンだけを適用します |
| 両方の組が空 | このキーが存在する前とまったく同じ解決をします。ピンを書いたことのないプロジェクトは何も変わりません |

ピンが適用されるのは監査の入口だけです。タスク委任経路（`codex_task` · `glm_task`）のモデル解決には影響せず、同じフィールドはウェブコンソールの Audit パネルでも編集できます。

## workflow.yaml — todo

バックログキュー (todo) の**案内表面**を切るスイッチです。セッション開始時のキュー要約行、ステータスラインの TODO セグメント、自然言語リクエストの todo ワークフローへの推論ルーティング — この 3 つがこのキー 1 つで静かになります。

```yaml
workflow:
    todo:
        enabled: false   # 明示的なオフ — キーがなければオンと読まれます
```

| キー | 値 | 説明 |
|----|-----|------|
| `todo.enabled` | (キーなし、既定) | オン。配布テンプレートにはこのブロックがそもそもなく、ほとんどのプロジェクトが置かれている状態です |
| `todo.enabled` | `false` | セッション開始時の要約・ステータスラインの TODO セグメント・スキルの自動ルーティングが切れます |

**切ってもコマンドは消えません。** `moai todo` CLI は登録されたまま全動詞が動作し、名前で直接呼んだ `/moai todo` もそのまま実行されます。この境界は意図されたものです — スイッチは「頼んでいない人にキューを見せる表面」だけを静かにし、キューを実際に使う人の経路には触れません。設定ファイルを読めない失敗経路でもオンと解釈されます (fail-open)。

小さな一回きりのプロジェクトで、セッションごとに出るバックログ要約がノイズに感じられるときにこのキーを使います。キュー自体の運用方法は [moai todo](/ja/utility-commands/moai-todo/) のページにあります。

## crosssession.yaml — セッション間メッセージ

自分の他の Claude Code セッションから届くメッセージを、このセッションがどう扱うかを決めます。`moai cc` · `moai glm` · `moai cg` のランチャーが起動時にこれらの値を一時的な `--settings` ファイルへ移し、ウェブコンソールは設定 seam を通じてこのファイルを編集します。ランチャーを経由せず素の `claude` で起動したセッションは、このファイルを読みません。

```yaml
crosssession:
  inbound: ""             # "" | accept | hold | refuse
  isolate_machines: false # デフォルト — マシン外へ出るメッセージに承認を求めない
  dialog_expiry: ""       # "" | 60s | 5m | 10m | never
```

| キー | 値 | 説明 |
|------|-----|------|
| `inbound` | `""` (デフォルト) | 2つのセッションの権限モードの種別から、Claude Code がメッセージごとに判断します |
| `inbound` | `accept` | メッセージをそのまま配送します。人が見ていないワーカーセッションにメッセージを受け取らせるにはこの値が必要です |
| `inbound` | `hold` | メッセージごとに承認を待たせます。こうして保留したメッセージは期限切れにならず、後で `accept` が適用されたときに配送されます |
| `inbound` | `refuse` | メッセージを破棄します |
| `isolate_machines` | `false` (デフォルト) | **このマシンの外のセッションへメッセージが出るとき、承認を求めません。** 同じマシン内のセッション同士は値に関わらずマシンを離れませんが、マシンの外のセッションへ向かうメッセージは Anthropic のサーバーを経由します — その経路を無承認で許すという意味なので、デフォルトのままにするかどうかは判断して決めてください |
| `isolate_machines` | `true` | メッセージがマシンを離れる前に必ず明示的な承認を求めます(`bypassPermissions` モードでも)。どの設定スコープであれ1つでも `true` があれば適用されるため、リポジトリにチェックインされたプロジェクトファイルは要求を有効化できても無効化はできません — 無効化するには `true` を書いたスコープをすべて取り除きます |
| `dialog_expiry` | `""` (デフォルト) | Claude Code の5分のデフォルトをそのまま使います |
| `dialog_expiry` | `60s` · `5m` · `10m` · `never` | **デフォルト判断で**保留されたメッセージの承認ダイアログの期限です。`never` はセッションが終わるまで保留します。明示的な `inbound: hold` で保留したメッセージには適用されません |

## 関連ドキュメント

- [settings.jsonガイド](/ja/advanced/settings-json) — Claude Codeランタイム設定
- [ハーネスプロファイルと評価](/ja/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/ja/cli-reference/doctor) — `moai doctor config` で結合設定検査
