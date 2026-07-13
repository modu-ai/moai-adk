---
title: Statusline システム — 3-line レイアウト完全ガイド
weight: 78
draft: false
---

Claude Code と moai-adk-go 統合のための **カスタム statusline システム** です。トークノミクスは測定から始まります — コンテキスト使用率(CW%)、プロンプトキャッシュヒット率、rate limit 消費率をターミナル下部に常時表示し、トークン運用状況をひと目で読み取れるようにします。Claude Code v2.1.139 から effort/thinking、v2.1.145 から workspace.repo + pr フィールドが stdin JSON に追加され、豊富なコンテキストを表示できます。

> MoAI ワークフローは PR 中心です。すべての SPEC は plan-PR → run-PR → sync-PR サイクルを生成するため、statusline に現在の PR 番号 + レビュー状態 + コンテキスト使用率 + handoff 勧告を即座に表示すると開発効率が大きく向上します。

## 概要

### 最終レイアウト (3-line v3)

```
🤖 Opus 4.7 │ 🧠 xhigh·t │ 💾 67% │ 🔅 v2.1.146 │ 🗿 v3.0.0-rc6 │ ⏳ 4h 52m │ 💬 MoAI
🪫 CW: ███████░░░ 72% (⚠️/clear) │ 🔋 5H: █████░░░░░ 56% (46m) │ 🔋 7D: █░░░░░░░░░ 13% (May 28)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk (🅱️ main ↑5 +2) │ 💾 +0 M1 ?1 │ 💌 PR #1234 (⌥approved)
```

- **Line 1 (Info)**: モデル · effort/thinking · キャッシュヒット率 · Claude Code バージョン · MoAI バージョン · セッション時間 · output style
- **Line 2 (Usage bars)**: CW (context window) · 5H (rolling) · 7D (rolling) — 各 bar は絵文字 + label + bar + % + reset 情報
- **Line 3 (Git/PR)**: ディレクトリ · リポジトリ+ブランチ統合 · git status · アクティブ SPEC task · PR 情報

### データフロー

```
Claude Code stdin (JSON)
    ↓
internal/statusline/types.go (StdinData パース)
    ↓
internal/statusline/builder.go (CollectMemory, CollectMetrics, etc.)
    ↓
internal/statusline/renderer.go (3-line v3 layout)
    ↓
.moai/status_line.sh → ターミナル表示
```

## Line 1 — Info (7 segments)

### Model

- **フォーマット**: `🤖 <model display name>`
- **データソース**: stdin `model.display_name` (または string shorthand)
- **例**: `🤖 Opus 4.7`, `🤖 Sonnet 4.6`, `🤖 Haiku 4.5`
- **非表示条件**: `model` field 不在または `data.Metrics.Model == ""`
- **セグメントキー**: `model`

### Effort / Thinking

- **フォーマット**: `🧠 <level>[·t]`
- **データソース**: stdin `effort.level` + `thinking.enabled` (Claude Code v2.1.139+)
- **Level 値**: `low` / `medium` / `high` / `xhigh` / `max`
- **`·t` サフィックス**: `thinking.enabled == true` のとき追加 (extended reasoning 有効)
- **例**:
  - `🧠 xhigh·t` (xhigh effort + thinking 有効)
  - `🧠 high` (high effort, thinking なし)
  - `·t` (effort 不在 + thinking のみ有効)
- **非表示条件**: `effort` + `thinking` の両方が不在 (effort.level 空文字列を含む)
- **セグメントキー**: `effort_thinking`

いまのセッションがどの推論深度で回っているかを常時確認できるという点で、このセグメントはモデルポリシーが実際に適用されているかを検証する窓でもあります。

### キャッシュヒット率

- **フォーマット**: `💾 <N>%` (N = cache_read / (cache_read + cache_creation) × 100、小数点切り捨て)
- **データソース**: stdin `current_usage.cache_read_tokens` + `current_usage.cache_creation_tokens`
- **例**: `💾 28%` (cache_read 2000, cache_creation 5000 → 2000/7000)
- **非表示条件**: `current_usage` 不在 · `cache_creation == 0` (fresh cache write なし) · 両方 0 — 値をでっち上げず静かに省略 (graceful degradation)
- **トグル**: statusline.yaml の `cache_hit: false` → 非表示 (default-on)
- **セグメントキー**: `cache_hit`
- **参考**: 同じ `💾` 絵文字が Line 3 Git Status (`💾 +N M? ?`) にも使用 — 本セグメントは Line 1 に位置しパーセント形式で区別。prompt-cache 再利用率モニタリング (SPEC-TOKEN-EFFICIENCY-001 P0-2)

