---
title: Windows ガイド
weight: 40
draft: false
---

## サポート環境

| 環境 | サポート状況 | 備考 |
|------|----------|------|
| **WSL (推奨)** | ✅ 完全サポート | 最良の体験 |
| **PowerShell 7.x+** | ✅ サポート | 代替環境 |
| PowerShell 5.x (レガシー) | ❌ 未サポート | Windows PowerShell |
| cmd.exe | ❌ 未サポート | コマンドプロンプト |

**必須要件:**
- [Git for Windows](https://gitforwindows.org/) のインストールが必須
- WSLまたはPowerShell 7.x以上

## インストール方法

### WSL (推奨)

WSLはWindows上でLinux環境を提供し、MoAI-ADKのすべての機能を完全にサポートします。

```bash
# WSLのインストール(管理者権限のPowerShellで実行)
wsl --install

# WSL内でMoAI-ADKをインストール
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **参考**: 最良の体験のためにWSLの使用を推奨します。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## 非ASCIIユーザー名パスエラー

### 症状

Windowsのユーザー名に日本語、中国語などの非ASCII文字が含まれている場合、`EINVAL`エラーが発生することがあります。これはWindowsの8.3短縮ファイル名変換の過程で発生する問題です。

```
Error: EINVAL: invalid argument, open 'C:\Users\田中太郎\AppData\Local\Temp\...'
```

### 解決方法1: 代替一時ディレクトリの設定 (推奨)

ASCII文字のみを含むパスに一時ディレクトリを作成します:

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

環境変数を恒久的に設定するには、システム環境変数に`MOAI_TEMP_DIR`を追加してください。

### 解決方法2: 8.3ファイル名生成の無効化

管理者権限で実行:

```bash
fsutil 8dot3name set 1
```

> **注意**: この設定はシステム全体に影響します。一部のレガシープログラムが影響を受ける可能性があります。

### 解決方法3: ASCIIユーザーアカウントの作成

英語名で新しいWindowsユーザーアカウントを作成すると、パスの問題を根本的に解決できます。

## WSLセットアップガイド

### WSLのインストール

```powershell
# 管理者権限のPowerShellで実行
wsl --install

# デフォルトディストリビューション: Ubuntu (推奨)
# 再起動後にユーザー名とパスワードを設定
```

### プロジェクトファイルへのアクセス

WSLからWindowsファイルへアクセス:

```bash
# Windowsファイルシステムへのアクセス
cd /mnt/c/Users/ユーザー名/projects/

# WSLネイティブファイルシステムを使用 (より高速)
cd ~/projects/
```

> **パフォーマンスのヒント**: WSLネイティブファイルシステム(`~/`配下)で作業すると、クロスファイルシステムのオーバーヘッドなしに最適なパフォーマンスが得られます。

### VS Code連携

1. VS Codeに[WSL拡張機能](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl)をインストール
2. WSLターミナルで`code .`を実行
3. VS Codeが自動的にWSLモードで開く

## CGモードでのtmux使用

[CGモード](/ja/multi-llm/cg-mode)を使用するにはtmuxが必要です。WSLでインストール:

```bash
# Ubuntu/Debian
sudo apt install tmux

# tmuxセッションを開始
tmux new -s moai

# CGモードを実行
moai cg
```

## トラブルシューティング

| 問題 | 原因 | 解決 |
|------|------|------|
| `moai: command not found` | PATHにGo binディレクトリが含まれていない | `export PATH="$HOME/go/bin:$PATH"`を`.bashrc`に追加 |
| `EINVAL`エラー | 非ASCIIユーザー名 | 上記の[非ASCIIユーザー名パスエラー](#非asciiユーザー名パスエラー)を参照 |
| 権限拒否 | インストールスクリプトの権限 | `chmod +x install.sh`後に再実行 |
| Gitコマンド失敗 | Git for Windows未インストール | [Git for Windows](https://gitforwindows.org/)をインストール |
| tmuxがない | CGモードを実行できない | `sudo apt install tmux` (WSL内で) |

## 次のステップ

- [インストール](/ja/getting-started/installation) — インストールの詳細ガイド
- [初期設定](/ja/getting-started/init-wizard) — プロジェクトの初期化
- [CGモード](/ja/multi-llm/cg-mode) — Claude + GLMハイブリッドモード
