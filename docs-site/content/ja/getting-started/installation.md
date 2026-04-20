---
title: インストール
weight: 30
draft: false
---
# インストール

MoAI-ADK 2.x をシステムにインストールする方法を説明します。

## 前提条件

インストール前に以下を確認してください:

### 1. Claude Code

MoAI-ADK は Claude Code 上で動作する拡張フレームワークです。先に Claude Code がインストールされている必要があります。

```bash
claude --version
```

まだインストールしていない場合は、[Claude Code 公式ドキュメント](https://docs.anthropic.com/en/docs/claude-code)を参照してください。

### 2. Git (必須)

MoAI-ADK は Git ベースのワークフローを使用します。システムに Git がインストールされている必要があります。

```bash
git --version
```

{{< callout type="warning" >}}
**Windows ユーザー**: **Git Bash** または **WSL** 環境で使用してください。コマンドプロンプト (cmd.exe) はサポートされていません。

Git がインストールされていない場合:
- **Windows**: [git-scm.com](https://git-scm.com) から Git for Windows をインストールしてください。Git Bash が同梱されています。
- **macOS**: `xcode-select --install` または [git-scm.com](https://git-scm.com)
- **Linux**: `sudo apt install git` (Ubuntu/Debian) または `sudo dnf install git` (Fedora)
{{< /callout >}}

### システム要件

| 項目 | 要件 |
|------|------|
| **OS** | macOS、Linux、Windows (Git Bash / WSL) |
| **アーキテクチャ** | amd64、arm64 |
| **メモリ** | 最小 4GB RAM |
| **ディスク** | 最小 100MB の空き容量 |

## インストール方法

### 方法 1: クイックインストール (推奨)

1 つのコマンドで最新バージョンを自動インストールします。

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
インストールスクリプトは自動的にプラットフォームを検出し、GitHub からプリビルドバイナリをダウンロードし、SHA256 チェックサムを検証し、PATH を設定します。Python や別途のランタイムは不要です。
{{< /callout >}}

インストール後に確認します:

```bash
moai version
```

#### インストールオプション

```bash
# 特定バージョンをインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version 2.0.0

# カスタムディレクトリにインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### 方法 2: ソースからビルド

Go 開発環境がある場合、ソースから直接ビルドできます。

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

ビルドされたバイナリは `./bin/moai` に生成されます。PATH が通ったディレクトリにコピーしてください:

```bash
cp ./bin/moai ~/.local/bin/
```

### インストール場所

インストールスクリプトは以下の順序でインストールディレクトリを決定します:

| プラットフォーム | 優先順位 |
|----------------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## 1.x からのマイグレーション

{{< callout type="error" >}}
**MoAI-ADK 1.x (Python 版) ユーザーは、必ず既存バージョンを先にアンインストールしてください。**

1.x と 2.x は同じ `moai` コマンドを使用するため、古いバージョンが残っていると競合が発生します。
{{< /callout >}}

### ステップ 1: 既存の 1.x を削除

```bash
# uv でインストールした場合
uv tool uninstall moai-adk

# pip でインストールした場合
pip uninstall moai-adk
```

### ステップ 2: 既存設定のバックアップ (任意)

```bash
# 既存の設定をバックアップしたい場合
cp -r ~/.moai ~/.moai-v1-backup
```

### ステップ 3: 2.x をインストール

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### ステップ 4: インストール確認

```bash
moai version
# 出力例: moai v2.x.x (commit: abc1234, built: 2026-01-15)
```

{{< callout type="info" >}}
2.x は単一の Go バイナリで、Python ランタイムや仮想環境は不要です。起動時間が約 800ms から約 5ms に大幅に改善されました。
{{< /callout >}}

## WSL サポート

Windows での WSL (Windows Subsystem for Linux) 環境における MoAI-ADK のインストールと使用方法を説明します。

### WSL のインストール

WSL がインストールされていない場合、PowerShell (管理者権限) で以下のコマンドを実行してください:

```powershell
wsl --install
```

インストール後、Windows を再起動すると Ubuntu が自動的にインストールされます。

### WSL での MoAI-ADK インストール

WSL ターミナルで Linux と同じコマンドを使用します:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### パス処理

Windows パスと WSL パスを区別する必要があります:

| Windows パス | WSL パス |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**推奨**: WSL の Linux ファイルシステム (`~/projects/`) にプロジェクトを作成すると、I/O パフォーマンスが 2-5 倍向上します。Windows ファイルシステム (`/mnt/c/`) へのアクセスはパフォーマンスが低下する可能性があります。
{{< /callout >}}

### WSL ベストプラクティス

1. **Linux ファイルシステムを使用**: プロジェクトは `~/projects/` ディレクトリに作成
2. **Git 認証情報の設定**: Windows とは別に WSL で Git 認証情報を構成
3. **推奨ターミナル**: Windows Terminal を使用して複数の WSL ディストリビューションを管理

### WSL トラブルシューティング

#### PATH が読み込まれない

```bash
# ~/.bashrc または ~/.zshrc に追加
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### フック/MCP サーバーの実行権限問題

```bash
# 実行権限を付与
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Windows パスアクセスの遅延

プロジェクトを Linux ファイルシステムに移動してください:

```bash
# Windows から WSL に移動
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip と uv ツールの競合

MoAI-ADK 1.x (Python 版) ユーザーが直面する一般的な問題です。

### 問題の説明

pip と uv は異なる場所にパッケージをインストールします。両方のツールを併用すると、`moai` コマンドが予期しないバージョンを実行する可能性があります。

### 症状

- `moai version` を実行すると 1.x バージョンが表示される
- `command not found: moai` エラーが発生
- `which moai` と異なるパスから実行される

### 原因

1. pip はシステム Python パスにインストール
2. uv tool は `~/.local/bin` または `~/.cargo/bin` にインストール
3. PATH の順序により異なるバージョンが実行される

### 解決策

#### クリーン再インストール

```bash
# 1. 既存のすべてのバージョンを削除
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. 残ったバイナリを確認・削除
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. 2.x をインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. 確認
moai version
```

#### シェル設定の更新

```bash
# ~/.bashrc または ~/.zshrc に追加
export PATH="$HOME/.local/bin:$PATH"

# 設定を適用
source ~/.bashrc  # または source ~/.zshrc
```

### 予防策

1. MoAI-ADK 2.x は Python に依存しない Go バイナリです
2. 2.x をインストールする前に 1.x (Python 版) をアンインストールしてください
3. pip と uv tool を同時に使用しないでください

## トラブルシューティング

### 問題: コマンドが見つからない

```bash
command not found: moai
```

**解決策:**

1. ターミナルを再起動してください
2. PATH を確認してください:

```bash
echo $PATH
```

3. バイナリの場所を確認してください:

```bash
which moai || ls ~/.local/bin/moai
```

4. PATH に手動で追加してください:

```bash
# Bash/Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### 問題: 権限エラー

```bash
Permission denied
```

**解決策:**

```bash
chmod +x ~/.local/bin/moai
```

### 問題: 1.x と 2.x の競合

古いバージョンの `moai` が実行される場合:

```bash
# どの moai が実行されているか確認
which moai

# 1.x が残っている場合は削除
uv tool uninstall moai-adk
# または
pip uninstall moai-adk

# ターミナル再起動後に 2.x を確認
moai version
```

## インストール後の次のステップ

インストール完了後、プロジェクトを初期化してください:

### 新しいプロジェクトを作成

```bash
moai init my-project
```

### 既存プロジェクトに適用

```bash
cd my-existing-project
moai init
```

## アップグレード

最新バージョンにアップグレードするには:

```bash
moai update
```

### アップデートオプション

```bash
# バージョン確認のみ (アップデートしない)
moai update --check

# テンプレート同期のみ (パッケージアップグレードをスキップ)
moai update --templates-only

# 設定編集モード (初期化ウィザードを再実行)
moai update --config
moai update -c

# バックアップなしで強制アップデート
moai update --force

# 自動承認モード (すべての確認を自動承認)
moai update --yes
```

### マージ戦略

```bash
# 自動マージを強制 (デフォルト)
moai update --merge

# 手動マージを強制
moai update --manual
```

{{< callout type="info" >}}
**自動保存項目**: ユーザー設定、カスタムエージェント、カスタムコマンド、カスタムスキル、カスタムフック、SPEC ドキュメント、レポートはアップデート時に自動的に保存されます。
{{< /callout >}}

詳細は[アップデートガイド](https://adk.mo.ai.kr/getting-started/update)を参照してください。

## アンインストール

MoAI-ADK を完全に削除するには:

```bash
# バイナリを削除
rm $(which moai)

# 設定ディレクトリを削除 (任意)
rm -rf ~/.moai
```

---

## 次のステップ

[初期セットアップウィザード](./init-wizard)で MoAI-ADK の設定方法を学んでください。
