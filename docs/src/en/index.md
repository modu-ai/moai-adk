---
title: MoAI-ADK (Agentic Development Kit)
description: AI-driven SPEC-First TDD development framework providing seamless workflow from specifications through testing, coding, and documentation
lang: en
---

# MoAI-ADK (Agentic Development Kit)

[日本語](index.md) | [English](../en/index.md) | [한국어](../ko/index.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![Coverage](https://img.shields.io/badge/coverage-97.7%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **MoAI-ADKはAIと共に仕様(SPEC) → テスト(TDD) → コード → ドキュメントを自然に繋ぐ開発ワークフローを提供します。**

---

## 1. MoAI-ADKの概要

MoAI-ADKは3つのコア原則でAI協力開発を革新します。以下のナビゲーションであなたの状況に合ったセクションに移動してください。

MoAI-ADKを**初めてお使いの場合**は「MoAI-ADKとは？」から始めてください。
**素早く始めたい場合**は「5分クイックスタート」で直接進めます。
**既にインストールして概念を理解したい場合**は「核心概念の簡単な理解」をおすすめします。

| 質問 | 直ぐに見る |
| --- | --- |
| 初めてですが何ですか？ | [MoAI-ADKとは？](#moai-adkとは) |
| どうやって始めますか？ | [5分クイックスタート](#5分クイックスタート) |
| 基本フローが知りたいです | [基本ワークフロー (0 → 3)](#基本ワークフロー-0--3) |
| Plan / Run / Syncコマンドは何をしますか？ | [核心コマンド要約](#核心コマンド要約) |
| SPEC·TDD·TAGとは何ですか？ | [核心概念の簡単な理解](#核心概念の簡単な理解) |
| エージェント/Skillsが知りたいです | [サブエージェント＆スキル概要](#サブエージェント--スキル概要) |
| Claude Code Hooksが知りたいです | [Claude Code Hooksガイド](#claude-code-hooksガイド) |
| より深く学びたいです | [追加資料](#追加資料) |

---

## MoAI-ADKとは？

### 問題：AI開発の信頼性危機

今日、多くの開発者がClaudeやChatGPTの助けを求めていますが、一つの根本的な疑念を拭えません。**「このAIが作ったコードを本当に信頼できるか？」**

現実はこうです。AIに「ログイン機能を作ってください」と頼むと、文法的に完璧なコードが出てきます。しかし、以下のような問題が繰り返されます：

- **要件不明確**：「正確に何を作るべきか」という基本的な質問が答えられません。メール/パスワードログイン？OAuth？2FAは？すべて推測に依存します。
- **テスト漏れ**：ほとんどのAIは「happy path」のみテストします。間違ったパスワードは？ネットワークエラーは？3ヶ月後のプロダクションでバグが爆発します。
- **ドキュメント不一致**：コードが修正されてもドキュメントはそのままです。「このコードがなぜここにあるの？」という質問が繰り返されます。
- **コンテキスト損失**：同じプロジェクトでも毎回最初から説明する必要があります。プロジェクトの構造、決定理由、以前の試みが記録されません。
- **変更影響把握不可能**：要件が変更された時、どのコードが影響を受けるか追跡できません。

### 解決策：SPEC-First TDD with Alfred SuperAgent

**MoAI-ADK**(MoAI Agentic Development Kit)はこれらの問題を**体系的に解決**するように設計されたオープンソースフレームワークです。

核心原理は単純ですが強力です：

> **「コードがなければテストもなく、テストがなければSPECもない」**

より正確には逆順です：

> **「SPECが先に出る。SPECがなければテストもない。テストとコードがなければドキュメントも完成ではない」**

この順序を守る時、失敗しないエージェンティックコーディングを体験できます：

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span> 明確な要件**
`/alfred:1-plan`コマンドでSPECを先に書きます。「ログイン機能」という曖昧な要求が「WHEN 有効な認証情報が提供されたら、JWTトークンを発行すべき」という**明確な要件**に変換されます。Alfredのspec-builderがEARS文法を使いわずか3分で専門的なSPECを作成してくれます。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span> テスト保証**
`/alfred:2-run`で自動的にテスト駆動開発(TDD)を進めます。RED(失敗するテスト) → GREEN(最小実装) → REFACTOR(コード整理)の順で進み、**テストカバレッジは85%以上を保証**します。もはや「後でテスト」はありません。テストがコード作成をリードします。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span> ドキュメント自動同期**
`/alfred:3-sync`コマンド一つでコード、テスト、ドキュメントがすべて**最新状態で同期**されます。README、CHANGELOG、APIドキュメント、そしてLiving Documentまで自動的に更新されます。6ヶ月後でもコードとドキュメントは一致します。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span> @TAGシステムで追跡**
すべてのコードとテスト、ドキュメントに`@TAG:ID`を付けます。後で要件が変更されたら、`rg "@SPEC:EX-AUTH-001"`の一つのコマンドで関連するテスト、実装、ドキュメントを**すべて見つけられます**。リファクタリング時に自信が生まれます。

**<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_5</span> Alfredがコンテキストを記憶**
AIエージェントたちが協力してプロジェクトの構造、決定理由、作業履歴を**すべて記憶**します。同じ質問を繰り返す必要がありません。

### MoAI-ADKの核心3つの約束

初心者も覚えられるように、MoAI-ADKの価値は3つに単純化されます：

**第一に、SPECがコードより先**
何を作るか明確に定義して始めます。SPECを書いているうちに実装前に問題を発見できます。チームメンバーとの意思疎通コストが大幅に減ります。

**第二に、テストがコードをリードする (TDD)**
実装前にテストを先に書きます(RED)。テストを通過させる最小実装をします(GREEN)。その後コードを整理します(REFACTOR)。結果：バグが少なく、リファクタリングに自信が生まれ、誰でも理解できるコード。

**第三に、ドキュメントとコードは常に一致する**
`/alfred:3-sync`一つのコマンドですべてのドキュメントが自動更新されます。README、CHANGELOG、APIドキュメント、Living Documentがコードと常に同期されます。半年前のコードを修正しようとする時の絶望感がなくなります。

---

## なぜ必要なのか？

### AI開発の現実的な課題

現代のAI協力開発は多様な挑戦に直面しています。MoAI-ADKはこれらすべての問題を**体系的に解決**します：

| 懸念 | 従来方式の問題 | MoAI-ADKの解決 |
| --- | --- | --- |
| 「AIコードを信頼できない」 | テストなしの実装、検証方法不明確 | SPEC → TEST → CODE順序強制、カバレッジ85%+保証 |
| 「毎回同じ説明繰り返し」 | コンテキスト損失、プロジェクト履歴未記録 | Alfredがすべての情報記憶、19個AIチーム協力 |
| 「プロンプト作成困難」 | 良いプロンプトを作る方法を知らない | `/alfred`コマンドが標準化されたプロンプト自動提供 |
| 「ドキュメントが常に古い」 | コード修正後ドキュメント更新忘れ | `/alfred:3-sync`一つのコマンドで自動同期 |
| 「どこを修正したか分からない」 | コード検索困難、意図不明確 | @TAGチェーンでSPEC → TEST → CODE → DOC連結 |
| 「チームオンボーディング時間が長い」 | 新しいチームメンバーがコード文脈把握不可能 | SPECを読めば意図をすぐ理解可能 |

### 今すぐ体験できる利益

MoAI-ADKを導入する瞬間から以下を感じられます：

- **開発速度向上**：明確なSPECで往復説明時間短縮
- **バグ減少**：SPECベーステストで事前発見
- **コード理解度向上**：@TAGとSPECで意図をすぐ把握
- **維持管理コスト削減**：コードとドキュメント常に一致
- **チーム協業効率化**：SPECとTAGで明確なコミュニケーション

---

## ⚡ 3分超高速スタート

MoAI-ADKで**3ステップだけ**で最初のプロジェクトを始めましょう。初心者でも5分以内に完了できます。

### ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span>：インストール（約1分）

#### UVインストールコマンド

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows (PowerShell)
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 実際の出力（例）

```bash
# UVバージョン確認
uv --version
✓ uv 0.5.1 is already installed

$ uv --version
uv 0.5.1
```

#### 次へ：MoAI-ADKインストール

```bash
uv tool install moai-adk

# 結果: <span class="material-icons">check_circle</span> Installed moai-adk
```

**検証**：

```bash
moai-adk --version
# 出力: MoAI-ADK v1.0.0
```

---

### ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span>：最初のプロジェクト作成（約2分）

#### コマンド

```bash
moai-adk init hello-world
cd hello-world
```

#### 実際に作成されるもの

```
hello-world/
├── .moai/              <span class="material-icons">check_circle</span> Alfred設定
├── .claude/            <span class="material-icons">check_circle</span> Claude Code自動化
└── CLAUDE.md           <span class="material-icons">check_circle</span> プロジェクトガイド
```

#### 検証：核心ファイル確認

```bash
# 核心設定ファイル確認
ls -la .moai/config.json  # <span class="material-icons">check_circle</span> 存在するか？
ls -la .claude/commands/  # <span class="material-icons">check_circle</span> コマンドがあるか？

# または一度に確認
moai-adk doctor
```

**出力例**：

```
<span class="material-icons">check_circle</span> Python 3.13.0
<span class="material-icons">check_circle</span> uv 0.5.1
<span class="material-icons">check_circle</span> .moai/ directory initialized
<span class="material-icons">check_circle</span> .claude/ directory ready
<span class="material-icons">check_circle</span> 16 agents configured
<span class="material-icons">check_circle</span> 74 skills loaded
```

---

### ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span>：Alfred開始（約1-2分）

#### Claude Code実行

```bash
claude
```

#### Claude Codeで以下を入力

```
/alfred:0-project
```

#### Alfredが聞くこと

```
Q1: プロジェクト名は？
A: hello-world

Q2: プロジェクト目標は？
A: MoAI-ADK学習

Q3: 主な開発言語は？
A: python

Q4: モードは？
A: personal (ローカル開発用)
```

#### 結果：プロジェクト準備完了！ <span class="material-icons">check_circle</span>

```
<span class="material-icons">check_circle</span> プロジェクト初期化完了
<span class="material-icons">check_circle</span> .moai/config.jsonに設定保存
<span class="material-icons">check_circle</span> .moai/project/にドキュメント作成
<span class="material-icons">check_circle</span> Alfredがスキル推薦完了

次のステップ: /alfred:1-plan "最初の機能説明"
```

---

## <span class="material-icons">rocket_launch</span> 次へ：10分で最初の機能完成

今実際に**機能を作ってドキュメントも自動生成**してみましょう！

> **→ 次のセクション：["最初の10分実践：Hello World API"](#-最初の10分実践-hello-world-api) に移動**

このセクションでは：

- <span class="material-icons">check_circle</span> 簡単なAPIをSPECで定義する
- <span class="material-icons">check_circle</span> TDD (RED → GREEN → REFACTOR)完全体験
- <span class="material-icons">check_circle</span> 自動ドキュメント生成体験
- <span class="material-icons">check_circle</span> @TAGシステム理解

---

[詳細なインストールガイド](getting-started/installation.md) | [クイックスタート](getting-started/quick-start.md) | [概念説明](getting-started/concepts.md) | [Alfredコマンド](guides/alfred/index.md)