キャッシュヒット率はコンテキストダイエットの効果測定器です — 常時ロード指針を減らすと、この数字が上がるのをすぐに確認できます。

### Claude Code バージョン

- **フォーマット**: `🔅 v<version>` (default) または `🔅 cc v<version>` (full mode)
- **データソース**: stdin `version` 文字列
- **例**: `🔅 v2.1.146`
- **非表示条件**: `version` 空文字列
- **セグメントキー**: `claude_version`

### MoAI バージョン

- **フォーマット**: `🗿 v<current>` または更新可能時 `🗿 v<current> -> 🗿 v<latest>`
- **データソース**: `.moai/config/sections/system.yaml` `moai.version` + バックグラウンド update checker の結果
- **例**:
  - `🗿 v2.20.0-rc1` (最新)
  - `🗿 v2.18.0 -> 🗿 v2.20.0-rc1` (更新勧告)
- **セグメントキー**: `moai_version`

### セッション時間

- **フォーマット**: `⏳ <X>h <Y>m` (≥1h) / `⏳ <X>m` (<1h) / `⏳ <X>d <Y>h` (≥24h)
- **データソース**: stdin `cost.total_duration_ms`
- **例**: `⏳ 4h 52m`, `⏳ 35m`, `⏳ 1d 3h`
- **セグメントキー**: `session_time`

### Output Style

- **フォーマット**: `💬 <style name>`
- **データソース**: stdin `output_style.name`
- **例**: `💬 MoAI`, `💬 R2-D2`, `💬 default`
- **非表示条件**: `output_style.name` 空文字列
- **セグメントキー**: `output_style`

## Line 2 — Usage Bars (3 segments)

### CW (Context Window)

- **フォーマット**: `<icon> CW: <bar> <pct>% [(⚠️/clear)]`
- **データソース**:
  - bar: `context_window.context_window_size` × auto-compact threshold (default 85%) → scaled budget
  - パーセンテージ: `context_window.used_percentage` (事前計算) または `current_usage` tokens 合算
  - `(⚠️/clear)` 有効条件: `shouldShowHandoffGuide(data) == true`
- **絵文字**:
  - `🔋` (正常、<50% scaled)
  - `🪫` (警告、50-79% scaled)
  - `🪫` (危険、≥80% scaled、色追加)
- **`(⚠️/clear)` handoff suffix**:
  - 1M context モデル (Opus 4.7): used_percentage ≥50% (raw context_window_size 基準)
  - 200K context モデル (Sonnet/Haiku): used_percentage ≥90%
  - 意味: 次の turn 開始前に `/clear` 勧告 + paste-ready resume message 活用
- **例**: `🪫 CW: ███████░░░ 72% (⚠️/clear)`
- **セグメントキー**: `context`

### 5H (5 時間 rolling rate limit)

- **フォーマット**: `🔋 5H: <bar> <pct>% [(<reset>)]`
- **データソース**: stdin `rate_limits.five_hour.{used_percentage, resets_at}`
- **Reset フォーマット**:
  - <60 分: `(Nm)` (例: `(47m)`)
  - <24 時間: `(Nh Nm)` (例: `(2h 15m)`)
  - ≥24 時間: `(Mon DD)` (例: `(May 28)`)
- **例**: `🔋 5H: █████░░░░░ 56% (47m)`
- **データ不在**: `rate_limits.five_hour == null` → bar 0%, reset `(rolling)`
- **セグメントキー**: `usage_5h`

### 7D (7 日 rolling rate limit)

- **フォーマット**: `🔋 7D: <bar> <pct>% [(<reset>)]`
- **データソース**: stdin `rate_limits.seven_day.{used_percentage, resets_at}`
- **Reset フォーマット**: `(Mon DD)` (絶対日付)
- **例**: `🔋 7D: █░░░░░░░░░ 13% (May 28)`
- **セグメントキー**: `usage_7d`

