---
title: クイックスタート
description: 5分でMoAI-ADKの基本ワークフローを体験するガイド
lang: ja
---

# クイックスタートガイド

MoAI-ADKで**3ステップだけ**で最初のプロジェクトを始めましょう。初心者でも5分以内に完了できます。

## 前提条件

- ✅ MoAI-ADKインストール完了
- ✅ Claude Codeインストール済み
- ✅ Git初期化済み

まだの場合は[インストールガイド](installation.md)を参照してください。

---

## ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span>：プロジェクト作成（2分）

### 新規プロジェクト作成

```bash
# 新しいプロジェクトを作成
moai-adk init hello-api
cd hello-api

# 構造確認
ls -la
```

### 生成されるもの

```
hello-api/
├── .moai/              ✅ Alfred設定
├── .claude/            ✅ Claude Code自動化
└── CLAUDE.md           ✅ プロジェクトガイド
```

### 検証

```bash
# 診断実行
moai-adk doctor
```

期待される出力：
```
✅ Python 3.13.0
✅ uv 0.5.1
✅ .moai/ directory initialized
✅ .claude/ directory ready
✅ 16 agents configured
✅ 74 skills loaded
```

---

## ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span>：Alfred開始（1分）

### Claude Code実行

```bash
claude
```

### プロジェクト初期化

```
/alfred:0-project
```

Alfredが以下を質問します：

```
Q1: プロジェクト名は？
A: hello-api

Q2: プロジェクト目標は？
A: MoAI-ADK学習

Q3: 主な開発言語は？
A: python

Q4: モードは？
A: personal (ローカル開発用)
```

### 結果確認

```
✅ プロジェクト初期化完了
✅ 設定が.moai/config.jsonに保存
✅ .moai/project/にドキュメント作成
✅ Alfredがスキル推薦完了
```

---

## ステップ<span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span>：最初の機能作成（5分）

### SPEC作成（1分）

```bash
/alfred:1-plan "GET /helloエンドポイント - クエリパラメータnameを受け取って挨拶を返す"
```

Alfredが自動生成：
```
✅ SPEC ID: HELLO-001
✅ ファイル: .moai/specs/SPEC-HELLO-001/spec.md
✅ ブランチ: feature/SPEC-HELLO-001
```

### TDD実装（3分）

```bash
/alfred:2-run HELLO-001
```

AlfredがTDDサイクルを自動実行：
- 🔴 **RED**: 失敗するテストを先に作成
- 🟢 **GREEN**: テストを通過させる最小実装
- ♻️ **REFACTOR**: コードを整理・改善

### ドキュメント同期（1分）

```bash
/alfred:3-sync
```

自動的に実行：
```
✅ docs/api/hello.md - APIドキュメント作成
✅ README.md - API使用法追加
✅ CHANGELOG.md - v0.1.0リリースノート追加
✅ @TAGチェーン検証 - すべての@TAG確認
```

---

## 🎉 5分後：あなたが得たもの

### 生成されたファイル

```
hello-api/
├── .moai/specs/SPEC-HELLO-001/
│   ├── spec.md              ← 要件ドキュメント
│   └── plan.md              ← 計画
├── tests/test_hello.py      ← テスト（100%カバレッジ）
├── src/hello/
│   ├── api.py               ← API実装
│   └── __init__.py
├── docs/api/hello.md        ← APIドキュメント
├── README.md                ← 更新済み
└── CHANGELOG.md             ← v0.1.0リリースノート
```

### Git履歴

```bash
git log --oneline | head -4
```

期待される出力：
```
c1d2e3f ♻️ refactor(HELLO-001): add name length validation
b2c3d4e 🟢 feat(HELLO-001): implement hello API
a3b4c5d 🔴 test(HELLO-001): add failing hello API tests
d4e5f6g Merge branch 'develop'
```

### 学んだこと

- ✅ **SPEC**: EARS形式で要件を明確に定義
- ✅ **TDD**: RED → GREEN → REFACTORサイクル体験
- ✅ **自動化**: ドキュメントがコードと一緒に自動生成
- ✅ **追跡性**: @TAGシステムですべてのステップが連結
- ✅ **品質**: テスト100%、明確な実装、自動ドキュメント化

---

## <span class="material-icons">search</span> 検証してみよう

### APIテスト実行

```bash
pytest tests/test_hello.py -v
```

期待される出力：
```
✅ test_hello_with_name_should_return_personalized_greeting PASSED
✅ test_hello_without_name_should_return_default_greeting PASSED
✅ test_hello_with_long_name_should_return_400 PASSED
✅ 3 passed in 0.05s
```

### @TAGチェーン確認

```bash
rg '@(SPEC|TEST|CODE|DOC):HELLO-001' -n
```

期待される出力：
```
.moai/specs/SPEC-HELLO-001/spec.md:7:# @SPEC:EX-HELLO-001: Hello World API
tests/test_hello.py:3:# @TEST:EX-HELLO-002 | SPEC: SPEC-HELLO-001.md
src/hello/api.py:3:# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md
docs/api/hello.md:24:- @SPEC:EX-HELLO-001
```

### 生成されたドキュメント確認

```bash
cat docs/api/hello.md
cat README.md
cat CHANGELOG.md
```

---

## 🚀 次のステップ

### もっと複雑な機能に挑戦

```bash
# 次の機能開始
/alfred:1-plan "ユーザーデータベース照会API"
```

### 学習を深める

- **概念理解**: [概念ガイド](concepts.md)で核心原理を学習
- **Alfredコマンド**: [Alfredガイド](../guides/alfred/index.md)で全コマンドを学習
- **TDD詳説**: [TDDガイド](../guides/tdd/index.md)でテスト駆動開発を深く理解

### 実践例

- **Todo API**: [Todo API例](../guides/project/init.md)で実践的なアプリケーション作成
- **認証システム**: 複雑な認証機能の実装
- **データベース連携**: 永続化データの実装

---

## 💡 ヒントとコツ

### 成功のためのヒント

1. **小さく始める**: 最初は簡単なAPIから
2. **SPECに集中**: 明確な要件が高品質なコードを作る
3. **TDDを信頼**: テストが最初にコードをリードする
4. **頻繁に同期**: `/alfred:3-sync`を定期的に実行
5. **@TAGを活用**: すべてのコードに適切なTAGを付ける

### よくある質問

**Q: 既存プロジェクトに追加できますか？**
A: はい。`moai-adk init .`で既存コードを変更せずに`.moai/`構造のみ追加します。

**Q: テストはどのように実行しますか？**
A: `/alfred:2-run`が先に実行し、必要なら`pytest`などを再実行します。

**Q: ドキュメントが常に最新であることを確認する方法は？**
A: `/alfred:3-sync`が同期レポートを作成します。プルリクエストでレポートを確認してください。

---

## 🎯 成功基準

5分後、以下が達成できれば成功です：

- ✅ MoAI-ADKプロジェクト作成完了
- ✅ 最初のAPI機能実装完了
- ✅ テスト100%通過
- ✅ 自動生成されたドキュメント確認
- ✅ @TAGシステム理解
- ✅ Git履歴にTDDサイクル記録

---

**🎊 おめでとうございます！** あなたは5分でMoAI-ADKの基本ワークフローをマスターしました。次は[概念ガイド](concepts.md)で背后的な原理を学び、より高度な機能に挑戦しましょう。