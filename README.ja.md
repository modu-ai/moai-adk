# MoAI-ADK (Agentic Development Kit)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-87.84%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **MoAI-ADKは、SPEC→TEST (TDD)→コード→ドキュメントをAIとともに滑らかにつなぐ開発ワークフローを提供します。**

---

## 1. MoAI-ADK を一目で

| 質問 | ショートカット |
| --- | --- |
| 初めてです。どんなツール？ | [MoAI-ADKとは？](#moai-adkとは) |
| どうやって始めるの？ | [5分クイックスタート](#5分クイックスタート) |
| 基本フローが知りたい | [コアワークフロー (0 → 3)](#コアワークフロー-0--3) |
| Plan / Run / Sync は何をする？ | [主要コマンドまとめ](#主要コマンドまとめ) |
| SPEC・TDD・TAG とは？ | [主要コンセプトを理解する](#主要コンセプトを理解する) |
| エージェントとSkillsが知りたい | [Sub-agent と Skills の概要](#sub-agent-と-skills-の概要) |
| もっと学びたい | [追加リソース](#追加リソース) |

---

## MoAI-ADKとは？

MoAI-ADK（MoAI Agentic Development Kit）は、**開発プロセスのすべてのステップにAIを組み込むオープンソースのツールキット**です。Alfred SuperAgentが「まずSPECを作り、テスト（TDD）で検証し、ドキュメントとコードを常に同期させる」という原則を代わりに守ってくれます。

初めて使う場合は、次の3つだけ覚えてください。

1. **何を作るのか（SPEC）** を先に言語化する。
2. **テストを先に書き（TDD）**、その後で実装する。
3. **ドキュメント / README / CHANGELOG** を自動で最新に保つ。

この流れを4つの `/alfred` コマンドで繰り返せば、プロジェクト全体の整合性が保たれます。

---

## なぜ必要なのか？

| 課題 | MoAI-ADK の支援内容 |
| --- | --- |
| 「AIが書いたコードを信頼しづらい」 | SPEC → TEST → IMPLEMENTATION → DOCS を TAG チェーンで連結 |
| 「文脈がなく毎回同じ質問になる」 | Alfred が主要ドキュメントと履歴を覚えて案内 |
| 「プロンプト作成が難しい」 | `/alfred` コマンドと準備済み Skills が標準プロンプトを提供 |
| 「ドキュメントとコードが乖離する」 | `/alfred:3-sync` が README／CHANGELOG／Living Doc を自動整合 |

---

## 5分クイックスタート

```bash
# 1.（任意）uv をインストール — pip より速い Python パッケージマネージャー
curl -LsSf https://astral.sh/uv/install.sh | sh

# 2. MoAI-ADK をインストール（tool モード: グローバル隔離実行）
uv tool install moai-adk

# 3. 新規プロジェクトを開始
moai-adk init my-project
cd my-project

# 4. Claude Code（または CLI）から Alfred を呼び出す
claude  # Claude Code を起動して以下のコマンドを実行
/alfred:0-project "プロジェクト名"
```

> 🔍 確認コマンド: `moai-adk doctor` — Python/uv バージョン、`.moai/` 構造、エージェント／Skills の準備状況をチェックします。

---

## MoAI-ADK を最新に保つ

### 現在のバージョンを確認
```bash
# インストールされているバージョンを確認
moai-adk --version

# PyPI の最新バージョンを確認
uv tool list  # moai-adk の現在のバージョンを表示
```

### アップグレード方法

#### 方法1: 個別ツールのみアップグレード（推奨）
```bash
# moai-adk だけを最新バージョンにアップグレード
uv tool upgrade moai-adk
```

#### 方法2: インストール済みツールを一括アップグレード
```bash
# uv tool の全ツールを最新バージョンに更新
uv tool update
```

#### 方法3: 指定バージョンをインストール
```bash
# 指定バージョンを再インストール（例: 0.4.2）
uv tool install moai-adk==0.4.2
```

### アップデート後の確認
```bash
# インストール済みバージョンを確認
moai-adk --version

# プロジェクトが正常稼働するか確認
moai-adk doctor

# 既存プロジェクトに新テンプレートを適用（任意）
moai-adk init .  # 既存コードは維持し、.moai/ 構造のみ更新
```

> 💡 **Tip**: メジャー／マイナー更新があったら `moai-adk init .` を実行し、最新のエージェント／Skills／テンプレートを取り込んでください。既存のコードやカスタマイズは安全です。

---

## コアワークフロー (0 → 3)

Alfred は4つのコマンドでプロジェクトを推進します。

```mermaid
%%{init: {'theme':'neutral'}}%%
graph TD
    Start([ユーザー要求]) --> Init[0. Init<br/>/alfred:0-project]
    Init --> Plan[1. Plan & SPEC<br/>/alfred:1-plan]
    Plan --> Run[2. Run & TDD<br/>/alfred:2-run]
    Run --> Sync[3. Sync & Docs<br/>/alfred:3-sync]
    Sync --> Plan
    Sync -.-> End([リリース])
```

### 0. INIT — プロジェクト準備
- プロジェクト紹介、ターゲット、言語、モード（ロケール）をヒアリング
- `.moai/config.json` と `.moai/project/*` の5文書を自動生成
- 言語検出と推奨 Skill Pack（Foundation + Essentials + Domain/Language）を配置
- テンプレート整理、初期 Git／バックアップチェックを実施

### 1. PLAN — 方向性を揃える
- `/alfred:1-plan` が EARS 形式 SPEC（`@SPEC:ID` 付き）を作成
- Plan Board、実装アイデア、リスク整理を生成
- チームモードでは自動でブランチ／Draft PR を作成

### 2. RUN — テスト駆動開発
- フェーズ1 `implementation-planner`: ライブラリ、フォルダー、TAG 設計
- フェーズ2 `tdd-implementer`: RED（失敗テスト）→ GREEN（最小実装）→ REFACTOR
- `quality-gate` が TRUST 5 の原則とカバレッジ変化を検証

### 3. SYNC — ドキュメントと PR を整理
- Living Doc、README、CHANGELOG などを同期
- TAG チェーンを検証し、孤立した TAG を復旧
- Sync Report を生成し、Draft → Ready for Review、`--auto-merge` に対応

---

## 主要コマンドまとめ

| コマンド | 内容 | 主なアウトプット |
| --- | --- | --- |
| `/alfred:0-project` | プロジェクト説明の収集、設定／文書生成、Skill 提案 | `.moai/config.json`, `.moai/project/*`, 初期レポート |
| `/alfred:1-plan <説明>` | 要件分析、SPEC 下書き、Plan Board 作成 | `.moai/specs/SPEC-*/spec.md`, plan/acceptance 文書, フィーチャーブランチ |
| `/alfred:2-run <SPEC-ID>` | TDD 実行、テスト／実装／リファクタリング、品質検証 | `tests/`, `src/` 実装, 品質レポート, TAG 連携 |
| `/alfred:3-sync` | ドキュメント／README／CHANGELOG 同期、TAG／PR 状態を整理 | `docs/`, `.moai/reports/sync-report.md`, レビュー準備済み PR |

> ❗ すべてのコマンドは **Phase 0（任意）→ Phase 1 → Phase 2 → Phase 3** のループを守ります。Alfred が現在の状況と次のステップを自動で報告します。

---

## 主要コンセプトを理解する

### SPEC-First
- **なぜ?** 家を建てる前に設計図が必要なように、実装前に要件を整理します。
- **どうやって?** `/alfred:1-plan` が “WHEN… THEN…” 構造を持つ EARS 形式 SPEC を生成します。
- **結果:** `@SPEC:ID` 付きの文書 + Plan Board + 受け入れ基準。

### TDD (Test-Driven Development)
- **RED**: まず失敗するテストを書く。
- **GREEN**: テストが通る最小限のコードを書く。
- **REFACTOR**: 構造を整え重複を除去する。
- `/alfred:2-run` がフローを自動化し、RED/GREEN/REFACTOR のログを残します。

### TAG システム
- `@SPEC:ID` → `@TEST:ID` → `@CODE:ID` → `@DOC:ID` を連結します。
- TAG を検索すれば関連する SPEC・テスト・ドキュメントを一括で追跡可能。
- `/alfred:3-sync` が TAG インベントリを確認し、孤立した TAG を通知します。

### TRUST 5 原則
1. **Test First** — テストを必ず先に書く
2. **Readable** — 短い関数と一貫したスタイルを保つ
3. **Unified** — アーキテクチャと型／契約を整合させる
4. **Secured** — 入力検証、情報保護、静的解析を行う
5. **Trackable** — TAG、Git 履歴、ドキュメントを連動させる

> 詳細なルールは `.moai/memory/development-guide.md` を参照してください。

---

## 最初のハンズオン: Todo API 例

1. **Plan**
   ```bash
   /alfred:1-plan "Todo の追加・取得・更新・削除 API"
   ```
   Alfred が SPEC (`.moai/specs/SPEC-TODO-001/spec.md`) と plan/acceptance 文書を生成します。

2. **Run**
   ```bash
   /alfred:2-run TODO-001
   ```
   テスト (`tests/test_todo_api.py`)、実装 (`src/todo/`)、レポートが自動で作成されます。

3. **Sync**
   ```bash
   /alfred:3-sync
   ```
   `docs/api/todo.md`、TAG チェーン、Sync Report を更新します。

4. **確認コマンド**
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):TODO-001' -n
   pytest tests/test_todo_api.py -v
   cat docs/api/todo.md
   ```

> わずか15分で、SPEC → TDD → ドキュメントが連携した Todo API を完成できます。

---

## Sub-agent と Skills の概要

Alfred は **19名のチーム**（SuperAgent 1 + コア Sub-agent 10 + 0-project Sub-agent 6 + ビルトイン 2）と **44個の Claude Skills** を組み合わせて動作します。

### コア Sub-agent（Plan → Run → Sync）

| Sub-agent | モデル | 役割 |
| --- | --- | --- |
| project-manager 📋 | Sonnet | プロジェクト初期化とメタデータインタビュー |
| spec-builder 🏗️ | Sonnet | Plan Board 作成、EARS SPEC 執筆 |
| code-builder 💎 | Sonnet | `implementation-planner` と `tdd-implementer` で TDD を一貫実行 |
| doc-syncer 📖 | Haiku | Living Doc、README、CHANGELOG を同期 |
| tag-agent 🏷️ | Haiku | TAG インベントリ管理、孤立 TAG の検出 |
| git-manager 🚀 | Haiku | GitFlow、Draft/Ready、Auto Merge を管理 |
| debug-helper 🔍 | Sonnet | 失敗分析と fix-forward 戦略の提案 |
| trust-checker ✅ | Haiku | TRUST 5 の品質ゲートを検証 |
| quality-gate 🛡️ | Haiku | カバレッジ変化とリリース阻害要因をレビュー |
| cc-manager 🛠️ | Sonnet | Claude Code セッション最適化と Skills 配備 |

### Skills（段階的開示）
- **Foundation (6)**: TRUST, TAG, SPEC, EARS, Git, 言語検出
- **Essentials (4)**: Debug, Refactor, Review, Performance
- **Domain (10)**: Backend, Web API, Security, Data, Mobile など
- **Language (23)**: Python, TypeScript, Go, Rust, Java, Swift など主要言語パック
- **Claude Code Ops (1)**: セッション設定、出力スタイル管理

> Skills は `.claude/skills/` に 500語以下のガイドとして保存されています。必要なときだけ読み込み、コンテキストコストを抑えます。

---

## AI モデル選択ガイド

| シーン | デフォルトモデル | 理由 |
| --- | --- | --- |
| SPEC／設計／リファクタリング／問題解決 | **Claude 4.5 Sonnet** | 深い推論と構造化された文章が得意 |
| ドキュメント同期、TAG チェック、Git 自動化 | **Claude 4.5 Haiku** | 繰り返しの速い作業や文字列処理が得意 |

- パターン化された作業は Haiku から始め、判断が難しい場合は Sonnet に切り替えましょう。
- 手動でモデルを切り替えた場合は、理由をログに残すとチームの理解が進みます。

---

## よくある質問 (FAQ)

- **Q. 既存プロジェクトにも導入できますか？**  
  A. はい。`moai-adk init .` を実行すれば `.moai/` 構造だけを追加し、コードは触りません。
- **Q. テストはどうやって実行しますか？**  
  A. まず `/alfred:2-run` が実行します。必要に応じて `pytest` や `pnpm test` などを再実行してください。
- **Q. ドキュメントが常に最新か確認するには？**  
  A. `/alfred:3-sync` が Sync Report を生成します。Pull Request で確認してください。
- **Q. 手動で進めてもいいですか？**  
  A. 可能ですが、SPEC → TEST → CODE → DOC の順序と TAG の付与は必須です。

---

## 追加リソース

| 目的 | リソース |
| --- | --- |
| Skills 詳細構造 | `docs/skills/overview.md` および Tier ごとのドキュメント |
| Sub-agent 詳細 | `docs/agents/overview.md` |
| ワークフローガイド | `docs/guides/workflow/`（Plan/Run/Sync） |
| 開発ガイドライン | `.moai/memory/development-guide.md`, `.moai/memory/spec-metadata.md` |
| アップデート計画 | `CHANGELOG.md`, `UPDATE-PLAN-0.4.0.md` |

---

## コミュニティとサポート

- GitHub リポジトリ: <https://github.com/modu-ai/moai-adk>
- Issues & Discussions: バグ報告、機能要望、アイデアを歓迎します。
- PyPI: <https://pypi.org/project/moai-adk/>
- 連絡先: `CONTRIBUTING.md` のガイドラインを参照してください。

> 🙌 「SPEC がなければ CODE もない」— Alfred とともに、一貫した AI 開発文化を体験しましょう。

