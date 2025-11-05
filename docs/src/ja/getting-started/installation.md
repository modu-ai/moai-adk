---
title: インストールガイド
description: MoAI-ADKのインストールと初期設定の完全ガイド
lang: ja
---

# インストールガイド

MoAI-ADKをインストールして最初のプロジェクトを準備するまでの完全なガイドです。

## 要件

- **Python 3.13+** (必須)
- **uv** (推奨パッケージマネージャー)
- **Git** (バージョン管理用)
- **Claude Code** (AI協力開発用)

## ステップ1：uvのインストール

uvはMoAI-ADKが推奨する最新のPythonパッケージマネージャーです。

### macOS/Linux

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc  # または ~/.zshrc
```

### Windows (PowerShell)

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### 検証

```bash
uv --version
# 期待される出力: uv 0.5.1 (またはそれ以降)
```

## ステップ2：MoAI-ADKのインストール

### 基本インストール

```bash
uv tool install moai-adk
```

### 検証

```bash
moai-adk --version
# 期待される出力: moai-adk version 1.0.0
```

### インストール済みツールの確認

```bash
uv tool list | grep moai-adk
```

## ステップ3：Python環境の確認

### Pythonバージョン確認

```bash
python --version
# 期待される出力: Python 3.13.x (またはそれ以降)
```

### Python 3.13がない場合

**pyenvを使用する方法（推奨）**：

```bash
# pyenvのインストール
curl https://pyenv.run | bash

# Python 3.13のインストール
pyenv install 3.13
pyenv global 3.13

# 検証
python --version
```

**uvで自動管理する方法**：

```bash
# uvが自動的にPython 3.13をダウンロード
uv python install 3.13
uv python pin 3.13

# 検証
python --version
```

## ステップ4：Gitの設定

### Gitのインストール

**macOS**：
```bash
# Homebrewでインストール
brew install git

# またはXcode Command Line Tools
xcode-select --install
```

**Ubuntu/Debian**：
```bash
sudo apt update
sudo apt install git -y
```

**Windows**：
```powershell
winget install Git.Git
```

### 検証

```bash
git --version
# 期待される出力: git version 2.x.x
```

## ステップ5：Claude Codeのインストール

### インストール

**macOS**：
```bash
brew install claude-code
```

**他のプラットフォーム**：
```bash
npm install -g @anthropic-ai/claude-code
```

### 検証

```bash
claude --version
# 期待される出力: claude version 1.5.0 (またはそれ以降)
```

### 認証

```bash
claude auth login
```

## ステップ6：システム診断

すべての要件が満たされているか確認します：

```bash
moai-adk doctor
```

### 正常な出力例

```
Running system diagnostics...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ Check                                    ┃ Status ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                           │   ✓    │
│ uv installed                             │   ✓    │
│ Git installed                            │   ✓    │
│ Claude Code installed                   │   ✓    │
│ Project structure (.moai/)               │   ✓    │
│ Config file (.moai/config.json)          │   ✓    │
└──────────────────────────────────────────┴────────┘

✓ All checks passed
```

## ステップ7：最初のプロジェクト作成

### 新規プロジェクト

```bash
moai-adk init hello-world
cd hello-world
```

### 既存プロジェクトへの追加

```bash
cd your-existing-project
moai-adk init .
```

### 生成される構造

```
hello-world/
├── .moai/                          # MoAI-ADKプロジェクト設定
│   ├── config.json                 # プロジェクト設定（言語、モード、所有者）
│   ├── project/                    # プロジェクト情報
│   │   ├── product.md              # 製品ビジョンと目標
│   │   ├── structure.md            # ディレクトリ構造
│   │   └── tech.md                 # 技術スタックとアーキテクチャ
│   ├── memory/                     # Alfredの知識ベース（8個ファイル）
│   ├── specs/                      # SPECファイル
│   └── reports/                    # 分析レポート
├── .claude/                        # Claude Code自動化
│   ├── agents/                     # 16個サブエージェント（専門家含む）
│   ├── commands/                   # 4個Alfredコマンド
│   ├── skills/                     # 74個Claude Skills
│   ├── hooks/                      # 5個イベント自動化フック
│   └── settings.json               # Claude Code設定
└── CLAUDE.md                       # Alfredの核心指示
```

## トラブルシューティング

### よくある問題

#### 1. uvが見つからない

**症状**：
```bash
bash: uv: command not found
```

**解決策**：
```bash
# PATHに手動で追加
export PATH="$HOME/.cargo/bin:$PATH"

# シェルを再起動
source ~/.bashrc  # または ~/.zshrc
```

#### 2. Pythonバージョンが古い

**症状**：
```
Python 3.8 found, but 3.13+ required
```

**解決策**：
```bash
# pyenvでPython 3.13をインストール
pyenv install 3.13
pyenv global 3.13
```

#### 3. Claude Codeが認識されない

**症状**：
```
Command not found: claude
```

**解決策**：
```bash
# 再インストール
npm install -g @anthropic-ai/claude-code

# 認証
claude auth login
```

#### 4. プロジェクト初期化エラー

**症状**：
```
Error: .moai directory already exists
```

**解決策**：
```bash
# 既存プロジェクトの場合
cd your-project
moai-adk init .

# 新しいプロジェクトの場合
rm -rf .moai .claude
moai-adk init fresh-project
```

### 詳細診断

より詳細な診断情報が必要な場合：

```bash
moai-adk doctor --verbose
```

### ヘルプ

問題が解決しない場合：

1. **GitHub Issues**：同様の問題を検索
2. **GitHub Discussions**：質問を投稿
3. **コミュニティ**：Discordでリアルタイム質問

報告する際には以下の情報を含めてください：

- `moai-adk doctor --verbose`の出力
- エラーメッセージ全体
- オペレーティングシステムとバージョン
- 実行したコマンド

## 次のステップ

インストールが完了したら：

1. **クイックスタート**：[クイックスタートガイド](quick-start.md)で5分で最初の機能を作成
2. **概念学習**：[概念ガイド](concepts.md)で核心概念を理解
3. **Alfredコマンド**：[Alfredガイド](../guides/alfred/index.md)でコマンドを学習

---

**✅ おめでとうございます！** MoAI-ADKのインストールが完了しました。次は[クイックスタート](quick-start.md)で最初のプロジェクトを始めましょう。