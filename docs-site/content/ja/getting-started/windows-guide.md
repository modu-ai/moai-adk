---
title: Windows 利用ガイド
weight: 40
draft: false
---

Windows で MoAI-ADK を使う際に知っておくべき環境要件と、よくある落とし穴をまとめました。結論から言えば **WSL が最も快適です** — ネイティブ Windows 環境で遭遇するパス・権限の問題のほとんどは、WSL では発生しません。

## サポート環境

| 環境 | サポート状況 | 備考 |
|------|----------|------|
| **WSL (推奨)** | {{< icon check ok >}} 完全サポート | 最適な体験 |
| **PowerShell 7.x+** | {{< icon check ok >}} サポート | 代替環境 |
| PowerShell 5.x (レガシー) | {{< icon x danger >}} 非サポート | Windows PowerShell |
| cmd.exe | {{< icon x danger >}} 非サポート | コマンドプロンプト |

**必須要件:**
- [Git for Windows](https://gitforwindows.org/) のインストール必須
- WSL または PowerShell 7.x 以上

## インストール方法

### WSL (推奨)

WSL は Windows 上で Linux 環境を提供し、MoAI-ADK のすべての機能を完全にサポートします。

```bash
# WSL のインストール (管理者 PowerShell で実行)
wsl --install

# WSL 内で MoAI-ADK をインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **参考**: 最適な体験のために WSL の利用を推奨します。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## 非 ASCII ユーザー名のパスエラー

### 問題の症状

Windows のユーザー名に日本語、韓国語、中国語などの非 ASCII 文字が含まれる場合、`EINVAL` エラーが発生することがあります。これは Windows の 8.3 短縮ファイル名変換の過程で発生する問題です。

```
Error: EINVAL: invalid argument, open 'C:\Users\홍길동\AppData\Local\Temp\...'
```

### 解決方法 1: 代替一時ディレクトリの設定 (推奨)

ASCII 文字のみを含むパスに一時ディレクトリを作成します:

```bash
# Command Prompt
set MOAI_TEMP_DIR=C:\temp
mkdir C:\temp 2>/dev/null
```

```powershell
# PowerShell
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force
```

環境変数を恒久的に設定するには、システム環境変数に `MOAI_TEMP_DIR` を追加してください。

### 解決方法 2: 8.3 ファイル名生成の無効化

管理者権限で実行:

```bash
fsutil 8dot3name set 1
```

> **注意**: この設定はシステム全体に影響します。一部のレガシープログラムに影響が及ぶ可能性があります。

### 解決方法 3: ASCII ユーザーアカウントの作成

英語名で新しい Windows ユーザーアカウントを作成すれば、パス問題を根本的に解決できます。

## WSL セットアップガイド

### WSL のインストール

```powershell
# 管理者 PowerShell で実行
wsl --install

# 既定ディストリビューション: Ubuntu (推奨)
# 再起動後、ユーザー名とパスワードを設定
```

### プロジェクトファイルへのアクセス

WSL から Windows のファイルにアクセス:

```bash
# Windows ファイルシステムへのアクセス
cd /mnt/c/Users/ユーザー名/projects/

# WSL ネイティブファイルシステムを使用 (より高速)
cd ~/projects/
```

> **パフォーマンスのヒント**: WSL ネイティブファイルシステム(`~/` 配下)で作業すると、クロスファイルシステムのオーバーヘッドなしに最適なパフォーマンスが得られます。

### VS Code 連携

1. VS Code に [WSL 拡張機能](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) をインストール
2. WSL ターミナルで `code .` を実行
3. VS Code が自動的に WSL モードで開く

## CG モードでの tmux 利用

[CG モード](/ja/multi-llm/cg-mode)を使うには tmux が必要です。WSL でのインストール:

```bash
# Ubuntu/Debian
sudo apt install tmux

# tmux セッションの開始
tmux new -s moai

# CG モードの実行
moai cg
```

## トラブルシューティング

| 問題 | 原因 | 解決 |
|------|------|------|
| `moai: command not found` | PATH に Go bin ディレクトリが未登録 | `export PATH="$HOME/go/bin:$PATH"` を `.bashrc` に追加 |
| `EINVAL` エラー | 非 ASCII ユーザー名 | 上記の [非 ASCII ユーザー名のパスエラー](#非-ascii-ユーザー名のパスエラー) を参照 |
| 権限拒否 | インストールスクリプトの権限 | `chmod +x install.sh` の後に再実行 |
| Git コマンドの失敗 | Git for Windows 未インストール | [Git for Windows](https://gitforwindows.org/) をインストール |
| tmux がない | CG モードを実行できない | `sudo apt install tmux` (WSL で) |

## 次のステップ

- [インストール](/ja/getting-started/installation) — インストール詳細ガイド
- [初期設定](/ja/getting-started/init-wizard) — プロジェクトの初期化
- [CG モード](/ja/multi-llm/cg-mode) — Claude + GLM ハイブリッドモード
