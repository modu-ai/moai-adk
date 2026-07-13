---
title: インストール
weight: 30
draft: false
---

MoAI-ADK をシステムにインストールする方法を案内します。インストールされるのは Go でビルドされた単一バイナリ1つです — Python も、仮想環境も、パッケージマネージャーも必要ありません。

## ライセンス

MoAI-ADK {{< version >}} 以降は **Apache-2.0 ライセンス** の下で配布されます。

商用利用、修正、配布は自由で、ソースコードの公開義務はありません。詳細は [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) を参照してください。

{{< callout type="info" >}}
**参考**: MoAI-ADK 1.x (Python 版) は GPL-3.0 ライセンスでした。v2.0.0 から Go 言語で書き直され、Apache-2.0 に変更されました。
{{< /callout >}}

## 前提条件

インストール前に以下の項目を確認してください:

### 1. Claude Code

MoAI-ADK は Claude Code 上で動作する拡張フレームワークです。まず Claude Code がインストールされている必要があります。

```bash
claude --version
```

まだインストールしていない場合は [Claude Code 公式ドキュメント](https://docs.anthropic.com/en/docs/claude-code) を参照してください。

### 2. Git (必須)

MoAI-ADK は Git ベースのワークフローを使用します。システムに Git がインストールされている必要があります。

```bash
git --version
```

{{< callout type="warning" >}}
**Windows ユーザー**: 必ず **Git Bash** または **WSL** 環境で使用してください。Command Prompt (cmd.exe) はサポートされていません。

Git がインストールされていない場合:
- **Windows**: [git-scm.com](https://git-scm.com) から Git for Windows をインストールしてください。Git Bash も一緒にインストールされます。
- **macOS**: `xcode-select --install` または [git-scm.com](https://git-scm.com)
- **Linux**: `sudo apt install git` (Ubuntu/Debian) または `sudo dnf install git` (Fedora)
{{< /callout >}}

### システム要件

| 項目 | 要件 |
|------|---------|
| **OS** | macOS, Linux, Windows (Git Bash / WSL) |
| **アーキテクチャ** | amd64, arm64 |
| **メモリ** | 最低 4GB RAM |
| **ディスク** | 最低 100MB の空き容量 |

## インストール方法

### 方法 1: クイックインストール (推奨)

1つのコマンドで最新バージョンを自動インストールします。

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
インストールスクリプトは自動的にプラットフォームを検出し、GitHub からビルド済みバイナリをダウンロードし、SHA256 チェックサムを検証し、PATH を設定します。Python や別途のランタイムは必要ありません。
{{< /callout >}}

インストールが完了したら確認しましょう:

```bash
moai version
```

#### インストールオプション

```bash
# 特定バージョンをインストール (希望するリリースタグを指定)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version <リリースタグ>

# カスタムディレクトリにインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### 方法 2: ソースからビルド

Go 開発環境がある場合は、ソースから直接ビルドできます。

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

ビルドされたバイナリは `./bin/moai` に生成されます。PATH の通った場所にコピーしてください:

```bash
cp ./bin/moai ~/.local/bin/
```

### インストール場所

インストールスクリプトは次の順序でインストールディレクトリを決定します:

| プラットフォーム | 優先順位 |
|--------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## 1.x ユーザーのマイグレーション

{{< callout type="error" >}}
**MoAI-ADK 1.x (Python 版) のユーザーは、必ず先に既存バージョンを削除してください。**

1.x と 2.x は同じ `moai` コマンドを使用するため、既存バージョンが残っていると衝突が発生します。
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

### ステップ 3: 2.x のインストール

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### ステップ 4: インストールの確認

```bash
moai version
# 出力例: moai <バージョン> (commit: <ハッシュ>, built: <ビルド日>)
```

{{< callout type="info" >}}
Go エディション (v2.0+) は単一バイナリで、Python ランタイムや仮想環境は必要ありません。起動時間は約 800ms から 5ms へと大幅に向上しました。
{{< /callout >}}

## WSL サポート

Windows ユーザー向けに、WSL (Windows Subsystem for Linux) 環境でのインストールと使用方法を案内します。

### WSL のインストール

WSL がインストールされていない場合、PowerShell (管理者権限) で次のコマンドを実行してください:

```powershell
wsl --install
```

インストール後に Windows を再起動すると、Ubuntu が自動的にインストールされます。

### WSL での MoAI-ADK インストール

WSL ターミナルで Linux と同じコマンドを使用します:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### パスの扱い

WSL では Windows パスと WSL パスを区別する必要があります:

| Windows パス | WSL パス |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**推奨**: WSL の Linux ファイルシステム (`~/projects/`) にプロジェクトを作成すると I/O 性能が 2-5 倍向上します。Windows ファイルシステム (`/mnt/c/`) にアクセスすると性能が低下する可能性があります。
{{< /callout >}}

### WSL のベストプラクティス

1. **Linux ファイルシステムを使用**: プロジェクトは `~/projects/` ディレクトリに作成
2. **Git 認証情報の設定**: Windows とは別に WSL で Git 認証情報を構成
3. **推奨ターミナル**: Windows Terminal を使って複数の WSL ディストリビューションを管理

### WSL のトラブルシューティング

#### PATH が読み込まれない

```bash
# ~/.bashrc または ~/.zshrc に追加
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### フック/MCP サーバーの実行権限の問題

```bash
# 実行権限を付与
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Windows パスへのアクセスが遅い

Linux ファイルシステムにプロジェクトを移動してください:

```bash
# Windows から WSL へ移動
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip と uv ツールの衝突

MoAI-ADK 1.x (Python 版) のユーザーが遭遇しうる一般的な問題です。

### 問題の説明

pip と uv は異なる場所にパッケージをインストールします。両方のツールを混用すると、`moai` コマンドが予期しないバージョンを実行することがあります。

### 症状

- `moai version` を実行すると 1.x のバージョンが表示される
- `command not found: moai` エラーが発生する
- `which moai` とは別のパスから実行される

### 原因

1. pip はシステムの Python パスにインストール
2. uv tool は `~/.local/bin` または `~/.cargo/bin` にインストール
3. PATH の順序によって異なるバージョンが実行される

### 解決方法

#### 完全削除後に再インストール

```bash
# 1. すべての既存バージョンを削除
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. 残っているバイナリの確認と削除
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. 2.x のインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. 確認
moai version
```

#### シェル設定の更新

```bash
# ~/.bashrc または ~/.zshrc に追加
export PATH="$HOME/.local/bin:$PATH"

# 設定の適用
source ~/.bashrc  # または source ~/.zshrc
```

### 予防方法

1. MoAI-ADK 2.x は Python と無関係な Go バイナリです
2. 1.x (Python 版) を削除してから 2.x をインストールしてください
3. pip と uv tool を同時に使わないでください

## インストールのトラブルシューティング

### 問題: コマンドが見つからない

```bash
command not found: moai
```

**解決方法:**

1. ターミナルを再起動してください
2. PATH 設定を確認してください:

```bash
echo $PATH
```

3. バイナリのインストール場所を確認してください:

```bash
which moai || ls ~/.local/bin/moai
```

4. PATH に手動で追加してください:

```bash
# Bash/Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### 問題: 権限拒否

```bash
Permission denied
```

**解決方法:**

```bash
chmod +x ~/.local/bin/moai
```

### 問題: 1.x と 2.x の衝突

以前のバージョンの `moai` コマンドが実行される場合:

```bash
# どの moai が実行されるか確認
which moai

# 1.x が残っていれば削除
uv tool uninstall moai-adk
# または
pip uninstall moai-adk

# ターミナル再起動後に 2.x を確認
moai version
```

## インストール後の次のステップ

インストールが完了したらプロジェクトを初期化しましょう:

### 新規プロジェクトの作成

```bash
moai init my-project
```

### 既存プロジェクトへの適用

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

# テンプレート同期のみ (パッケージのアップグレードをスキップ)
moai update --templates-only

# 設定編集モード (初期化ウィザードの再実行)
moai update --config
moai update -c

# バックアップなしで強制アップデート
moai update --force

# 自動承認モード (すべての確認を自動承認)
moai update --yes
```

### マージ戦略

```bash
# 自動マージを強制 (既定値)
moai update --merge

# 手動マージを強制
moai update --manual
```

{{< callout type="info" >}}
**自動保存項目**: ユーザー設定、カスタムエージェント、カスタムコマンド、カスタムスキル、カスタムフック、SPEC 文書、レポートはアップデート時に自動的に保存されます。
{{< /callout >}}

詳細は [アップデートガイド](https://adk.mo.ai.kr/getting-started/update) を参照してください。

## アンインストール

MoAI-ADK を完全に削除するには、バイナリと設定ディレクトリを削除してください:

```bash
# バイナリの削除 (which moai の結果を削除)
rm "$(which moai)"

# 設定ディレクトリの削除 (任意)
rm -rf "$HOME/.moai"
```

---

## 次のステップ

[初期設定ウィザード](./init-wizard) で MoAI-ADK の構成方法を確認しましょう。