サブスクリプションプランのユーザーにとって 5H/7D bar は事実上の予算ゲージです — rate limit が枯渇する前に重い作業を配置するか、CG モードで GLM ワーカーに渡すかを、この 2 つの bar を見て判断できます。

## Line 3 — Git / PR (5 segments)

### Directory

- **フォーマット**: `📁 <directory name>`
- **データソース**: stdin `workspace.project_dir` (basename) または `cwd`
- **例**: `📁 moai-adk-go`, `📁 my-project`
- **非表示条件**: `data.Directory` 空文字列
- **セグメントキー**: `directory`

### Repo + Branch (統合セグメント)

- **フォーマット**: `🔀 <owner>/<name> (🅱️ <branch>[ ↑N][ ↓N][ +N])`
- **データソース**:
  - `🔀 owner/name`: stdin `workspace.repo.{host, owner, name}` (Claude Code v2.1.145+)
  - `🅱️ branch`: ローカル git `branch --show-current`
  - `↑N`: ahead count (origin/<branch> 対比)
  - `↓N`: behind count
  - `+N`: dirty count = Modified + Staged + Untracked
- **例**:
  - `🔀 modu-ai/moai-adk (🅱️ main ↑3 +2)` (repo + branch + ahead + dirty)
  - `🔀 modu-ai/moai-adk (🅱️ main)` (clean branch, no ahead)
  - `🔀 (🅱️ feat/auth ↑2 ↓1 +6)` (repo 情報不在の fallback)
- **非表示条件**:
  - branch 空文字列 → セグメント全体を非表示
  - repo nil 時は fallback (括弧内の branch のみ表示)
- **Worktree モード**: `worktree` segment 有効時、branch に `[WT] ` prefix
- **セグメントキー**: `git_branch` (combined)

### Git Status

- **フォーマット**: `💾 +<staged> M<modified> ?<untracked>`
- **データソース**: ローカル git `git status --porcelain` のパース
- **例**: `💾 +0 M1 ?1` (staged 0, modified 1, untracked 1)
- **非表示条件**: git 利用不可
- **参考**: 以前の mailbox 4 種絵文字 (`📬`/`📫`/`📪`/`📭`) は廃止、統一された `💾` を使用
- **セグメントキー**: `git_status`

### Task (アクティブ SPEC workflow)

- **フォーマット**: `📋 [<command> <SPEC-ID>-<stage>]`
- **データソース**: `~/.moai/state/last-session-state.json` の `active_task` フィールド (該当ファイルの書き込み時のみ表示)
- **例**: `📋 [/moai run SPEC-V3R5-STATUSLINE-001-implement]`
- **非表示条件**: ファイル不在または `active_task` nil → segment 非表示
- **セグメントキー**: `task` (opt-in default off)

### PR (アクティブ GitHub Pull Request)

- **フォーマット**: `💌 PR #<number> (⌥<review_state>)` (state ありのとき) / `💌 PR #<number>` (state 空文字列)
- **データソース**: stdin `pr.{number, url, review_state}` (Claude Code v2.1.146+)
- **Review state 値**: `approved` / `pending` / `changes_requested` / `draft` / その他 (raw passthrough)
- **色分け** (review_state 部分):
  - `approved`: 緑 (Success)
  - `pending`: 黄 (Warning)
  - `changes_requested`: 赤 (Error)
  - `draft`: 灰 (Muted)
  - その他: 色なし (raw passthrough)
- **例**:
  - `💌 PR #1234 (⌥approved)` (緑)
  - `💌 PR #1023 (⌥pending)` (黄)
  - `💌 PR #7 (⌥changes_requested)` (赤)
  - `💌 PR #99 (⌥draft)` (灰)
  - `💌 PR #100` (state なし)
- **非表示条件**:
  - `pr` フィールド不在 (PR なしまたは v2.1.145 以下)
  - `pr.number == 0`
  - `SegmentPR` config が明示的に false
- **セグメントキー**: `pr` (default on per v2.20.0-rc1)

## 設定

### 基本構造

