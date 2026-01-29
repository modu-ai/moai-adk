# 🗿 MoAI-ADK: Agentic AI 開発フレームワーク

![MoAI-ADK](./assets/images/readme/moai-adk-og.png)

**利用可能な言語:** [🇰🇷 한국어](./README.ko.md) | [🇺🇸 English](./README.md) | [🇯🇵 日本語](./README.ja.md) | [🇨🇳 中文](./README.zh.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: Copyleft](https://img.shields.io/badge/License-Copyleft--3.0-blue.svg)](./LICENSE)
[![Python](https://img.shields.io/badge/Python-3.11--3.14-blue)](https://www.python.org/)

> **"バイブコーディングの目的は迅速な生産性ではなく、コード品質である"**

MoAI-ADKは**高品質なコードを作成するAI開発環境**を提供します。SPEC-First DDD (Domain-Driven Development)、動作保存による継続的リファクタリング、そして20個の専門AIエージェントが一緒に働きます。

---

## 🎁 MoAI-ADK スポンサー: z.ai GLM 4.7

**💎 コスト効率の良いAI開発のための最適ソリューション**

MoAI-ADKは**z.ai GLM 4.7**とのパートナーシップを通じて、開発者に経済的なAI開発環境を提供します。

### 🚀 GLM 4.7 特別特典

| 特典                  | 説明                                       |
| --------------------- | ------------------------------------------ |
| **💰 70% コスト削減** | Claude比 1/7価格で同等性能                 |
| **⚡ 高速応答**       | 最適化されたインフラで低レイテンシ応答提供 |
| **🔄 互換性**         | Claude Codeと完全互換、別途コード修正不要  |
| **📈 無制限使用**     | 日次/週次トークンリミットなしで自由に使用  |

### 🎁 登録特別割引

**👉 [GLM 4.7 登録 (10% 追加割引)](https://z.ai/subscribe?ic=1NDV03BGWU)**

このリンクから登録すると:

- ✅ **追加10%割引**特典
- ✅ **MoAIオープンソース開発**に貢献 (リワードクレジットはオープンソースプロジェクトに使用されます)

### 💡 使用ガイド

```bash
# 1. GLM APIキー発行
上記リンクから登録後APIキー発行

# 2. MoAI-ADKでGLM設定
moai glm YOUR_API_KEY
```

> **💡 ヒント**: Worktree環境でGLM 4.7を活用すると、Opusで設計してGLMで大量実装してコストを最大70%削減できます。

---

## 🌟 核心価値

- **🎯 SPEC-First**: 明確な仕様書で90%再作業削減
- **🔵 DDD (Domain-Driven Development)**: 分析-保存-改善サイクルによる動作保存リファクタリング
- **🤖 AIオーケストレーション**: 20個専門エージェント + 52個スキル
- **🧠 Sequential Thinking MCP**: 段階的推論による構造化された問題解決
- **🌐 多言語ルーティング**: 韓国語/英語/日本語/中国語自動サポート
- **🌳 Worktree並列開発**: 完全分離環境で無制限並列作業
- **🏆 MoAI Rank**: バイブコーディングリーダーボードでモチベーション
- **🔗 Ralph-Style LSP統合 (NEW v1.9.0)**: リアルタイム品質フィードバックのためのLSPベース自律ワークフロー

---

## 🎯 Ralph-Style LSP統合 (NEW v1.9.0)

### LSP統合概要

MoAI-ADKは、LSP（Language Server Protocol）診断統合によるRalphスタイル自律ワークフローをサポートするようになりました。システムはワークフロー開始時にLSP診断状態をキャプチャし、実行中に診断状態をモニタリングし、品質しきい値が満たされると自動的にフェーズを完了します。

### 主な機能

**LSPベースラインキャプチャ**:
- フェーズ開始時の自動LSP診断キャプチャ
- エラー、警告、タイプエラー、リントエラーの追跡
- 回帰検出のためのベースライン使用

**完了マーカー**:
- Planフェーズ: SPEC作成完了、ベースライン記録
- Runフェーズ: 0エラー、0タイプエラー、カバレッジ >= 85%
- Syncフェーズ: 0エラー、<10警告

**実行モード**:
- Interactive（デフォルト）: 各ステップで手動承認
- Autonomous（選択可能）: 完了まで連続ループ

**ループ防止**:
- 最大100回反復
- 進捗なし検出（5回反復）
- 停滞時の代替戦略

**構成**:
```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  autonomous:
    user_approval_required: false
    continuous_loop: true
    completion_marker_based: true
    lsp_feedback_integration: true
```

---

> **📚 詳細は公式オンラインドキュメントを参照してください:** [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

## 1. 30秒インストール

> **⚠️ Windowsユーザー**: MoAI-ADKは**PowerShell**と**WSL (Windows Subsystem for Linux)**をサポートしています。コマンドプロンプト(cmd.exe)は**サポートされていません**。PowerShell、Windows Terminal、またはWSLを使用してください。

### 📋 前提条件

**必要な依存関係:**

- **Python 3.11+**: MoAI-ADKにはPython 3.11以降が必要です
- **PyYAML 6.0+**: Hookスクリプト実行に必要です（`uv run --with pyyaml`で自動インストール）
- **uv**: Pythonパッケージャー（MoAI-ADKインストールに含まれています）

**参考**: PyYAMLは以下の機能に必要です:
- AST-grep マルチドキュメントYAMLパース
- 設定ファイルの読み書き
- SPECファイルYAMLフロントマatterパース
- スキルメタデータ処理

PyYAMLがない場合、Hookは自動的に`uv run --with pyyaml`を使用してインストールを試みます。

### 🚀 方法1: クイックインストール (推奨)

```bash
curl -LsSf https://modu-ai.github.io/moai-adk/install.sh | sh
```

### 🔧 方法2: 手動インストール

```bash
# Step 1: uv インストール (macOS/Linux)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Step 2: MoAI-ADK インストール
uv tool install moai-adk
```

### 🎨 対話型設定ウィザード

`moai init`コマンド実行時に**9段階対話型ウィザード**が開始されます:

![MoAI-ADK Init Wizard](./assets/images/readme/init-wizard-banner.png)

---

#### Step 1: 言語選択

対話言語を選択します。以降の案内がすべて選択した言語で表示されます。

```text
🌐 Language Selection
❯ Select your conversation language: [↑↓] Navigate  [Enter] Select
❯ Japanese (日本語)
  English
  Korean (한국어)
  Chinese (中文)
```

---

#### Step 2: 名前入力

ユーザー名を入力します。AIがパーソナライズされた応答を提供します。

```text
👤 ユーザー設定
❯ ユーザー名を入力してください (選択事項):
```

---

#### Step 3: GLM APIキー入力

Z.AI社のGLM APIキーを入力します。

```text
🔑 APIキー入力
GLM CodePlan API key (optional - press Enter to skip)

✓ 既存GLM APIキーが見つかりました: 99c1a2df...
Enterを押すと既存キー維持、新しいキーを入力すると交換されます

? GLM APIキーを入力してください:
```

> 🎁 **GLM登録特典**: GLMアカウントがない場合は下記リンクで登録してください!
>
> **👉 [GLM 登録 (10% 追加割引)](https://z.ai/subscribe?ic=1NDV03BGWU)**
>
> このリンクから登録すると**追加10%割引**特典を受けられます。
> また、リンクを通じた登録時に発生するリワードは**MoAIオープンソース開発**に使用されます。🙏

---

#### Step 4: プロジェクト設定

プロジェクト名を入力します。

```text
📁 プロジェクト設定
❯ プロジェクト名: MoAI-ADK
```

---

#### Step 5: Git設定

Gitモードを選択します。

```text
🔀 Git設定
❯ Gitモードを選択してください: [↑↓] Navigate  [Enter] Select
❯ manual (ローカルのみ) - ローカルリポジトリのみ使用
  personal (GitHub個人) - GitHub個人アカウント使用
  team (GitHubチーム) - GitHubチーム/組織使用
```

---

#### Step 6: GitHubユーザー名

personal/team選択時にGitHubユーザー名を入力します。

```text
❯ GitHubユーザー名:
```

---

#### Step 7: コミットメッセージ言語

Gitコミットメッセージに使用する言語を選択します。

```text
🗣️ 出力言語設定
❯ コミットメッセージ言語: [↑↓] Navigate  [Enter] Select
  English
❯ Japanese (日本語)
  Korean (한국어)
  Chinese (中文)
```

---

#### Step 8: コードコメント言語

コードコメントに使用する言語を選択します。

```text
❯ コードコメント言語: [↑↓] Navigate  [Enter] Select
  English
❯ Japanese (日本語)
  Korean (한국어)
  Chinese (中文)
```

---

#### Step 9: ドキュメント言語

ドキュメントに使用する言語を選択します。

```text
❯ ドキュメント言語: [↑↓] Navigate  [Enter] Select
  English
❯ Japanese (日本語)
  Korean (한국어)
  Chinese (中文)
```

> 💡 **トークン最適化戦略**: エージェントに指示する内部プロンプトは**英語で固定**されています。
>
> **理由**: 非英語圏言語はClaudeで**12%~20%トークンを追加消費**します。無限反復エージェント作業が多くなるとコストと週次トークンリミットに大きな影響を与えるため、MoAIは内部エージェント指示は英語で固定し**一般対話のみユーザー言語で提供**します。
>
> これがMoAIの**トークン浪費を減らすための取り組み**です。

---

#### インストール完了

すべての設定が完了すると5段階インストールが自動進行します:

```text
🚀 Starting installation...

Phase 1: Preparation and backup...        ████████████████ 100%
Phase 2: Creating directory structure...  ████████████████ 100%
Phase 3: Installing resources...          ████████████████ 100%
Phase 4: Generating configurations...     ████████████████ 100%
Phase 5: Validation and finalization...   ████████████████ 100%

✅ Initialization Completed Successfully!
────────────────────────────────────────────────────────────────

📊 Summary:
  📁 Location:   /path/to/my-project
  🌐 Language:   Auto-detect (use /moai project)
  🔀 Git:        manual (github-flow, branch: manual)
  🌍 Locale:     ja
  📄 Files:      47 created
  ⏱️  Duration:   1234ms

🚀 Next Steps:
  1. Run cd my-project to enter the project
  2. Run /moai project in Claude Code for full setup
  3. Start developing with MoAI-ADK!
```

### 既存プロジェクトに追加

```bash
cd your-existing-project
moai init .
# 既存ファイルはそのまま維持されます
```

### WSL (Windows Subsystem for Linux) サポート

MoAI-ADKはWindows 10およびWindows 11上で**WSL 1**と**WSL 2**を完全にサポートしています。

#### WSLへのインストール

```bash
# WSLでMoAI-ADKをインストール
uv tool install moai-adk

# プロジェクトの初期化
cd your-project-directory
moai-adk init
```

#### パス処理

MoAI-ADKはWindowsとWSLのパス形式を自動的に変換します:

- **Windowsパス**: `C:\Users\goos\project` → **WSLパス**: `/mnt/c/Users/goos/project`
- 手動設定は不要です
- Linuxファイルシステム(`/home/user/`)とWindowsファイルシステム(`/mnt/c/`)の両方でシームレスに動作します

#### ベストプラクティス

**推奨**: 最適なパフォーマンスのためにLinuxファイルシステムにプロジェクトを配置してください
```bash
# ✅ 最高のパフォーマンス
cd ~/projects
moai-adk init
```

**サポート済み**: Windowsファイルシステム上のプロジェクトも使用可能です
```bash
# ✅ 動作しますが、わずかなオーバーヘッドが発生する可能性があります
cd /mnt/c/Users/YourName/projects
moai-adk init
```

#### WSLのトラブルシューティング

**WSL環境の確認:**
```bash
# WSLで実行中か確認
echo $WSL_DISTRO_NAME

# CLAUDE_PROJECT_DIRの確認 (Claude Codeが設定)
echo $CLAUDE_PROJECT_DIR
```

**パスの問題:**
- フックが失敗する場合は、`CLAUDE_PROJECT_DIR`が正しく設定されているか確認してください
- MoAI-ADKはWindowsパスをWSL形式に自動変換します
- `.claude/settings.json`で正しいパス参照を確認してください

**関連Issue:**
- [Issue #295: WSL Support Request](https://github.com/modu-ai/moai-adk/issues/295)
- [Claude Code Issue #19653: WSL Path Handling](https://github.com/anthropics/claude-code/issues/19653)

---

### 🔄 MoAI-ADK アップデート

既存プロジェクトを最新バージョンにアップデートします。

```bash
moai update
```

**3段階スマートアップデートワークフロー**:

```text
Stage 1: 📦 パッケージバージョン確認
         └─ PyPIで最新バージョン確認 → 必要時自動アップグレード

Stage 2: 🔍 Configバージョン比較
         └─ パッケージテンプレート vs プロジェクト設定比較
         └─ 同一ならスキップ (70-80% 性能向上)

Stage 3: 📄 テンプレート同期
         └─ バックアップ作成 → テンプレートアップデート → ユーザー設定復元
```

**主要オプション**:

```bash
# バージョンのみ確認 (アップデートなし)
moai update --check

# テンプレートのみ同期 (パッケージアップグレードスキップ)
moai update --templates-only

# 設定編集モード (initウィザード再実行)
moai update --config
moai update -c

# バックアップなし強制アップデート
moai update --force

# All is well~ 自動モード (すべての確認自動承認)
moai update --yes
```

**マージ戦略選択**:

```text
🔀 Choose merge strategy:
  [1] Auto-merge (default)
      → テンプレート + ユーザー変更事項自動保存
  [2] Manual merge
      → バックアップ + マージガイド作成 (直接制御)
```

```bash
# Auto-merge強制 (デフォルト)
moai update --merge

# Manual merge強制
moai update --manual
```

**自動保存項目**:

| 項目                     | 説明                                         |
| ------------------------ | -------------------------------------------- |
| **ユーザー設定**         | `.claude/settings.local.json` (MCP, GLM設定) |
| **カスタムエージェント** | テンプレートにないユーザー生成エージェント   |
| **カスタムコマンド**     | ユーザー定義スラッシュコマンド               |
| **カスタムスキル**       | ユーザー定義スキル                           |
| **カスタムフック**       | ユーザー定義フックスクリプト                 |
| **SPECドキュメント**     | `.moai/specs/` フォルダ全体                  |
| **レポート**             | `.moai/reports/` フォルダ全体                |

> 💡 **アップデートヒント**: `moai update -c`でいつでも言語、APIキー、Git設定を変更できます。
> ユーザーのコマンド、エージェント、スキル、フックはmoai以外のフォルダに生成して使用すると良いです。

---

## ⚠️ 既知の問題と解決策

### pipとuv toolの競合

**問題点**: pipとuv toolの両方でMoAI-ADKをインストールした場合、バージョンの競合が発生する可能性があります。

**症状**:
```bash
# moai updateは最新バージョンを表示
moai update
✓ Package already up to date (1.5.0)

# しかし実際のコマンドは古いバージョンを使用
which moai
~/.pyenv/shims/moai  # pipバージョンを使用 (例: 1.1.0)

# hookでimportエラー
ModuleNotFoundError: No module named 'yaml'
```

**根本原因**:
- `uv tool install`は `~/.local/bin/moai` にインストール
- `pip install`は `~/.pyenv/shims/moai` にインストール
- PATHの優先順位によって使用されるバージョンが決まる
- **WindowsユーザーはPython環境の違いによりより深刻な問題が発生する可能性**

**解決策**:

#### オプション1: uv toolのみ使用 (推奨)

```bash
# pipバージョンをアンインストール
pip uninstall moai-adk -y

# uv toolがPATHで優先されるように設定
export PATH="$HOME/.local/bin:$PATH"

# 確認
which moai  # ~/.local/bin/moaiが表示されるはず
moai --version  # 最新バージョンが表示されるはず
```

#### オプション2: シェル設定を更新

**macOS/Linux (~/.zshrcまたは~/.bashrc)**:
```bash
# pyenv初期化の後に追加
# ===== UV Tool Priority =====
export PATH="$HOME/.local/bin:$PATH"
```

**Windows (PowerShell $PROFILE)**:
```powershell
# $PROFILEに追加
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"
```

#### オプション3: uvで強制再インストール

```bash
# すべてのバージョンをアンインストール
pip uninstall moai-adk -y
uv tool uninstall moai-adk

# uvで再インストール
uv tool install moai-adk

# 確認
uv tool list
moai --version
```

**予防**:
- 常に `uv tool install moai-adk` でインストール
- pipとuvを混用しない
- 定期的に `which moai` でアクティブなインストールを確認

---

## 2. プロジェクトドキュメント生成 (選択事項)

新規プロジェクトや既存プロジェクトで**Claude Codeがプロジェクトを理解するのを助ける**プロジェクトドキュメントを自動生成できます:

```text
> /moai project
```

### 生成される3つのファイル

| ファイル                     | 目標             | 主要内容                                                               |
| ---------------------------- | ---------------- | ---------------------------------------------------------------------- |
| `.moai/project/product.md`   | **製品概要**     | プロジェクト名/説明、ターゲットユーザー、核心機能、使用事例            |
| `.moai/project/structure.md` | **構造分析**     | ディレクトリツリー、主要フォルダ目的、核心ファイル位置、モジュール構成 |
| `.moai/project/tech.md`      | **技術スタック** | 使用技術、フレームワーク選択理由、開発環境、ビルド/デプロイ設定        |

### なぜ必要なのか？

- **コンテキスト提供**: Claude Codeがプロジェクト文脈を迅速把握
- **一貫性維持**: チームメンバー間のプロジェクト理解度共有
- **オンボーディング加速**: 新規開発者のプロジェクト把握時間短縮
- **AI協業最適化**: より正確なコード提案とレビュー可能

> 💡 **ヒント**: プロジェクト初期または構造変更時に`/moai project`を実行すると最新状態でドキュメントが更新されます。

---

## 3. 核心コマンド集

### 🎯 `/moai project` - プロジェクト初期化

```bash
> /moai project
```

プロジェクトの現在状態を自動分析して最適の開発環境を構成します。プログラミング言語とフレームワークを検出し、Gitワークフローと品質保証基準を自動設定します。すべての構成が完了するとすぐに開発を開始できる準備状態になります。

**作業内容**:

- ✅ プロジェクト構造分析
- ✅ プログラミング言語/フレームワーク検出
- ✅ `.moai/config/config.yaml` 生成
- ✅ Gitワークフロー設定
- ✅ セッションメモリシステム構成
- ✅ 品質保証基準設定

---

### 📋 `/moai plan` - SPEC作成

```bash
> /moai plan "機能説明"
```

EARS形式を使用して曖昧さのない仕様書を自動生成します。要求事項定義、成功基準、テストシナリオを含めて開発方向を明確に提示します。生成されたSPECは開発チームとAIが同一理解を共有する単一ソース（Source of Truth）として作動します。

**自動生成**:

- EARS形式仕様書
- 要求事項定義
- 成功基準
- テストシナリオ

**例**:

```bash
> /moai plan "ユーザープロフィールページ"
# → SPEC-002 生成

> /moai plan "決済API"
# → SPEC-003 生成
```

**重要**: 必ず次に`> /clear`実行

---

### 💻 `/moai run` - 実装 (DDD)

```bash
> /moai run SPEC-001
```

DDD (Domain-Driven Development) 方法論で実装を実行します：

**DDDサイクル** (ANALYZE-PRESERVE-IMPROVE):

- 🔍 **ANALYZE**: ドメイン境界と結合度分析
- 🛡️ **PRESERVE**: 特性テストで動作保存
- ✨ **IMPROVE**: 漸進的構造改善

**検証項目**:

- テストカバレッジ >= 85%
- リンティング通過
- タイプ検査通過
- セキュリティ検査通過
- ✅ TRUST 5検証

---

### 📚 `/moai sync` - ドキュメント同期

```bash
> /clear  # 同期前に常にclearを実行してセッション初期化後品質検査を実行
> /moai sync SPEC-001
```

品質検証を開始にドキュメント同期、Gitコミット、PR自動化を行います。APIドキュメント、アーキテクチャ図、README、CHANGELOGを自動生成して最新状態を維持します。変更事項を自動コミットしてチームモードではPRをDraftからReadyに転換します。

**自動実行される作業**:

1. **Phase 1: 品質検証**
   - テスト実行 (pytest, jest, go test等)
   - リンター検査 (ruff, eslint, golangci-lint等)
   - タイプチェッカー (mypy, tsc, go vet等)
   - コードレビュー (manager-quality)

2. **Phase 2-3: ドキュメント同期**
   - APIドキュメント自動生成
   - アーキテクチャ図更新
   - README更新
   - SPECドキュメント同期

3. **Phase 4: Git自動化**
   - 変更事項コミット
   - PR Draft → Ready転換
   - (選択) Auto-merge

**実行モード**:

- `auto` (デフォルト): 変更されたファイルのみ選択同期
- `force`: 全体ドキュメント再生成
- `status`: 状態確認のみ実行
- `project`: プロジェクト全体同期

**詳細**: コマンドファイル参照

---

### 🚀 `/moai` - 完全自律自動化

```bash
> /moai "機能説明"
```

ユーザーが目標を提示するとAIが自ら探索、計画、実装、検証をすべて行います。並列探索でコードベースを分析し、自律ループを通じて問題を自ら修正します。完了マーカー(`<moai>DONE</moai>`)を検知すると自動終了して開発者は最終結果のみ確認すれば良いです。

#### コンセプトとワークフロー

```mermaid
flowchart TB
    Start([ユーザーリクエスト<br/>/moai '機能説明']) --> Phase0[Phase 0: 並列探索]

    Phase0 --> Explore[🔍 Explore Agent<br/>コードベース構造分析]
    Phase0 --> Research[📚 Research Agent<br/>技術文書調査]
    Phase0 --> Quality[✅ Quality Agent<br/>品質状態評価]

    Explore --> Phase1[Phase 1: SPEC作成]
    Research --> Phase1
    Quality --> Phase1

    Phase1 --> Spec[📋 EARS形式SPEC文書<br/>要件仕様書]

    Spec --> Phase2[Phase 2: DDD実装]

    Phase2 --> Analyze[🔍 ANALYZE: ドメイン分析]
    Analyze --> Preserve[🛡️ PRESERVE: 動作保存]
    Preserve --> Improve[✨ IMPROVE: 漸進的改善]

    Improve --> Check{品質検証<br/>TRUST 5}

    Check -->|通過| Phase3[Phase 3: ドキュメント同期]
    Check -->|失敗| Loop[🔄 自律ループ<br/>問題自動修正]

    Loop --> Analyze

    Phase3 --> Docs[📚 README、APIドキュメント<br/>自動更新]

    Docs --> Done{<moai>DONE</moai><br/>完了マーカー検知}

    Done -->|はい| End([✅ 完了<br/>結果のみ伝達])
    Done -->|いいえ| Loop

    style Phase0 fill:#e1f5fe
    style Phase1 fill:#fff3e0
    style Phase2 fill:#f3e5f5
    style Phase3 fill:#e8f5e9
    style Done fill:#c8e6c9
    style End fill:#4caf50,color:#fff
```

#### 詳細プロセス

**一度に実行**:

1. **Phase 1: 並列探索** (3-4倍高速分析)
   - **Explore Agent**: コードベース構造、パターン、関連ファイル分析
   - **Research Agent**: 技術文書、ベストプラクティス調査
   - **Quality Agent**: 現在の品質状態、潜在的問題特定

2. **Phase 2: SPEC作成** (EARS形式)
   - 明確な要件定義
   - 受諾基準仕様書
   - ユーザーストーリー作成

3. **Phase 3: DDD実装** (自律ループ)
   - **ANALYZE**: ドメイン境界と結合度分析
   - **PRESERVE**: 特性テストで動作保存
   - **IMPROVE**: 漸進的コード品質改善
   - **ループ**: 品質検証失敗時自動問題修正

4. **Phase 4: ドキュメント同期**
   - README、APIドキュメント自動更新
   - CHANGELOG自動生成
   - ユーザーガイド最新化

#### いつ使用するか？

| 状況                       | 説明                         | 例                           |
| -------------------------- | ---------------------------- | ---------------------------- |
| **新機能開発**             | 最初から最後までAIが自動処理 | "JWT認証システム追加"        |
| **複雑なリファクタリング** | 複数ファイルに影響大きな変更 | "データベース層再構成"       |
| **バグ修正**               | 原因特定から修正まで自動化   | "ログイン失敗バグ修正"       |
| **SPECベース開発**         | SPEC文書のある機能実装       | `/moai SPEC-AUTH-001` |

**オプション**:

- `--loop`: 自律反復修正活性化 (AIが自ら問題解決)
- `--max N`: 最大反復回数指定 (デフォルト: 100)
- `--sequential` / `--seq`: 順次探索 (デバッグ用) - 並列がデフォルト
- `--branch`: 機能ブランチ自動作成
- `--pr`: 完了後Pull Request作成
- `--resume SPEC`: 続きから

> **パフォーマンス**: 並列探索がデフォルトになり、3-4倍高速な分析が可能です。`--sequential`はデバッグ用にのみ使用してください。

**例**:

```bash
# 基本自律実行 (並列がデフォルト)
> /moai "JWT認証追加"

# 自動ループ + 順次探索 (デバッグ用)
> /moai "JWT認証" --loop --seq

# 続きから
> /moai resume SPEC-AUTH-001

# UltraThinkモード (Sequential Thinkingで深層分析)
> /moai "JWT認証追加" --ultrathink
```

**UltraThinkモード** (`--ultrathink`): Sequential Thinking MCPを自動的に適用してリクエストを深層分析し、最適な実行プランを生成する強化された分析モードです。

`--ultrathink`が追加されると、MoAIが構造化された推論を有効にして：
- 複雑な問題を管理可能なステップに分解
- 各サブタスクを適切なエージェントにマッピング
- 並列 vs 順次実行の機会を特定
- 最適なエージェント委任戦略を生成

**UltraThink出力例**:
```
thought: "リクエスト分析: ユーザーがJWT認証を希望。次が含まれます: ユーザーモデル(backend)、APIエンドポイント(backend)、ログインフォーム(frontend)、認証コンテキスト(frontend)、テスト(testing)。"

thought: "サブタスク分解: 1) ユーザーモデル → expert-backend, 2) JWT API → expert-backend, 3) ログインフォーム → expert-frontend, 4) 認証コンテキスト → expert-frontend, 5) テスト → expert-testing。"

thought: "実行戦略: Phase 1 - ユーザーモデル + APIのためにexpert-backendを並列実行。Phase 2 - UIのためにexpert-frontend実行。Phase 3 - テストのためにexpert-testing実行。"

thought: "最終プラン: Use the expert-backend subagent (parallel), then Use the expert-frontend subagent, then Use the expert-testing subagent."
```

次の場合に `--ultrathink` を使用してください：
- 複雑なマルチドメインタスク (backend + frontend + testing)
- 複数のファイルに影響するアーキテクチャ決定
- 分析が必要なパフォーマンス最適化
- セキュリティレビューが必要
- 振る舞い保存リファクタリング

---

### 🔁 `/moai loop` - 自律反復修正

```bash
> /moai loop
```

AIが自らLSPエラー、テスト失敗、カバレッジ不足を診断して修正を反復します。並列診断でLSP、AST-grep、Tests、Coverageを同時実行して3-4倍速く問題を解決します。完了マーカーを検知または最大反復回数に到達するまで自律的に実行されます。

#### コンセプトとワークフロー

```mermaid
flowchart TB
    Start([ユーザーリクエスト<br/>/moai loop]) --> Parallel[並列診断]

    Parallel --> LSP[LSP診断<br/>タイプエラー、未定義]
    Parallel --> AST[AST-grep<br/>パターン検査、セキュリティ]
    Parallel --> Tests[テスト実行<br/>単体、統合]
    Parallel --> Coverage[カバレッジ<br/>85%目標]

    LSP --> Collect[問題収集]
    AST --> Collect
    Tests --> Collect
    Coverage --> Collect

    Collect --> HasIssues{問題あり?}

    HasIssues -->|いいえ| Done[<moai>DONE</moai><br/>完了マーカー]
    HasIssues -->|はい| CreateTODO[TODO生成<br/>優先順位別ソート]

    CreateTODO --> Process[順次処理]

    Process --> Fix1[Level 1: 即時修正<br/>import整列、空白]
    Process --> Fix2[Level 2: 安全修正<br/>変数名、タイプ]
    Process --> Fix3[Level 3: 承認修正<br/>ロジック変更]

    Fix1 --> Validate[検証]
    Fix2 --> Validate
    Fix3 --> Validate

    Validate --> ReCheck{再診断?}

    ReCheck -->|はい| Parallel
    ReCheck -->|いいえ| MaxIter{最大反復<br/>100回到達?}

    MaxIter -->|いいえ| Parallel
    MaxIter -->|はい| Snapshot[スナップショット保存<br/>後で再開可能]

    Done --> End([✅ 完了])
    Snapshot --> End([⏸️ 一時停止])

    style Parallel fill:#e1f5fe
    style Collect fill:#fff3e0
    style Process fill:#f3e5f5
    style Validate fill:#e8f5e9
    style Done fill:#c8e6c9
    style End fill:#4caf50,color:#fff
```

#### 並列診断詳細

**並列診断** (3.75倍高速):

```mermaid
flowchart TB
    Start([並列診断開始]) --> Parallel

    subgraph Parallel[同時実行]
        direction TB
        LSP[LSP診断]
        AST[AST-grep検査]
        TESTS[テスト実行]
        COVERAGE[カバレッジ確認]
    end

    LSP --> Collect[問題統合と優先順位<br/>Level 1 → 2 → 3順序処理]
    AST --> Collect
    TESTS --> Collect
    COVERAGE --> Collect

    style Start fill:#e3f2fd
    style Parallel fill:#f3f4f6
    style LSP fill:#fff9c4
    style AST fill:#ffccbc
    style TESTS fill:#c8e6c9
    style COVERAGE fill:#b2dfdb
    style Collect fill:#e1bee7
```

#### 📖 AST-grepとは？

> **"grepはテキストを探すが、AST-grepはコード構造を探す"**

**概念**:

AST-grepは**構造的コード検査ツール**です。通常のgrepや正規表現がテキストを検索するのとは異なり、AST-grepはコードの**抽象構文木(Abstract Syntax Tree)**を分析してコードの**構造とパターン**を検査します。

**テキスト検索 vs 構造検索**:

| 特徴     | grep/正規表現               | AST-grep                      |
| -------- | --------------------------- | ----------------------------- |
| 検索対象 | テキスト文字列              | コード構造 (AST)              |
| 例       | `print("hello")`            | `print(__)`                   |
| 意味     | "print"という文字列検索     | print関数呼び出しパターン検索 |
| 空白感応 | はい (空白、インデント重要) | いいえ (構造のみ分析)         |
| 変数区別 | 困難 (例: `x=1`, `y=1`は別) | 可能 (全変数代入パターン)     |

**動作原理**:

```mermaid
flowchart LR
    Source[ソースコード<br/>def foo x:<br/>    return x + 1] --> AST[AST分析]

    AST --> |変換| Tree[抽象構文木<br/>Function<br/> Call<br/>]

    Tree --> Pattern[パターンマッチング]

    Pattern --> Result1[✓ 関数定義]
    Pattern --> Result2[✓ return文]
    Pattern --> Result3[✓ 加算演算]

    style Source fill:#e3f2fd
    style AST fill:#fff3e0
    style Tree fill:#f3e5f5
    style Pattern fill:#e8f5e9
    style Result1 fill:#c8e6c9
    style Result2 fill:#c8e6c9
    style Result3 fill:#c8e6c9
```

**AST-grepが検知すること**:

1. **セキュリティ脆弱性**
   - SQLインジェクションパターン: `execute(f"SELECT * FROM users WHERE id={user_input}")`
   - ハードコードされたパスワード: `password = "123456"`
   - 安全でない関数使用: `eval(user_input)`

2. **コードスメル (Code Smells)**
   - 重複コード: 類似構造の反復
   - 長い関数: 複雑すぎる
   - マジックナンバー: `if x == 42` (意味のない数字)

3. **アンチパターン (Anti-patterns)**
   - 空のexceptブロック: `except: pass`
   - 大域変数修正
   - 循環依存

4. **ベストプラクティス違反**
   - タイプヒント欠如
   - ドキュメント化欠如
   - エラー処理欠如

**例シナリオ**:

```python
# AST-grepが問題を見つけるコード例
def process_user_input(user_input):
    # ⚠️ 警告: eval使用 (セキュリティ脆弱性)
    result = eval(user_input)

    # ⚠️ 警告: 空のexcept (アンチパターン)
    try:
        save_to_database(result)
    except:
        pass

    # ⚠️ 警告: マジックナンバー (コードスメル)
    if result > 42:
        return True
```

**なぜ重要か？**

- **正確性**: コードの意味を理解して検査するので誤検知(False Positive)が少ない
- **40言語サポート**: Python、TypeScript、Go、Rust、Javaなど多様な言語で動作
- **自動修正可能**: パターンを見つけるだけでなく自動修正提案も生成
- **セキュリティ強化**: OWASP Top 10などセキュリティ脆弱性を自動検知

**MoAI-ADKでの活用**:

`/moai loop`と`/moai fix`コマンドでAST-grepは並列診断の中核構成要素として動作します:

- **LSP**: タイプエラー、定義検索
- **AST-grep**: 構造的パターン、セキュリティ脆弱性 ← **これが私たちの関心事!**
- **Tests**: テスト失敗
- **Coverage**: カバレッジ不足

この4つが**同時実行**されてコード品質を3.75倍速く診断します。

---

#### 詳細プロセス

**自律ループフロー**:

1. **並列診断** (同時実行)
   - **LSP**: タイプエラー、未定義、潜在的バグ
   - **AST-grep**: コードパターン検査、セキュリティ脆弱性
   - **Tests**: 単体テスト、統合テスト実行
   - **Coverage**: 85%カバレッジ目標達成確認

2. **TODO生成** (優先順位別)
   - Level 1: 即時修正 (import整列、空白、フォーマット)
   - Level 2: 安全修正 (変数名、タイプ追加)
   - Level 3: 承認修正 (ロジック変更、API修正)
   - Level 4: 手動必要 (セキュリティ、アーキテクチャ)

3. **順次修正**
   - TODO項目を一つずつ処理
   - 各修正後検証
   - 失敗時再診断

4. **反復または完了**
   - 全問題解決時 `<moai>DONE</moai>` マーカー
   - 最大100回反復後スナップショット保存

#### いつ使用するか？

| 状況                 | 説明                     | 例                           |
| -------------------- | ------------------------ | ---------------------------- |
| **実装後品質確保**   | コード作成後自動品質改善 | 機能実装後 `/moai loop` 実行 |
| **テスト失敗修正**   | テスト失敗を自動分析修正 | テスト実行後失敗時           |
| **カバレッジ向上**   | 85%目標を自動達成        | 新コード作成後               |
| **リファクタリング** | コード品質を継続的に改善 | 定期的実行で保守             |

**オプション**:

- `--max N`: 最大反復回数 (デフォルト: 100)
- `--auto`: 自動修正活性化 (Level 1-3)
- `--sequential` / `--seq`: 順次診断 (デバッグ用) - 並列がデフォルト
- `--errors`: エラーのみ修正
- `--coverage`: カバレッジ包含 (100%目標)
- `--resume ID`: スナップショット復元

> **パフォーマンス**: 並列診断がデフォルトになり、LSP、AST-grep、Tests、Coverageを同時実行します (3.75倍高速)。

**例**:

```bash
# 基本自律ループ (並列がデフォルト)
> /moai loop

# 順次 + 自動修正 (デバッグ用)
> /moai loop --seq --auto

# 最大50回反復
> /moai loop --max 50

# スナップショット復元
> /moai loop --resume latest
```

---

### 🔧 `/moai fix` - 単発自動修正

```bash
> /moai fix
```

**「一回の実行で、一度に修正」**

LSPエラー、linting問題、AST-grepパターンを並列でスキャンして、検出された問題を一回の実行で一度に修正します。Level 1-2は即時修正し、Level 3はユーザー承認後修正し、Level 4は手動修正が必要だと報告します。`--dry`オプションでプレビュー確認後、実際の修正を適用できます。

---

#### 🎯 コンセプトとワークフロー

`/moai fix`は**一回の実行で終わる単発修正**コマンドです。自律的にループする`/moai loop`とは異なり、スキャン→修正→報告の順序で一度だけ実行します。

```mermaid
flowchart TB
    Start[/moai fix 開始] --> Scan[並列スキャン実行]

    Scan --> LSP[LSP診断]
    Scan --> AST[AST-grep検査]
    Scan --> Linter[Linter実行]
    Scan --> Test[テスト実行]

    LSP --> Gather[結果収集]
    AST --> Gather
    Linter --> Gather
    Test --> Gather

    Gather --> Analyze[問題分析と分類]

    Analyze --> Level1{Level判定}
    Level1 -->|Level 1<br/>即時修正| Fix1[自動修正]
    Level1 -->|Level 2<br/>安全修正| Fix2[ログ記録後修正]
    Level1 -->|Level 3<br/>承認必要| Ask{ユーザー承認}
    Level1 -->|Level 4<br/>手動必要| Report[報告のみ]

    Ask -->|承認| Fix3[修正実行]
    Ask -->|拒否| Skip[スキップ]

    Fix1 --> Verify[修正検証]
    Fix2 --> Verify
    Fix3 --> Verify
    Skip --> Verify
    Report --> End[完了報告]

    Verify --> TestCheck{テスト実行?}
    TestCheck -->|是| TestRun[テスト実行]
    TestCheck -->|否| End
    TestRun --> TestPass{パス?}
    TestPass -->|是| End
    TestPass -->|否| Ask

    style Start fill:#e1f5ff
    style End fill:#c8e6c9
    style LSP fill:#fff3e0
    style AST fill:#fff3e0
    style Linter fill:#fff3e0
    style Test fill:#fff3e0
    style Fix1 fill:#c8e6c9
    style Fix2 fill:#c8e6c9
    style Fix3 fill:#c8e6c9
    style Report fill:#ffcdd2
    style Ask fill:#fff9c4
```

---

#### 🔍 並列スキャン詳細

**4つの診断ツールを並列実行して、3.75倍高速にスキャンします**。

```mermaid
flowchart TB
    Start[並列スキャン開始] --> Parallel[並列実行]

    Parallel --> LSP[LSP診断<br/>-------------------<br/>• TypeScriptエラー<br/>• Pythonタイプエラー<br/>• 未定義変数<br/>• インポートエラー]
    Parallel --> AST[AST-grep検査<br/>-------------------<br/>• セキュリティ脆弱性<br/>• コードスメル<br/>• アンチパターン<br/>• 構造的問題]
    Parallel --> Linter[Linter実行<br/>-------------------<br/>• コードスタイル<br/>• フォーマット<br/>• 未使用インポート<br/>• 命名規則]
    Parallel --> Test[テスト検査<br/>-------------------<br/>• 失敗テスト<br/>• カバレッジ不足<br/>• テストエラー<br/>• スキップテスト]

    LSP --> Merge[結果統合]
    AST --> Merge
    Linter --> Merge
    Test --> Merge

    Merge --> Result[統合結果レポート]

    Result --> Stats[統計: <br/>• LSP: 12件検出<br/>• AST: 5件検出<br/>• Linter: 8件検出<br/>• Test: 3件検出<br/><br/>合計: 28件]

    Stats --> Classify[レベル分類:<br/>• Level 1: 15件<br/>• Level 2: 8件<br/>• Level 3: 4件<br/>• Level 4: 1件]

    Classify --> End[修正開始]

    style Start fill:#e1f5ff
    style End fill:#c8e6c9
    style LSP fill:#fff3e0
    style AST fill:#ffe0b2
    style Linter fill:#e1bee7
    style Test fill:#c8e6c9
    style Parallel fill:#ffecb3
    style Merge fill:#b2dfdb
```

**並列実行によるパフォーマンス向上**:

- 順次実行: 150秒 (LSP 45s + AST 35s + Linter 40s + Test 30s)
- 並列実行: 40秒 (最も遅いツールの時間)
- **高速化**: 3.75倍向上

---

#### 📋 修正レベルとプロセス

各問題は複雑度に応じて4つのレベルに分類されます。

| Level | 説明     | 承認     | 例                                   | 自動修正可否 |
| ----- | -------- | -------- | ------------------------------------ | ------------ |
| 1     | 即時修正 | 不要     | インポート整列、空白、フォーマット   | ✅ 可能      |
| 2     | 安全修正 | ログのみ | 変数名、タイプ追加、単純なリファクタ | ✅ 可能      |
| 3     | 承認必要 | 必要     | ロジック変更、API修正、メソッド置換  | ⚠️ 承認後    |
| 4     | 手動必要 | 不可     | セキュリティ、アーキテクチャ         | ❌ 不可能    |

**修正プロセス**:

1. **スキャン並列実行**: LSP、AST-grep、Linter、Testを同時に実行
2. **結果統合**: すべての診断結果を統合して重複排除
3. **レベル分類**: 各問題をレベル1-4に分類
4. **修正実行**:
   - Level 1: 即時修正（ログのみ記録）
   - Level 2: 安全修正（変更内容をログ記録）
   - Level 3: ユーザー承認後修正
   - Level 4: 手動修正が必要と報告
5. **検証**: 修正後にテスト実行して回帰検証
6. **完了報告**: 修正内容と統計を報告

---

#### 🎬 いつ使うべきか

| 状況                         | 推奨コマンド | 理由                           |
| ---------------------------- | ------------ | ------------------------------ |
| 日常的なコード品質維持       | `/moai fix`  | 迅速な単発修正、ループ不要     |
| CI失敗の原因を一括修正       | `/moai fix`  | 一回の実行で全問題修正         |
| 新機能実装後のクリーンアップ | `/moai fix`  | 一括フォーマットとスタイル修正 |
| 複雑な再発するバグ           | `/moai loop` | 継続的な修正と検証が必要       |
| 大規模リファクタリング       | `/moai loop` | 多段階修正と段階的検証が必要   |
| PR作成前の最終確認           | `/moai fix`  | 1回の実行でクリーンアップ      |
| レガシーコードの大幅改善     | `/moai loop` | 何度も反復して段階的に改善     |

---

#### 🔀 `/moai fix` vs `/moai loop` 選択ガイド

**どちらを使うか迷ったら、このフローチャートで決定できます**。

```mermaid
flowchart TB
    Start[問題発見] --> Q1{問題の性質は?}

    Q1 -->|単純なエラー<br/>フォーマット<br/>スタイル| SimpleCheck{修正回数は?}
    Q1 -->|複雑なロジック<br/>再発する問題<br/>複数ファイル| ComplexCheck{検証が必要?}

    SimpleCheck -->|1回で完了| Fix[**/moai fix**<br/>単発修正]
    SimpleCheck -->|複数回必要| Loop[**/moai loop**<br/>自律反復]

    ComplexCheck -->|継続的検証が必要| Loop
    ComplexCheck -->|一度だけ| Complexity{複雑度は?}

    Complexity -->|アーキテクチャ変更| Manual[手動修正]
    Complexity -->|通常範囲| Fix

    Fix --> FixExample[例:<br/>• import整列<br/>• フォーマット<br/>• タイプエラー修正<br/>• 未使用変数削除<br/>• LSPエラー修正]
    Loop --> LoopExample[例:<br/>• 複雑なバグ修正<br/>• 大規模リファクタ<br/>• 再発する問題<br/>• 多段階改善<br/>• 継続的検証が必要]
    Manual --> ManualExample[例:<br/>• セキュリティ修正<br/>• アーキテクチャ変更<br/>• API設計変更]

    style Start fill:#e1f5ff
    style Fix fill:#c8e6c9
    style Loop fill:#fff9c4
    style Manual fill:#ffcdd2
    style FixExample fill:#e8f5e9
    style LoopExample fill:#fffde7
    style ManualExample fill:#ffebee
```

**要約**:

- **`/moai fix`**: 一回の実行で終わる、日常的な問題修正に最適
- **`/moai loop`**: 継続的な修正と検証が必要な複雑な問題に最適
- **手動修正**: アーキテクチャ変更やセキュリティ修正など、人的判断が必要

---

#### 🛠️ オプションと使用例

```bash
# 基本修正（並列がデフォルト）
> /moai fix

# プレビューのみ（実際修正なし）
> /moai fix --dry

# 順次スキャン（デバッグ用）
> /moai fix --seq

# レベル3以下のみ修正
> /moai fix --level 3

# エラーのみ修正
> /moai fix --errors

# セキュリティ検査包含
> /moai fix --security

# 特定ファイル
> /moai fix src/auth.py

# フォーマットスキップ
> /moai fix --no-fmt
```

---

> **💡 パフォーマンス**: 並列スキャンがデフォルトになり、LSP、AST-grep、Linterを同時実行します（3.75倍高速）。

---

## 4. 🤖 All is Well - エージェンティック自律自動化

**MoAI-ADKの最も強力な機能**: AIが自ら探索し、計画し、実装し、検証を完了マーカー検出まで繰り返します。

### 核心コンセプト

```text
ユーザー: "認証機能追加"
  ↓
AI: 探索 → 計画 → 実装 → 検証 → 繰り返し
  ↓
AI: すべての問題解決
  ↓
AI: <moai>DONE</moai>  ← 完了マーカー
```

### 3つのコマンド階層

| コマンド       | タイプ     | 説明                                          |
| -------------- | ---------- | --------------------------------------------- |
| `/moai fix`    | 単発       | 1回スキャン + 自動修正                        |
| `/moai loop`   | 自律ループ | 完了マーカー または 最大回数まで繰り返し修正  |
| `/moai`        | 完全自律   | 目標 → SPEC → 実装 → ドキュメント全過程自動化 |

### コマンドチェーン関係

```text
/moai
  ├── Phase 0: 並列探索 (Explore + Research + Quality)
  ├── Phase 1: /moai plan (SPEC生成)
  ├── Phase 2: /moai run (DDD実装)
  │     └── /moai loop (自律ループ)
  │           └── /moai fix (単発修正) × N回
  └── Phase 3: /moai sync (ドキュメント同期)
```

### 完了マーカー

AIが作業完了時にマーカーを追加して自律ループが終了します:

| マーカー                | 説明     |
| ----------------------- | -------- |
| `<moai>DONE</moai>`     | 作業完了 |
| `<moai>COMPLETE</moai>` | 全体完了 |
| `<moai:done />`         | XML形式  |

### 自動修正レベル

| Level | 説明     | 承認     | 例                           |
| ----- | -------- | -------- | ---------------------------- |
| 1     | 即時修正 | 不要     | import整列、空白             |
| 2     | 安全修正 | ログのみ | 変数名、タイプ追加           |
| 3     | 承認必要 | 必要     | ロジック変更、API修正        |
| 4     | 手動必要 | 不可     | セキュリティ、アーキテクチャ |

### クイックスタート

```bash
# 1回修正 (並列がデフォルト)
> /moai fix

# 自律ループ (完了マーカーまで、並列がデフォルト)
> /moai loop --auto

# 完全自律自動化 (All is Well!、並列がデフォルト)
> /moai "JWT認証追加" --loop

# 続きから
> /moai resume SPEC-AUTH-001
```

---

## 5. MoAIオーケストレーターとSub-Agents

### 🎩 MoAI - Strategic Orchestrator (戦略的オーケストレーター)

**役割**: ユーザーリクエストを分析して適切な専門エージェントに委任

**作業フロー**:

1. **Understand**: リクエスト分析及び明確化
2. **Plan**: Planエージェントを通じ実行計画策定
3. **Execute**: 専門エージェントに作業委任 (順次/並列)
4. **Integrate**: 結果統合及びユーザー報告

### 🌐 多言語自動ルーティング (NEW)

MoAIは4つの言語リクエストを自動認識して正しいエージェントを呼び出します:

| リクエスト言語 | 例                          | 呼出エージェント |
| -------------- | --------------------------- | ---------------- |
| 英語           | "Design backend API"        | expert-backend   |
| 韓国語         | "백엔드 API 설계해줘"       | expert-backend   |
| 日本語         | "バックエンドAPIを設計して" | expert-backend   |
| 中国語         | "设计后端API"               | expert-backend   |

---

### 🔧 Tier 1: ドメイン専門家 (9個)

| エージェント           | 専門分野                          | 使用例                       |
| ---------------------- | --------------------------------- | ---------------------------- |
| **expert-backend**     | FastAPI, Django, DB設計           | API設計、クエリ最適化        |
| **expert-frontend**    | React, Vue, Next.js               | UIコンポーネント、状態管理   |
| **expert-stitch**      | Google Stitch, UI/UXデザイン      | AI駆動UI生成                 |
| **expert-security**    | セキュリティ分析、OWASP           | セキュリティ監査、脆弱性分析 |
| **expert-devops**      | Docker, K8s, CI/CD                | デプロイ自動化、インフラ     |
| **expert-debug**       | バグ分析、性能                    | 問題診断、ボトルネック解決   |
| **expert-performance** | プロファイリング、最適化          | 応答時間改善                 |
| **expert-refactoring** | コードリファクタリング、AST-Grep  | 大規模コード変換             |
| **expert-testing**     | テスト戦略、E2E                   | テスト計画、カバレッジ       |

---

### 🎯 Tier 2: ワークフロー管理者 (7個)

| エージェント           | 役割                      | 自動呼出時期      |
| --------------------- | ------------------------- | ----------------- |
| **manager-spec**      | SPEC作成 (EARS)          | `/moai plan`    |
| **manager-ddd**       | DDD実装実行               | `/moai run`     |
| **manager-docs**      | ドキュメント自動生成      | `/moai sync`    |
| **manager-quality**   | TRUST 5検証               | 実装完了後        |
| **manager-strategy**  | 実行戦略策立               | 複雑な企画時      |
| **manager-project**   | プロジェクト初期化 & 設定 | `/moai project` |
| **manager-git**       | Gitワークフロー           | ブランチ/PR管理   |

---

### 🏗️ Tier 3: Claude Code Builder (4個)

| エージェント        | 役割               | 使用例                     |
| ------------------- | ------------------ | -------------------------- |
| **builder-agent**   | 新エージェント生成 | 組織専門家エージェント     |
| **builder-skill**   | 新スキル生成       | チーム専用スキルモジュール |
| **builder-command** | 新コマンド生成     | カスタムワークフロー       |
| **builder-plugin**  | プラグイン生成     | デプロイ用プラグイン       |

---

### 🧠 Sequential Thinking MCP サポート

すべてのエージェントは、深層分析のためにSequential Thinking MCPを使用する `--ultrathink` フラグをサポートしています：

**使用方法**:
```bash
> /moai "JWT認証追加" --ultrathink
```

**エージェント別 UltraThink 例**:

| エージェントタイプ    | UltraThink 深層分析フォーカー                              |
| ------------------- | --------------------------------------------------------- |
| **expert-backend**  | API設計パターン、データベーススキーマ、クエリ最適化        |
| **expert-frontend** | コンポーネントアーキテクチャ、状態管理、UI/UX設計           |
| **expert-security** | 脅威分析、脆弱性パターン、OWASP準拠                        |
| **expert-devops**   | デプロイ戦略、CI/CDパイプライン、インフラ                   |
| **manager-ddd**     | リファクタリング戦略、振る舞い保存、レガシーコード          |
| **manager-spec**    | 要件分析、受入基準、ユーザーストーリー                      |

`--ultrathink`がエージェント呼び出しに追加されると、エージェントはSequential Thinking MCPを有効にして：
- 複雑な問題を管理可能なステップに分解
- ドメイン別パターンとベストプラクティスを分析
- 適切な実行戦略にサブタスクをマッピング
- 最適な実装計画を生成

---

## 6. Agent-Skills

### 📚 スキルライブラリ構造

```text
🏗️ Foundation (6)    → 核心哲学、実行ルール
🎯 Domain (4)        → ドメイン専門知識
💻 Language (16)     → 16個プログラミング言語
🚀 Platform (10)     → クラウド/BaaS統合
📋 Workflow (8)      → 自動化ワークフロー
📚 Library (3)       → 特殊ライブラリ
🛠️ Tool (2)          → 開発ツール
📑 Docs (1)          → ドキュメント生成
📊 Formats (1)       → データフォーマット処理
🖥️ Framework (1)     → アプリケーションフレームワーク
```

### よく使うスキル組合せ

| 目的                 | スキル組合せ                                                             |
| -------------------- | ------------------------------------------------------------------------ |
| **バックエンドAPI**  | `moai-lang-python` + `moai-domain-backend` + `moai-platform-supabase`    |
| **フロントエンドUI** | `moai-lang-typescript` + `moai-domain-frontend` + `moai-library-shadcn`  |
| **ドキュメント生成** | `moai-library-nextra` + `moai-workflow-docs` + `moai-library-mermaid`    |
| **テスト**           | `moai-lang-python` + `moai-workflow-testing` + `moai-foundation-quality` |

### スキル使用法

```python
# 方法1: 直接呼出 (Agent)
Skill("moai-lang-python")

# 方法2: MoAI自動選択 (一般ユーザー)
"FastAPIサーバー作成して"
→ MoAIが自動でmoai-lang-python選択
```

---

## 7. Google Stitch MCP - AI駆動UI/UXデザイン

### 概要

**Google Stitch**はテキスト説明からUI画面を生成するAI駆動デザインツールです。MoAI-ADKはStitch MCP統合を通じてデザインコンテキスト抽出、画面生成、コードエクスポートを自動化します。

### 主要機能

| 機能                          | 説明                                        |
| ----------------------------- | ------------------------------------------- |
| `generate_screen_from_text`   | テキスト説明からUI画面を生成                |
| `extract_design_context`      | 既存画面から"Design DNA"を抽出 (色、フォント、レイアウト) |
| `fetch_screen_code`           | 生成された画面のHTML/CSS/JSコードをダウンロード     |
| `fetch_screen_image`          | 画面スクリーンショットをダウンロード                      |
| `create_project` / `list_projects` | Stitchプロジェクト管理                   |

### クイックスタート

**1. Google Cloud設定**:

```bash
# Stitch API有効化
gcloud beta services mcp enable stitch.googleapis.com

# 認証
gcloud auth application-default login

# 環境変数設定 (.bashrc または .zshrc)
export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"
```

**2. MCP設定** (`.claude/.mcp.json`):

```json
{
  "mcpServers": {
    "stitch": {
      "command": "npx",
      "args": ["-y", "stitch-mcp"],
      "env": {
        "GOOGLE_CLOUD_PROJECT": "YOUR_PROJECT_ID"
      }
    }
  }
}
```

### 使用例

**ログイン画面の生成**:

```text
> ログイン画面を作成してください。メール入力、パスワード入力（表示/非表示トグル）、
> ログインボタン、パスワードリセットリンク、ソーシャルログイン（Google、Apple）を含みます。
> モバイル: 縦スタック。デスクトップ: 400pxカード中央配置。
```

**デザイン一貫性維持ワークフロー**:

1. `extract_design_context`: 既存画面からデザイントークンを抽出
2. `generate_screen_from_text`: 抽出されたコンテキストで新画面を生成
3. `fetch_screen_code`: プロダクションコードをエクスポート

### プロンプト作成のコツ

| 項目 | 推奨 |
| ---- | ---- |
| **コンポーネント** | ボタン、入力、カードなど必要なUI要素を明示 |
| **レイアウト** | single-column、grid、sidebarなどを指定 |
| **レスポンシブ** | モバイル/デスクトップ動作を明示 |
| **スタイル** | 色、フォント、ホバー効果を指定 |

> **注意**: 一度に一つの画面、一つか二つの修正のみをリクエストすることが最良の結果を得る方法です。

### 詳細ドキュメント

完全なプロンプトテンプレート、エラー処理、高度なパターンについてはスキルドキュメントを参照してください:

- **スキル**: `.claude/skills/moai-platform-stitch/SKILL.md`
- **エージェント**: `expert-stitch` (UI/UXデザイン専門エージェント)

---

## 7.1 Memory MCP - セッション間永続ストレージ

### 概要

**Memory MCP**はClaude Codeセッション間に永続ストレージを提供し、MoAIがユーザー設定、プロジェクトコンテキスト、学習パターンを記憶できるようにします。

### 主な機能

| 機能 | 説明 |
| --- | --- |
| **ユーザー設定** | 会話言語、コーディングスタイル、命名規則を記憶 |
| **プロジェクトコンテキスト** | 技術スタック、アーキテクチャ決定、プロジェクト慣習を保存 |
| **学習パターン** | よく使用するライブラリ、一般的なエラー解決策を保存 |
| **セッション状態** | 最後に作業したSPEC、保留中のタスクを追跡 |

### メモリカテゴリ

| 接頭辞 | カテゴリ | 例 |
| --- | --- | --- |
| `user_` | ユーザー設定 | `user_language`, `user_coding_style` |
| `project_` | プロジェクトコンテキスト | `project_tech_stack`, `project_architecture` |
| `pattern_` | 学習パターン | `pattern_preferred_libraries`, `pattern_error_resolutions` |
| `session_` | セッション状態 | `session_last_spec`, `session_pending_tasks` |

### インストール

Claude Code設定にMemory MCPを追加:

```json
// .claude/settings.local.json
{
  "mcpServers": {
    "memory": {
      "command": "npx",
      "args": ["-y", "@anthropic/memory-mcp-server"]
    }
  }
}
```

### 使用例

**ユーザー設定の保存**:
```
「会話は日本語でお願いします」
→ MoAI保存: user_language = "ja"
```

**修正から学習**:
```
「Python変数はsnake_caseで書いてください」
→ MoAI保存: user_coding_style = "snake_case"
```

**コンテキストの取得**:
```
「最後に作業していたSPECは何でしたか？」
→ MoAI取得: session_last_spec
```

### ベストプラクティス

- 説明的でカテゴリ化されたキー名を使用
- 値は簡潔に（1000文字以下）
- 機密認証情報は保存しない
- 個人データではなく設定を保存

---

## 8. TRUST 5品質原則

MoAI-ADKのすべてプロジェクトは**TRUST 5**品質フレームワークに従います。

### 🏆 TRUST 5 = Test + Readable + Unified + Secured + Trackable

```mermaid
graph TD
    T1["🔴 T: Test-First<br/>━━━━━━━━<br/>• DDD Analyze-Preserve-Improve<br/>• 85%+ カバレッジ<br/>• 自動テスト"]
    R["📖 R: Readable<br/>━━━━━━━━<br/>• 明確な命名<br/>• コードコメント<br/>• リンター準拠"]
    U["🔄 U: Unified<br/>━━━━━━━━<br/>• 一貫したスタイル<br/>• 標準パターン<br/>• エラー処理"]
    S["🔒 S: Secured<br/>━━━━━━━━<br/>• OWASP Top 10<br/>• 脆弱性スキャン<br/>• 暗号化ポリシー"]
    T2["📋 T: Trackable<br/>━━━━━━━━<br/>• 明確なコミット<br/>• イシュー追跡<br/>• CHANGELOG"]

    T1 --> R --> U --> S --> T2 --> Deploy["✅ Production Ready"]
```

### T - Test-First (テスト優先)

**原則**: すべての実装はテストから開始

**検証**:

- テストカバレッジ >= 85%
- ドメイン境界分析 (Analyze)
- 動作保存テスト (Preserve)
- 漸進的改善 (Improve)

### R - Readable (可読性)

**原則**: コードは明確で理解しやすいべき

**検証**:

- 明確な変数名
- 複雑なロジックにコメント
- コードレビュー通過
- リンター検査通過

### U - Unified (統一性)

**原則**: プロジェクト全体に一貫したスタイル維持

**検証**:

- プロジェクトスタイルガイド準拠
- 一貫した命名規則
- 統一されたエラー処理
- 標準ドキュメント形式

### S - Secured (セキュリティ)

**原則**: すべてのコードはセキュリティ検査通過

**検証**:

- OWASP Top 10チェック
- 依存関係脆弱性スキャン
- 暗号化ポリシー準拠
- アクセス制御検証

### T - Trackable (追跡可能性)

**原則**: すべての変更は明確に追跡可能

**検証**:

- 明確なコミットメッセージ
- イシュー追跡 (GitHub Issues)
- CHANGELOG維持
- コードレビュー記録

---

## 9. 自動品質検査

### 🔍 AST-Grep基盤構造的検査

**AST-Grep**はテキストではなく**コード構造**を分析します:

| 機能                         | 説明                  | 例                                           |
| ---------------------------- | --------------------- | -------------------------------------------- |
| **構造的検索**               | ASTパターンマッチング | パラメータ化されないSQLクエリ探求            |
| **セキュリティスキャン**     | 自動脆弱点探知        | SQL Injection, XSS, ハードコードされた秘密鍵 |
| **パターンリファクタリング** | 安全コード変換        | 変数名一括変更、関数抽出                     |
| **多言語サポート**           | 40+ 言語              | Python, TypeScript, Go, Rust...              |

### 自動検査フロー

```text
コード作成
    ↓
[Hook] AST-Grep 自動スキャン
    ↓
⚠️  脆弱点発見時即時通知
    ↓
✅ 安全コードでリファクタリング
```

**検出例**:

```bash
⚠️  AST-Grep: Potential SQL injection in src/auth.py:47
   Pattern: execute(f"SELECT * FROM users WHERE id={user_id}")
   Suggestion: execute("SELECT * FROM users WHERE id=%s", (user_id,))
```

---

### 🛡️ セキュリティガード - コマンド保護

MoAI-ADKには危険な操作から保護する**セキュリティガードフック**が含まれています：

| カテゴリ | 保護対象コマンド | プラットフォーム |
|----------|------------------|------------------|
| **データベース削除** | `supabase db reset`, `neon database delete`, `pscale database delete` | 全て |
| **SQL危険コマンド** | `DROP DATABASE`, `DROP SCHEMA`, `TRUNCATE TABLE` | 全て |
| **ファイル削除** | `rm -rf /`, `rm -rf ~`, `rm -rf .git` | Unix |
| **ファイル削除** | `rd /s /q C:\`, `Remove-Item -Recurse -Force` | Windows |
| **Git危険コマンド** | `git push --force origin main`, `git branch -D main` | 全て |
| **クラウドインフラ** | `terraform destroy`, `az group delete`, `aws delete-*` | 全て |
| **Dockerクリーンアップ** | `docker system prune -a`, `docker volume prune`, `docker image prune -a` | 全て |

**保護レベル**：

| レベル | 動作 | 例 |
|--------|------|-----|
| **拒否** | 即座にブロック | `rm -rf /`, `DROP DATABASE`, `docker system prune -a` |
| **確認** | ユーザー確認が必要 | `git reset --hard`, `prisma migrate reset` |
| **許可** | 正常に進行 | 安全な操作 |

**動作方式**：

```text
ユーザーがコマンドを実行
    ↓
[Hook] セキュリティガードスキャン
    ↓
⚠️  危険パターン検出 → ブロックまたは確認
    ↓
✅ 安全に進行
```

> **注意**: セキュリティガードはプラットフォーム固有のコマンドパターンでUnix（macOS/Linux）とWindowsの両方のユーザーを保護します。

**手動実行**: 実際にこれらのコマンドが必要な場合は、ターミナルで直接実行してください：

```bash
# Dockerクリーンアップ（必要時に手動実行）
docker system prune -a        # 未使用のイメージ、コンテナ、ネットワークを全て削除
docker volume prune           # 未使用ボリュームを削除（⚠️ データ損失リスク）
docker image prune -a         # 未使用イメージを全て削除

# データベース操作（必要時に手動実行）
supabase db reset             # ローカルデータベースをリセット
DROP DATABASE dbname;         # SQL: データベースを削除

# ファイル操作（必要時に手動実行）
rm -rf node_modules           # node_modulesを削除
```

---

## 10. 📊 Statuslineカスタマイズ

MoAI-ADKはClaude Codeターミナルにリアルタイム状態情報を表示するカスタマイズ可能なstatuslineを提供します。

### デフォルトレイアウト

```text
🤖 Opus 4.5 | 💰 152K/200K | 💬 MoAI | 📁 MoAI-ADK | 📊 +0 M58 ?5 | 💾 57.7MB | 🔀 main
```

### 使用可能なコンポーネント

| アイコン | コンポーネント | 説明                                            | 設定キー         |
| -------- | -------------- | ----------------------------------------------- | ---------------- |
| 🤖       | モデル         | Claudeモデル (Opus, Sonnet等)                   | `model`          |
| 💰       | コンテキスト   | コンテキストウィンドウ使用量 (例: 77K/200K)     | `context_window` |
| 💬       | スタイル       | アクティブアウトプットスタイル (例: MoAI)       | `output_style`   |
| 📁       | ディレクトリ   | 現在のプロジェクト名                            | `directory`      |
| 📊       | Git状態        | ステージング/修正/追跡されないファイル数        | `git_status`     |
| 💾       | メモリ         | プロセスメモリ使用量                            | `memory_usage`   |
| 🔀       | ブランチ       | 現在のGitブランチ                               | `branch`         |
| 🔅       | バージョン     | Claude Codeバージョン (オプション)              | `version`        |

### 設定

`.moai/config/statusline-config.yaml`で表示項目を設定します:

```yaml
display:
  model: true # 🤖 Claudeモデル
  context_window: true # 💰 コンテキストウィンドウ
  output_style: true # 💬 アウトプットスタイル
  directory: true # 📁 プロジェクト名
  git_status: true # 📊 Git状態
  memory_usage: true # 💾 メモリ使用量
  branch: true # 🔀 Gitブランチ
  version: 1.1.0 # 🔅 バージョン (オプション)
  active_task: true # アクティブタスク
```

### メモリコレクター

`memory_usage`が有効な場合、MoAI-ADKは`psutil`を使用してリアルタイムメモリ使用量を収集します:

- **プロセスメモリ**: 現在のPythonプロセスのRSS (常駐セットサイズ)
- **キャッシング**: 10秒TTLで性能最適化
- **クロスプラットフォーム**: macOS, Linux, Windowsサポート
- **グレースフルデグラデーション**: psutilが使用不可な場合は"N/A"表示

### ディスプレイモード

| モード     | 最大長  | 使用ケース       |
| ---------- | ------- | ---------------- |
| `compact`  | 80文字  | 標準ターミナル   |
| `extended` | 120文字 | ワイドターミナル |
| `minimal`  | 40文字  | 狭いターミナル   |

モード設定:

```bash
export MOAI_STATUSLINE_MODE=extended
```

---

## 11. 🌳 Worktree並列開発

MoAI-ADKの核心革新: **Worktreeで完全分離、無制限並列開発**

### 💡 なぜWorktreeなのか？

**問題点**: `moai glm`/`moai cc`でLLMを変更すると**すべて開かれたセッション**に適用されます。同一セッションでモデルを変更すると認証エラーでその後進行が困難です。

**解決策**: Git Worktreeで各SPECを完全分離して独立LLM設定維持

---

### 📦 Worktreeワークフロー

```text
┌─────────────────────────────────────────────────────────────────┐
│  ターミナル1 (Claude Opus) - SPEC設計専用                        │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ cd my-project                                                │
│  $ claude                                                        │
│                                                                  │
│  > /moai plan "ユーザー認証システム" --worktree                   │
│  ✅ SPEC-AUTH-001 生成完了                                      │
│  ✅ Worktree生成: ~/moai/worktrees/my-project/SPEC-AUTH-001     │
│  ✅ Branch: feature/SPEC-AUTH-001                                │
│                                                                  │
│  > /moai plan "決済システム" --worktree                          │
│  ✅ SPEC-PAY-002 生成完了                                       │
│  ✅ Worktree生成: ~/moai/worktrees/my-project/SPEC-PAY-002      │
│                                                                  │
│  > /moai plan "ダッシュボードUI" --worktree                         │
│  ✅ SPEC-UI-003 生成完                                        │
│  ✅ Worktree生成: ~/moai/worktrees/my-project/SPEC-UI-003       │
│                                                                  │
│  💡 OpusですべてSPEC計画完了 (セッション維持中...)                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  ターミナル2 - SPEC-AUTH-001 Worktree (GLM 4.7)                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-worktree go SPEC-AUTH-001                                │
│  # または省略形: moai-wt go SPEC-AUTH-001                          │
│                                                                  │
│  📁 現在位置: ~/moai/worktrees/my-project/SPEC-AUTH-001        │
│  🔀 Branch: feature/SPEC-AUTH-001                                │
│                                                                  │
│  $ moai glm                                                       │
│  ✅ Switched to GLM backend                                      │
│                                                                  │
│  $ claude                                                        │
│  > /moai run SPEC-AUTH-001                                     │
│  🔄 DDD実行中... (Analyze → Preserve → Improve)                 │
│  ✅ 実装完了!                                                   │
│  ✅ テスト通過 (Coverage: 92%)                                  │
│                                                                  │
│  > /moai sync SPEC-AUTH-001                                    │
│  ✅ ドキュメント同期完了                                             │
│                                                                  │
│  # 完了後マージ                                                   │
│  $ git checkout main                                             │
│  $ git merge feature/SPEC-AUTH-001                               │
│  $ moai-worktree clean --merged-only                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  ターミナル3 - SPEC-PAY-002 Worktree (GLM 4.7)                     │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-wt go SPEC-PAY-002                                       │
│  $ moai glm                                                       │
│  $ claude                                                        │
│                                                                  │
│  > /moai SPEC-PAY-002                                     │
│  🔄 Plan → Run → Sync 自動実行                                  │
│  ✅ 完了!                                                        │
│                                                                  │
│  $ git checkout main && git merge feature/SPEC-PAY-002           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  ターミナル4 - SPEC-UI-003 Worktree (GLM 4.7)                      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-wt go SPEC-UI-003                                        │
│  $ moai glm                                                       │
│  $ claude                                                        │
│  > /moai SPEC-UI-003                                      │
│  ✅ 完了!                                                        │
└─────────────────────────────────────────────────────────────────┘
```

---

### 🎯 核心ワークフロー

#### Phase 1: Claude 4.5 Opusで計画 (ターミナル1)

```bash
/moai plan "機能説明" --worktree
```

- ✅ SPECドキュメント生成
- ✅ Worktree自動生成
- ✅ 機能ブランチ自動生成

#### Phase 2: GLM 4.7で実装 (ターミナル2, 3, 4...)

```bash
moai-wt go SPEC-ID
moai glm
claude
> /moai run SPEC-ID
> /moai sync SPEC-ID
```

- ✅ 独立した作業環境
- ✅ GLMコスト効率
- ✅ 衝突なし並列開発

**Phase 3: マージ及び整理**

```bash
# 方法1: 一括完了 (推奨)
moai-wt done SPEC-ID              # checkout main → merge → cleanup
moai-wt done SPEC-ID --push       # 上記 + リモートプッシュ

# 方法2: 手動処理
git checkout main
git merge feature/SPEC-ID
moai-wt clean --merged-only
```

---

### ✨ Worktree長所

| 長所           | 説明                                  |
| -------------- | ------------------------------------- |
| **完全分離**   | 各SPECが独立Git状態、ファイル衝突なし |
| **LLM独立**    | 各Worktreeで別LLM設定可能             |
| **無制限並列** | 依存性なし無制限SPEC並列開発          |
| **安全マージ** | 完了SPECのみ順次的にmainにマージ      |

---

### 📊 Worktreeコマンド

| コマンド                 | 説明                           | 使用例                         |
| ------------------------ | ------------------------------ | ------------------------------ |
| `moai-wt new SPEC-ID`    | 新Worktree生成                 | `moai-wt new SPEC-AUTH-001`    |
| `moai-wt go SPEC-ID`     | Worktree進入 (新シェル開く)    | `moai-wt go SPEC-AUTH-001`     |
| `moai-wt list`           | Worktreeリスト確認             | `moai-wt list`                 |
| `moai-wt done SPEC-ID`   | マージ後整理 (checkout→merge)  | `moai-wt done SPEC-AUTH-001`   |
| `moai-wt remove SPEC-ID` | Worktree削除                   | `moai-wt remove SPEC-AUTH-001` |
| `moai-wt status`         | Worktree状態及びレジストリ確認 | `moai-wt status`               |
| `moai-wt sync [SPEC-ID]` | Worktree同期                   | `moai-wt sync --all`           |
| `moai-wt clean`          | マージWorktree整理             | `moai-wt clean --merged-only`  |
| `moai-wt recover`        | ディスクからレジストリ復復     | `moai-wt recover`              |
| `moai-wt config`         | Worktree設定確認               | `moai-wt config root`          |

---

## 12. CLAUDE.mdの理解

MoAI-ADKインストール後、プロジェクトルートに生成される`CLAUDE.md`は**MoAI（AIオーケストレーター）の実行指示書**です。このファイルはClaude Codeがプロジェクトでどのように動作するかを定義します。

### CLAUDE.mdとは？

`CLAUDE.md`はClaude Codeがセッション開始時に自動的に読み込むプロジェクト設定ファイルです。MoAI-ADKではこのファイルを通じて**MoAIオーケストレーター**の行動規則を定義します。

```text
📁 プロジェクトルート
├── CLAUDE.md              ← MoAI実行指示書（変更非推奨）
├── CLAUDE.local.md        ← 個人カスタム指示（オプション）
├── .claude/
│   ├── settings.json      ← Claude Code設定（更新時に上書き）
│   ├── settings.local.json← 個人設定（オプション、上書きなし）
│   ├── agents/            ← サブエージェント定義
│   ├── commands/          ← スラッシュコマンド
│   └── skills/            ← スキル定義
└── .moai/
    └── config/            ← MoAI設定
```

### CLAUDE.mdコア構造

| セクション                      | 説明                   | 主要内容                                               |
| ------------------------------- | ---------------------- | ------------------------------------------------------ |
| **Core Identity**               | MoAIの役割定義         | 戦略的オーケストレーター、HARDルール                   |
| **Request Processing Pipeline** | リクエスト処理フロー   | Analyze → Route → Execute → Report                     |
| **Command Reference**           | コマンド分類           | Type A (Workflow), Type B (Utility), Type C (Feedback) |
| **Agent Catalog**               | サブエージェントリスト | Manager 8個, Expert 8個, Builder 4個                   |
| **SPEC-Based Workflow**         | SPEC基盤開発           | Plan → Run → Sync フロー                               |
| **Quality Gates**               | 品質検証ルール         | HARD/SOFTルールチェックリスト                          |
| **Configuration Reference**     | 設定参照               | 言語、出力形式ルール                                   |

### 使用方法：変更しないでください

> **推奨**：`CLAUDE.md`は**変更せずそのまま使用**してください。

**理由**：

- MoAI-ADKアップデート時に自動的に最新バージョンに置き換えられます
- 変更するとアップデート時に競合が発生する可能性があります
- エージェント間の一貫した動作を保証します

```bash
# アップデート時CLAUDE.mdは自動的に最新化
moai update
```

### カスタマイズ：CLAUDE.local.mdを使用

追加指示が必要な場合は**`CLAUDE.local.md`**ファイルを作成してください。

```bash
# プロジェクトルートにCLAUDE.local.mdを作成
touch CLAUDE.local.md
```

**CLAUDE.local.md例**：

```markdown
# プロジェクトローカル指示

## コーディングスタイル

- すべての関数に型ヒント必須
- docstringはGoogleスタイル使用

## プロジェクト特殊ルール

- APIレスポンスは常にsnake_case使用
- テストファイルはtest\_プレフィックス必須

## 禁止事項

- console.log使用禁止（logger使用）
- any型使用禁止
```

**利点**：

- `CLAUDE.md`アップデートと競合なし
- プロジェクト別カスタム設定可能
- `.gitignore`に追加して個人設定維持可能

### CLAUDE.md vs CLAUDE.local.md

| 区分             | CLAUDE.md      | CLAUDE.local.md           |
| ---------------- | -------------- | ------------------------- |
| **目的**         | MoAI実行指示   | 個人/プロジェクト追加指示 |
| **変更**         | 非推奨         | 自由に変更可能            |
| **アップデート** | MoAIが自動管理 | ユーザーが直接管理        |
| **Git**          | コミット対象   | 選択（.gitignore可能）    |
| **優先順位**     | 基本ルール     | 追加/オーバーライドルール |

### 設定カスタマイズ：settings.local.jsonの使用

v1.8.12から、`.claude/settings.json`は`moai update`実行時に**常に上書き**されます。個人設定が必要な場合は**`settings.local.json`**ファイルを作成してください。

```bash
# .claude/ディレクトリにsettings.local.jsonを作成
touch .claude/settings.local.json
```

**settings.local.json例**：

```json
{
  "env": {
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "DISABLE_NON_ESSENTIAL_MODEL_CALLS": "1",
    "MY_CUSTOM_VAR": "value"
  },
  "permissions": {
    "allow": [
      "Bash(docker:*)",
      "Bash(kubectl:*)"
    ]
  }
}
```

**利点**：

- `settings.json`アップデートと競合なし
- 個人環境変数と権限設定可能
- `.gitignore`に追加してプライベート設定維持可能

### settings.json vs settings.local.json

| 区分             | settings.json     | settings.local.json          |
| ---------------- | ----------------- | ---------------------------- |
| **目的**         | MoAIデフォルト設定 | 個人/プロジェクト追加設定    |
| **変更**         | 非推奨            | 自由に変更可能               |
| **アップデート** | MoAIが上書き      | ユーザーが直接管理           |
| **Git**          | コミット対象      | 選択（.gitignore可能）       |
| **優先順位**     | 基本設定          | マージ（高い優先順位）       |

### コアルール（HARD Rules）

`CLAUDE.md`に定義された**HARDルール**は必ず遵守されます：

1. **Language-Aware Responses**：ユーザー言語で応答
2. **Parallel Execution**：独立タスクは並列実行
3. **No XML in User Responses**：ユーザー応答にXMLタグ非表示

これらのルールは`CLAUDE.local.md`でもオーバーライドできません。

---

## 13. MoAI Rank 紹介

**エージェンティックコーディングの新次元**: あなたのコーディング旅を追跡して、グローバル開発者たちと競争してください！

### なぜMoAI Rankなのか？

| 機能                            | 説明                         |
| ------------------------------- | ---------------------------- |
| **📊 トークントラッキング**     | セッション別AI使用量自動記録 |
| **🏆 グローバルリーダーボード** | 日間/週間/月間/全体順位      |
| **🎭 コーディングスタイル分析** | あなただけの開発パターン発見 |
| **📈 ダッシュボード**           | 可視化された統計とインサイト |

---

### 🚀 CLIコマンド

```bash
❯ moai rank
Usage: moai rank [OPTIONS] COMMAND [ARGS]...

  MoAI Rank - Token usage leaderboard.

  Track your Claude Code token usage and compete on the leaderboard.
  Visit https://rank.mo.ai.kr for the web dashboard.

Commands:
  register   Register with MoAI Rank via GitHub OAuth.
  status     Show your current rank and statistics.
  exclude    Exclude a project from session tracking.
  include    Re-include a previously excluded project.
  logout     Remove stored MoAI Rank credentials.
```

---

### Step 1: GitHub OAuth登録

```bash
❯ moai rank login

╭──────────────────────────── Registration ────────────────────────────╮
│ MoAI Rank Registration                                               │
│                                                                      │
│ This will open your browser to authorize with GitHub.                │
│ After authorization, your API key will be stored securely.           │
╰──────────────────────────────────────────────────────────────────────╯

Opening browser for GitHub authorization...
Waiting for authorization (timeout: 5 minutes)...

╭───────────────────────── Registration Complete ──────────────────────╮
│ Successfully registered as your-github-id                            │
│                                                                      │
│ API Key: moai_rank_a9011fac_c...                                     │
│ Stored in: ~/.moai/rank/credentials.json                             │
╰──────────────────────────────────────────────────────────────────────╯

╭───────────────────────── Global Hook Installed ──────────────────────╮
│ Session tracking hook installed globally.                            │
│                                                                      │
│ Your Claude Code sessions will be automatically tracked.             │
│ Hook location: ~/.claude/hooks/moai/session_end__rank_submit.py      │
│                                                                      │
│ To exclude specific projects:                                        │
│   moai rank exclude /path/to/project                                 │
╰──────────────────────────────────────────────────────────────────────╯
```

---

### Step 2: セッションデータ同期

既存のClaude CodeセッションデータをMoAI Rankに同期します。

```bash
❯ moai rank sync

Syncing 2577 session(s) to MoAI Rank
Phase 1: Parsing transcripts (parallel: 20 workers)

  Parsing transcripts ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (2577/2577)

Phase 2: Submitting 1873 session(s) (batch mode)
Batch size: 100 | Batches: 19

  Submitting batches ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (19/19)

Sync Complete
  ✓ Submitted: 1169
  ○ Skipped:   704 (no usage or duplicate)
  ✗ Failed:    0
```

---

### Step 3: 私の順位確認

```bash
❯ moai rank status

╭────────────────────────────── MoAI Rank ─────────────────────────────╮
│ your-github-id                                                       │
│                                                                      │
│ 🏆 Global Rank: #42                                                  │
╰──────────────────────────────────────────────────────────────────────╯
╭───── Daily ──────╮  ╭───── Weekly ─────╮  ╭──── Monthly ─────╮  ╭──── All Time ────╮
│ #12              │  │ #28              │  │ #42              │  │ #156             │
╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯
╭─────────────────────────── Token Usage ──────────────────────────────╮
│ 1,247,832 total tokens                                               │
│                                                                      │
│ Input  ██████████████░░░░░░ 847,291 (68%)                            │
│ Output ██████░░░░░░░░░░░░░░ 400,541 (32%)                            │
│                                                                      │
│ Sessions: 47                                                         │
╰──────────────────────────────────────────────────────────────────────╯

● Hook: Installed  |  https://rank.mo.ai.kr
```

---

### Step 3: ウェブダッシュボード

![MoAI Rank Dashboard](./assets/images/readme/moai-rank-dashboard.png)

**[https://rank.mo.ai.kr](https://rank.mo.ai.kr)**

ダッシュボードで:

- トークン使用量推移
- ツール使用統計
- モデル別使用分析
- 週間/月間レポート

📖 **詳細**: [modu-ai/moai-rank](https://github.com/modu-ai/moai-rank)リポジトリを参照してください。

---

### Step 4: 収集メトリック

| メトリック           | 説明                                  |
| -------------------- | ------------------------------------- |
| **トークン使用量**   | 入力/出力トークン、キャッシュトークン |
| **ツール使用**       | Read, Edit, Bash等使用回数            |
| **モデル使用**       | Opus, Sonnet, Haiku別分量             |
| **コードメトリック** | 追加/削除ライン、修正ファイル         |
| **セッション情報**   | 継続時間、ターン数、タイムスタンプ    |

### 🔒 プライバシー保護

```bash
# 現在プロジェクト除外
moai rank exclude

# 特定パス除外
moai rank exclude /path/to/private

# ワイルドカードパターン
moai rank exclude "*/confidential/*"

# 除外リスト確認
moai rank list-excluded
```

**保証**: 収集データは**数値メトリックのみ** (コード内容、ファイルパス非転送)

---

## 14. FAQ 5個

### Q1: SPECは常に必要ですか？

| 条件            | SPEC必要有無    |
| --------------- | --------------- |
| 1-2ファイル修正 | 選択 (省略可能) |
| 3-5ファイル修正 | 推奨            |
| 10+ファイル修正 | 必須            |
| 新機能追加      | 推奨            |
| バグ修正        | 選択            |

### Q2: MCPサーバーインストールが必要ですか？

**必須 (2個)**:

- **Context7**: 最新ライブラリドキュメント、Skill参照生成
- **Sequential Thinking**: 複雑なタスクでの構造化された問題解決と段階的推論

**推奨**:

- **Memory MCP**: セッション間永続ストレージ（ユーザー設定、プロジェクトコンテキスト、学習パターン）

**選択**:

- claude-in-chrome: ブラウザでClaude使用及びウェブ自動化テスト
- Playwright: ウェブ自動化テスト
- Figma: デザインシステム

### Q3: MoAI Rankは費用がかかりますか？

無料です。セッションデータのみ自動収集します。

### Q4: GLM設定は必須ですか？

違います。Claudeのみ使用でも良いです。ただ、コスト削減のために推奨します。

### Q5: 既存プロジェクトにも適用可能ですか？

はい。`moai init .`で既存ファイルはそのまま維持されます。

---

## 15. コミュニティ & サポート

### 🌐 参加する

- **Discord (公式)**: [https://discord.gg/umywNygN](https://discord.gg/umywNygN)
- **GitHub**: [https://github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **開発者ブログ**: [https://goos.kim](https://goos.kim)

### 🆘 サポート

- メール: [support@mo.ai.kr](mailto:support@mo.ai.kr)
- ドキュメント: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

---

## 15. Star History

[![Star History Chart](https://api.star-history.com/svg?repos=modu-ai/moai-adk&type=date&legend=top-left)](https://www.star-history.com/#modu-ai/moai-adk&type=date&legend=top-left)

---

## 16. ライセンス

Copyleft License (COPYLEFT-3.0) - [LICENSE](./LICENSE)

---

## 17. 🙏 Made with ❤️ by MoAI-ADK Team

**Last Updated:** 2026-01-28
**Philosophy**: SPEC-First DDD + Agent Orchestration + Hybrid LLM
**MoAI**: MoAIは"みんなのためのAI (Modu-ui AI)"を意味します。

> **"無限可能主義 - みんなのAI"**
