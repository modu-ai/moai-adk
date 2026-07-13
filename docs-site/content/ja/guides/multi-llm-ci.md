---
title: "Multi-LLM CI ガイド"
description: "GitHub Actions で複数の AI モデルによるコードレビューを自動化"
draft: false
weight: 10
---

MoAI-ADK の Multi-LLM CI 機能で、GitHub Actions に複数の LLM コードレビューを
設定する方法を案内します。レビュアーも 1 つのモデルに縛られる理由はありません —
モデルごとに強みと単価が異なるため、PR レビューにもマルチ LLM 配分というトークノミクス
の観点がそのまま適用されます。

## 概要

### Multi-LLM CI とは?

MoAI-ADK の Multi-LLM CI 機能は、GitHub Actions で複数の AI モデルによる同時コードレビューを行う統合 CI/CD パイプラインを提供します。

### サポートする LLM

| LLM | 提供者 | トリガー方式 | 特徴 |
|-----|--------|-------------|------|
| **Claude** | Anthropic | `/claude` コメント | Issue/PR レビュー、OAuth 認証 |
| **Codex** | OpenAI | PR open 自動 | プライベートリポジトリ専用 |
| **Gemini** | Google | PR open 自動 | API Key 認証 |
| **GLM** | Zhipu AI | PR open 自動 | トークン認証 |

## はじめる

### 前提条件

- MoAI-ADK のインストール (macOS · Linux · Windows)
- GitHub repository
- 各 LLM のアカウントと API トークン

### 初期設定

```bash
moai github init
```

このコマンドが行う作業:
- `.github/workflows/` ディレクトリの作成
- workflow テンプレートの配置
- composite actions の配置
- GitHub Secrets 設定ガイド

### LLM 認証設定

```bash
# Claude (OAuth)
moai github auth claude

# Codex (プライベートリポジトリ)
moai github auth codex

# Gemini
moai github auth gemini

# GLM
moai github auth glm
```

### GitHub Secrets の設定

各 LLM に必要な Secrets:
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth トークン
- `CODEX_AUTH_JSON` - Codex 認証 JSON (base64 エンコード)
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API Token

### 最初の PR テスト

PR を作成すると、自動で LLM Panel コメントが追加されます:

```markdown
## LLM Code Review Status

| LLM | Status |
|-----|--------|
| Claude | Pending (add `/claude` comment) |
| Codex | ✓ Ready |
| Gemini | ⚠️ Token missing |
| GLM | ✓ Ready |

Trigger individual reviews:
- Add `/claude` comment to trigger Claude
- Add `/codex` comment to trigger Codex
- Add `/gemini` comment to trigger Gemini
- Add `/glm` comment to trigger GLM
```

## LLM 認証の設定

### Claude の設定

#### OAuth トークンの発行