`.moai/config/sections/statusline.yaml` で segment の有効化を管理します。

```yaml
statusline:
  theme: catppuccin-mocha    # カラーテーマ
  segments:
    # Line 1
    model: true
    effort_thinking: true
    claude_version: true
    moai_version: true
    session_time: true
    output_style: true

    # Line 2
    context: true
    usage_5h: true
    usage_7d: true

    # Line 3
    directory: true
    git_branch: true       # combined repo+branch
    git_status: true
    task: true             # opt-in default off in older versions
    pr: true               # default on per v2.20.0-rc1
    worktree: false
```

### Segment 有効化マトリクス

| セグメント | ライン | デフォルト有効 | stdin field |
|---------|------|----------|-------------|
| `model` | L1 | ✓ | `model.display_name` |
| `effort_thinking` | L1 | ✓ | `effort.level` + `thinking.enabled` |
| `claude_version` | L1 | ✓ | `version` |
| `moai_version` | L1 | ✓ | (ローカル config) |
| `session_time` | L1 | ✓ | `cost.total_duration_ms` |
| `output_style` | L1 | ✓ | `output_style.name` |
| `context` | L2 | ✓ | `context_window.*` |
| `usage_5h` | L2 | ✓ | `rate_limits.five_hour.*` |
| `usage_7d` | L2 | ✓ | `rate_limits.seven_day.*` |
| `directory` | L3 | ✓ | `workspace.project_dir` |
| `git_branch` (combined) | L3 | ✓ | `workspace.repo.*` + local git |
| `git_status` | L3 | ✓ | local git |
| `task` | L3 | opt-in | `~/.moai/state/last-session-state.json` |
| `pr` | L3 | ✓ (v2.20.0-rc1+) | `pr.*` (Claude Code v2.1.146+) |
| `worktree` | L3 | ✗ opt-in | `workspace.git_worktree` |

## Handoff Guide — `(⚠️/clear)` 勧告基準

CW bar の `(⚠️/clear)` suffix は、コンテキスト使用量がモデル別の閾値を超えると有効になります。これは SSE stall リスクを事前に防ぎ、paste-ready resume message の活用を推奨する視覚的なマーカーです。

| モデルクラス | Context Window | 閾値 | 勧告時点 |
|------------|----------------|--------|----------|
| **1M context** (Opus 4.7) | 1,000,000 tokens | **≥50%** | ~500K トークン使用 |
| **200K context** (Sonnet, Haiku) | 200,000 tokens | **≥90%** | ~180K トークン使用 |
| その他 / 不明 | — | 表示しない | (安全な default) |

> 閾値は `internal/statusline/renderer.go shouldShowHandoffGuide()` 関数で強制されます。この閾値は `.claude/rules/moai/workflow/context-window-management.md` の HARD rule と一致します。

有効化時のユーザーフローは次のとおりです。

1. `(⚠️/clear)` marker の表示
2. 進行中の作業を `progress.md` などに保存
3. orchestrator が paste-ready resume message を生成 (session-handoff.md の 6-block フォーマット)
4. `/clear` 実行後に resume message を貼り付け
5. 新しいセッションで作業を継続

## stdin JSON スキーマ参照

