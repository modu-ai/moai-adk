---
title: "マルチLLM CIガイド"
description: "GitHub Actionsで複数のAIモデルによるコードレビューを自動化"
date: 2026-04-27
draft: false
weight: 10
---

# マルチLLM CIガイド

MoAI-ADKのマルチLLM CI機能を使用して、GitHub Actionsで複数のLLMによるコードレビューを設定する方法を説明します。

## 概要

### マルチLLM CIとは？

MoAI-ADKのマルチLLM CI機能は、GitHub Actionsで複数のAIモデルを同時にコードレビューを実行する統合CI/CDパイプラインを提供します。

### 対応LLM

| LLM | プロバイダー | トリガー方法 | 特徴 |
|-----|----------|-------------|------|
| **Claude** | Anthropic | `/claude` コメント | Issue/PRレビュー、OAuth認証 |
| **Codex** | OpenAI | PRオープン時自動 | ⚠️ プライベートリポジトリのみ |
| **Gemini** | Google | PRオープン時自動 | API Key認証 |
| **GLM** | Zhipu AI | PRオープン時自動 | トークン認証 |

### ユーザーメメリット

- **同時マルチLLMレビュー**: 1つのPRで複数のLLMのフィードバックを同時に取得
- **統一管理**: `moai github` CLIによる一貫した設定管理
- **セキュア認証**: 各LLM専用の認証処理
- **言語検出**: プロジェクト言語を自動検出し適切なLLMを割り当て

## はじめに

### 前提条件

- macOS (arm64) - v1.0基準
- Go 1.23+
- GitHubリポジトリ
- 各LLMアカウントとAPIトークン

### 初期設定

```bash
moai github init
```

このコマンドが実行する処理:
- `.github/workflows/` ディレクトリ作成
- workflowテンプレート配布
- composite actions配布
- GitHub Secrets設定ガイド

### LLM認証設定

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

### GitHub Secrets設定

各LLMに必要なSecrets:
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuthトークン
- `CODEX_AUTH_JSON` - Codex認証JSON (base64エンコード)
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM APIトークン

### 最初のPRテスト

PRを作成すると自動的にLLM Panelコメントが追加されます:

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

## LLM認証設定

### Claude設定

#### OAuthトークン発行

1. [Claude Code](https://claude.ai/download) インストール
2. ログイン後OAuthトークン発行
3. `.claude/settings.local.json`に自動保存

#### moai github auth claude

```bash
moai github auth claude
```

**対話型設定プロセス:**
```
Claude OAuthトークンが見つかりません。
Claude Codeをインストールしてログインしますか？ (y/n): y

[確認済み] OAuthトークンがsettings.local.jsonに保存されました。
GitHub Secret: CLAUDE_CODE_OAUTH_TOKENに次の値を設定してください:
<token-value>
```

### Codex設定 (プライベートリポジトリのみ)

#### 認証JSON作成

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
OpenAI auth.jsonファイルパス: ~/.codex/auth.json
ファイルを読み込みGitHub Secretを生成します...
⚠️ Codexはプライベートリポジトリでのみ使用可能です (REQ-SEC-001)

生成されたSecret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini設定

```bash
moai github auth gemini
```

API Key入力後、自動的にGitHub Secret設定ガイド提供。

### GLM設定

```bash
moai github auth glm
```

GLMトークンパス (`~/.moai/.env.glm`) から自動読み取り。

## Workflowテンプレート解説

### llm-panel.yml

**トリガー:** PRオープン時

**役割:** 各LLMの状態を視覚的に表示するパネルコメント自動作成

**備考:** `/claude`, `/codex`, `/gemini`, `/glm` コメントで個別レビュートトリガー

### claude.yml / claude-code-review.yml

- **claude.yml**: Issueトリガー (初期レビュー)
- **claude-code-review.yml**: PRトリガー (変更点レビュー)

**特徴:** `/claude` コメントのみでトリガー

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
- PR synchronize時自動トリガー

### glm-review.yml

- GLM専用環境設定 (setup-glm-env action)
- 環境変数自動注入

### Composite Actions

#### detect-language

**入力:** repositoryルートパス
**出力:** language環境変数 (`detected_language`)

**対応言語:** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift (16言語)

#### setup-glm-env

GLMチームモードで必要な環境変数設定:
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## 高度な設定

### github-actions.yamlカスタマイズ

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

#### 言語別LLM割り当て

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

### Runnerバージョン管理

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

#### Doctor統合

```bash
moai doctor
```

runnerバージョンチェックがシステム診断に統合されます (T-27)。

## トラブルシューティング

### PRコメントトリガーが動作しない場合

#### チ�荷リスト

1. ✅ GitHub Actions workflowが有効になっているか?
   - Repository → Actions → workflows で確認

2. ✅ GitHub Secretsが設定されているか?
   - Settings → Secrets and variables → Actions

3. ✅ Workflow permissionsが正しいか?
   - `contents: read`, `pull-requests: write` が必要

### LLM別エラー対応

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**解決:** `moai github auth claude` 再実行

#### Codex

**Error:** `repository visibility check failed`
**原因:** 公開リポジトリでCodex使用を試行
**解決:** リポジトリをプライベートに変更

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**解決:** Google Cloud Consoleでquota増加

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**解決:** `~/.moai/.env.glm` トークン確認

## 次のステップ

- [CLIリファレンス](/docs/commands/)
- [Workflow設定リファレンス](/docs/configuration/)
- [セキュリティポリシー確認](/docs/security/)