1. [Claude Code](https://claude.ai/download) をインストール
2. ログイン後、OAuth トークンを発行
3. `.claude/settings.local.json` に自動保存

#### moai github auth claude

```bash
moai github auth claude
```

**対話型設定プロセス:**
```
Claude OAuth トークンが見つかりません。
Claude Code をインストールしてログインしますか? (y/n): y

[確認済み] OAuth トークンが settings.local.json に保存されました。
GitHub Secret: CLAUDE_CODE_OAUTH_TOKEN に次の値を設定してください:
<token-value>
```

### Codex の設定 (プライベートリポジトリ専用)

#### 認証 JSON の作成

```json
{
  "token": "sk-...",
  "base_url": "https://api.openai.com/v1"
}
```

#### moai github auth codex

```bash
moai github auth codex
```

**対話型設定:**
```
OpenAI auth.json ファイルのパス: ~/.codex/auth.json
ファイルを読み取って GitHub Secret を生成します...
注意: Codex はプライベートリポジトリでのみ使用可能です (REQ-SEC-001)

生成された Secret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini の設定

```bash
moai github auth gemini
```

API Key の入力後、自動で GitHub Secret 設定ガイドを提供。

### GLM の設定

```bash
moai github auth glm
```

GLM トークンのパス (`~/.moai/.env.glm`) から自動で読み取り。

## Workflow テンプレートの理解

### llm-panel.yml

**トリガー:** PR opened

**役割:** 各 LLM の状態を視覚的に表示するパネルコメントの自動生成

**備考:** `/claude`、`/codex`、`/gemini`、`/glm` コメントで個別レビューをトリガー

### claude.yml / claude-code-review.yml

- **claude.yml**: Issue トリガー (ドラフトレビュー)
- **claude-code-review.yml**: PR トリガー (変更内容レビュー)

**特徴:** `/claude` コメントでのみトリガー

### codex-review.yml

**セキュリティ制約:**
- `private` リポジトリでのみ動作 (REQ-SEC-001)
- `visibility` チェックで公開リポジトリをブロック

**workflow:**
```yaml
private-guard:
  runs-on: ubuntu-latest
  steps:
    - name: Check Repository Visibility
      run: |
        if [[ "${{ github.repository_visibility }}" == "public" ]]; then
          echo "::error::Codex review is restricted to private repositories"
          exit 1
        fi
```

### gemini-review.yml

- 自動言語検出 (detect-language action)
- PR synchronized 時に自動トリガー

### glm-review.yml

- GLM 専用の環境設定 (setup-glm-env action)
- 環境変数の自動注入

### Composite Actions

#### detect-language

**入力:** repository ルートパス
**出力:** language 環境変数 (`detected_language`)

**サポート言語:** Go、Python、TypeScript、JavaScript、Rust、Java、Kotlin、C#、Ruby、PHP、Elixir、C++、Scala、R、Flutter、Swift (16 言語)

#### setup-glm-env

GLM チームモードで必要な環境変数の設定:
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## 高度な設定

### github-actions.yaml のカスタマイズ

#### 基本構造

```yaml
# .moai/config/sections/github-actions.yaml
llm_review:
  enabled: true
  runners:
    claude: true
    codex: true
    gemini: true
    glm: true
  triggers:
    on_pr_open: true
    on_comment:
      claude: "/claude"
      codex: "/codex"
      gemini: "/gemini"
      glm: "/glm"
```

#### 言語別 LLM 割り当て

```yaml
language_rules:
  go:
    - gemini
    - claude
  python:
    - claude
    - glm
  typescript:
    - codex
    - claude
```

### Runner バージョン管理

#### 自動アップデート確認

```bash
moai github status
```

**出力例:**
```
✓ GitHub Actions Runner
  Version: 2.700.1 (10 days old)
  Status: OK

⚠️ Update available: 2.701.0
Run: moai doctor --fix
```

#### Doctor 統合

```bash
moai doctor
```

runner のバージョンチェックがシステム診断に統合されます。

## トラブルシューティング

### PR コメントトリガーが動作しないとき

#### Checklist

1. GitHub Actions workflow が有効になっているか?
   - Repository → Actions → workflows を確認

2. GitHub Secrets が設定されているか?
   - Settings → Secrets and variables → Actions

3. Workflow permissions が正しいか?
   - `contents: read`、`pull-requests: write` が必要

### LLM 別エラー対応

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**解決:** `moai github auth claude` を再実行

#### Codex

**Error:** `repository visibility check failed`
**原因:** 公開リポジトリで Codex を使用しようとした
**解決:** リポジトリをプライベートに切り替え

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**解決:** Google Cloud Console で quota を増設

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**解決:** `~/.moai/.env.glm` のトークンを確認

## 次のステップ

- [CLI リファレンス参照](/ja/workflow-commands/)
- [Workflow 設定参照](/ja/advanced/settings-json/)
- [セキュリティポリシーの確認](/ja/advanced/security-notes/)