Claude Code が statusline スクリプトに渡す stdin JSON の全フィールド一覧は [公式 docs Available data](https://code.claude.com/docs/en/statusline#available-data) を参照してください。moai-adk-go は次のフィールドを活用します。

```json
{
  "session_id": "abc...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/path/to/cwd",
  "model": {"id": "claude-opus-4-7", "display_name": "Opus 4.7"},
  "workspace": {
    "current_dir": "...",
    "project_dir": "...",
    "git_worktree": "feature-xyz",
    "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"}
  },
  "version": "2.1.146",
  "output_style": {"name": "MoAI"},
  "cost": {
    "total_cost_usd": 1.234,
    "total_duration_ms": 17520000,
    "total_lines_added": 156,
    "total_lines_removed": 23
  },
  "context_window": {
    "used_percentage": 62,
    "context_window_size": 1000000,
    "total_input_tokens": 620000,
    "total_output_tokens": 0,
    "current_usage": {
      "input_tokens": 8500,
      "output_tokens": 1200,
      "cache_creation_input_tokens": 5000,
      "cache_read_input_tokens": 605300
    }
  },
  "exceeds_200k_tokens": true,
  "effort": {"level": "xhigh"},
  "thinking": {"enabled": true},
  "rate_limits": {
    "five_hour": {"used_percentage": 56, "resets_at": 1779286800},
    "seven_day": {"used_percentage": 13, "resets_at": 1779832400}
  },
  "pr": {
    "number": 1234,
    "url": "https://github.com/modu-ai/moai-adk/pull/1234",
    "review_state": "approved"
  }
}
```

## バージョン履歴

- **v2.20.0-rc1 layout v3** (2026-05-22): 3-line layout 再設計 — repo+branch 統合セグメント、directory L3 先頭、`🪫 CW:` 絵文字を前へ、`(⚠️/clear)` handoff suffix、`💾` git status 統一、`💌 PR #N (⌥state)` 形式
- **v2.20.0-rc1 STATUSLINE-STDINFIELDS-001** (2026-05-21): `workspace.repo` + `exceeds_200k_tokens` + `pr` stdin フィールドマッピング追加、1M context handoff threshold 75% → 50%
- **v2.20.0-rc1 STATUSLINE-V2145-001** (2026-05-20): PR segment 追加 (v2.1.145+ stdin)、4-locale docs 同期
- **v2.1.139** (Claude Code): `effort.level` + `thinking.enabled` stdin JSON 追加
- **v2.1.146** (Claude Code): `workspace.repo` + `pr` stdin JSON 追加

## トラブルシューティング

### Statusline に PR が出ない

- Claude Code バージョン確認: `🔅 v2.1.146` 以上が必要 (v2.1.145 は stdin に `pr` フィールドを含まない)
- 現在の branch に OPEN PR があるか確認: `gh pr view`
- `statusline.yaml` に `pr: false` と明示されていないか確認

### `(⚠️/clear)` が表示されない

- 1M context モデル: used_percentage 50% 未満 → 正常 (まだ閾値未達)
- 200K context モデル: used_percentage 90% 未満 → 正常
- 閾値超過なのに表示されない: `shouldShowHandoffGuide` 関数の `MemoryData.ContextWindowSize` マッピングを確認 (boundary defect の可能性)

### 色が表示されない

- ターミナルが ANSI 256-color をサポートしているか確認
- `theme: catppuccin-mocha` が環境に適合しているか確認
- `NO_COLOR=1` 環境変数の設定有無を確認

### 検証コマンド

```bash
# stdin fixture で statusline の実出力を確認
NOW=$(date +%s)
echo '{"session_id":"test","model":{"display_name":"Opus 4.7"},"workspace":{"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}},"version":"2.1.146","output_style":{"name":"MoAI"},"context_window":{"used_percentage":62,"context_window_size":1000000},"exceeds_200k_tokens":true,"effort":{"level":"xhigh"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":56,"resets_at":'$((NOW + 2820))'},"seven_day":{"used_percentage":13,"resets_at":'$((NOW + 518400))'}},"cost":{"total_duration_ms":17520000},"pr":{"number":1234,"url":"https://github.com/modu-ai/moai-adk/pull/1234","review_state":"approved"}}' | moai statusline
```

## `/cd` キャッシュ保存ディレクトリ切り替え (CC 2.1.169+)

Claude Code 2.1.169+ はセッションの作業ディレクトリを **プロンプトキャッシュを保持しながら** 変更する `/cd <path>` コマンドを提供します — statusline の `cwd` フィールドが新しいディレクトリを反映するよう更新されますが、進行中の推論コンテキストは再構築されません。これは新しいターミナルセッションを開くことに対するキャッシュ保存の代替手段です: `/cd` は蓄積されたコンテキストを維持し、新しいターミナルはゼロから cold-start します。statusline がコンテキスト損失なしに離れたい `cwd` を表示するとき (例: セッション中に L2 worktree へ切り替え)、`/cd` が摩擦の少ない経路です。resume-pattern 統合は [セッションハンドオフ](/ja/workflow-commands/moai-sync)を参照してください。

## 関連ドキュメント

- [Settings JSON](/ja/advanced/settings-json) — Claude Code `statusLine` フィールド設定
